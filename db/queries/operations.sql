-- ============================================================
-- LOOKUP QUERIES
-- ============================================================

-- name: ListTransportModeServiceLookups :many
SELECT
    transport_mode_id,
    transport_mode_name,
    movement_type_id,
    movement_type_name,
    service_type_id,
    service_type_name,
    service_subcategory_id,
    service_subcategory_name
FROM trans_move_service_lu
ORDER BY transport_mode_id, movement_type_id, service_type_id, service_subcategory_id;

-- name: ListJobStatusLookups :many
SELECT
    job_status_id,
    job_status_name,
    job_status_desc
FROM job_status_lu
ORDER BY job_status_id;

-- name: ListDocumentStatusLookups :many
SELECT
    doc_status_id,
    doc_status_name,
    doc_status_desc
FROM document_status_lu
ORDER BY doc_status_id;

-- name: ListRoleDetailsLookups :many
SELECT
    role_id,
    role_name,
    role_desc
FROM role_details_lu
ORDER BY role_id;

-- name: ListPriorityLookups :many
SELECT
    priority_id,
    priority_label
FROM priority_lu
ORDER BY priority_id;



-- ============================================================
-- JOB CRUD QUERIES
-- ============================================================

-- name: GetNextJobSequence :one
SELECT COALESCE(MAX(
    CASE 
        WHEN job_code ~ ('^' || sqlc.arg(prefix) || '-[0-9]+$')
        THEN CAST(SUBSTRING(job_code FROM LENGTH(sqlc.arg(prefix)) + 2) AS INTEGER)
        ELSE 0
    END
), 0) + 1 AS next_seq
FROM ops_job
WHERE job_code LIKE sqlc.arg(prefix) || '%';

-- name: ListJobs :many
SELECT
    j.id,
    j.job_code,
    j.enquiry_number,
    j.job_type,
    j.transport_mode,
    j.service_type,
    j.customer_id,
    cust.name AS customer_name,
    j.agent_id,
    agent.name AS agent_name,
    j.shipment_origin,
    j.destination_city,
    j.destination_state,
    j.destination_country,
    j.source_city,
    j.source_state,
    j.source_country,
    j.status,
    j.priority_level,
    j.sales_executive_id,
    se.name AS sales_executive_name,
    se.email AS sales_executive_email,
    se.role AS sales_executive_role,
    j.operations_exec_id,
    oe.name AS operations_exec_name,
    oe.email AS operations_exec_email,
    oe.role AS operations_exec_role,
    j.created_at,
    j.modified_at,
    j.is_active
FROM ops_job j
LEFT JOIN party_master cust ON cust.id = j.customer_id
LEFT JOIN party_master agent ON agent.id = j.agent_id
LEFT JOIN employee_master se ON se.id = j.sales_executive_id
LEFT JOIN employee_master oe ON oe.id = j.operations_exec_id
WHERE j.is_active
  AND (sqlc.narg(status)::text IS NULL OR j.status = sqlc.narg(status))
  AND (sqlc.narg(customer_id)::uuid IS NULL OR j.customer_id = sqlc.narg(customer_id))
  AND (sqlc.narg(job_type)::text IS NULL OR j.job_type = sqlc.narg(job_type))
ORDER BY j.created_at DESC
LIMIT sqlc.arg(row_limit);

-- name: GetJob :one
SELECT
    j.id,
    j.job_code,
    j.enquiry_number,
    j.job_type,
    j.transport_mode,
    j.service_type,
    j.service_subcategory,
    j.parent_job_id,
    j.customer_id,
    cust.name AS customer_name,
    j.agent_id,
    agent.name AS agent_name,
    j.shipment_origin,
    j.destination_city,
    j.destination_state,
    j.destination_country,
    j.source_city,
    j.source_state,
    j.source_country,
    j.branch_id,
    j.branch_name,
    j.inco_term_code,
    j.commodity,
    j.classification,
    j.sales_executive_id,
    se.name AS sales_executive_name,
    se.email AS sales_executive_email,
    se.role AS sales_executive_role,
    j.operations_exec_id,
    oe.name AS operations_exec_name,
    oe.email AS operations_exec_email,
    oe.role AS operations_exec_role,
    j.cs_executive_id,
    ce.name AS cs_executive_name,
    ce.email AS cs_executive_email,
    ce.role AS cs_executive_role,
    j.agent_deadline,
    j.shipment_ready_date,
    j.status,
    j.priority_level,
    j.created_at,
    j.created_by,
    j.modified_at,
    j.modified_by,
    j.is_active
