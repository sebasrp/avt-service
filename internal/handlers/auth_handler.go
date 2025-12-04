package handlers

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sebasr/avt-service/internal/auth"
	"github.com/sebasr/avt-service/internal/email"
	"github.com/sebasr/avt-service/internal/middleware"
	"github.com/sebasr/avt-service/internal/models"
	"github.com/sebasr/avt-service/internal/repository"
)

// Default reset token TTL (12 hours)
const defaultResetTokenTTL = 12 * time.Hour

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtService       *auth.JWTService
	emailService     email.Service
	resetTokenTTL    time.Duration
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	jwtService *auth.JWTService,
) *AuthHandler {
	return &AuthHandler{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
		resetTokenTTL:    defaultResetTokenTTL,
	}
}

// WithEmailService sets the email service for password reset functionality
func (h *AuthHandler) WithEmailService(emailService email.Service) *AuthHandler {
	h.emailService = emailService
	return h
}

// WithResetTokenTTL sets the reset token TTL
func (h *AuthHandler) WithResetTokenTTL(ttl time.Duration) *AuthHandler {
	h.resetTokenTTL = ttl
	return h
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest represents the token refresh request body
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// ForgotPasswordRequest represents the forgot password request body
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents the password reset request body
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8,max=72"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	User         UserInfo  `json:"user"`
}

// UserInfo represents basic user information
type UserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
}

// Register handles user registration
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Normalize email
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Check if user already exists
	existingUser, err := h.userRepo.GetByEmail(c.Request.Context(), email)
	if err == nil && existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "user_exists",
			"message": "A user with this email already exists",
		})
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to process registration",
		})
		return
	}

	// Create user
	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "user_exists",
				"message": "A user with this email already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to create user",
		})
		return
	}

	// Generate tokens
	accessToken, err := h.jwtService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to generate access token",
		})
		return
	}

	refreshTokenString, expiresAt, err := h.jwtService.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to generate refresh token",
		})
		return
	}

	// Store refresh token
	refreshToken := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: auth.HashToken(refreshTokenString),
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		UserAgent: c.Request.UserAgent(),
		IPAddress: c.ClientIP(),
	}

	if err := h.refreshTokenRepo.Create(c.Request.Context(), refreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to create session",
		})
		return
	}

	// Return tokens
	c.JSON(http.StatusCreated, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		ExpiresAt:    expiresAt,
		User: UserInfo{
			ID:            user.ID.String(),
			Email:         user.Email,
			EmailVerified: user.EmailVerified,
		},
	})
}

// Login handles user login
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Normalize email
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Get user
	user, err := h.userRepo.GetByEmail(c.Request.Context(), email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_credentials",
				"message": "Invalid email or password",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to authenticate",
		})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "account_disabled",
			"message": "This account has been disabled",
		})
		return
	}

	// Verify password
	if !auth.VerifyPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid_credentials",
			"message": "Invalid email or password",
		})
		return
	}

	// Update last login (non-blocking)
	_ = h.userRepo.UpdateLastLogin(c.Request.Context(), user.ID)

	// Generate tokens
	accessToken, err := h.jwtService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to generate access token",
		})
		return
	}

	refreshTokenString, expiresAt, err := h.jwtService.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to generate refresh token",
		})
		return
	}

	// Store refresh token
	refreshToken := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: auth.HashToken(refreshTokenString),
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		UserAgent: c.Request.UserAgent(),
		IPAddress: c.ClientIP(),
	}

	if err := h.refreshTokenRepo.Create(c.Request.Context(), refreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to create session",
		})
		return
	}

	// Return tokens
	c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		ExpiresAt:    expiresAt,
		User: UserInfo{
			ID:            user.ID.String(),
			Email:         user.Email,
			EmailVerified: user.EmailVerified,
		},
	})
}

