# AVT Service

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

Example:

```bash
PORT=3000 DATABASE_URL="postgres://user:pass@localhost:5432/telemetry?sslmode=disable" go run cmd/server/main.go
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

### Telemetry Ingestion

**Endpoint:** `POST /api/telemetry`

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

**Example with curl:**

```bash
curl -X POST http://localhost:8080/api/telemetry \
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

**Endpoint:** `POST /api/telemetry/batch`

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

**Example with curl:**

```bash
curl -X POST http://localhost:8080/api/telemetry/batch \
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
