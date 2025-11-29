package operations

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"frego-operations/internal/db"
	sqlc "frego-operations/internal/db/sqlc"
)

// Repository wraps sqlc queries for operations domain.
type Repository struct {
	tenantSessions *db.TenantSessionManager
}

func New(pool *pgxpool.Pool) *Repository {
	return NewWithSessions(db.NewTenantSessionManager(pool, pool, "operations"))
}

func NewWithSessions(sessions *db.TenantSessionManager) *Repository {
	return &Repository{
		tenantSessions: sessions,
	}
}

func (r *Repository) withQueries(ctx context.Context, fn func(*sqlc.Queries) error) error {
	return r.tenantSessions.WithTenantTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := sqlc.New(tx)
		return fn(q)
	})
}

// ============================================================
// LOOKUP METHODS
// ============================================================

func (r *Repository) ListTransportModeServiceLookups(ctx context.Context) ([]sqlc.ListTransportModeServiceLookupsRow, error) {
	var rows []sqlc.ListTransportModeServiceLookupsRow
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		rows, err = q.ListTransportModeServiceLookups(ctx)
		return err
	})
	return rows, err
}

func (r *Repository) ListJobStatusLookups(ctx context.Context) ([]sqlc.ListJobStatusLookupsRow, error) {
	var rows []sqlc.ListJobStatusLookupsRow
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		rows, err = q.ListJobStatusLookups(ctx)
		return err
	})
	return rows, err
}

func (r *Repository) ListDocumentStatusLookups(ctx context.Context) ([]sqlc.ListDocumentStatusLookupsRow, error) {
	var rows []sqlc.ListDocumentStatusLookupsRow
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		rows, err = q.ListDocumentStatusLookups(ctx)
		return err
	})
	return rows, err
}

func (r *Repository) ListPriorityLookups(ctx context.Context) ([]sqlc.ListPriorityLookupsRow, error) {
	var rows []sqlc.ListPriorityLookupsRow
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		rows, err = q.ListPriorityLookups(ctx)
		return err
	})
	return rows, err
}

func (r *Repository) ListRoleDetailsLookups(ctx context.Context) ([]sqlc.ListRoleDetailsLookupsRow, error) {
	var rows []sqlc.ListRoleDetailsLookupsRow
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		rows, err = q.ListRoleDetailsLookups(ctx)
		return err
	})
	return rows, err
}

// ============================================================
// JOB CRUD METHODS
// ============================================================

func (r *Repository) GetNextJobSequence(ctx context.Context, prefix string) (int32, error) {
	var seq int32
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		seq, err = q.GetNextJobSequence(ctx, pgtype.Text{String: prefix, Valid: true})
		return err
	})
	return seq, err
}

func (r *Repository) ListJobs(ctx context.Context, status, jobType *string, customerID *uuid.UUID, limit int32) ([]sqlc.ListJobsRow, error) {
	var rows []sqlc.ListJobsRow
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		rows, err = q.ListJobs(ctx, sqlc.ListJobsParams{
			Status:     NullTextFromString(status),
			CustomerID: NullUUIDFromUUID(customerID),
			JobType:    NullTextFromString(jobType),
			RowLimit:   limit,
		})
		return err
	})
	return rows, err
}

func (r *Repository) GetJob(ctx context.Context, id uuid.UUID) (sqlc.GetJobRow, error) {
	var row sqlc.GetJobRow
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		row, err = q.GetJob(ctx, id)
		return err
	})
	return row, err
}

func (r *Repository) CreateJob(ctx context.Context, params sqlc.CreateJobParams) (sqlc.OpsJob, error) {
	var job sqlc.OpsJob
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		job, err = q.CreateJob(ctx, params)
		return err
	})
	return job, err
}

func (r *Repository) UpdateJob(ctx context.Context, params sqlc.UpdateJobParams) (sqlc.OpsJob, error) {
	var job sqlc.OpsJob
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		job, err = q.UpdateJob(ctx, params)
		return err
	})
	return job, err
}

func (r *Repository) ArchiveJob(ctx context.Context, id uuid.UUID, actor string) error {
	return r.withQueries(ctx, func(q *sqlc.Queries) error {
		return q.ArchiveJob(ctx, sqlc.ArchiveJobParams{
			ID:    id,
			Actor: pgtype.Text{String: actor, Valid: true},
		})
	})
}

// ============================================================
// JOB PACKAGE METHODS
// ============================================================

func (r *Repository) ListJobPackages(ctx context.Context, jobID uuid.UUID) ([]sqlc.OpsPackage, error) {
	var rows []sqlc.OpsPackage
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		rows, err = q.ListJobPackages(ctx, jobID)
		return err
	})
	return rows, err
}

