package common

import "context"

type principalKey struct{}

// Principal represents the authenticated identity extracted from upstream auth.
type Principal struct {
	Username string
	Subject  string
	Email    string
	TenantID string
	Roles    []string
	Scopes   []string
	Claims   map[string]any
}

// HasRole reports whether the principal is assigned the specified role.
func (p Principal) HasRole(role string) bool {
	for _, r := range p.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// WithPrincipal stores the authenticated principal on the supplied context.
func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, principal)
}

// PrincipalFromContext retrieves the principal from context if present.
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	val := ctx.Value(principalKey{})
	if val == nil {
		return Principal{}, false
	}
	principal, ok := val.(Principal)
	if !ok {
		return Principal{}, false
	}
	return principal, true
}
