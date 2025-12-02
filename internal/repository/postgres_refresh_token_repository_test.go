package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sebasr/avt-service/internal/database"
	"github.com/sebasr/avt-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresRefreshTokenRepository_Create(t *testing.T) {
	db, cleanup := setupRefreshTokenTestDB(t)
	defer cleanup()

	repo := NewPostgresRefreshTokenRepository(db.DB)
	userRepo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a test user first
	user := &models.User{
		ID:           uuid.New(),
		Email:        "token@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create a refresh token
	token := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "test-hash-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UserAgent: "Mozilla/5.0",
		IPAddress: "192.168.1.1",
	}

	err = repo.Create(ctx, token)
	assert.NoError(t, err)

	// Verify token was created
	retrieved, err := repo.GetByHash(ctx, token.TokenHash)
	require.NoError(t, err)
	assert.Equal(t, token.ID, retrieved.ID)
	assert.Equal(t, token.UserID, retrieved.UserID)
	assert.Equal(t, token.TokenHash, retrieved.TokenHash)
	assert.Equal(t, token.UserAgent, retrieved.UserAgent)
	assert.Equal(t, token.IPAddress, retrieved.IPAddress)
}

func TestPostgresRefreshTokenRepository_GetByHash(t *testing.T) {
	db, cleanup := setupRefreshTokenTestDB(t)
	defer cleanup()

	repo := NewPostgresRefreshTokenRepository(db.DB)
	userRepo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "getbyhash@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create a refresh token
	token := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "unique-hash-456",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UserAgent: "Chrome/99.0",
		IPAddress: "10.0.0.1",
	}
	err = repo.Create(ctx, token)
	require.NoError(t, err)

	// Retrieve by hash
	retrieved, err := repo.GetByHash(ctx, token.TokenHash)
	require.NoError(t, err)
	assert.Equal(t, token.ID, retrieved.ID)
	assert.Equal(t, token.UserID, retrieved.UserID)
	assert.Nil(t, retrieved.RevokedAt)
	assert.Nil(t, retrieved.ReplacedBy)
}

func TestPostgresRefreshTokenRepository_GetByHash_NotFound(t *testing.T) {
	db, cleanup := setupRefreshTokenTestDB(t)
	defer cleanup()

	repo := NewPostgresRefreshTokenRepository(db.DB)
	ctx := context.Background()

	_, err := repo.GetByHash(ctx, "non-existent-hash")
	assert.ErrorIs(t, err, ErrRefreshTokenNotFound)
}

func TestPostgresRefreshTokenRepository_GetByHash_Expired(t *testing.T) {
	db, cleanup := setupRefreshTokenTestDB(t)
	defer cleanup()

	repo := NewPostgresRefreshTokenRepository(db.DB)
	userRepo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "expired@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create an expired token
	token := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "expired-hash",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		CreatedAt: time.Now().Add(-2 * time.Hour),
		UserAgent: "Safari/15.0",
		IPAddress: "172.16.0.1",
	}
	err = repo.Create(ctx, token)
	require.NoError(t, err)

	// Try to retrieve expired token
	_, err = repo.GetByHash(ctx, token.TokenHash)
	assert.ErrorIs(t, err, ErrRefreshTokenNotFound)
}

func TestPostgresRefreshTokenRepository_GetByHash_Revoked(t *testing.T) {
	db, cleanup := setupRefreshTokenTestDB(t)
	defer cleanup()

	repo := NewPostgresRefreshTokenRepository(db.DB)
	userRepo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "revoked@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create a token
	token := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "revoked-hash",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UserAgent: "Firefox/100.0",
		IPAddress: "192.168.0.1",
	}
	err = repo.Create(ctx, token)
	require.NoError(t, err)

	// Revoke the token
	err = repo.Revoke(ctx, token.ID)
	require.NoError(t, err)

	// Try to retrieve revoked token
	_, err = repo.GetByHash(ctx, token.TokenHash)
	assert.ErrorIs(t, err, ErrRefreshTokenRevoked)
}

func TestPostgresRefreshTokenRepository_Revoke(t *testing.T) {
	db, cleanup := setupRefreshTokenTestDB(t)
	defer cleanup()

	repo := NewPostgresRefreshTokenRepository(db.DB)
	userRepo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "revoke@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create a token
	token := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "to-revoke-hash",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UserAgent: "Edge/99.0",
		IPAddress: "10.1.1.1",
	}
	err = repo.Create(ctx, token)
	require.NoError(t, err)

	// Revoke the token
	err = repo.Revoke(ctx, token.ID)
	assert.NoError(t, err)

	// Verify token is revoked
	_, err = repo.GetByHash(ctx, token.TokenHash)
	assert.ErrorIs(t, err, ErrRefreshTokenRevoked)
}

