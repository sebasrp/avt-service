# AVT Service

[![CI/CD Pipeline](https://github.com/sebasrp/avt-service/actions/workflows/ci-cd.yml/badge.svg)](https://github.com/sebasrp/avt-service/actions/workflows/ci-cd.yml)

A high-performance backend service for Automatic Vehicle Telemetry ingestion and storage using Go, Gin, and TimescaleDB.

## Quick Start

### 1. Set up the development environment

This will install linters, download dependencies, start the database, and run migrations:

```bash
make dev-setup
```

Or manually:

```bash
# Install linters
make install-linter

# Download dependencies
go mod download

# Start database
docker-compose up -d

# Wait for database to be ready (5-10 seconds)

# Run migrations
make migrate
```

### 2. Run the service

```bash
make run
# or
go run cmd/server/main.go
```

The service will be available at `http://localhost:8080`.

## Installation

Clone the repository:

```bash
git clone https://github.com/sebasr/avt-service.git
cd avt-service
```

## Development

### Building

The build process automatically runs formatting, linting, and tests:

```bash
make build
```

This will:

1. Format all Go files
2. Run linter checks
3. Run all tests
4. Build the binary to `bin/server`

Or build manually:

```bash
go build -o bin/server cmd/server/main.go
```

Run the binary:

```bash
./bin/server
```

### Available Commands

#### Development

```bash
make dev-setup       # Set up local development environment
make run             # Run the server directly
make build           # Build the application (with fmt, lint, test)
make clean           # Remove build artifacts
```

#### Database

```bash
make docker-up       # Start Docker containers (TimescaleDB)
make docker-down     # Stop Docker containers
make migrate         # Run database migrations
make migrate-down    # Rollback last migration
make migrate-create NAME=my_migration # Create new migration
make db-shell        # Open psql shell to database
```

#### Dependencies

```bash
make deps            # Download and tidy dependencies
make install-linter  # Install golangci-lint
```

## Database Setup

### Using Docker (Recommended for Development)

Start the TimescaleDB container:

```bash
make docker-up
```

Run migrations:

```bash
make migrate
```

## Configuration

The service is configured via environment variables:

### Database Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `DATABASE_URL` | - | Full PostgreSQL connection string |
| `DB_HOST` | `localhost` | Database host |
| `DB_PORT` | `5432` | Database port |
| `DB_NAME` | `telemetry_dev` | Database name |
| `DB_USER` | `telemetry_user` | Database user |
| `DB_PASSWORD` | `telemetry_pass` | Database password |
| `DB_SSLMODE` | `disable` | PostgreSQL SSL mode |
| `DB_MAX_CONNECTIONS` | `25` | Maximum database connections |
| `DB_MAX_IDLE_CONNECTIONS` | `5` | Maximum idle connections |
| `DB_CONNECTION_MAX_LIFETIME` | `5m` | Maximum connection lifetime |

### Authentication Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `JWT_SECRET` | - | **Required** Secret key for JWT signing (use strong random string) |
| `JWT_ACCESS_TOKEN_TTL` | `1h` | Access token expiration time |
| `JWT_REFRESH_TOKEN_TTL` | `720h` (30 days) | Refresh token expiration time |

Example:

```bash
# Generate a strong JWT secret
JWT_SECRET=$(openssl rand -base64 32)

# Run with authentication enabled
PORT=3000 \
DATABASE_URL="postgres://user:pass@localhost:5432/telemetry?sslmode=disable" \
JWT_SECRET="your-secret-key-here" \
JWT_ACCESS_TOKEN_TTL="1h" \
JWT_REFRESH_TOKEN_TTL="720h" \
go run cmd/server/main.go
```

## Running the Service

### Development

```bash
make run
```

### Production

Build the binary:

```bash
make build
```

Run the binary:

```bash
./bin/server
```

Or use Docker:

```bash
docker build -t avt-service .
docker run -p 8080:8080 --env-file .env avt-service
```

## API Endpoints

### Authentication

The service supports JWT-based authentication for user management and device ownership.

#### Register

**Endpoint:** `POST /api/v1/auth/register`

Create a new user account.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123"
}
```

**Response:** 201 Created
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "emailVerified": false
  },
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expiresAt": "2024-01-10T09:51:08Z"
}
```

#### Login

**Endpoint:** `POST /api/v1/auth/login`

Authenticate and receive access/refresh tokens.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123"
}
```

**Response:** 200 OK
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "emailVerified": false
  },
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expiresAt": "2024-01-10T09:51:08Z"
}
```

