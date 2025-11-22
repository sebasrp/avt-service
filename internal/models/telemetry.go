// Package models contains data models for the AVT service.
package models

import (
	"fmt"
	"time"
)

// TelemetryData represents complete telemetry data from a RaceBox device
type TelemetryData struct {
	// Database ID
	ID int64 `json:"id,omitempty" db:"id"`

	// UTC timestamp
	Timestamp time.Time `json:"timestamp" db:"recorded_at"`

	// Device and session identifiers
	DeviceID  string  `json:"deviceId,omitempty" db:"device_id"`
	SessionID *string `json:"sessionId,omitempty" db:"session_id"`

	// GPS time of week in milliseconds
	ITOW int64 `json:"iTOW" db:"itow"`

	// GPS data
	GPS GpsData `json:"gps"`

	// Motion data
	Motion MotionData `json:"motion"`

	// Battery level (0-100%) or input voltage for Micro (in volts)
	Battery float64 `json:"battery" db:"battery"`

	// Whether device is charging (Mini/Mini S only)
	IsCharging bool `json:"isCharging" db:"is_charging"`

	// Time accuracy in nanoseconds
	TimeAccuracy int64 `json:"timeAccuracy" db:"time_accuracy"`

	// Validity flags
	ValidityFlags int `json:"validityFlags" db:"validity_flags"`
}

// GpsData represents GPS data from the RaceBox device
type GpsData struct {
	// Latitude in degrees
	Latitude float64 `json:"latitude" db:"latitude"`

	// Longitude in degrees
	Longitude float64 `json:"longitude" db:"longitude"`

	// WGS altitude in meters
	WgsAltitude float64 `json:"wgsAltitude" db:"wgs_altitude"`

	// MSL altitude in meters
	MslAltitude float64 `json:"mslAltitude" db:"msl_altitude"`

	// Speed in km/h
	Speed float64 `json:"speed" db:"speed"`

	// Heading in degrees (0-360, where 0 is North)
	Heading float64 `json:"heading" db:"heading"`

	// Number of satellites used in solution
	NumSatellites int `json:"numSatellites" db:"num_satellites"`

	// Fix status (0: no fix, 2: 2D fix, 3: 3D fix)
	FixStatus int `json:"fixStatus" db:"fix_status"`

	// Horizontal accuracy in meters
	HorizontalAccuracy float64 `json:"horizontalAccuracy" db:"horizontal_accuracy"`

	// Vertical accuracy in meters
	VerticalAccuracy float64 `json:"verticalAccuracy" db:"vertical_accuracy"`

	// Speed accuracy in km/h
	SpeedAccuracy float64 `json:"speedAccuracy" db:"speed_accuracy"`

	// Heading accuracy in degrees
	HeadingAccuracy float64 `json:"headingAccuracy" db:"heading_accuracy"`

	// PDOP (Position Dilution of Precision)
	PDOP float64 `json:"pdop" db:"pdop"`

	// Whether the fix is valid
	IsFixValid bool `json:"isFixValid" db:"is_fix_valid"`
}

// MotionData represents motion sensor data from the RaceBox device
type MotionData struct {
	// G-force on X axis (front/back)
	GForceX float64 `json:"gForceX" db:"g_force_x"`

	// G-force on Y axis (right/left)
	GForceY float64 `json:"gForceY" db:"g_force_y"`

	// G-force on Z axis (up/down)
	GForceZ float64 `json:"gForceZ" db:"g_force_z"`

	// Rotation rate on X axis (roll) in degrees per second
	RotationX float64 `json:"rotationX" db:"rotation_x"`

	// Rotation rate on Y axis (pitch) in degrees per second
	RotationY float64 `json:"rotationY" db:"rotation_y"`

	// Rotation rate on Z axis (yaw) in degrees per second
	RotationZ float64 `json:"rotationZ" db:"rotation_z"`
}

// Validate validates the telemetry data for correctness
func (t *TelemetryData) Validate() error {
	// Validate timestamp
	if t.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}

	// Validate GPS data
	if err := t.GPS.Validate(); err != nil {
		return fmt.Errorf("GPS validation failed: %w", err)
	}

	// Validate Motion data
	if err := t.Motion.Validate(); err != nil {
		return fmt.Errorf("motion validation failed: %w", err)
	}

	// Validate battery level (0-100% for percentage, or 0-30V for voltage)
	if t.Battery < 0 || t.Battery > 100 {
		// Allow higher values for voltage readings (up to 30V)
		if t.Battery > 30 {
			return fmt.Errorf("invalid battery value: %.2f (must be 0-100%% or 0-30V)", t.Battery)
		}
	}

	return nil
}

