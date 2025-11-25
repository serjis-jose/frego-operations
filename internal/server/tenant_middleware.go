package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tenantIDHeader = "X-Tenant-ID"

// TenantMiddleware extracts tenant ID from request and sets it in context
func TenantMiddleware(logger *slog.Logger, pool *pgxpool.Pool, defaultTenant uuid.UUID) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantIDStr := r.Header.Get(tenantIDHeader)

			var tenantID uuid.UUID
			var err error

			if tenantIDStr != "" {
				tenantID, err = uuid.Parse(tenantIDStr)
				if err != nil {
					logger.Warn("invalid tenant ID in header, using default",
						slog.String("tenant_id", tenantIDStr),
						slog.Any("error", err),
					)
					tenantID = defaultTenant
				}
			} else {
				tenantID = defaultTenant
			}

			// Verify tenant exists and is active
			var exists bool
			err = pool.QueryRow(r.Context(), `
				SELECT EXISTS (
					SELECT 1 FROM registry.tenant_registry 
					WHERE tenant_id = $1 AND is_active = true
				)
			`, tenantID).Scan(&exists)

			if err != nil && err != pgx.ErrNoRows {
				logger.Error("failed to verify tenant", slog.Any("error", err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if !exists {
				logger.Warn("tenant not found or inactive", slog.String("tenant_id", tenantID.String()))
				http.Error(w, "Tenant not found or inactive", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), "tenant_id", tenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetTenantID extracts tenant ID from context
func GetTenantID(ctx context.Context) (uuid.UUID, error) {
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("tenant ID not found in context")
	}
	return tenantID, nil
}
