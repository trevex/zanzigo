
CREATE TABLE tuples (
    uuid TEXT NOT NULL,
    object_type TEXT NOT NULL,
    object_id TEXT NOT NULL,
    object_relation TEXT NOT NULL,
    subject_type TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    subject_relation TEXT NOT NULL,
    PRIMARY KEY (object_type, object_id, object_relation, subject_type, subject_id, subject_relation)
);

CREATE INDEX idx_tuple ON tuples (object_type, object_id, object_relation, subject_type, subject_id, subject_relation);
CREATE INDEX idx_tuples_partial_for_usersets ON tuples (object_type, object_id, object_relation) WHERE subject_relation <> '';
CREATE INDEX idx_tuples_partial_for_indirect ON tuples (object_type, object_id, subject_type);
CREATE UNIQUE INDEX idx_tuples_uuid ON tuples (uuid);
-- We also want to speed up common list operations
CREATE INDEX idx_tuples_partial_for_list_obj ON tuples (object_type, object_id);
CREATE INDEX idx_tuples_partial_for_list_obj_rel ON tuples (object_type, object_id, object_relation);
CREATE INDEX idx_tuples_partial_for_list_sub ON tuples (subject_type, subject_id);
CREATE INDEX idx_tuples_partial_for_list_sub_rel ON tuples (subject_type, subject_id, subject_relation);
