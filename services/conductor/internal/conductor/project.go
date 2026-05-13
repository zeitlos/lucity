package conductor

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/pkg/tenant"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

// gRPC call timeouts. Short for quick lookups, long for operations that
// clone repos or touch external systems.
const (
	grpcTimeout     = 10 * time.Second
	grpcLongTimeout = 60 * time.Second
)

// projectIDPattern validates project slugs (same rules as workspace IDs).
var projectIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`)

// TODO: Rename to Project once refactoring is done.
type Project = platform.Project
type ProjectID = platform.ProjectID

type ServiceInstance struct {
	ID                   platform.ServiceID
	Name                 string
	Image                string
	Port                 int
	Framework            string
	SourceURL            string
	ContextPath          string
	GitHubInstallationID int64
	StartCommand         string
	CustomStartCommand   string
	ImageTag             string
	Ready                bool
	Replicas             int
	Scaling              ScalingConfig
	Resources            *ServiceResources
	Domains              []string
	Deployments          []Deployment
	InitialDeploy        *DeployOp
}

type ServiceResources struct {
	CpuMillicores      int
	MemoryMB           int
	CpuLimitMillicores int
	MemoryLimitMB      int
}

type ScalingConfig struct {
	Replicas    int
	Autoscaling *AutoscalingConfig
}

type AutoscalingConfig struct {
	Enabled     bool
	MinReplicas int
	MaxReplicas int
	TargetCPU   int
}

type Deployment struct {
	ID                  platform.DeploymentID
	ImageTag            string
	Active              bool
	Timestamp           time.Time
	Revision            string
	Message             string
	SourceCommitMessage string
	SourceURL           string // full URL to commit on GitHub
}

func (c *Client) Projects(ctx context.Context, workspace string) ([]Project, error) {
	return c.platform.Projects(ctx, workspace)
}

func (c *Client) Project(ctx context.Context, id platform.ProjectID) (*Project, error) {
	return c.platform.Project(ctx, id)
}

// slugFromName derives a URL-safe slug from a display name.
func slugFromName(name string) string {
	s := strings.ToLower(name)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	s = strings.Trim(b.String(), "-")
	// Collapse consecutive hyphens
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	if len(s) > 63 {
		s = s[:63]
	}
	return s
}

func (c *Client) CreateProject(ctx context.Context, ws, slug, displayName string) (*Project, error) {
	// Derive slug from display name if not provided
	if slug == "" {
		slug = slugFromName(displayName)
	}
	if !projectIDPattern.MatchString(slug) {
		return nil, fmt.Errorf("invalid project ID %q: must be 3-63 lowercase alphanumeric characters or hyphens", slug)
	}

	// 1. Create GitOps repo
	initCtx, initCancel := context.WithTimeout(ctx, grpcLongTimeout)
	defer initCancel()
	repoURL, err := c.Packager.InitProject(initCtx, ws, slug, displayName)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// 2. Deploy the default development environment via ArgoCD
	envName := "development"
	ns := labels.NamespaceFor(ws, slug, envName)
	deployCtx, deployCancel := context.WithTimeout(ctx, grpcTimeout)
	defer deployCancel()
	if _, err := c.Deployer.DeployEnvironment(deployCtx, ws, slug, envName, repoURL, ns); err != nil {
		slog.Warn("failed to deploy development environment", "project", slug, "error", err)
	}

	projectID, err := platform.ParseProjectID(ws + "/" + slug)

	if err != nil {
		return nil, err
	}

	// envID, err := platform.ParseEnvironmentID(ws + "/" + slug + "/" + envName)

	// if err != nil {
	// 	return nil, err
	// }

	_ = ws
	return c.Project(ctx, projectID)
}

func (c *Client) DeleteProject(ctx context.Context, project platform.ProjectID) (bool, error) {
	id := project.Name
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return false, err
	}

	// 1. Fetch project to discover all environments
	getCtx, getCancel := context.WithTimeout(ctx, grpcTimeout)
	defer getCancel()
	info, err := c.Packager.GetProject(getCtx, ws, id)
	if err != nil {
		return false, fmt.Errorf("failed to get project for deletion: %w", err)
	}

	// 2. Remove ArgoCD Application for each environment (best-effort)
	for _, env := range info.Environments {
		rmCtx, rmCancel := context.WithTimeout(ctx, grpcTimeout)
		if err := c.Deployer.RemoveDeployment(rmCtx, ws, id, env); err != nil {
			slog.Warn("failed to remove deployment during project deletion",
				"project", id, "environment", env, "error", err)
		}
		rmCancel()
	}

	// 3. Remove ArgoCD repository credential (best-effort)
	repoCtx, repoCancel := context.WithTimeout(ctx, grpcTimeout)
	defer repoCancel()
	if err := c.Deployer.DeleteRepository(repoCtx, ws, id); err != nil {
		slog.Warn("failed to delete ArgoCD repository credential",
			"project", id, "error", err)
	}

	// 4. Delete OCI images from registry (best-effort)
	imgCtx, imgCancel := context.WithTimeout(ctx, grpcTimeout)
	defer imgCancel()
	if _, err := c.Builder.DeleteImages(imgCtx, ws, id); err != nil {
		slog.Warn("failed to delete registry images",
			"project", id, "error", err)
	}

	// 5. Delete GitOps repo
	delCtx, delCancel := context.WithTimeout(ctx, grpcTimeout)
	defer delCancel()
	if err := c.Packager.DeleteProject(delCtx, ws, id); err != nil {
		return false, fmt.Errorf("failed to delete project: %w", err)
	}
	return true, nil
}
