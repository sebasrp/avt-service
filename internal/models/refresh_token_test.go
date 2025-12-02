package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRefreshToken_IsValid(t *testing.T) {
	now := time.Now()
	userID := uuid.New()

	tests := []struct {
		name     string
		token    *RefreshToken
		expected bool
	}{
		{
			name: "valid token - not expired, not revoked",
			token: &RefreshToken{
				ID:        uuid.New(),
				UserID:    userID,
				TokenHash: "hash123",
				ExpiresAt: now.Add(24 * time.Hour),
				CreatedAt: now,
				RevokedAt: nil,
			},
			expected: true,
		},
		{
			name: "invalid token - expired",
			token: &RefreshToken{
				ID:        uuid.New(),
				UserID:    userID,
				TokenHash: "hash123",
				ExpiresAt: now.Add(-1 * time.Hour),
				CreatedAt: now.Add(-25 * time.Hour),
				RevokedAt: nil,
			},
			expected: false,
		},
		{
			name: "invalid token - revoked",
			token: &RefreshToken{
				ID:        uuid.New(),
				UserID:    userID,
				TokenHash: "hash123",
				ExpiresAt: now.Add(24 * time.Hour),
				CreatedAt: now,
				RevokedAt: &now,
			},
			expected: false,
		},
		{
			name: "invalid token - both expired and revoked",
			token: &RefreshToken{
				ID:        uuid.New(),
				UserID:    userID,
				TokenHash: "hash123",
				ExpiresAt: now.Add(-1 * time.Hour),
				CreatedAt: now.Add(-25 * time.Hour),
				RevokedAt: &now,
			},
			expected: false,
		},
		{
			name: "valid token - expires in 1 second",
			token: &RefreshToken{
				ID:        uuid.New(),
				UserID:    userID,
				TokenHash: "hash123",
				ExpiresAt: now.Add(1 * time.Second),
				CreatedAt: now,
				RevokedAt: nil,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.token.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRefreshToken_IsExpired(t *testing.T) {
	now := time.Now()
	userID := uuid.New()

	tests := []struct {
		name     string
		token    *RefreshToken
		expected bool
	}{
		{
			name: "not expired - expires in future",
			token: &RefreshToken{
				ID:        uuid.New(),
				UserID:    userID,
				ExpiresAt: now.Add(24 * time.Hour),
			},
			expected: false,
		},
		{
			name: "expired - expired 1 hour ago",
			token: &RefreshToken{
				ID:        uuid.New(),
				UserID:    userID,
				ExpiresAt: now.Add(-1 * time.Hour),
			},
			expected: true,
		},
		{
			name: "expired - expired just now",
			token: &RefreshToken{
				ID:        uuid.New(),
				UserID:    userID,
				ExpiresAt: now.Add(-1 * time.Millisecond),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.token.IsExpired()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRefreshToken_IsRevoked(t *testing.T) {
	now := time.Now()
	userID := uuid.New()

	tests := []struct {
		name     string
		token    *RefreshToken
		expected bool
	}{
		{
			name: "not revoked",
			token: &RefreshToken{
				ID:        uuid.New(),
				UserID:    userID,
				RevokedAt: nil,
			},
			expected: false,
		},
		{
			name: "revoked",
			token: &RefreshToken{
				ID:        uuid.New(),
				UserID:    userID,
				RevokedAt: &now,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.token.IsRevoked()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRefreshToken_ToResponse(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	tokenID := uuid.New()
	replacedByID := uuid.New()
	revokedAt := now.Add(-1 * time.Hour)

	token := &RefreshToken{
		ID:         tokenID,
		UserID:     userID,
		TokenHash:  "hash123",
		ExpiresAt:  now.Add(24 * time.Hour),
		CreatedAt:  now,
		RevokedAt:  &revokedAt,
		ReplacedBy: &replacedByID,
		UserAgent:  "Mozilla/5.0",
		IPAddress:  "192.168.1.1",
	}

	response := token.ToResponse()

	assert.Equal(t, tokenID, response.ID)
	assert.Equal(t, userID, response.UserID)
	assert.Equal(t, token.ExpiresAt, response.ExpiresAt)
	assert.Equal(t, token.CreatedAt, response.CreatedAt)
	assert.Equal(t, &revokedAt, response.RevokedAt)
	assert.Equal(t, &replacedByID, response.ReplacedBy)
	assert.Equal(t, "Mozilla/5.0", response.UserAgent)
	assert.Equal(t, "192.168.1.1", response.IPAddress)
	assert.False(t, response.IsValid) // Should be false because it's revoked
}

func TestRefreshToken_ToResponse_ValidToken(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	tokenID := uuid.New()

	token := &RefreshToken{
		ID:        tokenID,
		UserID:    userID,
		TokenHash: "hash123",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
		RevokedAt: nil,
		UserAgent: "Mozilla/5.0",
		IPAddress: "192.168.1.1",
	}

	response := token.ToResponse()

	assert.Equal(t, tokenID, response.ID)
	assert.Equal(t, userID, response.UserID)
	assert.True(t, response.IsValid) // Should be true - not revoked and not expired
	assert.Nil(t, response.RevokedAt)
	assert.Nil(t, response.ReplacedBy)
}
