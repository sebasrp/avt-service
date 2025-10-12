#!/bin/bash
# Script to run database migrations

set -e

# Default values
DB_URL="${DATABASE_URL:-postgres://telemetry_user:telemetry_pass@localhost:5432/telemetry_dev?sslmode=disable}"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-internal/database/migrations}"

# Find migrate tool in common locations
MIGRATE_BIN=""
if command -v migrate &> /dev/null; then
    MIGRATE_BIN="migrate"
elif [ -f "$HOME/go/bin/migrate" ]; then
    MIGRATE_BIN="$HOME/go/bin/migrate"
elif [ -f "$(go env GOPATH)/bin/migrate" ]; then
    MIGRATE_BIN="$(go env GOPATH)/bin/migrate"
else
    echo "Error: migrate tool is not installed"
    echo "Install it with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    echo "Make sure $(go env GOPATH)/bin is in your PATH"
    exit 1
fi

# Run migrations
echo "Running migrations from $MIGRATIONS_DIR to $DB_URL"
"$MIGRATE_BIN" -path "$MIGRATIONS_DIR" -database "$DB_URL" up

echo "Migrations completed successfully"

