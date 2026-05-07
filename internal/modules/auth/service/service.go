package service

import (
	"context"

	"github.com/M1ralai/notly-api/internal/modules/auth/dto"
)

type AuthService interface {
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error)
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error)
	VerifyEmail(ctx context.Context, req *dto.VerifyEmailRequest) (*dto.AuthResponse, error)
	ResendCode(ctx context.Context, req *dto.ResendCodeRequest) error
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, error)
	Logout(ctx context.Context, req *dto.LogoutRequest) error
}
