# 🚀 Quick Start Guide

## Prerequisites
- Go 1.25.1+
- Docker & Docker Compose
- PostgreSQL 14+ (or use Docker)
- Make

## 1. Move to Separate Repository (Recommended)

```bash
# From frego-backend directory
cd ..
mv frego-backend/frego-operations-microservice ./frego-operations
cd frego-operations

# Initialize git
git init
git add .
git commit -m "Initial commit: Finance microservice"

# Add your remote
git remote add origin <your-repo-url>
git push -u origin main
```

## 2. Local Development Setup

### Option A: Using Docker Compose (Easiest)

```bash
# Start PostgreSQL
docker-compose up -d postgres

# Wait for DB to be ready
sleep 5

# Provision a test tenant
docker-compose exec postgres psql -U postgres -d frego_finance_db -c \
  "CALL ensure_finance_tenant_schema('550e8400-e29b-41d4-a716-446655440000', 'finance_demo');"

# Verify tenant was created
docker-compose exec postgres psql -U postgres -d frego_finance_db -c \
  "SELECT * FROM tenant_registry;"

# Install Go dependencies
go mod download

# Generate code (requires tools)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
make generate

# Run the service
make run
```

### Option B: Manual Setup

```bash
# 1. Start PostgreSQL
docker run -d \
  --name frego-operations-db \
  -e POSTGRES_DB=frego_finance_db \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5433:5432 \
  postgres:16-alpine

# 2. Run provisioning script
psql postgres://postgres:postgres@localhost:5433/frego_finance_db \
  -f db/provision_tenant.sql

# 3. Provision a tenant
./scripts/provision_tenant.sh 550e8400-e29b-41d4-a716-446655440000 finance_demo

# 4. Set environment variables
export DB_URL=postgres://postgres:postgres@localhost:5433/frego_finance_db
export HTTP_ADDRESS=:8080
export ENVIRONMENT=development
export DEFAULT_TENANT=550e8400-e29b-41d4-a716-446655440000

# 5. Generate code
make generate

# 6. Run the service
make run
```

## 3. Test the Service

```bash
# Health check
curl http://localhost:8080/health

# Should return: OK

# Test with tenant header (will need auth in production)
curl -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000" \
     http://localhost:8080/finance/api/v1/invoices

# Expected: 401 Unauthorized (auth not configured yet)
```

## 4. Configure Authentication (Optional for Dev)

Update `.env`:

```env
KEYCLOAK_ISSUER=https://your-keycloak.com/realms/frego
KEYCLOAK_AUDIENCE=frego-operations
KEYCLOAK_TENANT_CLAIM=tenant_id
```

## 5. Next Steps

### Implement Your First Feature

1. **Create Invoice Endpoint**
   - Implement `CreateInvoice` in `internal/service/finance/service.go`
   - Implement `CreateInvoice` in `internal/repository/finance/repository.go`
   - Implement handler in `internal/api/handler.go`

2. **Test It**
   ```bash
   curl -X POST http://localhost:8080/finance/api/v1/invoices \
     -H "Content-Type: application/json" \
     -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000" \
     -d '{
       "invoice_no": "INV-001",
       "invoice_type": "STANDARD",
       "invoice_date": "2025-11-22",
       "customer_id": "123e4567-e89b-12d3-a456-426614174000",
       "currency_code": "AED",
       "lines": [
         {
           "line_no": 1,
           "item_description": "Freight Charges",
           "amount_without_tax": 1000.00
         }
       ]
     }'
   ```

## 6. Development Workflow

```bash
# Make changes to code

# Regenerate if needed
make generate

# Run tests
go test ./...

# Build
make build

# Run
./bin/finance-server
```

## 7. Database Management

### View Tenant Schemas
```sql
SELECT * FROM tenant_registry;
```

### Query Tenant Data
```sql
SET search_path TO finance_demo, public;
SELECT * FROM ar_invoice;
```

### Add New Tenant
```bash
./scripts/provision_tenant.sh <new-tenant-uuid> <schema-name>
```

## 8. Troubleshooting

### Service won't start
```bash
# Check database connection
psql $DB_URL

# Check environment variables
env | grep DB_URL
```

### Code generation fails
```bash
# Install tools
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# Try again
make generate
```

### Tenant provisioning fails
```bash
# Check if procedure exists
psql $DB_URL -c "\df ensure_finance_tenant_schema"

# Re-run provisioning script
psql $DB_URL -f db/provision_tenant.sql
```

## 9. Production Deployment

See [DEPLOYMENT.md](docs/DEPLOYMENT.md) for:
- Kubernetes deployment
- Environment configuration
- Scaling strategies
- Monitoring setup

## 10. Resources

- **Architecture**: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- **Contributing**: [CONTRIBUTING.md](CONTRIBUTING.md)
- **API Docs**: [api/finance_openapi.yaml](api/finance_openapi.yaml)
- **Implementation Summary**: [IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md)

---

## 🎯 Success Checklist

- [ ] Repository moved to separate location
- [ ] PostgreSQL running
- [ ] Tenant provisioned
- [ ] Code generated successfully
- [ ] Service starts without errors
- [ ] Health check returns OK
- [ ] Can query tenant data

**Happy coding! 🚀**
