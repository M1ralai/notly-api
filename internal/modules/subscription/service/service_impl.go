package service

import (
	"context"
	"time"

	"github.com/M1ralai/notly-api/internal/modules/subscription/domain"
	"github.com/M1ralai/notly-api/internal/modules/subscription/repository"
)

type subscriptionService struct {
	entitlementRepo repository.EntitlementRepository
}

func NewSubscriptionService(entitlementRepo repository.EntitlementRepository) Service {
	return &subscriptionService{entitlementRepo: entitlementRepo}
}

func FreeStatus() *PremiumStatus {
	return &PremiumStatus{
		IsPremium:   false,
		PremiumPlan: "free",
		Status:      domain.StatusExpired,
	}
}

func (s *subscriptionService) GetPremiumStatus(ctx context.Context, userID int) (*PremiumStatus, error) {
	if s.entitlementRepo == nil {
		return FreeStatus(), nil
	}

	now := time.Now()
	entitlement, err := s.entitlementRepo.GetActiveByUserID(ctx, userID, now)
	if err != nil {
		return nil, err
	}
	if entitlement == nil || !entitlement.HasPremiumAccess(now) {
		return FreeStatus(), nil
	}

	return &PremiumStatus{
		IsPremium:        true,
		PremiumPlan:      entitlement.Plan,
		PremiumExpiresAt: entitlement.ExpiresAt,
		Provider:         entitlement.Provider,
		ProductID:        entitlement.ProductID,
		Status:           entitlement.Status,
	}, nil
}

func (s *subscriptionService) HasPremiumAccess(ctx context.Context, userID int) (bool, error) {
	status, err := s.GetPremiumStatus(ctx, userID)
	if err != nil {
		return false, err
	}
	return status.IsPremium, nil
}

func (s *subscriptionService) RequirePremium(ctx context.Context, userID int) error {
	hasAccess, err := s.HasPremiumAccess(ctx, userID)
	if err != nil {
		return err
	}
	if !hasAccess {
		return ErrPremiumRequired
	}
	return nil
}

func (s *subscriptionService) UpsertEntitlement(ctx context.Context, input UpsertEntitlementInput) (*PremiumStatus, error) {
	if input.Provider == "" {
		input.Provider = domain.ProviderAdmin
	}
	if input.ProductID == "" {
		input.ProductID = domain.ProductNotlyProMonthly
	}
	if input.Plan == "" {
		input.Plan = domain.PlanMonthly
	}
	if input.Status == "" {
		input.Status = domain.StatusActive
	}
	if input.Environment == "" {
		input.Environment = "production"
	}

	now := time.Now()
	entitlement := &domain.Entitlement{
		UserID:                input.UserID,
		Provider:              input.Provider,
		ProductID:             input.ProductID,
		Plan:                  input.Plan,
		Status:                input.Status,
		TransactionID:         input.TransactionID,
		OriginalTransactionID: input.OriginalTransactionID,
		PurchaseTokenHash:     input.PurchaseTokenHash,
		ExpiresAt:             input.ExpiresAt,
		Environment:           input.Environment,
		RawPayload:            input.RawPayload,
		LastVerifiedAt:        &now,
	}

	saved, err := s.entitlementRepo.Upsert(ctx, entitlement)
	if err != nil {
		return nil, err
	}
	if !saved.HasPremiumAccess(now) {
		return FreeStatus(), nil
	}
	return &PremiumStatus{
		IsPremium:        true,
		PremiumPlan:      saved.Plan,
		PremiumExpiresAt: saved.ExpiresAt,
		Provider:         saved.Provider,
		ProductID:        saved.ProductID,
		Status:           saved.Status,
	}, nil
}
