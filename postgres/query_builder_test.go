package postgres

import (
	"testing"

	"github.com/trevex/zanzigo"
	"github.com/trevex/zanzigo/testsuite"

	"github.com/stretchr/testify/require"
)

func TestPostgresSelectQueryFor(t *testing.T) {
	resolver, err := zanzigo.NewResolver(testsuite.Model, storage, 16)
	require.NoError(t, err)

	ruleset := resolver.RulesetFor("doc", "viewer")
	query, err := SelectQueryFor(ruleset, true, "$%d")
	require.NoError(t, err)
	expectedQuery := standardizeSpaces(`
		(SELECT 0 AS rule_index, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='doc' AND object_id=$%d AND (object_relation='editor' OR object_relation='owner' OR object_relation='viewer') AND subject_type=$%d AND subject_id=$%d AND subject_relation=$%d)
		UNION ALL
		(SELECT 1 AS rule_index, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='doc' AND object_id=$%d AND (object_relation='editor' OR object_relation='owner' OR object_relation='viewer') AND subject_relation <> '')
		UNION ALL
		(SELECT 2 AS rule_index, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='doc' AND object_id=$%d AND (object_relation='parent') AND subject_type='folder')
	`)
	require.Equal(t, expectedQuery, query)
}

func TestPostgresFunctionFor(t *testing.T) {
	storageFunctions, err := NewPostgresStorage(databaseURL, UseFunctions())
	require.NoError(t, err)
	defer storageFunctions.Close()

	resolver, err := zanzigo.NewResolver(testsuite.Model, storageFunctions, 16)
	require.NoError(t, err)

	ruleset := resolver.RulesetFor("doc", "viewer")

	decl, query, err := FunctionFor("zanzigo_doc_viewer", ruleset)
	require.NoError(t, err)
	expectedQuery := `SELECT zanzigo_doc_viewer($1, $2, $3, $4)`
	decl = standardizeSpaces(decl)
	expectedDecl := standardizeSpaces(`
CREATE OR REPLACE FUNCTION zanzigo_doc_viewer(TEXT, TEXT, TEXT, TEXT) RETURNS BOOLEAN LANGUAGE 'plpgsql' AS $$
DECLARE
	mt RECORD;
	result BOOLEAN;
BEGIN
	FOR mt IN
		(SELECT 0 AS rule_index, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='doc' AND object_id=$1 AND (object_relation='editor' OR object_relation='owner' OR object_relation='viewer') AND subject_type=$2 AND subject_id=$3 AND subject_relation=$4)
		UNION ALL
		(SELECT 1 AS rule_index, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='doc' AND object_id=$1 AND (object_relation='editor' OR object_relation='owner' OR object_relation='viewer') AND subject_relation <> '')
		UNION ALL
		(SELECT 2 AS rule_index, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='doc' AND object_id=$1 AND (object_relation='parent') AND subject_type='folder') ORDER BY rule_index
	LOOP
		IF mt.rule_index = 0 THEN
			RETURN TRUE;
		ELSIF mt.rule_index = 1 THEN
			EXECUTE FORMAT('SELECT zanzigo_%s_%s($1, $2, $3, $4)', mt.subject_type, mt.subject_relation) USING mt.subject_id, $2, $3, $4 INTO result;
			IF result = TRUE THEN
				RETURN TRUE;
			END IF;
		ELSIF mt.rule_index = 2 THEN
			SELECT zanzigo_folder_editor(mt.subject_id, $2, $3, $4) INTO result;
			IF result = TRUE THEN
				RETURN TRUE;
			END IF;
			SELECT zanzigo_folder_owner(mt.subject_id, $2, $3, $4) INTO result;
			IF result = TRUE THEN
				RETURN TRUE;
			END IF;
			SELECT zanzigo_folder_viewer(mt.subject_id, $2, $3, $4) INTO result;
			IF result = TRUE THEN
				RETURN TRUE;
			END IF;
		END IF;
	END LOOP;
	RETURN FALSE;
END;
$$;`)
	require.Equal(t, expectedDecl, decl)
	require.Equal(t, expectedQuery, query)

	_, err = storageFunctions.createOrReplaceFunctionFor("doc", "viewer", ruleset)
	if err != nil {
		t.Fatalf("Expected function to be created, but failed with: %v", err)
	}

}