#### Refresh Token

**Endpoint:** `POST /api/v1/auth/refresh`

Get a new access token using a refresh token.

**Request Body:**
```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:** 200 OK
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expiresAt": "2024-01-10T10:51:08Z"
}
```

#### Logout

**Endpoint:** `POST /api/v1/auth/logout`

Revoke all refresh tokens for the authenticated user.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** 200 OK
```json
{
  "message": "Successfully logged out"
}
```

#### Get User Profile

**Endpoint:** `GET /api/v1/users/me`

Get the authenticated user's profile.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** 200 OK
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "emailVerified": false,
  "createdAt": "2024-01-01T00:00:00Z",
  "profile": {
    "displayName": "John Doe",
    "timezone": "UTC",
    "unitsPreference": "metric"
  }
}
```

#### Update User Profile

**Endpoint:** `PATCH /api/v1/users/me`

Update the authenticated user's profile.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "displayName": "John Doe",
  "timezone": "America/New_York",
  "unitsPreference": "imperial"
}
```

**Response:** 200 OK

#### Change Password

**Endpoint:** `POST /api/v1/users/me/change-password`

Change the authenticated user's password.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "currentPassword": "oldPassword123",
  "newPassword": "newSecurePassword456"
}
```

**Response:** 200 OK

### Device Management

#### List Devices

**Endpoint:** `GET /api/v1/devices`

List all devices owned by the authenticated user.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** 200 OK
```json
{
  "devices": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440000",
      "deviceId": "device-001",
      "deviceName": "My Car",
      "deviceModel": "Tesla Model 3",
      "claimedAt": "2024-01-01T00:00:00Z",
      "lastSeenAt": "2024-01-10T08:51:08Z",
      "isActive": true
    }
  ]
}
```

#### Get Device

**Endpoint:** `GET /api/v1/devices/:id`

Get details of a specific device.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** 200 OK

#### Update Device

**Endpoint:** `PATCH /api/v1/devices/:id`

Update device information.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "deviceName": "My Updated Car Name",
  "metadata": {
    "color": "red",
    "year": 2023
  }
}
```

**Response:** 200 OK

#### Deactivate Device

**Endpoint:** `DELETE /api/v1/devices/:id`

Deactivate a device (soft delete).

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** 200 OK

### Error Responses

All endpoints return consistent error responses:

**400 Bad Request** - Invalid request data
```json
{
  "error": "validation error",
  "details": "email is required"
}
```

**401 Unauthorized** - Missing or invalid authentication
```json
{
  "error": "unauthorized",
  "details": "invalid or expired token"
}
```

**403 Forbidden** - Insufficient permissions
```json
{
  "error": "forbidden",
  "details": "you do not have permission to access this device"
}
```

**404 Not Found** - Resource not found
```json
{
  "error": "not found",
  "details": "device not found"
}
```

**409 Conflict** - Resource conflict
```json
{
  "error": "conflict",
  "details": "email already registered"
}
```

**429 Too Many Requests** - Rate limit exceeded
```json
{
  "error": "rate limit exceeded",
  "details": "too many requests, please try again later"
}
```

**500 Internal Server Error** - Server error
```json
{
  "error": "internal server error",
  "details": "an unexpected error occurred"
}
```

### Token Format

**Access Token:**
- Type: JWT (JSON Web Token)
- Algorithm: HS256
- Expiration: 1 hour (configurable via `JWT_ACCESS_TOKEN_TTL`)
- Claims:
  - `sub`: User ID (UUID)
  - `email`: User email
  - `exp`: Expiration timestamp
  - `iat`: Issued at timestamp

**Refresh Token:**
- Type: JWT (JSON Web Token)
- Algorithm: HS256
- Expiration: 30 days (configurable via `JWT_REFRESH_TOKEN_TTL`)
- Claims:
  - `sub`: User ID (UUID)
  - `email`: User email
  - `jti`: Unique token ID (for rotation)
  - `exp`: Expiration timestamp
  - `iat`: Issued at timestamp

