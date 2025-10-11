package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sebasr/avt-service/internal/handlers"
	"github.com/sebasr/avt-service/internal/models"
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

func TestTelemetryEndpoint(t *testing.T) {
	router := New()

	// Create sample telemetry data
	telemetry := models.TelemetryData{
		ITOW:      118286240,
		Timestamp: time.Now().UTC(),
		GPS: models.GpsData{
			Latitude:           42.6719035,
			Longitude:          23.2887238,
			WgsAltitude:        625.761,
			MslAltitude:        590.095,
			Speed:              125.5,
			Heading:            270.5,
			NumSatellites:      11,
			FixStatus:          3,
			HorizontalAccuracy: 0.924,
			VerticalAccuracy:   1.836,
			SpeedAccuracy:      0.704,
			HeadingAccuracy:    145.26856,
			PDOP:               3.0,
			IsFixValid:         true,
		},
		Motion: models.MotionData{
			GForceX:   -0.003,
			GForceY:   0.113,
			GForceZ:   0.974,
			RotationX: 2.09,
			RotationY: 0.86,
			RotationZ: 0.04,
		},
		Battery:       89.0,
		IsCharging:    false,
		TimeAccuracy:  25,
		ValidityFlags: 7,
	}

	body, err := json.Marshal(telemetry)
	if err != nil {
		t.Fatalf("Failed to marshal telemetry: %v", err)
	}

	req, _ := http.NewRequest("POST", "/api/telemetry", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify response contains success message
	if msg, ok := response["message"].(string); !ok || msg != "Telemetry data received successfully" {
		t.Errorf("Expected success message, got %v", response["message"])
	}

	// Verify response contains timestamp
	if _, ok := response["timestamp"]; !ok {
		t.Error("Expected timestamp in response")
	}
}
