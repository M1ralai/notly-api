package dto

import (
	"time"

	"github.com/M1ralai/notly-api/internal/modules/pomodoro/domain"
)

type PomodoroSessionResponse struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	CourseID        *int      `json:"course_id,omitempty"`
	DurationMinutes int       `json:"duration_minutes"`
	Notes           string    `json:"notes,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type PomodoroSettingsResponse struct {
	UserID       int       `json:"user_id"`
	StudyPresets []int     `json:"study_presets"`
	StudyColor   string    `json:"study_color"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func ToSessionResponse(s *domain.PomodoroSession) *PomodoroSessionResponse {
	if s == nil {
		return nil
	}
	return &PomodoroSessionResponse{
		ID:              s.ID,
		UserID:          s.UserID,
		CourseID:        s.CourseID,
		DurationMinutes: s.DurationMinutes,
		Notes:           s.Notes,
		CreatedAt:       s.CreatedAt,
	}
}

func ToSessionResponseList(sessions []domain.PomodoroSession) []*PomodoroSessionResponse {
	if sessions == nil {
		return nil
	}
	result := make([]*PomodoroSessionResponse, len(sessions))
	for i, s := range sessions {
		sCopy := s
		result[i] = ToSessionResponse(&sCopy)
	}
	return result
}

func ToSettingsResponse(s *domain.PomodoroSettings) *PomodoroSettingsResponse {
	if s == nil {
		return nil
	}
	return &PomodoroSettingsResponse{
		UserID:       s.UserID,
		StudyPresets: s.StudyPresets,
		StudyColor:   s.StudyColor,
		UpdatedAt:    s.UpdatedAt,
	}
}