FROM ops_job j
LEFT JOIN party_master cust ON cust.id = j.customer_id
LEFT JOIN party_master agent ON agent.id = j.agent_id

LEFT JOIN employee_master se ON se.id = j.sales_executive_id
LEFT JOIN employee_master oe ON oe.id = j.operations_exec_id
LEFT JOIN employee_master ce ON ce.id = j.cs_executive_id
WHERE j.id = sqlc.arg(id);

-- name: CreateJob :one
INSERT INTO ops_job (
    job_code,
    enquiry_number,
    job_type,
    transport_mode,
    service_type,
    service_subcategory,
    parent_job_id,
    customer_id,
    agent_id,
    shipment_origin,
    destination_city,
    destination_state,
    destination_country,
    source_city,
    source_state,
    source_country,
    branch_id,
    branch_name,
    inco_term_code,
    commodity,
    classification,
    sales_executive_id,
    sales_executive_name,
    operations_exec_id,
    operations_exec_name,
    cs_executive_id,
    cs_executive_name,
    agent_deadline,
    shipment_ready_date,
    status,
    priority_level,
    created_at,
    created_by,
    is_active
) VALUES (
    sqlc.arg(job_code),
    sqlc.narg(enquiry_number),
    sqlc.narg(job_type),
    sqlc.narg(transport_mode),
    sqlc.narg(service_type),
    sqlc.narg(service_subcategory),
    sqlc.narg(parent_job_id),
    sqlc.narg(customer_id),
    sqlc.narg(agent_id),
    sqlc.narg(shipment_origin),
    sqlc.narg(destination_city),
    sqlc.narg(destination_state),
    sqlc.narg(destination_country),
    sqlc.narg(source_city),
    sqlc.narg(source_state),
    sqlc.narg(source_country),
    sqlc.narg(branch_id),
    sqlc.narg(branch_name),
    sqlc.narg(inco_term_code),
    sqlc.narg(commodity),
    sqlc.narg(classification),
    sqlc.narg(sales_executive_id),
    sqlc.narg(sales_executive_name),
    sqlc.narg(operations_exec_id),
    sqlc.narg(operations_exec_name),
    sqlc.narg(cs_executive_id),
    sqlc.narg(cs_executive_name),
    sqlc.narg(agent_deadline),
    sqlc.narg(shipment_ready_date),
    sqlc.narg(status),
    sqlc.narg(priority_level),
    now(),
    sqlc.arg(actor),
    true
) RETURNING *;

