# Testing Documentation

## Overview

The AVT service includes comprehensive unit and integration tests for all major components, with special focus on the batch telemetry endpoint.

## Test Coverage

### Handlers Package (95.3% coverage)

Tests cover both single and batch telemetry endpoints with various scenarios:

#### Single Telemetry Tests
- Valid telemetry data
- Invalid JSON payloads
- Missing required fields (timestamp)
- Minimal valid data
- Content-Type validation
- Database error handling

#### Batch Telemetry Tests
- Valid batches with multiple records
- Single record batches
- Empty batches
- Invalid JSON
- Missing timestamps in batch records
- Batch size validation (max 1000 records)
- Database error handling
- Session ID support
- Content-Type validation

### Server Package (100% coverage)

Integration tests verifying the complete HTTP layer:

- Full telemetry endpoint workflow
- Batch telemetry endpoint workflow
- Validation error handling
- Large payload handling (1000 records)
- Non-existent route handling (404)

### Repository Package (64.6% coverage)

Database integration tests:

- Single record saving
- Batch record saving
- Time range queries
- Recent data retrieval
- Session-based queries

## Running Tests

### All Tests

```bash
go test ./...
```

### With Coverage

```bash
go test ./... -cover
```

### With Detailed Coverage Report

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Specific Package

```bash
# Handlers only
go test ./internal/handlers -v

# Batch tests only
go test ./internal/handlers -v -run Batch

# Server integration tests
go test ./internal/server -v
```

### Skip Integration Tests

Some tests require a database connection. To skip these in CI or development:

```bash
go test ./... -short
```

## Test Scenarios

### Batch Endpoint Test Coverage

#### Success Cases

1. **Multiple Records**: Batch with 2-3 telemetry records
2. **Single Record**: Batch with just one record
3. **Maximum Size**: Batch with exactly 1000 records
4. **With Session ID**: Batch records with session identifier

#### Validation Errors

1. **Empty Batch**: Array with no records
2. **Invalid JSON**: Malformed JSON payload
3. **Missing Timestamp**: Records without required timestamp field
4. **Oversized Batch**: More than 1000 records (returns 400)

#### Error Handling

1. **Database Errors**: Simulated database connection failures
2. **Partial Validation**: Early detection of invalid records

### Expected Response Format

#### Success Response (201 Created)

```json
{
  "message": "Batch telemetry data received successfully (N records)",
  "count": N,
  "ids": [12345, 12346, ...]
}
```

#### Error Response (400/500)

```json
{
  "error": "Description of the error",
  "details": "Additional error context (optional)"
}
```

## Mock Repository

The test suite uses a mock repository implementation that:

- Returns sequential IDs starting from 1
- Simulates successful saves by default
- Can be configured to return errors
- Implements all repository interface methods

Example:

```go
mockRepo := repository.NewMockRepository()
mockRepo.SaveBatchFunc = func(_ context.Context, _ []*models.TelemetryData) error {
    return errors.New("simulated error")
}
```

## Test Best Practices

1. **Isolation**: Each test is independent and uses fresh mock instances
2. **Validation**: Response status codes and body content are verified
3. **Edge Cases**: Boundary conditions are tested (empty, max size, etc.)
4. **Error Paths**: Both success and failure scenarios are covered
5. **Real Data**: Tests use realistic telemetry data structures

## Continuous Integration

Tests are designed to run in CI environments:

- No external dependencies for unit tests
- Integration tests can be skipped with `-short` flag
- Fast execution (< 1 second for most tests)
- Clear error messages for debugging

## Coverage Goals

- **Handlers**: > 90% (currently 95.3%)
- **Server**: 100% (achieved)
- **Repository**: > 60% (currently 64.6%)

## Future Test Additions

Potential areas for expansion:

- [ ] Concurrent batch upload stress tests
- [ ] Performance benchmarks for large batches
- [ ] Database transaction rollback scenarios
- [ ] Rate limiting tests
- [ ] Authentication/authorization tests (when implemented)

