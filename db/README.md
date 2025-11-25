# Database Files

This directory contains all database-related SQL files for the Finance microservice.

## 📁 Files

### **Core Schema Files**

| File | Purpose | Usage |
|------|---------|-------|
| **provision_tenant.sql** | Complete tenant provisioning procedure | Run once in frego_finance_db to create procedure |
| **schema.sql** | Static schema reference (for documentation/IDE) | Not executed, just for reference |
| **tenant_registry_schema.sql** | Shared tenant registry schema | Run once in frego_tenant_db |

### **Query Files**

| Directory | Purpose |
|-----------|---------|
| **queries/finance.sql** | SQLC query definitions for finance operations |

---

## 🚀 Usage

### Initial Setup

```bash
# 1. Create tenant registry database (shared)
createdb frego_tenant_db
psql frego_tenant_db -f db/tenant_registry_schema.sql

# 2. Create finance database
createdb frego_finance_db

# 3. Install provisioning procedure
psql frego_finance_db -f db/provision_tenant.sql
```

### Provision a Tenant

```bash
# Option 1: Use the script (recommended)
./scripts/provision_tenant_complete.sh "Company Name" "email@example.com" "finance"

# Option 2: Manual SQL
psql frego_finance_db <<SQL
CALL ensure_finance_tenant_schema(
    '550e8400-e29b-41d4-a716-446655440000'::uuid,
    'finance_acme'
);
SQL
```

---

## 📝 File Details

### **provision_tenant.sql**
Complete stored procedure that creates all finance tables in a tenant-specific schema.

**Tables created:**
- **Lookups**: currency_lu, payment_term_lu, gl_account_lu, tax_code_lu, etc.
- **Ledger**: journal_entry_header, journal_entry_lines, general_ledger
- **AR**: ar_invoice, ar_receipt, ar_credit_note (with lines)
- **AP**: ap_vendor_invoice, ap_payment_*, ap_debit_note (with lines)
- **Audit**: approval_history with triggers

**Key features:**
- UUIDs for external references (no hard FKs to operations DB)
- Approval workflow with audit triggers
- JSONB for supporting documents
- Complete indexing strategy

### **tenant_registry_schema.sql**
Shared tenant registry that all services query.

**Tables:**
- `tenant_registry` - Main tenant metadata with module subscriptions
- `tenant_module_log` - Audit log of provisioning activities

**Functions:**
- `tenant_has_module(uuid, text)` - Check if tenant has a module
- `get_tenant_schema(uuid, text)` - Get schema name for tenant/module

**Procedures:**
- `register_tenant()` - Register new tenant
- `subscribe_tenant_to_module()` - Subscribe to a module
- `mark_module_provisioned()` - Mark provisioning complete/failed

### **schema.sql**
Static copy of the finance schema for:
- IDE autocomplete
- Documentation
- Schema comparison tools
- **Not executed** - just for reference

### **queries/finance.sql**
SQLC query definitions for code generation.

**Query types:**
- CRUD operations for invoices, receipts, payments
- Reporting queries
- Lookup queries
- Transaction queries

---

## 🔄 Schema Updates

When updating the schema:

1. **Update provision_tenant.sql**
   ```sql
   -- Add new table or column in the EXECUTE format() blocks
   ```

2. **Update schema.sql** (for reference)
   ```bash
   # Copy relevant sections from provision_tenant.sql
   ```

3. **Update queries/finance.sql** (if needed)
   ```sql
   -- Add new queries for SQLC
   ```

4. **Regenerate code**
   ```bash
   make generate
   ```

---

## 📊 Schema Structure

```
finance_<tenant> schema:
├── Lookups (12 tables)
│   ├── currency_lu
│   ├── payment_term_lu
│   ├── gl_account_lu
│   └── ...
├── Ledger Engine (3 tables)
│   ├── journal_entry_header
│   ├── journal_entry_lines
│   └── general_ledger
├── AR Module (6 tables)
│   ├── ar_invoice
│   ├── ar_invoice_line
│   ├── ar_receipt
│   ├── ar_receipt_invoice_allocation
│   └── ar_credit_note
├── AP Module (7 tables)
│   ├── ap_vendor_invoice
│   ├── ap_vendor_invoice_line
│   ├── ap_payment_against_invoice
│   ├── ap_payment_without_invoice
│   └── ap_debit_note
└── Audit (1 table)
    └── approval_history
```

**Total: 30+ tables per tenant schema**

---

## ⚠️ Important Notes

1. **External References**: Tables reference `party_id`, `job_id`, `employee_id` as UUIDs without foreign keys
2. **Multi-Tenancy**: Each tenant gets their own schema (e.g., `finance_acme`, `finance_xyz`)
3. **Shared Registry**: All services query `frego_tenant_db` for schema mappings
4. **Approval Triggers**: Automatic audit trail when approval status changes

---

## 🆘 Troubleshooting

### Procedure doesn't exist
```sql
-- Check if procedure exists
SELECT routine_name FROM information_schema.routines 
WHERE routine_schema = 'public' 
  AND routine_name = 'ensure_finance_tenant_schema';

-- Reinstall if needed
psql frego_finance_db -f db/provision_tenant.sql
```

### Schema already exists
```sql
-- Drop and recreate (WARNING: deletes all data)
DROP SCHEMA IF EXISTS finance_acme CASCADE;

-- Then run provisioning again
CALL ensure_finance_tenant_schema('uuid', 'finance_acme');
```

### Query tenant schemas
```sql
-- List all finance schemas
SELECT schema_name 
FROM information_schema.schemata 
WHERE schema_name LIKE 'finance_%';

-- Count tables in schema
SELECT COUNT(*) 
FROM information_schema.tables 
WHERE table_schema = 'finance_acme';
```

---

For more information, see:
- **Migration Guide**: `../docs/MIGRATION_TO_SHARED_TENANT_DB.md`
- **Architecture**: `../docs/ARCHITECTURE.md`
