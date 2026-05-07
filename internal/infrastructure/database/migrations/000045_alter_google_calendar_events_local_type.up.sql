ALTER TABLE google_calendar_events DROP CONSTRAINT google_calendar_events_local_type_check;
ALTER TABLE google_calendar_events ALTER COLUMN local_type TYPE VARCHAR(50);
ALTER TABLE google_calendar_events ADD CONSTRAINT google_calendar_events_local_type_check CHECK (local_type IN ('task', 'habit', 'course_schedule'));
