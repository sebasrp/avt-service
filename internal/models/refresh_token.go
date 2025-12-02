package models

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token for user authentication
type RefreshToken struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	UserID     uuid.UUID  `json:"userId" db:"user_id"`
	TokenHash  string     `json:"-" db:"token_hash"` // Never expose in JSON - stored as SHA256 hash
	ExpiresAt  time.Time  `json:"expiresAt" db:"expires_at"`
	CreatedAt  time.Time  `json:"createdAt" db:"created_at"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty" db:"revoked_at"`
	ReplacedBy *uuid.UUID `json:"replacedBy,omitempty" db:"replaced_by"` // ID of the token that replaced this one
	UserAgent  string     `json:"userAgent,omitempty" db:"user_agent"`
	IPAddress  string     `json:"ipAddress,omitempty" db:"ip_address"`
}

// IsValid checks if the refresh token is still valid
func (rt *RefreshToken) IsValid() bool {
	// Token is invalid if it's been revoked
	if rt.RevokedAt != nil {
		return false
	}

	// Token is invalid if it's expired
	if time.Now().After(rt.ExpiresAt) {
		return false
	}

	return true
}

// IsExpired checks if the token has expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsRevoked checks if the token has been revoked
func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil
}

// RefreshTokenResponse represents a refresh token for API responses
type RefreshTokenResponse struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"userId"`
	ExpiresAt  time.Time  `json:"expiresAt"`
	CreatedAt  time.Time  `json:"createdAt"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty"`
	ReplacedBy *uuid.UUID `json:"replacedBy,omitempty"`
	UserAgent  string     `json:"userAgent,omitempty"`
	IPAddress  string     `json:"ipAddress,omitempty"`
	IsValid    bool       `json:"isValid"`
}

// ToResponse converts a RefreshToken to a RefreshTokenResponse (safe for API)
func (rt *RefreshToken) ToResponse() *RefreshTokenResponse {
	return &RefreshTokenResponse{
		ID:         rt.ID,
		UserID:     rt.UserID,
		ExpiresAt:  rt.ExpiresAt,
		CreatedAt:  rt.CreatedAt,
		RevokedAt:  rt.RevokedAt,
		ReplacedBy: rt.ReplacedBy,
		UserAgent:  rt.UserAgent,
		IPAddress:  rt.IPAddress,
		IsValid:    rt.IsValid(),
	}
}
