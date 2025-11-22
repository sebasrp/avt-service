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
	"github.com/sebasr/avt-service/internal/repository"
)

func init() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
}

func TestTelemetryEndpoint(t *testing.T) {
	// Create a mock repository
	mockRepo := repository.NewMockRepository()
	router := New(mockRepo)

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
	// Create a mock repository
	mockRepo := repository.NewMockRepository()
	router := New(mockRepo)

	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 404 for non-existent routes
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestBatchTelemetryEndpoint(t *testing.T) {
	// Create a mock repository
	mockRepo := repository.NewMockRepository()
	router := New(mockRepo)

	now := time.Now().UTC()

	// Create sample telemetry batch
	batch := []models.TelemetryData{
		{
			ITOW:      118286240,
			Timestamp: now,
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
		{
			ITOW:      118286340,
			Timestamp: now.Add(100 * time.Millisecond),
			GPS: models.GpsData{
				Latitude:           42.6719136,
				Longitude:          23.2887339,
				WgsAltitude:        625.8,
				MslAltitude:        590.1,
				Speed:              125.6,
				Heading:            270.6,
				NumSatellites:      11,
				FixStatus:          3,
				HorizontalAccuracy: 0.925,
				VerticalAccuracy:   1.837,
				SpeedAccuracy:      0.705,
				HeadingAccuracy:    145.27,
				PDOP:               3.0,
				IsFixValid:         true,
			},
			Motion: models.MotionData{
				GForceX:   -0.002,
				GForceY:   0.112,
				GForceZ:   0.975,
				RotationX: 2.10,
				RotationY: 0.87,
				RotationZ: 0.05,
			},
			Battery:       88.9,
			IsCharging:    false,
			TimeAccuracy:  25,
			ValidityFlags: 7,
		},
		{
			ITOW:      118286440,
			Timestamp: now.Add(200 * time.Millisecond),
			GPS: models.GpsData{
				Latitude:  42.6719237,
				Longitude: 23.2887440,
				Speed:     125.7,
			},
			Motion: models.MotionData{
				GForceX: -0.001,
				GForceY: 0.111,
				GForceZ: 0.976,
			},
			Battery:       88.8,
			TimeAccuracy:  25,
			ValidityFlags: 7,
		},
	}

	body, err := json.Marshal(batch)
	if err != nil {
		t.Fatalf("Failed to marshal batch: %v", err)
	}

	req, _ := http.NewRequest("POST", "/api/telemetry/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify response contains success message
	expectedMsg := "Batch telemetry data received successfully (3 records)"
	if msg, ok := response["message"].(string); !ok || msg != expectedMsg {
		t.Errorf("Expected message '%s', got %v", expectedMsg, response["message"])
	}

	// Verify response contains count
	if count, ok := response["count"].(float64); !ok || count != 3 {
		t.Errorf("Expected count 3, got %v", response["count"])
	}

	// Verify response contains IDs array
	if ids, ok := response["ids"].([]interface{}); !ok {
		t.Error("Expected ids array in response")
	} else if len(ids) != 3 {
		t.Errorf("Expected 3 IDs in response, got %d", len(ids))
	}
}

func TestBatchTelemetryEndpointValidation(t *testing.T) {
	mockRepo := repository.NewMockRepository()
	router := New(mockRepo)

	tests := []struct {
		name           string
		batch          interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "empty batch",
			batch:          []models.TelemetryData{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Empty batch",
		},
		{
			name: "missing timestamp",
			batch: []models.TelemetryData{
				{
					ITOW:   118286240,
					GPS:    models.GpsData{},
					Motion: models.MotionData{},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Validation failed for record 0",
		},
		{
			name:           "invalid JSON",
			batch:          "not valid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid JSON payload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.batch.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.batch)
				if err != nil {
					t.Fatalf("Failed to marshal batch: %v", err)
				}
			}

			req, _ := http.NewRequest("POST", "/api/telemetry/batch", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Parse response
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			// Check error message
			if errMsg, ok := response["error"].(string); !ok || errMsg != tt.expectedError {
				t.Errorf("Expected error '%s', got %v", tt.expectedError, response["error"])
			}
		})
	}
}

func TestBatchTelemetryEndpointLargePayload(t *testing.T) {
	mockRepo := repository.NewMockRepository()
	router := New(mockRepo)

	now := time.Now().UTC()

	// Create a batch with exactly 1000 records (should succeed)
	batch := make([]models.TelemetryData, 1000)
	for i := 0; i < 1000; i++ {
		batch[i] = models.TelemetryData{
			ITOW:          int64(118286240 + i),
			Timestamp:     now.Add(time.Duration(i) * time.Millisecond),
			GPS:           models.GpsData{Latitude: 42.0, Longitude: 23.0},
			Motion:        models.MotionData{},
			Battery:       85.0,
			TimeAccuracy:  25,
			ValidityFlags: 7,
		}
	}

	body, _ := json.Marshal(batch)
	req, _ := http.NewRequest("POST", "/api/telemetry/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should succeed with 1000 records
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d for 1000 records, got %d", http.StatusCreated, w.Code)
	}

	// Now test with 1001 records (should fail)
	largeBatch := make([]models.TelemetryData, 1001)
	for i := 0; i < 1001; i++ {
		largeBatch[i] = models.TelemetryData{
			ITOW:          int64(118286240 + i),
			Timestamp:     now.Add(time.Duration(i) * time.Millisecond),
			GPS:           models.GpsData{Latitude: 42.0, Longitude: 23.0},
			Motion:        models.MotionData{},
			Battery:       85.0,
			TimeAccuracy:  25,
			ValidityFlags: 7,
		}
	}

	body, _ = json.Marshal(largeBatch)
	req, _ = http.NewRequest("POST", "/api/telemetry/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should fail with 1001 records
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for 1001 records, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	expectedError := "Batch too large (max 1000 records)"
	if errMsg, ok := response["error"].(string); !ok || errMsg != expectedError {
		t.Errorf("Expected error '%s', got %v", expectedError, response["error"])
	}
}
