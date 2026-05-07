DROP INDEX IF EXISTS idx_courses_semester_id;
ALTER TABLE courses DROP COLUMN IF EXISTS semester_id;
