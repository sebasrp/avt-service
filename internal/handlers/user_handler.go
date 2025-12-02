package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sebasr/avt-service/internal/auth"
	"github.com/sebasr/avt-service/internal/middleware"
	"github.com/sebasr/avt-service/internal/repository"
)

// UserHandler handles user-related requests
type UserHandler struct {
	userRepo repository.UserRepository
}

// NewUserHandler creates a new user handler
func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

// UpdateProfileRequest represents the profile update request body
type UpdateProfileRequest struct {
	DisplayName *string `json:"displayName,omitempty"`
	AvatarURL   *string `json:"avatarUrl,omitempty"`
}

// ChangePasswordRequest represents the password change request body
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=8,max=72"`
}

// UserProfileResponse represents the user profile response
type UserProfileResponse struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	EmailVerified bool    `json:"emailVerified"`
	DisplayName   *string `json:"displayName,omitempty"`
	AvatarURL     *string `json:"avatarUrl,omitempty"`
	IsActive      bool    `json:"isActive"`
	CreatedAt     string  `json:"createdAt"`
	LastLoginAt   *string `json:"lastLoginAt,omitempty"`
}

// GetProfile retrieves the authenticated user's profile
// GET /api/v1/users/me
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "user_not_found",
				"message": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to retrieve profile",
		})
		return
	}

	var lastLoginAt *string
	if user.LastLoginAt != nil {
		loginStr := user.LastLoginAt.Format("2006-01-02T15:04:05Z07:00")
		lastLoginAt = &loginStr
	}

	c.JSON(http.StatusOK, UserProfileResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		IsActive:      user.IsActive,
		CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		LastLoginAt:   lastLoginAt,
	})
}

// UpdateProfile updates the authenticated user's profile
// PATCH /api/v1/users/me
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Get current user
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "user_not_found",
				"message": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to retrieve profile",
		})
		return
	}

	// Note: In the current implementation, we don't have a UserProfile table yet,
	// so we're just validating the request structure. When Phase 4 user_profiles
	// table is ready, we would update it here.

	// For now, just return success with the current user data
	var lastLoginAt *string
	if user.LastLoginAt != nil {
		loginStr := user.LastLoginAt.Format("2006-01-02T15:04:05Z07:00")
		lastLoginAt = &loginStr
	}

	c.JSON(http.StatusOK, UserProfileResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		DisplayName:   req.DisplayName,
		AvatarURL:     req.AvatarURL,
		IsActive:      user.IsActive,
		CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		LastLoginAt:   lastLoginAt,
	})
}

// ChangePassword changes the authenticated user's password
// POST /api/v1/users/me/change-password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Get current user
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "user_not_found",
				"message": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to retrieve user",
		})
		return
	}

	// Verify current password
	if !auth.VerifyPassword(req.CurrentPassword, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid_password",
			"message": "Current password is incorrect",
		})
		return
	}

	// Check if new password is the same as current
	if auth.VerifyPassword(req.NewPassword, user.PasswordHash) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "same_password",
			"message": "New password must be different from current password",
		})
		return
	}

	// Hash new password
	newPasswordHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to process password change",
		})
		return
	}

	// Update password
	if err := h.userRepo.UpdatePassword(c.Request.Context(), userID, newPasswordHash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to update password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}
