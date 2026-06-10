package conductor

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

const grpcTimeout = 10 * time.Second

// projectIDPattern validates project slugs (same rules as workspace IDs).
var projectIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`)

type Project = platform.Project
type ProjectID = platform.ProjectID
type Deployment = platform.Deployment

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

	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}

	if len(s) > 63 {
		s = s[:63]
	}

	return s
}

func (c *Client) CreateProject(ctx context.Context, ws, slug, displayName string) (*Project, error) {
	if slug == "" {
		slug = slugFromName(displayName)
	}

	if !projectIDPattern.MatchString(slug) {
		return nil, fmt.Errorf("invalid project ID %q: must be 3-63 lowercase alphanumeric characters or hyphens", slug)
	}

	projectID := platform.ProjectID{
		Workspace: ws,
		Name:      slug,
	}

	if _, err := c.CreateEnvironment(ctx, projectID, "development", nil, EcoTier); err != nil {
		return nil, fmt.Errorf("create default environment: %w", err)
	}

	return c.Project(ctx, projectID)
}

func (c *Client) DeleteProject(ctx context.Context, project platform.ProjectID) error {
	envs, err := c.platform.Environments(ctx, project)

	if err != nil {
		return err
	}

	for _, env := range envs {
		if err := c.checkEnvironmentEmpty(ctx, env.ID); err != nil {
			return err
		}
	}

	for _, env := range envs {
		if err := c.environment.Delete(ctx, env.ID); err != nil {
			return err
		}
	}

	return nil
}
