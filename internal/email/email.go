// Package email provides email service functionality for the AVT service.
package email

import "context"

// Service defines the interface for sending emails.
// Implementations include Mailgun for production and Mock for testing.
type Service interface {
	// SendPasswordResetEmail sends a password reset link to the user.
	// The resetToken is included in the email as part of the reset link.
	// Returns an error if the email fails to send.
	SendPasswordResetEmail(ctx context.Context, to, resetToken string) error

	// SendPasswordChangedEmail notifies the user that their password was changed.
	// This is a security notification to alert users of potential unauthorized access.
	// Returns an error if the email fails to send.
	SendPasswordChangedEmail(ctx context.Context, to string) error
}
