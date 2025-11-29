package operations

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// LOOKUP DTOs
// ============================================================

type ServiceSubcategory struct {
	ServiceSubcategoryID   int16
	ServiceSubcategoryName string
}

type ServiceType struct {
	ServiceTypeID        int16
	ServiceTypeName      string
	ServiceSubcategories []ServiceSubcategory
}

type MovementType struct {
	MovementTypeID   int16
	MovementTypeName string
	ServiceTypes     []ServiceType
}

type TransportMode struct {
	TransportModeID   int16
	TransportModeName string
	MovementTypes     []MovementType
}

// JobStatus represents a job status lookup
type JobStatus struct {
	JobStatusID   int16
	JobStatusName string
	JobStatusDesc *string
}

// DocumentStatus represents a document status lookup
type DocumentStatus struct {
	DocStatusID   int16
	DocStatusName string
	DocStatusDesc *string
}

// PriorityLevel represents a priority level lookup
type PriorityLevel struct {
	PriorityID    int16
	PriorityLabel string
}

// RoleDetails represents a role details lookup
type RoleDetails struct {
	RoleID   int16
	RoleName string
	RoleDesc *string
}

// SalesExecutiveLookup represents a sales executive lookup entry grouped by branch.
type SalesExecutiveLookup struct {
	SalesExecID   uuid.UUID
	SalesExecName string
	BranchID      *uuid.UUID
	BranchName    *string
}

// CSExecutiveLookup represents a CS executive lookup entry grouped by branch.
type CSExecutiveLookup struct {
	CSExecID   uuid.UUID
	CSExecName string
	BranchID   *uuid.UUID
	BranchName *string
}

// BranchLookup represents a branch lookup entry.
type BranchLookup struct {
	BranchID   uuid.UUID
	BranchName string
	IsActive   bool
}

// IncotermLookup represents an available incoterm option.
type IncotermLookup struct {
	ID      uuid.UUID
	Code    string
	Name    string
	Version int32
}

// OperationsLookups aggregates all operations-related lookup data
type OperationsLookups struct {
	TransportModes       []string
	MovementTypes        map[string][]string
	ServiceTypes         map[string][]string
	ServiceSubcategories map[string][]string
	Incoterms            []IncotermLookup
	JobStatuses          []JobStatus
	DocumentStatuses     []DocumentStatus
	PriorityLevels       []PriorityLevel
	RoleDetails          []RoleDetails
	SalesExecutives      []SalesExecutiveLookup
	CSExecutives         []CSExecutiveLookup
	Branches             []BranchLookup
}

// ============================================================
// JOB DTOs
// ============================================================

// Employee reference for jobs
type Employee struct {
	ID    uuid.UUID
	Name  string
	Email string
	Role  string
}

// JobListItem represents a job in list view
type JobListItem struct {
	ID                  uuid.UUID
	JobCode             string
	EnquiryNumber       *string
	JobType             *string
	TransportMode       *string
	ServiceType         *string
	CustomerID          *uuid.UUID
	CustomerName        *string
	AgentID             *uuid.UUID
	AgentName           *string
	ShipmentOrigin      *string
	DestinationCity     *string
	DestinationState    *string
	DestinationCountry  *string
	SourceCity          *string
	SourceState         *string
	SourceCountry       *string
	Status              *string
	PriorityLevel       *string
	SalesExecutive      Employee
	OperationsExecutive Employee
	CreatedAt           time.Time
	ModifiedAt          *time.Time
	IsActive            bool
}

// JobDetail represents a complete job with all related data
type JobDetail struct {
	ID                  uuid.UUID
	JobCode             string
	EnquiryNumber       *string
	JobType             *string
	TransportMode       *string
	ServiceType         *string
	ServiceSubcategory  *string
	ParentJobID         *uuid.UUID
	CustomerID          *uuid.UUID
	CustomerName        *string
	AgentID             *uuid.UUID
	AgentName           *string
	ShipmentOrigin      *string
	DestinationCity     *string
	DestinationState    *string
	DestinationCountry  *string
	SourceCity          *string
	SourceState         *string
	SourceCountry       *string
	BranchID            *uuid.UUID
	BranchName          *string
	IncotermCode        *string
	Commodity           *string
	Classification      *string
	SalesExecutive      Employee
	OperationsExecutive Employee
	CSExecutive         Employee
	AgentDeadline       *time.Time
	ShipmentReadyDate   *time.Time
	Status              *string
	PriorityLevel       *string
	CreatedAt           time.Time
	CreatedBy           *string
	ModifiedAt          *time.Time
	ModifiedBy          *string
	IsActive            bool
	Packages            []Package
	Carrier             *Carrier
	Documents           []Document
	Billing             []Billing
	Provisions          []Provision
	Tracking            *Tracking
}

