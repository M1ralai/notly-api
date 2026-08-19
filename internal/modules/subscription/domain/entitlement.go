package domain

import "time"

const (
	ProviderApple      = "apple"
	ProviderGoogle     = "google"
	ProviderRevenueCat = "revenuecat"
	ProviderStripe     = "stripe"
	ProviderRCBilling  = "rc_billing"
	ProviderAdmin      = "admin"

	StatusActive       = "active"
	StatusExpired      = "expired"
	StatusCancelled    = "cancelled"
	StatusRevoked      = "revoked"
	StatusGracePeriod  = "grace_period"
	StatusBillingIssue = "billing_issue"
	StatusInTrial      = "in_trial"

	ProductNotlyProMonthly = "notly_pro_monthly"
	ProductNotlyProYearly  = "notly_pro_yearly"
	PlanMonthly            = "monthly"
	PlanYearly             = "yearly"
	PlanAnnual             = "annual"
	PlanLifetime           = "lifetime"
)

type Entitlement struct {
	ID                    int
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
	LastVerifiedAt        *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (e *Entitlement) HasPremiumAccess(now time.Time) bool {
	if e == nil {
		return false
	}
	if e.Status != StatusActive && e.Status != StatusGracePeriod {
		return false
	}
	return e.ExpiresAt == nil || e.ExpiresAt.After(now)
}
