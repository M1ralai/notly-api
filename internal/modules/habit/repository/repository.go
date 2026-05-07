package repository

import (
	"context"
	"time"

	"github.com/M1ralai/notly-api/internal/modules/habit/domain"
)

type HabitRepository interface {
	Create(ctx context.Context, habit *domain.Habit) (*domain.Habit, error)
	GetByID(ctx context.Context, id int) (*domain.Habit, error)
	GetByUserID(ctx context.Context, userID int) ([]*domain.Habit, error)
	GetActiveHabits(ctx context.Context, userID int) ([]*domain.Habit, error)
	Update(ctx context.Context, habit *domain.Habit) error
	Delete(ctx context.Context, id int) error

	GetUpdatedSince(ctx context.Context, userID int, since time.Time) ([]*domain.Habit, error)
	GetDeletedSince(ctx context.Context, userID int, since time.Time) ([]int, error)

	LogHabit(ctx context.Context, habitID int, logDate time.Time, count int, notes string) error
	SkipHabit(ctx context.Context, habitID int, logDate time.Time, notes string) error
	GetLogsForDate(ctx context.Context, habitID int, date time.Time) (*HabitLogModel, error)
	GetLogsByDateRange(ctx context.Context, habitID int, start, end time.Time) ([]*HabitLogModel, error)
	HasLogForToday(ctx context.Context, habitID int) (bool, error)
	HasSkippedToday(ctx context.Context, habitID int) (bool, error)
	DeleteTodayLog(ctx context.Context, habitID int, userID int) error
}
