package repository

import (
	"context"
	"time"

	"github.com/M1ralai/notly-api/internal/modules/calendar/domain"
)

// EventMappingRepository manages Google Calendar event mappings
type EventMappingRepository interface {
	Create(ctx context.Context, event *domain.GoogleCalendarEvent) error
	GetByLocalEvent(ctx context.Context, userID, localID int, localType string, date time.Time) (*domain.GoogleCalendarEvent, error)
	GetByLocalID(ctx context.Context, userID, localID int, localType string) (*domain.GoogleCalendarEvent, error)
	Delete(ctx context.Context, id int) error
	DeleteByUser(ctx context.Context, userID int) error
}
