package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sebasr/avt-service/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestMiddleware() (*AuthMiddleware, *auth.JWTService) {
	jwtService := auth.NewJWTService("test-secret-key", 1*time.Hour, 24*time.Hour)
	middleware := NewAuthMiddleware(jwtService)
	return middleware, jwtService
}

func TestAuthMiddleware_Required_ValidToken(t *testing.T) {
	middleware, jwtService := setupTestMiddleware()

	// Create a valid token
	userID := uuid.New()
	email := "test@example.com"
	token, err := jwtService.GenerateAccessToken(userID, email)
	require.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handlerCalled := false
	var capturedUserID uuid.UUID
	var capturedEmail string
	var getUserIDErr error
	var getUserEmailErr error

	router.GET("/protected", middleware.Required(), func(c *gin.Context) {
		handlerCalled = true
		capturedUserID, getUserIDErr = GetUserID(c)
		capturedEmail, getUserEmailErr = GetUserEmail(c)
		c.Status(http.StatusOK)
	})

	// Create test request
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	router.ServeHTTP(w, req)

	// Verify handler was called
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, getUserIDErr)
	assert.NoError(t, getUserEmailErr)
	assert.Equal(t, userID, capturedUserID)
	assert.Equal(t, email, capturedEmail)
}

func TestAuthMiddleware_Required_NoToken(t *testing.T) {
	middleware, _ := setupTestMiddleware()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	handlerCalled := false
	router.GET("/protected", middleware.Required(), func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)

	router.ServeHTTP(w, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing authorization header")
}

func TestAuthMiddleware_Required_InvalidToken(t *testing.T) {
	middleware, _ := setupTestMiddleware()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	handlerCalled := false
	router.GET("/protected", middleware.Required(), func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	router.ServeHTTP(w, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_Required_ExpiredToken(t *testing.T) {
	// Create JWT service with very short TTL
	jwtService := auth.NewJWTService("test-secret", 1*time.Millisecond, 1*time.Hour)
	middleware := NewAuthMiddleware(jwtService)

	userID := uuid.New()
	email := "expired@example.com"
	token, err := jwtService.GenerateAccessToken(userID, email)
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	handlerCalled := false
	router.GET("/protected", middleware.Required(), func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	router.ServeHTTP(w, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "token has expired")
}

func TestAuthMiddleware_Required_MalformedAuthHeader(t *testing.T) {
	middleware, _ := setupTestMiddleware()

	tests := []struct {
		name   string
		header string
	}{
		{"missing Bearer prefix", "some-token"},
		{"wrong prefix", "Basic some-token"},
		{"empty token", "Bearer "},
		{"only Bearer", "Bearer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			handlerCalled := false
			router.GET("/protected", middleware.Required(), func(c *gin.Context) {
				handlerCalled = true
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", tt.header)

			router.ServeHTTP(w, req)

			assert.False(t, handlerCalled)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestAuthMiddleware_Optional_ValidToken(t *testing.T) {
	middleware, jwtService := setupTestMiddleware()

	userID := uuid.New()
	email := "optional@example.com"
	token, err := jwtService.GenerateAccessToken(userID, email)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	handlerCalled := false
	var capturedUserID uuid.UUID
	var capturedEmail string

	router.GET("/optional", middleware.Optional(), func(c *gin.Context) {
		handlerCalled = true
		var err error
		capturedUserID, err = GetUserID(c)
		require.NoError(t, err)
		capturedEmail, err = GetUserEmail(c)
		require.NoError(t, err)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	router.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, userID, capturedUserID)
	assert.Equal(t, email, capturedEmail)
}

func TestAuthMiddleware_Optional_NoToken(t *testing.T) {
	middleware, _ := setupTestMiddleware()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	handlerCalled := false
	var userIDExists bool

	router.GET("/optional", middleware.Optional(), func(c *gin.Context) {
		handlerCalled = true
		_, err := GetUserID(c)
		userIDExists = (err == nil)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/optional", nil)

	router.ServeHTTP(w, req)

	// Handler should be called even without token
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, userIDExists)
}

func TestAuthMiddleware_Optional_InvalidToken(t *testing.T) {
	middleware, _ := setupTestMiddleware()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	handlerCalled := false
	var userIDExists bool

	router.GET("/optional", middleware.Optional(), func(c *gin.Context) {
		handlerCalled = true
		_, err := GetUserID(c)
		userIDExists = (err == nil)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	router.ServeHTTP(w, req)

	// Handler should be called even with invalid token
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, userIDExists)
}

func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	userID := uuid.New()
	c.Set(string(UserIDKey), userID)

	retrievedID, err := GetUserID(c)
	assert.NoError(t, err)
	assert.Equal(t, userID, retrievedID)
}

func TestGetUserID_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	_, err := GetUserID(c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not authenticated")
}

func TestGetUserID_InvalidType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	c.Set(string(UserIDKey), "not-a-uuid")

	_, err := GetUserID(c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user ID format")
}

func TestGetUserEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	email := "test@example.com"
	c.Set(string(UserEmailKey), email)

	retrievedEmail, err := GetUserEmail(c)
	assert.NoError(t, err)
	assert.Equal(t, email, retrievedEmail)
}

func TestGetUserEmail_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	_, err := GetUserEmail(c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not authenticated")
}

func TestMustGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	userID := uuid.New()
	c.Set(string(UserIDKey), userID)

	retrievedID := MustGetUserID(c)
	assert.Equal(t, userID, retrievedID)
}

func TestMustGetUserID_Panics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	assert.Panics(t, func() {
		MustGetUserID(c)
	})
}

func TestMustGetUserEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	email := "test@example.com"
	c.Set(string(UserEmailKey), email)

	retrievedEmail := MustGetUserEmail(c)
	assert.Equal(t, email, retrievedEmail)
}

func TestMustGetUserEmail_Panics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	assert.Panics(t, func() {
		MustGetUserEmail(c)
	})
}
