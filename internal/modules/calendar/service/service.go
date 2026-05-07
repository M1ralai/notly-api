package service

import (
	"context"
	"time"

	"github.com/M1ralai/notly-api/internal/modules/calendar/dto"
)

type CalendarService interface {
	// OAuth Flow
	GetGoogleAuthURL(ctx context.Context, userID int) (string, error)
	HandleGoogleCallback(ctx context.Context, userID int, code string) (*dto.IntegrationResponse, error)
	DisconnectGoogle(ctx context.Context, userID int) error

	// Sync Operations
	SyncGoogle(ctx context.Context, userID int) error
	GetSyncStatus(ctx context.Context, userID int) (*dto.SyncStatusResponse, error)

	// Integration Management
	GetIntegrations(ctx context.Context, userID int) ([]*dto.IntegrationResponse, error)

	// Task/Habit Done/Undone sync
	MarkDone(ctx context.Context, userID int, localID int, entityType string, title string, date time.Time) error
	MarkUndone(ctx context.Context, userID int, localID int, entityType string, date time.Time) error

	// Timed Events (Tasks)
	CreateAllDayEvent(ctx context.Context, userID int, localID int, entityType string, title, description string, date time.Time) error
	UpdateAllDayEvent(ctx context.Context, userID int, localID int, entityType string, title, description string, date time.Time) error
	CreateTimedEvent(ctx context.Context, userID int, localID int, entityType string, title, description string, startTime, endTime time.Time, recurrence []string, notificationEnabled bool, notificationMethod string, notificationMinutes int) error
	UpdateTimedEvent(ctx context.Context, userID int, localID int, entityType string, title, description string, startTime, endTime time.Time, recurrence []string, notificationEnabled bool, notificationMethod string, notificationMinutes int) error

	// Delete Operations
	DeleteEvent(ctx context.Context, userID int, localID int, entityType string) error

	// Queue Operations
	QueueSync(ctx context.Context, userID int, eventID int, action string) error
	ProcessSyncQueue(ctx context.Context, limit int) (int, error)
}
