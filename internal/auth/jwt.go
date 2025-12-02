package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	// ErrInvalidToken is returned when the token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when the token has expired
	ErrExpiredToken = errors.New("token has expired")
	// ErrInvalidClaims is returned when the token claims are invalid
	ErrInvalidClaims = errors.New("invalid token claims")
)

// Claims represents the JWT claims for authentication
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token generation and validation
type JWTService struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// NewJWTService creates a new JWT service instance
func NewJWTService(secret string, accessTTL, refreshTTL time.Duration) *JWTService {
	return &JWTService{
		secret:          []byte(secret),
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

// GenerateAccessToken generates a new access token for a user
func (s *JWTService) GenerateAccessToken(userID uuid.UUID, email string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID.String(),
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "avt-service",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a new refresh token for a user
// Returns the token string and its expiration time
func (s *JWTService) GenerateRefreshToken(userID uuid.UUID, email string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(s.refreshTokenTTL)

	claims := &Claims{
		UserID: userID.String(),
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(), // Unique JWT ID to prevent duplicate tokens
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "avt-service",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// ValidateToken validates a JWT token and returns its claims
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	// Extract claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	// Validate user ID is a valid UUID
	if _, err := uuid.Parse(claims.UserID); err != nil {
		return nil, fmt.Errorf("%w: invalid user ID format", ErrInvalidClaims)
	}

	return claims, nil
}

// GetAccessTokenTTL returns the access token TTL
func (s *JWTService) GetAccessTokenTTL() time.Duration {
	return s.accessTokenTTL
}

// GetRefreshTokenTTL returns the refresh token TTL
func (s *JWTService) GetRefreshTokenTTL() time.Duration {
	return s.refreshTokenTTL
}
