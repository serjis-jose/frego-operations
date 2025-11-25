package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"frego-operations/internal/common"
)

var errMissingToken = errors.New("auth: bearer token missing")

// Middleware enforces Keycloak authentication across the HTTP surface.
func Middleware(logger *slog.Logger, authenticator *Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			normalizedPath := strings.TrimSuffix(r.URL.Path, "/")
			// Bypass authentication for health check and swagger endpoints
			if r.Method == http.MethodGet && (strings.HasSuffix(normalizedPath, "/health") || strings.HasSuffix(normalizedPath, "/swagger") || strings.HasSuffix(normalizedPath, "/swagger.yaml")) {
				next.ServeHTTP(w, r)
				return
			}

			rawToken, err := bearerToken(r.Header.Get("Authorization"))
			if err != nil {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}

			principal, err := authenticator.Authenticate(r.Context(), rawToken)
			if err != nil {
				switch {
				case errors.Is(err, ErrTenantMissing):
					http.Error(w, "tenant context missing in token", http.StatusForbidden)
				case errors.Is(err, ErrAudienceMismatch):
					http.Error(w, "token audience not accepted", http.StatusForbidden)
				default:
					logger.Warn("auth: token verification failed", slog.Any("error", err))
					http.Error(w, "invalid bearer token", http.StatusUnauthorized)
				}
				return
			}

			ctx := common.WithPrincipal(r.Context(), principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(header string) (string, error) {
	if header == "" {
		return "", errMissingToken
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		return "", errMissingToken
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return "", errMissingToken
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errMissingToken
	}
	return token, nil
}
