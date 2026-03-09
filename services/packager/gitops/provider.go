package gitops

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeitlos/lucity/pkg/labels"
)

// RepoSuffix is appended to project names to form GitOps repo names.
const RepoSuffix = "-gitops"

// ServiceDef describes a service configured in the project's GitOps repo.
type ServiceDef struct {
	Name        string
	Image       string // image repository path (e.g., localhost:5000/myapp/api)
	Port        int
	Framework   string // detected framework for dashboard icons (e.g., "nextjs", "vite")
	SourceURL   string // GitHub repo URL for this service
	ContextPath string // subdirectory within the repo (monorepo support)
}

// DatabaseDef describes a PostgreSQL database configured in the project's GitOps repo.
type DatabaseDef struct {
	Name      string
	Version   string // e.g., "16"
	Instances int    // CNPG cluster instances
	Size      string // e.g., "10Gi"
}

// DatabaseRef references a key in a CNPG-generated Kubernetes Secret.
type DatabaseRef struct {
	Database string
	Key      string
}

// ServiceRef references another service's internal URL.
type ServiceRef struct {
	Service string
}

// Provider abstracts Git repository operations for GitOps repos.
// Implementation: Soft-serve.
type Provider interface {
	// CreateRepo creates a GitOps repo with the standard directory structure
	// and an initial commit. Returns the repo clone URL.
	CreateRepo(ctx context.Context, project string) (repoURL string, err error)

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
	// commitPrefix controls the git commit message prefix (e.g., "deploy", "rollback").
	// If empty, defaults to "deploy".
	UpdateImageTag(ctx context.Context, project, environment, service, tag, digest, commitPrefix string) error

	// Services reads the services defined in the project's base/values.yaml.
	Services(ctx context.Context, project string) ([]ServiceDef, error)

	// CreateEnvironment creates a new environment directory with values.yaml
	// in the GitOps repo. If fromEnvironment is set, copies its values as a starting point,
	// strips all domains, and regenerates platform domains using workloadDomain.
	// Returns the names of services that received new platform domains.
	CreateEnvironment(ctx context.Context, project, environment, fromEnvironment, workloadDomain string) (serviceNames []string, err error)

	// DeleteEnvironment removes an environment directory from the GitOps repo.
	DeleteEnvironment(ctx context.Context, project, environment string) error

	// Promote copies the image tag for a service from one environment to another.
	// Returns the promoted image tag.
	Promote(ctx context.Context, project, service, fromEnv, toEnv string) (imageTag string, err error)

	// DeploymentHistory returns the deployment history for a service in an environment,
	// parsed from the GitOps repo's git log. Returns entries in reverse chronological order.
	DeploymentHistory(ctx context.Context, project, environment, service string) ([]DeploymentEntry, error)

	// AddDomain adds a domain hostname to a service in an environment.
	AddDomain(ctx context.Context, project, environment, service, hostname string) error

	// RemoveDomain removes a domain hostname from a service in an environment.
	RemoveDomain(ctx context.Context, project, environment, service, hostname string) error

	// AllDomains returns all domain hostnames across all projects and environments.
	AllDomains(ctx context.Context) ([]string, error)

	// EnvironmentServices reads per-environment service state (image tags, domains)
	// from the environment's values.yaml.
	EnvironmentServices(ctx context.Context, project, environment string) ([]ServiceInstanceMeta, error)

	// RepoFiles returns raw file contents from the GitOps repo, keyed by relative path.
	// Excludes the chart/ directory (the embedded chart is used instead during ejection).
	RepoFiles(ctx context.Context, project string) (map[string][]byte, error)

	// SharedVariables returns all shared variables for an environment.
	SharedVariables(ctx context.Context, project, environment string) (map[string]string, error)

	// SetSharedVariables replaces all shared variables for an environment.
	// Also propagates value changes to any services that reference shared vars via sharedRefs.
	SetSharedVariables(ctx context.Context, project, environment string, vars map[string]string) error

	// ServiceVariables returns all variables and shared refs for a service in an environment.
	ServiceVariables(ctx context.Context, project, environment, service string) (vars map[string]string, sharedRefs []string, databaseRefs map[string]DatabaseRef, serviceRefs map[string]ServiceRef, err error)

	// SetServiceVariables replaces all variables for a service in an environment.
	// Direct values come from vars. Keys listed in sharedRefs are resolved from the
	// environment's shared variables and merged into the service's env map.
	SetServiceVariables(ctx context.Context, project, environment, service string, vars map[string]string, sharedRefs []string, databaseRefs map[string]DatabaseRef, serviceRefs map[string]ServiceRef) error

	// AddDatabase adds a PostgreSQL database definition to base/values.yaml.
	AddDatabase(ctx context.Context, project string, db DatabaseDef) error

	// RemoveDatabase removes a database definition from base/values.yaml.
	RemoveDatabase(ctx context.Context, project, name string) error

	// Databases reads the database definitions from base/values.yaml.
	Databases(ctx context.Context, project string) ([]DatabaseDef, error)

	// SyncChart overwrites the embedded lucity-app chart in the project's GitOps repo
	// with the current version. No-op if the chart is already up to date.
	SyncChart(ctx context.Context, project string) error
}

// DeploymentEntry represents a single deployment event parsed from a git commit.
type DeploymentEntry struct {
	ImageTag  string
	Revision  string // git commit SHA
	Timestamp time.Time
	Author    string
}

// maxDeploymentHistory is the maximum number of deployment history entries to return.
const maxDeploymentHistory = 20

// parseDeployCommit checks if a commit message represents a deployment of the given
// service to the given environment. Returns the image tag if matched.
func parseDeployCommit(message, environment, service string) (imageTag string, ok bool) {
	// Match: deploy(<env>): <service> <tag>
	// Match: rollback(<env>): <service> <tag>
	for _, prefix := range []string{"deploy", "rollback"} {
		full := fmt.Sprintf("%s(%s): %s ", prefix, environment, service)
		if strings.HasPrefix(message, full) {
			tag := strings.TrimSpace(message[len(full):])
			if tag != "" {
				return tag, true
			}
		}
	}

	// Match: promote(<env>): <service> ...
	promotePrefix := fmt.Sprintf("promote(%s): %s ", environment, service)
	if strings.HasPrefix(message, promotePrefix) {
		// Softserve format: promote(<toEnv>): <service> <fromEnv> <toEnv>
		// The tag isn't in the message — mark as promoted.
		rest := strings.TrimSpace(message[len(promotePrefix):])
		parts := strings.Fields(rest)
		if len(parts) >= 1 {
			return fmt.Sprintf("promoted from %s", parts[0]), true
		}
		return "promoted", true
	}

	return "", false
}

// ServiceInstanceMeta describes a service's state in a specific environment.
type ServiceInstanceMeta struct {
	Name     string
	ImageTag string
	Domains  []string // domain hostnames from per-environment values.yaml
}

// EnvironmentMeta holds metadata about a project environment.
type EnvironmentMeta struct {
	Name     string
	Services []ServiceInstanceMeta
}

// ProjectMeta holds metadata about a project, read from its GitOps repo.
type ProjectMeta struct {
	Name             string // e.g., "warm-wren"
	RepoURL          string
	Environments     []string
	EnvironmentInfos []EnvironmentMeta
	Services         []ServiceDef
	Databases        []DatabaseDef
	CreatedAt        time.Time
}

// NamespaceFor derives the K8s namespace from workspace, project, and environment.
// "acme" + "api" + "production" → "acme-api-production"
func NamespaceFor(workspace, project, environment string) string {
	return labels.NamespaceFor(workspace, project, environment)
}
