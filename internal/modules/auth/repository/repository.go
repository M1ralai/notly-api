package repository

import (
	"context"

	"github.com/M1ralai/notly-api/internal/modules/auth/domain"
)

// RefreshTokenRepository defines the contract for refresh token persistence.
type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *domain.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	DeleteByTokenHash(ctx context.Context, tokenHash string) error
	DeleteAllByUserID(ctx context.Context, userID int) error
}
