package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTService(t *testing.T) {
	secret := "test-secret-key"
	accessTTL := time.Hour
	refreshTTL := 24 * time.Hour

	service := NewJWTService(secret, accessTTL, refreshTTL)

	assert.NotNil(t, service)
	assert.Equal(t, []byte(secret), service.secret)
	assert.Equal(t, accessTTL, service.accessTokenTTL)
	assert.Equal(t, refreshTTL, service.refreshTokenTTL)
}

func TestGenerateAccessToken(t *testing.T) {
	service := NewJWTService("test-secret", time.Hour, 24*time.Hour)
	userID := uuid.New()
	email := "test@example.com"

	token, err := service.GenerateAccessToken(userID, email)

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the token
	claims, err := service.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, "avt-service", claims.RegisteredClaims.Issuer)
	assert.Equal(t, userID.String(), claims.RegisteredClaims.Subject)
}

func TestGenerateRefreshToken(t *testing.T) {
	service := NewJWTService("test-secret", time.Hour, 24*time.Hour)
	userID := uuid.New()
	email := "test@example.com"

	token, expiresAt, err := service.GenerateRefreshToken(userID, email)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.False(t, expiresAt.IsZero())

	// Verify expiration is approximately 24 hours from now
	expectedExpiry := time.Now().Add(24 * time.Hour)
	assert.WithinDuration(t, expectedExpiry, expiresAt, time.Second)

	// Validate the token
	claims, err := service.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), claims.UserID)
	assert.Equal(t, email, claims.Email)
}

func TestValidateToken_ValidToken(t *testing.T) {
	service := NewJWTService("test-secret", time.Hour, 24*time.Hour)
	userID := uuid.New()
	email := "test@example.com"

	token, err := service.GenerateAccessToken(userID, email)
	require.NoError(t, err)

	claims, err := service.ValidateToken(token)

	require.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID.String(), claims.UserID)
	assert.Equal(t, email, claims.Email)
}

func TestValidateToken_EmptyToken(t *testing.T) {
	service := NewJWTService("test-secret", time.Hour, 24*time.Hour)

	claims, err := service.ValidateToken("")

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	service := NewJWTService("test-secret", time.Hour, 24*time.Hour)

	claims, err := service.ValidateToken("invalid.token.string")

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	service := NewJWTService("test-secret", -time.Hour, 24*time.Hour) // Negative TTL = already expired
	userID := uuid.New()
	email := "test@example.com"

	token, err := service.GenerateAccessToken(userID, email)
	require.NoError(t, err)

	claims, err := service.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	service1 := NewJWTService("secret1", time.Hour, 24*time.Hour)
	service2 := NewJWTService("secret2", time.Hour, 24*time.Hour)

	userID := uuid.New()
	email := "test@example.com"

	// Generate token with service1
	token, err := service1.GenerateAccessToken(userID, email)
	require.NoError(t, err)

	// Try to validate with service2 (different secret)
	claims, err := service2.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestValidateToken_InvalidSigningMethod(t *testing.T) {
	service := NewJWTService("test-secret", time.Hour, 24*time.Hour)
	userID := uuid.New()

	// Create a token with a different signing method (RS256 instead of HS256)
	claims := &Claims{
		UserID: userID.String(),
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// This would fail in practice because RS256 requires a private key,
	// but we're testing the validation logic
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	result, err := service.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestValidateToken_InvalidUserIDFormat(t *testing.T) {
	service := NewJWTService("test-secret", time.Hour, 24*time.Hour)

	// Create a token with invalid UUID format
	claims := &Claims{
		UserID: "not-a-valid-uuid",
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "avt-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(service.secret)
	require.NoError(t, err)

	result, err := service.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidClaims)
}

func TestTokenUniqueness(t *testing.T) {
	service := NewJWTService("test-secret", time.Hour, 24*time.Hour)
	userID := uuid.New()
	email := "test@example.com"

	// Generate two access tokens for the same user
	token1, err1 := service.GenerateAccessToken(userID, email)
	require.NoError(t, err1)

	// Sleep for more than 1 second to ensure different timestamps (JWT uses second precision)
	time.Sleep(1100 * time.Millisecond)

	token2, err2 := service.GenerateAccessToken(userID, email)
	require.NoError(t, err2)

	// Tokens should be different due to different issued-at times
	assert.NotEqual(t, token1, token2)

	// Both tokens should be valid
	claims1, err := service.ValidateToken(token1)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), claims1.UserID)

	claims2, err := service.ValidateToken(token2)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), claims2.UserID)
}

func TestGetAccessTokenTTL(t *testing.T) {
	accessTTL := time.Hour
	service := NewJWTService("test-secret", accessTTL, 24*time.Hour)

	assert.Equal(t, accessTTL, service.GetAccessTokenTTL())
}

func TestGetRefreshTokenTTL(t *testing.T) {
	refreshTTL := 24 * time.Hour
	service := NewJWTService("test-secret", time.Hour, refreshTTL)

	assert.Equal(t, refreshTTL, service.GetRefreshTokenTTL())
}

func TestAccessAndRefreshTokenDifferentTTLs(t *testing.T) {
	service := NewJWTService("test-secret", time.Minute, time.Hour)
	userID := uuid.New()
	email := "test@example.com"

	// Generate both token types
	accessToken, err := service.GenerateAccessToken(userID, email)
	require.NoError(t, err)

	refreshToken, expiresAt, err := service.GenerateRefreshToken(userID, email)
	require.NoError(t, err)

	// Parse and verify different expiration times
	accessClaims, err := service.ValidateToken(accessToken)
	require.NoError(t, err)

	refreshClaims, err := service.ValidateToken(refreshToken)
	require.NoError(t, err)

	// Access token should expire sooner than refresh token
	assert.True(t, accessClaims.RegisteredClaims.ExpiresAt.Before(refreshClaims.RegisteredClaims.ExpiresAt.Time))

	// Refresh token expiry should match returned expiresAt
	assert.WithinDuration(t, expiresAt, refreshClaims.RegisteredClaims.ExpiresAt.Time, time.Second)
}

func BenchmarkGenerateAccessToken(b *testing.B) {
	service := NewJWTService("test-secret", time.Hour, 24*time.Hour)
	userID := uuid.New()
	email := "test@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GenerateAccessToken(userID, email)
	}
}

func BenchmarkGenerateRefreshToken(b *testing.B) {
	service := NewJWTService("test-secret", time.Hour, 24*time.Hour)
	userID := uuid.New()
	email := "test@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = service.GenerateRefreshToken(userID, email)
	}
}

func BenchmarkValidateToken(b *testing.B) {
	service := NewJWTService("test-secret", time.Hour, 24*time.Hour)
	userID := uuid.New()
	email := "test@example.com"
	token, _ := service.GenerateAccessToken(userID, email)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ValidateToken(token)
	}
}
