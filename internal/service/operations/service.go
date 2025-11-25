package operations

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"frego-operations/internal/common"
	sqlc "frego-operations/internal/db/sqlc"
	operationsdto "frego-operations/internal/dto/operations"
	"frego-operations/internal/logging"
	repository "frego-operations/internal/repository/operations"
)

// Service orchestrates business logic for operations.
type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// ============================================================
// LOOKUP METHODS
// ============================================================

// GetLookups retrieves all operations-related lookup data.
func (s *Service) GetLookups(ctx context.Context) (operationsdto.OperationsLookups, error) {
	logger := logging.FromContext(ctx)
	logger.Info("fetching operations lookups")

	result := operationsdto.OperationsLookups{
		TransportModes:       []string{},
		MovementTypes:        map[string][]string{},
		ServiceTypes:         map[string][]string{},
		ServiceSubcategories: map[string][]string{},
		Incoterms:            []operationsdto.IncotermLookup{},
		JobStatuses:          []operationsdto.JobStatus{},
		DocumentStatuses:     []operationsdto.DocumentStatus{},
		PriorityLevels:       []operationsdto.PriorityLevel{},
		RoleDetails:          []operationsdto.RoleDetails{},
		SalesExecutives:      []operationsdto.SalesExecutiveLookup{},
		CSExecutives:         []operationsdto.CSExecutiveLookup{},
		Branches:             []operationsdto.BranchLookup{},
	}

	// Fetch transport mode lookups and build hierarchy
	transportRows, err := s.repo.ListTransportModeServiceLookups(ctx)
	if err != nil {
		logger.Error("failed to list transport mode service lookups", slog.Any("error", err))
		return result, fmt.Errorf("operations: list transport mode services: %w", err)
	}

	transportSeen := make(map[string]struct{})
	movementSeen := make(map[string]map[string]struct{})    // mode -> movement set
	serviceSeen := make(map[string]map[string]struct{})     // movement -> service set
	subcategorySeen := make(map[string]map[string]struct{}) // service -> subcategory set

	for _, row := range transportRows {
		mode := common.PgtypeTextToString(row.TransportModeName)
		if _, ok := transportSeen[mode]; !ok {
			result.TransportModes = append(result.TransportModes, mode)
			transportSeen[mode] = struct{}{}
			movementSeen[mode] = make(map[string]struct{})
		}
		movement := common.PgtypeTextToString(row.MovementTypeName)
		if _, ok := movementSeen[mode][movement]; !ok {
			result.MovementTypes[mode] = append(result.MovementTypes[mode], movement)
			movementSeen[mode][movement] = struct{}{}
			if _, ok := serviceSeen[movement]; !ok {
				serviceSeen[movement] = make(map[string]struct{})
			}
		}
		service := common.PgtypeTextToString(row.ServiceTypeName)
		if _, ok := serviceSeen[movement][service]; !ok {
			result.ServiceTypes[movement] = append(result.ServiceTypes[movement], service)
			serviceSeen[movement][service] = struct{}{}
			if _, ok := subcategorySeen[service]; !ok {
				subcategorySeen[service] = make(map[string]struct{})
			}
		}
		subcategory := common.PgtypeTextToString(row.ServiceSubcategoryName)
		if _, ok := subcategorySeen[service][subcategory]; !ok && subcategory != "" {
			result.ServiceSubcategories[service] = append(result.ServiceSubcategories[service], subcategory)
			subcategorySeen[service][subcategory] = struct{}{}
		}
	}

	// Fetch job statuses
	jobStatusRows, err := s.repo.ListJobStatusLookups(ctx)
	if err != nil {
		logger.Error("failed to list job status lookups", slog.Any("error", err))
		return result, fmt.Errorf("operations: list job statuses: %w", err)
	}
	for _, row := range jobStatusRows {
		desc := common.PgtypeTextToStringPtr(row.JobStatusDesc)
		result.JobStatuses = append(result.JobStatuses, operationsdto.JobStatus{
			JobStatusID:   row.JobStatusID.Int16,
			JobStatusName: common.PgtypeTextToString(row.JobStatusName),
			JobStatusDesc: desc,
		})
	}

	// Fetch document statuses
	docStatusRows, err := s.repo.ListDocumentStatusLookups(ctx)
	if err != nil {
		logger.Error("failed to list document status lookups", slog.Any("error", err))
		return result, fmt.Errorf("operations: list document statuses: %w", err)
	}
	for _, row := range docStatusRows {
		desc := common.PgtypeTextToStringPtr(row.DocStatusDesc)
		result.DocumentStatuses = append(result.DocumentStatuses, operationsdto.DocumentStatus{
			DocStatusID:   row.DocStatusID.Int16,
			DocStatusName: common.PgtypeTextToString(row.DocStatusName),
			DocStatusDesc: desc,
		})
	}

	// Fetch priority levels
	priorityRows, err := s.repo.ListPriorityLookups(ctx)
	if err != nil {
		logger.Error("failed to list priority lookups", slog.Any("error", err))
		return result, fmt.Errorf("operations: list priorities: %w", err)
	}
	for _, row := range priorityRows {
		result.PriorityLevels = append(result.PriorityLevels, operationsdto.PriorityLevel{
			PriorityID:    row.PriorityID,
			PriorityLabel: row.PriorityLabel,
		})
	}

	// Fetch role details
	roleRows, err := s.repo.ListRoleDetailsLookups(ctx)
	if err != nil {
		logger.Error("failed to list role details lookups", slog.Any("error", err))
		return result, fmt.Errorf("operations: list role details: %w", err)
	}
	for _, row := range roleRows {
		desc := common.PgtypeTextToStringPtr(row.RoleDesc)
		result.RoleDetails = append(result.RoleDetails, operationsdto.RoleDetails{
			RoleID:   row.RoleID.Int16,
			RoleName: common.PgtypeTextToString(row.RoleName),
			RoleDesc: desc,
		})
	}

	branchRows, err := s.repo.ListBranchLookups(ctx)
	if err != nil {
		logger.Error("failed to list branch lookups", slog.Any("error", err))
		return result, fmt.Errorf("operations: list branches: %w", err)
	}
	for _, row := range branchRows {
		result.Branches = append(result.Branches, operationsdto.BranchLookup{
			BranchID:   row.BranchID,
			BranchName: row.BranchName,
			IsActive:   common.BoolValue(row.IsActive),
		})
	}

	// Fetch sales executive lookups grouped by branch
	salesExecRows, err := s.repo.ListSalesExecutiveLookups(ctx)
	if err != nil {
		logger.Error("failed to list sales executive lookups", slog.Any("error", err))
		return result, fmt.Errorf("operations: list sales executives: %w", err)
	}
	for _, row := range salesExecRows {
		exec := operationsdto.SalesExecutiveLookup{
			SalesExecID:   row.SalesExecID,
			SalesExecName: row.SalesExecName,
			BranchID:      common.UUIDPtr(row.BranchID),
			BranchName:    common.TextPtr(row.BranchName),
		}
		result.SalesExecutives = append(result.SalesExecutives, exec)
	}

	// Fetch CS executive lookups grouped by branch
	csExecRows, err := s.repo.ListCSExecutiveLookups(ctx)
	if err != nil {
		logger.Error("failed to list cs executive lookups", slog.Any("error", err))
		return result, fmt.Errorf("operations: list cs executives: %w", err)
	}
	for _, row := range csExecRows {
		exec := operationsdto.CSExecutiveLookup{
			CSExecID:   row.CsExecID,
			CSExecName: row.CsExecName,
			BranchID:   common.UUIDPtr(row.BranchID),
			BranchName: common.TextPtr(row.BranchName),
		}
		result.CSExecutives = append(result.CSExecutives, exec)
	}

	logger.Info("fetched operations lookups",
		slog.Int("transport_modes", len(result.TransportModes)),
		slog.Int("incoterms", len(result.Incoterms)),
		slog.Int("job_statuses", len(result.JobStatuses)),
		slog.Int("document_statuses", len(result.DocumentStatuses)),
		slog.Int("role_details", len(result.RoleDetails)),
		slog.Int("branches", len(result.Branches)),
		slog.Int("sales_executives", len(result.SalesExecutives)),
		slog.Int("cs_executives", len(result.CSExecutives)),
	)

	return result, nil
}

