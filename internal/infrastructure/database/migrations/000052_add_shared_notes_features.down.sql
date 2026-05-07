-- 000052_add_shared_notes_features.down.sql

DROP TABLE IF EXISTS note_collaborators;
DROP TABLE IF EXISTS note_attachments;

ALTER TABLE notes
    DROP COLUMN IF EXISTS share_token,
    DROP COLUMN IF EXISTS is_public;
