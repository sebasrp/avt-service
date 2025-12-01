package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError error
	}{
		{
			name:        "valid password",
			password:    "securePassword123",
			expectError: nil,
		},
		{
			name:        "minimum length password",
			password:    "12345678",
			expectError: nil,
		},
		{
			name:        "empty password",
			password:    "",
			expectError: ErrPasswordEmpty,
		},
		{
			name:        "too short password",
			password:    "1234567",
			expectError: ErrPasswordTooShort,
		},
		{
			name:        "too long password",
			password:    strings.Repeat("a", 73),
			expectError: ErrPasswordTooLong,
		},
		{
			name:        "maximum length password",
			password:    strings.Repeat("a", 72),
			expectError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)

			if tt.expectError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectError, err)
				assert.Empty(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)
				// Verify the hash is a valid bcrypt hash (starts with $2a$, $2b$, or $2y$)
				assert.True(t, strings.HasPrefix(hash, "$2a$") ||
					strings.HasPrefix(hash, "$2b$") ||
					strings.HasPrefix(hash, "$2y$"))
				// Verify the hash is different from the plaintext password
				assert.NotEqual(t, tt.password, hash)
			}
		})
	}
}

func TestHashPasswordUniqueness(t *testing.T) {
	password := "testPassword123"

	// Hash the same password multiple times
	hash1, err1 := HashPassword(password)
	require.NoError(t, err1)

	hash2, err2 := HashPassword(password)
	require.NoError(t, err2)

	// Each hash should be different due to random salt
	assert.NotEqual(t, hash1, hash2, "bcrypt should generate different hashes for the same password")
}

func TestVerifyPassword(t *testing.T) {
	validPassword := "securePassword123"
	validHash, err := HashPassword(validPassword)
	require.NoError(t, err)

	tests := []struct {
		name     string
		password string
		hash     string
		expected bool
	}{
		{
			name:     "correct password",
			password: validPassword,
			hash:     validHash,
			expected: true,
		},
		{
			name:     "wrong password",
			password: "wrongPassword",
			hash:     validHash,
			expected: false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     validHash,
			expected: false,
		},
		{
			name:     "empty hash",
			password: validPassword,
			hash:     "",
			expected: false,
		},
		{
			name:     "both empty",
			password: "",
			hash:     "",
			expected: false,
		},
		{
			name:     "invalid hash format",
			password: validPassword,
			hash:     "not-a-valid-bcrypt-hash",
			expected: false,
		},
		{
			name:     "case sensitive password",
			password: "SECUREPASSWORD123",
			hash:     validHash,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyPassword(tt.password, tt.hash)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHashAndVerifyPassword(t *testing.T) {
	tests := []struct {
		password      string
		wrongPassword string
	}{
		{"password123", "password124"},
		{"P@ssw0rd!", "P@ssw0rd?"},
		{"12345678", "12345679"},
		{"VeryLongPasswordWithSpecialCharacters!@#$%^&*()", "VeryLongPasswordWithSpecialCharacters!@#$%^&*()X"},
		{strings.Repeat("a", 72), strings.Repeat("b", 72)}, // Maximum length - different char
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			// Hash the password
			hash, err := HashPassword(tt.password)
			require.NoError(t, err)
			require.NotEmpty(t, hash)

			// Verify correct password
			assert.True(t, VerifyPassword(tt.password, hash))

			// Verify incorrect password
			assert.False(t, VerifyPassword(tt.wrongPassword, hash))
		})
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := "benchmarkPassword123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = HashPassword(password)
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	password := "benchmarkPassword123"
	hash, _ := HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = VerifyPassword(password, hash)
	}
}