func TestPostgresRefreshTokenRepository_Revoke_NotFound(t *testing.T) {
	db, cleanup := setupRefreshTokenTestDB(t)
	defer cleanup()

	repo := NewPostgresRefreshTokenRepository(db.DB)
	ctx := context.Background()

	// Try to revoke non-existent token
	err := repo.Revoke(ctx, uuid.New())
	assert.ErrorIs(t, err, ErrRefreshTokenNotFound)
}

func TestPostgresRefreshTokenRepository_RevokeByHash(t *testing.T) {
	db, cleanup := setupRefreshTokenTestDB(t)
	defer cleanup()

	repo := NewPostgresRefreshTokenRepository(db.DB)
	userRepo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "revokebyhash@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create a token
	token := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "revoke-by-hash-test",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UserAgent: "Opera/88.0",
		IPAddress: "192.168.100.1",
	}
	err = repo.Create(ctx, token)
	require.NoError(t, err)

	// Revoke by hash
	err = repo.RevokeByHash(ctx, token.TokenHash)
	assert.NoError(t, err)

	// Verify token is revoked
	_, err = repo.GetByHash(ctx, token.TokenHash)
	assert.ErrorIs(t, err, ErrRefreshTokenRevoked)
}

func TestPostgresRefreshTokenRepository_RevokeByHash_NotFound(t *testing.T) {
	db, cleanup := setupRefreshTokenTestDB(t)
	defer cleanup()

	repo := NewPostgresRefreshTokenRepository(db.DB)
	ctx := context.Background()

	// Try to revoke non-existent token
	err := repo.RevokeByHash(ctx, "non-existent-hash")
	assert.ErrorIs(t, err, ErrRefreshTokenNotFound)
}

func TestPostgresRefreshTokenRepository_RevokeAllForUser(t *testing.T) {
	db, cleanup := setupRefreshTokenTestDB(t)
	defer cleanup()

	repo := NewPostgresRefreshTokenRepository(db.DB)
	userRepo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "revokeall@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create multiple tokens for the user
	token1 := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "token-1-hash",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UserAgent: "Device 1",
		IPAddress: "10.0.0.1",
	}
	token2 := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "token-2-hash",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UserAgent: "Device 2",
		IPAddress: "10.0.0.2",
	}
	token3 := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "token-3-hash",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UserAgent: "Device 3",
		IPAddress: "10.0.0.3",
	}

	err = repo.Create(ctx, token1)
	require.NoError(t, err)
	err = repo.Create(ctx, token2)
	require.NoError(t, err)
	err = repo.Create(ctx, token3)
	require.NoError(t, err)

	// Revoke all tokens for user
	err = repo.RevokeAllForUser(ctx, user.ID)
	assert.NoError(t, err)

	// Verify all tokens are revoked
	_, err = repo.GetByHash(ctx, token1.TokenHash)
	assert.ErrorIs(t, err, ErrRefreshTokenRevoked)
	_, err = repo.GetByHash(ctx, token2.TokenHash)
	assert.ErrorIs(t, err, ErrRefreshTokenRevoked)
	_, err = repo.GetByHash(ctx, token3.TokenHash)
	assert.ErrorIs(t, err, ErrRefreshTokenRevoked)
}

func TestPostgresRefreshTokenRepository_DeleteExpired(t *testing.T) {
	db, cleanup := setupRefreshTokenTestDB(t)
	defer cleanup()

	repo := NewPostgresRefreshTokenRepository(db.DB)
	userRepo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "deleteexpired@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create some expired tokens
	expiredToken1 := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "expired-1",
		ExpiresAt: time.Now().Add(-2 * time.Hour),
		CreatedAt: time.Now().Add(-3 * time.Hour),
		UserAgent: "Old Device 1",
		IPAddress: "192.168.1.1",
	}
	expiredToken2 := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "expired-2",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		CreatedAt: time.Now().Add(-2 * time.Hour),
		UserAgent: "Old Device 2",
		IPAddress: "192.168.1.2",
	}

	// Create a valid token
	validToken := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "valid-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UserAgent: "Current Device",
		IPAddress: "192.168.1.100",
	}

	err = repo.Create(ctx, expiredToken1)
	require.NoError(t, err)
	err = repo.Create(ctx, expiredToken2)
	require.NoError(t, err)
	err = repo.Create(ctx, validToken)
	require.NoError(t, err)

	// Delete expired tokens
	count, err := repo.DeleteExpired(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Verify expired tokens are deleted
	_, err = repo.GetByHash(ctx, expiredToken1.TokenHash)
	assert.ErrorIs(t, err, ErrRefreshTokenNotFound)
	_, err = repo.GetByHash(ctx, expiredToken2.TokenHash)
	assert.ErrorIs(t, err, ErrRefreshTokenNotFound)

	// Verify valid token still exists
	retrieved, err := repo.GetByHash(ctx, validToken.TokenHash)
	assert.NoError(t, err)
	assert.Equal(t, validToken.ID, retrieved.ID)
}

// setupRefreshTokenTestDB creates a test database with the necessary tables
func setupRefreshTokenTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()
	return setupTestDB(t)
}
