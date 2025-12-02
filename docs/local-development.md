# Local Development Guide

This guide provides instructions for working with the AVT Service authentication system during local development.

## Table of Contents

- [Quick Start](#quick-start)
- [Creating Test Users](#creating-test-users)
- [Getting Test Tokens](#getting-test-tokens)
- [Testing Protected Endpoints](#testing-protected-endpoints)
- [Debugging Auth Issues](#debugging-auth-issues)
- [Common Scenarios](#common-scenarios)

---

## Quick Start

### 1. Start the Development Environment

```bash
# Start database
make docker-up

# Run migrations
make migrate

# Start server with authentication
JWT_SECRET="dev-secret-key-change-in-production" make run
```

### 2. Verify Server is Running

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-10T08:51:08Z",
  "database": "connected"
}
```

---

## Creating Test Users

### Using curl

```bash
# Register a new test user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "testpass123"
  }'
```

Expected response:
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "test@example.com",
    "emailVerified": false
  },
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expiresAt": "2024-01-10T09:51:08Z"
}
```

### Using the Database Directly

For testing purposes, you can also create users directly in the database:

```bash
# Connect to database
make db-shell

# Create a user with hashed password
INSERT INTO users (email, password_hash, email_verified, is_active)
VALUES (
  'admin@example.com',
  '$2a$10$YourBcryptHashHere',  -- Use bcrypt to hash "password123"
  true,
  true
);
```

To generate a bcrypt hash in Go:

```go
package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 10)
    fmt.Println(string(hash))
}
```

---

## Getting Test Tokens

### Login to Get Tokens

```bash
# Login with test user
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "testpass123"
  }'
```

Save the tokens from the response:
```bash
# Save to environment variables for easy reuse
export ACCESS_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
export REFRESH_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Refresh Access Token

When your access token expires (after 1 hour by default):

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refreshToken\": \"$REFRESH_TOKEN\"}"
```

---

## Testing Protected Endpoints

### Access User Profile

```bash
curl http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Update User Profile

```bash
curl -X PATCH http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "displayName": "Test User",
    "timezone": "America/New_York",
    "unitsPreference": "imperial"
  }'
```

### List User Devices

```bash
curl http://localhost:8080/api/v1/devices \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Upload Telemetry with Authentication

```bash
curl -X POST http://localhost:8080/api/v1/telemetry \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "deviceId": "device-001",
    "iTOW": 118286240,
    "timestamp": "2024-01-10T08:51:08.239Z",
    "gps": {
      "latitude": 42.6719035,
      "longitude": 23.2887238,
      "wgsAltitude": 625.761,
      "mslAltitude": 590.095,
      "speed": 125.5,
      "heading": 270.5,
      "numSatellites": 11,
      "fixStatus": 3,
      "horizontalAccuracy": 0.924,
      "verticalAccuracy": 1.836,
      "speedAccuracy": 0.704,
      "headingAccuracy": 145.26856,
      "pdop": 3.0,
      "isFixValid": true
    },
    "motion": {
      "gForceX": -0.003,
      "gForceY": 0.113,
      "gForceZ": 0.974,
      "rotationX": 2.09,
      "rotationY": 0.86,
      "rotationZ": 0.04
    },
    "battery": 89.0,
    "isCharging": false,
    "timeAccuracy": 25,
    "validityFlags": 7
  }'
```

---

## Debugging Auth Issues

### Common Issues and Solutions

#### 1. "unauthorized" Error

**Problem:** Getting 401 Unauthorized response

**Possible Causes:**
- Access token expired (tokens expire after 1 hour)
- Invalid token format
- Wrong JWT secret
- Token not included in Authorization header

**Solutions:**

```bash
# Check if token is expired by decoding it (use jwt.io)
echo $ACCESS_TOKEN | cut -d'.' -f2 | base64 -d

# Refresh the token
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refreshToken\": \"$REFRESH_TOKEN\"}"

# Verify Authorization header format
curl -v http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $ACCESS_TOKEN"
# Should see: Authorization: Bearer eyJ...
```

#### 2. "invalid or expired token" Error

**Problem:** Token validation fails

**Solutions:**

```bash
# Verify JWT_SECRET matches between server restarts
echo $JWT_SECRET

# Login again to get fresh tokens
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "testpass123"
  }'
```

#### 3. "email already registered" Error

**Problem:** Trying to register with existing email

**Solutions:**

```bash
# Use a different email
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test2@example.com",
    "password": "testpass123"
  }'

# Or delete the existing user from database
make db-shell
DELETE FROM users WHERE email = 'test@example.com';
```

#### 4. Database Connection Issues

**Problem:** Cannot connect to database

**Solutions:**

```bash
# Check if database is running
docker ps | grep timescale

# Restart database
make docker-down
make docker-up

# Wait for database to be ready
sleep 5

# Run migrations
make migrate
```

### Enable Debug Logging

Add verbose logging to see what's happening:

```bash
# Run server with debug output
GIN_MODE=debug JWT_SECRET="dev-secret-key" go run cmd/server/main.go
```

### Inspect Database State

```bash
# Connect to database
make db-shell

# Check users
SELECT id, email, email_verified, created_at FROM users;

# Check refresh tokens
SELECT id, user_id, expires_at, revoked_at FROM refresh_tokens;

# Check devices
SELECT id, device_id, user_id, device_name, claimed_at FROM devices;
```

---

## Common Scenarios

### Scenario 1: Testing Device Claiming

```bash
# 1. Register and login
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "pass123"}' \
  | jq -r '.accessToken' > /tmp/token.txt

export ACCESS_TOKEN=$(cat /tmp/token.txt)

# 2. Upload telemetry with device ID (device will be auto-claimed)
curl -X POST http://localhost:8080/api/v1/telemetry \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "deviceId": "my-test-device",
    "iTOW": 118286240,
    "timestamp": "2024-01-10T08:51:08.239Z",
    "gps": {...},
    "motion": {...}
  }'

# 3. Verify device was claimed
curl http://localhost:8080/api/v1/devices \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Scenario 2: Testing Multiple Users

```bash
# Create two users
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "user1@example.com", "password": "pass123"}' \
  | jq -r '.accessToken' > /tmp/user1_token.txt

curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "user2@example.com", "password": "pass123"}' \
  | jq -r '.accessToken' > /tmp/user2_token.txt

# Each user uploads telemetry with their own device
curl -X POST http://localhost:8080/api/v1/telemetry \
  -H "Authorization: Bearer $(cat /tmp/user1_token.txt)" \
  -H "Content-Type: application/json" \
  -d '{"deviceId": "device-user1", ...}'

curl -X POST http://localhost:8080/api/v1/telemetry \
  -H "Authorization: Bearer $(cat /tmp/user2_token.txt)" \
  -H "Content-Type: application/json" \
  -d '{"deviceId": "device-user2", ...}'

# Verify each user only sees their own devices
curl http://localhost:8080/api/v1/devices \
  -H "Authorization: Bearer $(cat /tmp/user1_token.txt)"
```

### Scenario 3: Testing Token Refresh Flow

```bash
# 1. Login and save tokens
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "testpass123"}')

