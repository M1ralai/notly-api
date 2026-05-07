package email

import (
	"context"
	"fmt"
	"os"

	"github.com/resend/resend-go/v2"
)

type EmailService interface {
	SendVerificationCode(to, code string) error
}

type ResendEmailService struct {
	client *resend.Client
	from   string
}

func NewResendEmailService() *ResendEmailService {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		fmt.Println("⚠️  RESEND_API_KEY not set - emails will NOT be sent")
	}

	fromEmail := os.Getenv("EMAIL_FROM_ADDRESS")
	if fromEmail == "" {
		fromEmail = "onboarding@resend.dev" // Resend test email
	}

	client := resend.NewClient(apiKey)

	return &ResendEmailService{
		client: client,
		from:   fromEmail,
	}
}

func (s *ResendEmailService) SendVerificationCode(to, code string) error {
	if os.Getenv("RESEND_API_KEY") == "" {
		// Fallback to console logging if no API key
		fmt.Printf("\n===== EMAIL (Console Fallback) =====\n")
		fmt.Printf("To: %s\n", to)
		fmt.Printf("Subject: Verify your email\n")
		fmt.Printf("Your verification code: %s\n", code)
		fmt.Printf("====================================\n\n")
		return nil
	}

	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Subject: "Verify your email - SK-ED",
		Html:    s.generateVerificationEmailHTML(code),
		Text:    fmt.Sprintf("Your verification code is: %s\n\nThis code will expire in 15 minutes.", code),
	}

	_, err := s.client.Emails.SendWithContext(context.Background(), params)
	if err != nil {
		fmt.Printf("❌ Failed to send email via Resend: %v\n", err)
		// Still log to console as backup
		fmt.Printf("\n===== VERIFICATION CODE (Resend Failed) =====\n")
		fmt.Printf("Email: %s\n", to)
		fmt.Printf("Code: %s\n", code)
		fmt.Printf("=============================================\n\n")
		return err
	}

	fmt.Printf("✅ Verification email sent to %s\n", to)
	return nil
}

func (s *ResendEmailService) generateVerificationEmailHTML(code string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify Your Email</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #f5f5f5;">
    <table width="100%%" cellpadding="0" cellspacing="0" style="background-color: #f5f5f5; padding: 40px 0;">
        <tr>
            <td align="center">
                <table width="600" cellpadding="0" cellspacing="0" style="background-color: #ffffff; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
                    <tr>
                        <td style="padding: 40px 40px 20px 40px; text-align: center;">
                            <h1 style="margin: 0; color: #2563eb; font-size: 28px;">SK-ED</h1>
                        </td>
                    </tr>
                    <tr>
                        <td style="padding: 0 40px 40px 40px;">
                            <h2 style="margin: 0 0 20px 0; color: #1f2937; font-size: 24px; font-weight: 600;">Verify Your Email</h2>
                            <p style="margin: 0 0 30px 0; color: #6b7280; font-size: 16px; line-height: 1.5;">
                                Welcome! Please use the verification code below to complete your registration:
                            </p>
                            <div style="background-color: #f3f4f6; border-radius: 8px; padding: 30px; text-align: center; margin-bottom: 30px;">
                                <div style="font-size: 36px; font-weight: bold; letter-spacing: 8px; color: #2563eb; font-family: 'Courier New', monospace;">
                                    %s
                                </div>
                            </div>
                            <p style="margin: 0 0 20px 0; color: #6b7280; font-size: 14px; line-height: 1.5;">
                                This code will expire in <strong>15 minutes</strong>.
                            </p>
                            <p style="margin: 0; color: #9ca3af; font-size: 12px; line-height: 1.5;">
                                If you didn't request this email, you can safely ignore it.
                            </p>
                        </td>
                    </tr>
                    <tr>
                        <td style="padding: 20px 40px; background-color: #f9fafb; border-bottom-left-radius: 8px; border-bottom-right-radius: 8px;">
                            <p style="margin: 0; color: #9ca3af; font-size: 12px; text-align: center;">
                                © 2026 SK-ED. All rights reserved.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, code)
}
