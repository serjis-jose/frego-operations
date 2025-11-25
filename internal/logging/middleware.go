package logging

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// InjectMiddleware attaches the provided logger to the request context, optionally
// enriching it with middleware-provided metadata like the request ID. Downstream
// handlers can retrieve the logger via FromContext.
func InjectMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			enriched := logger
			if reqID := middleware.GetReqID(ctx); reqID != "" {
				enriched = enriched.With(slog.String("request_id", reqID))
			}
			next.ServeHTTP(w, r.WithContext(WithContext(ctx, enriched)))
		})
	}
}
