-- 000051_align_events_schema_for_sync.down.sql

-- Drop index
DROP INDEX IF EXISTS idx_events_life_area;

-- Revert column names
DO $$ 
BEGIN 
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'events' AND column_name = 'start_time') THEN
        ALTER TABLE events RENAME COLUMN start_time TO start_datetime;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'events' AND column_name = 'end_time') THEN
        ALTER TABLE events RENAME COLUMN end_time TO end_datetime;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'events' AND column_name = 'is_all_day') THEN
        ALTER TABLE events RENAME COLUMN is_all_day TO all_day;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'events' AND column_name = 'recurrence') THEN
        ALTER TABLE events RENAME COLUMN recurrence TO recurrence_rule;
    END IF;
END $$;

-- Drop life_area_id column
ALTER TABLE events DROP COLUMN IF EXISTS life_area_id;
