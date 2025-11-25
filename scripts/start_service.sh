#!/usr/bin/env bash
# Start the Operations Service locally
# Assumes databases are running on localhost

set -euo pipefail

# Default Configuration
export HTTP_ADDRESS=${HTTP_ADDRESS:-":8092"}
export ENVIRONMENT=${ENVIRONMENT:-"development"}

# Database Configuration (Localhost)
export TENANT_DB_URL=${TENANT_DB_URL:-"postgres://erp_user:password@localhost:5432/frego_core_db?sslmode=disable"}
export DB_URL=${DB_URL:-"postgres://postgres:postgres@localhost:5432/frego_operations_db?sslmode=disable"}

# Mock Security for local dev
export KEYCLOAK_ISSUER=${KEYCLOAK_ISSUER:-"http://localhost:8080/realms/frego"}
export DEFAULT_TENANT=${DEFAULT_TENANT:-"550e8400-e29b-41d4-a716-446655440000"}

echo "============================================"
echo "Starting Operations Microservice (Local)"
echo "============================================"
echo "Tenant DB:     $TENANT_DB_URL"
echo "Operations DB: $DB_URL"
echo "Port:          $HTTP_ADDRESS"
echo "============================================"

# Check if databases are reachable
if ! psql "$TENANT_DB_URL" -c "SELECT 1" > /dev/null 2>&1; then
    echo "Error: Cannot connect to Tenant DB at localhost:5432"
    echo "Run docker compose up postgres-db first"
    exit 1
fi

if ! psql "$DB_URL" -c "SELECT 1" > /dev/null 2>&1; then
    echo "Error: Cannot connect to Operations DB at localhost:5432"
    echo "Run docker compose up postgres-db first"
    exit 1
fi

# Generate code if needed
if [ ! -f "internal/api/operations.gen.go" ]; then
    echo "Warning: API code not generated. Run 'make generate' or 'oapi-codegen' manually"
fi

# Run the service
echo "Building and running..."
go run cmd/server/main.go
