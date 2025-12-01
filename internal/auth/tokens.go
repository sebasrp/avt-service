package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

const (
	// DefaultTokenLength is the default length for secure tokens (32 bytes = 256 bits)
	DefaultTokenLength = 32
)

var (
	// ErrNoToken is returned when no token is provided
	ErrNoToken = errors.New("no token provided")
	// ErrInvalidFormat is returned when the token format is invalid
	ErrInvalidFormat = errors.New("invalid token format")
	// ErrTokenGeneration is returned when token generation fails
	ErrTokenGeneration = errors.New("failed to generate token")
)

// HashToken creates a SHA256 hash of the token for secure storage
// The hash is returned as a hex-encoded string
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// GenerateSecureToken generates a cryptographically secure random token
// Returns a base64 URL-encoded string of random bytes
// Used for email verification tokens, password reset tokens, etc.
func GenerateSecureToken() (string, error) {
	return GenerateSecureTokenWithLength(DefaultTokenLength)
}

// GenerateSecureTokenWithLength generates a cryptographically secure random token
// with a specified byte length. The resulting base64 string will be longer.
func GenerateSecureTokenWithLength(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("%w: length must be positive", ErrTokenGeneration)
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("%w: %v", ErrTokenGeneration, err)
	}

	// Use URL-safe base64 encoding (no padding)
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// VerifyTokenHash compares a plain token with its hash
func VerifyTokenHash(token, hash string) bool {
	if token == "" || hash == "" {
		return false
	}
	computedHash := HashToken(token)
	return computedHash == hash
}
