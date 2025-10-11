# avt-service

A simple backend Service for Automatic Vehicle Telemetry

## Installation

1. Clone the repository
2. Install dependencies:

```bash
go mod download
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

### Quick Commands

```bash
make run          # Run the server directly
make check        # Run all checks (fmt, lint, test)
make clean        # Remove build artifacts
make deps         # Download dependencies
```

## Running the Service

Start the server:

```bash
make run
# or
go run cmd/server/main.go
```

The service will listen on port 8080 by default. You can change the port using the `PORT` environment variable:

```bash
PORT=3000 make run
```

## API Endpoints

### Telemetry Ingestion

**Endpoint:** `POST /api/telemetry`

Receives telemetry data and logs it to the console.

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
  "timestamp": "2022-01-10T08:51:08.239Z"
}
```

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

The telemetry data will be logged to the console in a structured format for monitoring and debugging.

## License

See LICENSE file for details.
