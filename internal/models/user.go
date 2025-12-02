package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user account in the system
type User struct {
	ID                         uuid.UUID  `json:"id" db:"id"`
	Email                      string     `json:"email" db:"email"`
	PasswordHash               string     `json:"-" db:"password_hash"` // Never expose in JSON
	EmailVerified              bool       `json:"emailVerified" db:"email_verified"`
	VerificationToken          *string    `json:"-" db:"verification_token"` // Never expose in JSON
	VerificationTokenExpiresAt *time.Time `json:"-" db:"verification_token_expires_at"`
	ResetToken                 *string    `json:"-" db:"reset_token"` // Never expose in JSON
	ResetTokenExpiresAt        *time.Time `json:"-" db:"reset_token_expires_at"`
	CreatedAt                  time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt                  time.Time  `json:"updatedAt" db:"updated_at"`
	LastLoginAt                *time.Time `json:"lastLoginAt,omitempty" db:"last_login_at"`
	IsActive                   bool       `json:"isActive" db:"is_active"`
}

// UserProfile represents user profile information
type UserProfile struct {
	UserID               uuid.UUID `json:"userId" db:"user_id"`
	DisplayName          *string   `json:"displayName,omitempty" db:"display_name"`
	AvatarURL            *string   `json:"avatarUrl,omitempty" db:"avatar_url"`
	Timezone             string    `json:"timezone" db:"timezone"`
	UnitsPreference      string    `json:"unitsPreference" db:"units_preference"` // "metric" or "imperial"
	NotificationsEnabled bool      `json:"notificationsEnabled" db:"notifications_enabled"`
	CreatedAt            time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt            time.Time `json:"updatedAt" db:"updated_at"`
}

// UserResponse represents a user for API responses (excludes sensitive fields)
type UserResponse struct {
	ID            uuid.UUID  `json:"id"`
	Email         string     `json:"email"`
	EmailVerified bool       `json:"emailVerified"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	LastLoginAt   *time.Time `json:"lastLoginAt,omitempty"`
	IsActive      bool       `json:"isActive"`
}

// ToResponse converts a User to a UserResponse (safe for API)
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:            u.ID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
		LastLoginAt:   u.LastLoginAt,
		IsActive:      u.IsActive,
	}
}

// UserProfileResponse represents a user profile for API responses
type UserProfileResponse struct {
	UserID               uuid.UUID `json:"userId"`
	DisplayName          *string   `json:"displayName,omitempty"`
	AvatarURL            *string   `json:"avatarUrl,omitempty"`
	Timezone             string    `json:"timezone"`
	UnitsPreference      string    `json:"unitsPreference"`
	NotificationsEnabled bool      `json:"notificationsEnabled"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

// ToResponse converts a UserProfile to a UserProfileResponse
func (up *UserProfile) ToResponse() *UserProfileResponse {
	return &UserProfileResponse{
		UserID:               up.UserID,
		DisplayName:          up.DisplayName,
		AvatarURL:            up.AvatarURL,
		Timezone:             up.Timezone,
		UnitsPreference:      up.UnitsPreference,
		NotificationsEnabled: up.NotificationsEnabled,
		CreatedAt:            up.CreatedAt,
		UpdatedAt:            up.UpdatedAt,
	}
}

// UserWithProfile combines user and profile information
type UserWithProfile struct {
	User    *UserResponse        `json:"user"`
	Profile *UserProfileResponse `json:"profile,omitempty"`
}
