package email

import (
	"context"
	"testing"
)

func TestMockService_SendPasswordResetEmail(t *testing.T) {
	service := NewMockService()
	ctx := context.Background()

	// Send first email
	err := service.SendPasswordResetEmail(ctx, "user1@example.com", "token123")
	if err != nil {
		t.Fatalf("SendPasswordResetEmail() error = %v", err)
	}

	// Send second email
	err = service.SendPasswordResetEmail(ctx, "user2@example.com", "token456")
	if err != nil {
		t.Fatalf("SendPasswordResetEmail() error = %v", err)
	}

	// Verify emails were stored
	emails := service.GetPasswordResetEmails()
	if len(emails) != 2 {
		t.Errorf("GetPasswordResetEmails() count = %d, want 2", len(emails))
	}

	// Verify first email
	if emails[0].To != "user1@example.com" {
		t.Errorf("Email[0].To = %q, want %q", emails[0].To, "user1@example.com")
	}
	if emails[0].Token != "token123" {
		t.Errorf("Email[0].Token = %q, want %q", emails[0].Token, "token123")
	}

	// Verify second email
	if emails[1].To != "user2@example.com" {
		t.Errorf("Email[1].To = %q, want %q", emails[1].To, "user2@example.com")
	}
	if emails[1].Token != "token456" {
		t.Errorf("Email[1].Token = %q, want %q", emails[1].Token, "token456")
	}
}

func TestMockService_SendPasswordChangedEmail(t *testing.T) {
	service := NewMockService()
	ctx := context.Background()

	// Send first email
	err := service.SendPasswordChangedEmail(ctx, "user1@example.com")
	if err != nil {
		t.Fatalf("SendPasswordChangedEmail() error = %v", err)
	}

	// Send second email
	err = service.SendPasswordChangedEmail(ctx, "user2@example.com")
	if err != nil {
		t.Fatalf("SendPasswordChangedEmail() error = %v", err)
	}

	// Verify emails were stored
	emails := service.GetPasswordChangedEmails()
	if len(emails) != 2 {
		t.Errorf("GetPasswordChangedEmails() count = %d, want 2", len(emails))
	}

	// Verify first email
	if emails[0].To != "user1@example.com" {
		t.Errorf("Email[0].To = %q, want %q", emails[0].To, "user1@example.com")
	}

	// Verify second email
	if emails[1].To != "user2@example.com" {
		t.Errorf("Email[1].To = %q, want %q", emails[1].To, "user2@example.com")
	}
}

func TestMockService_Reset(t *testing.T) {
	service := NewMockService()
	ctx := context.Background()

	// Send some emails
	_ = service.SendPasswordResetEmail(ctx, "user@example.com", "token123")
	_ = service.SendPasswordChangedEmail(ctx, "user@example.com")

	// Verify emails exist
	if len(service.GetPasswordResetEmails()) != 1 {
		t.Error("Expected 1 password reset email before reset")
	}
	if len(service.GetPasswordChangedEmails()) != 1 {
		t.Error("Expected 1 password changed email before reset")
	}

	// Reset
	service.Reset()

	// Verify emails were cleared
	if len(service.GetPasswordResetEmails()) != 0 {
		t.Error("Expected 0 password reset emails after reset")
	}
	if len(service.GetPasswordChangedEmails()) != 0 {
		t.Error("Expected 0 password changed emails after reset")
	}
}

func TestMockService_ConcurrentAccess(t *testing.T) {
	service := NewMockService()
	ctx := context.Background()

	// Test concurrent writes
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(_ int) {
			_ = service.SendPasswordResetEmail(ctx, "user@example.com", "token")
			_ = service.SendPasswordChangedEmail(ctx, "user@example.com")
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all emails were recorded
	resetEmails := service.GetPasswordResetEmails()
	changedEmails := service.GetPasswordChangedEmails()

	if len(resetEmails) != numGoroutines {
		t.Errorf("GetPasswordResetEmails() count = %d, want %d", len(resetEmails), numGoroutines)
	}
	if len(changedEmails) != numGoroutines {
		t.Errorf("GetPasswordChangedEmails() count = %d, want %d", len(changedEmails), numGoroutines)
	}
}

func TestMockService_GettersReturnCopy(t *testing.T) {
	service := NewMockService()
	ctx := context.Background()

	// Send an email
	_ = service.SendPasswordResetEmail(ctx, "user@example.com", "token123")

	// Get emails
	emails1 := service.GetPasswordResetEmails()
	emails2 := service.GetPasswordResetEmails()

	// Modify first slice
	if len(emails1) > 0 {
		emails1[0].To = "modified@example.com"
	}

	// Verify second slice wasn't affected (proving it's a copy)
	if len(emails2) > 0 && emails2[0].To != "user@example.com" {
		t.Error("Getter should return a copy, not a reference to internal slice")
	}

	// Verify internal state wasn't modified
	emails3 := service.GetPasswordResetEmails()
	if len(emails3) > 0 && emails3[0].To != "user@example.com" {
		t.Error("Modifying returned slice should not affect internal state")
	}
}

func TestMockService_ImplementsInterface(_ *testing.T) {
	// This test verifies that MockService implements the Service interface
	var _ Service = (*MockService)(nil)
}
