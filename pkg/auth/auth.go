package auth

import (
	"context"
	"errors"
)

// WorkspaceRole represents a user's role within a workspace.
type WorkspaceRole string

const (
	WorkspaceRoleDeployer WorkspaceRole = "deployer"
	WorkspaceRoleUser     WorkspaceRole = "user"
	WorkspaceRoleAdmin    WorkspaceRole = "admin"
)

var roleRank = map[WorkspaceRole]int{
	WorkspaceRoleDeployer: 1,
	WorkspaceRoleUser:     2,
	WorkspaceRoleAdmin:    3,
}

func (r WorkspaceRole) Satisfies(required WorkspaceRole) bool {
	return roleRank[r] >= roleRank[required]
}

// WorkspaceMembership represents a user's membership in a workspace.
type WorkspaceMembership struct {
	Workspace string
	Role      WorkspaceRole
}

type contextKey struct{}

// Claims represents the authenticated user's identity and roles.
type Claims struct {
	Subject    string
	Name       string
	Email      string
	AvatarURL  string
	Workspaces []WorkspaceMembership
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

func NewContext(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, contextKey{}, claims)
}

func FromContext(ctx context.Context) (*Claims, error) {
	claims, set := ctx.Value(contextKey{}).(*Claims)

	if !set || claims == nil {
		return nil, errors.New("unauthenticated")
	}

	return claims, nil
}
