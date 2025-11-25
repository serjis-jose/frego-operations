package common

import "context"

type tenantKey struct{}

type tenantInfo struct {
	ID          string
	Schema      string
	DisplayName string
}

// WithTenant returns a new context carrying the provided tenant identifier and schema.
func WithTenant(ctx context.Context, tenantID, schema, displayName string) context.Context {
	return context.WithValue(ctx, tenantKey{}, tenantInfo{
		ID:          tenantID,
		Schema:      schema,
		DisplayName: displayName,
	})
}

// TenantFromContext extracts the tenant identifier and schema from context.
func TenantFromContext(ctx context.Context) (tenantID string, schema string, ok bool) {
	val := ctx.Value(tenantKey{})
	if val == nil {
		return "", "", false
	}
	info, castOK := val.(tenantInfo)
	if !castOK || info.ID == "" {
		return "", "", false
	}
	return info.ID, info.Schema, true
}

// TenantDisplayNameFromContext returns the tenant display name if present.
func TenantDisplayNameFromContext(ctx context.Context) (string, bool) {
	val := ctx.Value(tenantKey{})
	if val == nil {
		return "", false
	}
	info, castOK := val.(tenantInfo)
	if !castOK || info.ID == "" {
		return "", false
	}
	return info.DisplayName, true
}
