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
	"github.com/sebasr/avt-service/internal/email"
	"github.com/sebasr/avt-service/internal/middleware"
	"github.com/sebasr/avt-service/internal/models"
	"github.com/sebasr/avt-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuthTest() (*AuthHandler, *repository.MockUserRepository, *repository.MockRefreshTokenRepository, *auth.JWTService) {
	userRepo := repository.NewMockUserRepository()
	refreshTokenRepo := repository.NewMockRefreshTokenRepository()
	jwtService := auth.NewJWTService("test-secret", 1*time.Hour, 24*time.Hour)
	handler := NewAuthHandler(userRepo, refreshTokenRepo, jwtService)

	gin.SetMode(gin.TestMode)

	return handler, userRepo, refreshTokenRepo, jwtService
}

func TestAuthHandler_Register_Success(t *testing.T) {
	handler, userRepo, refreshTokenRepo, _ := setupAuthTest()

	var capturedUser *models.User
	var capturedRefreshToken *models.RefreshToken

	userRepo.GetByEmailFunc = func(_ context.Context, _ string) (*models.User, error) {
		return nil, repository.ErrUserNotFound
	}

	userRepo.CreateFunc = func(_ context.Context, user *models.User) error {
		capturedUser = user
		return nil
	}

	refreshTokenRepo.CreateFunc = func(_ context.Context, token *models.RefreshToken) error {
		capturedRefreshToken = token
		return nil
	}

	reqBody := RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Register(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, "test@example.com", response.User.Email)
	assert.False(t, response.User.EmailVerified)

	assert.NotNil(t, capturedUser)
	assert.Equal(t, "test@example.com", capturedUser.Email)
	assert.NotEmpty(t, capturedUser.PasswordHash)
	assert.True(t, capturedUser.IsActive)

	assert.NotNil(t, capturedRefreshToken)
	assert.Equal(t, capturedUser.ID, capturedRefreshToken.UserID)
}

func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	handler, userRepo, _, _ := setupAuthTest()

	existingUser := &models.User{
		ID:    uuid.New(),
		Email: "existing@example.com",
	}

	userRepo.GetByEmailFunc = func(_ context.Context, email string) (*models.User, error) {
		if email == "existing@example.com" {
			return existingUser, nil
		}
		return nil, repository.ErrUserNotFound
	}

	reqBody := RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Register(c)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "user_exists")
}

func TestAuthHandler_Register_InvalidRequest(t *testing.T) {
	handler, _, _, _ := setupAuthTest()

	tests := []struct {
		name    string
		body    interface{}
		wantErr string
	}{
		{
			name:    "missing email",
			body:    map[string]string{"password": "password123"},
			wantErr: "invalid_request",
		},
		{
			name:    "invalid email",
			body:    map[string]string{"email": "not-an-email", "password": "password123"},
			wantErr: "invalid_request",
		},
		{
			name:    "password too short",
			body:    map[string]string{"email": "test@example.com", "password": "short"},
			wantErr: "invalid_request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.Register(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantErr)
		})
	}
}

func TestAuthHandler_Login_Success(t *testing.T) {
	handler, userRepo, refreshTokenRepo, _ := setupAuthTest()

	passwordHash, _ := auth.HashPassword("password123")
	user := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		IsActive:     true,
	}

	userRepo.GetByEmailFunc = func(_ context.Context, email string) (*models.User, error) {
		if email == "test@example.com" {
			return user, nil
		}
		return nil, repository.ErrUserNotFound
	}

	userRepo.UpdateLastLoginFunc = func(_ context.Context, _ uuid.UUID) error {
		return nil
	}

	refreshTokenRepo.CreateFunc = func(_ context.Context, _ *models.RefreshToken) error {
		return nil
	}

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, "test@example.com", response.User.Email)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	handler, userRepo, _, _ := setupAuthTest()

	passwordHash, _ := auth.HashPassword("correctpassword")
	user := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		IsActive:     true,
	}

	userRepo.GetByEmailFunc = func(_ context.Context, email string) (*models.User, error) {
		if email == "test@example.com" {
			return user, nil
		}
		return nil, repository.ErrUserNotFound
	}

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_credentials")
}

