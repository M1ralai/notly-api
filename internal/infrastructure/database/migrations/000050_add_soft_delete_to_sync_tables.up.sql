-- 000050_add_soft_delete_to_sync_tables.up.sql

ALTER TABLE tasks ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP NULL;
ALTER TABLE habits ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP NULL;
ALTER TABLE life_areas ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP NULL;
ALTER TABLE life_areas ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW();
ALTER TABLE goals ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP NULL;

-- For academic/course management
ALTER TABLE courses ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP NULL;
ALTER TABLE course_components ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP NULL;
ALTER TABLE course_schedules ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP NULL;
ALTER TABLE semesters ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP NULL;

-- For events and dates
ALTER TABLE events ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP NULL;
ALTER TABLE milestones ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP NULL;

-- Add indexes on deleted_at to optimize sync queries
CREATE INDEX IF NOT EXISTS idx_tasks_deleted_at ON tasks(deleted_at);
CREATE INDEX IF NOT EXISTS idx_habits_deleted_at ON habits(deleted_at);
CREATE INDEX IF NOT EXISTS idx_life_areas_deleted_at ON life_areas(deleted_at);
CREATE INDEX IF NOT EXISTS idx_goals_deleted_at ON goals(deleted_at);
CREATE INDEX IF NOT EXISTS idx_courses_deleted_at ON courses(deleted_at);
CREATE INDEX IF NOT EXISTS idx_events_deleted_at ON events(deleted_at);
