package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sebasr/avt-service/internal/models"
)

// MockRefreshTokenRepository is a mock implementation of RefreshTokenRepository for testing
type MockRefreshTokenRepository struct {
	CreateFunc           func(ctx context.Context, token *models.RefreshToken) error
	GetByHashFunc        func(ctx context.Context, hash string) (*models.RefreshToken, error)
	RevokeFunc           func(ctx context.Context, id uuid.UUID) error
	RevokeByHashFunc     func(ctx context.Context, hash string) error
	RevokeAllForUserFunc func(ctx context.Context, userID uuid.UUID) error
	DeleteExpiredFunc    func(ctx context.Context) (int64, error)
}

// NewMockRefreshTokenRepository creates a new mock refresh token repository
func NewMockRefreshTokenRepository() *MockRefreshTokenRepository {
	return &MockRefreshTokenRepository{
		CreateFunc: func(_ context.Context, _ *models.RefreshToken) error {
			return nil
		},
		GetByHashFunc: func(_ context.Context, _ string) (*models.RefreshToken, error) {
			return nil, ErrRefreshTokenNotFound
		},
		RevokeFunc: func(_ context.Context, _ uuid.UUID) error {
			return nil
		},
		RevokeByHashFunc: func(_ context.Context, _ string) error {
			return nil
		},
		RevokeAllForUserFunc: func(_ context.Context, _ uuid.UUID) error {
			return nil
		},
		DeleteExpiredFunc: func(_ context.Context) (int64, error) {
			return 0, nil
		},
	}
}

// Create implements RefreshTokenRepository.Create
func (m *MockRefreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) error {
	return m.CreateFunc(ctx, token)
}

// GetByHash implements RefreshTokenRepository.GetByHash
func (m *MockRefreshTokenRepository) GetByHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	return m.GetByHashFunc(ctx, hash)
}

// Revoke implements RefreshTokenRepository.Revoke
func (m *MockRefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	return m.RevokeFunc(ctx, id)
}

// RevokeByHash implements RefreshTokenRepository.RevokeByHash
func (m *MockRefreshTokenRepository) RevokeByHash(ctx context.Context, hash string) error {
	return m.RevokeByHashFunc(ctx, hash)
}

// RevokeAllForUser implements RefreshTokenRepository.RevokeAllForUser
func (m *MockRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	return m.RevokeAllForUserFunc(ctx, userID)
}

// DeleteExpired implements RefreshTokenRepository.DeleteExpired
func (m *MockRefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	return m.DeleteExpiredFunc(ctx)
}
