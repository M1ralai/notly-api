package domain

import "time"

type Semester struct {
	ID        int
	UserID    int
	Name      string
	StartDate time.Time
	EndDate   time.Time
	IsCurrent bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsActive checks if the current date is within the semester dates
func (s *Semester) IsActive() bool {
	now := time.Now()
	return (now.Equal(s.StartDate) || now.After(s.StartDate)) &&
		(now.Equal(s.EndDate) || now.Before(s.EndDate))
}
