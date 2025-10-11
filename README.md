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

### Testing with curl

Test the hello endpoint:

```bash
curl http://localhost:8080/
```

Test the health endpoint:

```bash
curl http://localhost:8080/health
```

Test the greeting endpoint:

```bash
curl http://localhost:8080/api/greeting/YourName
```

## License

See LICENSE file for details.