// ============================================================
// JOB CRUD METHODS
// ============================================================

// generateJobCode generates a unique job code in the format FRG-YYYYMM-NNNN
func (s *Service) generateJobCode(ctx context.Context) (string, error) {
	logger := logging.FromContext(ctx)

	// Get current year and month
	now := time.Now()
	prefix := fmt.Sprintf("FRG-%d%02d", now.Year(), now.Month())

	// Get next sequence number for this month
	nextSeq, err := s.repo.GetNextJobSequence(ctx, prefix)
	if err != nil {
		logger.Error("failed to get next job sequence", slog.Any("error", err))
		return "", fmt.Errorf("operations: generate job code: %w", err)
	}

	// Format as FRG-YYYYMM-NNNN
	jobCode := fmt.Sprintf("%s-%04d", prefix, nextSeq)
	logger.Info("generated job code", slog.String("jobCode", jobCode))

	return jobCode, nil
}

// ListJobs retrieves a list of jobs with optional filters.
func (s *Service) ListJobs(ctx context.Context, status, jobType *string, customerID *uuid.UUID, limit int32) ([]operationsdto.JobListItem, error) {
	logger := logging.FromContext(ctx)
	logger.Info("listing jobs", slog.Any("status", status), slog.Any("jobType", jobType), slog.Any("customerID", customerID))

	rows, err := s.repo.ListJobs(ctx, status, jobType, customerID, limit)
	if err != nil {
		logger.Error("failed to list jobs", slog.Any("error", err))
		return nil, fmt.Errorf("operations: list jobs: %w", err)
	}

	result := make([]operationsdto.JobListItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, operationsdto.JobListItem{
			ID:                 row.ID,
			JobCode:            row.JobCode,
			EnquiryNumber:      common.PgtypeTextToStringPtr(row.EnquiryNumber),
			JobType:            common.PgtypeTextToStringPtr(row.JobType),
			TransportMode:      common.PgtypeTextToStringPtr(row.TransportMode),
			ServiceType:        common.PgtypeTextToStringPtr(row.ServiceType),
			CustomerID:         uuidFromPgtype(row.CustomerID),
			CustomerName:       common.PgtypeTextToStringPtr(row.CustomerName),
			AgentID:            uuidFromPgtype(row.AgentID),
			AgentName:          common.PgtypeTextToStringPtr(row.AgentName),
			ShipmentOrigin:     common.PgtypeTextToStringPtr(row.ShipmentOrigin),
			DestinationCity:    common.PgtypeTextToStringPtr(row.DestinationCity),
			DestinationState:   common.PgtypeTextToStringPtr(row.DestinationState),
			DestinationCountry: common.PgtypeTextToStringPtr(row.DestinationCountry),
			SourceCity:         common.PgtypeTextToStringPtr(row.SourceCity),
			SourceState:        common.PgtypeTextToStringPtr(row.SourceState),
			SourceCountry:      common.PgtypeTextToStringPtr(row.SourceCountry),
			Status:             common.PgtypeTextToStringPtr(row.Status),
			PriorityLevel:      int2ToStringPtr(row.PriorityLevel),
			SalesExecutive: operationsdto.Employee{
				ID:    row.SalesExecutiveID.Bytes,
				Name:  common.PgtypeTextToString(row.SalesExecutiveName),
				Email: common.PgtypeTextToString(row.SalesExecutiveEmail),
				Role:  common.PgtypeTextToString(row.SalesExecutiveRole),
			},
			OperationsExecutive: operationsdto.Employee{
				ID:    row.OperationsExecID.Bytes,
				Name:  common.PgtypeTextToString(row.OperationsExecName),
				Email: common.PgtypeTextToString(row.OperationsExecEmail),
				Role:  common.PgtypeTextToString(row.OperationsExecRole),
			},
			CreatedAt:  row.CreatedAt.Time,
			ModifiedAt: timeFromPgtype(row.ModifiedAt),
			IsActive:   row.IsActive.Bool,
		})
	}

	logger.Info("listed jobs", slog.Int("count", len(result)))
	return result, nil
}

