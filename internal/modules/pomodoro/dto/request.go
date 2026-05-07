package dto

type UpdatePomodoroSettingsRequest struct {
	StudyPresets []int  `json:"study_presets" validate:"required,min=1"`
	StudyColor   string `json:"study_color" validate:"required"`
}

type CreatePomodoroSessionRequest struct {
	CourseID        *int   `json:"course_id" validate:"omitempty"`
	DurationMinutes int    `json:"duration_minutes" validate:"required,min=1"`
	Notes           string `json:"notes"`
}
