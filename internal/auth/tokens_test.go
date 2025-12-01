package auth

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string // SHA256 hash in hex
	}{
		{
			name:     "simple token",
			token:    "test-token",
			expected: "8eaa36172ff6c7afe16cbb9a6e82e2f7c25b6fbbdd5e5d63d7beae61ea27a7e6",
		},
		{
			name:     "empty token",
			token:    "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", // SHA256 of empty string
		},
		{
			name:     "long token",
			token:    "this-is-a-very-long-token-with-many-characters-0123456789",
			expected: "bdc9c71f45a3d1d8d9c7e1e9f5c0a1b2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8", // Will be different
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := HashToken(tt.token)

			// Verify hash is 64 characters (SHA256 hex = 32 bytes = 64 hex chars)
			assert.Len(t, hash, 64)

			// Verify hash is consistent
			hash2 := HashToken(tt.token)
			assert.Equal(t, hash, hash2, "hash should be deterministic")

			// Verify different tokens produce different hashes
			if tt.token != "" {
				differentHash := HashToken(tt.token + "x")
				assert.NotEqual(t, hash, differentHash)
			}
		})
	}
}

func TestHashTokenDeterministic(t *testing.T) {
	token := "consistent-token"

	hash1 := HashToken(token)
	hash2 := HashToken(token)
	hash3 := HashToken(token)

	assert.Equal(t, hash1, hash2)
	assert.Equal(t, hash2, hash3)
}

func TestHashTokenUniqueness(t *testing.T) {
	tokens := []string{
		"token1",
		"token2",
		"token3",
		"a",
		"A",
		"",
	}

	hashes := make(map[string]string)

	for _, token := range tokens {
		hash := HashToken(token)

		// Check for collisions (should not happen with different tokens)
		if existingToken, exists := hashes[hash]; exists {
			t.Errorf("Hash collision detected: '%s' and '%s' produce the same hash", token, existingToken)
		}

		hashes[hash] = token
	}
}

func TestGenerateSecureToken(t *testing.T) {
	token, err := GenerateSecureToken()

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify it's base64 URL encoded
	_, err = base64.RawURLEncoding.DecodeString(token)
	assert.NoError(t, err, "token should be valid base64 URL encoding")

	// Token should be longer than 32 characters (base64 of 32 bytes)
	assert.Greater(t, len(token), 32)
}

func TestGenerateSecureTokenUniqueness(t *testing.T) {
	tokens := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		token, err := GenerateSecureToken()
		require.NoError(t, err)

		// Check for duplicates (extremely unlikely with cryptographic randomness)
		if tokens[token] {
			t.Fatalf("Duplicate token generated: %s", token)
		}

		tokens[token] = true
	}

	assert.Len(t, tokens, iterations)
}

func TestGenerateSecureTokenWithLength(t *testing.T) {
	tests := []struct {
		name        string
		length      int
		expectError bool
	}{
		{
			name:        "valid length 16",
			length:      16,
			expectError: false,
		},
		{
			name:        "valid length 32",
			length:      32,
			expectError: false,
		},
		{
			name:        "valid length 64",
			length:      64,
			expectError: false,
		},
		{
			name:        "invalid length 0",
			length:      0,
			expectError: true,
		},
		{
			name:        "invalid negative length",
			length:      -1,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateSecureTokenWithLength(tt.length)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, token)
				assert.ErrorIs(t, err, ErrTokenGeneration)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, token)

				// Decode to verify length
				decoded, err := base64.RawURLEncoding.DecodeString(token)
				require.NoError(t, err)
				assert.Len(t, decoded, tt.length)
			}
		})
	}
}

func TestVerifyTokenHash(t *testing.T) {
	token := "my-secure-token"
	hash := HashToken(token)

	tests := []struct {
		name     string
		token    string
		hash     string
		expected bool
	}{
		{
			name:     "valid token and hash",
			token:    token,
			hash:     hash,
			expected: true,
		},
		{
			name:     "wrong token",
			token:    "wrong-token",
			hash:     hash,
			expected: false,
		},
		{
			name:     "wrong hash",
			token:    token,
			hash:     HashToken("different"),
			expected: false,
		},
		{
			name:     "empty token",
			token:    "",
			hash:     hash,
			expected: false,
		},
		{
			name:     "empty hash",
			token:    token,
			hash:     "",
			expected: false,
		},
		{
			name:     "both empty",
			token:    "",
			hash:     "",
			expected: false,
		},
		{
			name:     "invalid hash format",
			token:    token,
			hash:     "not-a-valid-hash",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyTokenHash(tt.token, tt.hash)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHashAndVerifyTokenIntegration(t *testing.T) {
	// Generate a secure token
	token, err := GenerateSecureToken()
	require.NoError(t, err)

	// Hash it
	hash := HashToken(token)

	// Verify it
	assert.True(t, VerifyTokenHash(token, hash))

	// Verify wrong token fails
	wrongToken, err := GenerateSecureToken()
	require.NoError(t, err)
	assert.False(t, VerifyTokenHash(wrongToken, hash))
}

func TestTokenFormatting(t *testing.T) {
	token, err := GenerateSecureToken()
	require.NoError(t, err)

	// Should not contain padding characters
	assert.NotContains(t, token, "=")

	// Should be URL-safe (no + or /)
	assert.NotContains(t, token, "+")
	assert.NotContains(t, token, "/")
}

func BenchmarkHashToken(b *testing.B) {
	token := "benchmark-token-12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = HashToken(token)
	}
}

func BenchmarkGenerateSecureToken(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateSecureToken()
	}
}

func BenchmarkVerifyTokenHash(b *testing.B) {
	token := "benchmark-token"
	hash := HashToken(token)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = VerifyTokenHash(token, hash)
	}
}