// Validate validates GPS data for correctness
func (g *GpsData) Validate() error {
	// Validate latitude range
	if g.Latitude < -90 || g.Latitude > 90 {
		return fmt.Errorf("invalid latitude: %.7f (must be between -90 and 90)", g.Latitude)
	}

	// Validate longitude range
	if g.Longitude < -180 || g.Longitude > 180 {
		return fmt.Errorf("invalid longitude: %.7f (must be between -180 and 180)", g.Longitude)
	}

	// Validate speed (reasonable maximum: 500 km/h for racing)
	if g.Speed < 0 || g.Speed > 500 {
		return fmt.Errorf("invalid speed: %.2f km/h (must be between 0 and 500)", g.Speed)
	}

	// Validate heading range
	if g.Heading < 0 || g.Heading > 360 {
		return fmt.Errorf("invalid heading: %.2f degrees (must be between 0 and 360)", g.Heading)
	}

	// Validate altitude (reasonable range: -500m to 9000m)
	if g.WgsAltitude < -500 || g.WgsAltitude > 9000 {
		return fmt.Errorf("invalid WGS altitude: %.2f m (must be between -500 and 9000)", g.WgsAltitude)
	}

	if g.MslAltitude < -500 || g.MslAltitude > 9000 {
		return fmt.Errorf("invalid MSL altitude: %.2f m (must be between -500 and 9000)", g.MslAltitude)
	}

	// Validate number of satellites
	if g.NumSatellites < 0 || g.NumSatellites > 50 {
		return fmt.Errorf("invalid number of satellites: %d (must be between 0 and 50)", g.NumSatellites)
	}

	// Validate fix status
	if g.FixStatus < 0 || g.FixStatus > 3 {
		return fmt.Errorf("invalid fix status: %d (must be 0, 2, or 3)", g.FixStatus)
	}

	// Validate accuracy values (must be non-negative)
	if g.HorizontalAccuracy < 0 {
		return fmt.Errorf("invalid horizontal accuracy: %.2f (must be non-negative)", g.HorizontalAccuracy)
	}

	if g.VerticalAccuracy < 0 {
		return fmt.Errorf("invalid vertical accuracy: %.2f (must be non-negative)", g.VerticalAccuracy)
	}

	if g.SpeedAccuracy < 0 {
		return fmt.Errorf("invalid speed accuracy: %.2f (must be non-negative)", g.SpeedAccuracy)
	}

	if g.HeadingAccuracy < 0 || g.HeadingAccuracy > 360 {
		return fmt.Errorf("invalid heading accuracy: %.2f (must be between 0 and 360)", g.HeadingAccuracy)
	}

	// Validate PDOP (reasonable range: 0-50)
	if g.PDOP < 0 || g.PDOP > 50 {
		return fmt.Errorf("invalid PDOP: %.2f (must be between 0 and 50)", g.PDOP)
	}

	return nil
}

// Validate validates motion sensor data for correctness
func (m *MotionData) Validate() error {
	// Validate G-forces (reasonable range: -10g to +10g for racing)
	if m.GForceX < -10 || m.GForceX > 10 {
		return fmt.Errorf("invalid G-force X: %.3f (must be between -10 and 10)", m.GForceX)
	}

	if m.GForceY < -10 || m.GForceY > 10 {
		return fmt.Errorf("invalid G-force Y: %.3f (must be between -10 and 10)", m.GForceY)
	}

	if m.GForceZ < -10 || m.GForceZ > 10 {
		return fmt.Errorf("invalid G-force Z: %.3f (must be between -10 and 10)", m.GForceZ)
	}

	// Validate rotation rates (reasonable range: -360 to +360 degrees/second)
	if m.RotationX < -360 || m.RotationX > 360 {
		return fmt.Errorf("invalid rotation X: %.2f deg/s (must be between -360 and 360)", m.RotationX)
	}

	if m.RotationY < -360 || m.RotationY > 360 {
		return fmt.Errorf("invalid rotation Y: %.2f deg/s (must be between -360 and 360)", m.RotationY)
	}

	if m.RotationZ < -360 || m.RotationZ > 360 {
		return fmt.Errorf("invalid rotation Z: %.2f deg/s (must be between -360 and 360)", m.RotationZ)
	}

	return nil
}
