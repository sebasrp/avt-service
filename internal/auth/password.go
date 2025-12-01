package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost is the default bcrypt cost parameter
	// Higher cost = more secure but slower
	// 10 is a good balance for production use
	DefaultCost = bcrypt.DefaultCost
)

var (
	// ErrPasswordTooShort is returned when password is too short
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	// ErrPasswordTooLong is returned when password is too long
	ErrPasswordTooLong = errors.New("password must be at most 72 characters")
	// ErrPasswordEmpty is returned when password is empty
	ErrPasswordEmpty = errors.New("password cannot be empty")
)

// HashPassword hashes a plaintext password using bcrypt
// Returns the hashed password or an error if hashing fails
func HashPassword(password string) (string, error) {
	// Validate password length
	if len(password) == 0 {
		return "", ErrPasswordEmpty
	}
	if len(password) < 8 {
		return "", ErrPasswordTooShort
	}
	if len(password) > 72 {
		// bcrypt has a maximum password length of 72 bytes
		return "", ErrPasswordTooLong
	}

	// Generate bcrypt hash
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

// VerifyPassword compares a plaintext password with a hashed password
// Returns true if the password matches, false otherwise
func VerifyPassword(password, hash string) bool {
	// Validate inputs
	if password == "" || hash == "" {
		return false
	}

	// Compare password with hash
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
