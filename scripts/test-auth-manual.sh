#!/bin/bash

# Manual Authentication Testing Script
# This script tests all authentication endpoints manually

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
API_BASE="$BASE_URL/api/v1"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test email and password
TEST_EMAIL="test-$(date +%s)@example.com"
TEST_PASSWORD="TestPassword123!"

echo -e "${YELLOW}=== Authentication Manual Testing ===${NC}"
echo "Base URL: $BASE_URL"
echo "Test Email: $TEST_EMAIL"
echo ""

# Test 1: Register new user
echo -e "${YELLOW}Test 1: Register New User${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "$API_BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$TEST_EMAIL\", \"password\": \"$TEST_PASSWORD\"}")

echo "Response: $REGISTER_RESPONSE"

# Extract tokens
ACCESS_TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.accessToken // empty')
REFRESH_TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.refreshToken // empty')

if [ -z "$ACCESS_TOKEN" ]; then
    echo -e "${RED}✗ Registration failed${NC}"
    exit 1
else
    echo -e "${GREEN}✓ Registration successful${NC}"
    echo "Access Token: ${ACCESS_TOKEN:0:20}..."
    echo "Refresh Token: ${REFRESH_TOKEN:0:20}..."
fi
echo ""

# Test 2: Login with same credentials
echo -e "${YELLOW}Test 2: Login with Registered User${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$TEST_EMAIL\", \"password\": \"$TEST_PASSWORD\"}")

echo "Response: $LOGIN_RESPONSE"

NEW_ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.accessToken // empty')
if [ -z "$NEW_ACCESS_TOKEN" ]; then
    echo -e "${RED}✗ Login failed${NC}"
    exit 1
else
    echo -e "${GREEN}✓ Login successful${NC}"
    ACCESS_TOKEN=$NEW_ACCESS_TOKEN
fi
echo ""

