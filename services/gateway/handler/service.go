package handler

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
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

func (c *Client) DetectServices(ctx context.Context, sourceURL string) ([]DetectedService, error) {
	ctx = auth.OutgoingContext(ctx)

	// Call builder to detect services (long — clones repo)
	detectCtx, detectCancel := context.WithTimeout(ctx, grpcLongTimeout)
	defer detectCancel()
	detectResp, err := c.Builder.DetectServices(detectCtx, &builder.DetectServicesRequest{
		SourceUrl: sourceURL,
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

func (c *Client) AddService(ctx context.Context, projectID, name string, port int, framework, sourceURL, contextPath string) (*Service, error) {
	ctx = auth.OutgoingContext(ctx)

	// Derive image path using cluster-internal address so pods can pull it.
	image := deriveImagePath(c.RegistryImagePrefix, projectID, name)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	_, err := c.Packager.AddService(callCtx, &packager.AddServiceRequest{
		Project:     projectID,
		Service:     name,
		Image:       image,
		Port:        int32(port),
		Framework:   framework,
		SourceUrl:   sourceURL,
		ContextPath: contextPath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add service: %w", err)
	}

	// Fire-and-forget: build the service so it has an image immediately.
	registry := deriveImagePath(c.RegistryPushURL, projectID, name)
	token := auth.TokenFrom(ctx)
	go func() {
		buildCtx := auth.WithToken(context.Background(), token)
		buildCtx = auth.OutgoingContext(buildCtx)
		buildCtx, cancel := context.WithTimeout(buildCtx, grpcTimeout)
		defer cancel()
		_, err := c.Builder.StartBuild(buildCtx, &builder.StartBuildRequest{
			SourceUrl:   sourceURL,
			GitRef:      "",
			Service:     name,
			Registry:    registry,
			ContextPath: contextPath,
		})
		if err != nil {
			slog.Warn("failed to start initial build", "project", projectID, "service", name, "error", err)
		} else {
			slog.Info("started initial build", "project", projectID, "service", name)
		}
	}()

	return &Service{
		Name:        name,
		Image:       image,
		Port:        port,
		Framework:   framework,
		SourceURL:   sourceURL,
		ContextPath: contextPath,
	}, nil
}

func (c *Client) RemoveService(ctx context.Context, projectID, service string) (bool, error) {
	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	_, err := c.Packager.RemoveService(callCtx, &packager.RemoveServiceRequest{
		Project: projectID,
		Service: service,
	})
	if err != nil {
		return false, fmt.Errorf("failed to remove service: %w", err)
	}
	return true, nil
}

func (c *Client) StartBuild(ctx context.Context, projectID, service, gitRef string) (*Build, error) {
	ctx = auth.OutgoingContext(ctx)

	// Look up source URL and context path from the service definition
	sourceURL, contextPath, err := c.serviceSourceInfo(ctx, projectID, service)
	if err != nil {
		return nil, err
	}

	registry := deriveImagePath(c.RegistryPushURL, projectID, service)

	buildCtx, buildCancel := context.WithTimeout(ctx, grpcTimeout)
	defer buildCancel()
	buildResp, err := c.Builder.StartBuild(buildCtx, &builder.StartBuildRequest{
		SourceUrl:   sourceURL,
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

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.Builder.BuildStatus(callCtx, &builder.BuildStatusRequest{BuildId: buildID})
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

	updateCtx, updateCancel := context.WithTimeout(ctx, grpcTimeout)
	defer updateCancel()
	_, err := c.Packager.UpdateImageTag(updateCtx, &packager.UpdateImageTagRequest{
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
	syncCtx, syncCancel := context.WithTimeout(ctx, grpcTimeout)
	defer syncCancel()
	_, err = c.Deployer.SyncDeployment(syncCtx, &deployer.SyncDeploymentRequest{
		Project:     projectID,
		Environment: environment,
	})
	if err != nil {
		slog.Warn("failed to trigger sync", "project", projectID, "environment", environment, "error", err)
	}

	return true, nil
}

// serviceSourceInfo looks up the source URL and context path for a service
// from the project's service definitions in the GitOps repo.
func (c *Client) serviceSourceInfo(ctx context.Context, projectID, service string) (sourceURL, contextPath string, err error) {
	getCtx, getCancel := context.WithTimeout(ctx, grpcTimeout)
	defer getCancel()
	resp, err := c.Packager.GetProject(getCtx, &packager.GetProjectRequest{Project: projectID})
	if err != nil {
		return "", "", fmt.Errorf("failed to get project: %w", err)
	}

	for _, svc := range resp.Project.Services {
		if svc.Name == service {
			return svc.SourceUrl, svc.ContextPath, nil
		}
	}
	return "", "", fmt.Errorf("service %q not found in project %q", service, projectID)
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
	RolloutHealth  string
	RolloutMessage string
}

func deployOpFromState(s *deploy.State) *DeployOp {
	return &DeployOp{
		ID:          s.ID,
		Phase:       string(s.Phase),
		BuildID:     s.BuildID,
		ImageRef:    s.ImageRef,
		Digest:      s.Digest,
		Error:       s.Error,
		RolloutHealth:  s.RolloutHealth,
		RolloutMessage: s.RolloutMessage,
	}
}

// Deploy starts a unified build+deploy operation. It triggers a build and,
// on success, automatically updates the image tag and syncs ArgoCD.
func (c *Client) Deploy(ctx context.Context, projectID, service, environment, gitRef string) (*DeployOp, error) {
	ctx = auth.OutgoingContext(ctx)

	// Look up source URL and context path from the service definition
	sourceURL, contextPath, err := c.serviceSourceInfo(ctx, projectID, service)
	if err != nil {
		return nil, err
	}

	registry := deriveImagePath(c.RegistryPushURL, projectID, service)

	// Start the build
	startCtx, startCancel := context.WithTimeout(ctx, grpcTimeout)
	defer startCancel()
	buildResp, err := c.Builder.StartBuild(startCtx, &builder.StartBuildRequest{
		SourceUrl:   sourceURL,
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

// maxBuildDuration is the maximum time to wait for a build to complete
// before failing the deploy. Prevents goroutine leaks from hung builds.
const maxBuildDuration = 30 * time.Minute

// runDeploy streams build logs from the builder and, on success, deploys the image.
func (c *Client) runDeploy(token, deployID, projectID, service, environment, buildID string) {
	ctx := auth.WithToken(context.Background(), token)
	ctx = auth.OutgoingContext(ctx)

	c.DeployTracker.AppendLog(deployID, "Queued for build...")

	// Stream build logs in a background goroutine.
	go c.streamBuildLogs(ctx, deployID, buildID)

	// Poll build status for phase transitions.
	deadline := time.Now().Add(maxBuildDuration)
	for time.Now().Before(deadline) {
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
			c.DeployTracker.AppendLog(deployID, "Build succeeded")
			c.finalizeDeploy(ctx, deployID, projectID, service, environment, status.ImageRef, status.Digest)
			return
		case builder.BuildPhase_BUILD_PHASE_FAILED:
			c.DeployTracker.AppendLog(deployID, fmt.Sprintf("Build failed: %s", status.Error))
			c.DeployTracker.Fail(deployID, status.Error)
			return
		}
	}

	// Build timed out — fail the deploy to prevent goroutine leaks.
	c.DeployTracker.AppendLog(deployID, fmt.Sprintf("Build timed out after %s", maxBuildDuration))
	c.DeployTracker.Fail(deployID, fmt.Sprintf("build timed out after %s", maxBuildDuration))
	slog.Error("deploy: build timed out", "deployId", deployID, "buildId", buildID, "timeout", maxBuildDuration)
}

// streamBuildLogs opens a gRPC stream to the builder and forwards log lines
// into the deploy tracker. Runs until the stream ends or an error occurs.
func (c *Client) streamBuildLogs(ctx context.Context, deployID, buildID string) {
	stream, err := c.Builder.BuildLogs(ctx, &builder.BuildLogsRequest{BuildId: buildID, Offset: 0})
	if err != nil {
		slog.Debug("deploy: failed to open build log stream", "deployId", deployID, "error", err)
		return
	}
	for {
		entry, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			slog.Debug("deploy: build log stream ended", "deployId", deployID, "error", err)
			return
		}
		c.DeployTracker.AppendLog(deployID, entry.Line)
	}
}

// finalizeDeploy updates the GitOps repo, triggers ArgoCD sync, and monitors rollout health.
func (c *Client) finalizeDeploy(ctx context.Context, deployID, projectID, service, environment, imageRef, digest string) {
	c.DeployTracker.Update(deployID, deploy.PhaseDeploying)

	tag := extractTag(imageRef)

	c.DeployTracker.AppendLog(deployID, fmt.Sprintf("Updating GitOps repo (tag: %s)", tag))
	_, err := c.Packager.UpdateImageTag(ctx, &packager.UpdateImageTagRequest{
		Project:     projectID,
		Environment: environment,
		Service:     service,
		Tag:         tag,
		Digest:      digest,
	})
	if err != nil {
		c.DeployTracker.AppendLog(deployID, fmt.Sprintf("Failed to update image tag: %v", err))
		c.DeployTracker.Fail(deployID, fmt.Sprintf("failed to update image tag: %v", err))
		return
	}

	c.DeployTracker.AppendLog(deployID, "Triggering ArgoCD sync...")
	// Trigger ArgoCD sync (best-effort)
	_, err = c.Deployer.SyncDeployment(ctx, &deployer.SyncDeploymentRequest{
		Project:     projectID,
		Environment: environment,
	})
	if err != nil {
		slog.Warn("deploy: failed to trigger sync", "deployId", deployID, "error", err)
		c.DeployTracker.AppendLog(deployID, fmt.Sprintf("Warning: sync trigger failed (%v), relying on auto-sync", err))
	}

	c.DeployTracker.AppendLog(deployID, "Waiting for rollout...")

	// Poll ArgoCD for rollout health. This catches ImagePullBackOff, CrashLoopBackOff, etc.
	// Timeout after 2 minutes — pods should start well within that window.
	deadline := time.Now().Add(2 * time.Minute)
	lastHealth := ""
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
		c.DeployTracker.UpdateRolloutHealth(deployID, health, resp.Message)

		// Log health changes.
		if health != lastHealth {
			msg := fmt.Sprintf("ArgoCD: %s", health)
			if resp.Message != "" {
				msg += fmt.Sprintf(" — %s", resp.Message)
			}
			c.DeployTracker.AppendLog(deployID, msg)
			lastHealth = health
		}

		switch resp.Status {
		case deployer.DeploymentStatus_DEPLOYMENT_STATUS_SYNCED:
			// Healthy + Synced — rollout succeeded
			c.DeployTracker.AppendLog(deployID, "Deploy succeeded")
			c.DeployTracker.Succeed(deployID, imageRef, digest)
			slog.Info("deploy succeeded", "deployId", deployID, "project", projectID, "service", service, "environment", environment, "tag", tag)
			return
		case deployer.DeploymentStatus_DEPLOYMENT_STATUS_DEGRADED:
			// Degraded — pods failed (ImagePullBackOff, CrashLoopBackOff, etc.)
			c.DeployTracker.AppendLog(deployID, fmt.Sprintf("Deploy failed: %s", resp.Message))
			c.DeployTracker.Fail(deployID, resp.Message)
			slog.Warn("deploy failed: ArgoCD reports degraded", "deployId", deployID, "project", projectID, "environment", environment, "message", resp.Message)
			return
		}
		// PROGRESSING, OUT_OF_SYNC, UNKNOWN — keep polling
	}

	// Timeout: mark succeeded but with the last known health status.
	// The rollout may still be in progress — we just stop tracking.
	c.DeployTracker.AppendLog(deployID, "Deploy completed (ArgoCD still progressing)")
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

// DeployLogs returns a channel of log lines for a deploy. The channel receives
// existing log lines (backlog) followed by new lines as they arrive. The channel
// is closed when the deploy reaches a terminal phase. The returned function
// unsubscribes from further updates.
func (c *Client) DeployLogs(ctx context.Context, deployID string) (<-chan string, func(), error) {
	s := c.DeployTracker.Get(deployID)
	if s == nil {
		return nil, nil, fmt.Errorf("deploy %q not found", deployID)
	}

	out := make(chan string, 128)
	sub, unsub := c.DeployTracker.Subscribe(deployID)

	go func() {
		defer close(out)

		// Send backlog.
		backlog := c.DeployTracker.LogLines(deployID, 0)
		for _, line := range backlog {
			select {
			case out <- line:
			case <-ctx.Done():
				return
			}
		}

		// Stream new lines from subscriber channel.
		done := c.DeployTracker.Done(deployID)
		for {
			select {
			case line, ok := <-sub:
				if !ok {
					return // deploy finished, channel closed
				}
				select {
				case out <- line:
				case <-ctx.Done():
					return
				}
			case <-done:
				// Drain any remaining lines in the subscriber channel.
				for line := range sub {
					select {
					case out <- line:
					case <-ctx.Done():
						return
					}
				}
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, unsub, nil
}

// Rollback updates the image tag to a previous value without rebuilding.
func (c *Client) Rollback(ctx context.Context, projectID, service, environment, imageTag string) (bool, error) {
	ctx = auth.OutgoingContext(ctx)

	updateCtx, updateCancel := context.WithTimeout(ctx, grpcTimeout)
	defer updateCancel()
	_, err := c.Packager.UpdateImageTag(updateCtx, &packager.UpdateImageTagRequest{
		Project:      projectID,
		Environment:  environment,
		Service:      service,
		Tag:          imageTag,
		CommitPrefix: "rollback",
	})
	if err != nil {
		return false, fmt.Errorf("failed to rollback: %w", err)
	}

	// Trigger ArgoCD sync
	syncCtx, syncCancel := context.WithTimeout(ctx, grpcTimeout)
	defer syncCancel()
	_, err = c.Deployer.SyncDeployment(syncCtx, &deployer.SyncDeploymentRequest{
		Project:     projectID,
		Environment: environment,
	})
	if err != nil {
		slog.Warn("failed to trigger sync after rollback", "project", projectID, "environment", environment, "error", err)
	}

	return true, nil
}

// Domain represents a domain hostname with its type and DNS status.
type Domain struct {
	Hostname  string
	Type      string // "PLATFORM" or "CUSTOM"
	DnsStatus string // "VALID", "PENDING", "MISCONFIGURED", or "ERROR"
}

// DnsCheck holds the result of a live DNS verification.
type DnsCheck struct {
	Hostname       string
	Status         string // "VALID", "PENDING", "MISCONFIGURED", "ERROR"
	CnameTarget    string // actual CNAME target found, empty if none
	ExpectedTarget string // platform's domain target
	Message        string // human-readable explanation
}

// PlatformConfig returns platform-level configuration for domain management.
func (c *Client) PlatformConfig() (workloadDomain, domainTarget string) {
	return c.WorkloadDomain, c.DomainTarget
}

// IsPlatformDomain checks if a hostname is a platform-generated domain.
func (c *Client) IsPlatformDomain(hostname string) bool {
	return strings.HasSuffix(hostname, "."+c.WorkloadDomain)
}

// CheckDns performs a live DNS check for a custom domain.
// It verifies that the domain has a CNAME record pointing to the platform's domain target.
func (c *Client) CheckDns(hostname string) DnsCheck {
	result := DnsCheck{
		Hostname:       hostname,
		ExpectedTarget: c.DomainTarget,
	}

	if c.IsPlatformDomain(hostname) {
		result.Status = "VALID"
		result.Message = "Platform domain"
		return result
	}

	lookupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := &net.Resolver{}

	// Check CNAME record.
	// Go's LookupCNAME returns the hostname itself (with trailing dot) when no CNAME exists.
	cname, err := resolver.LookupCNAME(lookupCtx, hostname)
	if err == nil && cname != "" {
		cname = strings.TrimSuffix(cname, ".")
		normalized := strings.TrimSuffix(hostname, ".")

		// If CNAME differs from the input hostname, a real CNAME record exists.
		if !strings.EqualFold(cname, normalized) {
			result.CnameTarget = cname
			expected := strings.TrimSuffix(c.DomainTarget, ".")
			if strings.EqualFold(cname, expected) {
				result.Status = "VALID"
				result.Message = "CNAME record verified"
				return result
			}
			result.Status = "MISCONFIGURED"
			result.Message = fmt.Sprintf("CNAME record points to %s, expected %s", cname, c.DomainTarget)
			return result
		}
	}

	// No CNAME found. Check if the domain resolves at all (A record).
	addrs, lookupErr := resolver.LookupHost(lookupCtx, hostname)
	if lookupErr != nil || len(addrs) == 0 {
		result.Status = "PENDING"
		result.Message = "No DNS record found. Add a CNAME record pointing to " + c.DomainTarget
		return result
	}

	// Domain resolves via A record but has no CNAME to the platform target.
	result.Status = "MISCONFIGURED"
	result.Message = fmt.Sprintf("Domain resolves via A record (%s) but has no CNAME to %s", addrs[0], c.DomainTarget)
	return result
}

// BuildDomain constructs a Domain struct with type and DNS status.
func (c *Client) BuildDomain(hostname string) Domain {
	domainType := "CUSTOM"
	if c.IsPlatformDomain(hostname) {
		domainType = "PLATFORM"
	}
	check := c.CheckDns(hostname)
	return Domain{
		Hostname:  hostname,
		Type:      domainType,
		DnsStatus: check.Status,
	}
}

// GenerateDomain creates a platform domain for a service in an environment.
// Format: {service}-{env}.{workloadDomain}. Appends a numeric suffix on collision.
func (c *Client) GenerateDomain(ctx context.Context, projectID, service, environment string) (*Domain, error) {
	ctx = auth.OutgoingContext(ctx)

	// Get all existing domains for collision detection
	allCtx, allCancel := context.WithTimeout(ctx, grpcTimeout)
	defer allCancel()
	allResp, err := c.Packager.AllDomains(allCtx, &packager.AllDomainsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list all domains: %w", err)
	}

	existing := make(map[string]bool, len(allResp.Hostnames))
	for _, h := range allResp.Hostnames {
		existing[h] = true
	}

	// Generate hostname: {service}-{env}.{workloadDomain}
	base := fmt.Sprintf("%s-%s.%s", service, environment, c.WorkloadDomain)
	hostname := base
	for i := 2; existing[hostname]; i++ {
		hostname = fmt.Sprintf("%s-%s-%d.%s", service, environment, i, c.WorkloadDomain)
	}

	// Add the domain
	addCtx, addCancel := context.WithTimeout(ctx, grpcTimeout)
	defer addCancel()
	_, err = c.Packager.AddDomain(addCtx, &packager.AddDomainRequest{
		Project:     projectID,
		Environment: environment,
		Service:     service,
		Hostname:    hostname,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add platform domain: %w", err)
	}

	// Trigger ArgoCD sync
	syncCtx, syncCancel := context.WithTimeout(ctx, grpcTimeout)
	defer syncCancel()
	_, err = c.Deployer.SyncDeployment(syncCtx, &deployer.SyncDeploymentRequest{
		Project:     projectID,
		Environment: environment,
	})
	if err != nil {
		slog.Warn("failed to trigger sync after domain add", "project", projectID, "environment", environment, "error", err)
	}

	d := c.BuildDomain(hostname)
	return &d, nil
}

// validateHostname checks that a hostname is a valid domain name.
func validateHostname(hostname string) error {
	// Strip common protocol prefixes users might paste.
	for _, prefix := range []string{"https://", "http://", "www."} {
		hostname = strings.TrimPrefix(hostname, prefix)
	}
	// Strip trailing dot (FQDN notation).
	hostname = strings.TrimSuffix(hostname, ".")
	// Strip trailing slash.
	hostname = strings.TrimRight(hostname, "/")

	if len(hostname) < 4 || len(hostname) > 253 {
		return fmt.Errorf("hostname must be between 4 and 253 characters")
	}

	// Must contain at least one dot (e.g. "example.com").
	if !strings.Contains(hostname, ".") {
		return fmt.Errorf("hostname must be a fully qualified domain name (e.g. api.example.com)")
	}

	labels := strings.Split(hostname, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return fmt.Errorf("each part of the hostname must be between 1 and 63 characters")
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return fmt.Errorf("hostname labels cannot start or end with a hyphen")
		}
		for _, ch := range label {
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-') {
				return fmt.Errorf("hostname contains invalid character %q — only letters, digits, and hyphens are allowed", ch)
			}
		}
	}

	return nil
}

// AddCustomDomain adds a user-specified custom domain to a service.
func (c *Client) AddCustomDomain(ctx context.Context, projectID, service, environment, hostname string) (*Domain, error) {
	ctx = auth.OutgoingContext(ctx)

	if err := validateHostname(hostname); err != nil {
		return nil, fmt.Errorf("invalid hostname: %w", err)
	}

	// Reject platform domains — those should use GenerateDomain.
	if c.IsPlatformDomain(hostname) {
		return nil, fmt.Errorf("cannot add a platform domain as a custom domain — use Generate Domain instead")
	}

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	_, err := c.Packager.AddDomain(callCtx, &packager.AddDomainRequest{
		Project:     projectID,
		Environment: environment,
		Service:     service,
		Hostname:    hostname,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add custom domain: %w", err)
	}

	// Trigger ArgoCD sync
	syncCtx, syncCancel := context.WithTimeout(ctx, grpcTimeout)
	defer syncCancel()
	_, err = c.Deployer.SyncDeployment(syncCtx, &deployer.SyncDeploymentRequest{
		Project:     projectID,
		Environment: environment,
	})
	if err != nil {
		slog.Warn("failed to trigger sync after domain add", "project", projectID, "environment", environment, "error", err)
	}

	d := c.BuildDomain(hostname)
	return &d, nil
}

// RemoveDomain removes a domain from a service in an environment.
func (c *Client) RemoveDomain(ctx context.Context, projectID, service, environment, hostname string) (bool, error) {
	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	_, err := c.Packager.RemoveDomain(callCtx, &packager.RemoveDomainRequest{
		Project:     projectID,
		Environment: environment,
		Service:     service,
		Hostname:    hostname,
	})
	if err != nil {
		return false, fmt.Errorf("failed to remove domain: %w", err)
	}

	// Trigger ArgoCD sync
	syncCtx, syncCancel := context.WithTimeout(ctx, grpcTimeout)
	defer syncCancel()
	_, err = c.Deployer.SyncDeployment(syncCtx, &deployer.SyncDeploymentRequest{
		Project:     projectID,
		Environment: environment,
	})
	if err != nil {
		slog.Warn("failed to trigger sync after domain remove", "project", projectID, "environment", environment, "error", err)
	}

	return true, nil
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