ACCESS_TOKEN=$(echo $RESPONSE | jq -r '.accessToken')
REFRESH_TOKEN=$(echo $RESPONSE | jq -r '.refreshToken')

# 2. Use access token
curl http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 3. Wait for token to expire (or manually expire it in database)
# For testing, you can set JWT_ACCESS_TOKEN_TTL=10s

# 4. Try to use expired token (should fail)
curl http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 5. Refresh to get new tokens
NEW_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refreshToken\": \"$REFRESH_TOKEN\"}")

NEW_ACCESS_TOKEN=$(echo $NEW_RESPONSE | jq -r '.accessToken')
NEW_REFRESH_TOKEN=$(echo $NEW_RESPONSE | jq -r '.refreshToken')

# 6. Use new access token
curl http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $NEW_ACCESS_TOKEN"
```

### Scenario 4: Testing Logout

```bash
# 1. Login
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "testpass123"}')

ACCESS_TOKEN=$(echo $RESPONSE | jq -r '.accessToken')
REFRESH_TOKEN=$(echo $RESPONSE | jq -r '.refreshToken')

# 2. Logout (revokes all refresh tokens)
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 3. Try to refresh (should fail)
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refreshToken\": \"$REFRESH_TOKEN\"}"
# Expected: {"error": "unauthorized", "details": "invalid or expired token"}
```

---

## Testing Scripts

### Complete Test Script

Save this as `scripts/test-auth.sh`:

```bash
#!/bin/bash

set -e

BASE_URL="http://localhost:8080"
EMAIL="test-$(date +%s)@example.com"
PASSWORD="testpass123"

echo "=== Testing Authentication Flow ==="
echo

# 1. Register
echo "1. Registering user: $EMAIL"
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL\", \"password\": \"$PASSWORD\"}")