// GetJob retrieves a job by ID with all related data.
func (s *Service) GetJob(ctx context.Context, jobID uuid.UUID) (operationsdto.JobDetail, error) {
	logger := logging.FromContext(ctx)
	logger.Info("fetching job", slog.String("jobID", jobID.String()))

	// Fetch job
	job, err := s.repo.GetJob(ctx, jobID)
	if err != nil {
		logger.Error("failed to get job", slog.Any("error", err))
		return operationsdto.JobDetail{}, fmt.Errorf("operations: get job: %w", err)
	}

	detail := operationsdto.JobDetail{
		ID:                 job.ID,
		JobCode:            job.JobCode,
		EnquiryNumber:      common.PgtypeTextToStringPtr(job.EnquiryNumber),
		JobType:            common.PgtypeTextToStringPtr(job.JobType),
		TransportMode:      common.PgtypeTextToStringPtr(job.TransportMode),
		ServiceType:        common.PgtypeTextToStringPtr(job.ServiceType),
		ServiceSubcategory: common.PgtypeTextToStringPtr(job.ServiceSubcategory),
		ParentJobID:        uuidFromPgtype(job.ParentJobID),
		CustomerID:         uuidFromPgtype(job.CustomerID),
		CustomerName:       common.PgtypeTextToStringPtr(job.CustomerName),
		AgentID:            uuidFromPgtype(job.AgentID),
		AgentName:          common.PgtypeTextToStringPtr(job.AgentName),
		ShipmentOrigin:     common.PgtypeTextToStringPtr(job.ShipmentOrigin),
		DestinationCity:    common.PgtypeTextToStringPtr(job.DestinationCity),
		DestinationState:   common.PgtypeTextToStringPtr(job.DestinationState),
		DestinationCountry: common.PgtypeTextToStringPtr(job.DestinationCountry),
		SourceCity:         common.PgtypeTextToStringPtr(job.SourceCity),
		SourceState:        common.PgtypeTextToStringPtr(job.SourceState),
		SourceCountry:      common.PgtypeTextToStringPtr(job.SourceCountry),
		BranchID:           uuidFromPgtype(job.BranchID),
		BranchName:         common.PgtypeTextToStringPtr(job.BranchName),
		IncotermCode:       common.PgtypeTextToStringPtr(job.IncoTermCode),
		Commodity:          common.PgtypeTextToStringPtr(job.Commodity),
		Classification:     common.PgtypeTextToStringPtr(job.Classification),
		SalesExecutive: operationsdto.Employee{
			ID:    job.SalesExecutiveID.Bytes,
			Name:  common.PgtypeTextToString(job.SalesExecutiveName),
			Email: common.PgtypeTextToString(job.SalesExecutiveEmail),
			Role:  common.PgtypeTextToString(job.SalesExecutiveRole),
		},
		OperationsExecutive: operationsdto.Employee{
			ID:    job.OperationsExecID.Bytes,
			Name:  common.PgtypeTextToString(job.OperationsExecName),
			Email: common.PgtypeTextToString(job.OperationsExecEmail),
			Role:  common.PgtypeTextToString(job.OperationsExecRole),
		},
		CSExecutive: operationsdto.Employee{
			ID:    job.CsExecutiveID.Bytes,
			Name:  common.PgtypeTextToString(job.CsExecutiveName),
			Email: common.PgtypeTextToString(job.CsExecutiveEmail),
			Role:  common.PgtypeTextToString(job.CsExecutiveRole),
		},
		AgentDeadline:     timeFromPgtype(job.AgentDeadline),
		ShipmentReadyDate: timeFromPgtype(job.ShipmentReadyDate),
		Status:            common.PgtypeTextToStringPtr(job.Status),
		PriorityLevel:     int2ToStringPtr(job.PriorityLevel),
		CreatedAt:         job.CreatedAt.Time,
		CreatedBy:         common.PgtypeTextToStringPtr(job.CreatedBy),
		ModifiedAt:        timeFromPgtype(job.ModifiedAt),
		ModifiedBy:        common.PgtypeTextToStringPtr(job.ModifiedBy),
		IsActive:          job.IsActive.Bool,
		Packages:          []operationsdto.Package{},
		Documents:         []operationsdto.Document{},
		Billing:           []operationsdto.Billing{},
		Provisions:        []operationsdto.Provision{},
	}

	// Fetch packages
	packages, _ := s.repo.ListJobPackages(ctx, jobID)
	for _, pkg := range packages {
		detail.Packages = append(detail.Packages, packageFromSqlc(pkg))
	}

	// Fetch carrier
	carrier, err := s.repo.GetJobCarrier(ctx, jobID)
	if err == nil {
		c := carrierFromSqlc(carrier)
		detail.Carrier = &c
	}

	// Fetch documents
	docs, _ := s.repo.ListJobDocuments(ctx, jobID)
	for _, doc := range docs {
		detail.Documents = append(detail.Documents, documentFromSqlc(doc))
	}

	// Fetch billing
	billings, _ := s.repo.ListJobBilling(ctx, jobID)
	for _, b := range billings {
		detail.Billing = append(detail.Billing, billingFromSqlc(b))
	}

	// Fetch provisions
	provisions, _ := s.repo.ListJobProvisions(ctx, jobID)
	for _, p := range provisions {
		detail.Provisions = append(detail.Provisions, provisionFromSqlc(p))
	}

	// Fetch tracking
	tracking, err := s.repo.GetJobTracking(ctx, jobID)
	if err == nil {
		t := trackingFromSqlc(tracking)
		detail.Tracking = &t
	}

	logger.Info("fetched job", slog.String("jobCode", detail.JobCode))
	return detail, nil
}

