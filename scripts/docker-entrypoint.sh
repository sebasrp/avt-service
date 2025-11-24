#!/bin/sh
set -e

echo "Starting AVT Service..."

# Wait for database to be ready
echo "Waiting for database to be ready..."
until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER"; do
  echo "Database is unavailable - sleeping"
  sleep 2
done

echo "Database is ready!"

# Run migrations
echo "Running database migrations..."
export DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"
export MIGRATIONS_DIR="internal/database/migrations"

if [ -f "./run-migrations.sh" ]; then
    ./run-migrations.sh
else
    echo "Warning: Migration script not found, running migrations directly..."
    migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" up
fi

echo "Migrations completed successfully!"

# Start the application
echo "Starting server..."
exec ./server