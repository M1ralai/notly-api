package domain

import "time"

type Course struct {
	ID          int
	UserID      int
	Name        string
	Code        string
	Instructor  string
	Credits     float64
	SemesterID  *int
	Type        string
	Color       string
	SyllabusURL string
	FinalGrade  string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Components  []*Component
	Schedules   []*Schedule
	Resources   []*Resource
}

type Component struct {
	ID             int
	CourseID       int
	Type           string
	Name           string
	Weight         float64
	MaxScore       float64
	AchievedScore  *float64
	DueDate        *time.Time
	CompletionDate *time.Time
	IsCompleted    bool
	Notes          string
	DisplayOrder   int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Schedule struct {
	ID                   int
	CourseID             int
	DayOfWeek            string
	StartTime            string // HH:MM format
	EndTime              string // HH:MM format
	Location             string
	NotificationsEnabled bool
	NotificationType     string
	ReminderTime         int
	CreatedAt            time.Time
}

type Resource struct {
	ID            int
	CourseID      int
	ComponentID   *int
	Title         string
	Type          string
	URL           string
	FilePath      string
	Description   string
	Tags          []string
	IsPrimary     bool
	FileSizeBytes int64
	MimeType      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (c *Course) IsCompleted() bool {
	return c.FinalGrade != ""
}
