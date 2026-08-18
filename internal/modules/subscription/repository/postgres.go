package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/M1ralai/notly-api/internal/modules/subscription/domain"
	"github.com/jmoiron/sqlx"
)

type postgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) EntitlementRepository {
	return &postgresRepository{db: db}
}

const entitlementSelectColumns = `
	id, user_id, provider, product_id, plan, status, transaction_id,
	original_transaction_id, purchase_token_hash, expires_at, environment,
	raw_payload, last_verified_at, created_at, updated_at`

func (r *postgresRepository) Upsert(ctx context.Context, entitlement *domain.Entitlement) (*domain.Entitlement, error) {
	query := `
		INSERT INTO subscription_entitlements (
			user_id, provider, product_id, plan, status, transaction_id,
			original_transaction_id, purchase_token_hash, expires_at, environment,
			raw_payload, last_verified_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
		ON CONFLICT (user_id, provider, product_id)
		DO UPDATE SET
			plan = EXCLUDED.plan,
			status = EXCLUDED.status,
			transaction_id = EXCLUDED.transaction_id,
			original_transaction_id = COALESCE(EXCLUDED.original_transaction_id, subscription_entitlements.original_transaction_id),
			purchase_token_hash = COALESCE(EXCLUDED.purchase_token_hash, subscription_entitlements.purchase_token_hash),
			expires_at = EXCLUDED.expires_at,
			environment = EXCLUDED.environment,
			raw_payload = EXCLUDED.raw_payload,
			last_verified_at = EXCLUDED.last_verified_at,
			updated_at = NOW()
		RETURNING ` + entitlementSelectColumns

	model := FromDomain(entitlement)
	var saved EntitlementModel
	err := r.db.GetContext(
		ctx,
		&saved,
		query,
		model.UserID,
		model.Provider,
		model.ProductID,
		model.Plan,
		model.Status,
		model.TransactionID,
		model.OriginalTransactionID,
		model.PurchaseTokenHash,
		model.ExpiresAt,
		model.Environment,
		model.RawPayload,
		model.LastVerifiedAt,
	)
	if err != nil {
		return nil, err
	}
	return saved.ToDomain(), nil
}

func (r *postgresRepository) GetActiveByUserID(ctx context.Context, userID int, now time.Time) (*domain.Entitlement, error) {
	query := `
		SELECT ` + entitlementSelectColumns + `
		FROM subscription_entitlements
		WHERE user_id = $1
		  AND status IN ('active', 'grace_period')
		  AND (expires_at IS NULL OR expires_at > $2)
		ORDER BY expires_at DESC NULLS FIRST, updated_at DESC
		LIMIT 1
	`

	var model EntitlementModel
	err := r.db.GetContext(ctx, &model, query, userID, now)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return model.ToDomain(), nil
}
