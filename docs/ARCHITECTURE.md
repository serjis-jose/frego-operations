# Finance Module Architecture

## Overview

The Finance module is a standalone microservice that handles all financial operations for Frego ERP, including:
- **Accounts Receivable (AR)**: Customer invoices, receipts, credit notes
- **Accounts Payable (AP)**: Vendor invoices, payments, debit notes
- **General Ledger (GL)**: Chart of accounts, journal entries, financial reporting

## Multi-Tenancy Strategy

### Schema-per-Tenant
Each tenant gets a dedicated PostgreSQL schema (e.g., `finance_tenant_abc`). This provides:
- **Data Isolation**: Complete separation of tenant data
- **Performance**: No query-level filtering overhead
- **Flexibility**: Per-tenant schema migrations if needed
- **Security**: Database-level isolation

### Tenant Provisioning Flow
1. Tenant subscribes to Finance module
2. System calls `ensure_finance_tenant_schema(tenant_id, schema_name)`
3. Procedure creates schema and all finance tables
4. Tenant is ready to use finance features

## Database Design

### External References
The finance module references external entities (Party, Job, Employee) by UUID only:
- No hard Foreign Key constraints to external databases
- Enables true microservice independence
- Data consistency handled via events or API validation

### Replicated Lookups
Essential lookups are replicated locally:
- `currency_lu`
- `payment_term_lu`
- `branch_lu`

This ensures data integrity within the finance module without external dependencies.

## Integration Patterns

### Event-Driven
```
Operations Service → JobCompleted Event → Finance Service → Create Invoice
```

### API-Driven
```
Operations Service → POST /finance/api/v1/invoices → Finance Service
```

### Hybrid
- Critical operations use synchronous API calls
- Non-critical updates use asynchronous events

## Security

### Authentication
- JWT tokens from Keycloak
- Bearer token in Authorization header

### Authorization
- Tenant ID extracted from JWT claims
- Validated against tenant_registry
- Search path set to tenant schema

### Data Access
```
Request → Auth Middleware → Tenant Middleware → Handler
                ↓                    ↓
           Verify JWT        Set tenant context
                              Validate tenant exists
```

## Deployment

### Containerization
- Docker image built from Dockerfile
- Multi-stage build for smaller image size
- Distroless base for security

### Orchestration
- Kubernetes deployment with HPA
- Separate database instance
- Redis for caching (optional)

### Scaling
- Horizontal scaling via replicas
- Database connection pooling
- Read replicas for reporting queries

## Development Workflow

1. **Schema Changes**: Update `db/schema.sql` and `db/provision_tenant.sql`
2. **Code Generation**: Run `make generate` to generate API and DB code
3. **Testing**: Run `go test ./...`
4. **Building**: Run `make build` or `make docker-build`
5. **Deployment**: Push to container registry and deploy

## Monitoring

### Metrics
- Request latency
- Error rates
- Database connection pool stats
- Tenant-specific metrics

### Logging
- Structured logging with slog
- Request ID tracing
- Tenant ID in all logs

### Health Checks
- `/health` endpoint
- Database connectivity check
- Dependency health checks
