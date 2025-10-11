package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sebasr/avt-service/internal/models"
)

// TelemetryHandler handles incoming telemetry data from RaceBox devices
func TelemetryHandler(c *gin.Context) {
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

	// Log the telemetry data to console
	logTelemetry(telemetry)

	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"message":   "Telemetry data received successfully",
		"timestamp": telemetry.Timestamp,
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
