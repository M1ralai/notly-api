ALTER TABLE course_resources
  DROP COLUMN IF EXISTS mime_type,
  DROP COLUMN IF EXISTS file_size_bytes;
