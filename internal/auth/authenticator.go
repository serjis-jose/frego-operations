package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"

	"frego-operations/internal/common"
)

// Config drives how the authenticator validates Keycloak-issued tokens.
type Config struct {
	Issuer      string
	Audiences   []string
	TenantClaim string
}

// ErrAudienceMismatch indicates the token was not intended for this API.
var ErrAudienceMismatch = errors.New("auth: token audience mismatch")

// ErrTenantMissing indicates the expected tenant claim was not present.
var ErrTenantMissing = errors.New("auth: tenant claim missing")

// Authenticator wraps verification of OIDC tokens and projection into principals.
type Authenticator struct {
	verifier    *oidc.IDTokenVerifier
	audiences   []string
	tenantClaim string
}

// NewAuthenticator builds an Authenticator backed by a cached OIDC provider.
func NewAuthenticator(ctx context.Context, cfg Config) (*Authenticator, error) {
	if strings.TrimSpace(cfg.Issuer) == "" {
		return nil, errors.New("auth: keycloak issuer must be configured")
	}
	if strings.TrimSpace(cfg.TenantClaim) == "" {
		cfg.TenantClaim = "tenantId"
	}

	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("auth: init oidc provider: %w", err)
	}

	oidcCfg := &oidc.Config{
		SkipClientIDCheck: true,
	}
	verifier := provider.Verifier(oidcCfg)

	return &Authenticator{
		verifier:    verifier,
		audiences:   normalizeAudiences(cfg.Audiences),
		tenantClaim: cfg.TenantClaim,
	}, nil
}

// Authenticate verifies the provided raw token, returning the associated principal.
func (a *Authenticator) Authenticate(ctx context.Context, rawToken string) (common.Principal, error) {
	idToken, err := a.verifier.Verify(ctx, rawToken)
	if err != nil {
		return common.Principal{}, fmt.Errorf("auth: verify token: %w", err)
	}

	if len(a.audiences) > 0 && !audAllowed(idToken.Audience, a.audiences) {
		return common.Principal{}, fmt.Errorf("%w: %v", ErrAudienceMismatch, idToken.Audience)
	}

	var claims map[string]any
	if err := idToken.Claims(&claims); err != nil {
		return common.Principal{}, fmt.Errorf("auth: decode claims: %w", err)
	}

	tenantID, err := extractTenantClaim(claims, a.tenantClaim)
	if err != nil {
		return common.Principal{}, err
	}

	principal := common.Principal{
		Username: stringOrFallback(claims, "preferred_username", idToken.Subject),
		Subject:  idToken.Subject,
		TenantID: tenantID,
		Claims:   claims,
	}

	if email, ok := stringClaim(claims["email"]); ok {
		principal.Email = email
	}
	if scope, ok := stringClaim(claims["scope"]); ok {
		principal.Scopes = fields(scope)
	}
	principal.Roles = extractRoles(claims)

	return principal, nil
}

func stringOrFallback(claims map[string]any, key, fallback string) string {
	if val, ok := stringClaim(claims[key]); ok && val != "" {
		return val
	}
	return fallback
}

func normalizeAudiences(auds []string) []string {
	var normalized []string
	for _, aud := range auds {
		trimmed := strings.TrimSpace(aud)
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}
	return normalized
}

func audAllowed(tokenAud, allowed []string) bool {
	for _, allowedAud := range allowed {
		for _, aud := range tokenAud {
			if aud == allowedAud {
				return true
			}
		}
	}
	return false
}

func extractTenantClaim(claims map[string]any, tenantClaim string) (string, error) {
	raw, ok := claims[tenantClaim]
	if !ok {
		return "", fmt.Errorf("%w: claim %q missing", ErrTenantMissing, tenantClaim)
	}
	if s, ok := stringClaim(raw); ok && s != "" {
		return s, nil
	}
	return "", fmt.Errorf("%w: claim %q malformed", ErrTenantMissing, tenantClaim)
}

func stringClaim(input any) (string, bool) {
	switch v := input.(type) {
	case string:
		return v, true
	case []string:
		if len(v) == 0 {
			return "", false
		}
		return v[0], true
	case fmt.Stringer:
		return v.String(), true
	case []any:
		if len(v) == 0 {
			return "", false
		}
		for _, candidate := range v {
			if s, ok := stringClaim(candidate); ok && s != "" {
				return s, true
			}
		}
	}
	return "", false
}

func extractRoles(claims map[string]any) []string {
	var roles []string
	if realmAccess, ok := nestedMap(claims, "realm_access"); ok {
		if realmRoles, ok := realmAccess["roles"]; ok {
			roles = append(roles, toStringSlice(realmRoles)...)
		}
	}
	if resourceAccess, ok := nestedMap(claims, "resource_access"); ok {
		for _, res := range resourceAccess {
			switch section := res.(type) {
			case map[string]any:
				if sectionRoles, ok := section["roles"]; ok {
					roles = append(roles, toStringSlice(sectionRoles)...)
				}
			}
		}
	}
	return dedupe(roles)
}

func nestedMap(claims map[string]any, key string) (map[string]any, bool) {
	raw, ok := claims[key]
	if !ok {
		return nil, false
	}
	m, ok := raw.(map[string]any)
	return m, ok
}

func toStringSlice(val any) []string {
	switch v := val.(type) {
	case []string:
		return append([]string(nil), v...)
	case []any:
		var out []string
		for _, item := range v {
			if s, ok := stringClaim(item); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		if s, ok := stringClaim(v); ok && s != "" {
			return []string{s}
		}
	}
	return nil
}

func dedupe(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	var result []string
	for _, v := range values {
		if v == "" {
			continue
		}
		if _, exists := seen[v]; exists {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}
	return result
}

func fields(scope string) []string {
	parts := strings.Fields(scope)
	if len(parts) == 0 {
		return nil
	}
	return parts
}
