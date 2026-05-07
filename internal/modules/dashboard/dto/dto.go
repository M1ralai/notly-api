package dto

import (
	"time"
)

type DashboardResponse struct {
	Tasks     []*TaskResponse     `json:"tasks"`
	Habits    []*HabitResponse    `json:"habits"`
	LifeAreas []*LifeAreaResponse `json:"life_areas"`
	Stats     *TaskStatsResponse  `json:"stats"`
}

// TaskResponse copied from task/dto to decouple modules
type TaskResponse struct {
	ID                   int        `json:"id"`
	UserID               int        `json:"user_id"`
	ParentTaskID         *int       `json:"parent_task_id,omitempty"`
	Title                string     `json:"title"`
	Description          string     `json:"description,omitempty"`
	DueDate              *time.Time `json:"due_date,omitempty"`
	EstimatedStart       *time.Time `json:"estimated_start,omitempty"`
	EstimatedEnd         *time.Time `json:"estimated_end,omitempty"`
	ActualStart          *time.Time `json:"actual_start,omitempty"`
	ActualEnd            *time.Time `json:"actual_end,omitempty"`
	Priority             string     `json:"priority"`
	IsCompleted          bool       `json:"is_completed"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
	ProgressPercentage   float64    `json:"progress_percentage"`
	CompletedSubtasks    int        `json:"completed_subtasks"`
	TotalSubtasks        int        `json:"total_subtasks"`
	EstimatedDurationMin int        `json:"estimated_duration_min,omitempty"`
	ActualDurationMin    int        `json:"actual_duration_min,omitempty"`
	IsOverdue            bool       `json:"is_overdue"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type TaskStatsResponse struct {
	CompletedToday int `json:"completed_today"`
	DueToday       int `json:"due_today"`
	DueTomorrow    int `json:"due_tomorrow"`
	Overdue        int `json:"overdue"`
}

// HabitResponse copied from habit/dto to decouple modules
type HabitResponse struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	LifeAreaID     *int      `json:"life_area_id,omitempty"`
	Title          string    `json:"title"`
	Icon           string    `json:"icon"`
	Description    string    `json:"description,omitempty"`
	Frequency      string    `json:"frequency"`
	FrequencyDays  []string  `json:"frequencyDays,omitempty"`
	IntervalDays   int       `json:"intervalDays,omitempty"`
	TargetCount    int       `json:"target_count"`
	CurrentStreak  int       `json:"current_streak"`
	BestStreak     int       `json:"bestStreak"`
	LongestStreak  int       `json:"longest_streak"`
	IsActive       bool      `json:"is_active"`
	CompletedToday bool      `json:"completed_today"`
	SkippedToday   bool      `json:"skipped_today"`
	TimeOfDay      string    `json:"timeOfDay,omitempty"`
	ReminderTime   string    `json:"reminderTime,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// LifeAreaResponse copied from lifearea/dto to decouple modules
type LifeAreaResponse struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	Name         string    `json:"name"`
	Icon         string    `json:"icon,omitempty"`
	Color        string    `json:"color,omitempty"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
}

