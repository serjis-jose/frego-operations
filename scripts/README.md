# Frego Operations Scripts

This directory contains utility scripts for managing the Operations microservice.

## Available Scripts

### Database Management

#### `bootstrap_db.sh`
Bootstrap the operations database with the provisioning procedure.

**Usage:**
```bash
./scripts/bootstrap_db.sh
```

**Environment Variables:**
- `OPERATIONS_DB_HOST` - Database host (default: `postgres-db`)
- `OPERATIONS_DB_PORT` - Database port (default: `5432`)
- `OPERATIONS_DB_NAME` - Database name (default: `frego_operations_db`)
- `OPERATIONS_DB_SUPERUSER` - Superuser name (default: `postgres`)
- `OPERATIONS_DB_SUPERUSER_PASSWORD` - Superuser password (default: `postgres`)
- `OPERATIONS_DB_OWNER` - Database owner (default: `erp_user`)

**What it does:**
1. Waits for PostgreSQL to be ready
2. Creates `frego_operations_db` if it doesn't exist
3. Installs `uuid-ossp` extension
4. Runs `db/provision_tenant.sql` to install the provisioning procedure

---

### Service Management

#### `start_service.sh`
Start the operations service locally (for development).

**Usage:**
```bash
./scripts/start_service.sh
```

**Prerequisites:**
- PostgreSQL running with `frego_tenant_db` and `frego_operations_db`
- Go 1.21+ installed

**Environment Variables:**
- `HTTP_ADDRESS` - Service port (default: `:8092`)
- `TENANT_DB_URL` - Tenant database URL
- `DB_URL` - Operations database URL

---

### Docker Management

#### `build_docker.sh`
Build the Docker image for the operations service.

**Usage:**
```bash
./scripts/build_docker.sh
```

**Environment Variables:**
- `IMAGE_NAME` - Image name (default: `frego-operations`)
- `IMAGE_TAG` - Image tag (default: `latest`)

**Example:**
```bash
IMAGE_NAME=my-operations IMAGE_TAG=v1.0.0 ./scripts/build_docker.sh
```

---

#### `docker_compose_up.sh`
Start all services using Docker Compose.

**Usage:**
```bash
./scripts/docker_compose_up.sh
```

**What it does:**
1. Starts PostgreSQL container
2. Runs database bootstrap
3. Starts operations service

---

### Build Scripts

#### `compile.sh`
Compile the operations service binary.

**Usage:**
```bash
./scripts/compile.sh
```

**Output:** `bin/operations-server`

---

## Typical Workflow

### First Time Setup

1. **Start PostgreSQL:**
   ```bash
   docker compose up postgres-db -d
   ```

2. **Bootstrap Database:**
   ```bash
   ./scripts/bootstrap_db.sh
   ```

3. **Start Service:**
   ```bash
   ./scripts/start_service.sh
   ```

### Using Docker

```bash
./scripts/docker_compose_up.sh
```

This will start everything in containers.

### Building for Production

```bash
./scripts/build_docker.sh
docker push your-registry/frego-operations:latest
```

---

## Notes

- All scripts use `set -euo pipefail` for safety
- Database scripts wait for PostgreSQL to be ready before proceeding
- Scripts can be run from any directory (they auto-detect their location)
