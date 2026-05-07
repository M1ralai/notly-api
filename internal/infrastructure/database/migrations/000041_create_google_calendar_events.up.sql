-- google_calendar_events table: Maps local habit/task completions to Google Calendar events
CREATE TABLE google_calendar_events (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    local_id INTEGER NOT NULL,
    local_type VARCHAR(10) NOT NULL CHECK (local_type IN ('task', 'habit')),
    google_event_id VARCHAR(255) NOT NULL,
    event_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, local_id, local_type, event_date)
);

CREATE INDEX idx_google_calendar_events_user ON google_calendar_events(user_id);
CREATE INDEX idx_google_calendar_events_lookup ON google_calendar_events(user_id, local_id, local_type, event_date);
