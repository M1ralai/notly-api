-- 000051_align_events_schema_for_sync.up.sql

-- Add missing life_area_id column
ALTER TABLE events ADD COLUMN IF NOT EXISTS life_area_id INTEGER REFERENCES life_areas(id) ON DELETE SET NULL;

-- Rename columns to match repository naming convention
DO $$ 
BEGIN 
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'events' AND column_name = 'start_datetime') THEN
        ALTER TABLE events RENAME COLUMN start_datetime TO start_time;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'events' AND column_name = 'end_datetime') THEN
        ALTER TABLE events RENAME COLUMN end_datetime TO end_time;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'events' AND column_name = 'all_day') THEN
        ALTER TABLE events RENAME COLUMN all_day TO is_all_day;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'events' AND column_name = 'recurrence_rule') THEN
        ALTER TABLE events RENAME COLUMN recurrence_rule TO recurrence;
    END IF;
END $$;

-- Add index for life_area_id
CREATE INDEX IF NOT EXISTS idx_events_life_area ON events(life_area_id);
