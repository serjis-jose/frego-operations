# Deployment Guide

## Prerequisites

- Docker and Docker Compose
- PostgreSQL 14+ (for production)
- Kubernetes cluster (for production)
- Keycloak instance for authentication

## Local Development

### Using Docker Compose

1. Start the services:
```bash
docker-compose up -d
```

2. Check logs:
```bash
docker-compose logs -f finance-service
```

3. Provision a tenant:
```bash
docker-compose exec postgres psql -U postgres -d frego_finance_db -c \
  "CALL ensure_finance_tenant_schema('550e8400-e29b-41d4-a716-446655440000', 'finance_demo');"
```

4. Test the API:
```bash
curl -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000" \
     http://localhost:8080/health
```

### Manual Setup

1. Start PostgreSQL:
```bash
docker run -d \
  --name frego-operations-db \
  -e POSTGRES_DB=frego_finance_db \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5433:5432 \
  postgres:16-alpine
```

2. Run provisioning script:
```bash
psql postgres://postgres:postgres@localhost:5433/frego_finance_db \
  -f db/provision_tenant.sql
```

3. Set environment variables:
```bash
export DB_URL=postgres://postgres:postgres@localhost:5433/frego_finance_db
export HTTP_ADDRESS=:8080
export ENVIRONMENT=development
```

4. Run the service:
```bash
go run ./cmd/server
```

## Production Deployment

### Kubernetes

1. Create namespace:
```bash
kubectl create namespace frego-operations
```

2. Create secrets:
```bash
kubectl create secret generic finance-db-credentials \
  --from-literal=url='postgres://user:pass@host:5432/frego_finance_db' \
  -n frego-operations
```

3. Deploy:
```bash
kubectl apply -f k8s/ -n frego-operations
```

### Environment Variables

Required:
- `DB_URL`: PostgreSQL connection string
- `KEYCLOAK_ISSUER`: Keycloak issuer URL
- `KEYCLOAK_AUDIENCE`: Expected audience in JWT

Optional:
- `HTTP_ADDRESS`: Server address (default: `:8080`)
- `ENVIRONMENT`: Environment name (default: `production`)
- `DEFAULT_TENANT`: Default tenant UUID
- `S3_BUCKET`: S3 bucket for documents
- `S3_REGION`: S3 region

### Database Migration

For production databases:

1. Backup the database
2. Run the provisioning script:
```bash
psql $DB_URL -f db/provision_tenant.sql
```

3. Provision tenants as needed:
```bash
./scripts/provision_tenant.sh <tenant-uuid> <schema-name>
```

### Health Checks

Configure health checks:
- Liveness: `GET /health`
- Readiness: `GET /health` (checks DB connectivity)

### Monitoring

Set up monitoring:
- Prometheus metrics endpoint (if enabled)
- Log aggregation (ELK, Datadog, etc.)
- APM tracing (Jaeger, Datadog, etc.)

## Scaling

### Horizontal Scaling

The service is stateless and can be scaled horizontally:

```bash
kubectl scale deployment finance-service --replicas=5 -n frego-operations
```

### Database Scaling

- Use connection pooling (configured via `DB_MAX_OPEN_CONNS`)
- Consider read replicas for reporting queries
- Monitor connection pool metrics

### Caching

Consider adding Redis for:
- Session caching
- Frequently accessed lookups
- Rate limiting

## Backup and Recovery

### Database Backup

```bash
pg_dump -Fc $DB_URL > frego_finance_backup.dump
```

### Restore

```bash
pg_restore -d $DB_URL frego_finance_backup.dump
```

## Troubleshooting

### Service won't start

Check:
1. Database connectivity: `psql $DB_URL`
2. Environment variables are set
3. Keycloak is accessible

### Tenant provisioning fails

Check:
1. Tenant UUID is valid
2. Schema name doesn't already exist
3. Database user has CREATE SCHEMA permission

### API returns 403

Check:
1. Tenant ID in header matches registered tenant
2. Tenant is active in tenant_registry
3. JWT token is valid and contains tenant claim