// CreateJob creates a new job with related entities.
func (s *Service) CreateJob(ctx context.Context, input operationsdto.CreateJobInput) (operationsdto.JobDetail, error) {
	logger := logging.FromContext(ctx)

	// Generate job code (always auto-generated)
	jobCode, err := s.generateJobCode(ctx)
	if err != nil {
		return operationsdto.JobDetail{}, err
	}
	logger.Info("creating job with generated code", slog.String("jobCode", jobCode))

	// Create job
	jobParams := sqlc.CreateJobParams{
		JobCode:            jobCode,
		EnquiryNumber:      repository.NullTextFromString(input.EnquiryNumber),
		JobType:            repository.NullTextFromString(input.JobType),
		TransportMode:      repository.NullTextFromString(input.TransportMode),
		ServiceType:        repository.NullTextFromString(input.ServiceType),
		ServiceSubcategory: repository.NullTextFromString(input.ServiceSubcategory),
		ParentJobID:        repository.NullUUIDFromUUID(input.ParentJobID),
		CustomerID:         repository.NullUUIDFromUUID(input.CustomerID),
		AgentID:            repository.NullUUIDFromUUID(input.AgentID),
		ShipmentOrigin:     repository.NullTextFromString(input.ShipmentOrigin),
		DestinationCity:    repository.NullTextFromString(input.DestinationCity),
		DestinationState:   repository.NullTextFromString(input.DestinationState),
		DestinationCountry: repository.NullTextFromString(input.DestinationCountry),
		SourceCity:         repository.NullTextFromString(input.SourceCity),
		SourceState:        repository.NullTextFromString(input.SourceState),
		SourceCountry:      repository.NullTextFromString(input.SourceCountry),
		BranchID:           repository.NullUUIDFromUUID(input.BranchID),
		IncoTermCode:       repository.NullTextFromString(input.IncotermCode),
		Commodity:          repository.NullTextFromString(input.Commodity),
		Classification:     repository.NullTextFromString(input.Classification),
		SalesExecutiveID:   repository.NullUUIDFromUUID(input.SalesExecutiveID),
		OperationsExecID:   repository.NullUUIDFromUUID(input.OperationsExecID),
		CsExecutiveID:      repository.NullUUIDFromUUID(input.CSExecutiveID),
		AgentDeadline:      timestampFromTime(input.AgentDeadline),
		ShipmentReadyDate:  timestampFromTime(input.ShipmentReadyDate),
		Status:             repository.NullTextFromString(input.Status),
		PriorityLevel:      int2FromString(input.PriorityLevel),
		Actor:              pgtype.Text{String: input.CreatedBy, Valid: true},
	}

	job, err := s.repo.CreateJob(ctx, jobParams)
	if err != nil {
		logger.Error("failed to create job", slog.Any("error", err))
		return operationsdto.JobDetail{}, fmt.Errorf("operations: create job: %w", err)
	}

	// Create packages
	for _, pkgInput := range input.Packages {
		pkgParams := sqlc.CreateJobPackageParams{
			JobID:        job.ID,
			PackageName:  repository.NullTextFromString(pkgInput.PackageName),
			PackageType:  repository.NullTextFromString(pkgInput.PackageType),
			Quantity:     int4FromInt32(pkgInput.Quantity),
			LengthMeters: numericFromFloat64(pkgInput.LengthMeters),
			WidthMeters:  numericFromFloat64(pkgInput.WidthMeters),
			HeightMeters: numericFromFloat64(pkgInput.HeightMeters),
			WeightKg:     numericFromFloat64(pkgInput.WeightKg),
			VolumeCbm:    numericFromFloat64(pkgInput.VolumeCbm),
			HsCode:       repository.NullTextFromString(pkgInput.HSCode),
			CargoType:    repository.NullTextFromString(pkgInput.CargoType),
			ContainerID:  repository.NullTextFromString(pkgInput.ContainerID),
			Notes:        repository.NullTextFromString(pkgInput.Notes),
			Actor:        pgtype.Text{String: input.CreatedBy, Valid: true},
		}
		_, err = s.repo.CreateJobPackage(ctx, pkgParams)
		if err != nil {
			logger.Warn("failed to create package", slog.Any("error", err))
		}
	}

	// Create carrier
	if input.Carrier != nil {
		carrierParams := sqlc.CreateJobCarrierParams{
			JobID:                  job.ID,
			CarrierPartyID:         repository.NullUUIDFromUUID(input.Carrier.CarrierPartyID),
			CarrierName:            repository.NullTextFromString(input.Carrier.CarrierName),
			VesselName:             repository.NullTextFromString(input.Carrier.VesselName),
			VoyageNumber:           repository.NullTextFromString(input.Carrier.VoyageNumber),
			FlightID:               repository.NullTextFromString(input.Carrier.FlightID),
			FlightDate:             dateFromTime(input.Carrier.FlightDate),
			VehicleNumber:          repository.NullTextFromString(input.Carrier.VehicleNumber),
			RouteDetails:           repository.NullTextFromString(input.Carrier.RouteDetails),
			DriverName:             repository.NullTextFromString(input.Carrier.DriverName),
			OriginPortStation:      repository.NullTextFromString(input.Carrier.OriginPortStation),
			DestinationPortStation: repository.NullTextFromString(input.Carrier.DestinationPortStation),
			AccountingInfo:         repository.NullTextFromString(input.Carrier.AccountingInfo),
			HandlingInfo:           repository.NullTextFromString(input.Carrier.HandlingInfo),
			SupportingDocUrl:       repository.NullTextFromString(input.Carrier.SupportingDocURL),
			Actor:                  pgtype.Text{String: input.CreatedBy, Valid: true},
		}
		_, err = s.repo.CreateJobCarrier(ctx, carrierParams)
		if err != nil {
			logger.Warn("failed to create carrier", slog.Any("error", err))
		}
	}

	// Create documents
	for _, docInput := range input.Documents {
		docParams := sqlc.CreateJobDocumentParams{
			JobID:       job.ID,
			DocTypeCode: repository.NullTextFromString(docInput.DocTypeCode),
			DocNumber:   repository.NullTextFromString(docInput.DocNumber),
			IssuedAt:    timestampFromString(docInput.IssuedAt),
			IssuedDate:  dateFromTime(docInput.IssuedDate),
			Description: repository.NullTextFromString(docInput.Description),
			FileKey:     repository.NullTextFromString(docInput.FileKey),
			FileRegion:  repository.NullTextFromString(docInput.FileRegion),
			Actor:       pgtype.Text{String: input.CreatedBy, Valid: true},
		}
		_, err = s.repo.CreateJobDocument(ctx, docParams)
		if err != nil {
			logger.Warn("failed to create document", slog.Any("error", err))
		}
	}

	// Create billing entries
	for _, billInput := range input.Billing {
		billParams := sqlc.CreateJobBillingParams{
			JobID:                 job.ID,
			ActivityType:          repository.NullTextFromString(billInput.ActivityType),
			ActivityCode:          repository.NullTextFromString(billInput.ActivityCode),
			BillingPartyID:        repository.NullUUIDFromUUID(billInput.BillingPartyID),
			PoNumber:              repository.NullTextFromString(billInput.PONumber),
			PoDate:                dateFromTime(billInput.PODate),
			CurrencyCode:          repository.NullTextFromString(billInput.CurrencyCode),
			Amount:                numericFromFloat64(billInput.Amount),
			Description:           repository.NullTextFromString(billInput.Description),
			Notes:                 repository.NullTextFromString(billInput.Notes),
			SupportingDocUrl:      repository.NullTextFromString(billInput.SupportingDocURL),
			AmountPrimaryCurrency: numericFromFloat64(billInput.AmountPrimaryCurrency),
			Actor:                 pgtype.Text{String: input.CreatedBy, Valid: true},
		}
		_, err = s.repo.CreateJobBilling(ctx, billParams)
		if err != nil {
			logger.Warn("failed to create billing", slog.Any("error", err))
		}
	}

	// Create provision entries
	for _, provInput := range input.Provisions {
		provParams := sqlc.CreateJobProvisionParams{
			JobID:                 job.ID,
			ActivityType:          repository.NullTextFromString(provInput.ActivityType),
			ActivityCode:          repository.NullTextFromString(provInput.ActivityCode),
			CostPartyID:           repository.NullUUIDFromUUID(provInput.CostPartyID),
			InvoiceNumber:         repository.NullTextFromString(provInput.InvoiceNumber),
			InvoiceDate:           dateFromTime(provInput.InvoiceDate),
			CurrencyCode:          repository.NullTextFromString(provInput.CurrencyCode),
			Amount:                numericFromFloat64(provInput.Amount),
			PaymentPriority:       repository.NullTextFromString(provInput.PaymentPriority),
			Notes:                 repository.NullTextFromString(provInput.Notes),
			SupportingDocUrl:      repository.NullTextFromString(provInput.SupportingDocURL),
			AmountPrimaryCurrency: numericFromFloat64(provInput.AmountPrimaryCurrency),
			Profit:                numericFromFloat64(provInput.Profit),
			Actor:                 pgtype.Text{String: input.CreatedBy, Valid: true},
		}
		_, err = s.repo.CreateJobProvision(ctx, provParams)
		if err != nil {
			logger.Warn("failed to create provision", slog.Any("error", err))
		}
	}

	// Create tracking
	if input.Tracking != nil {
		trackParams := sqlc.UpsertJobTrackingParams{
			JobID:          job.ID,
			EtdDate:        timestampFromTime(input.Tracking.ETDDate),
			EtaDate:        timestampFromTime(input.Tracking.ETADate),
			AtdDate:        timestampFromTime(input.Tracking.ATDDate),
			AtaDate:        timestampFromTime(input.Tracking.ATADate),
			JobStatus:      repository.NullTextFromString(input.Tracking.JobStatus),
			PodStatus:      repository.NullTextFromString(input.Tracking.PODStatus),
			DocumentStatus: repository.NullTextFromString(input.Tracking.DocumentStatus),
			Notes:          repository.NullTextFromString(input.Tracking.Notes),
			Actor:          pgtype.Text{String: input.CreatedBy, Valid: true},
		}
		_, err = s.repo.UpsertJobTracking(ctx, trackParams)
		if err != nil {
			logger.Warn("failed to create tracking", slog.Any("error", err))
		}
	}

	logger.Info("created job", slog.String("jobID", job.ID.String()))
	return s.GetJob(ctx, job.ID)
}