-- name: UpdateJob :one
UPDATE ops_job
SET
    enquiry_number = COALESCE(sqlc.narg(enquiry_number), enquiry_number),
    job_type = COALESCE(sqlc.narg(job_type), job_type),
    transport_mode = COALESCE(sqlc.narg(transport_mode), transport_mode),
    service_type = COALESCE(sqlc.narg(service_type), service_type),
    service_subcategory = COALESCE(sqlc.narg(service_subcategory), service_subcategory),
    parent_job_id = COALESCE(sqlc.narg(parent_job_id), parent_job_id),
    customer_id = COALESCE(sqlc.narg(customer_id), customer_id),
    agent_id = COALESCE(sqlc.narg(agent_id), agent_id),
    shipment_origin = COALESCE(sqlc.narg(shipment_origin), shipment_origin),
    destination_city = COALESCE(sqlc.narg(destination_city), destination_city),
    destination_state = COALESCE(sqlc.narg(destination_state), destination_state),
    destination_country = COALESCE(sqlc.narg(destination_country), destination_country),
    source_city = COALESCE(sqlc.narg(source_city), source_city),
    source_state = COALESCE(sqlc.narg(source_state), source_state),
    source_country = COALESCE(sqlc.narg(source_country), source_country),
    branch_id = COALESCE(sqlc.narg(branch_id), branch_id),
    branch_name = COALESCE(sqlc.narg(branch_name), branch_name),
    inco_term_code = COALESCE(sqlc.narg(inco_term_code), inco_term_code),
    commodity = COALESCE(sqlc.narg(commodity), commodity),
    classification = COALESCE(sqlc.narg(classification), classification),
    sales_executive_id = COALESCE(sqlc.narg(sales_executive_id), sales_executive_id),
    sales_executive_name = COALESCE(sqlc.narg(sales_executive_name), sales_executive_name),
    operations_exec_id = COALESCE(sqlc.narg(operations_exec_id), operations_exec_id),
    operations_exec_name = COALESCE(sqlc.narg(operations_exec_name), operations_exec_name),
    cs_executive_id = COALESCE(sqlc.narg(cs_executive_id), cs_executive_id),
    cs_executive_name = COALESCE(sqlc.narg(cs_executive_name), cs_executive_name),
    agent_deadline = COALESCE(sqlc.narg(agent_deadline), agent_deadline),
    shipment_ready_date = COALESCE(sqlc.narg(shipment_ready_date), shipment_ready_date),
    status = COALESCE(sqlc.narg(status), status),
    priority_level = COALESCE(sqlc.narg(priority_level), priority_level),
    modified_at = now(),
    modified_by = sqlc.arg(actor)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: ArchiveJob :exec
UPDATE ops_job
SET 
    is_active = false,
    modified_at = now(),
    modified_by = sqlc.arg(actor)
WHERE id = sqlc.arg(id);

-- ============================================================
-- JOB PACKAGE QUERIES
-- ============================================================

-- name: ListJobPackages :many
SELECT *
FROM ops_package
WHERE job_id = sqlc.arg(job_id)
  AND is_active
ORDER BY created_at;

-- name: CreateJobPackage :one
INSERT INTO ops_package (
    job_id,
    container_no,
    container_name,
    container_size,
    gross_weight_kg,
    net_weight_kg,
    volume,
    carrier_seal_no,
    commodity_cargo_description,
    package_type,
    cargo_type,
    no_of_packages,
    chargeable_weight,
    hs_code,
    temperature_control,
    created_at,
    created_by,
    is_active
) VALUES (
    sqlc.arg(job_id),
    sqlc.narg(container_no),
    sqlc.narg(container_name),
    sqlc.narg(container_size),
    sqlc.narg(gross_weight_kg),
    sqlc.narg(net_weight_kg),
    sqlc.narg(volume),
    sqlc.narg(carrier_seal_no),
    sqlc.narg(commodity_cargo_description),
    sqlc.narg(package_type),
    sqlc.narg(cargo_type),
    sqlc.narg(no_of_packages),
    sqlc.narg(chargeable_weight),
    sqlc.narg(hs_code),
    sqlc.arg(temperature_control),
    now(),
    sqlc.arg(actor),
    true
) RETURNING *;


-- ============================================================
-- JOB CARRIER QUERIES
-- ============================================================

-- name: GetJobCarriers :many
SELECT *
FROM ops_carrier
WHERE job_id = sqlc.arg(job_id)
  AND is_active
ORDER BY created_at;

