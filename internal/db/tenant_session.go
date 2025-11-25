package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TenantSessionManager manages tenant-specific database sessions
// with shared tenant registry
type TenantSessionManager struct {
	tenantPool  *pgxpool.Pool // Connection to shared frego_tenant_db
	servicePool *pgxpool.Pool // Connection to service-specific DB (frego_finance_db)
	serviceName string        // "finance", "operations", etc.
	mu          sync.RWMutex
}

// NewTenantSessionManager creates a new tenant session manager with shared registry
func NewTenantSessionManager(tenantPool, servicePool *pgxpool.Pool, serviceName string) *TenantSessionManager {
	return &TenantSessionManager{
		tenantPool:  tenantPool,
		servicePool: servicePool,
		serviceName: serviceName,
	}
}

// GetSession returns a database connection with the tenant schema set
func (m *TenantSessionManager) GetSession(ctx context.Context, tenantID uuid.UUID) (*pgxpool.Conn, error) {
	// 1. Query shared tenant registry for schema name
	schemaName, err := m.getTenantSchema(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get tenant schema: %w", err)
	}

	// 2. Get connection to service-specific DB
	conn, err := m.servicePool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire connection: %w", err)
	}

	// 3. Set search path to tenant schema
	_, err = conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s, public", pgx.Identifier{schemaName}.Sanitize()))
	if err != nil {
		conn.Release()
		return nil, fmt.Errorf("set search path: %w", err)
	}

	return conn, nil
}

// getTenantSchema fetches the schema name from shared tenant registry
func (m *TenantSessionManager) getTenantSchema(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var schemaName string
	var schemaColumn string

	// Determine which schema column to query based on service name
	switch m.serviceName {
	case "finance":
		schemaColumn = "finance_schema"
	case "operations":
		schemaColumn = "operations_schema"
	case "inventory":
		schemaColumn = "inventory_schema"
	case "hrms":
		schemaColumn = "hrms_schema"
	default:
		return "", fmt.Errorf("unknown service name: %s", m.serviceName)
	}

	// Query shared tenant registry
	query := fmt.Sprintf(`
		SELECT %s 
		FROM tenant_registry 
		WHERE tenant_id = $1 
		  AND is_active = true
		  AND $2 = ANY(modules_subscribed)
	`, schemaColumn)

	err := m.tenantPool.QueryRow(ctx, query, tenantID, m.serviceName).Scan(&schemaName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("tenant %s not found or not subscribed to %s module", tenantID, m.serviceName)
		}
		return "", fmt.Errorf("query tenant registry: %w", err)
	}

	if schemaName == "" {
		return "", fmt.Errorf("tenant %s has no schema for %s module", tenantID, m.serviceName)
	}

	return schemaName, nil
}

// VerifyTenantAccess verifies tenant has access to this service
func (m *TenantSessionManager) VerifyTenantAccess(ctx context.Context, tenantID uuid.UUID) error {
	var hasAccess bool
	err := m.tenantPool.QueryRow(ctx, `
		SELECT tenant_has_module($1, $2)
	`, tenantID, m.serviceName).Scan(&hasAccess)

	if err != nil {
		return fmt.Errorf("verify tenant access: %w", err)
	}

	if !hasAccess {
		return fmt.Errorf("tenant %s does not have access to %s module", tenantID, m.serviceName)
	}

	return nil
}

// ReleaseSession releases a tenant session back to the pool
func (m *TenantSessionManager) ReleaseSession(conn *pgxpool.Conn) {
	if conn != nil {
		conn.Release()
	}
}

// GetTenantInfo returns tenant information from shared registry
func (m *TenantSessionManager) GetTenantInfo(ctx context.Context, tenantID uuid.UUID) (*TenantInfo, error) {
	var info TenantInfo
	err := m.tenantPool.QueryRow(ctx, `
		SELECT 
			tenant_id,
			tenant_slug,
			tenant_name,
			contact_email,
			modules_subscribed,
			is_active
		FROM tenant_registry
		WHERE tenant_id = $1
	`, tenantID).Scan(
		&info.TenantID,
		&info.TenantSlug,
		&info.TenantName,
		&info.ContactEmail,
		&info.ModulesSubscribed,
		&info.IsActive,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tenant not found: %s", tenantID)
		}
		return nil, fmt.Errorf("get tenant info: %w", err)
	}

	return &info, nil
}

// TenantInfo represents tenant information from registry
type TenantInfo struct {
	TenantID          uuid.UUID
	TenantSlug        string
	TenantName        string
	ContactEmail      *string
	ModulesSubscribed []string
	IsActive          bool
}
