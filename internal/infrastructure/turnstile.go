package infrastructure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type TurnstileValidator struct {
	secretKey string
	client    *http.Client
}

type TurnstileVerifyRequest struct {
	Secret   string `json:"secret"`
	Response string `json:"response"`
}

type TurnstileVerifyResponse struct {
	Success     bool     `json:"success"`
	ChallengeTs string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
}

func NewTurnstileValidator() *TurnstileValidator {
	return &TurnstileValidator{
		secretKey: os.Getenv("TURNSTILE_SECRET_KEY"),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (v *TurnstileValidator) Verify(token string) error {
	if v.secretKey == "" {
		// If no secret key, skip validation (development mode)
		fmt.Println("⚠️  TURNSTILE_SECRET_KEY not set - skipping validation")
		return nil
	}

	reqBody := TurnstileVerifyRequest{
		Secret:   v.secretKey,
		Response: token,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := v.client.Post(
		"https://challenges.cloudflare.com/turnstile/v0/siteverify",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("failed to verify turnstile token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var verifyResp TurnstileVerifyResponse
	if err := json.Unmarshal(body, &verifyResp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !verifyResp.Success {
		return fmt.Errorf("turnstile verification failed: %v", verifyResp.ErrorCodes)
	}

	fmt.Printf("✅ Turnstile verification successful for hostname: %s\n", verifyResp.Hostname)
	return nil
}
