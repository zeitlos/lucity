package auth

import (
	"context"
	"strings"
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
	Subject    string
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

// ParseRauthyGroups converts Rauthy group names into workspace memberships.
// Group format:
//   - "ws:acme"         → workspace "acme", role "user"
//   - "ws:acme:admin"   → workspace "acme", role "admin"
//   - Groups without the "ws:" prefix are ignored
func ParseRauthyGroups(groups []string) []WorkspaceMembership {
	seen := make(map[string]int) // workspace → index in memberships
	var memberships []WorkspaceMembership

	for _, g := range groups {
		if !strings.HasPrefix(g, "ws:") {
			continue
		}
		parts := strings.SplitN(g, ":", 3)
		if len(parts) < 2 || parts[1] == "" {
			continue
		}
		workspace := parts[1]
		role := WorkspaceRoleUser
		if len(parts) == 3 && parts[2] == "admin" {
			role = WorkspaceRoleAdmin
		}

		if idx, ok := seen[workspace]; ok {
			// Upgrade to admin if we see the admin group
			if role == WorkspaceRoleAdmin {
				memberships[idx].Role = WorkspaceRoleAdmin
			}
			continue
		}

		seen[workspace] = len(memberships)
		memberships = append(memberships, WorkspaceMembership{
			Workspace: workspace,
			Role:      role,
		})
	}
	return memberships
}
