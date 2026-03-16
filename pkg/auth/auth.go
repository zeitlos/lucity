package auth

import (
	"context"
)

// Role represents a user's authorization level.
type Role string

const (
	RoleAnonymous Role = "ANONYMOUS"
	RoleUser      Role = "USER"
	RoleAdmin     Role = "ADMIN"
)

// WorkspaceRole represents a user's role within a workspace.
type WorkspaceRole string

const (
	WorkspaceRoleUser  WorkspaceRole = "user"
	WorkspaceRoleAdmin WorkspaceRole = "admin"
)

// WorkspaceMembership represents a user's membership in a workspace.
type WorkspaceMembership struct {
	Workspace string
	Role      WorkspaceRole
}

type contextKey struct{}

// Claims represents the authenticated user's identity and roles.
type Claims struct {
	Subject    string // OIDC subject (stable user identifier)
	Name       string // Display name
	Email      string
	Roles      []Role
	AvatarURL  string
	Workspaces []WorkspaceMembership
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

// IsMemberOf checks if the user is a member of the given workspace.
func (c *Claims) IsMemberOf(workspace string) bool {
	for _, m := range c.Workspaces {
		if m.Workspace == workspace {
			return true
		}
	}
	return false
}

// WorkspaceRoleIn returns the user's role in a workspace, or empty string if not a member.
func (c *Claims) WorkspaceRoleIn(workspace string) WorkspaceRole {
	for _, m := range c.Workspaces {
		if m.Workspace == workspace {
			return m.Role
		}
	}
	return ""
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