-- name: CreateJobCarrier :one
INSERT INTO ops_carrier (
    job_id,
    carrier_party_id,
    carrier_name,
    carrier_contact,
    vessel_name,
    voyage_number,
    flight_id,
    flight_date,
    airport_report_date,
    vehicle_number,
    vehicle_type,
    route_details,
    driver_name,
    driver_contact,
    origin_port_station,
    destination_port_station,
    origin_country,
    destination_country,
    accounting_info,
    handling_info,
    transport_document_reference,
    supporting_doc_url,
    file_region,
    description,
    created_at,
    created_by,
    is_active
) VALUES (
    sqlc.arg(job_id),
    sqlc.narg(carrier_party_id),
    sqlc.narg(carrier_name),
    sqlc.narg(carrier_contact),
    sqlc.narg(vessel_name),
    sqlc.narg(voyage_number),
    sqlc.narg(flight_id),
    sqlc.narg(flight_date),
    sqlc.narg(airport_report_date),
    sqlc.narg(vehicle_number),
    sqlc.narg(vehicle_type),
    sqlc.narg(route_details),
    sqlc.narg(driver_name),
    sqlc.narg(driver_contact),
    sqlc.narg(origin_port_station),
    sqlc.narg(destination_port_station),
    sqlc.narg(origin_country),
    sqlc.narg(destination_country),
    sqlc.narg(accounting_info),
    sqlc.narg(handling_info),
    sqlc.narg(transport_document_reference),
    sqlc.narg(doc_urls),
    sqlc.narg(file_region),
    sqlc.narg(description),
    now(),
    sqlc.arg(actor),
    true
) RETURNING *;

-- name: UpdateJobCarrier :one
UPDATE ops_carrier
SET
    carrier_party_id = COALESCE(sqlc.narg(carrier_party_id), carrier_party_id),
    carrier_name = COALESCE(sqlc.narg(carrier_name), carrier_name),
    carrier_contact = COALESCE(sqlc.narg(carrier_contact), carrier_contact),
    vessel_name = COALESCE(sqlc.narg(vessel_name), vessel_name),
    voyage_number = COALESCE(sqlc.narg(voyage_number), voyage_number),
    flight_id = COALESCE(sqlc.narg(flight_id), flight_id),
    flight_date = COALESCE(sqlc.narg(flight_date), flight_date),
    airport_report_date = COALESCE(sqlc.narg(airport_report_date), airport_report_date),
    vehicle_number = COALESCE(sqlc.narg(vehicle_number), vehicle_number),
    vehicle_type = COALESCE(sqlc.narg(vehicle_type), vehicle_type),
    route_details = COALESCE(sqlc.narg(route_details), route_details),
    driver_name = COALESCE(sqlc.narg(driver_name), driver_name),
    driver_contact = COALESCE(sqlc.narg(driver_contact), driver_contact),
    origin_port_station = COALESCE(sqlc.narg(origin_port_station), origin_port_station),
    destination_port_station = COALESCE(sqlc.narg(destination_port_station), destination_port_station),
    origin_country = COALESCE(sqlc.narg(origin_country), origin_country),
    destination_country = COALESCE(sqlc.narg(destination_country), destination_country),
    accounting_info = COALESCE(sqlc.narg(accounting_info), accounting_info),
    handling_info = COALESCE(sqlc.narg(handling_info), handling_info),
    transport_document_reference = COALESCE(sqlc.narg(transport_document_reference), transport_document_reference),
    supporting_doc_url = COALESCE(sqlc.narg(doc_urls), supporting_doc_url),
    file_region = COALESCE(sqlc.narg(file_region), file_region),
    description = COALESCE(sqlc.narg(description), description),
    modified_at = now(),
    modified_by = sqlc.arg(actor)
WHERE job_id = sqlc.arg(job_id)
  AND id = sqlc.arg(carrier_id)
  AND is_active = true
RETURNING *;

-- ============================================================
-- JOB DOCUMENT QUERIES
-- ============================================================

-- name: ListJobDocuments :many
SELECT *
FROM ops_job_document
WHERE job_id = sqlc.arg(job_id)
  AND is_active
ORDER BY created_at;

