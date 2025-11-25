# 🔄 Migration Guide: Shared Tenant Registry

## Overview

This guide helps you migrate from the old architecture (tenant_registry in each service DB) to the new architecture (shared tenant_registry database).

## Architecture Changes

### Before (❌ Old)
```
PostgreSQL
├── frego_erp_db
│   ├── public.tenant_registry     ← Duplicated
│   └── tenant_acme
└── frego_finance_db
    ├── public.tenant_registry     ← Duplicated
    └── finance_acme
```

### After (✅ New)
```
PostgreSQL
├── frego_tenant_db                ← NEW: Shared registry
│   └── public.tenant_registry
├── frego_erp_db                   ← Operations data only
│   └── tenant_acme
└── frego_finance_db               ← Finance data only
    └── finance_acme
```

---

## Migration Steps

### Step 1: Backup Everything

```bash
# Backup all databases
pg_dump frego_erp_db > backup_erp_$(date +%Y%m%d).sql
pg_dump frego_finance_db > backup_finance_$(date +%Y%m%d).sql
```

### Step 2: Run Migration Script

```bash
cd frego-operations

# Run the migration
./scripts/migrate_to_shared_tenant_db.sh
```

**What this does:**
1. Creates `frego_tenant_db`
2. Creates tenant_registry schema
3. Migrates existing tenants from both DBs
4. Sets up module subscriptions
5. Verifies migration

### Step 3: Update Configuration Files

#### Finance Service (.env)
```bash
# OLD
DB_URL=postgres://postgres:postgres@localhost:5433/frego_finance_db

# NEW - Add tenant DB
TENANT_DB_URL=postgres://postgres:postgres@localhost:5432/frego_tenant_db
DB_URL=postgres://postgres:postgres@localhost:5433/frego_finance_db
```

#### Operations Service (.env)
```bash
# NEW - Add tenant DB  
TENANT_DB_URL=postgres://postgres:postgres@localhost:5432/frego_tenant_db
DB_URL=postgres://postgres:postgres@localhost:5432/frego_erp_db
```

### Step 4: Update Application Code

The code has already been updated in `frego-operations`. For `frego-backend`, apply similar changes:

1. Update `internal/db/tenant_session.go`
2. Update `internal/config/config.go`
3. Update `cmd/server/main.go`

### Step 5: Test Migration

```bash
# Test 1: Verify shared registry
psql postgres://postgres:postgres@localhost:5432/frego_tenant_db -c "
SELECT 
    tenant_id,
    tenant_slug,
    array_to_string(modules_subscribed, ', ') as modules,
    operations_schema,
    finance_schema
FROM tenant_registry;
"

# Test 2: Provision new tenant
./scripts/provision_tenant_complete.sh "Test Company" "test@example.com" "operations,finance"

# Test 3: Start services
# Terminal 1: Finance service
cd frego-operations
make run

# Terminal 2: Operations service
cd frego-backend
make run

# Test 4: Make API calls with tenant header
curl -H "X-Tenant-ID: <tenant-uuid>" http://localhost:8080/frego-operations/api/v1/jobs
curl -H "X-Tenant-ID: <tenant-uuid>" http://localhost:8081/finance/api/v1/invoices
```

---

## Database Connection Summary

### After Migration

| Service | Tenant DB | Service DB |
|---------|-----------|------------|
| **Operations** | `frego_tenant_db:5432` (shared) | `frego_erp_db:5432` |
| **Finance** | `frego_tenant_db:5432` (shared) | `frego_finance_db:5433` |
| **Future Services** | `frego_tenant_db:5432` (shared) | Their own DB |

### Environment Variables

```bash
# Shared across ALL services
TENANT_DB_URL=postgres://postgres:postgres@localhost:5432/frego_tenant_db

# Finance service
DB_URL=postgres://postgres:postgres@localhost:5433/frego_finance_db

# Operations service
DB_URL=postgres://postgres:postgres@localhost:5432/frego_erp_db
```

---

## New Tenant Provisioning Flow

