package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/builder"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/packager"
	"github.com/zeitlos/lucity/services/gateway/deploy"
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

	// Derive image path using cluster-internal address so pods can pull it.
	image := deriveImagePath(c.RegistryImagePrefix, projectID, name)

	_, err := c.Packager.AddService(ctx, &packager.AddServiceRequest{
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

	registry := deriveImagePath(c.RegistryPushURL, projectID, service)

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

	// Trigger ArgoCD sync (best-effort — auto-sync will pick it up anyway)
	_, err = c.Deployer.SyncDeployment(ctx, &deployer.SyncDeploymentRequest{
		Project:     projectID,
		Environment: environment,
	})
	if err != nil {
		slog.Warn("failed to trigger sync", "project", projectID, "environment", environment, "error", err)
	}

	return true, nil
}

// deriveImagePath builds a registry image path from a project name.
// Project "zeitlos/myapp" + service "web" → "localhost:5000/myapp/web"
// The org prefix is stripped — OCI paths use only the short project name.
func deriveImagePath(registryURL, project, service string) string {
	// Use only the short name (after the slash) for the image namespace
	parts := strings.SplitN(project, "/", 2)
	name := project
	if len(parts) == 2 {
		name = parts[1]
	}
	return registryURL + "/" + name + "/" + service
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

// DeployOp represents the state of a unified build+deploy operation.
type DeployOp struct {
	ID          string
	Phase       string
	BuildID     string
	ImageRef    string
	Digest      string
	Error       string
	ArgoHealth  string
	ArgoMessage string
}

func deployOpFromState(s *deploy.State) *DeployOp {
	return &DeployOp{
		ID:          s.ID,
		Phase:       string(s.Phase),
		BuildID:     s.BuildID,
		ImageRef:    s.ImageRef,
		Digest:      s.Digest,
		Error:       s.Error,
		ArgoHealth:  s.ArgoHealth,
		ArgoMessage: s.ArgoMessage,
	}
}

// Deploy starts a unified build+deploy operation. It triggers a build and,
// on success, automatically updates the image tag and syncs ArgoCD.
func (c *Client) Deploy(ctx context.Context, projectID, service, environment, gitRef, contextPath string) (*DeployOp, error) {
	ctx = auth.OutgoingContext(ctx)

	// Get source URL from packager
	resp, err := c.Packager.GetProject(ctx, &packager.GetProjectRequest{Project: projectID})
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	registry := deriveImagePath(c.RegistryPushURL, projectID, service)

	// Start the build
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

	deployID := uuid.New().String()
	c.DeployTracker.Create(deployID, buildResp.BuildId, projectID, service, environment)

	// Run the deploy pipeline in the background.
	// Extract the token before spawning the goroutine — the HTTP request context
	// will be cancelled when the response is sent.
	token := auth.TokenFrom(ctx)
	go c.runDeploy(token, deployID, projectID, service, environment, buildResp.BuildId)

	return deployOpFromState(c.DeployTracker.Get(deployID)), nil
}

// DeployStatus returns the current state of a deploy operation.
func (c *Client) DeployStatus(ctx context.Context, deployID string) (*DeployOp, error) {
	s := c.DeployTracker.Get(deployID)
	if s == nil {
		return nil, fmt.Errorf("deploy %q not found", deployID)
	}
	return deployOpFromState(s), nil
}

// ActiveDeployment returns the in-flight deploy for a project/service/environment, or nil.
func (c *Client) ActiveDeployment(ctx context.Context, projectID, service, environment string) (*DeployOp, error) {
	s := c.DeployTracker.ActiveForService(projectID, service, environment)
	if s == nil {
		return nil, nil
	}
	return deployOpFromState(s), nil
}

// runDeploy polls the builder for build status and, on success, deploys the image.
func (c *Client) runDeploy(token, deployID, projectID, service, environment, buildID string) {
	ctx := auth.WithToken(context.Background(), token)
	ctx = auth.OutgoingContext(ctx)

	for {
		time.Sleep(2 * time.Second)

		status, err := c.Builder.BuildStatus(ctx, &builder.BuildStatusRequest{BuildId: buildID})
		if err != nil {
			slog.Error("deploy: failed to poll build status", "deployId", deployID, "buildId", buildID, "error", err)
			c.DeployTracker.Fail(deployID, fmt.Sprintf("failed to poll build status: %v", err))
			return
		}

		phase := buildPhaseToDeployPhase(status.Phase)
		c.DeployTracker.Update(deployID, phase)

		switch status.Phase {
		case builder.BuildPhase_BUILD_PHASE_SUCCEEDED:
			c.finalizeDeploy(ctx, deployID, projectID, service, environment, status.ImageRef, status.Digest)
			return
		case builder.BuildPhase_BUILD_PHASE_FAILED:
			c.DeployTracker.Fail(deployID, status.Error)
			return
		}
	}
}

// finalizeDeploy updates the GitOps repo, triggers ArgoCD sync, and monitors rollout health.
func (c *Client) finalizeDeploy(ctx context.Context, deployID, projectID, service, environment, imageRef, digest string) {
	c.DeployTracker.Update(deployID, deploy.PhaseDeploying)

	tag := extractTag(imageRef)

	_, err := c.Packager.UpdateImageTag(ctx, &packager.UpdateImageTagRequest{
		Project:     projectID,
		Environment: environment,
		Service:     service,
		Tag:         tag,
		Digest:      digest,
	})
	if err != nil {
		c.DeployTracker.Fail(deployID, fmt.Sprintf("failed to update image tag: %v", err))
		return
	}

	// Trigger ArgoCD sync (best-effort)
	_, err = c.Deployer.SyncDeployment(ctx, &deployer.SyncDeploymentRequest{
		Project:     projectID,
		Environment: environment,
	})
	if err != nil {
		slog.Warn("deploy: failed to trigger sync", "deployId", deployID, "error", err)
	}

	// Poll ArgoCD for rollout health. This catches ImagePullBackOff, CrashLoopBackOff, etc.
	// Timeout after 2 minutes — pods should start well within that window.
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		time.Sleep(3 * time.Second)

		resp, err := c.Deployer.GetDeploymentStatus(ctx, &deployer.GetDeploymentStatusRequest{
			Project:     projectID,
			Environment: environment,
		})
		if err != nil {
			slog.Debug("deploy: failed to poll ArgoCD status", "deployId", deployID, "error", err)
			continue
		}

		health := deploymentStatusToString(resp.Status)
		c.DeployTracker.UpdateArgoHealth(deployID, health, resp.Message)

		switch resp.Status {
		case deployer.DeploymentStatus_DEPLOYMENT_STATUS_SYNCED:
			// Healthy + Synced — rollout succeeded
			c.DeployTracker.Succeed(deployID, imageRef, digest)
			slog.Info("deploy succeeded", "deployId", deployID, "project", projectID, "service", service, "environment", environment, "tag", tag)
			return
		case deployer.DeploymentStatus_DEPLOYMENT_STATUS_DEGRADED:
			// Degraded — pods failed (ImagePullBackOff, CrashLoopBackOff, etc.)
			c.DeployTracker.Fail(deployID, resp.Message)
			slog.Warn("deploy failed: ArgoCD reports degraded", "deployId", deployID, "project", projectID, "environment", environment, "message", resp.Message)
			return
		}
		// PROGRESSING, OUT_OF_SYNC, UNKNOWN — keep polling
	}

	// Timeout: mark succeeded but with the last known health status.
	// The rollout may still be in progress — we just stop tracking.
	c.DeployTracker.Succeed(deployID, imageRef, digest)
	slog.Info("deploy completed (ArgoCD still progressing)", "deployId", deployID, "project", projectID, "service", service, "environment", environment, "tag", tag)
}

func buildPhaseToDeployPhase(phase builder.BuildPhase) deploy.Phase {
	switch phase {
	case builder.BuildPhase_BUILD_PHASE_QUEUED:
		return deploy.PhaseQueued
	case builder.BuildPhase_BUILD_PHASE_CLONING:
		return deploy.PhaseCloning
	case builder.BuildPhase_BUILD_PHASE_BUILDING:
		return deploy.PhaseBuilding
	case builder.BuildPhase_BUILD_PHASE_PUSHING:
		return deploy.PhasePushing
	case builder.BuildPhase_BUILD_PHASE_SUCCEEDED:
		return deploy.PhaseSucceeded
	case builder.BuildPhase_BUILD_PHASE_FAILED:
		return deploy.PhaseFailed
	default:
		return deploy.PhaseQueued
	}
}

func extractTag(imageRef string) string {
	// Find the last ":" that comes after the last "/" to avoid splitting on
	// the port in registry URLs like "localhost:5000/myapp/web:0a04266".
	if i := strings.LastIndex(imageRef, ":"); i >= 0 {
		if j := strings.LastIndex(imageRef, "/"); i > j {
			return imageRef[i+1:]
		}
	}
	return imageRef
}
