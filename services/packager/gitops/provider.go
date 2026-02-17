package gitops

import (
	"context"
	"time"
)

// ServiceDef describes a service configured in the project's GitOps repo.
type ServiceDef struct {
	Name      string
	Image     string // image repository path (e.g., ghcr.io/user/myapp/api)
	Port      int
	Public    bool
	Framework string // detected framework for dashboard icons (e.g., "nextjs", "vite")
}

// Provider abstracts Git repository operations for GitOps repos.
// Implementations: GitHub (default), Soft-serve (future).
type Provider interface {
	// CreateRepo creates a GitOps repo with the standard directory structure
	// and an initial commit. Returns the repo clone URL.
	CreateRepo(ctx context.Context, project, sourceURL string) (repoURL string, err error)

	// Repos lists all project GitOps repos and their metadata.
	Repos(ctx context.Context) ([]ProjectMeta, error)

	// Repo reads a single project's metadata from its GitOps repo.
	Repo(ctx context.Context, project string) (*ProjectMeta, error)

	// DeleteRepo removes a project's GitOps repo.
	DeleteRepo(ctx context.Context, project string) error

	// AddService adds a service definition to the project's base/values.yaml.
	AddService(ctx context.Context, project string, svc ServiceDef) error

	// RemoveService removes a service definition from the project's base/values.yaml.
	RemoveService(ctx context.Context, project, service string) error

	// UpdateImageTag updates the image tag for a service in an environment's values.yaml.
	UpdateImageTag(ctx context.Context, project, environment, service, tag, digest string) error

	// Services reads the services defined in the project's base/values.yaml.
	Services(ctx context.Context, project string) ([]ServiceDef, error)
}

// ProjectMeta holds metadata about a project, read from its GitOps repo.
type ProjectMeta struct {
	Name         string    // org-scoped: "zeitlos/myapp"
	SourceURL    string
	RepoURL      string
	Environments []string
	Services     []ServiceDef
	CreatedAt    time.Time
}
