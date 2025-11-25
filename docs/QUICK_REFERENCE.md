# 🚀 Quick Reference: Shared Tenant Registry

## Database URLs

```bash
# Shared (all services)
TENANT_DB_URL=postgres://postgres:postgres@localhost:5432/frego_tenant_db

# Finance service
DB_URL=postgres://postgres:postgres@localhost:5433/frego_finance_db

# Operations service  
DB_URL=postgres://postgres:postgres@localhost:5432/frego_erp_db
```

## Common Commands

### Setup & Migration
```bash
# Run migration (one time)
./scripts/migrate_to_shared_tenant_db.sh

# Verify migration
psql frego_tenant_db -c "SELECT * FROM tenant_registry;"
```

### Provision Tenant
```bash
# Full provisioning (all modules)
./scripts/provision_tenant_complete.sh "Company Name" "email@example.com" "operations,finance"

# Finance only
./scripts/provision_tenant_complete.sh "Finance Co" "finance@example.com" "finance"
```

### Query Tenant Info
```sql
-- List all tenants
SELECT tenant_slug, tenant_name, array_to_string(modules_subscribed, ', ') as modules
FROM tenant_registry;

-- Get tenant schema names
SELECT 
    tenant_slug,
    operations_schema,
    finance_schema
FROM tenant_registry
WHERE tenant_slug = 'acme';

-- Check provisioning log
SELECT * FROM tenant_module_log ORDER BY provisioned_at DESC LIMIT 10;
```

### Manual Provisioning
```sql
-- 1. Register tenant
CALL register_tenant(
    'uuid-here'::uuid,
    'acme',
    'Acme Corp',
    'admin@acme.com'
);

-- 2. Subscribe to module
CALL subscribe_tenant_to_module('uuid-here'::uuid, 'finance', 'finance_acme');

-- 3. Provision in service DB
-- (Run in frego_finance_db)
CALL ensure_finance_tenant_schema('uuid-here'::uuid, 'finance_acme');

-- 4. Mark complete
CALL mark_module_provisioned('uuid-here'::uuid, 'finance', 'success');
```

## File Locations

```
frego-operations-microservice/
├── db/
│   └── tenant_registry_schema.sql       ← Shared DB schema
├── scripts/
│   ├── migrate_to_shared_tenant_db.sh   ← Migration script
│   └── provision_tenant_complete.sh     ← Provisioning script
└── docs/
    ├── MIGRATION_TO_SHARED_TENANT_DB.md ← Migration guide
    └── SHARED_TENANT_REGISTRY_SUMMARY.md ← Summary
```

## Troubleshooting

```bash
# Can't connect to tenant DB
psql $TENANT_DB_URL -c "SELECT 1"

# Tenant not found
psql frego_tenant_db -c "SELECT * FROM tenant_registry WHERE tenant_slug = 'acme';"

# Module not provisioned
psql frego_tenant_db -c "SELECT * FROM tenant_module_log WHERE tenant_id = 'uuid';"

# Schema doesn't exist
psql frego_finance_db -c "\dn finance_*"
```

## API Testing

```bash
# Get tenant UUID
TENANT_ID=$(psql frego_tenant_db -tAc "SELECT tenant_id FROM tenant_registry WHERE tenant_slug = 'acme';")

# Test finance API
curl -H "X-Tenant-ID: $TENANT_ID" http://localhost:8081/health
curl -H "X-Tenant-ID: $TENANT_ID" http://localhost:8081/finance/api/v1/invoices

# Test operations API
curl -H "X-Tenant-ID: $TENANT_ID" http://localhost:8080/health
curl -H "X-Tenant-ID: $TENANT_ID" http://localhost:8080/frego-operations/api/v1/jobs
```

## Quick Checks

```bash
# 1. All databases exist?
psql -l | grep frego

# 2. Tenant registry populated?
psql frego_tenant_db -c "SELECT COUNT(*) FROM tenant_registry;"

# 3. Services can connect?
cd frego-operations-microservice && make run
cd frego-backend && make run

# 4. Schemas exist?
psql frego_finance_db -c "\dn"
psql frego_erp_db -c "\dn"
```
