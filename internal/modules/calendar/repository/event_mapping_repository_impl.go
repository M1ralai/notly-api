package repository

import (
	"context"
	"time"

	"github.com/M1ralai/notly-api/internal/modules/calendar/domain"
	"github.com/jmoiron/sqlx"
)

type eventMappingRepository struct {
	db *sqlx.DB
}

// NewEventMappingRepository creates a new event mapping repository
func NewEventMappingRepository(db *sqlx.DB) EventMappingRepository {
	return &eventMappingRepository{db: db}
}

func (r *eventMappingRepository) Create(ctx context.Context, event *domain.GoogleCalendarEvent) error {
	query := `
		INSERT INTO google_calendar_events (user_id, local_id, local_type, google_event_id, event_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	return r.db.QueryRowxContext(
		ctx, query,
		event.UserID, event.LocalID, event.LocalType, event.GoogleEventID, event.EventDate,
	).Scan(&event.ID, &event.CreatedAt)
}

func (r *eventMappingRepository) GetByLocalEvent(ctx context.Context, userID, localID int, localType string, date time.Time) (*domain.GoogleCalendarEvent, error) {
	query := `
		SELECT id, user_id, local_id, local_type, google_event_id, event_date, created_at
		FROM google_calendar_events
		WHERE user_id = $1 AND local_id = $2 AND local_type = $3 AND event_date = $4
	`
	var event domain.GoogleCalendarEvent
	err := r.db.GetContext(ctx, &event, query, userID, localID, localType, date)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	return &event, nil
}

func (r *eventMappingRepository) GetByLocalID(ctx context.Context, userID, localID int, localType string) (*domain.GoogleCalendarEvent, error) {
	query := `
		SELECT id, user_id, local_id, local_type, google_event_id, event_date, created_at
		FROM google_calendar_events
		WHERE user_id = $1 AND local_id = $2 AND local_type = $3
		LIMIT 1
	`
	var event domain.GoogleCalendarEvent
	err := r.db.GetContext(ctx, &event, query, userID, localID, localType)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	return &event, nil
}

func (r *eventMappingRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM google_calendar_events WHERE id = $1", id)
	return err
}

func (r *eventMappingRepository) DeleteByUser(ctx context.Context, userID int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM google_calendar_events WHERE user_id = $1", userID)
	return err
}
