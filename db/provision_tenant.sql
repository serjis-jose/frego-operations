BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
--  OPERATIONS TENANT SCHEMA PROVISIONING
-- ============================================================

CREATE OR REPLACE PROCEDURE ensure_operations_tenant_schema(
  p_tenant_id uuid DEFAULT NULL,
  p_schema text DEFAULT NULL,
  p_grant_role text DEFAULT 'erp_user'
)
LANGUAGE plpgsql
AS $$
DECLARE
  tenant_uuid    uuid := COALESCE(p_tenant_id, uuid_generate_v4());
  tenant_id_text text := trim(both from tenant_uuid::text);
  schema_input   text := NULLIF(trim(both from p_schema), '');
  tenant_schema  text;
BEGIN
  IF tenant_uuid IS NULL THEN
    RAISE EXCEPTION 'Tenant identifier must be provided.';
  END IF;

  -- Determine schema name
  IF schema_input IS NOT NULL THEN
    IF schema_input LIKE 'fin_%' THEN
      tenant_schema := 'ops_' || substring(schema_input from 5);
    ELSE
      tenant_schema := schema_input;
    END IF;
  ELSE
    tenant_schema := 'ops_' || regexp_replace(lower(tenant_id_text), '[^a-z0-9_]', '_', 'g');
  END IF;

  -- Create Schema
  EXECUTE format('CREATE SCHEMA IF NOT EXISTS %I', tenant_schema);
  EXECUTE format('SET search_path TO %I, public', tenant_schema);

  -- Create Tables
  EXECUTE format($ddl$

  -- Dependencies (Stubbed or Full for Joins)
  -- We include necessary tables for joins to work. 
  -- In a real microservice with separate DB, these would be synced or removed.

  CREATE TABLE IF NOT EXISTS currency_lu (
    id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    code         char(3) NOT NULL UNIQUE,
    name         text,
    created_at   timestamptz,
    created_by   text,
    modified_at  timestamptz,
    modified_by  text,
    is_active    boolean DEFAULT true
  );

  CREATE TABLE IF NOT EXISTS branch_lu (
    branch_id    uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    branch_name  text NOT NULL UNIQUE,
    is_active    boolean DEFAULT true
  );

  CREATE TABLE IF NOT EXISTS incoterm_lu (
    id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    code         text NOT NULL UNIQUE,
    name         text,
    version      smallint DEFAULT 2020,
    created_at   timestamptz,
    created_by   text,
    modified_at  timestamptz,
    modified_by  text,
    is_active    boolean DEFAULT true
  );

  CREATE TABLE IF NOT EXISTS document_type_lu (
    code         text PRIMARY KEY,
    label        text NOT NULL UNIQUE,
    created_at   timestamptz,
    created_by   text,
    modified_at  timestamptz,
    modified_by  text,
    is_active    boolean DEFAULT true
  );

  CREATE TABLE IF NOT EXISTS employee_master (
    id          uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        text NOT NULL,
    email       text NOT NULL UNIQUE,
    role        text NOT NULL,
    created_at  timestamptz,
    created_by  text,
    modified_at timestamptz,
    modified_by text,
    is_active   boolean DEFAULT true
  );

  CREATE TABLE IF NOT EXISTS cs_executive_lu (
    cs_exec_id   uuid PRIMARY KEY REFERENCES employee_master(id) ON DELETE CASCADE,
    cs_exec_name text NOT NULL,
    branch_id    uuid REFERENCES branch_lu(branch_id),
    created_at   timestamptz DEFAULT now(),
    created_by   text,
    modified_at  timestamptz,
    modified_by  text,
    is_active    boolean DEFAULT true
  );

  CREATE TABLE IF NOT EXISTS sales_executive_lu (
    sales_exec_id   uuid PRIMARY KEY REFERENCES employee_master(id) ON DELETE CASCADE,
    sales_exec_name text NOT NULL,
    branch_id       uuid REFERENCES branch_lu(branch_id),
    created_at      timestamptz DEFAULT now(),
    created_by      text,
    modified_at     timestamptz,
    modified_by     text,
    is_active       boolean DEFAULT true
  );

  CREATE TABLE IF NOT EXISTS party_master (
    id                   uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                 text NOT NULL,
    classification_id    uuid, -- Removed FK
    company_type_id      uuid, -- Removed FK
    sales_executive_id   uuid REFERENCES employee_master(id),
    sales_head_id        uuid REFERENCES employee_master(id),
    human_id             text UNIQUE,
    customer_group_id    uuid, -- Removed FK
    business_nature      text,
    vat_tax_number       text,
    trade_license_no     text,
    trade_license_expiry date,
    owner_id_type        text,
    owner_id_number      text,
    owner_id_expiry      date,
    status               text DEFAULT 'Active',
    created_at           timestamptz DEFAULT now(),
    created_by           text,
    modified_at          timestamptz,
    modified_by          text,
    is_active            boolean DEFAULT true
  );

  -- Operations Lookups

  CREATE TABLE IF NOT EXISTS trans_move_service_lu (
    transport_mode_id      smallint NOT NULL,
    transport_mode_name    text,
    movement_type_id       smallint NOT NULL,
    movement_type_name     text,
    service_type_id        smallint NOT NULL,
    service_type_name      text,
    service_subcategory_id smallint NOT NULL,
    service_subcategory_name text,
    PRIMARY KEY (transport_mode_id, movement_type_id, service_type_id, service_subcategory_id)
  );

  CREATE TABLE IF NOT EXISTS job_status_lu (
    job_status_id   smallint PRIMARY KEY,
    job_status_name text,
    job_status_desc text
  );

  CREATE TABLE IF NOT EXISTS document_status_lu (
    doc_status_id   smallint PRIMARY KEY,
    doc_status_name text,
    doc_status_desc text
  );

  CREATE TABLE IF NOT EXISTS role_details_lu (
    role_id     smallint PRIMARY KEY,
    role_name   text,
    role_desc   text
  );

  CREATE TABLE IF NOT EXISTS priority_lu (
    priority_id     smallint PRIMARY KEY,
    priority_label  text 
  );

  -- Operations Tables

  CREATE TABLE IF NOT EXISTS ops_job (
    id                  uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_code            text UNIQUE NOT NULL,
    enquiry_number      text,
    job_type            text,
    transport_mode      text,
    service_type        text,
    service_subcategory text,
    parent_job_id       uuid REFERENCES ops_job(id),
    customer_id         uuid REFERENCES party_master(id),
    agent_id            uuid REFERENCES party_master(id),
    shipment_origin     text,
    destination_city    text,
    destination_state   text,
    destination_country text,
    source_city         text,
    source_state        text,
    source_country      text,
    branch_id           uuid REFERENCES branch_lu(branch_id),
    inco_term_code      text REFERENCES incoterm_lu(code),
    commodity           text,
    classification      text,
    sales_executive_id  uuid REFERENCES employee_master(id),
    operations_exec_id  uuid REFERENCES employee_master(id),
    cs_executive_id     uuid REFERENCES cs_executive_lu(cs_exec_id),
    agent_deadline      timestamptz,
    shipment_ready_date timestamptz,
    status              text,
    priority_level      text,
    created_at          timestamptz DEFAULT now(),
    created_by          text,
    modified_at         timestamptz,
    modified_by         text,
    is_active           boolean DEFAULT true
  );

  CREATE TABLE IF NOT EXISTS ops_package (
    id              uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id          uuid NOT NULL REFERENCES ops_job(id) ON DELETE CASCADE,
    package_name    text,
    package_type    text,
    quantity        int,
    length_meters   numeric(8,3),
    width_meters    numeric(8,3),
    height_meters   numeric(8,3),
    weight_kg       numeric(10,2),
    volume_cbm      numeric(10,3),
    hs_code         text,
    cargo_type      text,
    container_id    text,
    notes           text,
    created_at      timestamptz DEFAULT now(),
    created_by      text,
    modified_at     timestamptz,
    modified_by     text,
    is_active       boolean DEFAULT true
  );

  CREATE TABLE IF NOT EXISTS ops_carrier (
    id                    uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id                uuid REFERENCES ops_job(id) ON DELETE CASCADE,
    carrier_party_id      uuid REFERENCES party_master(id),
    carrier_name          text,
    vessel_name           text,
    voyage_number         text,
    flight_id             text,
    flight_date           date,
    vehicle_number        text,
    route_details         text,
    driver_name           text,
    origin_port_station   text,
    destination_port_station text,
    accounting_info       text,
    handling_info         text,
    supporting_doc_url    text,
    created_at            timestamptz DEFAULT now(),
    created_by            text,
    modified_at           timestamptz,
    modified_by           text,
    is_active             boolean DEFAULT true
  );

  CREATE TABLE IF NOT EXISTS ops_job_document (
    id              uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id          uuid NOT NULL REFERENCES ops_job(id) ON DELETE CASCADE,
    doc_type_code   text REFERENCES document_type_lu(code),
    doc_number      text,
    issued_at       text,
    issued_date     timestamptz,
    description     text,
    file_key        text,
    file_region     text,
    created_at      timestamptz DEFAULT now(),
    created_by      text,
    modified_at     timestamptz,
    modified_by     text,
    is_active       boolean DEFAULT true
  );

  CREATE TABLE IF NOT EXISTS ops_billing (
    id                      uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id                  uuid REFERENCES ops_job(id) ON DELETE CASCADE,
    activity_type           text,
    activity_code           text,
    billing_party_id        uuid REFERENCES party_master(id),
    po_number               text,
    po_date                 date,
    currency_code           char(3) REFERENCES currency_lu(code),
    amount                  numeric(14,2),
    description             text,
    notes                   text,
    supporting_doc_url      text,
    amount_primary_currency numeric(14,2),
    created_at              timestamptz DEFAULT now(),
    created_by              text,
    modified_at             timestamptz,
    modified_by             text,
    is_active               boolean DEFAULT true
  );

  CREATE TABLE IF NOT EXISTS ops_provision (
    id                      uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id                  uuid REFERENCES ops_job(id) ON DELETE CASCADE,
    activity_type           text,
    activity_code           text,
    cost_party_id           uuid REFERENCES party_master(id),
    invoice_number          text,
    invoice_date            date,
    currency_code           char(3) REFERENCES currency_lu(code),
    amount                  numeric(14,2),
    payment_priority        text,
    notes                   text,
    supporting_doc_url      text,
    amount_primary_currency numeric(14,2),
    profit                  numeric(14,2),
    created_at              timestamptz DEFAULT now(),
    created_by              text,
    modified_at             timestamptz,
    modified_by             text,
    is_active               boolean DEFAULT true
  );

  CREATE TABLE IF NOT EXISTS ops_tracking (
    id               uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id           uuid REFERENCES ops_job(id) ON DELETE CASCADE,
    etd_date         timestamptz,
    eta_date         timestamptz,
    atd_date         timestamptz,
    ata_date         timestamptz,
    job_status       text,
    pod_status       text,
    document_status  text,
    notes            text,
    created_at       timestamptz DEFAULT now(),
    created_by       text,
    modified_at      timestamptz,
    modified_by      text,
    is_active        boolean DEFAULT true
  );

  $ddl$);

  -- Grant usage
  EXECUTE format('GRANT USAGE ON SCHEMA %I TO %I', tenant_schema, p_grant_role);
  EXECUTE format('GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA %I TO %I', tenant_schema, p_grant_role);
  EXECUTE format('GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA %I TO %I', tenant_schema, p_grant_role);

END;
$$;

COMMIT;
