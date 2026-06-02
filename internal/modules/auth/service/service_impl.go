package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	mathRand "math/rand"
	"os"
	"time"

	"github.com/M1ralai/notly-api/internal/infrastructure"
	"github.com/M1ralai/notly-api/internal/infrastructure/database"
	"github.com/M1ralai/notly-api/internal/infrastructure/email"
	"github.com/M1ralai/notly-api/internal/infrastructure/jobs"
	"github.com/M1ralai/notly-api/internal/infrastructure/logger"
	"github.com/M1ralai/notly-api/internal/modules/auth/domain"
	"github.com/M1ralai/notly-api/internal/modules/auth/dto"
	authRepo "github.com/M1ralai/notly-api/internal/modules/auth/repository"
	userDomain "github.com/M1ralai/notly-api/internal/modules/user/domain"
	userRepo "github.com/M1ralai/notly-api/internal/modules/user/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo           userRepo.UserRepository
	refreshTokenRepo   authRepo.RefreshTokenRepository
	logger             *logger.ZapLogger
	db                 *database.Database
	emailService       email.EmailService
	workerPool         *jobs.WorkerPool
	turnstileValidator *infrastructure.TurnstileValidator
}

func NewAuthService(
	userRepo userRepo.UserRepository,
	refreshTokenRepo authRepo.RefreshTokenRepository,
	logger *logger.ZapLogger,
	db *database.Database,
	emailService email.EmailService,
	workerPool *jobs.WorkerPool,
	turnstileValidator *infrastructure.TurnstileValidator,
) AuthService {
	return &authService{
		userRepo:           userRepo,
		refreshTokenRepo:   refreshTokenRepo,
		logger:             logger,
		db:                 db,
		emailService:       emailService,
		workerPool:         workerPool,
		turnstileValidator: turnstileValidator,
	}
}

