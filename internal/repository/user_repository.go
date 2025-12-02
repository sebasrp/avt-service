package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sebasr/avt-service/internal/models"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *models.User) error

	// GetByID retrieves a user by their ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)

	// GetByEmail retrieves a user by their email address
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// Update updates an existing user's information
	Update(ctx context.Context, user *models.User) error

	// UpdatePassword updates a user's password hash
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error

	// UpdateEmailVerification updates email verification status and clears verification token
	UpdateEmailVerification(ctx context.Context, id uuid.UUID, verified bool) error

	// SetVerificationToken sets the email verification token and expiry
	SetVerificationToken(ctx context.Context, id uuid.UUID, token string, expiresAt *time.Time) error

	// SetResetToken sets the password reset token and expiry
	SetResetToken(ctx context.Context, id uuid.UUID, token string, expiresAt *time.Time) error

	// UpdateLastLogin updates the user's last login timestamp
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
}
