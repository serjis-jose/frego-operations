package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Server represents the HTTP server
type Server struct {
	logger        *slog.Logger
	address       string
	server        *http.Server
	gracefulDelay time.Duration
}

// New creates a new HTTP server
func New(logger *slog.Logger, address string, handler http.Handler, gracefulDelay time.Duration) *Server {
	return &Server{
		logger:        logger,
		address:       address,
		gracefulDelay: gracefulDelay,
		server: &http.Server{
			Addr:              address,
			Handler:           handler,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       120 * time.Second,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("starting http server", slog.String("address", s.address))

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down http server", slog.Duration("graceful_delay", s.gracefulDelay))

	time.Sleep(s.gracefulDelay)

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	s.logger.Info("http server shutdown complete")
	return nil
}

// BuildRouter builds the main HTTP router
func BuildRouter(
	logger *slog.Logger,
	apiHandler http.Handler,
	tenantHandler http.Handler,
	corsMiddleware func(http.Handler) http.Handler,
	authMiddleware func(http.Handler) http.Handler,
	tenantMiddleware func(http.Handler) http.Handler,
) http.Handler {
	r := chi.NewRouter()

	// Global middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	// Health check endpoint (no auth required)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Internal tenant provisioning routes (no auth required - for backend-to-finance calls)
	if tenantHandler != nil {
		r.Mount("/", tenantHandler)
	}

	// API routes with auth and tenant middleware
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Use(tenantMiddleware)
		r.Mount("/finance/api/v1", apiHandler)
	})

	return r
}
