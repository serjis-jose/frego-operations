package tenant

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles tenant data access
type Repository struct {
	tenantPool     *pgxpool.Pool
	operationsPool *pgxpool.Pool
	dbUser         string
}

// New creates a new tenant repository
func New(tenantPool, operationsPool *pgxpool.Pool, dbUser string) *Repository {
	return &Repository{
		tenantPool:     tenantPool,
		operationsPool: operationsPool,
		dbUser:         dbUser,
	}
}

// GetTenantSchema returns the schema name for a tenant
func (r *Repository) GetTenantSchema(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var schemaName string
	err := r.tenantPool.QueryRow(ctx, `
		SELECT operations_schema 
		FROM registry.tenant_registry 
		WHERE tenant_id = $1 AND is_active = true
	`, tenantID).Scan(&schemaName)

	if err != nil {
		return "", fmt.Errorf("get tenant schema: %w", err)
	}

	return schemaName, nil
}
