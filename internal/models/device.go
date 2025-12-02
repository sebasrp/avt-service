package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Device represents a RaceBox device claimed by a user
type Device struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	DeviceID    string                 `json:"deviceId" db:"device_id"`                 // Hardware device ID
	UserID      uuid.UUID              `json:"userId" db:"user_id"`                     // Owner of the device
	DeviceName  *string                `json:"deviceName,omitempty" db:"device_name"`   // User-friendly name
	DeviceModel *string                `json:"deviceModel,omitempty" db:"device_model"` // e.g., "Mini S", "Micro"
	ClaimedAt   time.Time              `json:"claimedAt" db:"claimed_at"`               // When the device was claimed
	LastSeenAt  *time.Time             `json:"lastSeenAt,omitempty" db:"last_seen_at"`  // Last telemetry upload
	IsActive    bool                   `json:"isActive" db:"is_active"`                 // Whether device is active
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`        // Additional device info (JSONB)
	CreatedAt   time.Time              `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time              `json:"updatedAt" db:"updated_at"`
}

// MetadataJSON returns the metadata as a JSON string for database storage
func (d *Device) MetadataJSON() (string, error) {
	if d.Metadata == nil {
		return "{}", nil
	}

	bytes, err := json.Marshal(d.Metadata)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// SetMetadataFromJSON parses JSON string into metadata map
func (d *Device) SetMetadataFromJSON(jsonStr string) error {
	if jsonStr == "" || jsonStr == "{}" {
		d.Metadata = make(map[string]interface{})
		return nil
	}

	return json.Unmarshal([]byte(jsonStr), &d.Metadata)
}

// IsOnline checks if the device has been seen recently (within the last hour)
func (d *Device) IsOnline() bool {
	if d.LastSeenAt == nil {
		return false
	}

	// Consider device online if it was seen within the last hour
	return time.Since(*d.LastSeenAt) < time.Hour
}

// DeviceResponse represents a device for API responses
type DeviceResponse struct {
	ID          uuid.UUID              `json:"id"`
	DeviceID    string                 `json:"deviceId"`
	UserID      uuid.UUID              `json:"userId"`
	DeviceName  *string                `json:"deviceName,omitempty"`
	DeviceModel *string                `json:"deviceModel,omitempty"`
	ClaimedAt   time.Time              `json:"claimedAt"`
	LastSeenAt  *time.Time             `json:"lastSeenAt,omitempty"`
	IsActive    bool                   `json:"isActive"`
	IsOnline    bool                   `json:"isOnline"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

// ToResponse converts a Device to a DeviceResponse
func (d *Device) ToResponse() *DeviceResponse {
	return &DeviceResponse{
		ID:          d.ID,
		DeviceID:    d.DeviceID,
		UserID:      d.UserID,
		DeviceName:  d.DeviceName,
		DeviceModel: d.DeviceModel,
		ClaimedAt:   d.ClaimedAt,
		LastSeenAt:  d.LastSeenAt,
		IsActive:    d.IsActive,
		IsOnline:    d.IsOnline(),
		Metadata:    d.Metadata,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}