// Package represents a job package
type Package struct {
	ID                        uuid.UUID
	ContainerNo               *string
	ContainerName             *string
	ContainerSize             *string
	GrossWeightKg             *float64
	NetWeightKg               *float64
	Volume                    *float64
	CarrierSealNo             *string
	CommodityCargoDescription *string
	PackageType               *string
	CargoType                 *string
	NoOfPackages              *float64
	ChargeableWeight          *float64
	HSCode                    *string
	TemperatureControl        bool
}

// Carrier represents job carrier information
type Carrier struct {
	ID                     uuid.UUID
	CarrierPartyID         *uuid.UUID
	CarrierName            *string
	CarrierContact         *string
	VesselName             *string
	VoyageNumber           *string
	FlightID               *string
	FlightDate             *time.Time
	AirportReportDate      *time.Time
	VehicleNumber          *string
	VehicleType            *string
	RouteDetails           *string
	DriverName             *string
	DriverContact          *string
	OriginPortStation      *string
	DestinationPortStation *string
	OriginCountry          *string
	DestinationCountry     *string
	AccountingInfo         *string
	HandlingInfo           *string
	TransportDocumentRef   *string
	SupportingDocURLs      []string
	FileRegion             *string
	Description            *string
}

// Document represents a job document
type Document struct {
	ID          uuid.UUID
	DocTypeCode *string
	DocNumber   *string
	IssuedAt    *string
	IssuedDate  *time.Time
	Description *string
	FileKey     *string
	FileRegion  *string
}

// Billing represents job billing information
type Billing struct {
	ID                    uuid.UUID
	ActivityType          *string
	ActivityCode          *string
	BillingPartyID        *uuid.UUID
	BillingPartyName      *string
	PONumber              *string
	PODate                *time.Time
	CurrencyCode          *string
	Amount                *float64
	Description           *string
	Notes                 *string
	SupportingDocURLs     []string
	FileRegion            *string
	AmountPrimaryCurrency *float64
}

// Provision represents job provision/cost information
type Provision struct {
	ID                    uuid.UUID
	ActivityType          *string
	ActivityCode          *string
	CostPartyID           *uuid.UUID
	CostPartyName         *string
	InvoiceNumber         *string
	InvoiceDate           *time.Time
	CurrencyCode          *string
	Amount                *float64
	PaymentPriority       *string
	Notes                 *string
	SupportingDocURLs     []string
	FileRegion            *string
	AmountPrimaryCurrency *float64
	Profit                *float64
}

// Tracking represents job tracking information
type Tracking struct {
	ID             uuid.UUID
	ETDDate        *time.Time
	ETADate        *time.Time
	ATDDate        *time.Time
	ATADate        *time.Time
	JobStatus      *string
	PODDocURLs     []string
	FileRegion     *string
	DocumentStatus *string
	Notes          *string
}

// ============================================================
// INPUT DTOs
// ============================================================

// CreateJobInput represents input for creating a job
type CreateJobInput struct {
	JobCode            *string // Optional - will be auto-generated if not provided
	EnquiryNumber      *string
	JobType            *string
	TransportMode      *string
	ServiceType        *string
	ServiceSubcategory *string
	ParentJobID        *uuid.UUID
	CustomerID         *uuid.UUID
	AgentID            *uuid.UUID
	ShipmentOrigin     *string
	DestinationCity    *string
	DestinationState   *string
	DestinationCountry *string
	SourceCity         *string
	SourceState        *string
	SourceCountry      *string
	BranchID           *uuid.UUID
	BranchName         *string
	IncotermCode       *string
	Commodity          *string
	Classification     *string
	SalesExecutiveID   *uuid.UUID
	SalesExecutiveName *string
	OperationsExecID   *uuid.UUID
	OperationsExecName *string
	CSExecutiveID      *uuid.UUID
	CSExecutiveName    *string
	AgentDeadline      *time.Time
	ShipmentReadyDate  *time.Time
	Status             *string
	PriorityLevel      *string
	CreatedBy          string
	Packages           []PackageInput
	Carrier            *CarrierInput
	Documents          []DocumentInput
	Billing            []BillingInput
	Provisions         []ProvisionInput
	Tracking           *TrackingInput
}

