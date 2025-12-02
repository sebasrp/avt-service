package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sebasr/avt-service/internal/auth"
	"github.com/sebasr/avt-service/internal/middleware"
	"github.com/sebasr/avt-service/internal/models"
	"github.com/sebasr/avt-service/internal/repository"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtService       *auth.JWTService
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
	}
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
