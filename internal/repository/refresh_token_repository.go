package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sebasr/avt-service/internal/models"
)

// RefreshTokenRepository defines the interface for refresh token data access
type RefreshTokenRepository interface {
	// Create stores a new refresh token
	Create(ctx context.Context, token *models.RefreshToken) error

	// GetByHash retrieves a refresh token by its hash
	GetByHash(ctx context.Context, hash string) (*models.RefreshToken, error)

	// Revoke marks a refresh token as revoked by its ID
	Revoke(ctx context.Context, id uuid.UUID) error

	// RevokeByHash marks a refresh token as revoked by its hash
	RevokeByHash(ctx context.Context, hash string) error

	// RevokeAllForUser revokes all active refresh tokens for a specific user
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error

	// DeleteExpired removes all expired tokens and returns the count
	DeleteExpired(ctx context.Context) (int64, error)
}
