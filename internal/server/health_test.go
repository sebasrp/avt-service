package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthEndpoint(t *testing.T) {
	// Setup
	deps := newTestDeps()
	router := New(deps)

	t.Run("returns 200 OK", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns correct JSON structure", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check required fields exist
		assert.Contains(t, response, "status")
		assert.Contains(t, response, "timestamp")
		assert.Contains(t, response, "version")

		// Check field values
		assert.Equal(t, "healthy", response["status"])
		assert.Equal(t, "1.0.0", response["version"])
	})

	t.Run("returns valid RFC3339 timestamp", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		w := httptest.NewRecorder()

		beforeRequest := time.Now().UTC()
		router.ServeHTTP(w, req)
		afterRequest := time.Now().UTC()

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		timestampStr, ok := response["timestamp"].(string)
		require.True(t, ok, "timestamp should be a string")

		// Parse the timestamp
		timestamp, err := time.Parse(time.RFC3339, timestampStr)
		require.NoError(t, err, "timestamp should be valid RFC3339 format")

		// Verify timestamp is within reasonable range (within test execution time)
		assert.True(t, timestamp.After(beforeRequest.Add(-time.Second)), "timestamp should be after request start")
		assert.True(t, timestamp.Before(afterRequest.Add(time.Second)), "timestamp should be before request end")
	})

	t.Run("responds quickly for latency measurement", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		w := httptest.NewRecorder()

		start := time.Now()
		router.ServeHTTP(w, req)
		duration := time.Since(start)

		// Health check should respond in less than 100ms
		assert.Less(t, duration.Milliseconds(), int64(100),
			"health check should respond quickly for accurate latency measurement")
	})

	t.Run("includes request ID in response headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		requestID := w.Header().Get("X-Request-ID")
		assert.NotEmpty(t, requestID, "should include X-Request-ID header")
	})

	t.Run("accepts custom request ID", func(t *testing.T) {
		customID := "test-request-id-123"
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		req.Header.Set("X-Request-ID", customID)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		requestID := w.Header().Get("X-Request-ID")
		assert.Equal(t, customID, requestID, "should preserve custom request ID")
	})

	t.Run("handles concurrent requests", func(t *testing.T) {
		const numRequests = 10 // Reduced to stay within rate limit
		results := make(chan int, numRequests)

		// Use unique IP for this test to avoid rate limiting
		testIP := "192.0.3.1:12345"

		for i := 0; i < numRequests; i++ {
			go func() {
				req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
				req.RemoteAddr = testIP
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				results <- w.Code
			}()
		}

		// Collect all results
		for i := 0; i < numRequests; i++ {
			statusCode := <-results
			assert.Equal(t, http.StatusOK, statusCode)
		}
	})

	t.Run("does not accept POST method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/health", nil)
		req.RemoteAddr = "192.0.4.1:12345" // Unique IP to avoid rate limiting
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("does not accept PUT method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/health", nil)
		req.RemoteAddr = "192.0.5.1:12345" // Unique IP to avoid rate limiting
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("does not accept DELETE method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/health", nil)
		req.RemoteAddr = "192.0.6.1:12345" // Unique IP to avoid rate limiting
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHealthEndpointNetworkQualityMeasurement(t *testing.T) {
	deps := newTestDeps()
	router := New(deps)

	t.Run("simulates network quality detection workflow", func(t *testing.T) {
		// Simulate multiple pings to measure latency
		const numPings = 5
		latencies := make([]time.Duration, numPings)

		for i := 0; i < numPings; i++ {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
			w := httptest.NewRecorder()

			start := time.Now()
			router.ServeHTTP(w, req)
			latencies[i] = time.Since(start)

			assert.Equal(t, http.StatusOK, w.Code)
		}

		// Calculate average latency
		var totalLatency time.Duration
		for _, latency := range latencies {
			totalLatency += latency
		}
		avgLatency := totalLatency / time.Duration(numPings)

		t.Logf("Average latency over %d pings: %v", numPings, avgLatency)

		// Verify all latencies are reasonable
		for i, latency := range latencies {
			assert.Less(t, latency.Milliseconds(), int64(100),
				"ping %d latency should be under 100ms, got %v", i+1, latency)
		}
	})
}
