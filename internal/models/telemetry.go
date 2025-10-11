// Package models contains data models for the AVT service.
package models

import "time"

// TelemetryData represents complete telemetry data from a RaceBox device
type TelemetryData struct {
	// GPS time of week in milliseconds
	ITOW int64 `json:"iTOW"`

	// UTC timestamp
	Timestamp time.Time `json:"timestamp"`

	// GPS data
	GPS GpsData `json:"gps"`

	// Motion data
	Motion MotionData `json:"motion"`

	// Battery level (0-100%) or input voltage for Micro (in volts)
	Battery float64 `json:"battery"`

	// Whether device is charging (Mini/Mini S only)
	IsCharging bool `json:"isCharging"`

	// Time accuracy in nanoseconds
	TimeAccuracy int64 `json:"timeAccuracy"`

	// Validity flags
	ValidityFlags int `json:"validityFlags"`
}

// GpsData represents GPS data from the RaceBox device
type GpsData struct {
	// Latitude in degrees
	Latitude float64 `json:"latitude"`

	// Longitude in degrees
	Longitude float64 `json:"longitude"`

	// WGS altitude in meters
	WgsAltitude float64 `json:"wgsAltitude"`

	// MSL altitude in meters
	MslAltitude float64 `json:"mslAltitude"`

	// Speed in km/h
	Speed float64 `json:"speed"`

	// Heading in degrees (0-360, where 0 is North)
	Heading float64 `json:"heading"`

	// Number of satellites used in solution
	NumSatellites int `json:"numSatellites"`

	// Fix status (0: no fix, 2: 2D fix, 3: 3D fix)
	FixStatus int `json:"fixStatus"`

	// Horizontal accuracy in meters
	HorizontalAccuracy float64 `json:"horizontalAccuracy"`

	// Vertical accuracy in meters
	VerticalAccuracy float64 `json:"verticalAccuracy"`

	// Speed accuracy in km/h
	SpeedAccuracy float64 `json:"speedAccuracy"`

	// Heading accuracy in degrees
	HeadingAccuracy float64 `json:"headingAccuracy"`

	// PDOP (Position Dilution of Precision)
	PDOP float64 `json:"pdop"`

	// Whether the fix is valid
	IsFixValid bool `json:"isFixValid"`
}

// MotionData represents motion sensor data from the RaceBox device
type MotionData struct {
	// G-force on X axis (front/back)
	GForceX float64 `json:"gForceX"`

	// G-force on Y axis (right/left)
	GForceY float64 `json:"gForceY"`

	// G-force on Z axis (up/down)
	GForceZ float64 `json:"gForceZ"`

	// Rotation rate on X axis (roll) in degrees per second
	RotationX float64 `json:"rotationX"`

	// Rotation rate on Y axis (pitch) in degrees per second
	RotationY float64 `json:"rotationY"`

	// Rotation rate on Z axis (yaw) in degrees per second
	RotationZ float64 `json:"rotationZ"`
}
