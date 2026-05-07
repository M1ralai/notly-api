package repository

import (
	"context"
	"time"

	"github.com/M1ralai/notly-api/internal/modules/semester/domain"
)

type SemesterRepository interface {
	Create(ctx context.Context, semester *domain.Semester) (*domain.Semester, error)
	GetByID(ctx context.Context, id int) (*domain.Semester, error)
	GetByUserID(ctx context.Context, userID int) ([]*domain.Semester, error)
	GetCurrent(ctx context.Context, userID int) (*domain.Semester, error)
	Update(ctx context.Context, semester *domain.Semester) error
	Delete(ctx context.Context, id int) error
	GetUpdatedSince(ctx context.Context, userID int, since time.Time) ([]*domain.Semester, error)
	GetDeletedSince(ctx context.Context, userID int, since time.Time) ([]int, error)
}
