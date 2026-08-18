package repository

import (
	"time"

	"github.com/M1ralai/notly-api/internal/modules/subscription/domain"
)

type EntitlementModel struct {
	ID                    int        `db:"id"`
	UserID                int        `db:"user_id"`
	Provider              string     `db:"provider"`
	ProductID             string     `db:"product_id"`
	Plan                  string     `db:"plan"`
	Status                string     `db:"status"`
	TransactionID         *string    `db:"transaction_id"`
	OriginalTransactionID *string    `db:"original_transaction_id"`
	PurchaseTokenHash     *string    `db:"purchase_token_hash"`
	ExpiresAt             *time.Time `db:"expires_at"`
	Environment           string     `db:"environment"`
	RawPayload            []byte     `db:"raw_payload"`
	LastVerifiedAt        *time.Time `db:"last_verified_at"`
	CreatedAt             time.Time  `db:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at"`
}

func (m *EntitlementModel) ToDomain() *domain.Entitlement {
	return &domain.Entitlement{
		ID:                    m.ID,
		UserID:                m.UserID,
		Provider:              m.Provider,
		ProductID:             m.ProductID,
		Plan:                  m.Plan,
		Status:                m.Status,
		TransactionID:         valueOrEmpty(m.TransactionID),
		OriginalTransactionID: valueOrEmpty(m.OriginalTransactionID),
		PurchaseTokenHash:     valueOrEmpty(m.PurchaseTokenHash),
		ExpiresAt:             m.ExpiresAt,
		Environment:           m.Environment,
		RawPayload:            m.RawPayload,
		LastVerifiedAt:        m.LastVerifiedAt,
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
	}
}

func FromDomain(entitlement *domain.Entitlement) *EntitlementModel {
	return &EntitlementModel{
		ID:                    entitlement.ID,
		UserID:                entitlement.UserID,
		Provider:              entitlement.Provider,
		ProductID:             entitlement.ProductID,
		Plan:                  entitlement.Plan,
		Status:                entitlement.Status,
		TransactionID:         stringPtr(entitlement.TransactionID),
		OriginalTransactionID: stringPtr(entitlement.OriginalTransactionID),
		PurchaseTokenHash:     stringPtr(entitlement.PurchaseTokenHash),
		ExpiresAt:             entitlement.ExpiresAt,
		Environment:           entitlement.Environment,
		RawPayload:            entitlement.RawPayload,
		LastVerifiedAt:        entitlement.LastVerifiedAt,
		CreatedAt:             entitlement.CreatedAt,
		UpdatedAt:             entitlement.UpdatedAt,
	}
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