-- name: CreateJobDocument :one
INSERT INTO ops_job_document (
    job_id,
    doc_type_code,
    doc_number,
    issued_at,
    issued_date,
    description,
    file_key,
    file_region,
    supporting_doc_urls,
    bl_awb_uploads,
    house_doc_number,
    house_issued_at,
    house_issued_date,
    house_description,
    partial_bl_number,
    switch_bl_awb_number,
    switch_bl_awb_issued_at,
    switch_bl_awb_issued_date,
    switch_bl_awb_description,
    created_at,
    created_by,
    is_active
) VALUES (
    sqlc.arg(job_id),
    sqlc.narg(doc_type_code),
    sqlc.narg(doc_number),
    sqlc.narg(issued_at),
    sqlc.narg(issued_date),
    sqlc.narg(description),
    sqlc.narg(file_key),
    sqlc.narg(file_region),
    sqlc.narg(doc_urls),
    sqlc.narg(bl_awb_uploads),
    sqlc.narg(house_doc_number),
    sqlc.narg(house_issued_at),
    sqlc.narg(house_issued_date),
    sqlc.narg(house_description),
    sqlc.narg(partial_bl_number),
    sqlc.narg(switch_bl_awb_number),
    sqlc.narg(switch_bl_awb_issued_at),
    sqlc.narg(switch_bl_awb_issued_date),
    sqlc.narg(switch_bl_awb_description),
    now(),
    sqlc.arg(actor),
    true
) RETURNING *;

-- ============================================================
-- JOB PARTY QUERIES
-- ============================================================

-- name: GetJobParty :one
SELECT *
FROM ops_party
WHERE job_id = sqlc.arg(job_id)
  AND is_active;

-- name: UpsertJobParty :one
INSERT INTO ops_party (
    job_id,
    shipper_id,
    consignee_id,
    notify_party_id,
    switch_bl_shipper_id,
    switch_bl_consignee_id,
    switch_bl_notify_party_id,
    origin_agent_id,
    destination_agent_id,
    created_by,
    modified_by
)
VALUES (
    sqlc.arg(job_id),
    sqlc.narg(shipper_id),
    sqlc.narg(consignee_id),
    sqlc.narg(notify_party_id),
    sqlc.narg(switch_bl_shipper_id),
    sqlc.narg(switch_bl_consignee_id),
    sqlc.narg(switch_bl_notify_party_id),
    sqlc.narg(origin_agent_id),
    sqlc.narg(destination_agent_id),
    sqlc.arg(actor),
    sqlc.arg(actor)
)
ON CONFLICT (job_id) DO UPDATE
SET
    shipper_id = EXCLUDED.shipper_id,
    consignee_id = EXCLUDED.consignee_id,
    notify_party_id = EXCLUDED.notify_party_id,
    switch_bl_shipper_id = EXCLUDED.switch_bl_shipper_id,
    switch_bl_consignee_id = EXCLUDED.switch_bl_consignee_id,
    switch_bl_notify_party_id = EXCLUDED.switch_bl_notify_party_id,
    origin_agent_id = EXCLUDED.origin_agent_id,
    destination_agent_id = EXCLUDED.destination_agent_id,
    modified_at = now(),
    modified_by = EXCLUDED.modified_by,
    is_active = true
RETURNING *;

-- ============================================================
-- JOB BILLING QUERIES
-- ============================================================

-- name: ListJobBilling :many
SELECT
    b.*,
    p.name AS billing_party_name
FROM ops_billing b
LEFT JOIN party_master p ON p.id = b.billing_party_id
WHERE b.job_id = sqlc.arg(job_id)
  AND b.is_active
ORDER BY b.created_at;

