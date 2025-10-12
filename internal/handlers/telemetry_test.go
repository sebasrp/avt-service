package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sebasr/avt-service/internal/models"
	"github.com/sebasr/avt-service/internal/repository"
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
			// Create mock repository
			mockRepo := repository.NewMockRepository()
			handler := NewTelemetryHandler(mockRepo)

			// Create test router
			router := gin.New()
			router.POST("/api/telemetry", handler.HandlePost)

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
	// Create mock repository
	mockRepo := repository.NewMockRepository()
	handler := NewTelemetryHandler(mockRepo)

	router := gin.New()
	router.POST("/api/telemetry", handler.HandlePost)

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

func TestTelemetryHandlerDatabaseError(t *testing.T) {
	// Create mock repository that returns an error
	mockRepo := repository.NewMockRepository()
	mockRepo.SaveFunc = func(_ context.Context, _ *models.TelemetryData) error {
		return errors.New("database connection failed")
	}

	handler := NewTelemetryHandler(mockRepo)

	router := gin.New()
	router.POST("/api/telemetry", handler.HandlePost)

	telemetry := models.TelemetryData{
		Timestamp: time.Now().UTC(),
		GPS: models.GpsData{
			Latitude:  42.6719035,
			Longitude: 23.2887238,
		},
		Motion: models.MotionData{},
	}

	body, _ := json.Marshal(telemetry)
	req, _ := http.NewRequest("POST", "/api/telemetry", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	// Check response body
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if err, ok := response["error"].(string); !ok || err != "Failed to save telemetry data" {
		t.Errorf("Expected database error message, got %v", response["error"])
	}
}

func TestTelemetryHandler_BatchPost(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name: "valid batch with multiple records",
			payload: []models.TelemetryData{
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
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if msg, ok := resp["message"].(string); !ok || msg != "Batch telemetry data received successfully (2 records)" {
					t.Errorf("Expected batch success message, got %v", resp["message"])
				}
				if count, ok := resp["count"].(float64); !ok || count != 2 {
					t.Errorf("Expected count 2, got %v", resp["count"])
				}
				if ids, ok := resp["ids"].([]interface{}); !ok || len(ids) != 2 {
					t.Errorf("Expected 2 IDs in response, got %v", resp["ids"])
				}
			},
		},
		{
			name: "single record batch",
			payload: []models.TelemetryData{
				{
					ITOW:          118286240,
					Timestamp:     now,
					GPS:           models.GpsData{Latitude: 42.0, Longitude: 23.0},
					Motion:        models.MotionData{},
					Battery:       85.0,
					TimeAccuracy:  25,
					ValidityFlags: 7,
				},
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if count, ok := resp["count"].(float64); !ok || count != 1 {
					t.Errorf("Expected count 1, got %v", resp["count"])
				}
				if ids, ok := resp["ids"].([]interface{}); !ok || len(ids) != 1 {
					t.Errorf("Expected 1 ID in response, got %v", resp["ids"])
				}
			},
		},
		{
			name:           "empty batch",
			payload:        []models.TelemetryData{},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if err, ok := resp["error"].(string); !ok || err != "Empty batch" {
					t.Errorf("Expected empty batch error, got %v", resp["error"])
				}
			},
		},
		{
			name:           "invalid JSON",
			payload:        "not valid json",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if err, ok := resp["error"].(string); !ok || err != "Invalid JSON payload" {
					t.Errorf("Expected invalid JSON error, got %v", resp["error"])
				}
			},
		},
		{
			name: "missing timestamp in one record",
			payload: []models.TelemetryData{
				{
					ITOW:      118286240,
					Timestamp: now,
					GPS:       models.GpsData{},
					Motion:    models.MotionData{},
				},
				{
					ITOW:   118286340,
					GPS:    models.GpsData{},
					Motion: models.MotionData{},
					// Missing Timestamp
				},
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if err, ok := resp["error"].(string); !ok || err != "Missing timestamp in record 1" {
					t.Errorf("Expected missing timestamp error, got %v", resp["error"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock repository
			mockRepo := repository.NewMockRepository()
			handler := NewTelemetryHandler(mockRepo)

			// Create test router
			router := gin.New()
			router.POST("/api/telemetry/batch", handler.HandleBatchPost)

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
			req, err := http.NewRequest("POST", "/api/telemetry/batch", bytes.NewBuffer(body))
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
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
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

func TestTelemetryHandler_BatchPostTooLarge(t *testing.T) {
	// Create a batch with more than 1000 records
	now := time.Now().UTC()
	largeBatch := make([]models.TelemetryData, 1001)
	for i := 0; i < 1001; i++ {
		largeBatch[i] = models.TelemetryData{
			ITOW:      int64(118286240 + i),
			Timestamp: now.Add(time.Duration(i) * time.Millisecond),
			GPS:       models.GpsData{Latitude: 42.0, Longitude: 23.0},
			Motion:    models.MotionData{},
		}
	}

	mockRepo := repository.NewMockRepository()
	handler := NewTelemetryHandler(mockRepo)

	router := gin.New()
	router.POST("/api/telemetry/batch", handler.HandleBatchPost)

	body, _ := json.Marshal(largeBatch)
	req, _ := http.NewRequest("POST", "/api/telemetry/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	// Check response body
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if err, ok := response["error"].(string); !ok || err != "Batch too large (max 1000 records)" {
		t.Errorf("Expected batch too large error, got %v", response["error"])
	}
}

func TestTelemetryHandler_BatchPostDatabaseError(t *testing.T) {
	now := time.Now().UTC()
	batch := []models.TelemetryData{
		{
			ITOW:      118286240,
			Timestamp: now,
			GPS:       models.GpsData{Latitude: 42.0, Longitude: 23.0},
			Motion:    models.MotionData{},
		},
	}

	// Create mock repository that returns an error
	mockRepo := repository.NewMockRepository()
	mockRepo.SaveBatchFunc = func(_ context.Context, _ []*models.TelemetryData) error {
		return errors.New("database connection failed")
	}

	handler := NewTelemetryHandler(mockRepo)

	router := gin.New()
	router.POST("/api/telemetry/batch", handler.HandleBatchPost)

	body, _ := json.Marshal(batch)
	req, _ := http.NewRequest("POST", "/api/telemetry/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	// Check response body
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if err, ok := response["error"].(string); !ok || err != "Failed to save telemetry batch" {
		t.Errorf("Expected database error message, got %v", response["error"])
	}
}

func TestTelemetryHandler_BatchPostContentType(t *testing.T) {
	now := time.Now().UTC()
	batch := []models.TelemetryData{
		{
			ITOW:      118286240,
			Timestamp: now,
			GPS:       models.GpsData{},
			Motion:    models.MotionData{},
		},
	}

	mockRepo := repository.NewMockRepository()
	handler := NewTelemetryHandler(mockRepo)

	router := gin.New()
	router.POST("/api/telemetry/batch", handler.HandleBatchPost)

	body, _ := json.Marshal(batch)
	req, _ := http.NewRequest("POST", "/api/telemetry/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type 'application/json; charset=utf-8', got %q", contentType)
	}
}

func TestTelemetryHandler_BatchPostWithSessionID(t *testing.T) {
	now := time.Now().UTC()
	sessionID := "test-session-123"

	batch := []models.TelemetryData{
		{
			ITOW:      118286240,
			Timestamp: now,
			SessionID: &sessionID,
			GPS:       models.GpsData{},
			Motion:    models.MotionData{},
		},
	}

	mockRepo := repository.NewMockRepository()
	handler := NewTelemetryHandler(mockRepo)

	router := gin.New()
	router.POST("/api/telemetry/batch", handler.HandleBatchPost)

	body, _ := json.Marshal(batch)
	req, _ := http.NewRequest("POST", "/api/telemetry/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Check response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if count, ok := response["count"].(float64); !ok || count != 1 {
		t.Errorf("Expected count 1, got %v", response["count"])
	}
}
