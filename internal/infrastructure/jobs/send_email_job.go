package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/M1ralai/notly-api/internal/infrastructure/email"
)

// SendEmailJob sends a verification email asynchronously
type SendEmailJob struct {
	To           string
	Code         string
	EmailService email.EmailService
}

func (j *SendEmailJob) Name() string {
	return fmt.Sprintf("send_email:%s", j.To)
}

func (j *SendEmailJob) Execute(ctx context.Context) error {
	return j.EmailService.SendVerificationCode(j.To, j.Code)
}

func (j *SendEmailJob) Schedule() string {
	return "" // Manual-only job, not scheduled
}

func (j *SendEmailJob) Timeout() time.Duration {
	return 30 * time.Second // Email sending timeout
}

func (j *SendEmailJob) RetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries: 2,
		Delay:      2 * time.Second,
		Backoff:    3 * time.Second,
	}
}
