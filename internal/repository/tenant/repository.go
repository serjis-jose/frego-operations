package tenant

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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

var (
	schemaUnsafeChars     = regexp.MustCompile(`[^a-z0-9_]`)
	schemaMultiUnderscore = regexp.MustCompile(`_{2,}`)
)

// sanitizeSchemaSlug sanitizes a display name into a valid schema slug
func sanitizeSchemaSlug(value string) string {
	slug := strings.ToLower(strings.TrimSpace(value))
	slug = schemaUnsafeChars.ReplaceAllString(slug, "_")
	slug = schemaMultiUnderscore.ReplaceAllString(slug, "_")
	return strings.Trim(slug, "_")
}

// ProvisionTenant creates a new tenant schema
func (r *Repository) ProvisionTenant(ctx context.Context, tenantID uuid.UUID, displayName string) error {
	// 1. Compute schema name from displayName (matching backend logic)
	var schemaName string
	if displayName != "" {
		if slug := sanitizeSchemaSlug(displayName); slug != "" {
			schemaName = "ops_" + slug
		}
	}

	// If no schema name from displayName, generate from tenant_id
	if schemaName == "" {
		tenantIDStr := strings.ToLower(tenantID.String())
		sanitized := schemaUnsafeChars.ReplaceAllString(tenantIDStr, "_")
		schemaName = "ops_" + sanitized
	}

	// 2. Insert audit log entry with pending status
	_, logErr := r.tenantPool.Exec(ctx, `
		INSERT INTO tenant_module_log (tenant_id, module_name, action, schema_name, status, provisioned_by)
		VALUES ($1, 'operations', 'provision', $2, 'pending', current_user)
	`, tenantID, schemaName)
	if logErr != nil {
		// Continue even if log fails, but surface the error later
		logErr = fmt.Errorf("audit log insert failed: %w", logErr)
	}

	// 3. Provision schema in operations DB
	// Pass the computed schema name (stored procedure will use it as-is since it already has "ops_" prefix)
	_, err := r.operationsPool.Exec(ctx, `
		CALL ensure_operations_tenant_schema($1, $2, $3)
	`, tenantID, schemaName, r.dbUser)
	if err != nil {
		// Update audit log to failed
		_, _ = r.tenantPool.Exec(ctx, `
			UPDATE tenant_module_log SET status = 'failed', error_message = $3
			WHERE tenant_id = $1 AND module_name = 'operations' AND action = 'provision' AND status = 'pending'
		`, tenantID, schemaName, err.Error())
		if logErr != nil {
			return fmt.Errorf("%w; also %v", err, logErr)
		}
		return fmt.Errorf("provision tenant schema: %w", err)
	}

	// 4. Update tenant registry in tenant DB with the computed schema name
	// Note: We assume operations_schema column exists in tenant_registry
	_, err = r.tenantPool.Exec(ctx, `
		UPDATE tenant_registry 
		SET operations_schema = $2, modified_at = now()
		WHERE tenant_id = $1
	`, tenantID, schemaName)

	if err != nil {
		// Note: Schema is created but registry update failed. This is an inconsistency.
		// In a real system, we might want distributed transaction or compensation.
		return fmt.Errorf("update tenant registry: %w", err)
	}

	return nil
}

// GetTenantSchema returns the schema name for a tenant
func (r *Repository) GetTenantSchema(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var schemaName string
	err := r.tenantPool.QueryRow(ctx, `
		SELECT operations_schema 
		FROM tenant_registry 
		WHERE tenant_id = $1 AND is_active = true
	`, tenantID).Scan(&schemaName)

	if err != nil {
		return "", fmt.Errorf("get tenant schema: %w", err)
	}

	return schemaName, nil
}
