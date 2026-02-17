package auth

import "context"

// Role represents a user's authorization level.
type Role string

const (
	RoleAnonymous Role = "ANONYMOUS"
	RoleUser      Role = "USER"
	RoleAdmin     Role = "ADMIN"
)

type contextKey struct{}

// Claims represents the authenticated user's identity and roles.
type Claims struct {
	Subject     string
	Email       string
	Roles       []Role
	GitHubLogin string
	AvatarURL   string
}

// HasRole checks if the claims include the given role.
func (c *Claims) HasRole(role Role) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// WithClaims attaches claims to a context.
func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, contextKey{}, claims)
}

// FromContext extracts claims from a context.
// Returns nil if no claims are present.
func FromContext(ctx context.Context) *Claims {
	claims, _ := ctx.Value(contextKey{}).(*Claims)
	return claims
}