func (r *Repository) CreateJobPackage(ctx context.Context, params sqlc.CreateJobPackageParams) (sqlc.OpsPackage, error) {
	var pkg sqlc.OpsPackage
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		pkg, err = q.CreateJobPackage(ctx, params)
		return err
	})
	return pkg, err
}

// ============================================================
// JOB CARRIER METHODS
// ============================================================

func (r *Repository) GetJobCarriers(ctx context.Context, jobID uuid.UUID) ([]sqlc.OpsCarrier, error) {
	var carriers []sqlc.OpsCarrier
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		var bytes [16]byte
		copy(bytes[:], jobID[:])
		carriers, err = q.GetJobCarriers(ctx, pgtype.UUID{Bytes: bytes, Valid: true})
		return err
	})
	return carriers, err
}

func (r *Repository) CreateJobCarrier(ctx context.Context, params sqlc.CreateJobCarrierParams) (sqlc.OpsCarrier, error) {
	var carrier sqlc.OpsCarrier
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		carrier, err = q.CreateJobCarrier(ctx, params)
		return err
	})
	return carrier, err
}

func (r *Repository) UpdateJobCarrier(ctx context.Context, params sqlc.UpdateJobCarrierParams) (sqlc.OpsCarrier, error) {
	var carrier sqlc.OpsCarrier
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		carrier, err = q.UpdateJobCarrier(ctx, params)
		return err
	})
	return carrier, err
}

// ============================================================
// JOB DOCUMENT METHODS
// ============================================================

func (r *Repository) ListJobDocuments(ctx context.Context, jobID uuid.UUID) ([]sqlc.OpsJobDocument, error) {
	var rows []sqlc.OpsJobDocument
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		rows, err = q.ListJobDocuments(ctx, jobID)
		return err
	})
	return rows, err
}

func (r *Repository) CreateJobDocument(ctx context.Context, params sqlc.CreateJobDocumentParams) (sqlc.OpsJobDocument, error) {
	var doc sqlc.OpsJobDocument
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		doc, err = q.CreateJobDocument(ctx, params)
		return err
	})
	return doc, err
}

// ============================================================
// JOB BILLING METHODS
// ============================================================

func (r *Repository) ListJobBilling(ctx context.Context, jobID uuid.UUID) ([]sqlc.ListJobBillingRow, error) {
	var rows []sqlc.ListJobBillingRow
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		var bytes [16]byte
		copy(bytes[:], jobID[:])
		rows, err = q.ListJobBilling(ctx, pgtype.UUID{Bytes: bytes, Valid: true})
		return err
	})
	return rows, err
}

func (r *Repository) CreateJobBilling(ctx context.Context, params sqlc.CreateJobBillingParams) (sqlc.OpsBilling, error) {
	var billing sqlc.OpsBilling
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		billing, err = q.CreateJobBilling(ctx, params)
		return err
	})
	return billing, err
}

// ============================================================
// JOB PROVISION METHODS
// ============================================================

func (r *Repository) ListJobProvisions(ctx context.Context, jobID uuid.UUID) ([]sqlc.ListJobProvisionsRow, error) {
	var rows []sqlc.ListJobProvisionsRow
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		var bytes [16]byte
		copy(bytes[:], jobID[:])
		rows, err = q.ListJobProvisions(ctx, pgtype.UUID{Bytes: bytes, Valid: true})
		return err
	})
	return rows, err
}

func (r *Repository) CreateJobProvision(ctx context.Context, params sqlc.CreateJobProvisionParams) (sqlc.OpsProvision, error) {
	var provision sqlc.OpsProvision
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		provision, err = q.CreateJobProvision(ctx, params)
		return err
	})
	return provision, err
}

// ============================================================
// JOB TRACKING METHODS
// ============================================================

func (r *Repository) GetJobTracking(ctx context.Context, jobID uuid.UUID) (sqlc.OpsTracking, error) {
	var tracking sqlc.OpsTracking
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		var bytes [16]byte
		copy(bytes[:], jobID[:])
		tracking, err = q.GetJobTracking(ctx, pgtype.UUID{Bytes: bytes, Valid: true})
		return err
	})
	return tracking, err
}

func (r *Repository) UpsertJobTracking(ctx context.Context, params sqlc.UpsertJobTrackingParams) (sqlc.OpsTracking, error) {
	var tracking sqlc.OpsTracking
	err := r.withQueries(ctx, func(q *sqlc.Queries) error {
		var err error
		tracking, err = q.UpsertJobTracking(ctx, params)
		return err
	})
	return tracking, err
}

func NullTextFromString(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func NullUUIDFromUUID(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{Valid: false}
	}
	var bytes [16]byte
	copy(bytes[:], u[:])
	return pgtype.UUID{Bytes: bytes, Valid: true}
}
