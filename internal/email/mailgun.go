package email

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mailgun/mailgun-go/v5"
)

// MailgunService implements the Service interface using Mailgun's API.
type MailgunService struct {
	client      mailgun.Mailgun
	domain      string
	fromAddress string
	fromName    string
	appURL      string
}

// NewMailgunService creates a new Mailgun email service.
// domain: Mailgun domain (e.g., "mg.example.com")
// apiKey: Mailgun API key
// fromAddress: Sender email address (e.g., "noreply@example.com")
// fromName: Sender display name (e.g., "AVT Service")
// appURL: Frontend application URL for reset links (e.g., "https://app.example.com")
func NewMailgunService(domain, apiKey, fromAddress, fromName, appURL string) *MailgunService {
	// Trim whitespace from inputs (important when loaded from env files)
	domain = strings.TrimSpace(domain)
	apiKey = strings.TrimSpace(apiKey)
	fromAddress = strings.TrimSpace(fromAddress)
	fromName = strings.TrimSpace(fromName)
	appURL = strings.TrimSpace(appURL)

	// Mailgun v5 NewMailgun takes the API key as the parameter
	mg := mailgun.NewMailgun(apiKey)

	// Check if EU region should be used (set via MAILGUN_EU=true)
	if os.Getenv("MAILGUN_EU") == "true" {
		// Note: Mailgun v5 doesn't want the /v3 suffix - it adds it automatically
		_ = mg.SetAPIBase("https://api.eu.mailgun.net")
	}
	return &MailgunService{
		client:      mg,
		domain:      domain,
		fromAddress: fromAddress,
		fromName:    fromName,
		appURL:      appURL,
	}
}

// SendPasswordResetEmail sends a password reset link to the user.
func (s *MailgunService) SendPasswordResetEmail(ctx context.Context, to, resetToken string) error {
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.appURL, resetToken)

	subject := "Reset Your Password"
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f8f9fa; border-radius: 5px; padding: 30px; margin-bottom: 20px;">
        <h2 style="color: #2c3e50; margin-top: 0;">Password Reset Request</h2>
        <p>You requested to reset your password. Click the button below to proceed:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background-color: #007bff; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block; font-weight: bold;">Reset Password</a>
        </div>
        <p style="color: #666; font-size: 14px;">Or copy and paste this link into your browser:</p>
        <p style="word-break: break-all; background-color: #fff; padding: 10px; border-radius: 3px; font-size: 12px; border: 1px solid #ddd;">%s</p>
        <p style="color: #666; font-size: 14px; margin-top: 30px;">This link will expire in 12 hours.</p>
        <p style="color: #666; font-size: 14px;">If you didn't request this, you can safely ignore this email.</p>
    </div>
    <p style="color: #999; font-size: 12px; text-align: center;">This is an automated message, please do not reply.</p>
</body>
</html>`, resetLink, resetLink)

	textBody := fmt.Sprintf(`Password Reset Request

You requested to reset your password. Visit the link below to proceed:

%s

This link will expire in 12 hours.

If you didn't request this, you can safely ignore this email.

---
This is an automated message, please do not reply.`, resetLink)

	sender := fmt.Sprintf("%s <%s>", s.fromName, s.fromAddress)
	message := mailgun.NewMessage(s.domain, sender, subject, textBody, to)
	message.SetHTML(htmlBody)

	// Set timeout for the request
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := s.client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	return nil
}

// SendPasswordChangedEmail sends a notification that the password was changed.
func (s *MailgunService) SendPasswordChangedEmail(ctx context.Context, to string) error {
	subject := "Your Password Has Been Changed"
	htmlBody := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f8f9fa; border-radius: 5px; padding: 30px; margin-bottom: 20px;">
        <h2 style="color: #2c3e50; margin-top: 0;">Password Changed</h2>
        <p>Your password has been successfully changed.</p>
        <div style="background-color: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; margin: 20px 0;">
            <p style="margin: 0; color: #856404;"><strong>Security Alert:</strong> If you didn't make this change, please contact support immediately.</p>
        </div>
        <p style="color: #666; font-size: 14px;">For your security, all active sessions have been logged out. You'll need to log in again with your new password.</p>
    </div>
    <p style="color: #999; font-size: 12px; text-align: center;">This is an automated message, please do not reply.</p>
</body>
</html>`

	textBody := `Password Changed

Your password has been successfully changed.

SECURITY ALERT: If you didn't make this change, please contact support immediately.

For your security, all active sessions have been logged out. You'll need to log in again with your new password.

---
This is an automated message, please do not reply.`

	sender := fmt.Sprintf("%s <%s>", s.fromName, s.fromAddress)
	message := mailgun.NewMessage(s.domain, sender, subject, textBody, to)
	message.SetHTML(htmlBody)

	// Set timeout for the request
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := s.client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send password changed email: %w", err)
	}

	return nil
}
