# Frego Operations Microservice

A standalone microservice for managing freight forwarding operations, including jobs, packages, carriers, billing, and tracking.

## Overview

- **Service Name**: `frego-operations`
- **Port**: `8092`
- **Database**: `frego_operations_db` (PostgreSQL)
- **Schema Strategy**: Schema-per-tenant (e.g., `ops_tenant_abc`, `ops_tenant_xyz`)
- **Isolation**: Each tenant has a dedicated schema with all operations tables

## Architecture

### Multi-Tenant Design

The operations service uses a **schema-per-tenant** approach:

- **Shared Tenant Registry**: `frego_tenant_db` (shared across all microservices)
  - Stores tenant metadata
  - Maps tenants to their operations schemas
  
- **Operations Database**: `frego_operations_db`
  - Contains tenant-specific schemas (e.g., `ops_550e8400...`)
  - Each schema has complete operations tables

### Key Features

- **Job Management**: Create, update, and track freight forwarding jobs
- **Package Tracking**: Manage shipment packages and cargo details
- **Carrier Information**: Track carriers, vessels, flights, and vehicles
- **Billing & Provisions**: Handle job-related billing and cost provisions
- **Document Management**: Attach and track job-related documents
- **Multi-tenant Isolation**: Complete data isolation per tenant

## Database Schema

Each tenant schema contains:

- **Core Operations Tables**:
  - `ops_job` - Main job/shipment records
  - `ops_package` - Package details
  - `ops_carrier` - Carrier information
  - `ops_billing` - Billing entries
  - `ops_provision` - Cost provisions
  - `ops_tracking` - Shipment tracking
  - `ops_job_document` - Document attachments

- **Lookup Tables**:
  - `trans_move_service_lu` - Transport modes and services
  - `job_status_lu` - Job statuses
  - `document_status_lu` - Document statuses
  - `priority_lu` - Priority levels
  - `role_details_lu` - Role definitions

- **Shared Dependencies**:
  - `party_master` - Customers, agents, vendors
  - `employee_master` - Employees
  - `branch_lu` - Branch information
  - `currency_lu`, `country_lu`, etc.

## Setup

### Prerequisites

- Go 1.21+
- PostgreSQL 16+
- Docker & Docker Compose (optional)

### Local Development

1. Create the operations database:
```sql
CREATE DATABASE frego_operations_db;
```

2. Install the tenant provisioning procedure:
```bash
psql -d frego_operations_db -f db/provision_tenant.sql
```

### Tenant Provisioning

When a new tenant subscribes to the operations module:

```sql
CALL ensure_operations_tenant_schema(
  '550e8400-e29b-41d4-a716-446655440000'::uuid,  -- p_tenant_id
  'ops_acme'       -- p_schema (optional, defaults to ops_<uuid>)
);
```

This creates:
- A new schema (e.g., `ops_acme`)
- All operations tables
- All lookup tables
- Grants to `erp_user` role

## API Endpoints

### Operations Management

```
POST   /operations/api/v1/jobs           # Create job
GET    /operations/api/v1/jobs           # List jobs
GET    /operations/api/v1/jobs/{id}      # Get job details
PUT    /operations/api/v1/jobs/{id}      # Update job
DELETE /operations/api/v1/jobs/{id}      # Archive job
```

### Tenant Provisioning

```
POST   /operations/internal/provision-tenant  # Provision operations schema
```

## Running the Service

### Using Docker Compose

```bash
docker compose up -d
```

The service will be available at `http://localhost:8092`

### Building from Source

```bash
go build -o bin/operations-server ./cmd/server
./bin/operations-server
```

### Environment Variables

```bash
# Tenant Database (Shared)
TENANT_DB_URL=postgresql://erp_user:password@localhost:5432/frego_tenant_db

# Operations Database (Service-specific)
DB_URL=postgresql://erp_user:password@localhost:5432/frego_operations_db
DB_USER=erp_user

# Service Configuration
HTTP_ADDRESS=:8080
ENVIRONMENT=development

# Security
KEYCLOAK_ISSUER=http://keycloak:8080/realms/frego
KEYCLOAK_AUDIENCE=frego-app
KEYCLOAK_TENANT_CLAIM=tenantId
FREGO_INTERNAL_SECRET=your-secret-here
```

## API Documentation

OpenAPI documentation will be available at:
```
http://localhost:8092/operations/api/v1/docs
```

## Project Structure

```
frego-operations/
├── cmd/
│   └── server/           # Main application entry point
├── internal/
│   ├── api/              # HTTP handlers
│   ├── auth/             # Authentication middleware
│   ├── config/           # Configuration management
│   ├── db/               # Database connections
│   │   └── sqlc/         # Generated database code
│   ├── dto/              # Data transfer objects
│   │   ├── operations/   # Operations DTOs
│   │   └── tenant/       # Tenant DTOs
│   ├── logging/          # Logging utilities
│   ├── repository/       # Data access layer
│   │   ├── operations/   # Operations repository
│   │   └── tenant/       # Tenant repository
│   ├── server/           # HTTP server setup
│   ├── service/          # Business logic
│   │   ├── operations/   # Operations service
│   │   └── tenant/       # Tenant service
│   └── storage/          # File storage (S3)
├── db/
│   ├── schema.sql        # Database schema
│   ├── provision_tenant.sql  # Tenant provisioning procedure
│   └── queries/
│       └── operations.sql    # SQLC queries
├── scripts/              # Deployment scripts
├── docker-compose.yml    # Docker configuration
└── Dockerfile           # Container image
```

## Development

### Generate Database Code

```bash
sqlc generate
```

### Run Tests

```bash
go test ./...
```

### Build Docker Image

```bash
docker build -t frego-operations:latest .
```

## Integration with Other Services

The operations service integrates with:

- **frego-backend**: Shares party and employee data
- **frego-finance**: Provides job data for invoicing
- **frego-core**: Uses party management APIs

## License

Proprietary - Frego ERP System
