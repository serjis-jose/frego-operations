#!/usr/bin/env bash
# Start Finance Database using Docker Compose
# Note: Tenant DB is expected to be running via frego-backend (shared)

set -euo pipefail

echo "============================================"
echo "Starting Finance Database"
echo "============================================"

# Check if docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker is not running"
    exit 1
fi

# Start only the finance database
echo "Starting finance-db..."
docker-compose up -d finance-db

echo ""
echo "Waiting for finance database..."

# Wait for Finance DB
echo -n "  Waiting for frego-operations-db (5433)..."
until docker exec frego-operations-db pg_isready -U postgres -d frego_finance_db > /dev/null 2>&1; do
    echo -n "."
    sleep 1
done
echo " Ready!"

echo ""
echo "Checking for Shared Tenant DB..."
if docker ps | grep -q "frego-backend.*postgres"; then
    echo "  ✓ Found shared postgres-db (running)"
else
    echo "  ⚠️  Shared postgres-db NOT found (expected frego-backend container)!"
    echo "      Please start frego-backend first to provide the shared registry."
fi

echo ""
echo "============================================"
echo "Database Status"
echo "============================================"
echo "Tenant DB:  postgres://postgres:postgres@localhost:5432/frego_tenant_db (Shared)"
echo "Finance DB: postgres://postgres:postgres@localhost:5433/frego_finance_db"
echo ""
