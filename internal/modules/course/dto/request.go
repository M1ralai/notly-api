package dto

type CreateCourseRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Code        string  `json:"code,omitempty" validate:"omitempty,max=50"`
	Instructor  string  `json:"instructor,omitempty" validate:"omitempty,max=255"`
	Credits     float64 `json:"credits,omitempty"`
	SemesterID  int     `json:"semester_id" validate:"required"`
	Type        string  `json:"type,omitempty" validate:"omitempty,max=50"`
	Color       string  `json:"color,omitempty" validate:"omitempty,max=7"`
	SyllabusURL string  `json:"syllabus_url,omitempty" validate:"omitempty,url,max=500"`
}

type UpdateCourseRequest struct {
	Name        *string  `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Code        *string  `json:"code,omitempty" validate:"omitempty,max=50"`
	Instructor  *string  `json:"instructor,omitempty" validate:"omitempty,max=255"`
	Credits     *float64 `json:"credits,omitempty"`
	SemesterID  *int     `json:"semester_id,omitempty"`
	Type        *string  `json:"type,omitempty" validate:"omitempty,max=50"`
	Color       *string  `json:"color,omitempty" validate:"omitempty,max=7"`
	SyllabusURL *string  `json:"syllabus_url,omitempty" validate:"omitempty,url,max=500"`
	FinalGrade  *string  `json:"final_grade,omitempty" validate:"omitempty,max=10"`
	IsActive    *bool    `json:"is_active,omitempty"`
}

type CreateComponentRequest struct {
	CourseID      int      `json:"course_id" validate:"required"`
	Type          string   `json:"type" validate:"required,max=50"`
	Name          string   `json:"name" validate:"required,max=255"`
	Weight        *float64 `json:"weight,omitempty"`
	MaxScore      *float64 `json:"max_score,omitempty"`
	AchievedScore *float64 `json:"achieved_score,omitempty"`
	DueDate       *string  `json:"due_date,omitempty"`
	Notes         string   `json:"notes,omitempty"`
	DisplayOrder  int      `json:"display_order,omitempty"`
}

type UpdateComponentRequest struct {
	Type           *string  `json:"type,omitempty" validate:"omitempty,max=50"`
	Name           *string  `json:"name,omitempty" validate:"omitempty,max=255"`
	Weight         *float64 `json:"weight,omitempty"`
	MaxScore       *float64 `json:"max_score,omitempty"`
	AchievedScore  *float64 `json:"achieved_score,omitempty"`
	DueDate        *string  `json:"due_date,omitempty"`
	CompletionDate *string  `json:"completion_date,omitempty"`
	IsCompleted    *bool    `json:"is_completed,omitempty"`
	Notes          *string  `json:"notes,omitempty"`
	DisplayOrder   *int     `json:"display_order,omitempty"`
}

type CreateScheduleRequest struct {
	CourseID             int    `json:"course_id" validate:"required"`
	DayOfWeek            string `json:"day_of_week" validate:"required,oneof=Monday Tuesday Wednesday Thursday Friday Saturday Sunday"`
	StartTime            string `json:"start_time" validate:"required"`
	EndTime              string `json:"end_time" validate:"required"`
	Location             string `json:"location,omitempty" validate:"omitempty,max=255"`
	SyncToCalendar       bool   `json:"sync_to_calendar"`
	NotificationsEnabled bool   `json:"notifications_enabled"`
	NotificationType     string `json:"notification_type"`
	ReminderTime         int    `json:"reminder_time"`
}

type UpdateScheduleRequest struct {
	DayOfWeek            *string `json:"day_of_week,omitempty" validate:"omitempty,oneof=Monday Tuesday Wednesday Thursday Friday Saturday Sunday"`
	StartTime            *string `json:"start_time,omitempty"`
	EndTime              *string `json:"end_time,omitempty"`
	Location             *string `json:"location,omitempty" validate:"omitempty,max=255"`
	SyncToCalendar       *bool   `json:"sync_to_calendar"`
	NotificationsEnabled *bool   `json:"notifications_enabled"`
	NotificationType     *string `json:"notification_type"`
	ReminderTime         *int    `json:"reminder_time"`
}

type CreateResourceRequest struct {
	CourseID    int      `json:"course_id" validate:"required"`
	ComponentID *int     `json:"component_id,omitempty"`
	Title       string   `json:"title" validate:"required,max=255"`
	Type        string   `json:"type" validate:"required,oneof=file link video text"`
	URL         string   `json:"url,omitempty" validate:"omitempty,url,max=500"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	IsPrimary   bool     `json:"is_primary"`
}

type UpdateResourceRequest struct {
	ComponentID *int     `json:"component_id,omitempty"`
	Title       *string  `json:"title,omitempty" validate:"omitempty,max=255"`
	Type        *string  `json:"type,omitempty" validate:"omitempty,oneof=file link video text"`
	URL         *string  `json:"url,omitempty" validate:"omitempty,url,max=500"`
	Description *string  `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	IsPrimary   *bool    `json:"is_primary,omitempty"`
}
