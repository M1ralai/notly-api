package service

import (
	"context"
	"errors"
	"time"

	"github.com/M1ralai/notly-api/internal/infrastructure/jobs"
	"github.com/M1ralai/notly-api/internal/modules/auth/dto"
)

func (s *authService) VerifyEmail(ctx context.Context, req *dto.VerifyEmailRequest) (*dto.AuthResponse, error) {
	s.logger.Info("Verifying email", map[string]interface{}{
		"email":  req.Email,
		"action": "VERIFY_EMAIL",
	})

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Check if already verified
	if user.IsVerified {
		return nil, errors.New("email already verified")
	}

	// Find verification code
	query := `SELECT id, expires_at FROM verification_codes
	          WHERE user_id = $1 AND code = $2 AND type = 'REGISTER'
	          ORDER BY created_at DESC LIMIT 1`

	var codeID int
	var expiresAt time.Time
	err = s.db.Conn.QueryRowContext(ctx, query, user.ID, req.Code).Scan(&codeID, &expiresAt)
	if err != nil {
		s.logger.Info("verification code not found or invalid", map[string]interface{}{
			"user_id": user.ID,
			"action":  "VERIFY_FAILED",
		})
		return nil, errors.New("invalid or expired verification code")
	}

	// Check if code is expired
	if time.Now().After(expiresAt) {
		return nil, errors.New("verification code has expired")
	}

	// Update user to verified
	updateQuery := `UPDATE users SET is_verified = true, updated_at = NOW() WHERE id = $1`
	_, err = s.db.Conn.ExecContext(ctx, updateQuery, user.ID)
	if err != nil {
		s.logger.Error("failed to verify user", err, map[string]interface{}{
			"user_id": user.ID,
			"action":  "VERIFY_FAILED",
		})
		return nil, err
	}

	// Delete used verification code
	deleteQuery := `DELETE FROM verification_codes WHERE id = $1`
	_, err = s.db.Conn.ExecContext(ctx, deleteQuery, codeID)
	if err != nil {
		s.logger.Error("failed to delete verification code", err, map[string]interface{}{
			"code_id": codeID,
		})
		// Non-critical error, continue
	}

	// Update user object
	user.IsVerified = true

	// Generate token for verified user
	token, expiresAtToken, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	// Generate refresh token for verified user
	refreshToken, err := s.generateRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("email verified successfully", map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"action":  "EMAIL_VERIFIED",
	})

	return &dto.AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAtToken,
		User:         toAuthUserResponse(user),
	}, nil
}

func (s *authService) ResendCode(ctx context.Context, req *dto.ResendCodeRequest) error {
	s.logger.Info("Resending verification code", map[string]interface{}{
		"email":  req.Email,
		"action": "RESEND_CODE",
	})

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Check if already verified
	if user.IsVerified {
		return errors.New("email already verified")
	}

	// Delete old codes for this user
	deleteQuery := `DELETE FROM verification_codes WHERE user_id = $1 AND type = 'REGISTER'`
	_, err = s.db.Conn.ExecContext(ctx, deleteQuery, user.ID)
	if err != nil {
		s.logger.Error("failed to delete old verification codes", err, map[string]interface{}{
			"user_id": user.ID,
		})
		// Non-critical, continue
	}

	// Generate new 6-digit verification code
	code := s.generateVerificationCode()
	expiresAt := time.Now().Add(15 * time.Minute)

	// Save new verification code to database
	query := `INSERT INTO verification_codes (user_id, code, type, expires_at) VALUES ($1, $2, $3, $4)`
	_, err = s.db.Conn.ExecContext(ctx, query, user.ID, code, "REGISTER", expiresAt)
	if err != nil {
		s.logger.Error("failed to create new verification code", err, map[string]interface{}{
			"user_id": user.ID,
			"action":  "RESEND_CODE_FAILED",
		})
		return err
	}

	// Send verification email asynchronously via worker pool
	err = s.workerPool.SubmitAsync(&jobs.SendEmailJob{
		To:           user.Email,
		Code:         code,
		EmailService: s.emailService,
	})
	if err != nil {
		s.logger.Error("failed to queue verification email", err, map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
		})
		// Non-critical - code is in DB, user can try again
	}

	s.logger.Info("verification code resent", map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"action":  "CODE_RESENT",
	})

	return nil
}
