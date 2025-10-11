package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sebasr/avt-service/internal/models"
)

func TestTelemetryHandler(t *testing.T) {
	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name: "valid telemetry data",
			payload: models.TelemetryData{
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
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if msg, ok := resp["message"].(string); !ok || msg != "Telemetry data received successfully" {
					t.Errorf("Expected success message, got %v", resp["message"])
				}
				if _, ok := resp["timestamp"]; !ok {
					t.Error("Expected timestamp in response")
				}
			},
		},
		{
			name:           "invalid JSON",
			payload:        "invalid json",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if err, ok := resp["error"].(string); !ok || err != "Invalid JSON payload" {
					t.Errorf("Expected invalid JSON error, got %v", resp["error"])
				}
			},
		},
		{
			name: "missing timestamp",
			payload: models.TelemetryData{
				ITOW: 118286240,
				GPS: models.GpsData{
					Latitude:  42.6719035,
					Longitude: 23.2887238,
				},
				Motion: models.MotionData{
					GForceX: 0.5,
				},
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if err, ok := resp["error"].(string); !ok || err != "Missing required field: timestamp" {
					t.Errorf("Expected missing timestamp error, got %v", resp["error"])
				}
			},
		},
		{
			name: "minimal valid data",
			payload: models.TelemetryData{
				Timestamp: time.Now().UTC(),
				GPS:       models.GpsData{},
				Motion:    models.MotionData{},
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if msg, ok := resp["message"].(string); !ok || msg != "Telemetry data received successfully" {
					t.Errorf("Expected success message, got %v", resp["message"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test router
			router := gin.New()
			router.POST("/api/telemetry", TelemetryHandler)

			// Marshal payload to JSON
			var body []byte
			var err error
			if str, ok := tt.payload.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.payload)
				if err != nil {
					t.Fatalf("Failed to marshal payload: %v", err)
				}
			}

			// Create request
			req, err := http.NewRequest("POST", "/api/telemetry", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check response body
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}
		})
	}
}

func TestTelemetryHandlerContentType(t *testing.T) {
	router := gin.New()
	router.POST("/api/telemetry", TelemetryHandler)

	telemetry := models.TelemetryData{
		Timestamp: time.Now().UTC(),
		GPS:       models.GpsData{},
		Motion:    models.MotionData{},
	}

	body, _ := json.Marshal(telemetry)
	req, _ := http.NewRequest("POST", "/api/telemetry", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type 'application/json; charset=utf-8', got %q", contentType)
	}
}