// UpdateJobInput represents input for updating a job
type UpdateJobInput struct {
	EnquiryNumber      *string
	JobType            *string
	TransportMode      *string
	ServiceType        *string
	ServiceSubcategory *string
	ParentJobID        *uuid.UUID
	CustomerID         *uuid.UUID
	AgentID            *uuid.UUID
	ShipmentOrigin     *string
	DestinationCity    *string
	DestinationState   *string
	DestinationCountry *string
	SourceCity         *string
	SourceState        *string
	SourceCountry      *string
	BranchID           *uuid.UUID
	BranchName         *string
	IncotermCode       *string
	Commodity          *string
	Classification     *string
	SalesExecutiveID   *uuid.UUID
	SalesExecutiveName *string
	OperationsExecID   *uuid.UUID
	OperationsExecName *string
	CSExecutiveID      *uuid.UUID
	CSExecutiveName    *string
	AgentDeadline      *time.Time
	ShipmentReadyDate  *time.Time
	Status             *string
	PriorityLevel      *string
	ModifiedBy         string
	Packages           []PackageInput
	Carrier            *CarrierInput
	Documents          []DocumentInput
	Billing            []BillingInput
	Provisions         []ProvisionInput
	Tracking           *TrackingInput
}

// PackageInput represents input for creating a package
type PackageInput struct {
	PackageName  *string
	PackageType  *string
	Quantity     *int32
	LengthMeters *float64
	WidthMeters  *float64
	HeightMeters *float64
	WeightKg     *float64
	VolumeCbm    *float64
	HSCode       *string
	CargoType    *string
	ContainerID  *string
	Notes        *string
}

// CarrierInput represents input for creating carrier information
type CarrierInput struct {
	CarrierPartyID         *uuid.UUID
	CarrierName            *string
	VesselName             *string
	VoyageNumber           *string
	FlightID               *string
	FlightDate             *time.Time
	VehicleNumber          *string
	RouteDetails           *string
	DriverName             *string
	OriginPortStation      *string
	DestinationPortStation *string
	AccountingInfo         *string
	HandlingInfo           *string
	SupportingDocURL       *string
}

// DocumentInput represents input for creating a document
type DocumentInput struct {
	DocTypeCode *string
	DocNumber   *string
	IssuedAt    *string
	IssuedDate  *time.Time
	Description *string
	FileKey     *string
	FileRegion  *string
}

// BillingInput represents input for creating billing information
type BillingInput struct {
	ActivityType          *string
	ActivityCode          *string
	BillingPartyID        *uuid.UUID
	PONumber              *string
	PODate                *time.Time
	CurrencyCode          *string
	Amount                *float64
	Description           *string
	Notes                 *string
	SupportingDocURL      *string
	AmountPrimaryCurrency *float64
}

// ProvisionInput represents input for creating provision information
type ProvisionInput struct {
	ActivityType          *string
	ActivityCode          *string
	CostPartyID           *uuid.UUID
	InvoiceNumber         *string
	InvoiceDate           *time.Time
	CurrencyCode          *string
	Amount                *float64
	PaymentPriority       *string
	Notes                 *string
	SupportingDocURL      *string
	AmountPrimaryCurrency *float64
	Profit                *float64
}

// TrackingInput represents input for creating/updating tracking information
type TrackingInput struct {
	ETDDate        *time.Time
	ETADate        *time.Time
	ATDDate        *time.Time
	ATADate        *time.Time
	JobStatus      *string
	PODStatus      *string
	DocumentStatus *string
	Notes          *string
}
