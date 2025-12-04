package email

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// ConsoleService is an email service that logs emails to the console
// This is useful for local development and testing
type ConsoleService struct {
	fromAddress string
	fromName    string
	appURL      string
}

// NewConsoleService creates a new console-based email service
func NewConsoleService(fromAddress, fromName, appURL string) *ConsoleService {
	return &ConsoleService{
		fromAddress: fromAddress,
		fromName:    fromName,
		appURL:      appURL,
	}
}

// SendPasswordResetEmail logs the password reset email to the console
func (s *ConsoleService) SendPasswordResetEmail(_ context.Context, toEmail, resetToken string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", strings.TrimSuffix(s.appURL, "/"), resetToken)

	log.Println("========================================")
	log.Println("📧 PASSWORD RESET EMAIL (Console Mode)")
	log.Println("========================================")
	log.Printf("To: %s", toEmail)
	log.Printf("From: %s <%s>", s.fromName, s.fromAddress)
	log.Println("Subject: Password Reset Request")
	log.Println("----------------------------------------")
	log.Println("You requested a password reset for your account.")
	log.Println("")
	log.Printf("Reset URL: %s", resetURL)
	log.Printf("Reset Token: %s", resetToken)
	log.Println("")
	log.Println("This link will expire in 12 hours.")
	log.Println("========================================")

	return nil
}

// SendPasswordChangedEmail logs the password changed notification to the console
func (s *ConsoleService) SendPasswordChangedEmail(_ context.Context, toEmail string) error {
	log.Println("========================================")
	log.Println("📧 PASSWORD CHANGED EMAIL (Console Mode)")
	log.Println("========================================")
	log.Printf("To: %s", toEmail)
	log.Printf("From: %s <%s>", s.fromName, s.fromAddress)
	log.Println("Subject: Your Password Has Been Changed")
	log.Println("----------------------------------------")
	log.Println("Your password has been successfully changed.")
	log.Println("")
	log.Println("If you did not make this change, please contact support immediately.")
	log.Println("========================================")

	return nil
}
