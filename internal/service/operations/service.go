package operations

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
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
			JobStatusID:   row.JobStatusID,
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
			DocStatusID:   row.DocStatusID,
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
			PriorityLabel: common.PgtypeTextToString(row.PriorityLabel),
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
			RoleID:   row.RoleID,
			RoleName: common.PgtypeTextToString(row.RoleName),
			RoleDesc: desc,
		})
	}

	// Fetch activities
	activityRows, err := s.repo.ListActivityLookups(ctx)
	if err != nil {
		logger.Error("failed to list activity lookups", slog.Any("error", err))
		return result, fmt.Errorf("operations: list activities: %w", err)
	}
	for _, row := range activityRows {
		result.Activities = append(result.Activities, operationsdto.ActivityLookup{
			ActivityType: row.ActivityType,
			ActivityCode: row.ActivityCode,
		})
	}

	logger.Info("fetched operations lookups",
		slog.Int("transport_modes", len(result.TransportModes)),
		slog.Int("incoterms", len(result.Incoterms)),
		slog.Int("job_statuses", len(result.JobStatuses)),
		slog.Int("document_statuses", len(result.DocumentStatuses)),
		slog.Int("role_details", len(result.RoleDetails)),
		slog.Int("activities", len(result.Activities)),
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
			PriorityLevel:      textToStringPtr(row.PriorityLevel),
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
		PriorityLevel:     textToStringPtr(job.PriorityLevel),
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

	// Fetch carriers
	carriers, err := s.repo.GetJobCarriers(ctx, jobID)
	if err == nil && len(carriers) > 0 {
		c := carrierFromSqlc(carriers[0])
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
		BranchName:         repository.NullTextFromString(input.BranchName),
		IncoTermCode:       repository.NullTextFromString(input.IncotermCode),
		Commodity:          repository.NullTextFromString(input.Commodity),
		Classification:     repository.NullTextFromString(input.Classification),
		SalesExecutiveID:   repository.NullUUIDFromUUID(input.SalesExecutiveID),
		SalesExecutiveName: repository.NullTextFromString(input.SalesExecutiveName),
		OperationsExecID:   repository.NullUUIDFromUUID(input.OperationsExecID),
		OperationsExecName: repository.NullTextFromString(input.OperationsExecName),
		CsExecutiveID:      repository.NullUUIDFromUUID(input.CSExecutiveID),
		CsExecutiveName:    repository.NullTextFromString(input.CSExecutiveName),
		AgentDeadline:      timestampFromTime(input.AgentDeadline),
		ShipmentReadyDate:  timestampFromTime(input.ShipmentReadyDate),
		Status:             repository.NullTextFromString(input.Status),
		PriorityLevel:      textFromString(input.PriorityLevel),
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
			JobID:                     job.ID,
			ContainerNo:               repository.NullTextFromString(pkgInput.ContainerNo),
			ContainerType:             repository.NullTextFromString(pkgInput.ContainerType),
			ContainerSize:             repository.NullTextFromString(pkgInput.ContainerSize),
			GrossWeightKg:             numericFromFloat64(pkgInput.GrossWeightKg),
			NetWeightKg:               numericFromFloat64(pkgInput.NetWeightKg),
			Volume:                    numericFromFloat64(pkgInput.Volume),
			CarrierSealNo:             repository.NullTextFromString(pkgInput.CarrierSealNo),
			CommodityCargoDescription: repository.NullTextFromString(pkgInput.CommodityCargoDescription),
			PackageType:               repository.NullTextFromString(pkgInput.PackageType),
			CargoType:                 repository.NullTextFromString(pkgInput.CargoType),
			NoOfPackages:              numericFromFloat64(pkgInput.NoOfPackages),
			ChargeableWeight:          numericFromFloat64(pkgInput.ChargeableWeight),
			HsCode:                    repository.NullTextFromString(pkgInput.HSCode),
			TemperatureControl:        pgtype.Bool{Bool: pkgInput.TemperatureControl, Valid: true},
			Actor:                     pgtype.Text{String: input.CreatedBy, Valid: true},
		}
		_, err = s.repo.CreateJobPackage(ctx, pkgParams)
		if err != nil {
			logger.Warn("failed to create package", slog.Any("error", err))
		}
	}

	// Create carrier
	if input.Carrier != nil {
		carrierParams := sqlc.CreateJobCarrierParams{
			JobID:                      uuidToPgtype(job.ID),
			CarrierPartyID:             repository.NullUUIDFromUUID(input.Carrier.CarrierPartyID),
			CarrierName:                repository.NullTextFromString(input.Carrier.CarrierName),
			CarrierContact:             repository.NullTextFromString(input.Carrier.CarrierContact),
			VesselName:                 repository.NullTextFromString(input.Carrier.VesselName),
			VoyageNumber:               repository.NullTextFromString(input.Carrier.VoyageNumber),
			FlightID:                   repository.NullTextFromString(input.Carrier.FlightID),
			FlightDate:                 timestampFromTime(input.Carrier.FlightDate),
			AirportReportDate:          timestampFromTime(input.Carrier.AirportReportDate),
			VehicleNumber:              repository.NullTextFromString(input.Carrier.VehicleNumber),
			VehicleType:                repository.NullTextFromString(input.Carrier.VehicleType),
			RouteDetails:               repository.NullTextFromString(input.Carrier.RouteDetails),
			DriverName:                 repository.NullTextFromString(input.Carrier.DriverName),
			DriverContact:              repository.NullTextFromString(input.Carrier.DriverContact),
			OriginPortStation:          repository.NullTextFromString(input.Carrier.OriginPortStation),
			DestinationPortStation:     repository.NullTextFromString(input.Carrier.DestinationPortStation),
			OriginCountry:              repository.NullTextFromString(input.Carrier.OriginCountry),
			DestinationCountry:         repository.NullTextFromString(input.Carrier.DestinationCountry),
			AccountingInfo:             repository.NullTextFromString(input.Carrier.AccountingInfo),
			HandlingInfo:               repository.NullTextFromString(input.Carrier.HandlingInfo),
			TransportDocumentReference: repository.NullTextFromString(input.Carrier.TransportDocumentRef),
			DocUrls:                    input.Carrier.SupportingDocURLs,
			FileRegion:                 repository.NullTextFromString(input.Carrier.FileRegion),
			Description:                repository.NullTextFromString(input.Carrier.Description),
			Actor:                      pgtype.Text{String: input.CreatedBy, Valid: true},
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
			IssuedDate:  timestampFromTime(docInput.IssuedDate),
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
			JobID:                 uuidToPgtype(job.ID),
			ActivityType:          repository.NullTextFromString(billInput.ActivityType),
			ActivityCode:          repository.NullTextFromString(billInput.ActivityCode),
			BillingPartyID:        repository.NullUUIDFromUUID(billInput.BillingPartyID),
			PoNumber:              repository.NullTextFromString(billInput.PONumber),
			PoDate:                timestampFromTime(billInput.PODate),
			CurrencyCode:          repository.NullTextFromString(billInput.CurrencyCode),
			Quantity:              numericFromFloat64(billInput.Quantity),
			UnitPrice:             numericFromFloat64(billInput.UnitPrice),
			AmountWithoutTax:      numericFromFloat64(billInput.AmountWithoutTax),
			TaxCode:               repository.NullTextFromString(billInput.TaxCode),
			TaxAmount:             numericFromFloat64(billInput.TaxAmount),
			ExchangeRate:          numericFromFloat64(billInput.ExchangeRate),
			TotalAmount:           numericFromFloat64(billInput.TotalAmount),
			Description:           repository.NullTextFromString(billInput.Description),
			Notes:                 repository.NullTextFromString(billInput.Notes),
			DocUrls:               billInput.SupportingDocURLs,
			FileRegion:            repository.NullTextFromString(billInput.FileRegion),
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
			JobID:                 uuidToPgtype(job.ID),
			ActivityType:          repository.NullTextFromString(provInput.ActivityType),
			ActivityCode:          repository.NullTextFromString(provInput.ActivityCode),
			CostPartyID:           repository.NullUUIDFromUUID(provInput.CostPartyID),
			InvoiceNumber:         repository.NullTextFromString(provInput.InvoiceNumber),
			InvoiceDate:           timestampFromTime(provInput.InvoiceDate),
			CurrencyCode:          repository.NullTextFromString(provInput.CurrencyCode),
			Quantity:              numericFromFloat64(provInput.Quantity),
			UnitPrice:             numericFromFloat64(provInput.UnitPrice),
			AmountWithoutTax:      numericFromFloat64(provInput.AmountWithoutTax),
			TaxCode:               repository.NullTextFromString(provInput.TaxCode),
			TaxAmount:             numericFromFloat64(provInput.TaxAmount),
			TotalAmount:           numericFromFloat64(provInput.TotalAmount),
			PoNumber:              repository.NullTextFromString(provInput.PONumber),
			PoDate:                timestampFromTime(provInput.PODate),
			ExchangeRate:          numericFromFloat64(provInput.ExchangeRate),
			PaymentPriority:       repository.NullTextFromString(provInput.PaymentPriority),
			Notes:                 repository.NullTextFromString(provInput.Notes),
			DocUrls:               provInput.SupportingDocURLs,
			FileRegion:            repository.NullTextFromString(provInput.FileRegion),
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
			JobID:          uuidToPgtype(job.ID),
			EtdDate:        timestampFromTime(input.Tracking.ETDDate),
			EtaDate:        timestampFromTime(input.Tracking.ETADate),
			AtdDate:        timestampFromTime(input.Tracking.ATDDate),
			AtaDate:        timestampFromTime(input.Tracking.ATADate),
			JobStatus:      repository.NullTextFromString(input.Tracking.JobStatus),
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
		BranchName:         repository.NullTextFromString(input.BranchName),
		IncoTermCode:       repository.NullTextFromString(input.IncotermCode),
		Commodity:          repository.NullTextFromString(input.Commodity),
		Classification:     repository.NullTextFromString(input.Classification),
		SalesExecutiveID:   repository.NullUUIDFromUUID(input.SalesExecutiveID),
		SalesExecutiveName: repository.NullTextFromString(input.SalesExecutiveName),
		OperationsExecID:   repository.NullUUIDFromUUID(input.OperationsExecID),
		OperationsExecName: repository.NullTextFromString(input.OperationsExecName),
		CsExecutiveID:      repository.NullUUIDFromUUID(input.CSExecutiveID),
		CsExecutiveName:    repository.NullTextFromString(input.CSExecutiveName),
		AgentDeadline:      timestampFromTime(input.AgentDeadline),
		ShipmentReadyDate:  timestampFromTime(input.ShipmentReadyDate),
		Status:             repository.NullTextFromString(input.Status),
		PriorityLevel:      textFromString(input.PriorityLevel),
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
			JobID:                     jobID,
			ContainerNo:               repository.NullTextFromString(pkgInput.ContainerNo),
			ContainerType:             repository.NullTextFromString(pkgInput.ContainerType),
			ContainerSize:             repository.NullTextFromString(pkgInput.ContainerSize),
			GrossWeightKg:             numericFromFloat64(pkgInput.GrossWeightKg),
			NetWeightKg:               numericFromFloat64(pkgInput.NetWeightKg),
			Volume:                    numericFromFloat64(pkgInput.Volume),
			CarrierSealNo:             repository.NullTextFromString(pkgInput.CarrierSealNo),
			CommodityCargoDescription: repository.NullTextFromString(pkgInput.CommodityCargoDescription),
			PackageType:               repository.NullTextFromString(pkgInput.PackageType),
			CargoType:                 repository.NullTextFromString(pkgInput.CargoType),
			NoOfPackages:              numericFromFloat64(pkgInput.NoOfPackages),
			ChargeableWeight:          numericFromFloat64(pkgInput.ChargeableWeight),
			HsCode:                    repository.NullTextFromString(pkgInput.HSCode),
			TemperatureControl:        pgtype.Bool{Bool: pkgInput.TemperatureControl, Valid: true},
			Actor:                     pgtype.Text{String: input.ModifiedBy, Valid: true},
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
			JobID:                      uuidToPgtype(jobID),
			CarrierPartyID:             repository.NullUUIDFromUUID(input.Carrier.CarrierPartyID),
			CarrierName:                repository.NullTextFromString(input.Carrier.CarrierName),
			CarrierContact:             repository.NullTextFromString(input.Carrier.CarrierContact),
			VesselName:                 repository.NullTextFromString(input.Carrier.VesselName),
			VoyageNumber:               repository.NullTextFromString(input.Carrier.VoyageNumber),
			FlightID:                   repository.NullTextFromString(input.Carrier.FlightID),
			FlightDate:                 timestampFromTime(input.Carrier.FlightDate),
			AirportReportDate:          timestampFromTime(input.Carrier.AirportReportDate),
			VehicleNumber:              repository.NullTextFromString(input.Carrier.VehicleNumber),
			VehicleType:                repository.NullTextFromString(input.Carrier.VehicleType),
			RouteDetails:               repository.NullTextFromString(input.Carrier.RouteDetails),
			DriverName:                 repository.NullTextFromString(input.Carrier.DriverName),
			DriverContact:              repository.NullTextFromString(input.Carrier.DriverContact),
			OriginPortStation:          repository.NullTextFromString(input.Carrier.OriginPortStation),
			DestinationPortStation:     repository.NullTextFromString(input.Carrier.DestinationPortStation),
			OriginCountry:              repository.NullTextFromString(input.Carrier.OriginCountry),
			DestinationCountry:         repository.NullTextFromString(input.Carrier.DestinationCountry),
			AccountingInfo:             repository.NullTextFromString(input.Carrier.AccountingInfo),
			HandlingInfo:               repository.NullTextFromString(input.Carrier.HandlingInfo),
			TransportDocumentReference: repository.NullTextFromString(input.Carrier.TransportDocumentRef),
			DocUrls:                    input.Carrier.SupportingDocURLs,
			FileRegion:                 repository.NullTextFromString(input.Carrier.FileRegion),
			Description:                repository.NullTextFromString(input.Carrier.Description),
			Actor:                      pgtype.Text{String: input.ModifiedBy, Valid: true},
		}
		// We need the carrier ID to update. Since we don't have it in input, we assume single carrier per job.
		// But UpdateJobCarrier requires carrier_id.
		// Let's fetch existing carrier first.
		carriers, _ := s.repo.GetJobCarriers(ctx, jobID)
		if len(carriers) > 0 {
			updateParams.CarrierID = carriers[0].ID
			_, err = s.repo.UpdateJobCarrier(ctx, updateParams)
			if err != nil {
				logger.Warn("failed to update carrier", slog.Any("error", err))
			}
		} else {
			// Create new carrier
			createParams := sqlc.CreateJobCarrierParams{
				JobID:                      uuidToPgtype(jobID),
				CarrierPartyID:             repository.NullUUIDFromUUID(input.Carrier.CarrierPartyID),
				CarrierName:                repository.NullTextFromString(input.Carrier.CarrierName),
				CarrierContact:             repository.NullTextFromString(input.Carrier.CarrierContact),
				VesselName:                 repository.NullTextFromString(input.Carrier.VesselName),
				VoyageNumber:               repository.NullTextFromString(input.Carrier.VoyageNumber),
				FlightID:                   repository.NullTextFromString(input.Carrier.FlightID),
				FlightDate:                 timestampFromTime(input.Carrier.FlightDate),
				AirportReportDate:          timestampFromTime(input.Carrier.AirportReportDate),
				VehicleNumber:              repository.NullTextFromString(input.Carrier.VehicleNumber),
				VehicleType:                repository.NullTextFromString(input.Carrier.VehicleType),
				RouteDetails:               repository.NullTextFromString(input.Carrier.RouteDetails),
				DriverName:                 repository.NullTextFromString(input.Carrier.DriverName),
				DriverContact:              repository.NullTextFromString(input.Carrier.DriverContact),
				OriginPortStation:          repository.NullTextFromString(input.Carrier.OriginPortStation),
				DestinationPortStation:     repository.NullTextFromString(input.Carrier.DestinationPortStation),
				OriginCountry:              repository.NullTextFromString(input.Carrier.OriginCountry),
				DestinationCountry:         repository.NullTextFromString(input.Carrier.DestinationCountry),
				AccountingInfo:             repository.NullTextFromString(input.Carrier.AccountingInfo),
				HandlingInfo:               repository.NullTextFromString(input.Carrier.HandlingInfo),
				TransportDocumentReference: repository.NullTextFromString(input.Carrier.TransportDocumentRef),
				DocUrls:                    input.Carrier.SupportingDocURLs,
				FileRegion:                 repository.NullTextFromString(input.Carrier.FileRegion),
				Description:                repository.NullTextFromString(input.Carrier.Description),
				Actor:                      pgtype.Text{String: input.ModifiedBy, Valid: true},
			}
			_, err = s.repo.CreateJobCarrier(ctx, createParams)
			if err != nil {
				logger.Warn("failed to create carrier", slog.Any("error", err))
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
			IssuedDate:  timestampFromTime(docInput.IssuedDate),
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
			JobID:                 uuidToPgtype(jobID),
			ActivityType:          repository.NullTextFromString(billInput.ActivityType),
			ActivityCode:          repository.NullTextFromString(billInput.ActivityCode),
			BillingPartyID:        repository.NullUUIDFromUUID(billInput.BillingPartyID),
			PoNumber:              repository.NullTextFromString(billInput.PONumber),
			PoDate:                timestampFromTime(billInput.PODate),
			CurrencyCode:          repository.NullTextFromString(billInput.CurrencyCode),
			Quantity:              numericFromFloat64(billInput.Quantity),
			UnitPrice:             numericFromFloat64(billInput.UnitPrice),
			AmountWithoutTax:      numericFromFloat64(billInput.AmountWithoutTax),
			TaxCode:               repository.NullTextFromString(billInput.TaxCode),
			TaxAmount:             numericFromFloat64(billInput.TaxAmount),
			ExchangeRate:          numericFromFloat64(billInput.ExchangeRate),
			TotalAmount:           numericFromFloat64(billInput.TotalAmount),
			Description:           repository.NullTextFromString(billInput.Description),
			Notes:                 repository.NullTextFromString(billInput.Notes),
			DocUrls:               billInput.SupportingDocURLs,
			FileRegion:            repository.NullTextFromString(billInput.FileRegion),
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
			JobID:                 uuidToPgtype(jobID),
			ActivityType:          repository.NullTextFromString(provInput.ActivityType),
			ActivityCode:          repository.NullTextFromString(provInput.ActivityCode),
			CostPartyID:           repository.NullUUIDFromUUID(provInput.CostPartyID),
			InvoiceNumber:         repository.NullTextFromString(provInput.InvoiceNumber),
			InvoiceDate:           timestampFromTime(provInput.InvoiceDate),
			CurrencyCode:          repository.NullTextFromString(provInput.CurrencyCode),
			Quantity:              numericFromFloat64(provInput.Quantity),
			UnitPrice:             numericFromFloat64(provInput.UnitPrice),
			AmountWithoutTax:      numericFromFloat64(provInput.AmountWithoutTax),
			TaxCode:               repository.NullTextFromString(provInput.TaxCode),
			TaxAmount:             numericFromFloat64(provInput.TaxAmount),
			TotalAmount:           numericFromFloat64(provInput.TotalAmount),
			PoNumber:              repository.NullTextFromString(provInput.PONumber),
			PoDate:                timestampFromTime(provInput.PODate),
			ExchangeRate:          numericFromFloat64(provInput.ExchangeRate),
			PaymentPriority:       repository.NullTextFromString(provInput.PaymentPriority),
			Notes:                 repository.NullTextFromString(provInput.Notes),
			DocUrls:               provInput.SupportingDocURLs,
			FileRegion:            repository.NullTextFromString(provInput.FileRegion),
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
			JobID:          uuidToPgtype(jobID),
			EtdDate:        timestampFromTime(input.Tracking.ETDDate),
			EtaDate:        timestampFromTime(input.Tracking.ETADate),
			AtdDate:        timestampFromTime(input.Tracking.ATDDate),
			AtaDate:        timestampFromTime(input.Tracking.ATADate),
			JobStatus:      repository.NullTextFromString(input.Tracking.JobStatus),
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
// CONVERSION FUNCTIONS - SQLC to DTO
// ============================================================

func packageFromSqlc(pkg sqlc.OpsPackage) operationsdto.Package {
	return operationsdto.Package{
		ID:                        pkg.ID,
		ContainerNo:               common.PgtypeTextToStringPtr(pkg.ContainerNo),
		ContainerType:             common.PgtypeTextToStringPtr(pkg.ContainerType),
		ContainerSize:             common.PgtypeTextToStringPtr(pkg.ContainerSize),
		GrossWeightKg:             float64FromNumeric(pkg.GrossWeightKg),
		NetWeightKg:               float64FromNumeric(pkg.NetWeightKg),
		Volume:                    float64FromNumeric(pkg.Volume),
		CarrierSealNo:             common.PgtypeTextToStringPtr(pkg.CarrierSealNo),
		CommodityCargoDescription: common.PgtypeTextToStringPtr(pkg.CommodityCargoDescription),
		PackageType:               common.PgtypeTextToStringPtr(pkg.PackageType),
		CargoType:                 common.PgtypeTextToStringPtr(pkg.CargoType),
		NoOfPackages:              float64FromNumeric(pkg.NoOfPackages),
		ChargeableWeight:          float64FromNumeric(pkg.ChargeableWeight),
		HSCode:                    common.PgtypeTextToStringPtr(pkg.HsCode),
		TemperatureControl:        boolFromPgtype(pkg.TemperatureControl),
	}
}

func carrierFromSqlc(carrier sqlc.OpsCarrier) operationsdto.Carrier {
	var flightDate *time.Time
	if carrier.FlightDate.Valid {
		flightDate = &carrier.FlightDate.Time
	}
	var airportReportDate *time.Time
	if carrier.AirportReportDate.Valid {
		airportReportDate = &carrier.AirportReportDate.Time
	}
	return operationsdto.Carrier{
		ID:                     carrier.ID,
		CarrierPartyID:         uuidFromPgtype(carrier.CarrierPartyID),
		CarrierName:            common.PgtypeTextToStringPtr(carrier.CarrierName),
		CarrierContact:         common.PgtypeTextToStringPtr(carrier.CarrierContact),
		VesselName:             common.PgtypeTextToStringPtr(carrier.VesselName),
		VoyageNumber:           common.PgtypeTextToStringPtr(carrier.VoyageNumber),
		FlightID:               common.PgtypeTextToStringPtr(carrier.FlightID),
		FlightDate:             flightDate,
		AirportReportDate:      airportReportDate,
		VehicleNumber:          common.PgtypeTextToStringPtr(carrier.VehicleNumber),
		VehicleType:            common.PgtypeTextToStringPtr(carrier.VehicleType),
		RouteDetails:           common.PgtypeTextToStringPtr(carrier.RouteDetails),
		DriverName:             common.PgtypeTextToStringPtr(carrier.DriverName),
		DriverContact:          common.PgtypeTextToStringPtr(carrier.DriverContact),
		OriginPortStation:      common.PgtypeTextToStringPtr(carrier.OriginPortStation),
		DestinationPortStation: common.PgtypeTextToStringPtr(carrier.DestinationPortStation),
		OriginCountry:          common.PgtypeTextToStringPtr(carrier.OriginCountry),
		DestinationCountry:     common.PgtypeTextToStringPtr(carrier.DestinationCountry),
		AccountingInfo:         common.PgtypeTextToStringPtr(carrier.AccountingInfo),
		HandlingInfo:           common.PgtypeTextToStringPtr(carrier.HandlingInfo),
		TransportDocumentRef:   common.PgtypeTextToStringPtr(carrier.TransportDocumentReference),
		SupportingDocURLs:      stringsFromStringArray(carrier.SupportingDocUrl),
		FileRegion:             common.PgtypeTextToStringPtr(carrier.FileRegion),
		Description:            common.PgtypeTextToStringPtr(carrier.Description),
	}
}

func documentFromSqlc(doc sqlc.OpsJobDocument) operationsdto.Document {
	return operationsdto.Document{
		ID:          doc.ID,
		DocTypeCode: common.PgtypeTextToStringPtr(doc.DocTypeCode),
		DocNumber:   common.PgtypeTextToStringPtr(doc.DocNumber),
		IssuedAt:    common.PgtypeTextToStringPtr(doc.IssuedAt),
		IssuedDate:  timeFromTimestamptz(doc.IssuedDate),
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
		Quantity:              float64FromNumeric(b.Quantity),
		UnitPrice:             float64FromNumeric(b.UnitPrice),
		AmountWithoutTax:      float64FromNumeric(b.AmountWithoutTax),
		TaxCode:               common.PgtypeTextToStringPtr(b.TaxCode),
		TaxAmount:             float64FromNumeric(b.TaxAmount),
		ExchangeRate:          float64FromNumeric(b.ExchangeRate),
		TotalAmount:           float64FromNumeric(b.TotalAmount),
		Description:           common.PgtypeTextToStringPtr(b.Description),
		Notes:                 common.PgtypeTextToStringPtr(b.Notes),
		SupportingDocURLs:     stringsFromStringArray(b.SupportingDocUrl),
		FileRegion:            common.PgtypeTextToStringPtr(b.FileRegion),
		AmountPrimaryCurrency: float64FromNumeric(b.AmountPrimaryCurrency),
	}
}

func provisionFromSqlc(p sqlc.ListJobProvisionsRow) operationsdto.Provision {
	var invDate *time.Time
	if p.InvoiceDate.Valid {
		invDate = &p.InvoiceDate.Time
	}
	var poDate *time.Time
	if p.PoDate.Valid {
		poDate = &p.PoDate.Time
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
		Quantity:              float64FromNumeric(p.Quantity),
		UnitPrice:             float64FromNumeric(p.UnitPrice),
		AmountWithoutTax:      float64FromNumeric(p.AmountWithoutTax),
		TaxCode:               common.PgtypeTextToStringPtr(p.TaxCode),
		TaxAmount:             float64FromNumeric(p.TaxAmount),
		TotalAmount:           float64FromNumeric(p.TotalAmount),
		PONumber:              common.PgtypeTextToStringPtr(p.PoNumber),
		PODate:                poDate,
		ExchangeRate:          float64FromNumeric(p.ExchangeRate),
		PaymentPriority:       common.PgtypeTextToStringPtr(p.PaymentPriority),
		Notes:                 common.PgtypeTextToStringPtr(p.Notes),
		SupportingDocURLs:     stringsFromStringArray(p.SupportingDocUrl),
		FileRegion:            common.PgtypeTextToStringPtr(p.FileRegion),
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
		PODDocURLs:     stringsFromStringArray(t.PodDocUrls),
		FileRegion:     common.PgtypeTextToStringPtr(t.FileRegion),
		DocumentStatus: common.PgtypeTextToStringPtr(t.DocumentStatus),
		Notes:          common.PgtypeTextToStringPtr(t.Notes),
	}
}
