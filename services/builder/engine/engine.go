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
	ImageRef string // full image reference with tag (e.g., localhost:5000/proj/svc:abc123)
	Digest   string // image digest (e.g., sha256:...)
}

// BuildOpts configures a build.
type BuildOpts struct {
	BuildID     string // unique build identifier (used for K8s Job naming/labels)
	ContextPath string // subdirectory within repo, empty = root
	SourceURL   string // source repository URL (e.g., https://github.com/user/repo)
	GitRef      string // git ref to clone (branch, tag, SHA)
	GitHubToken string // GitHub OAuth token for cloning
	Registry    string // base image path without tag (e.g., host:5000/proj/svc)
	Insecure    bool   // allow HTTP (non-TLS) registry connections
}

// Engine abstracts the build backend.
type Engine interface {
	// Detect scans source code at the given path and returns detected services.
	Detect(ctx context.Context, repoPath string) ([]DetectResult, error)

	// Build builds a container image from source and pushes to the registry.
	Build(ctx context.Context, opts BuildOpts) (*BuildResult, error)
}
