package repository

import (
	"context"
	"time"

	"github.com/M1ralai/notly-api/internal/modules/lifearea/domain"
)

type LifeAreaRepository interface {
	Create(ctx context.Context, lifeArea *domain.LifeArea) (*domain.LifeArea, error)
	GetByID(ctx context.Context, id int) (*domain.LifeArea, error)
	GetByUserID(ctx context.Context, userID int) ([]*domain.LifeArea, error)
	Update(ctx context.Context, lifeArea *domain.LifeArea) error
	Delete(ctx context.Context, id int) error
	GetUpdatedSince(ctx context.Context, userID int, since time.Time) ([]*domain.LifeArea, error)
	GetDeletedSince(ctx context.Context, userID int, since time.Time) ([]int, error)
}