// UpdateJob updates an existing job.
func (s *Service) UpdateJob(ctx context.Context, jobID uuid.UUID, input operationsdto.UpdateJobInput) (operationsdto.JobDetail, error) {
	logger := logging.FromContext(ctx)
	logger.Info("updating job", slog.String("jobID", jobID.String()))

	params := sqlc.UpdateJobParams{
		ID:                 jobID,
		EnquiryNumber:      repository.NullTextFromString(input.EnquiryNumber),
		JobType:            repository.NullTextFromString(input.JobType),
		TransportMode:      repository.NullTextFromString(input.TransportMode),
		ServiceType:        repository.NullTextFromString(input.ServiceType),
		ServiceSubcategory: repository.NullTextFromString(input.ServiceSubcategory),
		ParentJobID:        repository.NullUUIDFromUUID(input.ParentJobID),
		CustomerID:         repository.NullUUIDFromUUID(input.CustomerID),
		AgentID:            repository.NullUUIDFromUUID(input.AgentID),
		ShipmentOrigin:     repository.NullTextFromString(input.ShipmentOrigin),
		DestinationCity:    repository.NullTextFromString(input.DestinationCity),
		DestinationState:   repository.NullTextFromString(input.DestinationState),
		DestinationCountry: repository.NullTextFromString(input.DestinationCountry),
		SourceCity:         repository.NullTextFromString(input.SourceCity),
		SourceState:        repository.NullTextFromString(input.SourceState),
		SourceCountry:      repository.NullTextFromString(input.SourceCountry),
		BranchID:           repository.NullUUIDFromUUID(input.BranchID),
		IncoTermCode:       repository.NullTextFromString(input.IncotermCode),
		Commodity:          repository.NullTextFromString(input.Commodity),
		Classification:     repository.NullTextFromString(input.Classification),
		SalesExecutiveID:   repository.NullUUIDFromUUID(input.SalesExecutiveID),
		OperationsExecID:   repository.NullUUIDFromUUID(input.OperationsExecID),
		CsExecutiveID:      repository.NullUUIDFromUUID(input.CSExecutiveID),
		AgentDeadline:      timestampFromTime(input.AgentDeadline),
		ShipmentReadyDate:  timestampFromTime(input.ShipmentReadyDate),
		Status:             repository.NullTextFromString(input.Status),
		PriorityLevel:      int2FromString(input.PriorityLevel),
		Actor:              pgtype.Text{String: input.ModifiedBy, Valid: true},
	}

	_, err := s.repo.UpdateJob(ctx, params)
	if err != nil {
		logger.Error("failed to update job", slog.Any("error", err))
		return operationsdto.JobDetail{}, fmt.Errorf("operations: update job: %w", err)
	}

	// Update packages if provided
	for _, pkgInput := range input.Packages {
		pkgParams := sqlc.CreateJobPackageParams{
			JobID:        jobID,
			PackageName:  repository.NullTextFromString(pkgInput.PackageName),
			PackageType:  repository.NullTextFromString(pkgInput.PackageType),
			Quantity:     int4FromInt32(pkgInput.Quantity),
			LengthMeters: numericFromFloat64(pkgInput.LengthMeters),
			WidthMeters:  numericFromFloat64(pkgInput.WidthMeters),
			HeightMeters: numericFromFloat64(pkgInput.HeightMeters),
			WeightKg:     numericFromFloat64(pkgInput.WeightKg),
			VolumeCbm:    numericFromFloat64(pkgInput.VolumeCbm),
			HsCode:       repository.NullTextFromString(pkgInput.HSCode),
			CargoType:    repository.NullTextFromString(pkgInput.CargoType),
			ContainerID:  repository.NullTextFromString(pkgInput.ContainerID),
			Notes:        repository.NullTextFromString(pkgInput.Notes),
			Actor:        pgtype.Text{String: input.ModifiedBy, Valid: true},
		}
		_, err = s.repo.CreateJobPackage(ctx, pkgParams)
		if err != nil {
			logger.Warn("failed to create package", slog.Any("error", err))
		}
	}

	// Update carrier if provided
	if input.Carrier != nil {
		// Try to update existing carrier first
		updateParams := sqlc.UpdateJobCarrierParams{
			JobID:                  jobID,
			CarrierPartyID:         repository.NullUUIDFromUUID(input.Carrier.CarrierPartyID),
			CarrierName:            repository.NullTextFromString(input.Carrier.CarrierName),
			VesselName:             repository.NullTextFromString(input.Carrier.VesselName),
			VoyageNumber:           repository.NullTextFromString(input.Carrier.VoyageNumber),
			FlightID:               repository.NullTextFromString(input.Carrier.FlightID),
			FlightDate:             dateFromTime(input.Carrier.FlightDate),
			VehicleNumber:          repository.NullTextFromString(input.Carrier.VehicleNumber),
			RouteDetails:           repository.NullTextFromString(input.Carrier.RouteDetails),
			DriverName:             repository.NullTextFromString(input.Carrier.DriverName),
			OriginPortStation:      repository.NullTextFromString(input.Carrier.OriginPortStation),
			DestinationPortStation: repository.NullTextFromString(input.Carrier.DestinationPortStation),
			AccountingInfo:         repository.NullTextFromString(input.Carrier.AccountingInfo),
			HandlingInfo:           repository.NullTextFromString(input.Carrier.HandlingInfo),
			SupportingDocUrl:       repository.NullTextFromString(input.Carrier.SupportingDocURL),
			Actor:                  pgtype.Text{String: input.ModifiedBy, Valid: true},
		}
		_, err = s.repo.UpdateJobCarrier(ctx, updateParams)
		if err != nil {
			// If update fails (no rows), create new carrier
			if errors.Is(err, pgx.ErrNoRows) {
				createParams := sqlc.CreateJobCarrierParams{
					JobID:                  jobID,
					CarrierPartyID:         repository.NullUUIDFromUUID(input.Carrier.CarrierPartyID),
					CarrierName:            repository.NullTextFromString(input.Carrier.CarrierName),
					VesselName:             repository.NullTextFromString(input.Carrier.VesselName),
					VoyageNumber:           repository.NullTextFromString(input.Carrier.VoyageNumber),
					FlightID:               repository.NullTextFromString(input.Carrier.FlightID),
					FlightDate:             dateFromTime(input.Carrier.FlightDate),
					VehicleNumber:          repository.NullTextFromString(input.Carrier.VehicleNumber),
					RouteDetails:           repository.NullTextFromString(input.Carrier.RouteDetails),
					DriverName:             repository.NullTextFromString(input.Carrier.DriverName),
					OriginPortStation:      repository.NullTextFromString(input.Carrier.OriginPortStation),
					DestinationPortStation: repository.NullTextFromString(input.Carrier.DestinationPortStation),
					AccountingInfo:         repository.NullTextFromString(input.Carrier.AccountingInfo),
					HandlingInfo:           repository.NullTextFromString(input.Carrier.HandlingInfo),
					SupportingDocUrl:       repository.NullTextFromString(input.Carrier.SupportingDocURL),
					Actor:                  pgtype.Text{String: input.ModifiedBy, Valid: true},
				}
				_, err = s.repo.CreateJobCarrier(ctx, createParams)
				if err != nil {
					logger.Warn("failed to create carrier", slog.Any("error", err))
				}
			} else {
				logger.Warn("failed to update carrier", slog.Any("error", err))
			}
		}
	}

	// Update documents if provided
	for _, docInput := range input.Documents {
		docParams := sqlc.CreateJobDocumentParams{
			JobID:       jobID,
			DocTypeCode: repository.NullTextFromString(docInput.DocTypeCode),
			DocNumber:   repository.NullTextFromString(docInput.DocNumber),
			IssuedAt:    timestampFromString(docInput.IssuedAt),
			IssuedDate:  dateFromTime(docInput.IssuedDate),
			Description: repository.NullTextFromString(docInput.Description),
			FileKey:     repository.NullTextFromString(docInput.FileKey),
			FileRegion:  repository.NullTextFromString(docInput.FileRegion),
			Actor:       pgtype.Text{String: input.ModifiedBy, Valid: true},
		}
		_, err = s.repo.CreateJobDocument(ctx, docParams)
		if err != nil {
			logger.Warn("failed to create document", slog.Any("error", err))
		}
	}

	// Update billing if provided
	for _, billInput := range input.Billing {
		billParams := sqlc.CreateJobBillingParams{
			JobID:                 jobID,
			ActivityType:          repository.NullTextFromString(billInput.ActivityType),
			ActivityCode:          repository.NullTextFromString(billInput.ActivityCode),
			BillingPartyID:        repository.NullUUIDFromUUID(billInput.BillingPartyID),
			PoNumber:              repository.NullTextFromString(billInput.PONumber),
			PoDate:                dateFromTime(billInput.PODate),
			CurrencyCode:          repository.NullTextFromString(billInput.CurrencyCode),
			Amount:                numericFromFloat64(billInput.Amount),
			Description:           repository.NullTextFromString(billInput.Description),
			Notes:                 repository.NullTextFromString(billInput.Notes),
			SupportingDocUrl:      repository.NullTextFromString(billInput.SupportingDocURL),
			AmountPrimaryCurrency: numericFromFloat64(billInput.AmountPrimaryCurrency),
			Actor:                 pgtype.Text{String: input.ModifiedBy, Valid: true},
		}
		_, err = s.repo.CreateJobBilling(ctx, billParams)
		if err != nil {
			logger.Warn("failed to create billing", slog.Any("error", err))
		}
	}

	// Update provisions if provided
	for _, provInput := range input.Provisions {
		provParams := sqlc.CreateJobProvisionParams{
			JobID:                 jobID,
			ActivityType:          repository.NullTextFromString(provInput.ActivityType),
			ActivityCode:          repository.NullTextFromString(provInput.ActivityCode),
			CostPartyID:           repository.NullUUIDFromUUID(provInput.CostPartyID),
			InvoiceNumber:         repository.NullTextFromString(provInput.InvoiceNumber),
			InvoiceDate:           dateFromTime(provInput.InvoiceDate),
			CurrencyCode:          repository.NullTextFromString(provInput.CurrencyCode),
			Amount:                numericFromFloat64(provInput.Amount),
			PaymentPriority:       repository.NullTextFromString(provInput.PaymentPriority),
			Notes:                 repository.NullTextFromString(provInput.Notes),
			SupportingDocUrl:      repository.NullTextFromString(provInput.SupportingDocURL),
			AmountPrimaryCurrency: numericFromFloat64(provInput.AmountPrimaryCurrency),
			Profit:                numericFromFloat64(provInput.Profit),
			Actor:                 pgtype.Text{String: input.ModifiedBy, Valid: true},
		}
		_, err = s.repo.CreateJobProvision(ctx, provParams)
		if err != nil {
			logger.Warn("failed to create provision", slog.Any("error", err))
		}
	}

	// Update tracking if provided
	if input.Tracking != nil {
		trackParams := sqlc.UpsertJobTrackingParams{
			JobID:          jobID,
			EtdDate:        timestampFromTime(input.Tracking.ETDDate),
			EtaDate:        timestampFromTime(input.Tracking.ETADate),
			AtdDate:        timestampFromTime(input.Tracking.ATDDate),
			AtaDate:        timestampFromTime(input.Tracking.ATADate),
			JobStatus:      repository.NullTextFromString(input.Tracking.JobStatus),
			PodStatus:      repository.NullTextFromString(input.Tracking.PODStatus),
			DocumentStatus: repository.NullTextFromString(input.Tracking.DocumentStatus),
			Notes:          repository.NullTextFromString(input.Tracking.Notes),
			Actor:          pgtype.Text{String: input.ModifiedBy, Valid: true},
		}
		_, err = s.repo.UpsertJobTracking(ctx, trackParams)
		if err != nil {
			logger.Warn("failed to update tracking", slog.Any("error", err))
		}
	}

	logger.Info("updated job", slog.String("jobID", jobID.String()))
	return s.GetJob(ctx, jobID)
}

