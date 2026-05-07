-- 000052_add_shared_notes_features.up.sql
-- ============================================================
-- Phase 1 – Media-first shared notes infrastructure
-- ============================================================

-- 1. Extend the existing notes table with sharing fields
ALTER TABLE notes
    ADD COLUMN IF NOT EXISTS is_public   BOOLEAN      NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS share_token VARCHAR(64)  UNIQUE;

-- 2. Attachment table – one note can have many media files
CREATE TABLE IF NOT EXISTS note_attachments (
    id              BIGSERIAL    PRIMARY KEY,
    note_id         BIGINT       NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    file_url        TEXT         NOT NULL,
    object_key      TEXT         NOT NULL,          -- MinIO key used for deletion
    file_type       VARCHAR(64)  NOT NULL,           -- image/jpeg, application/pdf …
    file_size_bytes BIGINT       NOT NULL DEFAULT 0,
    original_name   VARCHAR(255) NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_note_attachments_note_id ON note_attachments(note_id);

-- 3. Collaborator table – viewer / editor granularity
CREATE TABLE IF NOT EXISTS note_collaborators (
    id         BIGSERIAL    PRIMARY KEY,
    note_id    BIGINT       NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    user_id    BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role       VARCHAR(16)  NOT NULL DEFAULT 'viewer' CHECK (role IN ('viewer', 'editor')),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (note_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_note_collaborators_note_id ON note_collaborators(note_id);
CREATE INDEX IF NOT EXISTS idx_note_collaborators_user_id ON note_collaborators(user_id);
