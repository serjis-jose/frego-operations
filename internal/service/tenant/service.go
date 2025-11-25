package tenant

import (
	"context"

	"frego-operations/internal/repository/tenant"

	"github.com/google/uuid"
)

// Service handles tenant business logic
type Service struct {
	repo *tenant.Repository
}

// New creates a new tenant service
func New(repo *tenant.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// GetTenantSchema gets the schema name for a tenant
func (s *Service) GetTenantSchema(ctx context.Context, tenantID uuid.UUID) (string, error) {
	return s.repo.GetTenantSchema(ctx, tenantID)
}
