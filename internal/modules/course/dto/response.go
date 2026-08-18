package dto

import (
	"time"

	"github.com/M1ralai/notly-api/internal/modules/course/domain"
)

type CourseResponse struct {
	ID          int                  `json:"id"`
	UserID      int                  `json:"user_id"`
	Name        string               `json:"name"`
	Code        string               `json:"code,omitempty"`
	Instructor  string               `json:"instructor,omitempty"`
	Credits     float64              `json:"credits,omitempty"`
	SemesterID  *int                 `json:"semester_id,omitempty"`
	Semester    *SemesterInfo        `json:"semester,omitempty"`
	Type        string               `json:"type,omitempty"`
	Color       string               `json:"color,omitempty"`
	SyllabusURL string               `json:"syllabus_url,omitempty"`
	FinalGrade  string               `json:"final_grade,omitempty"`
	IsActive    bool                 `json:"is_active"`
	IsCompleted bool                 `json:"is_completed"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Components  []*ComponentResponse `json:"components,omitempty"`
	Schedules   []*ScheduleResponse  `json:"schedules,omitempty"`
	Resources   []*ResourceResponse  `json:"resources,omitempty"`
}

type SemesterInfo struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	IsCurrent bool   `json:"is_current"`
}

type ComponentResponse struct {
	ID             int        `json:"id"`
	CourseID       int        `json:"course_id"`
	Type           string     `json:"type"`
	Name           string     `json:"name"`
	Weight         float64    `json:"weight"`
	MaxScore       float64    `json:"max_score"`
	AchievedScore  *float64   `json:"achieved_score,omitempty"`
	DueDate        *time.Time `json:"due_date,omitempty"`
	CompletionDate *time.Time `json:"completion_date,omitempty"`
	IsCompleted    bool       `json:"is_completed"`
	Notes          string     `json:"notes,omitempty"`
	DisplayOrder   int        `json:"display_order"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ScheduleResponse struct {
	ID        int       `json:"id"`
	CourseID  int       `json:"course_id"`
	DayOfWeek string    `json:"day_of_week"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	Location  string    `json:"location,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ResourceResponse struct {
	ID            int       `json:"id"`
	CourseID      int       `json:"course_id"`
	ComponentID   *int      `json:"component_id,omitempty"`
	Title         string    `json:"title"`
	Type          string    `json:"type"`
	URL           string    `json:"url,omitempty"`
	FilePath      string    `json:"file_path,omitempty"`
	Description   string    `json:"description,omitempty"`
	Tags          []string  `json:"tags,omitempty"`
	IsPrimary     bool      `json:"is_primary"`
	FileSizeBytes int64     `json:"file_size_bytes,omitempty"`
	MimeType      string    `json:"mime_type,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func ToScheduleResponse(s *domain.Schedule) *ScheduleResponse {
	if s == nil {
		return nil
	}
	return &ScheduleResponse{
		ID:        s.ID,
		CourseID:  s.CourseID,
		DayOfWeek: s.DayOfWeek,
		StartTime: s.StartTime,
		EndTime:   s.EndTime,
		Location:  s.Location,
		CreatedAt: s.CreatedAt,
	}
}

func ToScheduleResponseList(schedules []*domain.Schedule) []*ScheduleResponse {
	if schedules == nil {
		return nil
	}
	result := make([]*ScheduleResponse, len(schedules))
	for i, s := range schedules {
		result[i] = ToScheduleResponse(s)
	}
	return result
}

func ToCourseResponse(c *domain.Course) *CourseResponse {
	return &CourseResponse{
		ID:          c.ID,
		UserID:      c.UserID,
		Name:        c.Name,
		Code:        c.Code,
		Instructor:  c.Instructor,
		Credits:     c.Credits,
		SemesterID:  c.SemesterID,
		Type:        c.Type,
		Color:       c.Color,
		SyllabusURL: c.SyllabusURL,
		FinalGrade:  c.FinalGrade,
		IsActive:    c.IsActive,
		IsCompleted: c.IsCompleted(),
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
		Components:  ToComponentResponseList(c.Components),
		Schedules:   ToScheduleResponseList(c.Schedules),
		Resources:   ToResourceResponseList(c.Resources),
	}
}

func ToComponentResponse(c *domain.Component) *ComponentResponse {
	if c == nil {
		return nil
	}
	return &ComponentResponse{
		ID:             c.ID,
		CourseID:       c.CourseID,
		Type:           c.Type,
		Name:           c.Name,
		Weight:         c.Weight,
		MaxScore:       c.MaxScore,
		AchievedScore:  c.AchievedScore,
		DueDate:        c.DueDate,
		CompletionDate: c.CompletionDate,
		IsCompleted:    c.IsCompleted,
		Notes:          c.Notes,
		DisplayOrder:   c.DisplayOrder,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

func ToComponentResponseList(components []*domain.Component) []*ComponentResponse {
	if components == nil {
		return nil
	}
	result := make([]*ComponentResponse, len(components))
	for i, c := range components {
		result[i] = ToComponentResponse(c)
	}
	return result
}

func ToResourceResponse(r *domain.Resource) *ResourceResponse {
	if r == nil {
		return nil
	}
	return &ResourceResponse{
		ID:            r.ID,
		CourseID:      r.CourseID,
		ComponentID:   r.ComponentID,
		Title:         r.Title,
		Type:          r.Type,
		URL:           r.URL,
		FilePath:      r.FilePath,
		Description:   r.Description,
		Tags:          r.Tags,
		IsPrimary:     r.IsPrimary,
		FileSizeBytes: r.FileSizeBytes,
		MimeType:      r.MimeType,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

func ToResourceResponseList(resources []*domain.Resource) []*ResourceResponse {
	if resources == nil {
		return nil
	}
	result := make([]*ResourceResponse, len(resources))
	for i, r := range resources {
		result[i] = ToResourceResponse(r)
	}
	return result
}

func ToCourseResponseList(courses []*domain.Course) []*CourseResponse {
	result := make([]*CourseResponse, len(courses))
	for i, c := range courses {
		result[i] = ToCourseResponse(c)
	}
	return result
}
