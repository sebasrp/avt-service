package email

import (
	"context"
	"sync"
)

// MockService is a mock email service implementation for testing.
// It stores sent emails in memory for verification in tests.
type MockService struct {
	mu                    sync.Mutex
	PasswordResetEmails   []MockEmail
	PasswordChangedEmails []MockEmail
}

// MockEmail represents an email that was sent by the mock service.
type MockEmail struct {
	To    string
	Token string // Only populated for password reset emails
}

// NewMockService creates a new mock email service.
func NewMockService() *MockService {
	return &MockService{
		PasswordResetEmails:   make([]MockEmail, 0),
		PasswordChangedEmails: make([]MockEmail, 0),
	}
}

// SendPasswordResetEmail records a password reset email.
func (s *MockService) SendPasswordResetEmail(_ context.Context, to, resetToken string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PasswordResetEmails = append(s.PasswordResetEmails, MockEmail{
		To:    to,
		Token: resetToken,
	})
	return nil
}

// SendPasswordChangedEmail records a password changed notification email.
func (s *MockService) SendPasswordChangedEmail(_ context.Context, to string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PasswordChangedEmails = append(s.PasswordChangedEmails, MockEmail{
		To: to,
	})
	return nil
}

// Reset clears all stored emails. Useful for test cleanup.
func (s *MockService) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PasswordResetEmails = make([]MockEmail, 0)
	s.PasswordChangedEmails = make([]MockEmail, 0)
}

// GetPasswordResetEmails returns a copy of all password reset emails sent.
func (s *MockService) GetPasswordResetEmails() []MockEmail {
	s.mu.Lock()
	defer s.mu.Unlock()
	emails := make([]MockEmail, len(s.PasswordResetEmails))
	copy(emails, s.PasswordResetEmails)
	return emails
}

// GetPasswordChangedEmails returns a copy of all password changed emails sent.
func (s *MockService) GetPasswordChangedEmails() []MockEmail {
	s.mu.Lock()
	defer s.mu.Unlock()
	emails := make([]MockEmail, len(s.PasswordChangedEmails))
	copy(emails, s.PasswordChangedEmails)
	return emails
}