type Claims struct {
	UserID   int    `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func (s *authService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	s.logger.Info("Login attempt", map[string]interface{}{
		"email":  req.Email,
		"action": "LOGIN",
	})

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		s.logger.Info("login failed - invalid password", map[string]interface{}{
			"user_id": user.ID,
			"email":   req.Email,
			"action":  "LOGIN_FAILED",
		})
		return nil, errors.New("invalid email or password")
	}

	// Generate access token
	token, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := s.generateRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("user logged in", map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"action":  "LOGIN",
	})

	return &dto.AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User: dto.UserResponse{
			ID:         user.ID,
			Email:      user.Email,
			FullName:   user.FullName,
			AvatarURL:  user.AvatarURL,
			Timezone:   user.Timezone,
			IsVerified: user.IsVerified,
			CreatedAt:  user.CreatedAt,
		},
	}, nil
}

func (s *authService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	s.logger.Info("Registering user", map[string]interface{}{
		"email":  req.Email,
		"action": "REGISTER",
	})

	if req.TurnstileToken != "" {
		if err := s.turnstileValidator.Verify(req.TurnstileToken); err != nil {
			s.logger.Error("turnstile verification failed", err, map[string]interface{}{
				"email":  req.Email,
				"action": "TURNSTILE_FAILED",
			})
			return nil, errors.New("bot verification failed")
		}
	} else {
		s.logger.Info("turnstile token not provided, skipping verification", map[string]interface{}{
			"email":  req.Email,
			"action": "TURNSTILE_SKIPPED",
		})
	}

	// Check if email already exists
	existing, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		// If already verified, reject registration
		if existing.IsVerified {
			return nil, errors.New("email already registered")
		}

		// If unverified, delete old codes and resend new verification code
		s.logger.Info("user already exists but unverified, resending verification code", map[string]interface{}{
			"user_id": existing.ID,
			"email":   existing.Email,
			"action":  "RESEND_ON_DUPLICATE_REGISTER",
		})

		// Delete old verification codes for this user
		deleteQuery := `DELETE FROM verification_codes WHERE user_id = $1 AND type = 'REGISTER'`
		_, err = s.db.Conn.ExecContext(ctx, deleteQuery, existing.ID)
		if err != nil {
			s.logger.Error("failed to delete old verification codes", err, map[string]interface{}{
				"user_id": existing.ID,
			})
			// Continue anyway - non-critical
		}

		// Generate new 6-digit verification code
		code := fmt.Sprintf("%06d", mathRand.Intn(1000000))
		expiresAt := time.Now().Add(15 * time.Minute)

		// Save new verification code to database
		insertQuery := `INSERT INTO verification_codes (user_id, code, type, expires_at) VALUES ($1, $2, $3, $4)`
		_, err = s.db.Conn.ExecContext(ctx, insertQuery, existing.ID, code, "REGISTER", expiresAt)
		if err != nil {
			s.logger.Error("failed to create new verification code", err, map[string]interface{}{
				"user_id": existing.ID,
			})
			return nil, err
		}

		// Send verification email asynchronously via worker pool
		err = s.workerPool.SubmitAsync(&jobs.SendEmailJob{
			To:           existing.Email,
			Code:         code,
			EmailService: s.emailService,
		})
		if err != nil {
			s.logger.Error("failed to queue verification email", err, map[string]interface{}{
				"user_id": existing.ID,
				"email":   existing.Email,
			})
			// Non-critical - user can resend
		}

		s.logger.Info("verification code resent to existing unverified user", map[string]interface{}{
			"user_id": existing.ID,
			"email":   existing.Email,
			"action":  "VERIFICATION_RESENT",
		})

		// Return response with empty token (user must verify first)
		return &dto.AuthResponse{
			Token: "",
			User: dto.UserResponse{
				ID:         existing.ID,
				Email:      existing.Email,
				FullName:   existing.FullName,
				Timezone:   existing.Timezone,
				IsVerified: existing.IsVerified,
				CreatedAt:  existing.CreatedAt,
			},
		}, nil
	}

	// No existing user - proceed with normal registration
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	timezone := req.Timezone
	if timezone == "" {
		timezone = "Europe/Istanbul"
	}

	now := time.Now()
	user := &userDomain.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
		Timezone:     timezone,
		IsVerified:   false, // New users start unverified
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	created, err := s.userRepo.Create(ctx, user)
	if err != nil {
		s.logger.Error("failed to create user", err, map[string]interface{}{
			"email":  req.Email,
			"action": "REGISTER_FAILED",
		})
		return nil, err
	}

	// Generate 6-digit verification code
	code := s.generateVerificationCode()
	expiresAt := time.Now().Add(15 * time.Minute)

	// Save verification code to database
	query := `INSERT INTO verification_codes (user_id, code, type, expires_at) VALUES ($1, $2, $3, $4)`
	_, err = s.db.Conn.ExecContext(ctx, query, created.ID, code, "REGISTER", expiresAt)
	if err != nil {
		s.logger.Error("failed to create verification code", err, map[string]interface{}{
			"user_id": created.ID,
			"action":  "VERIFICATION_CODE_FAILED",
		})
		return nil, err
	}

	// Send verification email asynchronously via worker pool
	err = s.workerPool.SubmitAsync(&jobs.SendEmailJob{
		To:           created.Email,
		Code:         code,
		EmailService: s.emailService,
	})
	if err != nil {
		s.logger.Error("failed to queue verification email", err, map[string]interface{}{
			"user_id": created.ID,
			"email":   created.Email,
		})
		// Non-critical - user can resend
	}

	s.logger.Info("user registered - verification required", map[string]interface{}{
		"user_id": created.ID,
		"email":   created.Email,
		"action":  "REGISTER",
	})

	// Return response WITHOUT token (user must verify first)
	return &dto.AuthResponse{
		Token:     "",
		ExpiresAt: time.Time{},
		User: dto.UserResponse{
			ID:         created.ID,
			Email:      created.Email,
			FullName:   created.FullName,
			AvatarURL:  created.AvatarURL,
			Timezone:   created.Timezone,
			IsVerified: created.IsVerified,
			CreatedAt:  created.CreatedAt,
		},
	}, nil
}

func (s *authService) generateVerificationCode() string {
	return fmt.Sprintf("%06d", mathRand.Intn(1000000))
}

func (s *authService) generateToken(user *userDomain.User) (string, time.Time, error) {
	jwtKey := []byte(os.Getenv("JWT_SECRET"))
	// Access token expires in 15 minutes
	expiresAt := time.Now().Add(15 * time.Minute)

	claims := &Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Email,
		Role:     "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-1 * time.Minute)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-1 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

func (s *authService) generateRefreshToken(ctx context.Context, userID int) (string, error) {
	// Generate 32 bytes of random data for the token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	refreshToken := hex.EncodeToString(b)

	// Hash the token before storing
	hash := sha256.Sum256([]byte(refreshToken))
	tokenHash := hex.EncodeToString(hash[:])

	// Refresh token expires in 7 days
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	rt := &domain.RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}

	if err := s.refreshTokenRepo.Create(ctx, rt); err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (s *authService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	// Hash the received refresh token to lookup in DB
	hash := sha256.Sum256([]byte(req.RefreshToken))
	tokenHash := hex.EncodeToString(hash[:])

	// Get refresh token from DB
	rt, err := s.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if rt == nil {
		return nil, errors.New("invalid refresh token")
	}

	// Check if expired
	if rt.IsExpired() {
		// Delete expired token
		_ = s.refreshTokenRepo.DeleteByTokenHash(ctx, tokenHash)
		return nil, errors.New("refresh token expired")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Rotate tokens: Delete used refresh token and generate a new one
	if err := s.refreshTokenRepo.DeleteByTokenHash(ctx, tokenHash); err != nil {
		s.logger.Error("failed to delete used refresh token", err, map[string]interface{}{
			"user_id": user.ID,
		})
		// Continue anyway to prevent blocking login
	}

	// Generate new access token
	newAccessToken, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	// Generate new refresh token
	newRefreshToken, err := s.generateRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("token refreshed", map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"action":  "REFRESH_TOKEN",
	})

	return &dto.AuthResponse{
		Token:        newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		User: dto.UserResponse{
			ID:         user.ID,
			Email:      user.Email,
			FullName:   user.FullName,
			AvatarURL:  user.AvatarURL,
			Timezone:   user.Timezone,
			IsVerified: user.IsVerified,
			CreatedAt:  user.CreatedAt,
		},
	}, nil
}

func (s *authService) Logout(ctx context.Context, req *dto.LogoutRequest) error {
	// Hash the received refresh token
	hash := sha256.Sum256([]byte(req.RefreshToken))
	tokenHash := hex.EncodeToString(hash[:])

	// Delete from DB
	if err := s.refreshTokenRepo.DeleteByTokenHash(ctx, tokenHash); err != nil {
		s.logger.Error("failed to delete refresh token on logout", err, nil)
		// Don't return error to user, just log it
	}

	s.logger.Info("user logged out", map[string]interface{}{
		"action": "LOGOUT",
	})

	return nil
}
