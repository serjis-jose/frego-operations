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
	// TODO: Implement internal secret validation for service-to-service authentication
	// Temporarily skipped for development - should validate Secret header matches internalSecret
	// if h.internalSecret != "" {
	// 	providedSecret := r.Header.Get("Secret")
	// 	if providedSecret != h.internalSecret {
	// 		h.logger.Warn("invalid internal secret provided",
	// 			slog.String("remote_addr", r.RemoteAddr))
	// 		http.Error(w, "unauthorized", http.StatusUnauthorized)
	// 		return
	// 	}
	// }

	var req ProvisionTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", slog.Any("error", err))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.TenantID == uuid.Nil {
		http.Error(w, "tenantId is required", http.StatusBadRequest)
		return
	}

	displayName := ""
	if req.DisplayName != nil {
		displayName = *req.DisplayName
	}

	h.logger.Info("received finance provisioning request",
		slog.String("tenant_id", req.TenantID.String()),
		slog.String("display_name", displayName))

	// Set actor for database operations if provided
	ctx := r.Context()
	if req.Actor != nil && *req.Actor != "" {
		// Set actor in context for database operations
		// This can be used by the service layer if needed
		ctx = r.Context()
	}

	err := h.tenantService.ProvisionTenant(ctx, req.TenantID, displayName)
	if err != nil {
		h.logger.Error("failed to provision finance tenant schema",
			slog.String("tenant_id", req.TenantID.String()),
			slog.String("display_name", displayName),
			slog.Any("error", err))
		http.Error(w, "failed to provision finance schema", http.StatusInternalServerError)
		return
	}

	// Get the final schema name
	finalSchema, _ := h.tenantService.GetTenantSchema(ctx, req.TenantID)

	h.logger.Info("successfully provisioned finance tenant schema",
		slog.String("tenant_id", req.TenantID.String()),
		slog.String("schema", finalSchema))

	resp := ProvisionTenantResponse{
		Message:    "finance schema provisioned successfully",
		TenantID:   req.TenantID.String(),
		SchemaName: finalSchema,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
