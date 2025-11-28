-- name: UpsertOrder :one
INSERT INTO orders (
    order_id,
    sales_team,
    sales_person,
    mode,
    customer_name,
    classification,
    agent_deadline,
    comments,
    commodity,
    destination_city,
    destination_country,
    incoterm,
    package_list,
    pickup_address,
    shipment_ready_date,
    shipment_type,
    source_city,
    source_country,
    status,
    order_created_on,
    order_created_by,
    is_deleted,
    template_name,
    email_logs,
    negotiated_quotes
)
VALUES (
    sqlc.arg(order_id),
    sqlc.narg(sales_team),
    sqlc.narg(sales_person),
    sqlc.narg(mode),
    sqlc.narg(customer_name),
    sqlc.narg(classification),
    sqlc.narg(agent_deadline),
    sqlc.narg(comments),
    sqlc.narg(commodity),
    sqlc.narg(destination_city),
    sqlc.narg(destination_country),
    sqlc.narg(incoterm),
    sqlc.narg(package_list),
    sqlc.narg(pickup_address),
    sqlc.narg(shipment_ready_date),
    sqlc.narg(shipment_type),
    sqlc.narg(source_city),
    sqlc.narg(source_country),
    sqlc.narg(status),
    sqlc.narg(order_created_on),
    sqlc.narg(order_created_by),
    sqlc.arg(is_deleted),
    sqlc.narg(template_name),
    sqlc.narg(email_logs),
    sqlc.narg(negotiated_quotes)
)
ON CONFLICT (order_id) DO UPDATE SET
    sales_team = EXCLUDED.sales_team,
    sales_person = EXCLUDED.sales_person,
    mode = EXCLUDED.mode,
    customer_name = EXCLUDED.customer_name,
    classification = EXCLUDED.classification,
    agent_deadline = EXCLUDED.agent_deadline,
    comments = EXCLUDED.comments,
    commodity = EXCLUDED.commodity,
    destination_city = EXCLUDED.destination_city,
    destination_country = EXCLUDED.destination_country,
    incoterm = EXCLUDED.incoterm,
    package_list = EXCLUDED.package_list,
    pickup_address = EXCLUDED.pickup_address,
    shipment_ready_date = EXCLUDED.shipment_ready_date,
    shipment_type = EXCLUDED.shipment_type,
    source_city = EXCLUDED.source_city,
    source_country = EXCLUDED.source_country,
    status = EXCLUDED.status,
    order_created_on = EXCLUDED.order_created_on,
    order_created_by = EXCLUDED.order_created_by,
    is_deleted = EXCLUDED.is_deleted,
    template_name = EXCLUDED.template_name,
    email_logs = EXCLUDED.email_logs,
    negotiated_quotes = EXCLUDED.negotiated_quotes
RETURNING *;

-- name: ListOrders :many
SELECT *
FROM orders
ORDER BY order_created_on DESC NULLS LAST, created_at DESC;

-- name: GetOrder :one
SELECT *
FROM orders
WHERE order_id = sqlc.arg(order_id);
