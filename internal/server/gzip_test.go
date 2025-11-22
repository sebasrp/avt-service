package server

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sebasr/avt-service/internal/models"
	"github.com/sebasr/avt-service/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestGzipDecompression(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		compress       bool
		contentType    string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Uncompressed request should work",
			compress:       false,
			contentType:    "application/json",
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "Gzip compressed request should work",
			compress:       true,
			contentType:    "application/json",
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock repository
			mockRepo := repository.NewMockRepository()

			// Create router with gzip middleware
			router := New(mockRepo)

			// Create test telemetry data
			telemetry := models.TelemetryData{
				Timestamp: time.Now().UTC(),
				ITOW:      118286240,
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

			// Marshal to JSON
			jsonData, err := json.Marshal(telemetry)
			assert.NoError(t, err)

			var body []byte
			headers := make(map[string]string)
			headers["Content-Type"] = tt.contentType

			if tt.compress {
				// Compress the data
				var buf bytes.Buffer
				gzipWriter := gzip.NewWriter(&buf)
				_, err := gzipWriter.Write(jsonData)
				assert.NoError(t, err)
				err = gzipWriter.Close()
				assert.NoError(t, err)
				body = buf.Bytes()
				headers["Content-Encoding"] = "gzip"
			} else {
				body = jsonData
			}

			// Create request
			req, err := http.NewRequest("POST", "/api/v1/telemetry", bytes.NewReader(body))
			assert.NoError(t, err)

			for key, value := range headers {
				req.Header.Set(key, value)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectError {
				// Verify response
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Telemetry data received successfully", response["message"])
			}
		})
	}
}

func TestGzipBatchDecompression(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		compress       bool
		batchSize      int
		expectedStatus int
	}{
		{
			name:           "Uncompressed batch should work",
			compress:       false,
			batchSize:      10,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Gzip compressed batch should work",
			compress:       true,
			batchSize:      10,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Large compressed batch should work",
			compress:       true,
			batchSize:      100,
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock repository
			mockRepo := repository.NewMockRepository()

			// Create router with gzip middleware
			router := New(mockRepo)

			// Create batch of telemetry data
			batch := make([]models.TelemetryData, tt.batchSize)
			now := time.Now().UTC()
			for i := 0; i < tt.batchSize; i++ {
				batch[i] = models.TelemetryData{
					Timestamp: now.Add(time.Duration(i) * time.Millisecond),
					ITOW:      int64(118286240 + i*100),
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
			}

			// Marshal to JSON
			jsonData, err := json.Marshal(batch)
			assert.NoError(t, err)

			var body []byte
			headers := make(map[string]string)
			headers["Content-Type"] = "application/json"

			if tt.compress {
				// Compress the data
				var buf bytes.Buffer
				gzipWriter := gzip.NewWriter(&buf)
				_, err := gzipWriter.Write(jsonData)
				assert.NoError(t, err)
				err = gzipWriter.Close()
				assert.NoError(t, err)
				body = buf.Bytes()
				headers["Content-Encoding"] = "gzip"

				// Log compression ratio
				compressionRatio := float64(len(body)) / float64(len(jsonData)) * 100
				t.Logf("Compression ratio: %.2f%% (original: %d bytes, compressed: %d bytes)",
					compressionRatio, len(jsonData), len(body))
			} else {
				body = jsonData
			}

			// Create request
			req, err := http.NewRequest("POST", "/api/v1/telemetry/batch", bytes.NewReader(body))
			assert.NoError(t, err)

			for key, value := range headers {
				req.Header.Set(key, value)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Verify response
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response["message"], "Batch telemetry data received successfully")
			assert.Equal(t, float64(tt.batchSize), response["count"])
		})
	}
}

func TestGzipInvalidData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock repository
	mockRepo := repository.NewMockRepository()

	// Create router with gzip middleware
	router := New(mockRepo)

	// Create invalid gzip data
	invalidGzipData := []byte("this is not valid gzip data")

	// Create request with gzip header but invalid data
	req, err := http.NewRequest("POST", "/api/v1/telemetry", bytes.NewReader(invalidGzipData))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Should return bad request
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
