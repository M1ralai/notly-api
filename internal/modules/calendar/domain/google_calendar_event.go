package domain

import "time"

// GoogleCalendarEvent maps local task/habit completions to Google Calendar event IDs
type GoogleCalendarEvent struct {
	ID            int
	UserID        int
	LocalID       int
	LocalType     string // "task" or "habit"
	GoogleEventID string
	EventDate     time.Time
	CreatedAt     time.Time
}
