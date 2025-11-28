BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
--  LOOKUP TABLES
-- ============================================================

CREATE TABLE IF NOT EXISTS trans_move_service_lu (
  transport_mode_id        smallint NOT NULL,
  transport_mode_name      text,
  movement_type_id         smallint NOT NULL,
  movement_type_name       text,
  service_type_id          smallint NOT NULL,
  service_type_name        text,
  service_subcategory_id   smallint NOT NULL,
  service_subcategory_name text,
  created_at               timestamptz,
  created_by               text,
  modified_at              timestamptz,
  modified_by              text,
  is_active                boolean DEFAULT true,
  PRIMARY KEY (transport_mode_id, movement_type_id, service_type_id, service_subcategory_id)
);

CREATE TABLE IF NOT EXISTS job_status_lu (
  job_status_id   smallint PRIMARY KEY,
  job_status_name text,
  job_status_desc text,
  created_at      timestamptz,
  created_by      text,
  modified_at     timestamptz,
  modified_by     text,
  is_active       boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS document_status_lu (
  doc_status_id   smallint PRIMARY KEY,
  doc_status_name text,
  doc_status_desc text,
  created_at      timestamptz,
  created_by      text,
  modified_at     timestamptz,
  modified_by     text,
  is_active       boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS role_details_lu (
  role_id     smallint PRIMARY KEY,
  role_name   text,
  role_desc   text,
  created_at  timestamptz,
  created_by  text,
  modified_at timestamptz,
  modified_by text,
  is_active   boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS priority_lu (
  priority_id     smallint PRIMARY KEY,
  priority_label  text,
  created_at      timestamptz,
  created_by      text,
  modified_at     timestamptz,
  modified_by     text,
  is_active       boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS branch_lu (
  branch_id    uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  branch_name  text NOT NULL UNIQUE,
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
--  EMPLOYEE MASTER (Soft reference - no FK to external DB)
-- ============================================================

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
  cs_exec_id   uuid PRIMARY KEY,
  cs_exec_name text NOT NULL,
  branch_id    uuid REFERENCES branch_lu(branch_id),
  created_at   timestamptz DEFAULT now(),
  created_by   text,
  modified_at  timestamptz,
  modified_by  text,
  is_active    boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS sales_executive_lu (
  sales_exec_id   uuid PRIMARY KEY,
  sales_exec_name text NOT NULL,
  branch_id       uuid REFERENCES branch_lu(branch_id),
  created_at      timestamptz DEFAULT now(),
  created_by      text,
  modified_at     timestamptz,
  modified_by     text,
  is_active       boolean DEFAULT true
);

-- ============================================================
--  PARTY MASTER (Soft reference - no FK to external DB)
-- ============================================================

CREATE TABLE IF NOT EXISTS party_master (
  id                   uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  name                 text NOT NULL,
  classification_id    uuid,
  company_type_id      uuid,
  sales_executive_id   uuid,
  sales_head_id        uuid,
  human_id             text UNIQUE,
  customer_group_id    uuid,
  business_nature      text,
  vat_tax_number       text,
  trade_license_no     text,
  trade_license_expiry date,
  owner_id_type        text,
  owner_id_number      text,
  owner_id_expiry      date,
  status               text CHECK (status IN ('Active','Inactive','Blacklisted')) DEFAULT 'Active',
  created_at           timestamptz DEFAULT now(),
  created_by           text,
  modified_at          timestamptz,
  modified_by          text,
  is_active            boolean DEFAULT true
);

-- ============================================================
--  OPERATIONS MODULE
-- ============================================================

CREATE TABLE IF NOT EXISTS ops_job (
  id                  uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  job_code            text UNIQUE NOT NULL,
  enquiry_number      text,
  job_type            text,
  transport_mode      text,
  service_type        text,
  service_subcategory text,
  parent_job_id       uuid REFERENCES ops_job(id),
  customer_id         uuid,
  agent_id            uuid,
  shipment_origin     text,
  destination_city    text,
  destination_state   text,
  destination_country text,
  source_city         text,
  source_state        text,
  source_country      text,
  branch_id           uuid REFERENCES branch_lu(branch_id),
  inco_term_code      text,
  commodity           text,
  classification      text,
  sales_executive_id  uuid,
  operations_exec_id  uuid,
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

CREATE INDEX IF NOT EXISTS idx_ops_job_job_code ON ops_job(job_code);

CREATE TABLE IF NOT EXISTS ops_package (
  id                         uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  job_id                     uuid NOT NULL REFERENCES ops_job(id) ON DELETE CASCADE,
  container_no               text,
  container_name             text,
  container_size             text,
  gross_weight_kg            numeric(10,3),
  net_weight_kg              numeric(10,3),
  volume                     numeric(10,3),
  carrier_seal_no            text,
  commodity_cargo_description text,
  package_type               text,
  cargo_type                 text,
  no_of_packages             numeric(20,2),
  chargeable_weight          numeric(10,3),
  hs_code                    text,
  temperature_control        boolean DEFAULT false,
  created_at                 timestamptz DEFAULT now(),
  created_by                 text,
  modified_at                timestamptz,
  modified_by                text,
  is_active                  boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS ops_carrier (
  id                    uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  job_id                uuid REFERENCES ops_job(id) ON DELETE CASCADE,
  carrier_party_id      uuid,
  carrier_name          text,
  carrier_contact       text,
  vessel_name           text,
  voyage_number         text,
  flight_id             text,
  flight_date           date,
  airport_report_date   timestamptz,
  vehicle_number        text,
  vehicle_type          text,
  route_details         text,
  driver_name           text,
  driver_contact        text,
  origin_port_station   text,
  destination_port_station text,
  origin_country        text,
  destination_country   text,
  accounting_info       text,
  handling_info         text,
  transport_document_reference text,
  supporting_doc_url    text[],
  file_region           text,
  description           text,
  created_at            timestamptz DEFAULT now(),
  created_by            text,
  modified_at           timestamptz,
  modified_by           text,
  is_active             boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS ops_party (
  job_id                     uuid PRIMARY KEY REFERENCES ops_job(id) ON DELETE CASCADE,
  shipper_id                 uuid,
  consignee_id               uuid,
  notify_party_id            uuid,
  switch_bl_shipper_id       uuid,
  switch_bl_consignee_id     uuid,
  switch_bl_notify_party_id  uuid,
  origin_agent_id            uuid,
  destination_agent_id       uuid,
  created_at                 timestamptz DEFAULT now(),
  created_by                 text,
  modified_at                timestamptz,
  modified_by                text,
  is_active                  boolean DEFAULT true
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
  supporting_doc_urls text[],
  bl_awb_uploads  text[],
  house_doc_number text,
  house_issued_at text,
  house_issued_date timestamptz,
  house_description text,
  partial_bl_number text,
  switch_bl_awb_number text,
  switch_bl_awb_issued_at text,
  switch_bl_awb_issued_date timestamptz,
  switch_bl_awb_description text,
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
  billing_party_id        uuid,
  po_number               text,
  po_date                 date,
  currency_code           char(3),
  amount                  numeric(14,2),
  description             text,
  notes                   text,
  supporting_doc_url      text[],
  file_region             text,
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
  cost_party_id           uuid,
  invoice_number          text,
  invoice_date            date,
  currency_code           char(3),
  amount                  numeric(14,2),
  payment_priority        text,
  notes                   text,
  supporting_doc_url      text[],
  file_region             text,
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
  pod_doc_urls     text[],
  file_region      text,
  document_status  text,
  notes            text,
  created_at       timestamptz DEFAULT now(),
  created_by       text,
  modified_at      timestamptz,
  modified_by      text,
  is_active        boolean DEFAULT true
);

-- ============================================================
--  ORDERS (PRICING TOOL)
-- ============================================================

CREATE TABLE IF NOT EXISTS orders (
  order_id             text PRIMARY KEY,
  sales_team           text,
  sales_person         text,
  mode                 text,
  customer_name        text,
  classification       text,
  agent_deadline       timestamptz,
  comments             text,
  commodity            text,
  destination_city     text,
  destination_country  text,
  incoterm             text,
  package_list         jsonb,
  pickup_address       text,
  shipment_ready_date  timestamptz,
  shipment_type        text,
  source_city          text,
  source_country       text,
  status               text,
  order_created_on     timestamptz,
  order_created_by     text,
  created_at           timestamptz NOT NULL DEFAULT now(),
  created_by           text,
  is_deleted           boolean       NOT NULL DEFAULT false,
  template_name        text,
  email_logs           jsonb,
  negotiated_quotes    jsonb,
  modified_at          timestamptz,
  modified_by          text,
  is_active            boolean       NOT NULL DEFAULT true
);

COMMIT;
