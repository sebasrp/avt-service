package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUser_ToResponse(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	lastLogin := now.Add(-1 * time.Hour)

	user := &User{
		ID:            userID,
		Email:         "test@example.com",
		PasswordHash:  "hashed-password-should-not-appear",
		EmailVerified: true,
		CreatedAt:     now.Add(-7 * 24 * time.Hour),
		UpdatedAt:     now,
		LastLoginAt:   &lastLogin,
		IsActive:      true,
	}

	response := user.ToResponse()

	// Verify exposed fields
	assert.Equal(t, userID, response.ID)
	assert.Equal(t, "test@example.com", response.Email)
	assert.True(t, response.EmailVerified)
	assert.Equal(t, user.CreatedAt, response.CreatedAt)
	assert.Equal(t, user.UpdatedAt, response.UpdatedAt)
	assert.Equal(t, &lastLogin, response.LastLoginAt)
	assert.True(t, response.IsActive)

	// Verify password hash is not in response (it has json:"-" tag)
	// We can't directly check this, but the struct definition ensures it
}

func TestUser_ToResponse_NoLastLogin(t *testing.T) {
	now := time.Now()
	userID := uuid.New()

	user := &User{
		ID:            userID,
		Email:         "new@example.com",
		PasswordHash:  "hashed-password",
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
		LastLoginAt:   nil,
		IsActive:      true,
	}

	response := user.ToResponse()

	assert.Equal(t, userID, response.ID)
	assert.Equal(t, "new@example.com", response.Email)
	assert.False(t, response.EmailVerified)
	assert.Nil(t, response.LastLoginAt)
	assert.True(t, response.IsActive)
}

func TestUser_ToResponse_Inactive(t *testing.T) {
	now := time.Now()
	userID := uuid.New()

	user := &User{
		ID:            userID,
		Email:         "inactive@example.com",
		PasswordHash:  "hashed-password",
		EmailVerified: true,
		CreatedAt:     now,
		UpdatedAt:     now,
		LastLoginAt:   nil,
		IsActive:      false, // Inactive user
	}

	response := user.ToResponse()

	assert.Equal(t, userID, response.ID)
	assert.False(t, response.IsActive)
}

func TestUserProfile_ToResponse(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	displayName := "John Doe"
	avatarURL := "https://example.com/avatar.jpg"

	profile := &UserProfile{
		UserID:               userID,
		DisplayName:          &displayName,
		AvatarURL:            &avatarURL,
		Timezone:             "America/New_York",
		UnitsPreference:      "imperial",
		NotificationsEnabled: true,
		CreatedAt:            now.Add(-7 * 24 * time.Hour),
		UpdatedAt:            now,
	}

	response := profile.ToResponse()

	assert.Equal(t, userID, response.UserID)
	assert.Equal(t, &displayName, response.DisplayName)
	assert.Equal(t, &avatarURL, response.AvatarURL)
	assert.Equal(t, "America/New_York", response.Timezone)
	assert.Equal(t, "imperial", response.UnitsPreference)
	assert.True(t, response.NotificationsEnabled)
	assert.Equal(t, profile.CreatedAt, response.CreatedAt)
	assert.Equal(t, profile.UpdatedAt, response.UpdatedAt)
}

func TestUserProfile_ToResponse_MinimalProfile(t *testing.T) {
	now := time.Now()
	userID := uuid.New()

	profile := &UserProfile{
		UserID:               userID,
		DisplayName:          nil,
		AvatarURL:            nil,
		Timezone:             "UTC",
		UnitsPreference:      "metric",
		NotificationsEnabled: false,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	response := profile.ToResponse()

	assert.Equal(t, userID, response.UserID)
	assert.Nil(t, response.DisplayName)
	assert.Nil(t, response.AvatarURL)
	assert.Equal(t, "UTC", response.Timezone)
	assert.Equal(t, "metric", response.UnitsPreference)
	assert.False(t, response.NotificationsEnabled)
}

func TestUserProfile_UnitsPreference(t *testing.T) {
	tests := []struct {
		name       string
		preference string
	}{
		{"metric", "metric"},
		{"imperial", "imperial"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := &UserProfile{
				UserID:          uuid.New(),
				Timezone:        "UTC",
				UnitsPreference: tt.preference,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}

			response := profile.ToResponse()
			assert.Equal(t, tt.preference, response.UnitsPreference)
		})
	}
}

func TestUserWithProfile(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	displayName := "Jane Smith"

	userResponse := &UserResponse{
		ID:            userID,
		Email:         "jane@example.com",
		EmailVerified: true,
		CreatedAt:     now,
		UpdatedAt:     now,
		LastLoginAt:   nil,
		IsActive:      true,
	}

	profileResponse := &UserProfileResponse{
		UserID:               userID,
		DisplayName:          &displayName,
		AvatarURL:            nil,
		Timezone:             "Europe/London",
		UnitsPreference:      "metric",
		NotificationsEnabled: true,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	userWithProfile := &UserWithProfile{
		User:    userResponse,
		Profile: profileResponse,
	}

	assert.NotNil(t, userWithProfile.User)
	assert.NotNil(t, userWithProfile.Profile)
	assert.Equal(t, userID, userWithProfile.User.ID)
	assert.Equal(t, userID, userWithProfile.Profile.UserID)
	assert.Equal(t, "jane@example.com", userWithProfile.User.Email)
	assert.Equal(t, &displayName, userWithProfile.Profile.DisplayName)
}

func TestUserWithProfile_NoProfile(t *testing.T) {
	now := time.Now()
	userID := uuid.New()

	userResponse := &UserResponse{
		ID:            userID,
		Email:         "user@example.com",
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
		LastLoginAt:   nil,
		IsActive:      true,
	}

	userWithProfile := &UserWithProfile{
		User:    userResponse,
		Profile: nil, // No profile yet
	}

	assert.NotNil(t, userWithProfile.User)
	assert.Nil(t, userWithProfile.Profile)
	assert.Equal(t, userID, userWithProfile.User.ID)
}
