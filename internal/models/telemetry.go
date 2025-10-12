// Package models contains data models for the AVT service.
package models

import "time"

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
