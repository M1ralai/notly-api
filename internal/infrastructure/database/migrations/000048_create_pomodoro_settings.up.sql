CREATE TABLE pomodoro_settings (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    study_presets INTEGER[] NOT NULL DEFAULT '{25, 45}',
    study_color VARCHAR(50) NOT NULL DEFAULT 'gray',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
