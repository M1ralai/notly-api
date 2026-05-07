-- Remove old semester string field
ALTER TABLE courses DROP COLUMN IF EXISTS semester;
DROP INDEX IF EXISTS idx_courses_semester;
