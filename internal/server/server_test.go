package server

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

func init() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
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

func TestNonExistentRoute(t *testing.T) {
	router := New()

	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 404 for non-existent routes
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