**Usage:**
```bash
# Include access token in Authorization header
curl -H "Authorization: Bearer <access_token>" \
  http://localhost:8080/api/v1/users/me
```

**Token Rotation:**
- Refresh tokens are rotated on each use
- Old refresh token is revoked when new one is issued
- Each refresh token has a unique `jti` claim to prevent reuse

### API Versioning

The API supports versioning to ensure backward compatibility:

- **Current Version:** `v1`
- **Versioned Endpoints:** `/api/v1/*`

### Request ID Tracking

Every request automatically receives a unique request ID for tracing and debugging:

- **Header:** `X-Request-ID`
- **Auto-generated:** UUID v4 format
- **Client-provided:** You can provide your own request ID in the `X-Request-ID` header
- **Response:** The request ID is returned in the response header

Example:

```bash
curl -X POST http://localhost:8080/api/v1/telemetry \
  -H "X-Request-ID: my-custom-id-123" \
  -H "Content-Type: application/json" \
  -d '{ ... }'
```

Response will include:

```
X-Request-ID: my-custom-id-123
```

### Telemetry Ingestion

**Endpoint:** `POST /api/v1/telemetry`

Receives telemetry data and stores it in TimescaleDB.

**Request Body:** JSON

```json
{
  "iTOW": 118286240,
  "timestamp": "2022-01-10T08:51:08.239Z",
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
}
```

**Response:** 201 Created

```json
{
  "message": "Telemetry data received successfully",
  "timestamp": "2022-01-10T08:51:08.239Z",
  "id": 12345
}
```

**Optional Fields:**

- `deviceId` (string): Device identifier
- `sessionId` (UUID): Session identifier for grouping telemetry data

**Example with curl (v1 API):**

```bash
curl -X POST http://localhost:8080/api/v1/telemetry \
  -H "Content-Type: application/json" \
  -d '{
    "iTOW": 118286240,
    "timestamp": "2022-01-10T08:51:08.239Z",
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

The telemetry data is stored in TimescaleDB and also logged to the console in a structured format for monitoring and debugging.

**Note:** For batch uploads, only the first and last records are logged to avoid excessive console output.

### Batch Telemetry Ingestion

**Endpoint:** `POST /api/v1/telemetry/batch`

Receives multiple telemetry data points in a single request for efficient batch processing.

**Request Body:** JSON Array

```json
[
  {
    "iTOW": 118286240,
    "timestamp": "2022-01-10T08:51:08.239Z",
    "gps": { ... },
    "motion": { ... },
    "battery": 89.0,
    "isCharging": false,
    "timeAccuracy": 25,
    "validityFlags": 7
  },
  {
    "iTOW": 118286340,
    "timestamp": "2022-01-10T08:51:08.339Z",
    "gps": { ... },
    "motion": { ... },
    "battery": 89.0,
    "isCharging": false,
    "timeAccuracy": 25,
    "validityFlags": 7
  }
]
```

**Response:** 201 Created

```json
{
  "message": "Batch telemetry data received successfully (2 records)",
  "count": 2,
  "ids": [12345, 12346]
}
```

**Constraints:**

- Maximum batch size: 1000 records
- All records must have valid timestamps
- Returns array of IDs for successfully saved records

**Example with curl (v1 API):**

```bash
curl -X POST http://localhost:8080/api/v1/telemetry/batch \
  -H "Content-Type: application/json" \
  -d '[
    {
      "iTOW": 118286240,
      "timestamp": "2022-01-10T08:51:08.239Z",
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
    }
  ]'
```

## Testing

The service includes comprehensive unit and integration tests.

### Run All Tests

```bash
make test
# or
go test ./...
```

### Run Tests with Coverage

```bash
make test-coverage
# or
go test ./... -cover
```

### Run Specific Tests

```bash
# Batch endpoint tests only
go test ./internal/handlers -v -run Batch

# Server integration tests
go test ./internal/server -v

# Skip integration tests (for CI)
go test ./... -short
```

### Test Coverage

- **Handlers**: 95.3% coverage
- **Server**: 100% coverage
- **Repository**: 64.6% coverage

For detailed testing documentation, see [docs/testing.md](docs/testing.md).

## Deployment

### Docker Deployment

Build the Docker image:

```bash
docker build -t avt-service:latest .
```

Run with Docker Compose:

```bash
docker-compose up
```

## License

See LICENSE file for details.