// ArchiveJob soft deletes a job.
func (s *Service) ArchiveJob(ctx context.Context, jobID uuid.UUID, actor string) error {
	logger := logging.FromContext(ctx)
	logger.Info("archiving job", slog.String("jobID", jobID.String()))

	err := s.repo.ArchiveJob(ctx, jobID, actor)
	if err != nil {
		logger.Error("failed to archive job", slog.Any("error", err))
		return fmt.Errorf("operations: archive job: %w", err)
	}

	logger.Info("archived job", slog.String("jobID", jobID.String()))
	return nil
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

func uuidFromPgtype(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	uid := uuid.UUID(u.Bytes)
	return &uid
}

func timeFromPgtype(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func timestampFromTime(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func int4FromInt32(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

func float8FromFloat64(f *float64) pgtype.Float8 {
	if f == nil {
		return pgtype.Float8{Valid: false}
	}
	return pgtype.Float8{Float64: *f, Valid: true}
}

func numericFromFloat64(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{Valid: false}
	}
	// Convert float64 to pgtype.Numeric
	var num pgtype.Numeric
	if err := num.Scan(fmt.Sprintf("%f", *f)); err != nil {
		return pgtype.Numeric{Valid: false}
	}
	return num
}

func dateFromTime(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: *t, Valid: true}
}

func uuidToPgtype(u uuid.UUID) pgtype.UUID {
	var pg pgtype.UUID
	copy(pg.Bytes[:], u[:])
	pg.Valid = true
	return pg
}

func packageFromSqlc(pkg sqlc.OpsPackage) operationsdto.Package {
	return operationsdto.Package{
		ID:           pkg.ID,
		PackageName:  common.PgtypeTextToStringPtr(pkg.PackageName),
		PackageType:  common.PgtypeTextToStringPtr(pkg.PackageType),
		Quantity:     int32FromPgtype(pkg.Quantity),
		LengthMeters: float64FromNumeric(pkg.LengthMeters),
		WidthMeters:  float64FromNumeric(pkg.WidthMeters),
		HeightMeters: float64FromNumeric(pkg.HeightMeters),
		WeightKg:     float64FromNumeric(pkg.WeightKg),
		VolumeCbm:    float64FromNumeric(pkg.VolumeCbm),
		HSCode:       common.PgtypeTextToStringPtr(pkg.HsCode),
		CargoType:    common.PgtypeTextToStringPtr(pkg.CargoType),
		ContainerID:  common.PgtypeTextToStringPtr(pkg.ContainerID),
		Notes:        common.PgtypeTextToStringPtr(pkg.Notes),
	}
}

func carrierFromSqlc(carrier sqlc.OpsCarrier) operationsdto.Carrier {
	var flightDate *time.Time
	if carrier.FlightDate.Valid {
		flightDate = &carrier.FlightDate.Time
	}
	return operationsdto.Carrier{
		ID:                     carrier.ID,
		CarrierPartyID:         uuidFromPgtype(carrier.CarrierPartyID),
		CarrierName:            common.PgtypeTextToStringPtr(carrier.CarrierName),
		VesselName:             common.PgtypeTextToStringPtr(carrier.VesselName),
		VoyageNumber:           common.PgtypeTextToStringPtr(carrier.VoyageNumber),
		FlightID:               common.PgtypeTextToStringPtr(carrier.FlightID),
		FlightDate:             flightDate,
		VehicleNumber:          common.PgtypeTextToStringPtr(carrier.VehicleNumber),
		RouteDetails:           common.PgtypeTextToStringPtr(carrier.RouteDetails),
		DriverName:             common.PgtypeTextToStringPtr(carrier.DriverName),
		OriginPortStation:      common.PgtypeTextToStringPtr(carrier.OriginPortStation),
		DestinationPortStation: common.PgtypeTextToStringPtr(carrier.DestinationPortStation),
		AccountingInfo:         common.PgtypeTextToStringPtr(carrier.AccountingInfo),
		HandlingInfo:           common.PgtypeTextToStringPtr(carrier.HandlingInfo),
		SupportingDocURL:       common.PgtypeTextToStringPtr(carrier.SupportingDocUrl),
	}
}

func documentFromSqlc(doc sqlc.OpsJobDocument) operationsdto.Document {
	return operationsdto.Document{
		ID:          doc.ID,
		DocTypeCode: common.PgtypeTextToStringPtr(doc.DocTypeCode),
		DocNumber:   common.PgtypeTextToStringPtr(doc.DocNumber),
		IssuedAt:    timestampToStringPtr(doc.IssuedAt),
		IssuedDate:  timeFromDate(doc.IssuedDate),
		Description: common.PgtypeTextToStringPtr(doc.Description),
		FileKey:     common.PgtypeTextToStringPtr(doc.FileKey),
		FileRegion:  common.PgtypeTextToStringPtr(doc.FileRegion),
	}
}

func billingFromSqlc(b sqlc.ListJobBillingRow) operationsdto.Billing {
	var poDate *time.Time
	if b.PoDate.Valid {
		poDate = &b.PoDate.Time
	}
	return operationsdto.Billing{
		ID:                    b.ID,
		ActivityType:          common.PgtypeTextToStringPtr(b.ActivityType),
		ActivityCode:          common.PgtypeTextToStringPtr(b.ActivityCode),
		BillingPartyID:        uuidFromPgtype(b.BillingPartyID),
		BillingPartyName:      common.PgtypeTextToStringPtr(b.BillingPartyName),
		PONumber:              common.PgtypeTextToStringPtr(b.PoNumber),
		PODate:                poDate,
		CurrencyCode:          common.PgtypeTextToStringPtr(b.CurrencyCode),
		Amount:                float64FromNumeric(b.Amount),
		Description:           common.PgtypeTextToStringPtr(b.Description),
		Notes:                 common.PgtypeTextToStringPtr(b.Notes),
		SupportingDocURL:      common.PgtypeTextToStringPtr(b.SupportingDocUrl),
		AmountPrimaryCurrency: float64FromNumeric(b.AmountPrimaryCurrency),
	}
}

func provisionFromSqlc(p sqlc.ListJobProvisionsRow) operationsdto.Provision {
	var invDate *time.Time
	if p.InvoiceDate.Valid {
		invDate = &p.InvoiceDate.Time
	}
	return operationsdto.Provision{
		ID:                    p.ID,
		ActivityType:          common.PgtypeTextToStringPtr(p.ActivityType),
		ActivityCode:          common.PgtypeTextToStringPtr(p.ActivityCode),
		CostPartyID:           uuidFromPgtype(p.CostPartyID),
		CostPartyName:         common.PgtypeTextToStringPtr(p.CostPartyName),
		InvoiceNumber:         common.PgtypeTextToStringPtr(p.InvoiceNumber),
		InvoiceDate:           invDate,
		CurrencyCode:          common.PgtypeTextToStringPtr(p.CurrencyCode),
		Amount:                float64FromNumeric(p.Amount),
		PaymentPriority:       common.PgtypeTextToStringPtr(p.PaymentPriority),
		Notes:                 common.PgtypeTextToStringPtr(p.Notes),
		SupportingDocURL:      common.PgtypeTextToStringPtr(p.SupportingDocUrl),
		AmountPrimaryCurrency: float64FromNumeric(p.AmountPrimaryCurrency),
		Profit:                float64FromNumeric(p.Profit),
	}
}

func trackingFromSqlc(t sqlc.OpsTracking) operationsdto.Tracking {
	return operationsdto.Tracking{
		ID:             t.ID,
		ETDDate:        timeFromPgtype(t.EtdDate),
		ETADate:        timeFromPgtype(t.EtaDate),
		ATDDate:        timeFromPgtype(t.AtdDate),
		ATADate:        timeFromPgtype(t.AtaDate),
		JobStatus:      common.PgtypeTextToStringPtr(t.JobStatus),
		PODStatus:      common.PgtypeTextToStringPtr(t.PodStatus),
		DocumentStatus: common.PgtypeTextToStringPtr(t.DocumentStatus),
		Notes:          common.PgtypeTextToStringPtr(t.Notes),
	}
}

func int32FromPgtype(i pgtype.Int4) *int32 {
	if !i.Valid {
		return nil
	}
	return &i.Int32
}

func float64FromPgtype(f pgtype.Float8) *float64 {
	if !f.Valid {
		return nil
	}
	return &f.Float64
}

func float64FromNumeric(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	// Convert pgtype.Numeric to float64
	f, err := n.Float64Value()
	if err != nil {
		return nil
	}
	if f.Valid {
		return &f.Float64
	}
	return nil
}

func timeFromDate(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	return &d.Time
}

func int2FromString(s *string) pgtype.Int2 {
	if s == nil {
		return pgtype.Int2{Valid: false}
	}
	var i int16
	_, err := fmt.Sscanf(*s, "%d", &i)
	if err != nil {
		return pgtype.Int2{Valid: false}
	}
	return pgtype.Int2{Int16: i, Valid: true}
}

func int2ToStringPtr(i pgtype.Int2) *string {
	if !i.Valid {
		return nil
	}
	s := fmt.Sprintf("%d", i.Int16)
	return &s
}

func timestampFromString(s *string) pgtype.Timestamptz {
	if s == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func timestampToStringPtr(t pgtype.Timestamptz) *string {
	if !t.Valid {
		return nil
	}
	s := t.Time.Format(time.RFC3339)
	return &s
}
