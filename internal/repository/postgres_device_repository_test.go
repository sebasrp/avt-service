package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sebasr/avt-service/internal/database"
	"github.com/sebasr/avt-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresDeviceRepository_Create(t *testing.T) {
	db, cleanup := setupDeviceTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db.DB)
	userRepo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a test user first
	user := &models.User{
		ID:           uuid.New(),
		Email:        "device@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create a device
	device := &models.Device{
		ID:          uuid.New(),
		DeviceID:    "RACEBOX-001",
		UserID:      user.ID,
		DeviceName:  stringPtr("My RaceBox"),
		DeviceModel: stringPtr("Mini S"),
		ClaimedAt:   time.Now(),
		IsActive:    true,
		Metadata:    map[string]interface{}{"firmware": "1.0.0"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = repo.Create(ctx, device)
	assert.NoError(t, err)

	// Verify device was created
	retrieved, err := repo.GetByID(ctx, device.ID)
	require.NoError(t, err)
	assert.Equal(t, device.ID, retrieved.ID)
	assert.Equal(t, device.DeviceID, retrieved.DeviceID)
	assert.Equal(t, device.UserID, retrieved.UserID)
	assert.Equal(t, *device.DeviceName, *retrieved.DeviceName)
	assert.Equal(t, *device.DeviceModel, *retrieved.DeviceModel)
	assert.Equal(t, device.IsActive, retrieved.IsActive)
	assert.NotNil(t, retrieved.Metadata)
	assert.Equal(t, "1.0.0", retrieved.Metadata["firmware"])
}

func TestPostgresDeviceRepository_Create_DuplicateDeviceID(t *testing.T) {
	db, cleanup := setupDeviceTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db.DB)
	userRepo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "duplicate@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create first device
	device1 := &models.Device{
		ID:        uuid.New(),
		DeviceID:  "RACEBOX-DUP",
		UserID:    user.ID,
		ClaimedAt: time.Now(),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = repo.Create(ctx, device1)
	require.NoError(t, err)

	// Try to create second device with same device_id
	device2 := &models.Device{
		ID:        uuid.New(),
		DeviceID:  "RACEBOX-DUP",
		UserID:    user.ID,
		ClaimedAt: time.Now(),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = repo.Create(ctx, device2)
	assert.ErrorIs(t, err, ErrDeviceExists)
}

func TestPostgresDeviceRepository_GetByID(t *testing.T) {
	db, cleanup := setupDeviceTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db.DB)
	userRepo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "getbyid@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create a device
	device := &models.Device{
		ID:          uuid.New(),
		DeviceID:    "RACEBOX-GET",
		UserID:      user.ID,
		DeviceName:  stringPtr("Test Device"),
		DeviceModel: stringPtr("Micro"),
		ClaimedAt:   time.Now(),
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = repo.Create(ctx, device)
	require.NoError(t, err)

	// Retrieve by ID
	retrieved, err := repo.GetByID(ctx, device.ID)
	require.NoError(t, err)
	assert.Equal(t, device.ID, retrieved.ID)
	assert.Equal(t, device.DeviceID, retrieved.DeviceID)
	assert.Equal(t, *device.DeviceName, *retrieved.DeviceName)
}

func TestPostgresDeviceRepository_GetByID_NotFound(t *testing.T) {
	db, cleanup := setupDeviceTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db.DB)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, uuid.New())
	assert.ErrorIs(t, err, ErrDeviceNotFound)
}

func TestPostgresDeviceRepository_GetByDeviceID(t *testing.T) {
	db, cleanup := setupDeviceTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db.DB)
	userRepo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "getbydeviceid@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create a device
	device := &models.Device{
		ID:        uuid.New(),
		DeviceID:  "RACEBOX-FIND",
		UserID:    user.ID,
		ClaimedAt: time.Now(),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = repo.Create(ctx, device)
	require.NoError(t, err)

	// Retrieve by device ID
	retrieved, err := repo.GetByDeviceID(ctx, device.DeviceID)
	require.NoError(t, err)
	assert.Equal(t, device.ID, retrieved.ID)
	assert.Equal(t, device.DeviceID, retrieved.DeviceID)
}

func TestPostgresDeviceRepository_GetByDeviceID_NotFound(t *testing.T) {
	db, cleanup := setupDeviceTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db.DB)
	ctx := context.Background()

	_, err := repo.GetByDeviceID(ctx, "NON-EXISTENT")
	assert.ErrorIs(t, err, ErrDeviceNotFound)
}

