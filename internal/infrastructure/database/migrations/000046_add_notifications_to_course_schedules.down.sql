ALTER TABLE course_schedules
    DROP COLUMN IF EXISTS notifications_enabled,
    DROP COLUMN IF EXISTS notification_type,
    DROP COLUMN IF EXISTS reminder_time;
