package repository

import (
	"context"
	"time"

	"github.com/M1ralai/notly-api/internal/modules/subscription/domain"
)

type EntitlementRepository interface {
	Upsert(ctx context.Context, entitlement *domain.Entitlement) (*domain.Entitlement, error)
	GetActiveByUserID(ctx context.Context, userID int, now time.Time) (*domain.Entitlement, error)
}
