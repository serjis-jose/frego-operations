BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
--  OPERATIONS MODULE SCHEMA
-- ============================================================
-- 
-- DESIGN PRINCIPLE: Operations module can operate standalone
-- 
-- External References (Core/Backend Tables):
--   - All references to external tables (party_master, employee_master, branch_lu)
--     are stored as UUID fields WITHOUT foreign key constraints
--   - This allows operations to work independently even if core tables don't exist
--   - UUID values can be inserted/used without validation against external tables
--   - Comments indicate which external table each UUID references
-- 
-- Internal References (Operations Tables):
--   - References within operations module (e.g., ops_job, ops_package)
--     use proper foreign key constraints for data integrity
-- 
-- ============================================================
--  CORE LOOKUPS (Internal to Operations)
-- ============================================================

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

-- ============================================================
--  OPERATIONS MODULE TABLES
-- ============================================================

CREATE TABLE IF NOT EXISTS ops_job (
  id                  uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  job_code            text UNIQUE NOT NULL,
  enquiry_number      text,
  job_type            text,
  transport_mode      text,
  service_type        text,
  service_subcategory text,
  parent_job_id       uuid, -- REFERENCES ops_job(id) - Internal FK kept
  customer_id         uuid, -- REFERENCES party_master(id) - External: UUID only
  customer_name       text, -- Snapshot: party_master.name (for standalone display)
  agent_id            uuid, -- REFERENCES party_master(id) - External: UUID only
  agent_name          text, -- Snapshot: party_master.name (for standalone display)
  shipment_origin     text,
  destination_city    text,
  destination_state   text,
  destination_country text,
  source_city         text,
  source_state        text,
  source_country      text,
  branch_id           uuid, -- REFERENCES branch_lu(branch_id) - External: UUID only
  branch_name         text, -- Snapshot: branch_lu.branch_name (for standalone display)
  inco_term_code      text REFERENCES incoterm_lu(code),
  commodity           text,
  classification      text,
  sales_executive_id  uuid, -- REFERENCES employee_master(id) - External: UUID only
  sales_executive_name text, -- Snapshot: employee_master.name (for standalone display)
  operations_exec_id  uuid, -- REFERENCES employee_master(id) - External: UUID only
  operations_exec_name text, -- Snapshot: employee_master.name (for standalone display)
  cs_executive_id     uuid, -- REFERENCES employee_master(id) - External: UUID only
  cs_executive_name   text, -- Snapshot: employee_master.name (for standalone display)
  agent_deadline      timestamptz,
  shipment_ready_date timestamptz,
  status              text,
  priority_level      text,
  created_at          timestamptz DEFAULT now(),
  created_by          text,
  modified_at         timestamptz,
  modified_by         text,
  is_active           boolean DEFAULT true,
  
  -- Internal FK for parent job
  CONSTRAINT fk_parent_job FOREIGN KEY (parent_job_id) REFERENCES ops_job(id)
);

CREATE INDEX IF NOT EXISTS idx_ops_job_customer ON ops_job(customer_id);
CREATE INDEX IF NOT EXISTS idx_ops_job_agent ON ops_job(agent_id);
CREATE INDEX IF NOT EXISTS idx_ops_job_branch ON ops_job(branch_id);
CREATE INDEX IF NOT EXISTS idx_ops_job_status ON ops_job(status);

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

CREATE INDEX IF NOT EXISTS idx_ops_package_job ON ops_package(job_id);

CREATE TABLE IF NOT EXISTS ops_carrier (
  id                    uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  job_id                uuid REFERENCES ops_job(id) ON DELETE CASCADE,
  carrier_party_id      uuid, -- REFERENCES party_master(id) - External: UUID only
  carrier_name          text, -- Snapshot: party_master.name (for standalone display)
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

CREATE INDEX IF NOT EXISTS idx_ops_carrier_job ON ops_carrier(job_id);

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

CREATE INDEX IF NOT EXISTS idx_ops_job_document_job ON ops_job_document(job_id);

CREATE TABLE IF NOT EXISTS ops_billing (
  id                      uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  job_id                  uuid REFERENCES ops_job(id) ON DELETE CASCADE,
  activity_type           text,
  activity_code           text,
  billing_party_id        uuid, -- REFERENCES party_master(id) - External: UUID only
  billing_party_name      text, -- Snapshot: party_master.name (for standalone display)
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

CREATE INDEX IF NOT EXISTS idx_ops_billing_job ON ops_billing(job_id);

CREATE TABLE IF NOT EXISTS ops_provision (
  id                      uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  job_id                  uuid REFERENCES ops_job(id) ON DELETE CASCADE,
  activity_type           text,
  activity_code           text,
  cost_party_id           uuid, -- REFERENCES party_master(id) - External: UUID only
  cost_party_name         text, -- Snapshot: party_master.name (for standalone display)
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

CREATE INDEX IF NOT EXISTS idx_ops_provision_job ON ops_provision(job_id);

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

CREATE INDEX IF NOT EXISTS idx_ops_tracking_job ON ops_tracking(job_id);

-- ============================================================
--  LOOKUP TABLES FOR OPERATIONS
-- ============================================================

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

COMMIT;
