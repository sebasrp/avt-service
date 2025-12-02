package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sebasr/avt-service/internal/middleware"
	"github.com/sebasr/avt-service/internal/repository"
)

// DeviceHandler handles device-related requests
type DeviceHandler struct {
	deviceRepo repository.DeviceRepository
}

// NewDeviceHandler creates a new device handler
func NewDeviceHandler(deviceRepo repository.DeviceRepository) *DeviceHandler {
	return &DeviceHandler{
		deviceRepo: deviceRepo,
	}
}

// UpdateDeviceRequest represents the device update request body
type UpdateDeviceRequest struct {
	DeviceName  *string                `json:"deviceName,omitempty"`
	DeviceModel *string                `json:"deviceModel,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// DeviceResponse represents a device in API responses
type DeviceResponse struct {
	ID          string                 `json:"id"`
	DeviceID    string                 `json:"deviceId"`
	DeviceName  *string                `json:"deviceName,omitempty"`
	DeviceModel *string                `json:"deviceModel,omitempty"`
	ClaimedAt   string                 `json:"claimedAt"`
	LastSeenAt  *string                `json:"lastSeenAt,omitempty"`
	IsActive    bool                   `json:"isActive"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   string                 `json:"createdAt"`
	UpdatedAt   string                 `json:"updatedAt"`
}

// ListDevices retrieves all devices for the authenticated user
// GET /api/v1/devices
func (h *DeviceHandler) ListDevices(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	devices, err := h.deviceRepo.ListByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to retrieve devices",
		})
		return
	}

	// Convert to response format
	response := make([]DeviceResponse, len(devices))
	for i, device := range devices {
		var lastSeenAt *string
		if device.LastSeenAt != nil {
			seenStr := device.LastSeenAt.Format("2006-01-02T15:04:05Z07:00")
			lastSeenAt = &seenStr
		}

		response[i] = DeviceResponse{
			ID:          device.ID.String(),
			DeviceID:    device.DeviceID,
			DeviceName:  device.DeviceName,
			DeviceModel: device.DeviceModel,
			ClaimedAt:   device.ClaimedAt.Format("2006-01-02T15:04:05Z07:00"),
			LastSeenAt:  lastSeenAt,
			IsActive:    device.IsActive,
			Metadata:    device.Metadata,
			CreatedAt:   device.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   device.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"devices": response,
		"total":   len(response),
	})
}

// GetDevice retrieves a specific device by ID
// GET /api/v1/devices/:id
func (h *DeviceHandler) GetDevice(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	deviceIDParam := c.Param("id")
	deviceID, err := uuid.Parse(deviceIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_device_id",
			"message": "Invalid device ID format",
		})
		return
	}

	device, err := h.deviceRepo.GetByID(c.Request.Context(), deviceID)
	if err != nil {
		if err == repository.ErrDeviceNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "device_not_found",
				"message": "Device not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to retrieve device",
		})
		return
	}

	// Verify device belongs to user
	if device.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": "You do not have access to this device",
		})
		return
	}

	var lastSeenAt *string
	if device.LastSeenAt != nil {
		seenStr := device.LastSeenAt.Format("2006-01-02T15:04:05Z07:00")
		lastSeenAt = &seenStr
	}

	c.JSON(http.StatusOK, DeviceResponse{
		ID:          device.ID.String(),
		DeviceID:    device.DeviceID,
		DeviceName:  device.DeviceName,
		DeviceModel: device.DeviceModel,
		ClaimedAt:   device.ClaimedAt.Format("2006-01-02T15:04:05Z07:00"),
		LastSeenAt:  lastSeenAt,
		IsActive:    device.IsActive,
		Metadata:    device.Metadata,
		CreatedAt:   device.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   device.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// UpdateDevice updates a device's information
// PATCH /api/v1/devices/:id
func (h *DeviceHandler) UpdateDevice(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	deviceIDParam := c.Param("id")
	deviceID, err := uuid.Parse(deviceIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_device_id",
			"message": "Invalid device ID format",
		})
		return
	}

	var req UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Get current device
	device, err := h.deviceRepo.GetByID(c.Request.Context(), deviceID)
	if err != nil {
		if err == repository.ErrDeviceNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "device_not_found",
				"message": "Device not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to retrieve device",
		})
		return
	}

	// Verify device belongs to user
	if device.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": "You do not have access to this device",
		})
		return
	}

	// Update fields if provided
	if req.DeviceName != nil {
		device.DeviceName = req.DeviceName
	}
	if req.DeviceModel != nil {
		device.DeviceModel = req.DeviceModel
	}
	if req.Metadata != nil {
		device.Metadata = req.Metadata
	}

	// Save updates
	if err := h.deviceRepo.Update(c.Request.Context(), device); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to update device",
		})
		return
	}

	var lastSeenAt *string
	if device.LastSeenAt != nil {
		seenStr := device.LastSeenAt.Format("2006-01-02T15:04:05Z07:00")
		lastSeenAt = &seenStr
	}

	c.JSON(http.StatusOK, DeviceResponse{
		ID:          device.ID.String(),
		DeviceID:    device.DeviceID,
		DeviceName:  device.DeviceName,
		DeviceModel: device.DeviceModel,
		ClaimedAt:   device.ClaimedAt.Format("2006-01-02T15:04:05Z07:00"),
		LastSeenAt:  lastSeenAt,
		IsActive:    device.IsActive,
		Metadata:    device.Metadata,
		CreatedAt:   device.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   device.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// DeactivateDevice deactivates a device
// DELETE /api/v1/devices/:id
func (h *DeviceHandler) DeactivateDevice(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	deviceIDParam := c.Param("id")
	deviceID, err := uuid.Parse(deviceIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_device_id",
			"message": "Invalid device ID format",
		})
		return
	}

	// Get current device
	device, err := h.deviceRepo.GetByID(c.Request.Context(), deviceID)
	if err != nil {
		if err == repository.ErrDeviceNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "device_not_found",
				"message": "Device not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to retrieve device",
		})
		return
	}

	// Verify device belongs to user
	if device.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": "You do not have access to this device",
		})
		return
	}

	// Deactivate device
	device.IsActive = false
	if err := h.deviceRepo.Update(c.Request.Context(), device); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to deactivate device",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Device deactivated successfully",
	})
}
