package repository

import (
	"context"
	"time"

	"github.com/M1ralai/notly-api/internal/modules/goal/domain"
)

type GoalRepository interface {
	Create(ctx context.Context, goal *domain.Goal) (*domain.Goal, error)
	GetByID(ctx context.Context, id int) (*domain.Goal, error)
	GetByUserID(ctx context.Context, userID int) ([]*domain.Goal, error)
	Update(ctx context.Context, goal *domain.Goal) error
	Delete(ctx context.Context, id int) error
	CountMilestones(ctx context.Context, goalID int) (total int, completed int, err error)
	GetUpdatedSince(ctx context.Context, userID int, since time.Time) ([]*domain.Goal, error)
	GetDeletedSince(ctx context.Context, userID int, since time.Time) ([]int, error)
}
