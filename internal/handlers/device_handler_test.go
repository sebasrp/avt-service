package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sebasr/avt-service/internal/middleware"
	"github.com/sebasr/avt-service/internal/models"
	"github.com/sebasr/avt-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDeviceTest() (*DeviceHandler, *repository.MockDeviceRepository) {
	deviceRepo := repository.NewMockDeviceRepository()
	handler := NewDeviceHandler(deviceRepo)

	gin.SetMode(gin.TestMode)

	return handler, deviceRepo
}

func TestDeviceHandler_ListDevices_Success(t *testing.T) {
	handler, deviceRepo := setupDeviceTest()

	userID := uuid.New()
	lastSeen := time.Now().Add(-1 * time.Hour)
	deviceName := "My RaceBox"
	deviceModel := "Mini S"

	devices := []*models.Device{
		{
			ID:          uuid.New(),
			DeviceID:    "RACEBOX-001",
			UserID:      userID,
			DeviceName:  &deviceName,
			DeviceModel: &deviceModel,
			ClaimedAt:   time.Now().Add(-30 * 24 * time.Hour),
			LastSeenAt:  &lastSeen,
			IsActive:    true,
			Metadata:    map[string]interface{}{"firmware": "1.0.0"},
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
		},
		{
			ID:        uuid.New(),
			DeviceID:  "RACEBOX-002",
			UserID:    userID,
			ClaimedAt: time.Now().Add(-10 * 24 * time.Hour),
			IsActive:  true,
			CreatedAt: time.Now().Add(-10 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-10 * 24 * time.Hour),
		},
	}

	deviceRepo.ListByUserIDFunc = func(_ context.Context, id uuid.UUID) ([]*models.Device, error) {
		if id == userID {
			return devices, nil
		}
		return []*models.Device{}, nil
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/devices", nil)
	c.Set(string(middleware.UserIDKey), userID)

	handler.ListDevices(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(2), response["total"])
	devicesArr := response["devices"].([]interface{})
	assert.Len(t, devicesArr, 2)
}

func TestDeviceHandler_ListDevices_Empty(t *testing.T) {
	handler, deviceRepo := setupDeviceTest()

	userID := uuid.New()

	deviceRepo.ListByUserIDFunc = func(_ context.Context, _ uuid.UUID) ([]*models.Device, error) {
		return []*models.Device{}, nil
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/devices", nil)
	c.Set(string(middleware.UserIDKey), userID)

	handler.ListDevices(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(0), response["total"])
}

func TestDeviceHandler_GetDevice_Success(t *testing.T) {
	handler, deviceRepo := setupDeviceTest()

	userID := uuid.New()
	deviceID := uuid.New()
	deviceName := "My Device"

	device := &models.Device{
		ID:         deviceID,
		DeviceID:   "RACEBOX-001",
		UserID:     userID,
		DeviceName: &deviceName,
		ClaimedAt:  time.Now().Add(-10 * 24 * time.Hour),
		IsActive:   true,
		CreatedAt:  time.Now().Add(-10 * 24 * time.Hour),
		UpdatedAt:  time.Now().Add(-1 * time.Hour),
	}

	deviceRepo.GetByIDFunc = func(_ context.Context, id uuid.UUID) (*models.Device, error) {
		if id == deviceID {
			return device, nil
		}
		return nil, repository.ErrDeviceNotFound
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/devices/"+deviceID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: deviceID.String()}}
	c.Set(string(middleware.UserIDKey), userID)

	handler.GetDevice(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response DeviceResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, deviceID.String(), response.ID)
	assert.Equal(t, "RACEBOX-001", response.DeviceID)
	assert.Equal(t, deviceName, *response.DeviceName)
}

func TestDeviceHandler_GetDevice_NotFound(t *testing.T) {
	handler, deviceRepo := setupDeviceTest()

	userID := uuid.New()
	deviceID := uuid.New()

	deviceRepo.GetByIDFunc = func(_ context.Context, _ uuid.UUID) (*models.Device, error) {
		return nil, repository.ErrDeviceNotFound
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/devices/"+deviceID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: deviceID.String()}}
	c.Set(string(middleware.UserIDKey), userID)

	handler.GetDevice(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "device_not_found")
}

func TestDeviceHandler_GetDevice_Forbidden(t *testing.T) {
	handler, deviceRepo := setupDeviceTest()

	userID := uuid.New()
	otherUserID := uuid.New()
	deviceID := uuid.New()

	device := &models.Device{
		ID:        deviceID,
		DeviceID:  "RACEBOX-001",
		UserID:    otherUserID, // Different user
		ClaimedAt: time.Now(),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	deviceRepo.GetByIDFunc = func(_ context.Context, id uuid.UUID) (*models.Device, error) {
		if id == deviceID {
			return device, nil
		}
		return nil, repository.ErrDeviceNotFound
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/devices/"+deviceID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: deviceID.String()}}
	c.Set(string(middleware.UserIDKey), userID)

	handler.GetDevice(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "forbidden")
}

func TestDeviceHandler_GetDevice_InvalidID(t *testing.T) {
	handler, _ := setupDeviceTest()

	userID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/devices/invalid-id", nil)
	c.Params = gin.Params{{Key: "id", Value: "invalid-id"}}
	c.Set(string(middleware.UserIDKey), userID)

	handler.GetDevice(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_device_id")
}

func TestDeviceHandler_UpdateDevice_Success(t *testing.T) {
	handler, deviceRepo := setupDeviceTest()

	userID := uuid.New()
	deviceID := uuid.New()

	device := &models.Device{
		ID:        deviceID,
		DeviceID:  "RACEBOX-001",
		UserID:    userID,
		ClaimedAt: time.Now().Add(-10 * 24 * time.Hour),
		IsActive:  true,
		CreatedAt: time.Now().Add(-10 * 24 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	deviceRepo.GetByIDFunc = func(_ context.Context, id uuid.UUID) (*models.Device, error) {
		if id == deviceID {
			return device, nil
		}
		return nil, repository.ErrDeviceNotFound
	}

	var updatedDevice *models.Device
	deviceRepo.UpdateFunc = func(_ context.Context, d *models.Device) error {
		updatedDevice = d
		return nil
	}

	newName := "Updated Name"
	newModel := "Mini S Pro"
	reqBody := UpdateDeviceRequest{
		DeviceName:  &newName,
		DeviceModel: &newModel,
		Metadata:    map[string]interface{}{"version": "2.0"},
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/devices/"+deviceID.String(), bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: deviceID.String()}}
	c.Set(string(middleware.UserIDKey), userID)

	handler.UpdateDevice(c)

	assert.Equal(t, http.StatusOK, w.Code)

	require.NotNil(t, updatedDevice)
	assert.Equal(t, newName, *updatedDevice.DeviceName)
	assert.Equal(t, newModel, *updatedDevice.DeviceModel)
	assert.Equal(t, "2.0", updatedDevice.Metadata["version"])
}

func TestDeviceHandler_UpdateDevice_NotFound(t *testing.T) {
	handler, deviceRepo := setupDeviceTest()

	userID := uuid.New()
	deviceID := uuid.New()

	deviceRepo.GetByIDFunc = func(_ context.Context, _ uuid.UUID) (*models.Device, error) {
		return nil, repository.ErrDeviceNotFound
	}

	newName := "Updated Name"
	reqBody := UpdateDeviceRequest{
		DeviceName: &newName,
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/devices/"+deviceID.String(), bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: deviceID.String()}}
	c.Set(string(middleware.UserIDKey), userID)

	handler.UpdateDevice(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "device_not_found")
}

func TestDeviceHandler_UpdateDevice_Forbidden(t *testing.T) {
	handler, deviceRepo := setupDeviceTest()

	userID := uuid.New()
	otherUserID := uuid.New()
	deviceID := uuid.New()

	device := &models.Device{
		ID:        deviceID,
		DeviceID:  "RACEBOX-001",
		UserID:    otherUserID,
		ClaimedAt: time.Now(),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	deviceRepo.GetByIDFunc = func(_ context.Context, id uuid.UUID) (*models.Device, error) {
		if id == deviceID {
			return device, nil
		}
		return nil, repository.ErrDeviceNotFound
	}

	newName := "Updated Name"
	reqBody := UpdateDeviceRequest{
		DeviceName: &newName,
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/devices/"+deviceID.String(), bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: deviceID.String()}}
	c.Set(string(middleware.UserIDKey), userID)

	handler.UpdateDevice(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "forbidden")
}

func TestDeviceHandler_DeactivateDevice_Success(t *testing.T) {
	handler, deviceRepo := setupDeviceTest()

	userID := uuid.New()
	deviceID := uuid.New()

	device := &models.Device{
		ID:        deviceID,
		DeviceID:  "RACEBOX-001",
		UserID:    userID,
		ClaimedAt: time.Now().Add(-10 * 24 * time.Hour),
		IsActive:  true,
		CreatedAt: time.Now().Add(-10 * 24 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	deviceRepo.GetByIDFunc = func(_ context.Context, id uuid.UUID) (*models.Device, error) {
		if id == deviceID {
			return device, nil
		}
		return nil, repository.ErrDeviceNotFound
	}

	var updatedDevice *models.Device
	deviceRepo.UpdateFunc = func(_ context.Context, d *models.Device) error {
		updatedDevice = d
		return nil
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/devices/"+deviceID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: deviceID.String()}}
	c.Set(string(middleware.UserIDKey), userID)

	handler.DeactivateDevice(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Device deactivated successfully")

	require.NotNil(t, updatedDevice)
	assert.False(t, updatedDevice.IsActive)
}

func TestDeviceHandler_DeactivateDevice_NotFound(t *testing.T) {
	handler, deviceRepo := setupDeviceTest()

	userID := uuid.New()
	deviceID := uuid.New()

	deviceRepo.GetByIDFunc = func(_ context.Context, _ uuid.UUID) (*models.Device, error) {
		return nil, repository.ErrDeviceNotFound
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/devices/"+deviceID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: deviceID.String()}}
	c.Set(string(middleware.UserIDKey), userID)

	handler.DeactivateDevice(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "device_not_found")
}

func TestDeviceHandler_DeactivateDevice_Forbidden(t *testing.T) {
	handler, deviceRepo := setupDeviceTest()

	userID := uuid.New()
	otherUserID := uuid.New()
	deviceID := uuid.New()

	device := &models.Device{
		ID:        deviceID,
		DeviceID:  "RACEBOX-001",
		UserID:    otherUserID,
		ClaimedAt: time.Now(),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	deviceRepo.GetByIDFunc = func(_ context.Context, id uuid.UUID) (*models.Device, error) {
		if id == deviceID {
			return device, nil
		}
		return nil, repository.ErrDeviceNotFound
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/devices/"+deviceID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: deviceID.String()}}
	c.Set(string(middleware.UserIDKey), userID)

	handler.DeactivateDevice(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "forbidden")
}
