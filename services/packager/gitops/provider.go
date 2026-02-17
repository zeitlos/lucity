package gitops

import (
	"context"
	"time"
)

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
}

// ProjectMeta holds metadata about a project, read from its GitOps repo.
type ProjectMeta struct {
	Name         string    // org-scoped: "zeitlos/myapp"
	SourceURL    string
	RepoURL      string
	Environments []string
	CreatedAt    time.Time
}
