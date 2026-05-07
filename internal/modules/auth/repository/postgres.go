package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/M1ralai/notly-api/internal/modules/auth/domain"
	"github.com/jmoiron/sqlx"
)

type postgresRepository struct {
	db *sqlx.DB
}

// NewPostgresRepository creates a new postgres-backed RefreshTokenRepository.
func NewPostgresRepository(db *sqlx.DB) RefreshTokenRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, rt *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(ctx, query, rt.UserID, rt.TokenHash, rt.ExpiresAt)
	return err
}

func (r *postgresRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	var model RefreshTokenModel
	err := r.db.GetContext(ctx, &model, query, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

func (r *postgresRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`
	_, err := r.db.ExecContext(ctx, query, tokenHash)
	return err
}

func (r *postgresRepository) DeleteAllByUserID(ctx context.Context, userID int) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
