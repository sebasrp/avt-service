package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sebasr/avt-service/internal/models"
)

// MockUserRepository is a mock implementation of UserRepository for testing
type MockUserRepository struct {
	CreateFunc                  func(ctx context.Context, user *models.User) error
	GetByIDFunc                 func(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmailFunc              func(ctx context.Context, email string) (*models.User, error)
	UpdateFunc                  func(ctx context.Context, user *models.User) error
	UpdatePasswordFunc          func(ctx context.Context, id uuid.UUID, passwordHash string) error
	UpdateEmailVerificationFunc func(ctx context.Context, id uuid.UUID, verified bool) error
	SetVerificationTokenFunc    func(ctx context.Context, id uuid.UUID, token string, expiresAt *time.Time) error
	SetResetTokenFunc           func(ctx context.Context, id uuid.UUID, token string, expiresAt *time.Time) error
	UpdateLastLoginFunc         func(ctx context.Context, id uuid.UUID) error
}

// NewMockUserRepository creates a new mock user repository
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		CreateFunc: func(_ context.Context, _ *models.User) error {
			return nil
		},
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return nil, ErrUserNotFound
		},
		GetByEmailFunc: func(_ context.Context, _ string) (*models.User, error) {
			return nil, ErrUserNotFound
		},
		UpdateFunc: func(_ context.Context, _ *models.User) error {
			return nil
		},
		UpdatePasswordFunc: func(_ context.Context, _ uuid.UUID, _ string) error {
			return nil
		},
		UpdateEmailVerificationFunc: func(_ context.Context, _ uuid.UUID, _ bool) error {
			return nil
		},
		SetVerificationTokenFunc: func(_ context.Context, _ uuid.UUID, _ string, _ *time.Time) error {
			return nil
		},
		SetResetTokenFunc: func(_ context.Context, _ uuid.UUID, _ string, _ *time.Time) error {
			return nil
		},
		UpdateLastLoginFunc: func(_ context.Context, _ uuid.UUID) error {
			return nil
		},
	}
}

// Create implements UserRepository.Create
func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	return m.CreateFunc(ctx, user)
}

// GetByID implements UserRepository.GetByID
func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return m.GetByIDFunc(ctx, id)
}

// GetByEmail implements UserRepository.GetByEmail
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return m.GetByEmailFunc(ctx, email)
}

// Update implements UserRepository.Update
func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	return m.UpdateFunc(ctx, user)
}

// UpdatePassword implements UserRepository.UpdatePassword
func (m *MockUserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return m.UpdatePasswordFunc(ctx, id, passwordHash)
}

// UpdateEmailVerification implements UserRepository.UpdateEmailVerification
func (m *MockUserRepository) UpdateEmailVerification(ctx context.Context, id uuid.UUID, verified bool) error {
	return m.UpdateEmailVerificationFunc(ctx, id, verified)
}

// SetVerificationToken implements UserRepository.SetVerificationToken
func (m *MockUserRepository) SetVerificationToken(ctx context.Context, id uuid.UUID, token string, expiresAt *time.Time) error {
	return m.SetVerificationTokenFunc(ctx, id, token, expiresAt)
}

// SetResetToken implements UserRepository.SetResetToken
func (m *MockUserRepository) SetResetToken(ctx context.Context, id uuid.UUID, token string, expiresAt *time.Time) error {
	return m.SetResetTokenFunc(ctx, id, token, expiresAt)
}

// UpdateLastLogin implements UserRepository.UpdateLastLogin
func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	return m.UpdateLastLoginFunc(ctx, id)
}
