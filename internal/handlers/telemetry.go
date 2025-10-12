// Package handlers contains HTTP request handlers for the AVT service.
package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sebasr/avt-service/internal/models"
	"github.com/sebasr/avt-service/internal/repository"
)

// TelemetryHandler handles telemetry-related HTTP requests
type TelemetryHandler struct {
	repo repository.TelemetryRepository
}

// NewTelemetryHandler creates a new telemetry handler with the given repository
func NewTelemetryHandler(repo repository.TelemetryRepository) *TelemetryHandler {
	return &TelemetryHandler{repo: repo}
}

// HandlePost handles incoming telemetry data from RaceBox devices
func (h *TelemetryHandler) HandlePost(c *gin.Context) {
	var telemetry models.TelemetryData

	// Parse JSON body
	if err := c.ShouldBindJSON(&telemetry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON payload",
		})
		return
	}

	// Validate required fields
	if telemetry.Timestamp.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required field: timestamp",
		})
		return
	}

	// Save to database
	if err := h.repo.Save(c.Request.Context(), &telemetry); err != nil {
		log.Printf("Error saving telemetry to database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save telemetry data",
		})
		return
	}

	// Log the telemetry data to console
	logTelemetry(telemetry)

	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"message":   "Telemetry data received successfully",
		"timestamp": telemetry.Timestamp,
		"id":        telemetry.ID,
	})
}

// HandleBatchPost handles incoming batch telemetry data from RaceBox devices
func (h *TelemetryHandler) HandleBatchPost(c *gin.Context) {
	var telemetryBatch []models.TelemetryData

	// Parse JSON body
	if err := c.ShouldBindJSON(&telemetryBatch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON payload",
			"details": err.Error(),
		})
		return
	}

	// Validate batch size
	if len(telemetryBatch) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Empty batch",
		})
		return
	}

	if len(telemetryBatch) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Batch too large (max 1000 records)",
		})
		return
	}

	// Validate each telemetry record
	for i, telemetry := range telemetryBatch {
		if telemetry.Timestamp.IsZero() {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Missing timestamp in record %d", i),
			})
			return
		}
	}

	// Convert to pointers for SaveBatch
	telemetryPointers := make([]*models.TelemetryData, len(telemetryBatch))
	for i := range telemetryBatch {
		telemetryPointers[i] = &telemetryBatch[i]
	}

	// Save batch to database
	if err := h.repo.SaveBatch(c.Request.Context(), telemetryPointers); err != nil {
		log.Printf("Error saving telemetry batch to database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save telemetry batch",
		})
		return
	}

	// Collect IDs of saved records
	savedIDs := make([]int64, len(telemetryBatch))
	for i, telemetry := range telemetryBatch {
		savedIDs[i] = telemetry.ID
		// Log first and last records only to avoid spam
		if i == 0 || i == len(telemetryBatch)-1 {
			logTelemetry(telemetry)
		}
	}

	log.Printf("Batch telemetry: Saved %d records", len(telemetryBatch))

	// Return success response with IDs
	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("Batch telemetry data received successfully (%d records)", len(telemetryBatch)),
		"count":   len(telemetryBatch),
		"ids":     savedIDs,
	})
}

// logTelemetry logs telemetry data in a structured format
func logTelemetry(data models.TelemetryData) {
	log.Printf("=== Telemetry Data Received ===")
	log.Printf("Timestamp: %s", data.Timestamp.Format("2006-01-02 15:04:05.000"))
	log.Printf("iTOW: %d ms", data.ITOW)
	log.Printf("Battery: %.1f%% (charging: %v)", data.Battery, data.IsCharging)

	// Log GPS data
	log.Printf("GPS:")
	log.Printf("  Position: %.7f°, %.7f°", data.GPS.Latitude, data.GPS.Longitude)
	log.Printf("  Altitude: %.1f m (MSL: %.1f m)", data.GPS.WgsAltitude, data.GPS.MslAltitude)
	log.Printf("  Speed: %.1f km/h, Heading: %.1f°", data.GPS.Speed, data.GPS.Heading)
	log.Printf("  Satellites: %d, Fix: %d (%s)",
		data.GPS.NumSatellites,
		data.GPS.FixStatus,
		fixStatusString(data.GPS.FixStatus, data.GPS.IsFixValid))
	log.Printf("  Accuracy: H=%.2fm, V=%.2fm, Speed=%.2fkm/h",
		data.GPS.HorizontalAccuracy,
		data.GPS.VerticalAccuracy,
		data.GPS.SpeedAccuracy)
	log.Printf("  PDOP: %.2f", data.GPS.PDOP)

	// Log Motion data
	log.Printf("Motion:")
	log.Printf("  G-Forces: X=%.3fg, Y=%.3fg, Z=%.3fg",
		data.Motion.GForceX,
		data.Motion.GForceY,
		data.Motion.GForceZ)
	log.Printf("  Rotation: X=%.2f°/s, Y=%.2f°/s, Z=%.2f°/s",
		data.Motion.RotationX,
		data.Motion.RotationY,
		data.Motion.RotationZ)
	log.Printf("==============================")
}

// fixStatusString converts fix status code to human-readable string
func fixStatusString(status int, isValid bool) string {
	if !isValid {
		return "Invalid"
	}
	switch status {
	case 0:
		return "No Fix"
	case 2:
		return "2D Fix"
	case 3:
		return "3D Fix"
	default:
		return fmt.Sprintf("Unknown (%d)", status)
	}
}
