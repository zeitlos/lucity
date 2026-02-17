package engine

import "context"

// DetectResult holds what was detected from a source directory.
type DetectResult struct {
	Name         string // suggested service name (e.g., "web")
	Provider     string // language provider (e.g., "node", "go", "python")
	Framework    string // detected framework (e.g., "nextjs", "vite", "django")
	StartCommand string // detected start command
	SuggestedPort int   // port based on framework heuristics
}

// BuildResult holds the output of a successful build.
type BuildResult struct {
	ImageRef string // full image reference with tag (e.g., ghcr.io/user/proj/svc:abc123)
	Digest   string // image digest (e.g., sha256:...)
}

// BuildOpts configures a build.
type BuildOpts struct {
	RepoPath    string // cloned source directory
	ImageName   string // full registry path with tag
	ContextPath string // subdirectory within repo, empty = root
	Token       string // OAuth token for registry push auth
}

// Engine abstracts the build backend.
// Implementations: LocalEngine (Docker + railpack), future: KubernetesEngine, DaggerEngine, GitHubActionsEngine.
type Engine interface {
	// Detect scans source code at the given path and returns detected services.
	Detect(ctx context.Context, repoPath string) ([]DetectResult, error)

	// Build builds a container image from source and pushes to the registry.
	Build(ctx context.Context, opts BuildOpts) (*BuildResult, error)
}