-- name: CreateJobBilling :one
INSERT INTO ops_billing (
    job_id,
    activity_type,
    activity_code,
    billing_party_id,
    po_number,
    po_date,
    currency_code,
    amount,
    description,
    notes,
    supporting_doc_url,
    file_region,
    amount_primary_currency,
    created_at,
    created_by,
    is_active
) VALUES (
    sqlc.arg(job_id),
    sqlc.narg(activity_type),
    sqlc.narg(activity_code),
    sqlc.narg(billing_party_id),
    sqlc.narg(po_number),
    sqlc.narg(po_date),
    sqlc.narg(currency_code),
    sqlc.narg(amount),
    sqlc.narg(description),
    sqlc.narg(notes),
    sqlc.narg(doc_urls),
    sqlc.narg(file_region),
    sqlc.narg(amount_primary_currency),
    now(),
    sqlc.arg(actor),
    true
) RETURNING *;

-- ============================================================
-- JOB PROVISION QUERIES
-- ============================================================

-- name: ListJobProvisions :many
SELECT
    p.*,
    pm.name AS cost_party_name
FROM ops_provision p
LEFT JOIN party_master pm ON pm.id = p.cost_party_id
WHERE p.job_id = sqlc.arg(job_id)
  AND p.is_active
ORDER BY p.created_at;

-- name: CreateJobProvision :one
INSERT INTO ops_provision (
    job_id,
    activity_type,
    activity_code,
    cost_party_id,
    invoice_number,
    invoice_date,
    currency_code,
    amount,
    payment_priority,
    notes,
    supporting_doc_url,
    file_region,
    amount_primary_currency,
    profit,
    created_at,
    created_by,
    is_active
) VALUES (
    sqlc.arg(job_id),
    sqlc.narg(activity_type),
    sqlc.narg(activity_code),
    sqlc.narg(cost_party_id),
    sqlc.narg(invoice_number),
    sqlc.narg(invoice_date),
    sqlc.narg(currency_code),
    sqlc.narg(amount),
    sqlc.narg(payment_priority),
    sqlc.narg(notes),
    sqlc.narg(doc_urls),
    sqlc.narg(file_region),
    sqlc.narg(amount_primary_currency),
    sqlc.narg(profit),
    now(),
    sqlc.arg(actor),
    true
) RETURNING *;

-- ============================================================
-- JOB TRACKING QUERIES
-- ============================================================

-- name: GetJobTracking :one
SELECT *
FROM ops_tracking
WHERE job_id = sqlc.arg(job_id)
  AND is_active
LIMIT 1;

-- name: UpsertJobTracking :one
INSERT INTO ops_tracking (
    job_id,
    etd_date,
    eta_date,
    atd_date,
    ata_date,
    job_status,
    pod_doc_urls,
    file_region,
    document_status,
    notes,
    created_at,
    created_by,
    is_active
) VALUES (
    sqlc.arg(job_id),
    sqlc.narg(etd_date),
    sqlc.narg(eta_date),
    sqlc.narg(atd_date),
    sqlc.narg(ata_date),
    sqlc.narg(job_status),
    sqlc.narg(doc_urls),
    sqlc.narg(file_region),
    sqlc.narg(document_status),
    sqlc.narg(notes),
    now(),
    sqlc.arg(actor),
    true
) 
ON CONFLICT (job_id) 
DO UPDATE SET
    etd_date = COALESCE(EXCLUDED.etd_date, ops_tracking.etd_date),
    eta_date = COALESCE(EXCLUDED.eta_date, ops_tracking.eta_date),
    atd_date = COALESCE(EXCLUDED.atd_date, ops_tracking.atd_date),
    ata_date = COALESCE(EXCLUDED.ata_date, ops_tracking.ata_date),
    job_status = COALESCE(EXCLUDED.job_status, ops_tracking.job_status),
    pod_doc_urls = COALESCE(EXCLUDED.pod_doc_urls, ops_tracking.pod_doc_urls),
    file_region = COALESCE(EXCLUDED.file_region, ops_tracking.file_region),
    document_status = COALESCE(EXCLUDED.document_status, ops_tracking.document_status),
    notes = COALESCE(EXCLUDED.notes, ops_tracking.notes),
    modified_at = now(),
    modified_by = EXCLUDED.created_by
RETURNING *;

