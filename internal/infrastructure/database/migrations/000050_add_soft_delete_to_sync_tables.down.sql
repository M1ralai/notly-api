-- 000050_add_soft_delete_to_sync_tables.down.sql

-- Drop indexes
DROP INDEX IF EXISTS idx_events_deleted_at;
DROP INDEX IF EXISTS idx_courses_deleted_at;
DROP INDEX IF EXISTS idx_goals_deleted_at;
DROP INDEX IF EXISTS idx_life_areas_deleted_at;
DROP INDEX IF EXISTS idx_habits_deleted_at;
DROP INDEX IF EXISTS idx_tasks_deleted_at;

-- Drop columns
ALTER TABLE milestones DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE events DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE semesters DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE course_schedules DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE course_components DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE courses DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE goals DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE life_areas DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE habits DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE tasks DROP COLUMN IF EXISTS deleted_at;