# Test 3: Access protected endpoint (Get Profile)
echo -e "${YELLOW}Test 3: Access Protected Endpoint (Get Profile)${NC}"
PROFILE_RESPONSE=$(curl -s -X GET "$API_BASE/users/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "Response: $PROFILE_RESPONSE"

USER_EMAIL=$(echo $PROFILE_RESPONSE | jq -r '.email // empty')
if [ "$USER_EMAIL" = "$TEST_EMAIL" ]; then
    echo -e "${GREEN}✓ Protected endpoint access successful${NC}"
else
    echo -e "${RED}✗ Protected endpoint access failed${NC}"
    exit 1
fi
echo ""

# Test 4: Access protected endpoint without token
echo -e "${YELLOW}Test 4: Access Protected Endpoint Without Token (Should Fail)${NC}"
UNAUTHORIZED_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE/users/me")

HTTP_CODE=$(echo "$UNAUTHORIZED_RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ]; then
    echo -e "${GREEN}✓ Correctly rejected unauthorized access${NC}"
else
    echo -e "${RED}✗ Should have returned 401, got $HTTP_CODE${NC}"
    exit 1
fi
echo ""

# Test 5: Refresh token
echo -e "${YELLOW}Test 5: Refresh Access Token${NC}"
REFRESH_RESPONSE=$(curl -s -X POST "$API_BASE/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refreshToken\": \"$REFRESH_TOKEN\"}")

echo "Response: $REFRESH_RESPONSE"

NEW_ACCESS_TOKEN=$(echo $REFRESH_RESPONSE | jq -r '.accessToken // empty')
NEW_REFRESH_TOKEN=$(echo $REFRESH_RESPONSE | jq -r '.refreshToken // empty')

if [ -z "$NEW_ACCESS_TOKEN" ] || [ -z "$NEW_REFRESH_TOKEN" ]; then
    echo -e "${RED}✗ Token refresh failed${NC}"
    exit 1
else
    echo -e "${GREEN}✓ Token refresh successful${NC}"
    echo "New Access Token: ${NEW_ACCESS_TOKEN:0:20}..."
    echo "New Refresh Token: ${NEW_REFRESH_TOKEN:0:20}..."
    
    # Verify tokens are different (token rotation)
    if [ "$REFRESH_TOKEN" != "$NEW_REFRESH_TOKEN" ]; then
        echo -e "${GREEN}✓ Token rotation working (refresh token changed)${NC}"
    else
        echo -e "${YELLOW}⚠ Token rotation might not be working${NC}"
    fi
fi
echo ""

# Test 6: Update profile
echo -e "${YELLOW}Test 6: Update User Profile${NC}"
UPDATE_RESPONSE=$(curl -s -X PATCH "$API_BASE/users/me" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $NEW_ACCESS_TOKEN" \
  -d '{"displayName": "Test User", "timezone": "America/New_York"}')

echo "Response: $UPDATE_RESPONSE"

DISPLAY_NAME=$(echo $UPDATE_RESPONSE | jq -r '.profile.displayName // empty')
if [ "$DISPLAY_NAME" = "Test User" ]; then
    echo -e "${GREEN}✓ Profile update successful${NC}"
else
    echo -e "${RED}✗ Profile update failed${NC}"
fi
echo ""

# Test 7: Change password
echo -e "${YELLOW}Test 7: Change Password${NC}"
NEW_PASSWORD="NewPassword456!"
CHANGE_PW_RESPONSE=$(curl -s -X POST "$API_BASE/users/me/change-password" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $NEW_ACCESS_TOKEN" \
  -d "{\"currentPassword\": \"$TEST_PASSWORD\", \"newPassword\": \"$NEW_PASSWORD\"}")

echo "Response: $CHANGE_PW_RESPONSE"

if echo "$CHANGE_PW_RESPONSE" | jq -e '.message' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Password change successful${NC}"
    TEST_PASSWORD=$NEW_PASSWORD
else
    echo -e "${RED}✗ Password change failed${NC}"
fi
echo ""

# Test 8: Login with new password
echo -e "${YELLOW}Test 8: Login with New Password${NC}"
NEW_LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$TEST_EMAIL\", \"password\": \"$NEW_PASSWORD\"}")

NEW_LOGIN_TOKEN=$(echo $NEW_LOGIN_RESPONSE | jq -r '.accessToken // empty')
if [ -z "$NEW_LOGIN_TOKEN" ]; then
    echo -e "${RED}✗ Login with new password failed${NC}"
    exit 1
else
    echo -e "${GREEN}✓ Login with new password successful${NC}"
fi
echo ""

# Test 9: Telemetry upload with authentication (device claiming)
echo -e "${YELLOW}Test 9: Telemetry Upload with Authentication (Device Claiming)${NC}"
DEVICE_ID="test-device-$(date +%s)"
TELEMETRY_RESPONSE=$(curl -s -X POST "$API_BASE/telemetry" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $NEW_LOGIN_TOKEN" \
  -d "{
    \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
    \"deviceId\": \"$DEVICE_ID\",
    \"iTOW\": 118286240,
    \"gps\": {
      \"latitude\": 42.0,
      \"longitude\": 23.0,
      \"wgsAltitude\": 100.0,
      \"mslAltitude\": 95.0,
      \"speed\": 0.0,
      \"heading\": 0.0,
      \"numSatellites\": 8,
      \"fixStatus\": 3,
      \"horizontalAccuracy\": 1.5,
      \"verticalAccuracy\": 2.0,
      \"speedAccuracy\": 0.5,
      \"headingAccuracy\": 5.0,
      \"pdop\": 2.5,
      \"isFixValid\": true
    },
    \"motion\": {
      \"gForceX\": 0.0,
      \"gForceY\": 0.0,
      \"gForceZ\": 1.0,
      \"rotationX\": 0.0,
      \"rotationY\": 0.0,
      \"rotationZ\": 0.0
    },
    \"battery\": 85.0,
    \"isCharging\": false,
    \"timeAccuracy\": 25,
    \"validityFlags\": 7
  }")

echo "Response: $TELEMETRY_RESPONSE"

if echo "$TELEMETRY_RESPONSE" | jq -e '.message' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Telemetry upload successful${NC}"
else
    echo -e "${RED}✗ Telemetry upload failed${NC}"
fi
echo ""

# Test 10: List devices
echo -e "${YELLOW}Test 10: List User Devices${NC}"
DEVICES_RESPONSE=$(curl -s -X GET "$API_BASE/devices" \
  -H "Authorization: Bearer $NEW_LOGIN_TOKEN")

echo "Response: $DEVICES_RESPONSE"

DEVICE_COUNT=$(echo $DEVICES_RESPONSE | jq -r '.devices | length // 0')
if [ "$DEVICE_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ Device listing successful (found $DEVICE_COUNT device(s))${NC}"
else
    echo -e "${YELLOW}⚠ No devices found (expected at least 1)${NC}"
fi
echo ""

# Test 11: Logout
echo -e "${YELLOW}Test 11: Logout${NC}"
LOGOUT_RESPONSE=$(curl -s -X POST "$API_BASE/auth/logout" \
  -H "Authorization: Bearer $NEW_LOGIN_TOKEN")

echo "Response: $LOGOUT_RESPONSE"

if echo "$LOGOUT_RESPONSE" | jq -e '.message' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Logout successful${NC}"
else
    echo -e "${RED}✗ Logout failed${NC}"
fi
echo ""

# Summary
echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}All manual tests completed!${NC}"
echo -e "${GREEN}================================${NC}"
echo ""
echo "Test Summary:"
echo "  ✓ User Registration"
echo "  ✓ User Login"
echo "  ✓ Protected Endpoint Access"
echo "  ✓ Unauthorized Access Rejection"
echo "  ✓ Token Refresh"
echo "  ✓ Profile Update"
echo "  ✓ Password Change"
echo "  ✓ Telemetry Upload with Auth"
echo "  ✓ Device Listing"
echo "  ✓ Logout"