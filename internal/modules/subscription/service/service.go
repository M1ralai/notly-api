package service

import (
	"context"
	"errors"
	"time"
)

const MonthlyPriceUSD = "3.99"

var ErrPremiumRequired = errors.New("premium required")

type PremiumStatus struct {
	IsPremium        bool       `json:"is_premium"`
	PremiumPlan      string     `json:"premium_plan"`
	PremiumExpiresAt *time.Time `json:"premium_expires_at,omitempty"`
	Provider         string     `json:"provider,omitempty"`
	ProductID        string     `json:"product_id,omitempty"`
	Status           string     `json:"status"`
}

type UpsertEntitlementInput struct {
	UserID                int
	Provider              string
	ProductID             string
	Plan                  string
	Status                string
	TransactionID         string
	OriginalTransactionID string
	PurchaseTokenHash     string
	ExpiresAt             *time.Time
	Environment           string
	RawPayload            []byte
}

type Service interface {
	GetPremiumStatus(ctx context.Context, userID int) (*PremiumStatus, error)
	HasPremiumAccess(ctx context.Context, userID int) (bool, error)
	RequirePremium(ctx context.Context, userID int) error
	UpsertEntitlement(ctx context.Context, input UpsertEntitlementInput) (*PremiumStatus, error)
}
