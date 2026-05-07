package domain

import (
	"context"
	"time"
)

type PomodoroSession struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	CourseID        *int      `json:"course_id"`
	DurationMinutes int       `json:"duration_minutes"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
}

type PomodoroSettings struct {
	UserID       int       `json:"user_id"`
	StudyPresets []int     `json:"study_presets"`
	StudyColor   string    `json:"study_color"`
	UpdatedAt    time.Time `json:"updated_at"`
}


type Repository interface {
	Save(ctx context.Context, session *PomodoroSession) error
	FindAllByUserID(ctx context.Context, userID int) ([]PomodoroSession, error)
	FindByCourseID(ctx context.Context, userID, courseID int) ([]PomodoroSession, error)
	GetSettings(ctx context.Context, userID int) (*PomodoroSettings, error)
	UpdateSettings(ctx context.Context, settings *PomodoroSettings) error
}
