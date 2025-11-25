-- ============================================================
-- Shared Tenant Registry Migration - SQL Version
-- ============================================================
-- IMPORTANT: Make sure you're connected to frego_tenant_db
-- 
-- Usage:
--   1. createdb frego_tenant_db
--   2. psql frego_tenant_db -f db/migration_to_shared_tenant_registry.sql
--
-- Or run manually:
--   psql -U postgres -d frego_tenant_db -f db/migration_to_shared_tenant_registry.sql
-- ============================================================

-- Step 1: Drop existing tables for clean migration
DROP TABLE IF EXISTS tenant_module_log CASCADE;
DROP TABLE IF EXISTS tenant_registry CASCADE;
DROP FUNCTION IF EXISTS tenant_has_module(uuid, text);
DROP FUNCTION IF EXISTS get_tenant_schema(uuid, text);
DROP PROCEDURE IF EXISTS register_tenant(uuid, text, text, text);
DROP PROCEDURE IF EXISTS subscribe_tenant_to_module(uuid, text, text);
DROP PROCEDURE IF EXISTS mark_module_provisioned(uuid, text, text, text);

-- Step 2: Create tenant registry schema
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE tenant_registry (
  tenant_id            uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_name          text NOT NULL,
  contact_email        text,
  
  -- Core Schemas
  operations_schema    text,
  finance_schema       text,
  
  -- Status & Audit
  is_active            boolean DEFAULT true,
  created_at           timestamptz DEFAULT now(),
  created_by           text,
  modified_at          timestamptz,
  modified_by          text,

  -- Helper for URL routing (kept as it's usually required for API)
  tenant_slug          text UNIQUE NOT NULL,
  
  CONSTRAINT tenant_slug_format CHECK (tenant_slug ~ '^[a-z0-9_]+$')
);

CREATE INDEX idx_tenant_slug ON tenant_registry(tenant_slug);
CREATE INDEX idx_tenant_active ON tenant_registry(is_active) WHERE is_active = true;

CREATE TABLE tenant_module_log (
  id                uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id         uuid NOT NULL REFERENCES tenant_registry(tenant_id),
  module_name       text NOT NULL,
  action            text NOT NULL,
  schema_name       text,
  status            text NOT NULL,
  error_message     text,
  provisioned_at    timestamptz DEFAULT now(),
  provisioned_by    text,
  
  CONSTRAINT valid_module_name CHECK (module_name IN ('operations', 'finance')),
  CONSTRAINT valid_action CHECK (action IN ('provision', 'deprovision', 'update')),
  CONSTRAINT valid_status CHECK (status IN ('pending', 'success', 'failed'))
);

CREATE INDEX idx_module_log_tenant ON tenant_module_log(tenant_id);
CREATE INDEX idx_module_log_module ON tenant_module_log(module_name);
CREATE INDEX idx_module_log_status ON tenant_module_log(status);

-- Step 3: Create helper functions
CREATE OR REPLACE FUNCTION tenant_has_module(p_tenant_id uuid, p_module text)
RETURNS boolean AS $$
DECLARE
  has_module boolean;
BEGIN
  -- Simplified logic: If schema is set, they have the module
  CASE p_module
    WHEN 'operations' THEN
      SELECT operations_schema IS NOT NULL INTO has_module FROM tenant_registry WHERE tenant_id = p_tenant_id;
    WHEN 'finance' THEN
      SELECT finance_schema IS NOT NULL INTO has_module FROM tenant_registry WHERE tenant_id = p_tenant_id;
    ELSE
      has_module := false;
  END CASE;
  
  RETURN COALESCE(has_module, false);
END;
$$ LANGUAGE plpgsql STABLE;

CREATE OR REPLACE FUNCTION get_tenant_schema(p_tenant_id uuid, p_module text)
RETURNS text AS $$
DECLARE
  schema_name text;
BEGIN
  CASE p_module
    WHEN 'operations' THEN
      SELECT operations_schema INTO schema_name FROM tenant_registry WHERE tenant_id = p_tenant_id;
    WHEN 'finance' THEN
      SELECT finance_schema INTO schema_name FROM tenant_registry WHERE tenant_id = p_tenant_id;
    ELSE
      RAISE EXCEPTION 'Invalid module: %', p_module;
  END CASE;
  
  RETURN schema_name;
END;
$$ LANGUAGE plpgsql STABLE;

-- Step 4: Create stored procedures
CREATE OR REPLACE PROCEDURE register_tenant(
  p_tenant_id uuid DEFAULT NULL,
  p_tenant_slug text DEFAULT NULL,
  p_tenant_name text DEFAULT NULL,
  p_contact_email text DEFAULT NULL
)
LANGUAGE plpgsql
AS $$
DECLARE
  v_tenant_id uuid := COALESCE(p_tenant_id, uuid_generate_v4());
  v_slug text := COALESCE(p_tenant_slug, regexp_replace(lower(p_tenant_name), '[^a-z0-9]', '_', 'g'));
  v_actor text := COALESCE(NULLIF(current_setting('app.actor', true), ''), session_user::text);
BEGIN
  INSERT INTO tenant_registry (tenant_id, tenant_slug, tenant_name, contact_email, created_by, modified_by)
  VALUES (v_tenant_id, v_slug, p_tenant_name, p_contact_email, v_actor, v_actor)
  ON CONFLICT (tenant_id) DO UPDATE SET
    tenant_name = EXCLUDED.tenant_name,
    contact_email = EXCLUDED.contact_email,
    modified_at = now(),
    modified_by = v_actor;
    
  RAISE NOTICE 'Tenant registered: % (slug: %)', v_tenant_id, v_slug;
END;
$$;

CREATE OR REPLACE PROCEDURE subscribe_tenant_to_module(
  p_tenant_id uuid,
  p_module text,
  p_schema_name text DEFAULT NULL
)
LANGUAGE plpgsql
AS $$
DECLARE
  v_schema_name text;
  v_actor text := COALESCE(NULLIF(current_setting('app.actor', true), ''), session_user::text);
BEGIN
  IF NOT EXISTS (SELECT 1 FROM tenant_registry WHERE tenant_id = p_tenant_id) THEN
    RAISE EXCEPTION 'Tenant % does not exist', p_tenant_id;
  END IF;
  
  IF p_schema_name IS NOT NULL THEN
    v_schema_name := p_schema_name;
  ELSE
    SELECT tenant_slug INTO v_schema_name FROM tenant_registry WHERE tenant_id = p_tenant_id;
    v_schema_name := p_module || '_' || v_schema_name;
  END IF;
  
  CASE p_module
    WHEN 'operations' THEN
      UPDATE tenant_registry SET
        operations_schema = v_schema_name,
        modified_at = now(),
        modified_by = v_actor
      WHERE tenant_id = p_tenant_id;
    WHEN 'finance' THEN
      UPDATE tenant_registry SET
        finance_schema = v_schema_name,
        modified_at = now(),
        modified_by = v_actor
      WHERE tenant_id = p_tenant_id;
    ELSE
      RAISE EXCEPTION 'Invalid module: %', p_module;
  END CASE;
  
  INSERT INTO tenant_module_log (tenant_id, module_name, action, schema_name, status, provisioned_by)
  VALUES (p_tenant_id, p_module, 'provision', v_schema_name, 'pending', v_actor);
  
  RAISE NOTICE 'Tenant % subscribed to % (schema: %)', p_tenant_id, p_module, v_schema_name;
END;
$$;

CREATE OR REPLACE PROCEDURE mark_module_provisioned(
  p_tenant_id uuid,
  p_module text,
  p_status text DEFAULT 'success',
  p_error_message text DEFAULT NULL
)
LANGUAGE plpgsql
AS $$
BEGIN
  UPDATE tenant_module_log SET
    status = p_status,
    error_message = p_error_message
  WHERE tenant_id = p_tenant_id
    AND module_name = p_module
    AND status = 'pending'
    AND id = (
      SELECT id FROM tenant_module_log
      WHERE tenant_id = p_tenant_id AND module_name = p_module AND status = 'pending'
      ORDER BY provisioned_at DESC
      LIMIT 1
    );
END;
$$;

-- Step 5: Verification
SELECT 'Tenant registry schema created successfully!' as status;
SELECT COUNT(*) as tenant_count FROM tenant_registry;

\echo ''
\echo '============================================'
\echo 'Migration Complete!'
\echo '============================================'
\echo ''
