package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sebasr/avt-service/internal/auth"
	"github.com/sebasr/avt-service/internal/middleware"
	"github.com/sebasr/avt-service/internal/models"
	"github.com/sebasr/avt-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserTest() (*UserHandler, *repository.MockUserRepository) {
	userRepo := repository.NewMockUserRepository()
	handler := NewUserHandler(userRepo)

	gin.SetMode(gin.TestMode)

	return handler, userRepo
}

func TestUserHandler_GetProfile_Success(t *testing.T) {
	handler, userRepo := setupUserTest()

	userID := uuid.New()
	lastLogin := time.Now().Add(-1 * time.Hour)
	user := &models.User{
		ID:            userID,
		Email:         "test@example.com",
		EmailVerified: true,
		IsActive:      true,
		CreatedAt:     time.Now().Add(-30 * 24 * time.Hour),
		LastLoginAt:   &lastLogin,
	}

	userRepo.GetByIDFunc = func(_ context.Context, id uuid.UUID) (*models.User, error) {
		if id == userID {
			return user, nil
		}
		return nil, repository.ErrUserNotFound
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	c.Set(string(middleware.UserIDKey), userID)

	handler.GetProfile(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response UserProfileResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, userID.String(), response.ID)
	assert.Equal(t, "test@example.com", response.Email)
	assert.True(t, response.EmailVerified)
	assert.True(t, response.IsActive)
	assert.NotEmpty(t, response.CreatedAt)
	assert.NotNil(t, response.LastLoginAt)
}

func TestUserHandler_GetProfile_UserNotFound(t *testing.T) {
	handler, userRepo := setupUserTest()

	userID := uuid.New()

	userRepo.GetByIDFunc = func(_ context.Context, _ uuid.UUID) (*models.User, error) {
		return nil, repository.ErrUserNotFound
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	c.Set(string(middleware.UserIDKey), userID)

	handler.GetProfile(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "user_not_found")
}

func TestUserHandler_UpdateProfile_Success(t *testing.T) {
	handler, userRepo := setupUserTest()

	userID := uuid.New()
	user := &models.User{
		ID:            userID,
		Email:         "test@example.com",
		EmailVerified: false,
		IsActive:      true,
		CreatedAt:     time.Now().Add(-10 * 24 * time.Hour),
	}

	userRepo.GetByIDFunc = func(_ context.Context, id uuid.UUID) (*models.User, error) {
		if id == userID {
			return user, nil
		}
		return nil, repository.ErrUserNotFound
	}

	displayName := "Test User"
	avatarURL := "https://example.com/avatar.jpg"
	reqBody := UpdateProfileRequest{
		DisplayName: &displayName,
		AvatarURL:   &avatarURL,
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/users/me", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware.UserIDKey), userID)

	handler.UpdateProfile(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response UserProfileResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, userID.String(), response.ID)
	assert.Equal(t, displayName, *response.DisplayName)
	assert.Equal(t, avatarURL, *response.AvatarURL)
}

func TestUserHandler_UpdateProfile_InvalidRequest(t *testing.T) {
	handler, _ := setupUserTest()

	userID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/users/me", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware.UserIDKey), userID)

	handler.UpdateProfile(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_request")
}

func TestUserHandler_ChangePassword_Success(t *testing.T) {
	handler, userRepo := setupUserTest()

	userID := uuid.New()
	currentPasswordHash, _ := auth.HashPassword("oldpassword123")
	user := &models.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: currentPasswordHash,
		IsActive:     true,
	}

	userRepo.GetByIDFunc = func(_ context.Context, id uuid.UUID) (*models.User, error) {
		if id == userID {
			return user, nil
		}
		return nil, repository.ErrUserNotFound
	}

	var updatedPasswordHash string
	userRepo.UpdatePasswordFunc = func(_ context.Context, id uuid.UUID, passwordHash string) error {
		if id == userID {
			updatedPasswordHash = passwordHash
			return nil
		}
		return repository.ErrUserNotFound
	}

	reqBody := ChangePasswordRequest{
		CurrentPassword: "oldpassword123",
		NewPassword:     "newpassword456",
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/users/me/change-password", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware.UserIDKey), userID)

	handler.ChangePassword(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Password changed successfully")
	assert.NotEmpty(t, updatedPasswordHash)
	assert.NotEqual(t, currentPasswordHash, updatedPasswordHash)
	assert.True(t, auth.VerifyPassword("newpassword456", updatedPasswordHash))
}

func TestUserHandler_ChangePassword_InvalidCurrentPassword(t *testing.T) {
	handler, userRepo := setupUserTest()

	userID := uuid.New()
	currentPasswordHash, _ := auth.HashPassword("correctpassword")
	user := &models.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: currentPasswordHash,
		IsActive:     true,
	}

	userRepo.GetByIDFunc = func(_ context.Context, id uuid.UUID) (*models.User, error) {
		if id == userID {
			return user, nil
		}
		return nil, repository.ErrUserNotFound
	}

	reqBody := ChangePasswordRequest{
		CurrentPassword: "wrongpassword",
		NewPassword:     "newpassword456",
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/users/me/change-password", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware.UserIDKey), userID)

	handler.ChangePassword(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_password")
}

func TestUserHandler_ChangePassword_SameAsCurrentPassword(t *testing.T) {
	handler, userRepo := setupUserTest()

	userID := uuid.New()
	currentPasswordHash, _ := auth.HashPassword("samepassword123")
	user := &models.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: currentPasswordHash,
		IsActive:     true,
	}

	userRepo.GetByIDFunc = func(_ context.Context, id uuid.UUID) (*models.User, error) {
		if id == userID {
			return user, nil
		}
		return nil, repository.ErrUserNotFound
	}

	reqBody := ChangePasswordRequest{
		CurrentPassword: "samepassword123",
		NewPassword:     "samepassword123",
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/users/me/change-password", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware.UserIDKey), userID)

	handler.ChangePassword(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "same_password")
}

func TestUserHandler_ChangePassword_InvalidRequest(t *testing.T) {
	handler, _ := setupUserTest()

	userID := uuid.New()

	tests := []struct {
		name    string
		body    interface{}
		wantErr string
	}{
		{
			name:    "missing current password",
			body:    map[string]string{"newPassword": "newpass123"},
			wantErr: "invalid_request",
		},
		{
			name:    "missing new password",
			body:    map[string]string{"currentPassword": "oldpass123"},
			wantErr: "invalid_request",
		},
		{
			name:    "new password too short",
			body:    map[string]string{"currentPassword": "oldpass123", "newPassword": "short"},
			wantErr: "invalid_request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/users/me/change-password", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Set(string(middleware.UserIDKey), userID)

			handler.ChangePassword(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantErr)
		})
	}
}

func TestUserHandler_ChangePassword_UserNotFound(t *testing.T) {
	handler, userRepo := setupUserTest()

	userID := uuid.New()

	userRepo.GetByIDFunc = func(_ context.Context, _ uuid.UUID) (*models.User, error) {
		return nil, repository.ErrUserNotFound
	}

	reqBody := ChangePasswordRequest{
		CurrentPassword: "oldpassword123",
		NewPassword:     "newpassword456",
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/users/me/change-password", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware.UserIDKey), userID)

	handler.ChangePassword(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "user_not_found")
}
