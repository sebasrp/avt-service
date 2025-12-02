package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDevice_IsOnline(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	deviceID := "DEVICE-001"

	tests := []struct {
		name     string
		device   *Device
		expected bool
	}{
		{
			name: "online - last seen 30 minutes ago",
			device: &Device{
				ID:         uuid.New(),
				DeviceID:   deviceID,
				UserID:     userID,
				LastSeenAt: func() *time.Time { t := now.Add(-30 * time.Minute); return &t }(),
				IsActive:   true,
			},
			expected: true,
		},
		{
			name: "online - last seen 59 minutes ago",
			device: &Device{
				ID:         uuid.New(),
				DeviceID:   deviceID,
				UserID:     userID,
				LastSeenAt: func() *time.Time { t := now.Add(-59 * time.Minute); return &t }(),
				IsActive:   true,
			},
			expected: true,
		},
		{
			name: "offline - last seen 2 hours ago",
			device: &Device{
				ID:         uuid.New(),
				DeviceID:   deviceID,
				UserID:     userID,
				LastSeenAt: func() *time.Time { t := now.Add(-2 * time.Hour); return &t }(),
				IsActive:   true,
			},
			expected: false,
		},
		{
			name: "offline - last seen 61 minutes ago",
			device: &Device{
				ID:         uuid.New(),
				DeviceID:   deviceID,
				UserID:     userID,
				LastSeenAt: func() *time.Time { t := now.Add(-61 * time.Minute); return &t }(),
				IsActive:   true,
			},
			expected: false,
		},
		{
			name: "offline - never seen",
			device: &Device{
				ID:         uuid.New(),
				DeviceID:   deviceID,
				UserID:     userID,
				LastSeenAt: nil,
				IsActive:   true,
			},
			expected: false,
		},
		{
			name: "online - last seen just now",
			device: &Device{
				ID:         uuid.New(),
				DeviceID:   deviceID,
				UserID:     userID,
				LastSeenAt: &now,
				IsActive:   true,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.device.IsOnline()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDevice_MetadataJSON(t *testing.T) {
	tests := []struct {
		name        string
		metadata    map[string]interface{}
		expectedErr bool
	}{
		{
			name: "simple metadata",
			metadata: map[string]interface{}{
				"firmware": "1.2.3",
				"model":    "Mini S",
			},
			expectedErr: false,
		},
		{
			name: "nested metadata",
			metadata: map[string]interface{}{
				"firmware": "1.2.3",
				"settings": map[string]interface{}{
					"rate": 10,
					"mode": "performance",
				},
			},
			expectedErr: false,
		},
		{
			name:        "nil metadata",
			metadata:    nil,
			expectedErr: false,
		},
		{
			name:        "empty metadata",
			metadata:    map[string]interface{}{},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &Device{
				ID:       uuid.New(),
				DeviceID: "TEST-001",
				UserID:   uuid.New(),
				Metadata: tt.metadata,
			}

			jsonStr, err := device.MetadataJSON()

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, jsonStr)

				switch {
				case tt.metadata == nil:
					assert.Equal(t, "{}", jsonStr)
				case len(tt.metadata) == 0:
					assert.Equal(t, "{}", jsonStr)
				default:
					// Should be valid JSON
					assert.Contains(t, jsonStr, "{")
					assert.Contains(t, jsonStr, "}")
				}
			}
		})
	}
}

func TestDevice_SetMetadataFromJSON(t *testing.T) {
	tests := []struct {
		name        string
		jsonStr     string
		expectedErr bool
		validate    func(*testing.T, *Device)
	}{
		{
			name:        "simple JSON",
			jsonStr:     `{"firmware":"1.2.3","model":"Mini S"}`,
			expectedErr: false,
			validate: func(t *testing.T, d *Device) {
				assert.Equal(t, "1.2.3", d.Metadata["firmware"])
				assert.Equal(t, "Mini S", d.Metadata["model"])
			},
		},
		{
			name:        "nested JSON",
			jsonStr:     `{"firmware":"1.2.3","settings":{"rate":10,"mode":"performance"}}`,
			expectedErr: false,
			validate: func(t *testing.T, d *Device) {
				assert.Equal(t, "1.2.3", d.Metadata["firmware"])
				settings, ok := d.Metadata["settings"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, float64(10), settings["rate"]) // JSON numbers are float64
				assert.Equal(t, "performance", settings["mode"])
			},
		},
		{
			name:        "empty JSON object",
			jsonStr:     "{}",
			expectedErr: false,
			validate: func(t *testing.T, d *Device) {
				assert.NotNil(t, d.Metadata)
				assert.Len(t, d.Metadata, 0)
			},
		},
		{
			name:        "empty string",
			jsonStr:     "",
			expectedErr: false,
			validate: func(t *testing.T, d *Device) {
				assert.NotNil(t, d.Metadata)
				assert.Len(t, d.Metadata, 0)
			},
		},
		{
			name:        "invalid JSON",
			jsonStr:     `{"firmware":"1.2.3"`,
			expectedErr: true,
			validate:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &Device{
				ID:       uuid.New(),
				DeviceID: "TEST-001",
				UserID:   uuid.New(),
			}

			err := device.SetMetadataFromJSON(tt.jsonStr)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, device)
				}
			}
		})
	}
}

