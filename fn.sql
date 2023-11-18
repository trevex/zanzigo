CREATE FUNCTION zanzigo_doc_viewer(TEXT, TEXT, TEXT, TEXT) RETURNS BOOLEAN AS $$
DECLARE
    mt RECORD
    query TEXT
BEGIN
    FOR mt IN
        (SELECT 1 AS command_type, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='doc' AND object_id=$1 AND (object_relation='editor' OR object_relation='owner' OR object_relation='viewer') AND subject_type=$2 AND subject_id=$3 AND subject_relation=$4) 
        UNION ALL 
        (SELECT 2 AS command_type, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='doc' AND object_id=$1 AND (object_relation='editor' OR object_relation='owner' OR object_relation='viewer') AND subject_relation <> '') 
        UNION ALL 
        (SELECT 3 AS command_type, object_type, object_id, object_relation, subject_type, subject_id, subject_relation FROM tuples WHERE object_type='doc' AND object_id=$1 AND (object_relation='parent') AND subject_type='folder') ORDER_BY command_type
    LOOP
        IF mt.command_type = 1 THEN
            RETURN TRUE;
        ELSIF mt.command_type = 2 THEN
            query = query||FORMAT('zanzigo_%s_%s(%s, %s, %s, %s)', mt.subject_type, mt.subject_relation, mt.subject_id, $2, $3, $4)||' OR ';
        ELSE
            /* LOOP OVER relations to subjects in code */
            query = query||FORMAT('zanzigo_%s_GENERATED(%s. %s. %s. %s)', mt.subject_type. mt.subject_id, $2, $3, $4)||' OR ';
        END IF
    END LOOP;
    query = 'FALSE';
    RETURN EXECUTE query;
END;
$$ LANGUAGE 'plpgsql' STABLE;