ACCESS_TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.accessToken')
REFRESH_TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.refreshToken')
USER_ID=$(echo $REGISTER_RESPONSE | jq -r '.user.id')

echo "✓ User registered: $USER_ID"
echo

# 2. Get profile
echo "2. Getting user profile"
curl -s "$BASE_URL/api/v1/users/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq
echo "✓ Profile retrieved"
echo

# 3. Update profile
echo "3. Updating profile"
curl -s -X PATCH "$BASE_URL/api/v1/users/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"displayName": "Test User", "timezone": "UTC"}' | jq
echo "✓ Profile updated"
echo

# 4. Upload telemetry (claims device)
echo "4. Uploading telemetry"
curl -s -X POST "$BASE_URL/api/v1/telemetry" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "deviceId": "test-device-001",
    "iTOW": 118286240,
    "timestamp": "2024-01-10T08:51:08.239Z",
    "gps": {
      "latitude": 42.6719035,
      "longitude": 23.2887238,
      "wgsAltitude": 625.761,
      "mslAltitude": 590.095,
      "speed": 125.5,
      "heading": 270.5,
      "numSatellites": 11,
      "fixStatus": 3,
      "horizontalAccuracy": 0.924,
      "verticalAccuracy": 1.836,
      "speedAccuracy": 0.704,
      "headingAccuracy": 145.26856,
      "pdop": 3.0,
      "isFixValid": true
    },
    "motion": {
      "gForceX": -0.003,
      "gForceY": 0.113,
      "gForceZ": 0.974,
      "rotationX": 2.09,
      "rotationY": 0.86,
      "rotationZ": 0.04
    },
    "battery": 89.0,
    "isCharging": false,
    "timeAccuracy": 25,
    "validityFlags": 7
  }' | jq
echo "✓ Telemetry uploaded"
echo

# 5. List devices
echo "5. Listing devices"
curl -s "$BASE_URL/api/v1/devices" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq
echo "✓ Devices listed"
echo

# 6. Refresh token
echo "6. Refreshing token"
REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refreshToken\": \"$REFRESH_TOKEN\"}")

NEW_ACCESS_TOKEN=$(echo $REFRESH_RESPONSE | jq -r '.accessToken')
echo "✓ Token refreshed"
echo

# 7. Logout
echo "7. Logging out"
curl -s -X POST "$BASE_URL/api/v1/auth/logout" \
  -H "Authorization: Bearer $NEW_ACCESS_TOKEN" | jq
echo "✓ Logged out"
echo

echo "=== All tests passed! ==="
```

Make it executable:
```bash
chmod +x scripts/test-auth.sh
./scripts/test-auth.sh
```

---

## Environment Variables for Testing

Create a `.env.local` file for local development:

```bash
# Server
PORT=8080

# Database
DATABASE_URL=postgres://telemetry_user:telemetry_pass@localhost:5432/telemetry_dev?sslmode=disable

# Authentication
JWT_SECRET=dev-secret-key-change-in-production
JWT_ACCESS_TOKEN_TTL=1h
JWT_REFRESH_TOKEN_TTL=720h

# For testing token expiration
# JWT_ACCESS_TOKEN_TTL=30s
# JWT_REFRESH_TOKEN_TTL=5m
```

Load it before running:
```bash
export $(cat .env.local | xargs)
go run cmd/server/main.go
```

---

## Tips and Best Practices

1. **Use unique emails for testing**: Add timestamps to avoid conflicts
   ```bash
   EMAIL="test-$(date +%s)@example.com"
   ```

2. **Save tokens to files**: Makes it easier to reuse them
   ```bash
   echo $ACCESS_TOKEN > /tmp/access_token.txt
   export ACCESS_TOKEN=$(cat /tmp/access_token.txt)
   ```

3. **Use jq for JSON parsing**: Makes responses more readable
   ```bash
   curl ... | jq
   ```

4. **Test with short token TTLs**: Set `JWT_ACCESS_TOKEN_TTL=30s` to test expiration quickly

5. **Clean up test data**: Regularly clean test users from database
   ```sql
   DELETE FROM users WHERE email LIKE 'test-%@example.com';
   ```

6. **Use Postman/Insomnia**: For interactive testing with saved requests

7. **Check server logs**: Run with `GIN_MODE=debug` to see detailed logs

---

## Next Steps

- See [authentication-architecture.md](authentication-architecture.md) for system design
- See [authentication-implementation-checklist.md](authentication-implementation-checklist.md) for implementation details
- See [testing.md](testing.md) for automated testing documentation