func TestAuthHandler_Login_UserNotFound(t *testing.T) {
	handler, userRepo, _, _ := setupAuthTest()

	userRepo.GetByEmailFunc = func(_ context.Context, _ string) (*models.User, error) {
		return nil, repository.ErrUserNotFound
	}

	reqBody := LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_credentials")
}

func TestAuthHandler_Login_InactiveUser(t *testing.T) {
	handler, userRepo, _, _ := setupAuthTest()

	passwordHash, _ := auth.HashPassword("password123")
	user := &models.User{
		ID:           uuid.New(),
		Email:        "inactive@example.com",
		PasswordHash: passwordHash,
		IsActive:     false,
	}

	userRepo.GetByEmailFunc = func(_ context.Context, _ string) (*models.User, error) {
		return user, nil
	}

	reqBody := LoginRequest{
		Email:    "inactive@example.com",
		Password: "password123",
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "account_disabled")
}

func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	handler, userRepo, refreshTokenRepo, jwtService := setupAuthTest()

	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		Email:    "test@example.com",
		IsActive: true,
	}

	// Generate a valid refresh token
	refreshTokenString, expiresAt, _ := jwtService.GenerateRefreshToken(userID, "test@example.com")
	tokenHash := auth.HashToken(refreshTokenString)

	storedToken := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	refreshTokenRepo.GetByHashFunc = func(_ context.Context, hash string) (*models.RefreshToken, error) {
		if hash == tokenHash {
			return storedToken, nil
		}
		return nil, repository.ErrRefreshTokenNotFound
	}

	refreshTokenRepo.RevokeByHashFunc = func(_ context.Context, _ string) error {
		return nil
	}

	refreshTokenRepo.CreateFunc = func(_ context.Context, _ *models.RefreshToken) error {
		return nil
	}

	userRepo.GetByIDFunc = func(_ context.Context, id uuid.UUID) (*models.User, error) {
		if id == userID {
			return user, nil
		}
		return nil, repository.ErrUserNotFound
	}

	reqBody := RefreshTokenRequest{
		RefreshToken: refreshTokenString,
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.RefreshToken(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
	// Note: Token might be the same if generated in the same second, which is fine
}

func TestAuthHandler_RefreshToken_InvalidToken(t *testing.T) {
	handler, _, _, _ := setupAuthTest()

	reqBody := RefreshTokenRequest{
		RefreshToken: "invalid-token",
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.RefreshToken(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_token")
}

func TestAuthHandler_RefreshToken_RevokedToken(t *testing.T) {
	handler, _, refreshTokenRepo, jwtService := setupAuthTest()

	userID := uuid.New()
	refreshTokenString, _, _ := jwtService.GenerateRefreshToken(userID, "test@example.com")
	tokenHash := auth.HashToken(refreshTokenString)

	refreshTokenRepo.GetByHashFunc = func(_ context.Context, hash string) (*models.RefreshToken, error) {
		if hash == tokenHash {
			return nil, repository.ErrRefreshTokenRevoked
		}
		return nil, repository.ErrRefreshTokenNotFound
	}

	reqBody := RefreshTokenRequest{
		RefreshToken: refreshTokenString,
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.RefreshToken(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_token")
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	handler, _, refreshTokenRepo, _ := setupAuthTest()

	userID := uuid.New()

	refreshTokenRepo.RevokeAllForUserFunc = func(_ context.Context, id uuid.UUID) error {
		assert.Equal(t, userID, id)
		return nil
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	c.Set(string(middleware.UserIDKey), userID)

	handler.Logout(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Successfully logged out")
}

func TestAuthHandler_Logout_Unauthorized(t *testing.T) {
	handler, _, _, _ := setupAuthTest()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	// Don't set user ID in context

	handler.Logout(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}

func TestAuthHandler_ForgotPassword_Success(t *testing.T) {
	handler, userRepo, _, _ := setupAuthTest()

	// Configure mock email service
	mockEmailService := email.NewMockService()
	handler = handler.WithEmailService(mockEmailService)

	user := &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		IsActive: true,
	}

	userRepo.GetByEmailFunc = func(_ context.Context, emailAddr string) (*models.User, error) {
		if emailAddr == "test@example.com" {
			return user, nil
		}
		return nil, repository.ErrUserNotFound
	}

	var setResetTokenCalled bool
	userRepo.SetResetTokenFunc = func(_ context.Context, id uuid.UUID, token string, expiresAt *time.Time) error {
		setResetTokenCalled = true
		assert.Equal(t, user.ID, id)
		assert.NotEmpty(t, token)
		assert.NotNil(t, expiresAt)
		assert.True(t, expiresAt.After(time.Now()))
		return nil
	}

	reqBody := ForgotPasswordRequest{
		Email: "test@example.com",
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ForgotPassword(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "If an account with that email exists")
	assert.True(t, setResetTokenCalled)

	// Verify email was sent
	emails := mockEmailService.GetPasswordResetEmails()
	assert.Len(t, emails, 1)
	assert.Equal(t, "test@example.com", emails[0].To)
}

func TestAuthHandler_ForgotPassword_UserNotFound(t *testing.T) {
	handler, userRepo, _, _ := setupAuthTest()

	userRepo.GetByEmailFunc = func(_ context.Context, _ string) (*models.User, error) {
		return nil, repository.ErrUserNotFound
	}

	// ForgotPassword should still return success to prevent email enumeration
	reqBody := ForgotPasswordRequest{
		Email: "nonexistent@example.com",
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ForgotPassword(c)

	// Should return success anyway (prevents email enumeration)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "If an account with that email exists")
}

func TestAuthHandler_ForgotPassword_InactiveUser(t *testing.T) {
	handler, userRepo, _, _ := setupAuthTest()

	user := &models.User{
		ID:       uuid.New(),
		Email:    "inactive@example.com",
		IsActive: false,
	}

	userRepo.GetByEmailFunc = func(_ context.Context, _ string) (*models.User, error) {
		return user, nil
	}

	var setResetTokenCalled bool
	userRepo.SetResetTokenFunc = func(_ context.Context, _ uuid.UUID, _ string, _ *time.Time) error {
		setResetTokenCalled = true
		return nil
	}

	reqBody := ForgotPasswordRequest{
		Email: "inactive@example.com",
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ForgotPassword(c)

	// Should return success but NOT call SetResetToken for inactive users
	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, setResetTokenCalled)
}

func TestAuthHandler_ForgotPassword_InvalidRequest(t *testing.T) {
	handler, _, _, _ := setupAuthTest()

	tests := []struct {
		name    string
		body    interface{}
		wantErr string
	}{
		{
			name:    "missing email",
			body:    map[string]string{},
			wantErr: "invalid_request",
		},
		{
			name:    "invalid email format",
			body:    map[string]string{"email": "not-an-email"},
			wantErr: "invalid_request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.ForgotPassword(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantErr)
		})
	}
}

func TestAuthHandler_ResetPassword_Success(t *testing.T) {
	handler, userRepo, refreshTokenRepo, _ := setupAuthTest()

	userID := uuid.New()
	resetToken := "test-reset-token"
	hashedToken := auth.HashToken(resetToken)
	expiresAt := time.Now().Add(1 * time.Hour)

	user := &models.User{
		ID:                  userID,
		Email:               "test@example.com",
		PasswordHash:        "old-hash",
		ResetToken:          &hashedToken,
		ResetTokenExpiresAt: &expiresAt,
		IsActive:            true,
	}

	userRepo.GetByResetTokenFunc = func(_ context.Context, token string) (*models.User, error) {
		if token == hashedToken {
			return user, nil
		}
		return nil, repository.ErrUserNotFound
	}

	var newPasswordHash string
	userRepo.UpdatePasswordFunc = func(_ context.Context, id uuid.UUID, hash string) error {
		assert.Equal(t, userID, id)
		newPasswordHash = hash
		return nil
	}

	var clearResetTokenCalled bool
	userRepo.ClearResetTokenFunc = func(_ context.Context, id uuid.UUID) error {
		clearResetTokenCalled = true
		assert.Equal(t, userID, id)
		return nil
	}

	var revokeAllCalled bool
	refreshTokenRepo.RevokeAllForUserFunc = func(_ context.Context, id uuid.UUID) error {
		revokeAllCalled = true
		assert.Equal(t, userID, id)
		return nil
	}

	reqBody := ResetPasswordRequest{
		Token:       resetToken,
		NewPassword: "newpassword123",
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ResetPassword(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Password has been reset successfully")
	assert.NotEmpty(t, newPasswordHash)
	assert.True(t, clearResetTokenCalled)
	assert.True(t, revokeAllCalled)
}

func TestAuthHandler_ResetPassword_InvalidToken(t *testing.T) {
	handler, userRepo, _, _ := setupAuthTest()

	userRepo.GetByResetTokenFunc = func(_ context.Context, _ string) (*models.User, error) {
		return nil, repository.ErrUserNotFound
	}

	reqBody := ResetPasswordRequest{
		Token:       "invalid-token",
		NewPassword: "newpassword123",
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ResetPassword(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_token")
}

func TestAuthHandler_ResetPassword_ExpiredToken(t *testing.T) {
	handler, userRepo, _, _ := setupAuthTest()

	userID := uuid.New()
	hashedToken := auth.HashToken("expired-token")
	// Token expired 1 hour ago
	expiresAt := time.Now().Add(-1 * time.Hour)

	user := &models.User{
		ID:                  userID,
		Email:               "test@example.com",
		ResetToken:          &hashedToken,
		ResetTokenExpiresAt: &expiresAt,
		IsActive:            true,
	}

	userRepo.GetByResetTokenFunc = func(_ context.Context, token string) (*models.User, error) {
		if token == hashedToken {
			return user, nil
		}
		return nil, repository.ErrUserNotFound
	}

	reqBody := ResetPasswordRequest{
		Token:       "expired-token",
		NewPassword: "newpassword123",
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ResetPassword(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "expired_token")
}

func TestAuthHandler_ResetPassword_InactiveUser(t *testing.T) {
	handler, userRepo, _, _ := setupAuthTest()

	userID := uuid.New()
	hashedToken := auth.HashToken("valid-token")
	expiresAt := time.Now().Add(1 * time.Hour)

	user := &models.User{
		ID:                  userID,
		Email:               "inactive@example.com",
		ResetToken:          &hashedToken,
		ResetTokenExpiresAt: &expiresAt,
		IsActive:            false, // Inactive user
	}

	userRepo.GetByResetTokenFunc = func(_ context.Context, token string) (*models.User, error) {
		if token == hashedToken {
			return user, nil
		}
		return nil, repository.ErrUserNotFound
	}

	reqBody := ResetPasswordRequest{
		Token:       "valid-token",
		NewPassword: "newpassword123",
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ResetPassword(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "account_disabled")
}

func TestAuthHandler_ResetPassword_InvalidRequest(t *testing.T) {
	handler, _, _, _ := setupAuthTest()

	tests := []struct {
		name    string
		body    interface{}
		wantErr string
	}{
		{
			name:    "missing token",
			body:    map[string]string{"newPassword": "newpassword123"},
			wantErr: "invalid_request",
		},
		{
			name:    "missing password",
			body:    map[string]string{"token": "some-token"},
			wantErr: "invalid_request",
		},
		{
			name:    "password too short",
			body:    map[string]string{"token": "some-token", "newPassword": "short"},
			wantErr: "invalid_request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.ResetPassword(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantErr)
		})
	}
}

func TestAuthHandler_ResetPassword_NilExpiresAt(t *testing.T) {
	handler, userRepo, _, _ := setupAuthTest()

	userID := uuid.New()
	hashedToken := auth.HashToken("token-with-nil-expiry")

	user := &models.User{
		ID:                  userID,
		Email:               "test@example.com",
		ResetToken:          &hashedToken,
		ResetTokenExpiresAt: nil, // No expiry set
		IsActive:            true,
	}

	userRepo.GetByResetTokenFunc = func(_ context.Context, token string) (*models.User, error) {
		if token == hashedToken {
			return user, nil
		}
		return nil, repository.ErrUserNotFound
	}

	reqBody := ResetPasswordRequest{
		Token:       "token-with-nil-expiry",
		NewPassword: "newpassword123",
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ResetPassword(c)

	// Should be treated as expired since there's no valid expiry
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "expired_token")
}
