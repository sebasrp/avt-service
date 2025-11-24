# Multi-stage build for smaller image
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server cmd/server/main.go

# Install golang-migrate tool
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata postgresql-client

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/server .
COPY --from=builder /app/internal/database/migrations ./internal/database/migrations
COPY --from=builder /app/scripts/run-migrations.sh ./run-migrations.sh
COPY --from=builder /app/scripts/docker-entrypoint.sh ./docker-entrypoint.sh
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

# Make scripts executable
RUN chmod +x ./run-migrations.sh ./docker-entrypoint.sh

# Expose port
EXPOSE 8080

# Use entrypoint script to run migrations before starting server
ENTRYPOINT ["./docker-entrypoint.sh"]