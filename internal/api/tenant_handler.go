package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"frego-operations/internal/service/tenant"
)

// TenantHandler handles tenant provisioning API requests
type TenantHandler struct {
	logger         *slog.Logger
	tenantService  *tenant.Service
	internalSecret string
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(logger *slog.Logger, tenantService *tenant.Service, internalSecret string) *TenantHandler {
	return &TenantHandler{
		logger:         logger,
		tenantService:  tenantService,
		internalSecret: internalSecret,
	}
}

// RegisterRoutes registers tenant-related routes
func (h *TenantHandler) RegisterRoutes(r chi.Router) {
	r.Post("/tenants/provision", h.ProvisionTenant)
}

// ProvisionTenantRequest defines the request body for tenant provisioning
type ProvisionTenantRequest struct {
	TenantID    uuid.UUID `json:"tenantId"`
	DisplayName *string   `json:"displayName,omitempty"`
	Actor       *string   `json:"actor,omitempty"`
}

// ProvisionTenantResponse defines the response for tenant provisioning
type ProvisionTenantResponse struct {
	Message    string `json:"message"`
	TenantID   string `json:"tenantId"`
	SchemaName string `json:"schemaName,omitempty"`
}

// ProvisionTenant handles the request to provision a finance schema for a tenant
func (h *TenantHandler) ProvisionTenant(w http.ResponseWriter, r *http.Request) {
	// Provisioning is now handled centrally by frego-backend.
	// This endpoint is deprecated and should not be used.
	resp := ProvisionTenantResponse{
		Message: "provisioning is now handled centrally",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
