ALTER TABLE course_schedules
    ADD COLUMN IF NOT EXISTS notifications_enabled BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS notification_type TEXT DEFAULT 'popup',
    ADD COLUMN IF NOT EXISTS reminder_time INTEGER DEFAULT 15;
