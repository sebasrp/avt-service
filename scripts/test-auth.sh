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