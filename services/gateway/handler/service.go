package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/builder"
	"github.com/zeitlos/lucity/pkg/packager"
)

type DetectedService struct {
	Name          string
	Provider      string
	Framework     string
	StartCommand  string
	SuggestedPort int
}

type Build struct {
	ID       string
	Phase    string
	ImageRef string
	Digest   string
	Error    string
}

func (c *Client) DetectServices(ctx context.Context, projectID string) ([]DetectedService, error) {
	ctx = auth.OutgoingContext(ctx)

	// Get source URL from packager
	resp, err := c.Packager.GetProject(ctx, &packager.GetProjectRequest{Project: projectID})
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Call builder to detect services
	detectResp, err := c.Builder.DetectServices(ctx, &builder.DetectServicesRequest{
		SourceUrl: resp.Project.SourceUrl,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to detect services: %w", err)
	}

	result := make([]DetectedService, 0, len(detectResp.Services))
	for _, s := range detectResp.Services {
		result = append(result, DetectedService{
			Name:          s.Name,
			Provider:      s.Provider,
			Framework:     s.Framework,
			StartCommand:  s.StartCommand,
			SuggestedPort: int(s.SuggestedPort),
		})
	}
	return result, nil
}

func (c *Client) AddService(ctx context.Context, projectID, name string, port int, public bool, framework string) (*Service, error) {
	ctx = auth.OutgoingContext(ctx)

	// Get project owner from the project name (format: "owner/repo-gitops")
	resp, err := c.Packager.GetProject(ctx, &packager.GetProjectRequest{Project: projectID})
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Derive image path from source URL: github.com/owner/repo → ghcr.io/owner/repo/service
	image := deriveImagePath(c.RegistryURL, resp.Project.SourceUrl, name)

	_, err = c.Packager.AddService(ctx, &packager.AddServiceRequest{
		Project:   projectID,
		Service:   name,
		Image:     image,
		Port:      int32(port),
		Public:    public,
		Framework: framework,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add service: %w", err)
	}

	return &Service{
		Name:      name,
		Image:     image,
		Port:      port,
		Public:    public,
		Framework: framework,
	}, nil
}

func (c *Client) RemoveService(ctx context.Context, projectID, service string) (bool, error) {
	ctx = auth.OutgoingContext(ctx)

	_, err := c.Packager.RemoveService(ctx, &packager.RemoveServiceRequest{
		Project: projectID,
		Service: service,
	})
	if err != nil {
		return false, fmt.Errorf("failed to remove service: %w", err)
	}
	return true, nil
}

func (c *Client) StartBuild(ctx context.Context, projectID, service, gitRef, contextPath string) (*Build, error) {
	ctx = auth.OutgoingContext(ctx)

	// Get source URL and image path from packager
	resp, err := c.Packager.GetProject(ctx, &packager.GetProjectRequest{Project: projectID})
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	registry := deriveImagePath(c.RegistryURL, resp.Project.SourceUrl, service)

	buildResp, err := c.Builder.StartBuild(ctx, &builder.StartBuildRequest{
		SourceUrl:   resp.Project.SourceUrl,
		GitRef:      gitRef,
		Service:     service,
		Registry:    registry,
		ContextPath: contextPath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start build: %w", err)
	}

	return &Build{
		ID:    buildResp.BuildId,
		Phase: "QUEUED",
	}, nil
}

func (c *Client) BuildStatus(ctx context.Context, buildID string) (*Build, error) {
	ctx = auth.OutgoingContext(ctx)

	resp, err := c.Builder.BuildStatus(ctx, &builder.BuildStatusRequest{BuildId: buildID})
	if err != nil {
		return nil, fmt.Errorf("failed to get build status: %w", err)
	}

	return &Build{
		ID:       buildID,
		Phase:    buildPhaseToString(resp.Phase),
		ImageRef: resp.ImageRef,
		Digest:   resp.Digest,
		Error:    resp.Error,
	}, nil
}

func (c *Client) DeployBuild(ctx context.Context, projectID, service, environment, tag, digest string) (bool, error) {
	ctx = auth.OutgoingContext(ctx)

	_, err := c.Packager.UpdateImageTag(ctx, &packager.UpdateImageTagRequest{
		Project:     projectID,
		Environment: environment,
		Service:     service,
		Tag:         tag,
		Digest:      digest,
	})
	if err != nil {
		return false, fmt.Errorf("failed to deploy build: %w", err)
	}
	return true, nil
}

// deriveImagePath builds a registry image path from a source URL.
// "https://github.com/owner/repo" + service "web" → "ghcr.io/owner/repo/web"
func deriveImagePath(registryURL, sourceURL, service string) string {
	// Strip protocol and github.com prefix
	path := sourceURL
	path = strings.TrimPrefix(path, "https://")
	path = strings.TrimPrefix(path, "http://")
	path = strings.TrimPrefix(path, "github.com/")
	path = strings.TrimSuffix(path, ".git")

	return registryURL + "/" + path + "/" + service
}

func buildPhaseToString(phase builder.BuildPhase) string {
	switch phase {
	case builder.BuildPhase_BUILD_PHASE_QUEUED:
		return "QUEUED"
	case builder.BuildPhase_BUILD_PHASE_CLONING:
		return "CLONING"
	case builder.BuildPhase_BUILD_PHASE_BUILDING:
		return "BUILDING"
	case builder.BuildPhase_BUILD_PHASE_PUSHING:
		return "PUSHING"
	case builder.BuildPhase_BUILD_PHASE_SUCCEEDED:
		return "SUCCEEDED"
	case builder.BuildPhase_BUILD_PHASE_FAILED:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}
