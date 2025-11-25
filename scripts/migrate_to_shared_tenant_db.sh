#!/usr/bin/env bash
# Migration Script: Set up Shared Tenant Registry
# This creates the frego_tenant_db and migrates existing tenants

set -euo pipefail

TENANT_DB_URL=${TENANT_DB_URL:-postgres://postgres:postgres@localhost:5432/frego_tenant_db}
ERP_DB_URL=${ERP_DB_URL:-postgres://postgres:postgres@localhost:5432/frego_operations_db}
FINANCE_DB_URL=${FINANCE_DB_URL:-postgres://postgres:postgres@localhost:5433/frego_finance_db}

echo "============================================"
echo "Shared Tenant Registry Migration"
echo "============================================"

# Step 1: Create frego_tenant_db
echo ""
echo "Step 1: Creating frego_tenant_db..."
if psql postgres://postgres:postgres@localhost:5432/postgres -lqt | cut -d \| -f 1 | grep -qw frego_tenant_db; then
    echo "  ✓ Database frego_tenant_db already exists"
else
    createdb -h localhost -p 5432 -U postgres frego_tenant_db
    echo "  ✓ Created frego_tenant_db"
fi

# Step 2: Initialize tenant registry schema (drop existing tables for clean migration)
echo ""
echo "Step 2: Initializing tenant registry schema..."
echo "  Dropping existing tenant registry tables (if any)..."
psql "${TENANT_DB_URL}" <<SQL
-- Drop existing tables for clean migration
DROP TABLE IF EXISTS tenant_module_log CASCADE;
DROP TABLE IF EXISTS tenant_registry CASCADE;
DROP FUNCTION IF EXISTS tenant_has_module(uuid, text);
DROP FUNCTION IF EXISTS get_tenant_schema(uuid, text);
DROP PROCEDURE IF EXISTS register_tenant(uuid, text, text, text);
DROP PROCEDURE IF EXISTS subscribe_tenant_to_module(uuid, text, text);
DROP PROCEDURE IF EXISTS mark_module_provisioned(uuid, text, text, text);
SQL

echo "  Creating fresh tenant registry schema..."
psql "${TENANT_DB_URL}" -f "$(dirname "$0")/../db/tenant_registry_schema.sql"
echo "  ✓ Schema initialized (clean state)"

# Step 3: Migrate existing tenants from frego_erp_db (if exists)
echo ""
echo "Step 3: Migrating existing tenants from frego_erp_db..."
if psql "${ERP_DB_URL}" -c "SELECT 1" >/dev/null 2>&1; then
    # Check if tenant_registry exists in ERP DB
    if psql "${ERP_DB_URL}" -c "SELECT 1 FROM tenant_registry LIMIT 1" >/dev/null 2>&1; then
        echo "  Found existing tenants in frego_erp_db"
        
        # Export tenants from ERP DB
        psql "${ERP_DB_URL}" -c "COPY (
            SELECT 
                tenant_id,
                schema_name as operations_schema,
                display_name as tenant_name,
                contact_email,
                is_active,
                created_at,
                created_by
            FROM tenant_registry
        ) TO STDOUT WITH CSV HEADER" | \
        psql "${TENANT_DB_URL}" -c "COPY tenant_registry(
            tenant_id,
            operations_schema,
            tenant_name,
            contact_email,
            is_active,
            created_at,
            created_by
        ) FROM STDIN WITH CSV HEADER"
        
        # Set tenant_slug and modules_subscribed
        psql "${TENANT_DB_URL}" <<SQL
UPDATE tenant_registry SET
    tenant_slug = regexp_replace(lower(operations_schema), '^tenant_', ''),
    modules_subscribed = ARRAY['operations']::text[]
WHERE tenant_slug IS NULL;
SQL
        
        echo "  ✓ Migrated $(psql "${TENANT_DB_URL}" -tAc "SELECT COUNT(*) FROM tenant_registry") tenants"
    else
        echo "  No existing tenants found in frego_erp_db"
    fi
else
    echo "  frego_erp_db not found, skipping migration"
fi

# Step 4: Migrate existing finance tenants (if exists)
echo ""
echo "Step 4: Migrating existing finance schemas..."
if psql "${FINANCE_DB_URL}" -c "SELECT 1" >/dev/null 2>&1; then
    if psql "${FINANCE_DB_URL}" -c "SELECT 1 FROM tenant_registry LIMIT 1" >/dev/null 2>&1; then
        echo "  Found existing finance tenants"
        
        # Update tenant_registry with finance schemas
        psql "${FINANCE_DB_URL}" -c "COPY (
            SELECT tenant_id, schema_name FROM tenant_registry
        ) TO STDOUT" | \
        while IFS=$'\t' read -r tenant_id schema_name; do
            psql "${TENANT_DB_URL}" <<SQL
UPDATE tenant_registry SET
    finance_schema = '${schema_name}',
    modules_subscribed = array_append(modules_subscribed, 'finance')
WHERE tenant_id = '${tenant_id}'::uuid
  AND NOT ('finance' = ANY(modules_subscribed));

INSERT INTO tenant_registry (tenant_id, tenant_slug, tenant_name, finance_schema, modules_subscribed)
SELECT 
    '${tenant_id}'::uuid,
    regexp_replace('${schema_name}', '^finance_', ''),
    regexp_replace('${schema_name}', '^finance_', ''),
    '${schema_name}',
    ARRAY['finance']::text[]
WHERE NOT EXISTS (SELECT 1 FROM tenant_registry WHERE tenant_id = '${tenant_id}'::uuid);
SQL
        done
        
        echo "  ✓ Migrated finance schemas"
    fi
else
    echo "  frego_finance_db not found, skipping"
fi

# Step 5: Verify migration
echo ""
echo "Step 5: Verifying migration..."
echo ""
psql "${TENANT_DB_URL}" <<SQL
SELECT 
    tenant_id,
    tenant_slug,
    tenant_name,
    array_to_string(modules_subscribed, ', ') as modules,
    operations_schema,
    finance_schema,
    is_active
FROM tenant_registry
ORDER BY created_at;
SQL

echo ""
echo "============================================"
echo "Migration Complete!"
echo "============================================"
echo ""
echo "Next steps:"
echo "  1. Update application configs to use TENANT_DB_URL"
echo "  2. Update both services to query shared tenant DB"
echo "  3. Test tenant provisioning with new flow"
