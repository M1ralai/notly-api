package repository

import (
	"time"

	"github.com/M1ralai/notly-api/internal/modules/auth/domain"
)

// RefreshTokenModel is the database representation of a refresh token.
type RefreshTokenModel struct {
	ID        int       `db:"id"`
	UserID    int       `db:"user_id"`
	TokenHash string    `db:"token_hash"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

// ToDomain converts the database model to the domain entity.
func (m *RefreshTokenModel) ToDomain() *domain.RefreshToken {
	return &domain.RefreshToken{
		ID:        m.ID,
		UserID:    m.UserID,
		TokenHash: m.TokenHash,
		ExpiresAt: m.ExpiresAt,
		CreatedAt: m.CreatedAt,
	}
}

// FromDomain converts a domain entity to the database model.
func FromDomain(rt *domain.RefreshToken) *RefreshTokenModel {
	return &RefreshTokenModel{
		ID:        rt.ID,
		UserID:    rt.UserID,
		TokenHash: rt.TokenHash,
		ExpiresAt: rt.ExpiresAt,
		CreatedAt: rt.CreatedAt,
	}
}
