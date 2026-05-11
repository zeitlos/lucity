package directory

import (
	"context"

	"github.com/zeitlos/lucity/pkg/auth"
)

type Provider interface {
	Workspaces(ctx context.Context) ([]Workspace, error)
	WorkspacesForUser(ctx context.Context, userID string) ([]Workspace, error)
	Workspace(ctx context.Context, id string) (*WorkspaceDetails, error)

	CreateWorkspace(ctx context.Context, id, name string) (*WorkspaceDetails, error)
	UpdateWorkspace(ctx context.Context, id, name string) (*WorkspaceDetails, error)
	DeleteWorkspace(ctx context.Context, id string) error

	InviteMember(ctx context.Context, workspaceID, email string, role Role) (*WorkspaceMember, error)
	UpdateMemberRole(ctx context.Context, workspaceID, userID string, role Role) (*WorkspaceMember, error)
	RemoveMember(ctx context.Context, workspaceID, userID string) error

	Projects(ctx, tenantID string) ([]Project, error)
	ProjectsForUser(ctx context.Context, userID string) ([]Workspace, error)
	Project(ctx, id string) (*Project, error)
}

type Role = auth.WorkspaceRole

type Workspace struct {
	ID        string
	Name      string
	Personal  bool
	Suspended bool
}

type WorkspaceDetails struct {
	Workspace

	Members []WorkspaceMember
}

type WorkspaceMember struct {
	ID    string
	Email string
	Name  string
	Role  Role
}

type Project struct {
	ID   string
	Name string
}
