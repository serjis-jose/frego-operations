#!/usr/bin/env bash
# Start the entire stack using Docker Compose

set -euo pipefail

echo "============================================"
echo "Starting Finance Stack (Docker)"
echo "============================================"

# Check if docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker is not running"
    exit 1
fi

# Build and start
echo "Building images..."
docker-compose build

echo "Starting services..."
docker-compose up -d

echo ""
echo "============================================"
echo "Stack is running!"
echo "============================================"
echo "Tenant DB:       localhost:5432"
echo "Finance DB:      localhost:5433"
echo "Finance Service: localhost:8081"
echo ""
echo "Logs:"
docker-compose logs -f
