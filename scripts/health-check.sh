#!/bin/bash

# Check if services are running
if ! docker compose -f /opt/avt-service/docker-compose.prod.yml ps | grep -q "Up"; then
    echo "⚠️  Services are not running!"
    docker compose -f /opt/avt-service/docker-compose.prod.yml up -d
fi

# Check API health
if ! curl -f http://localhost:8080/api/v1/health > /dev/null 2>&1; then
    echo "⚠️  API health check failed!"
    docker compose -f /opt/avt-service/docker-compose.prod.yml restart avt-service
fi