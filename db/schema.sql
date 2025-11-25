BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
--  OPERATIONS MODULE SCHEMA
-- ============================================================
-- 
-- This schema contains all tables needed for the Operations module
-- including shared lookups and dependencies
-- 

-- ============================================================
--  SHARED LOOKUPS
-- ============================================================

CREATE TABLE IF NOT EXISTS party_type_lu (
  type         text PRIMARY KEY,
  created_at   timestamptz,
  created_by   text,
  modified_at  timestamptz,
  modified_by  text,
  is_active    boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS classification_lu (
  id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  name         text NOT NULL UNIQUE,
  created_at   timestamptz,
  created_by   text,
  modified_at  timestamptz,
  modified_by  text,
  is_active    boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS customer_group_lu (
  id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  name         text NOT NULL UNIQUE,
  created_at   timestamptz,
  created_by   text,
  modified_at  timestamptz,
  modified_by  text,
  is_active    boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS company_type_lu (
  id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  name         text NOT NULL UNIQUE,
  created_at   timestamptz,
  created_by   text,
  modified_at  timestamptz,
  modified_by  text,
  is_active    boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS payment_term_lu (
  id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  code         text UNIQUE,
  name         text UNIQUE,
  days         int,
  created_at   timestamptz,
  created_by   text,
  modified_at  timestamptz,
  modified_by  text,
  is_active    boolean DEFAULT true
);

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

CREATE TABLE IF NOT EXISTS country_lu (
  country_id    serial PRIMARY KEY,
  country_name  text NOT NULL UNIQUE,
  country_code  char(3) NOT NULL UNIQUE,
  created_at    timestamptz DEFAULT now(),
  created_by    text,
  modified_at   timestamptz,
  modified_by   text,
  is_active     boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS state_lu (
  state_id     serial PRIMARY KEY,
  state_name   text NOT NULL,
  state_code   text,
  country_id   int NOT NULL REFERENCES country_lu(country_id),
  created_at   timestamptz DEFAULT now(),
  created_by   text,
  modified_at  timestamptz,
  modified_by  text,
  is_active    boolean DEFAULT true,
  UNIQUE (state_name, country_id),
  UNIQUE (state_code, country_id)
);

CREATE TABLE IF NOT EXISTS city_lu (
  city_id     serial PRIMARY KEY,
  city_name   text NOT NULL,
  city_code   text,
  state_id    int NOT NULL REFERENCES state_lu(state_id),
  created_at  timestamptz DEFAULT now(),
  created_by  text,
  modified_at timestamptz,
  modified_by text,
  is_active   boolean DEFAULT true,
  UNIQUE (city_name, state_id),
  UNIQUE (city_code, state_id)
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

CREATE TABLE IF NOT EXISTS network_lu (
  id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  name         text NOT NULL UNIQUE,
  network_type text,
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

CREATE TABLE IF NOT EXISTS criteria_lu (
  id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  name         text NOT NULL UNIQUE,
  created_at   timestamptz,
  created_by   text,
  modified_at  timestamptz,
  modified_by  text,
  is_active    boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS certification_lu (
  id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  name         text NOT NULL UNIQUE,
  created_at   timestamptz,
  created_by   text,
  modified_at  timestamptz,
  modified_by  text,
  is_active    boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS office_type_lu (
  code        text PRIMARY KEY,
  label       text NOT NULL UNIQUE,
  created_at  timestamptz DEFAULT now(),
  created_by  text,
  modified_at timestamptz,
  modified_by text,
  is_active   boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS shipping_mode_lu (
  name         text PRIMARY KEY,
  created_at   timestamptz DEFAULT now(),
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
--  EMPLOYEE MASTER
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

-- ============================================================
--  CS EXECUTIVE LOOKUP
-- ============================================================ 

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

-- ============================================================
--  SALES EXECUTIVE LOOKUP
-- ============================================================

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

-- ============================================================
--  PARTY MASTER
-- ============================================================

CREATE TABLE IF NOT EXISTS party_master (
  id                   uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  name                 text NOT NULL,
  classification_id    uuid NOT NULL REFERENCES classification_lu(id),
  company_type_id      uuid NOT NULL REFERENCES company_type_lu(id),
  sales_executive_id   uuid REFERENCES employee_master(id),
  sales_head_id        uuid REFERENCES employee_master(id),
  human_id             text UNIQUE,
  customer_group_id    uuid REFERENCES customer_group_lu(id),
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

CREATE TABLE IF NOT EXISTS party_role (
  party_id     uuid NOT NULL REFERENCES party_master(id) ON DELETE CASCADE,
  role_type    text NOT NULL REFERENCES party_type_lu(type),
  created_at   timestamptz DEFAULT now(),
  created_by   text,
  modified_at  timestamptz,
  modified_by  text,
  PRIMARY KEY (party_id, role_type)
);

CREATE TABLE IF NOT EXISTS party_address (
  id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  party_id     uuid NOT NULL REFERENCES party_master(id) ON DELETE CASCADE,
  street       text NOT NULL,
  city         text NOT NULL,
  state        text NOT NULL,
  country      text NOT NULL,
  postal_code  text NOT NULL,
  created_at   timestamptz DEFAULT now(),
  created_by   text,
  modified_at  timestamptz,
  modified_by  text,
  is_active    boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS party_contact_person (
  id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  party_id     uuid NOT NULL REFERENCES party_master(id) ON DELETE CASCADE,
  name         text NOT NULL,
  email        text NOT NULL,
  phone        text NOT NULL,
  designation  text,
  is_primary   boolean DEFAULT FALSE,
  created_at   timestamptz DEFAULT now(),
  created_by   text,
  modified_at  timestamptz,
  modified_by  text,
  is_active    boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS party_document (
  id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  party_id     uuid NOT NULL REFERENCES party_master(id) ON DELETE CASCADE,
  doc_type     text NOT NULL REFERENCES document_type_lu(code),
  doc_number   text,
  issued_on    date,
  expiry_on    date,
  file_key     text,
  file_region  text,
  display_name text,
  created_at   timestamptz DEFAULT now(),
  created_by   text,
  modified_at  timestamptz,
  modified_by  text,
  is_active    boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS party_finance (
  party_id                   uuid PRIMARY KEY REFERENCES party_master(id) ON DELETE CASCADE,
  credit_limit               numeric(14,2),
  credit_period_days         int,
  customs_duty_payment_limit numeric(14,2),
  payment_term_code          text REFERENCES payment_term_lu(code),
  preferred_currency_code    char(3) REFERENCES currency_lu(code),
  default_incoterm_code      text REFERENCES incoterm_lu(code),
  created_at                 timestamptz DEFAULT now(),
  created_by                 text,
  modified_at                timestamptz,
  modified_by                text,
  is_active                  boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS party_profile (
  party_id         uuid PRIMARY KEY REFERENCES party_master(id) ON DELETE CASCADE,
  company_id       text,
  office_type      text REFERENCES office_type_lu(code),
  years_in_business int,
  phone            text,
  mobile           text,
  email            text,
  website          text,
  created_at       timestamptz DEFAULT now(),
  created_by       text,
  modified_at      timestamptz,
  modified_by      text,
  is_active        boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS party_network (
  party_id     uuid NOT NULL REFERENCES party_master(id) ON DELETE CASCADE,
  network_id   uuid NOT NULL REFERENCES network_lu(id),
  created_at   timestamptz DEFAULT now(),
  created_by   text,
  modified_at  timestamptz,
  modified_by  text,
  is_active    boolean DEFAULT true,
  PRIMARY KEY (party_id, network_id)
);

CREATE TABLE IF NOT EXISTS party_criteria (
  party_id    uuid NOT NULL REFERENCES party_master(id) ON DELETE CASCADE,
  criteria_id uuid NOT NULL REFERENCES criteria_lu(id),
  created_at  timestamptz DEFAULT now(),
  created_by  text,
  modified_at timestamptz,
  modified_by text,
  is_active   boolean DEFAULT true,
  PRIMARY KEY (party_id, criteria_id)
);

CREATE TABLE IF NOT EXISTS party_certification (
  party_id          uuid NOT NULL REFERENCES party_master(id) ON DELETE CASCADE,
  certification_id  uuid NOT NULL REFERENCES certification_lu(id),
  created_at        timestamptz DEFAULT now(),
  created_by        text,
  modified_at       timestamptz,
  modified_by       text,
  is_active         boolean DEFAULT true,
  PRIMARY KEY (party_id, certification_id)
);

CREATE TABLE IF NOT EXISTS party_shipping_mode (
  party_id   uuid NOT NULL REFERENCES party_master(id) ON DELETE CASCADE,
  mode_id    text NOT NULL REFERENCES shipping_mode_lu(name),
  created_at timestamptz DEFAULT now(),
  created_by text,
  modified_at timestamptz,
  modified_by text,
  is_active  boolean DEFAULT true,
  PRIMARY KEY (party_id, mode_id)
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