func TestPostgresDeviceRepository_ListByUserID(t *testing.T) {
	db, cleanup := setupDeviceTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db.DB)
	userRepo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create test users
	user1 := &models.User{
		ID:           uuid.New(),
		Email:        "list1@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	user2 := &models.User{
		ID:           uuid.New(),
		Email:        "list2@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user1)
	require.NoError(t, err)
	err = userRepo.Create(ctx, user2)
	require.NoError(t, err)

	// Create devices for user1
	device1 := &models.Device{
		ID:        uuid.New(),
		DeviceID:  "RACEBOX-U1-1",
		UserID:    user1.ID,
		ClaimedAt: time.Now().Add(-2 * time.Hour),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	device2 := &models.Device{
		ID:        uuid.New(),
		DeviceID:  "RACEBOX-U1-2",
		UserID:    user1.ID,
		ClaimedAt: time.Now().Add(-1 * time.Hour),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	// Create device for user2
	device3 := &models.Device{
		ID:        uuid.New(),
		DeviceID:  "RACEBOX-U2-1",
		UserID:    user2.ID,
		ClaimedAt: time.Now(),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, device1)
	require.NoError(t, err)
	err = repo.Create(ctx, device2)
	require.NoError(t, err)
	err = repo.Create(ctx, device3)
	require.NoError(t, err)

	// List devices for user1
	devices, err := repo.ListByUserID(ctx, user1.ID)
	require.NoError(t, err)
	assert.Len(t, devices, 2)

	// Verify order (most recent first)
	assert.Equal(t, device2.DeviceID, devices[0].DeviceID)
	assert.Equal(t, device1.DeviceID, devices[1].DeviceID)

	// List devices for user2
	devices, err = repo.ListByUserID(ctx, user2.ID)
	require.NoError(t, err)
	assert.Len(t, devices, 1)
	assert.Equal(t, device3.DeviceID, devices[0].DeviceID)
}

func TestPostgresDeviceRepository_ListByUserID_Empty(t *testing.T) {
	db, cleanup := setupDeviceTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db.DB)
	ctx := context.Background()

	devices, err := repo.ListByUserID(ctx, uuid.New())
	assert.NoError(t, err)
	assert.Empty(t, devices)
}

func TestPostgresDeviceRepository_Update(t *testing.T) {
	db, cleanup := setupDeviceTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db.DB)
	userRepo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "update@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create a device
	device := &models.Device{
		ID:        uuid.New(),
		DeviceID:  "RACEBOX-UPD",
		UserID:    user.ID,
		ClaimedAt: time.Now(),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = repo.Create(ctx, device)
	require.NoError(t, err)

	// Update the device
	device.DeviceName = stringPtr("Updated Name")
	device.DeviceModel = stringPtr("Mini S Pro")
	device.IsActive = false
	device.Metadata = map[string]interface{}{"version": "2.0"}

	err = repo.Update(ctx, device)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(ctx, device.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", *retrieved.DeviceName)
	assert.Equal(t, "Mini S Pro", *retrieved.DeviceModel)
	assert.False(t, retrieved.IsActive)
	assert.Equal(t, "2.0", retrieved.Metadata["version"])
}

func TestPostgresDeviceRepository_Update_NotFound(t *testing.T) {
	db, cleanup := setupDeviceTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db.DB)
	ctx := context.Background()

	device := &models.Device{
		ID:        uuid.New(),
		DeviceID:  "RACEBOX-NONE",
		UserID:    uuid.New(),
		ClaimedAt: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Update(ctx, device)
	assert.ErrorIs(t, err, ErrDeviceNotFound)
}

func TestPostgresDeviceRepository_UpdateLastSeen(t *testing.T) {
	db, cleanup := setupDeviceTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db.DB)
	userRepo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// Create a test user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "lastseen@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create a device without last_seen_at
	device := &models.Device{
		ID:        uuid.New(),
		DeviceID:  "RACEBOX-SEEN",
		UserID:    user.ID,
		ClaimedAt: time.Now(),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = repo.Create(ctx, device)
	require.NoError(t, err)

	// Wait a moment to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Update last seen
	err = repo.UpdateLastSeen(ctx, device.DeviceID)
	assert.NoError(t, err)

	// Verify last_seen_at was updated
	retrieved, err := repo.GetByID(ctx, device.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved.LastSeenAt)
	assert.True(t, retrieved.LastSeenAt.After(device.CreatedAt))
}

func TestPostgresDeviceRepository_UpdateLastSeen_NotFound(t *testing.T) {
	db, cleanup := setupDeviceTestDB(t)
	defer cleanup()

	repo := NewPostgresDeviceRepository(db.DB)
	ctx := context.Background()

	err := repo.UpdateLastSeen(ctx, "NON-EXISTENT")
	assert.ErrorIs(t, err, ErrDeviceNotFound)
}

// setupDeviceTestDB creates a test database with the necessary tables
func setupDeviceTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()
	return setupTestDB(t)
}

// stringPtr is a helper to create string pointers
func stringPtr(s string) *string {
	return &s
}
