package api

import (
	"log/slog"

	operationsservice "frego-operations/internal/service/operations"
	tenantservice "frego-operations/internal/service/tenant"
)

// OperationsHandler handles operations API requests
type OperationsHandler struct {
	logger            *slog.Logger
	operationsService *operationsservice.Service
	tenantService     *tenantservice.Service
	internalSecret    string
}

// NewOperationsHandler creates a new operations handler
func NewOperationsHandler(
	logger *slog.Logger,
	operationsService *operationsservice.Service,
	tenantService *tenantservice.Service,
	internalSecret string,
) *OperationsHandler {
	return &OperationsHandler{
		logger:            logger,
		operationsService: operationsService,
		tenantService:     tenantService,
		internalSecret:    internalSecret,
	}
}

/*
// HealthCheck implements the health check endpoint
func (h *OperationsHandler) HealthCheck(ctx context.Context, request HealthCheckRequestObject) (HealthCheckResponseObject, error) {
	return HealthCheck200TextResponse("OK"), nil
}

// ProvisionTenant provisions operations schema for a tenant
func (h *OperationsHandler) ProvisionTenant(ctx context.Context, request ProvisionTenantRequestObject) (ProvisionTenantResponseObject, error) {
	if h.internalSecret == "" {
		h.logger.Warn("frego internal secret is missing; skipping provisioning auth check")
	}

	if request.Body == nil {
		return ProvisionTenant400JSONResponse{
			Code:    "INVALID_REQUEST",
			Message: "request body is required",
		}, nil
	}

	tenantID := request.Body.TenantId
	if tenantID == uuid.Nil {
		h.logger.Error("invalid tenant ID", slog.String("tenant_id", tenantID.String()))
		return ProvisionTenant400JSONResponse{
			Code:    "INVALID_TENANT_ID",
			Message: "tenantId is required",
		}, nil
	}

	displayName := ""
	if request.Body.DisplayName != nil {
		displayName = *request.Body.DisplayName
	}

	// Provision operations schema
	err := h.tenantService.ProvisionTenant(ctx, tenantID, displayName)
	if err != nil {
		h.logger.Error("failed to provision tenant", slog.Any("error", err))
		return ProvisionTenant500JSONResponse{
			Code:    "PROVISION_FAILED",
			Message: err.Error(),
		}, nil
	}

	// Get the final schema name
	schemaName, _ := h.tenantService.GetTenantSchema(ctx, tenantID)

	msg := "operations schema provisioned successfully"

	return ProvisionTenant200JSONResponse{
		Message:    &msg,
		TenantId:   &tenantID,
		SchemaName: &schemaName,
	}, nil
}

// ListJobs implements the list jobs endpoint
func (h *OperationsHandler) ListJobs(ctx context.Context, request ListJobsRequestObject) (ListJobsResponseObject, error) {
	jobs, err := h.operationsService.ListJobs(ctx, &request.Params)
	if err != nil {
		return nil, err
	}
	return ListJobs200JSONResponse{Items: jobs}, nil
}

// CreateJob implements the create job endpoint
func (h *OperationsHandler) CreateJob(ctx context.Context, request CreateJobRequestObject) (CreateJobResponseObject, error) {
	if request.Body == nil {
		return CreateJob400JSONResponse{Message: "request body required"}, nil
	}

	principal, ok := common.PrincipalFromContext(ctx)
	if !ok {
		return CreateJob400JSONResponse{Message: "unauthorized"}, nil
	}

	job, err := h.operationsService.CreateJob(ctx, *request.Body, principal.Username)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return CreateJob400JSONResponse{Message: "job code already exists"}, nil
			case "23503":
				return CreateJob400JSONResponse{Message: "invalid foreign key reference"}, nil
			}
		}
		return nil, err
	}

	return CreateJob201JSONResponse(job), nil
}

// GetJob implements the get job endpoint
func (h *OperationsHandler) GetJob(ctx context.Context, request GetJobRequestObject) (GetJobResponseObject, error) {
	job, err := h.operationsService.GetJob(ctx, request.JobId)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return GetJob404JSONResponse{NotFoundJSONResponse: NotFoundJSONResponse{Message: "job not found"}}, nil
		}
		return nil, err
	}

	return GetJob200JSONResponse(job), nil
}

// UpdateJob implements the update job endpoint
func (h *OperationsHandler) UpdateJob(ctx context.Context, request UpdateJobRequestObject) (UpdateJobResponseObject, error) {
	if request.Body == nil {
		return nil, fmt.Errorf("request body required")
	}

	principal, ok := common.PrincipalFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("unauthorized")
	}

	job, err := h.operationsService.UpdateJob(ctx, request.JobId, *request.Body, principal.Username)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return UpdateJob404JSONResponse{NotFoundJSONResponse: NotFoundJSONResponse{Message: "job not found"}}, nil
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, fmt.Errorf("invalid foreign key reference")
		}
		return nil, err
	}

	return UpdateJob200JSONResponse(job), nil
}

// ArchiveJob implements the archive job endpoint
func (h *OperationsHandler) ArchiveJob(ctx context.Context, request ArchiveJobRequestObject) (ArchiveJobResponseObject, error) {
	principal, ok := common.PrincipalFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("unauthorized")
	}

	err := h.operationsService.ArchiveJob(ctx, request.JobId, principal.Username)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return ArchiveJob404JSONResponse{NotFoundJSONResponse: NotFoundJSONResponse{Message: "job not found"}}, nil
		}
		return nil, err
	}

	return ArchiveJob204Response{}, nil
}
*/
