BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Lookup tables
CREATE TABLE IF NOT EXISTS trans_move_service_lu (
  transport_mode_id        smallint,
  transport_mode_name      text,
  movement_type_id         smallint,
  movement_type_name       text,
  service_type_id          smallint,
  service_type_name        text,
  service_subcategory_id   smallint,
  service_subcategory_name text
);

CREATE TABLE IF NOT EXISTS job_status_lu (
  job_status_id   smallint,
  job_status_name text,
  job_status_desc text
);

CREATE TABLE IF NOT EXISTS document_status_lu (
  doc_status_id   smallint,
  doc_status_name text,
  doc_status_desc text
);

CREATE TABLE IF NOT EXISTS role_details_lu (
  role_id     smallint,
  role_name   text,
  role_desc   text
);

CREATE TABLE IF NOT EXISTS priority_lu (
  priority_id    smallint PRIMARY KEY,
  priority_label text NOT NULL
);

CREATE TABLE IF NOT EXISTS branch_lu (
  branch_id   uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  branch_name text NOT NULL UNIQUE,
  is_active   boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS sales_executive_lu (
  sales_exec_id   uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  sales_exec_name text NOT NULL,
  branch_id       uuid REFERENCES branch_lu(branch_id),
  created_at      timestamptz DEFAULT now(),
  created_by      text,
  modified_at     timestamptz,
  modified_by     text,
  is_active       boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS cs_executive_lu (
  cs_exec_id   uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  cs_exec_name text NOT NULL,
  branch_id    uuid REFERENCES branch_lu(branch_id),
  created_at   timestamptz DEFAULT now(),
  created_by   text,
  modified_at  timestamptz,
  modified_by  text,
  is_active    boolean DEFAULT true
);

-- Core reference tables used in joins
CREATE TABLE IF NOT EXISTS party_master (
  id                   uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  name                 text NOT NULL,
  classification_id    uuid,
  company_type_id      uuid,
  sales_executive_id   uuid,
  sales_head_id        uuid,
  human_id             text,
  customer_group_id    uuid,
  business_nature      text,
  vat_tax_number       text,
  trade_license_no     text,
  trade_license_expiry date,
  owner_id_type        text,
  owner_id_number      text,
  owner_id_expiry      date,
  status               text DEFAULT 'Active',
  created_at           timestamptz,
  created_by           text,
  modified_at          timestamptz,
  modified_by          text,
  is_active            boolean DEFAULT true
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

-- Operations tables
CREATE TABLE IF NOT EXISTS ops_job (
  id                   uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  job_code             text NOT NULL UNIQUE,
  enquiry_number       text,
  job_type             text,
  transport_mode       text,
  service_type         text,
  service_subcategory  text,
  parent_job_id        uuid,
  customer_id          uuid REFERENCES party_master(id),
  agent_id             uuid REFERENCES party_master(id),
  shipment_origin      text,
  destination_city     text,
  destination_state    text,
  destination_country  text,
  source_city          text,
  source_state         text,
  source_country       text,
  branch_id            uuid REFERENCES branch_lu(branch_id),
  inco_term_code       text,
  commodity            text,
  classification       text,
  sales_executive_id   uuid REFERENCES employee_master(id),
  operations_exec_id   uuid REFERENCES employee_master(id),
  cs_executive_id      uuid REFERENCES employee_master(id),
  agent_deadline       timestamptz,
  shipment_ready_date  timestamptz,
  status               text,
  priority_level       smallint,
  created_at           timestamptz DEFAULT now(),
  created_by           text,
  modified_at          timestamptz,
  modified_by          text,
  is_active            boolean DEFAULT true
);

CREATE INDEX IF NOT EXISTS idx_ops_job_job_code ON ops_job(job_code);

CREATE TABLE IF NOT EXISTS ops_package (
  id             uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  job_id         uuid NOT NULL REFERENCES ops_job(id) ON DELETE CASCADE,
  package_name   text,
  package_type   text,
  quantity       int,
  length_meters  numeric,
  width_meters   numeric,
  height_meters  numeric,
  weight_kg      numeric,
  volume_cbm     numeric,
  hs_code        text,
  cargo_type     text,
  container_id   text,
  notes          text,
  created_at     timestamptz DEFAULT now(),
  created_by     text,
  modified_at    timestamptz,
  modified_by    text,
  is_active      boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS ops_carrier (
  id                     uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  job_id                 uuid NOT NULL REFERENCES ops_job(id) ON DELETE CASCADE,
  carrier_party_id       uuid,
  carrier_name           text,
  vessel_name            text,
  voyage_number          text,
  flight_id              text,
  flight_date            date,
  vehicle_number         text,
  route_details          text,
  driver_name            text,
  origin_port_station    text,
  destination_port_station text,
  accounting_info        text,
  handling_info          text,
  supporting_doc_url     text,
  created_at             timestamptz DEFAULT now(),
  created_by             text,
  modified_at            timestamptz,
  modified_by            text,
  is_active              boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS ops_job_document (
  id            uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  job_id        uuid NOT NULL REFERENCES ops_job(id) ON DELETE CASCADE,
  doc_type_code text,
  doc_number    text,
  issued_at     timestamptz,
  issued_date   date,
  description   text,
  file_key      text,
  file_region   text,
  created_at    timestamptz DEFAULT now(),
  created_by    text,
  modified_at   timestamptz,
  modified_by   text,
  is_active     boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS ops_billing (
  id                      uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  job_id                  uuid NOT NULL REFERENCES ops_job(id) ON DELETE CASCADE,
  activity_type           text,
  activity_code           text,
  billing_party_id        uuid REFERENCES party_master(id),
  po_number               text,
  po_date                 date,
  currency_code           char(3),
  amount                  numeric,
  description             text,
  notes                   text,
  supporting_doc_url      text,
  amount_primary_currency numeric,
  created_at              timestamptz DEFAULT now(),
  created_by              text,
  modified_at             timestamptz,
  modified_by             text,
  is_active               boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS ops_provision (
  id                      uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  job_id                  uuid NOT NULL REFERENCES ops_job(id) ON DELETE CASCADE,
  activity_type           text,
  activity_code           text,
  cost_party_id           uuid REFERENCES party_master(id),
  invoice_number          text,
  invoice_date            date,
  currency_code           char(3),
  amount                  numeric,
  payment_priority        text,
  notes                   text,
  supporting_doc_url      text,
  amount_primary_currency numeric,
  profit                  numeric,
  created_at              timestamptz DEFAULT now(),
  created_by              text,
  modified_at             timestamptz,
  modified_by             text,
  is_active               boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS ops_tracking (
  id               uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  job_id           uuid NOT NULL UNIQUE REFERENCES ops_job(id) ON DELETE CASCADE,
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

COMMIT;