// RefreshToken handles token refresh
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Validate the refresh token
	claims, err := h.jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid_token",
			"message": "Invalid or expired refresh token",
		})
		return
	}

	// Check if token exists in database and is not revoked
	tokenHash := auth.HashToken(req.RefreshToken)
	storedToken, err := h.refreshTokenRepo.GetByHash(c.Request.Context(), tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrRefreshTokenNotFound) || errors.Is(err, repository.ErrRefreshTokenRevoked) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_token",
				"message": "Invalid or revoked refresh token",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to validate token",
		})
		return
	}

	// Parse user ID from claims
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid_token",
			"message": "Invalid user ID in token",
		})
		return
	}

	// Get user to ensure they still exist and are active
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid_token",
			"message": "User not found",
		})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "account_disabled",
			"message": "This account has been disabled",
		})
		return
	}

	// Generate new tokens
	newAccessToken, err := h.jwtService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to generate access token",
		})
		return
	}

	newRefreshTokenString, expiresAt, err := h.jwtService.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to generate refresh token",
		})
		return
	}

	// Revoke old refresh token (token rotation, non-blocking)
	_ = h.refreshTokenRepo.RevokeByHash(c.Request.Context(), tokenHash)

	// Store new refresh token
	newRefreshToken := &models.RefreshToken{
		ID:         uuid.New(),
		UserID:     user.ID,
		TokenHash:  auth.HashToken(newRefreshTokenString),
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
		ReplacedBy: &storedToken.ID,
		UserAgent:  c.Request.UserAgent(),
		IPAddress:  c.ClientIP(),
	}

	if err := h.refreshTokenRepo.Create(c.Request.Context(), newRefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to create session",
		})
		return
	}

	// Return new tokens
	c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshTokenString,
		ExpiresAt:    expiresAt,
		User: UserInfo{
			ID:            user.ID.String(),
			Email:         user.Email,
			EmailVerified: user.EmailVerified,
		},
	})
}

// Logout handles user logout
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Not authenticated",
		})
		return
	}

	// Revoke all refresh tokens for this user
	if err := h.refreshTokenRepo.RevokeAllForUser(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to logout",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}

// ForgotPassword initiates the password reset flow
// POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Normalize email
	emailAddr := strings.ToLower(strings.TrimSpace(req.Email))

	// Always return success to prevent email enumeration attacks
	// We do the work asynchronously or just silently fail
	defer func() {
		c.JSON(http.StatusOK, gin.H{
			"message": "If an account with that email exists, a password reset link has been sent",
		})
	}()

	// Check if email service is configured
	if h.emailService == nil {
		log.Printf("Warning: Email service not configured, skipping password reset email for %s", emailAddr)
		return
	}

	// Look up user by email
	user, err := h.userRepo.GetByEmail(c.Request.Context(), emailAddr)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			// User not found - return success anyway to prevent enumeration
			return
		}
		log.Printf("Error looking up user for password reset: %v", err)
		return
	}

	// Check if user is active
	if !user.IsActive {
		return
	}

	// Generate secure reset token
	resetToken, err := auth.GenerateSecureToken()
	if err != nil {
		log.Printf("Error generating reset token: %v", err)
		return
	}

	// Hash the token for storage (we store the hash, send the plain token)
	hashedToken := auth.HashToken(resetToken)

	// Set expiration time
	expiresAt := time.Now().Add(h.resetTokenTTL)

	// Store the hashed token
	if err := h.userRepo.SetResetToken(c.Request.Context(), user.ID, hashedToken, &expiresAt); err != nil {
		log.Printf("Error storing reset token: %v", err)
		return
	}

	// Send the password reset email (with plain token)
	if err := h.emailService.SendPasswordResetEmail(c.Request.Context(), user.Email, resetToken); err != nil {
		log.Printf("Error sending password reset email: %v", err)
		// Don't return error to user - token is saved, they could try again
		return
	}
}

// ResetPassword completes the password reset flow
// POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Hash the provided token to look it up
	hashedToken := auth.HashToken(req.Token)

	// Find user by reset token hash
	user, err := h.userRepo.GetByResetToken(c.Request.Context(), hashedToken)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_token",
				"message": "Invalid or expired reset token",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to process reset request",
		})
		return
	}

	// Check if token is expired
	if user.ResetTokenExpiresAt == nil || user.ResetTokenExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "expired_token",
			"message": "Reset token has expired",
		})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "account_disabled",
			"message": "This account has been disabled",
		})
		return
	}

	// Hash the new password
	newPasswordHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to process password",
		})
		return
	}

	// Update the password
	if err := h.userRepo.UpdatePassword(c.Request.Context(), user.ID, newPasswordHash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to update password",
		})
		return
	}

	// Clear the reset token
	if err := h.userRepo.ClearResetToken(c.Request.Context(), user.ID); err != nil {
		log.Printf("Error clearing reset token: %v", err)
		// Non-critical, continue
	}

	// Revoke all refresh tokens for security
	if err := h.refreshTokenRepo.RevokeAllForUser(c.Request.Context(), user.ID); err != nil {
		log.Printf("Error revoking refresh tokens after password reset: %v", err)
		// Non-critical, continue
	}

	// Send password changed notification email
	if h.emailService != nil {
		if err := h.emailService.SendPasswordChangedEmail(c.Request.Context(), user.Email); err != nil {
			log.Printf("Error sending password changed email: %v", err)
			// Non-critical, continue
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password has been reset successfully",
	})
}
