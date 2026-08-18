package service

import (
	"context"
	"errors"
)

const MonthlyPriceUSD = "3.99"

var ErrPremiumRequired = errors.New("premium required")

type Service interface {
	HasPremiumAccess(ctx context.Context, userID int) (bool, error)
	RequirePremium(ctx context.Context, userID int) error
}