### Old Way (❌)
```bash
# Had to provision in each service separately
psql frego_erp_db -c "CALL ensure_party_tenant_schema(...)"
psql frego_finance_db -c "CALL ensure_finance_tenant_schema(...)"
```

### New Way (✅)
```bash
# Single command provisions across all services
./scripts/provision_tenant_complete.sh "Acme Corp" "admin@acme.com" "operations,finance"
```

**What happens:**
1. Tenant registered in shared `frego_tenant_db`
2. Modules subscribed (operations, finance, etc.)
3. Schemas provisioned in respective service DBs
4. All tracked in tenant_module_log

---

## Verifying the Migration

### 1. Check Tenant Registry
```sql
-- Connect to shared tenant DB
\c frego_tenant_db

-- View all tenants
SELECT * FROM tenant_registry;

-- Check a specific tenant
SELECT 
    tenant_name,
    array_to_string(modules_subscribed, ', ') as modules,
    operations_schema,
    finance_schema,
    is_active
FROM tenant_registry
WHERE tenant_slug = 'acme';
```

### 2. Check Module Provisioning Log
```sql
SELECT 
    t.tenant_name,
    l.module_name,
    l.schema_name,
    l.status,
    l.provisioned_at
FROM tenant_module_log l
JOIN tenant_registry t ON l.tenant_id = t.tenant_id
ORDER BY l.provisioned_at DESC;
```

### 3. Verify Schema Access
```sql
-- Check operations schema exists
\c frego_erp_db
\dn tenant_*

-- Check finance schema exists
\c frego_finance_db
\dn finance_*
```

---

## Troubleshooting

### Issue: Migration script fails

**Solution:**
```bash
# Check database connectivity
psql postgres://postgres:postgres@localhost:5432/postgres -c "SELECT version();"

# Check if databases exist
psql -l | grep frego

# Re-run with verbose output
bash -x ./scripts/migrate_to_shared_tenant_db.sh
```

### Issue: Service can't connect to tenant DB

**Solution:**
```bash
# Verify tenant DB URL in .env
cat .env | grep TENANT_DB_URL

# Test connection
psql $TENANT_DB_URL -c "SELECT 1"

# Check config loading
go run ./cmd/server | grep "tenant database"
```

### Issue: Tenant not found errors

**Solution:**
```sql
-- Check if tenant exists in registry
SELECT * FROM tenant_registry WHERE tenant_id = '<uuid>';

-- Check module subscription
SELECT 
    tenant_id,
    modules_subscribed,
    finance_schema
FROM tenant_registry 
WHERE tenant_id = '<uuid>';

-- Re-subscribe if needed
CALL subscribe_tenant_to_module('<uuid>'::uuid, 'finance', 'finance_acme');
```

---

## Rollback Plan

If you need to rollback:

```bash
# 1. Stop all services
docker compose down

# 2. Restore from backups
psql frego_erp_db < backup_erp_YYYYMMDD.sql
psql frego_finance_db < backup_finance_YYYYMMDD.sql

# 3. Revert code changes
git checkout main

# 4. Drop shared tenant DB
dropdb frego_tenant_db

# 5. Restart services with old config
```

---

## Benefits After Migration

✅ **Single Source of Truth** - One tenant registry for all services
✅ **Better Module Management** - Track which modules each tenant has
✅ **Easier Provisioning** - One command to provision across all services
✅ **Audit Trail** - tenant_module_log tracks all provisioning
✅ **Future-Proof** - Easy to add new modules (inventory, HRMS, etc.)
✅ **True Microservices** - Each service has its own DB, shares only tenant metadata

---

## Next Steps

1. ✅ Complete migration
2. ✅ Test with existing tenants
3. ✅ Provision new tenant using new flow
4. ✅ Update documentation
5. ✅ Train team on new provisioning process
6. Monitor for issues in production

---

## Support

For issues or questions:
- Check this migration guide
- Review the scripts in `scripts/`
- Check database logs: `tail -f /var/log/postgresql/*.log`
- Contact DevOps team

**Migration complete! 🎉**
