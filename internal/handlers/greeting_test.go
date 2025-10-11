package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGreetingHandler(t *testing.T) {
	tests := []struct {
		name        string
		paramName   string
		expectedMsg string
	}{
		{
			name:        "greeting with simple name",
			paramName:   "John",
			expectedMsg: "Hello, John!",
		},
		{
			name:        "greeting with complex name",
			paramName:   "Mary-Jane",
			expectedMsg: "Hello, Mary-Jane!",
		},
		{
			name:        "greeting with space",
			paramName:   "John Doe",
			expectedMsg: "Hello, John Doe!",
		},
		{
			name:        "greeting with single letter",
			paramName:   "A",
			expectedMsg: "Hello, A!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test router
			router := gin.New()
			router.GET("/api/greeting/:name", GreetingHandler)

			// Create a test request
			req, err := http.NewRequest("GET", "/api/greeting/"+tt.paramName, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Create a response recorder
			w := httptest.NewRecorder()

			// Perform the request
			router.ServeHTTP(w, req)

			// Check status code
			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}

			// Check response body
			var response GreetingResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Message != tt.expectedMsg {
				t.Errorf("Expected message %q, got %q", tt.expectedMsg, response.Message)
			}

			// Check content type
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json; charset=utf-8" {
				t.Errorf("Expected Content-Type %q, got %q", "application/json; charset=utf-8", contentType)
			}
		})
	}
}
