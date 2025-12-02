package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sebasr/avt-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresUserRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		Email:         "test@example.com",
		PasswordHash:  "hashed_password",
		EmailVerified: false,
		IsActive:      true,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.False(t, user.CreatedAt.IsZero())
	assert.False(t, user.UpdatedAt.IsZero())
}

func TestPostgresUserRepository_Create_DuplicateEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()

	user1 := &models.User{
		Email:        "duplicate@example.com",
		PasswordHash: "hash1",
		IsActive:     true,
	}

	err := repo.Create(ctx, user1)
	require.NoError(t, err)

	// Try to create another user with same email
	user2 := &models.User{
		Email:        "duplicate@example.com",
		PasswordHash: "hash2",
		IsActive:     true,
	}

	err = repo.Create(ctx, user2)
	assert.ErrorIs(t, err, ErrUserExists)
}

func TestPostgresUserRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a user
	user := &models.User{
		Email:        "getbyid@example.com",
		PasswordHash: "hashed_password",
		IsActive:     true,
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Get the user by ID
	retrieved, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Email, retrieved.Email)
	assert.Equal(t, user.PasswordHash, retrieved.PasswordHash)
	assert.Equal(t, user.IsActive, retrieved.IsActive)
}

func TestPostgresUserRepository_GetByID_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()

	nonExistentID := uuid.New()
	_, err := repo.GetByID(ctx, nonExistentID)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestPostgresUserRepository_GetByEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a user
	user := &models.User{
		Email:        "getbyemail@example.com",
		PasswordHash: "hashed_password",
		IsActive:     true,
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Get the user by email
	retrieved, err := repo.GetByEmail(ctx, user.Email)
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.PasswordHash, retrieved.PasswordHash)
}

func TestPostgresUserRepository_GetByEmail_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()

	_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestPostgresUserRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a user
	user := &models.User{
		Email:        "update@example.com",
		PasswordHash: "old_hash",
		IsActive:     true,
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Update the user
	user.PasswordHash = "new_hash"
	user.EmailVerified = true
	err = repo.Update(ctx, user)
	require.NoError(t, err)

	// Verify the update
	retrieved, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "new_hash", retrieved.PasswordHash)
	assert.True(t, retrieved.EmailVerified)
}

func TestPostgresUserRepository_UpdatePassword(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a user
	user := &models.User{
		Email:        "updatepwd@example.com",
		PasswordHash: "old_password_hash",
		IsActive:     true,
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Update password
	newHash := "new_password_hash"
	err = repo.UpdatePassword(ctx, user.ID, newHash)
	require.NoError(t, err)

	// Verify the update
	retrieved, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, newHash, retrieved.PasswordHash)
}

func TestPostgresUserRepository_UpdateEmailVerification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a user with verification token
	token := "verification_token"
	expiresAt := time.Now().Add(24 * time.Hour)
	user := &models.User{
		Email:                      "verify@example.com",
		PasswordHash:               "hash",
		EmailVerified:              false,
		VerificationToken:          &token,
		VerificationTokenExpiresAt: &expiresAt,
		IsActive:                   true,
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Update email verification
	err = repo.UpdateEmailVerification(ctx, user.ID, true)
	require.NoError(t, err)

	// Verify the update
	retrieved, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.True(t, retrieved.EmailVerified)
	assert.Nil(t, retrieved.VerificationToken)
	assert.Nil(t, retrieved.VerificationTokenExpiresAt)
}

func TestPostgresUserRepository_SetVerificationToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a user
	user := &models.User{
		Email:        "settoken@example.com",
		PasswordHash: "hash",
		IsActive:     true,
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Set verification token
	token := "new_verification_token"
	expiresAt := time.Now().Add(24 * time.Hour)
	err = repo.SetVerificationToken(ctx, user.ID, token, &expiresAt)
	require.NoError(t, err)

	// Verify the token was set
	retrieved, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved.VerificationToken)
	assert.Equal(t, token, *retrieved.VerificationToken)
	require.NotNil(t, retrieved.VerificationTokenExpiresAt)
	assert.WithinDuration(t, expiresAt, *retrieved.VerificationTokenExpiresAt, time.Second)
}

func TestPostgresUserRepository_SetResetToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a user
	user := &models.User{
		Email:        "resettoken@example.com",
		PasswordHash: "hash",
		IsActive:     true,
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Set reset token
	token := "reset_token"
	expiresAt := time.Now().Add(1 * time.Hour)
	err = repo.SetResetToken(ctx, user.ID, token, &expiresAt)
	require.NoError(t, err)

	// Verify the token was set
	retrieved, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved.ResetToken)
	assert.Equal(t, token, *retrieved.ResetToken)
	require.NotNil(t, retrieved.ResetTokenExpiresAt)
	assert.WithinDuration(t, expiresAt, *retrieved.ResetTokenExpiresAt, time.Second)
}

func TestPostgresUserRepository_UpdateLastLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a user
	user := &models.User{
		Email:        "lastlogin@example.com",
		PasswordHash: "hash",
		IsActive:     true,
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Initially, last login should be nil
	retrieved, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Nil(t, retrieved.LastLoginAt)

	// Update last login
	beforeUpdate := time.Now()
	err = repo.UpdateLastLogin(ctx, user.ID)
	require.NoError(t, err)

	// Verify last login was updated
	retrieved, err = repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved.LastLoginAt)
	assert.True(t, retrieved.LastLoginAt.After(beforeUpdate) || retrieved.LastLoginAt.Equal(beforeUpdate))
}
