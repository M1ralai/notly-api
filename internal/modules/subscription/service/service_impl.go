package service

import (
	"context"

	userRepo "github.com/M1ralai/notly-api/internal/modules/user/repository"
)

type subscriptionService struct {
	userRepo userRepo.UserRepository
}

func NewSubscriptionService(userRepo userRepo.UserRepository) Service {
	return &subscriptionService{userRepo: userRepo}
}

func (s *subscriptionService) HasPremiumAccess(ctx context.Context, userID int) (bool, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return false, err
	}
	return user.HasPremiumAccess(), nil
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
