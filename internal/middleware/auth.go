package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sebasr/avt-service/internal/auth"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// UserIDKey is the context key for the authenticated user's ID
	UserIDKey ContextKey = "user_id"

	// UserEmailKey is the context key for the authenticated user's email
	UserEmailKey ContextKey = "user_email"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	jwtService *auth.JWTService
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtService *auth.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

// Required returns a middleware that requires a valid JWT token
// Returns 401 Unauthorized if the token is missing or invalid
func (m *AuthMiddleware) Required() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := m.extractAndValidateToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Parse user ID from string to UUID
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "invalid user ID in token",
			})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set(string(UserIDKey), userID)
		c.Set(string(UserEmailKey), claims.Email)

		c.Next()
	}
}

// Optional returns a middleware that extracts user info if a valid token is present
// Continues execution even if the token is missing or invalid
func (m *AuthMiddleware) Optional() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := m.extractAndValidateToken(c)
		if err == nil && claims != nil {
			// Parse user ID from string to UUID
			userID, err := uuid.Parse(claims.UserID)
			if err == nil {
				// Set user information in context if token is valid
				c.Set(string(UserIDKey), userID)
				c.Set(string(UserEmailKey), claims.Email)
			}
		}

		// Continue regardless of authentication status
		c.Next()
	}
}

// extractAndValidateToken extracts the JWT token from the request and validates it
func (m *AuthMiddleware) extractAndValidateToken(c *gin.Context) (*auth.Claims, error) {
	// Extract token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, errors.New("missing authorization header")
	}

	// Check for Bearer token format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, errors.New("invalid authorization header format")
	}

	tokenString := parts[1]
	if tokenString == "" {
		return nil, errors.New("missing token")
	}

	// Validate token
	claims, err := m.jwtService.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

// GetUserID retrieves the authenticated user's ID from the context
func GetUserID(c *gin.Context) (uuid.UUID, error) {
	userID, exists := c.Get(string(UserIDKey))
	if !exists {
		return uuid.Nil, errors.New("user not authenticated")
	}

	id, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("invalid user ID format")
	}

	return id, nil
}

// GetUserEmail retrieves the authenticated user's email from the context
func GetUserEmail(c *gin.Context) (string, error) {
	email, exists := c.Get(string(UserEmailKey))
	if !exists {
		return "", errors.New("user not authenticated")
	}

	emailStr, ok := email.(string)
	if !ok {
		return "", errors.New("invalid email format")
	}

	return emailStr, nil
}

// MustGetUserID retrieves the user ID from context, panics if not found
// Use this only in handlers protected by Required() middleware
func MustGetUserID(c *gin.Context) uuid.UUID {
	userID, err := GetUserID(c)
	if err != nil {
		panic("user ID not found in context - ensure Required() middleware is applied")
	}
	return userID
}

// MustGetUserEmail retrieves the user email from context, panics if not found
// Use this only in handlers protected by Required() middleware
func MustGetUserEmail(c *gin.Context) string {
	email, err := GetUserEmail(c)
	if err != nil {
		panic("user email not found in context - ensure Required() middleware is applied")
	}
	return email
}
