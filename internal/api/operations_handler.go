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
	// Provisioning is now handled centrally by frego-backend.
	// This endpoint is deprecated and should not be used.
	msg := "provisioning is now handled centrally"
	return ProvisionTenant200JSONResponse{
		Message: &msg,
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
