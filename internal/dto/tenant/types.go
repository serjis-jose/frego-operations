package tenant

import (
	"time"

	"github.com/google/uuid"
)

// Tenant represents a tenant in the system
type Tenant struct {
	TenantID     uuid.UUID  `json:"tenant_id"`
	SchemaName   string     `json:"schema_name"`
	DisplayName  *string    `json:"display_name,omitempty"`
	ContactEmail *string    `json:"contact_email,omitempty"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	CreatedBy    *string    `json:"created_by,omitempty"`
	ModifiedAt   *time.Time `json:"modified_at,omitempty"`
	ModifiedBy   *string    `json:"modified_by,omitempty"`
}

// CreateTenantRequest represents a request to provision a new tenant
type CreateTenantRequest struct {
	TenantID     string  `json:"tenant_id" validate:"required,uuid"`
	SchemaName   *string `json:"schema_name,omitempty"`
	DisplayName  *string `json:"display_name,omitempty"`
	ContactEmail *string `json:"contact_email,omitempty" validate:"omitempty,email"`
}

// CreateTenantResponse represents the response after creating a tenant
type CreateTenantResponse struct {
	TenantID   uuid.UUID `json:"tenant_id"`
	SchemaName string    `json:"schema_name"`
	Message    string    `json:"message"`
}
