package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/sebasr/avt-service/internal/handlers"
)

func init() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
}

func TestServerRoutes(t *testing.T) {
	// Create the server
	router := New()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
		checkJSON      bool
	}{
		{
			name:           "root endpoint",
			method:         "GET",
			path:           "/",
			expectedStatus: http.StatusOK,
			expectedBody:   "Hello, World!",
			checkJSON:      false,
		},
		{
			name:           "health endpoint",
			method:         "GET",
			path:           "/health",
			expectedStatus: http.StatusOK,
			checkJSON:      true,
		},
		{
			name:           "greeting endpoint",
			method:         "GET",
			path:           "/api/greeting/TestUser",
			expectedStatus: http.StatusOK,
			checkJSON:      true,
		},
		{
			name:           "non-existent endpoint",
			method:         "GET",
			path:           "/nonexistent",
			expectedStatus: http.StatusNotFound,
			checkJSON:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check response body for non-JSON responses
			if !tt.checkJSON && tt.expectedBody != "" {
				if w.Body.String() != tt.expectedBody {
					t.Errorf("Expected body %q, got %q", tt.expectedBody, w.Body.String())
				}
			}

			// Validate JSON responses
			if tt.checkJSON && w.Code == http.StatusOK {
				var jsonResponse interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &jsonResponse); err != nil {
					t.Errorf("Expected valid JSON, got error: %v", err)
				}
			}
		})
	}
}

func TestHealthEndpointStructure(t *testing.T) {
	router := New()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response handlers.HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal health response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("Expected status 'ok', got %q", response.Status)
	}
}

func TestGreetingEndpointStructure(t *testing.T) {
	router := New()

	req, _ := http.NewRequest("GET", "/api/greeting/Alice", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response handlers.GreetingResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal greeting response: %v", err)
	}

	expectedMessage := "Hello, Alice!"
	if response.Message != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, response.Message)
	}
}