func TestDevice_MetadataRoundTrip(t *testing.T) {
	device := &Device{
		ID:       uuid.New(),
		DeviceID: "TEST-001",
		UserID:   uuid.New(),
		Metadata: map[string]interface{}{
			"firmware": "1.2.3",
			"model":    "Mini S",
			"settings": map[string]interface{}{
				"rate": 10,
				"mode": "performance",
			},
		},
	}

	// Convert to JSON
	jsonStr, err := device.MetadataJSON()
	require.NoError(t, err)

	// Create new device and set metadata from JSON
	newDevice := &Device{
		ID:       uuid.New(),
		DeviceID: "TEST-002",
		UserID:   uuid.New(),
	}
	err = newDevice.SetMetadataFromJSON(jsonStr)
	require.NoError(t, err)

	// Verify the metadata matches
	assert.Equal(t, device.Metadata["firmware"], newDevice.Metadata["firmware"])
	assert.Equal(t, device.Metadata["model"], newDevice.Metadata["model"])
}

func TestDevice_ToResponse(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	deviceUUID := uuid.New()
	lastSeen := now.Add(-30 * time.Minute)
	deviceName := "My RaceBox"
	deviceModel := "Mini S"

	device := &Device{
		ID:          deviceUUID,
		DeviceID:    "DEVICE-001",
		UserID:      userID,
		DeviceName:  &deviceName,
		DeviceModel: &deviceModel,
		ClaimedAt:   now.Add(-7 * 24 * time.Hour),
		LastSeenAt:  &lastSeen,
		IsActive:    true,
		Metadata: map[string]interface{}{
			"firmware": "1.2.3",
		},
		CreatedAt: now.Add(-7 * 24 * time.Hour),
		UpdatedAt: now,
	}

	response := device.ToResponse()

	assert.Equal(t, deviceUUID, response.ID)
	assert.Equal(t, "DEVICE-001", response.DeviceID)
	assert.Equal(t, userID, response.UserID)
	assert.Equal(t, &deviceName, response.DeviceName)
	assert.Equal(t, &deviceModel, response.DeviceModel)
	assert.Equal(t, device.ClaimedAt, response.ClaimedAt)
	assert.Equal(t, &lastSeen, response.LastSeenAt)
	assert.True(t, response.IsActive)
	assert.True(t, response.IsOnline) // Should be online - last seen 30 minutes ago
	assert.Equal(t, "1.2.3", response.Metadata["firmware"])
	assert.Equal(t, device.CreatedAt, response.CreatedAt)
	assert.Equal(t, device.UpdatedAt, response.UpdatedAt)
}

func TestDevice_ToResponse_Offline(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	lastSeen := now.Add(-2 * time.Hour) // 2 hours ago - should be offline

	device := &Device{
		ID:         uuid.New(),
		DeviceID:   "DEVICE-001",
		UserID:     userID,
		LastSeenAt: &lastSeen,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	response := device.ToResponse()

	assert.False(t, response.IsOnline) // Should be offline
}
