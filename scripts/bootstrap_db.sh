#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_DB_FILE=$(cd "$SCRIPT_DIR/.." && pwd)/db/provision_tenant.sql
CONTAINER_DB_FILE=/app/db/provision_tenant.sql

if [ -f "$CONTAINER_DB_FILE" ]; then
  PROVISION_FILE="$CONTAINER_DB_FILE"
else
  PROVISION_FILE="$REPO_DB_FILE"
fi

OPERATIONS_DB_HOST=${OPERATIONS_DB_HOST:-postgres-db}
OPERATIONS_DB_PORT=${OPERATIONS_DB_PORT:-5432}
OPERATIONS_DB_NAME=${OPERATIONS_DB_NAME:-frego_operations_db}
OPERATIONS_DB_SUPERUSER=${OPERATIONS_DB_SUPERUSER:-postgres}
OPERATIONS_DB_SUPERUSER_PASSWORD=${OPERATIONS_DB_SUPERUSER_PASSWORD:-postgres}
OPERATIONS_DB_OWNER=${OPERATIONS_DB_OWNER:-erp_user}

echo "==> Bootstrapping operations database"
echo "    Host:        ${OPERATIONS_DB_HOST}:${OPERATIONS_DB_PORT}"
echo "    Database:    ${OPERATIONS_DB_NAME}"
echo "    Superuser:   ${OPERATIONS_DB_SUPERUSER}"
echo "    Owner:       ${OPERATIONS_DB_OWNER}"
echo "    Provisioner: ${PROVISION_FILE}"

if ! command -v psql >/dev/null 2>&1; then
  echo "ERROR: psql command not found. Please install PostgreSQL client utilities." >&2
  exit 1
fi

export PGPASSWORD="${OPERATIONS_DB_SUPERUSER_PASSWORD}"

if command -v pg_isready >/dev/null 2>&1; then
  until pg_isready -h "${OPERATIONS_DB_HOST}" -p "${OPERATIONS_DB_PORT}" -U "${OPERATIONS_DB_SUPERUSER}" >/dev/null 2>&1; do
    echo "    waiting for Postgres at ${OPERATIONS_DB_HOST}:${OPERATIONS_DB_PORT}..."
    sleep 3
  done
else
  echo "    pg_isready not found; skipping readiness check"
fi

echo "==> Ensuring database exists..."
DB_EXISTS=$(
  psql -tAc "SELECT 1 FROM pg_database WHERE datname = '${OPERATIONS_DB_NAME}'" \
    -h "${OPERATIONS_DB_HOST}" \
    -p "${OPERATIONS_DB_PORT}" \
    -U "${OPERATIONS_DB_SUPERUSER}" \
    -d postgres | tr -d '[:space:]'
)

if [ "${DB_EXISTS}" != "1" ]; then
  echo "    creating database ${OPERATIONS_DB_NAME} owned by ${OPERATIONS_DB_OWNER}"
  psql -v ON_ERROR_STOP=1 \
    -h "${OPERATIONS_DB_HOST}" \
    -p "${OPERATIONS_DB_PORT}" \
    -U "${OPERATIONS_DB_SUPERUSER}" \
    -d postgres \
    -c "CREATE DATABASE \"${OPERATIONS_DB_NAME}\" OWNER \"${OPERATIONS_DB_OWNER}\""
else
  echo "    database ${OPERATIONS_DB_NAME} already exists"
  echo "    ensuring database ownership is set to ${OPERATIONS_DB_OWNER}"
  psql -v ON_ERROR_STOP=1 \
    -h "${OPERATIONS_DB_HOST}" \
    -p "${OPERATIONS_DB_PORT}" \
    -U "${OPERATIONS_DB_SUPERUSER}" \
    -d postgres \
    -c "ALTER DATABASE \"${OPERATIONS_DB_NAME}\" OWNER TO \"${OPERATIONS_DB_OWNER}\""
fi

echo "==> Creating uuid-ossp extension..."
psql -v ON_ERROR_STOP=1 \
  -h "${OPERATIONS_DB_HOST}" \
  -p "${OPERATIONS_DB_PORT}" \
  -U "${OPERATIONS_DB_SUPERUSER}" \
  -d "${OPERATIONS_DB_NAME}" \
  -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"

if [ ! -f "${PROVISION_FILE}" ]; then
  echo "ERROR: provisioning SQL not found at ${PROVISION_FILE}" >&2
  exit 1
fi

echo "==> Running provisioning script..."
psql -v ON_ERROR_STOP=1 \
  -h "${OPERATIONS_DB_HOST}" \
  -p "${OPERATIONS_DB_PORT}" \
  -U "${OPERATIONS_DB_SUPERUSER}" \
  -d "${OPERATIONS_DB_NAME}" \
  -f "${PROVISION_FILE}"

echo "==> Operations database bootstrap complete!"
