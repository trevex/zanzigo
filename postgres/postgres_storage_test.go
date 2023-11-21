package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/trevex/zanzigo"
	"github.com/trevex/zanzigo/testsuite"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
)

var (
	databaseURL = ""
	storage     zanzigo.Storage
)

func TestMain(m *testing.M) {
	var (
		pool     *dockertest.Pool
		resource *dockertest.Resource
		err      error
	)

	databaseURL = os.Getenv("TEST_POSTGRES_DATABASE_URL")

	if databaseURL == "" {
		pool, err = dockertest.NewPool("")
		if err != nil {
			log.Fatalf("Could not connect to docker: %s", err)
		}

		resource, err = pool.RunWithOptions(&dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "15.4",
			Env: []string{
				"POSTGRES_PASSWORD=zanzigo",
				"POSTGRES_USER=zanzigo",
				"POSTGRES_DB=zanzigo",
				"listen_addresses = '*'",
			},
		}, func(config *docker.HostConfig) {
			config.AutoRemove = true // Stopped container should be removed
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		})
		if err != nil {
			log.Fatalf("Could not start resource: %s", err)
		}
		_ = resource.Expire(300) // In any case container should be killed in 5min

		hostAndPort := resource.GetHostPort("5432/tcp")
		databaseURL = fmt.Sprintf("postgres://zanzigo:zanzigo@%s/zanzigo?sslmode=disable", hostAndPort)

		// We connect with exponential backoff (maximum wait 2min)
		pool.MaxWait = 120 * time.Second
		if err = pool.Retry(func() error {
			db, err := sql.Open("pgx", databaseURL)
			if err != nil {
				return err
			}
			return db.Ping()
		}); err != nil {
			log.Fatalf("Could not connect to postgres: %s", err)
		}
	}

	if err := RunMigrations(databaseURL); err != nil {
		log.Fatalf("Could not migrate db: %s", err)
	}

	storage, err = NewPostgresStorage(databaseURL)
	if err != nil {
		log.Fatalf("PostgresStorage creation failed: %v", err)
	}

	// Let's load the testsuite-data
	err = testsuite.Load(context.Background(), storage)
	if err != nil {
		log.Fatalf("Failed loading data into storage: %v", err)
	}

	code := m.Run()

	// os.Exit doesn't care for defer, so let's explicitly purge and close...
	storage.Close()
	if pool != nil {
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}

	os.Exit(code)
}

func TestPostgresWithTestSuite(t *testing.T) {
	storageFunctions, err := NewPostgresStorage(databaseURL, UseFunctions())
	if err != nil {
		t.Fatalf("PostgresStorage creation failed: %v", err)
	}
	defer storageFunctions.Close()
	testsuite.RunTestAll(t, map[string]testsuite.TestConfig{
		"queries": testsuite.TestConfig{
			Storage: storage,
			Expectations: testsuite.Expectations{
				UserdataCheckQueryTuple: zanzigo.MarkedTuple{
					CheckIndex: 0,
					RuleIndex:  2,
					Tuple:      zanzigo.TupleString("doc:mydoc#parent@folder:myfolder"),
				},
			},
		},
		"functions": testsuite.TestConfig{
			Storage: storageFunctions,
			Expectations: testsuite.Expectations{
				UserdataCheckQueryTuple: zanzigo.MarkedTuple{
					CheckIndex: 0,
					RuleIndex:  0,
					Tuple:      zanzigo.TupleString("doc:mydoc#viewer@user:myuser"),
				},
			},
		},
	})
}

func TestPostgresQueryBuilding(t *testing.T) {
	resolver, err := zanzigo.NewSequentialResolver(testsuite.Model, storage, 16)
	require.NoError(t, err)

	ruleset := resolver.RulesetFor("doc", "viewer")
	query := newPostgresQuery(ruleset)
	expectedQuery := standardizeSpaces(`
		(SELECT 0 AS rule_index, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='doc' AND object_id=%s AND (object_relation='editor' OR object_relation='owner' OR object_relation='viewer') AND subject_type=%s AND subject_id=%s AND subject_relation=%s)
		UNION ALL
		(SELECT 1 AS rule_index, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='doc' AND object_id=%s AND (object_relation='editor' OR object_relation='owner' OR object_relation='viewer') AND subject_relation <> '')
		UNION ALL
		(SELECT 2 AS rule_index, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='doc' AND object_id=%s AND (object_relation='parent') AND subject_type='folder')
	`)
	require.Equal(t, expectedQuery, query.query)
}

func TestPostgresFunctionBuilding(t *testing.T) {
	storageFunctions, err := NewPostgresStorage(databaseURL, UseFunctions())
	require.NoError(t, err)
	defer storageFunctions.Close()

	resolver, err := zanzigo.NewSequentialResolver(testsuite.Model, storageFunctions, 16)
	require.NoError(t, err)

	ruleset := resolver.RulesetFor("doc", "viewer")

	decl, query := postgresFunctionFor("doc", "viewer", ruleset)
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

	_, err = storageFunctions.newPostgresFunction("doc", "viewer", ruleset)
	if err != nil {
		t.Fatalf("Expected function to be created, but failed with: %v", err)
	}

}

func BenchmarkPostgres(b *testing.B) {
	storageFunctions, err := NewPostgresStorage(databaseURL, UseFunctions())
	require.NoError(b, err)

	defer storageFunctions.Close()
	testsuite.RunBenchmarkAll(b, map[string]zanzigo.Storage{
		"queries":   storage,
		"functions": storageFunctions,
	})
}

func standardizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
