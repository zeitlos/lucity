package conductor

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"regexp"
	"strings"
	"time"

	gh "github.com/google/go-github/v68/github"
	"github.com/google/uuid"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/tenant"
	"github.com/zeitlos/lucity/services/conductor/internal/api/deploy"
	"github.com/zeitlos/lucity/services/conductor/internal/data"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type ServiceID = platform.ServiceID
type DeploymentID = platform.DeploymentID
type Service = platform.Service
type DetectedService struct {
	Name          string
	Provider      string
	Framework     string
	StartCommand  string
	SuggestedPort int
}

func (c *Client) DetectServices(ctx context.Context, repository string, installationID int64) ([]DetectedService, error) {
	if _, err := tenant.FromContext(ctx); err != nil {
		return nil, err
	}
	sourceURL, err := c.resolveRepositoryURL(ctx, installationID, repository)
	if err != nil {
		return nil, err
	}
	ctx, err = c.withInstallationTokenForID(ctx, installationID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with GitHub: %w", err)
	}

	// Call builder to detect services (long — clones repo)
	detected, err := c.Builder.DetectServices(ctx, sourceURL, "")
	if err != nil {
		return nil, fmt.Errorf("failed to detect services: %w", err)
	}

	result := make([]DetectedService, 0, len(detected))
	for _, s := range detected {
		result = append(result, DetectedService{
			Name:          s.Name,
			Provider:      s.Provider,
			Framework:     s.Framework,
			StartCommand:  s.StartCommand,
			SuggestedPort: s.SuggestedPort,
		})
	}
	return result, nil
}

func (c *Client) AddService(ctx context.Context, environment platform.EnvironmentID, name string, port int, framework, startCommand, repository, contextPath string, installationID *int64, externalImage, customStartCommand string) (*Service, error) {
	ws := environment.Workspace
	projectID := environment.Project
	envName := environment.Name
	_ = envName
	// For source-based services, resolve repository to a verified clone URL.
	var sourceURL string
	var err error
	if repository != "" {
		if installationID == nil {
			return nil, fmt.Errorf("installationId is required when repository is set")
		}
		sourceURL, err = c.resolveRepositoryURL(ctx, *installationID, repository)
		if err != nil {
			return nil, err
		}
		ctx, err = c.withInstallationTokenForID(ctx, *installationID)
		if err != nil {
			return nil, fmt.Errorf("failed to authenticate with GitHub: %w", err)
		}
	} else if installationID != nil {
		ctx, err = c.withInstallationTokenForID(ctx, *installationID)
		if err != nil {
			return nil, fmt.Errorf("failed to authenticate with GitHub: %w", err)
		}
	}

	var ghInstallationID int64
	if installationID != nil {
		ghInstallationID = *installationID
	}

	// Derive name and port from image when not explicitly provided.
	if externalImage != "" {
		if name == "" {
			name = deriveServiceName(externalImage)
		}
		if port == 0 {
			port = defaultPortForImage(externalImage)
		}
	}

	// For external images, use the provided reference directly.
	// For source-based services, derive from the internal registry.
	var image, imageTag, imagePullPolicy string
	if externalImage != "" {
		image, imageTag = parseImageRef(externalImage)
		imagePullPolicy = "Always"
	} else {
		image = deriveImagePath(c.Config.RegistryImagePrefix, ws, projectID, name)
	}

	if err = c.Packager.AddService(ctx, ws, projectID, data.ServiceDef{
		Name:                 name,
		Image:                image,
		Port:                 port,
		Framework:            framework,
		SourceURL:            sourceURL,
		ContextPath:          contextPath,
		GitHubInstallationID: ghInstallationID,
		ImageTag:             imageTag,
		ImagePullPolicy:      imagePullPolicy,
		CustomStartCommand:   customStartCommand,
		StartCommand:         startCommand,
		Environment:          envName,
	}); err != nil {
		return nil, fmt.Errorf("failed to add service: %w", err)
	}

	service := &platform.Service{
		ID: platform.ServiceID{
			Workspace:   ws,
			Project:     projectID,
			Environment: envName,
			Name:        name,
		},
		Name: name,
	}

	// Trigger initial deploy for source-based services.
	if sourceURL != "" {
		registry := deriveImagePath(c.Config.RegistryPushURL, ws, projectID, name)

		buildID, err := c.Builder.StartBuild(ctx, sourceURL, "", name, registry, contextPath)
		if err != nil {
			slog.Warn("failed to start initial deploy", "project", projectID, "service", name, "error", err)
			return service, nil
		}

		deployID := uuid.New().String()
		c.DeployTracker.Create(deployID, buildID, projectID, name, envName)

		claims, err := auth.FromContext(ctx)

		if err != nil {
			return nil, err
		}

		go c.runDeploy(claims, ws, deployID, projectID, name, envName, buildID)
	}

	return service, nil
}

// wellKnownPorts maps common container image names to their default ports.
var wellKnownPorts = map[string]int{
	"nginx":         80,
	"httpd":         80,
	"apache":        80,
	"caddy":         80,
	"traefik":       80,
	"redis":         6379,
	"valkey":        6379,
	"postgres":      5432,
	"postgresql":    5432,
	"mysql":         3306,
	"mariadb":       3306,
	"mongo":         27017,
	"mongodb":       27017,
	"memcached":     11211,
	"rabbitmq":      5672,
	"nats":          4222,
	"elasticsearch": 9200,
	"opensearch":    9200,
	"minio":         9000,
	"grafana":       3000,
	"prometheus":    9090,
	"clickhouse":    8123,
	"influxdb":      8086,
	"vault":         8200,
	"consul":        8500,
	"etcd":          2379,
}

// defaultPortForImage returns a well-known port for the image, or 80 as fallback.
func defaultPortForImage(imageRef string) int {
	name := imageRef
	// Strip tag
	if i := strings.LastIndex(name, ":"); i >= 0 {
		if j := strings.LastIndex(name, "/"); i > j {
			name = name[:i]
		}
	}
	// Use last path segment (e.g., "bitnami/redis" → "redis")
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	if port, ok := wellKnownPorts[name]; ok {
		return port
	}
	return 80
}

// deriveServiceName extracts a service name from an image reference.
// e.g., "nginx:1.25" → "nginx", "ghcr.io/foo/my-app:v1" → "my-app"
func deriveServiceName(imageRef string) string {
	name := imageRef
	// Strip tag
	if i := strings.LastIndex(name, ":"); i >= 0 {
		if j := strings.LastIndex(name, "/"); i > j {
			name = name[:i]
		}
	}
	// Use last path segment
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	return name
}

// parseImageRef splits a container image reference into repository and tag.
// Handles registry:port/repo:tag by finding the last ":" after the last "/".
func parseImageRef(ref string) (repository, tag string) {
	if i := strings.LastIndex(ref, ":"); i >= 0 {
		if j := strings.LastIndex(ref, "/"); i > j {
			return ref[:i], ref[i+1:]
		}
	}
	return ref, "latest"
}

func (c *Client) RemoveService(ctx context.Context, svc platform.ServiceID) (bool, error) {
	projectID := svc.Project
	environment := svc.Environment
	service := svc.Name
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return false, err
	}

	if err := c.Packager.RemoveService(ctx, ws, projectID, environment, service); err != nil {
		return false, fmt.Errorf("failed to remove service: %w", err)
	}
	return true, nil
}

// SetCustomStartCommand sets or clears the custom start command for a service.
func (c *Client) SetCustomStartCommand(ctx context.Context, svc platform.ServiceID, command string) (bool, error) {
	projectID := svc.Project
	environment := svc.Environment
	service := svc.Name
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return false, err
	}

	if err := c.Packager.SetCustomStartCommand(ctx, ws, projectID, environment, service, command); err != nil {
		return false, fmt.Errorf("failed to set custom start command: %w", err)
	}
	return true, nil
}

// serviceSourceInfo looks up the source URL, context path, and GitHub installation ID
// for a service from the project's environment data in the GitOps repo.
func (c *Client) serviceSourceInfo(ctx context.Context, ws, projectID, service string) (sourceURL, contextPath string, installationID int64, err error) {
	info, err := c.Packager.GetProject(ctx, ws, projectID)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to get project: %w", err)
	}

	for _, env := range info.EnvironmentInfos {
		for _, svc := range env.Services {
			if svc.Name == service {
				return svc.SourceURL, svc.ContextPath, svc.GitHubInstallationID, nil
			}
		}
	}
	return "", "", 0, fmt.Errorf("service %q not found in project %q", service, projectID)
}

// withInstallationTokenForID mints a GitHub App installation token for the given
// installation ID and attaches it to the context for downstream gRPC calls.
func (c *Client) withInstallationTokenForID(ctx context.Context, installationID int64) (context.Context, error) {
	if c.GitHubApp == nil {
		return ctx, fmt.Errorf("github app not configured")
	}
	if installationID == 0 {
		return ctx, nil
	}

	token, err := c.GitHubApp.InstallationToken(ctx, installationID)
	if err != nil {
		return ctx, fmt.Errorf("failed to mint installation token: %w", err)
	}

	return auth.WithGitHubToken(ctx, token), nil
}

// deriveImagePath builds a registry image path scoped by workspace.
// workspace "acme" + project "api" + service "web" → "localhost:5000/acme/api/web"
func deriveImagePath(registryURL, workspace, project, service string) string {
	return registryURL + "/" + workspace + "/" + project + "/" + service
}

// DeployOp represents the state of a unified build+deploy operation.
type DeployOp struct {
	ID             string
	Phase          string
	BuildID        string
	ImageRef       string
	Digest         string
	Error          string
	RolloutHealth  string
	RolloutMessage string
	StartedAt      time.Time
}

func deployOpFromState(s *deploy.State) *DeployOp {
	return &DeployOp{
		ID:             s.ID,
		Phase:          string(s.Phase),
		BuildID:        s.BuildID,
		ImageRef:       s.ImageRef,
		Digest:         s.Digest,
		Error:          s.Error,
		RolloutHealth:  s.RolloutHealth,
		RolloutMessage: s.RolloutMessage,
		StartedAt:      s.StartedAt,
	}
}

// Deploy starts a unified build+deploy operation. It triggers a build and,
// on success, automatically updates the image tag and syncs ArgoCD.
func (c *Client) Deploy(ctx context.Context, svc platform.ServiceID, gitRef string) (*DeployOp, error) {
	ws := svc.Workspace
	projectID := svc.Project
	environment := svc.Environment
	service := svc.Name
	// Look up source URL, context path, and installation ID from the service definition
	sourceURL, contextPath, installationID, err := c.serviceSourceInfo(ctx, ws, projectID, service)
	if err != nil {
		return nil, err
	}
	if sourceURL == "" {
		return nil, fmt.Errorf("cannot deploy %q: service has no source repository (image-based services are deployed automatically)", service)
	}
	if installationID != 0 {
		ctx, err = c.withInstallationTokenForID(ctx, installationID)
		if err != nil {
			return nil, fmt.Errorf("failed to authenticate with GitHub: %w", err)
		}
	}

	registry := deriveImagePath(c.Config.RegistryPushURL, ws, projectID, service)

	// Start the build
	buildID, err := c.Builder.StartBuild(ctx, sourceURL, gitRef, service, registry, contextPath)
	if err != nil {
		return nil, fmt.Errorf("failed to start build: %w", err)
	}

	deployID := uuid.New().String()
	c.DeployTracker.Create(deployID, buildID, projectID, service, environment)

	// Run the deploy pipeline in the background.
	// Extract the token before spawning the goroutine — the HTTP request context
	// will be cancelled when the response is sent.
	claims, err := auth.FromContext(ctx)

	if err != nil {
		return nil, err
	}

	go c.runDeploy(claims, ws, deployID, projectID, service, environment, buildID)

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
func (c *Client) ActiveDeployment(ctx context.Context, svc platform.ServiceID) (*DeployOp, error) {
	projectID := svc.Project
	service := svc.Name
	environment := svc.Environment
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
func (c *Client) runDeploy(claims *auth.Claims, workspace, deployID, projectID, service, environment, buildID string) {
	// Build a base context carrying user identity for inproc calls.
	// Workspace is passed explicitly to each inproc method below.
	base := auth.NewContext(context.Background(), claims)

	slog.Info("deploy: goroutine started", "deployId", deployID, "buildId", buildID, "hasClaims", claims != nil, "workspace", workspace)
	c.DeployTracker.AppendLog(deployID, "Queued for build...")

	// Stream build logs in a background goroutine, with a long timeout
	// (longer than grpcCtx's polling timeout used elsewhere).
	logCtx, logCancel := context.WithTimeout(base, maxBuildDuration)
	go func() {
		defer logCancel()
		c.streamBuildLogs(logCtx, deployID, buildID)
	}()

	// Poll build status for phase transitions.
	deadline := time.Now().Add(maxBuildDuration)
	for time.Now().Before(deadline) {
		time.Sleep(2 * time.Second)

		status, err := c.Builder.BuildStatus(base, buildID)
		if err != nil {
			slog.Error("deploy: failed to poll build status", "deployId", deployID, "buildId", buildID, "error", err)
			c.DeployTracker.Fail(deployID, fmt.Sprintf("failed to poll build status: %v", err))
			return
		}

		phase := buildPhaseToDeployPhase(status.Phase)
		c.DeployTracker.Update(deployID, phase)

		switch status.Phase {
		case data.BuildPhaseSucceeded:
			c.DeployTracker.AppendLog(deployID, "Build succeeded")
			c.finalizeDeploy(base, workspace, deployID, projectID, service, environment, status.ImageRef, status.Digest)
			return
		case data.BuildPhaseFailed:
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

// streamBuildLogs reads from the builder's build-log channel and
// forwards lines into the deploy tracker. Runs until the channel
// closes or ctx is cancelled.
func (c *Client) streamBuildLogs(ctx context.Context, deployID, buildID string) {
	lines, err := c.Builder.BuildLogs(ctx, buildID, 0)
	if err != nil {
		slog.Warn("deploy: failed to open build log stream", "deployId", deployID, "error", err)
		return
	}
	for line := range lines {
		c.DeployTracker.AppendLog(deployID, line)
	}
}

// finalizeDeploy updates the GitOps repo, triggers ArgoCD sync, and monitors rollout health.
func (c *Client) finalizeDeploy(ctx context.Context, ws, deployID, projectID, service, environment, imageRef, digest string) {
	c.DeployTracker.Update(deployID, deploy.PhaseDeploying)

	tag, _ := imageParts(imageRef)

	c.DeployTracker.AppendLog(deployID, fmt.Sprintf("Updating GitOps repo (tag: %s)", tag))
	if err := c.Packager.UpdateImageTag(ctx, ws, projectID, environment, service, tag, digest, ""); err != nil {
		c.DeployTracker.AppendLog(deployID, fmt.Sprintf("Failed to update image tag: %v", err))
		c.DeployTracker.Fail(deployID, fmt.Sprintf("failed to update image tag: %v", err))
		return
	}

	c.DeployTracker.AppendLog(deployID, "Triggering ArgoCD sync...")
	// Trigger ArgoCD sync (best-effort)
	if _, err := c.Deployer.SyncDeployment(ctx, ws, projectID, environment); err != nil {
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

		st, message, err := c.Deployer.GetDeploymentStatus(ctx, ws, projectID, environment)
		if err != nil {
			slog.Warn("deploy: failed to poll ArgoCD status", "deployId", deployID, "error", err)
			continue
		}

		health := string(st)
		c.DeployTracker.UpdateRolloutHealth(deployID, health, message)

		// Log health changes.
		if health != lastHealth {
			msg := fmt.Sprintf("ArgoCD: %s", health)
			if message != "" {
				msg += fmt.Sprintf(" — %s", message)
			}
			c.DeployTracker.AppendLog(deployID, msg)
			lastHealth = health
		}

		switch st {
		case data.DeploymentStatusSynced:
			// Healthy + Synced — rollout succeeded
			c.DeployTracker.AppendLog(deployID, "Deploy succeeded")
			c.DeployTracker.Succeed(deployID, imageRef, digest)
			slog.Info("deploy succeeded", "deployId", deployID, "project", projectID, "service", service, "environment", environment, "tag", tag)
			return
		case data.DeploymentStatusDegraded:
			// Degraded — pods failed (ImagePullBackOff, CrashLoopBackOff, etc.)
			c.DeployTracker.AppendLog(deployID, fmt.Sprintf("Deploy failed: %s", message))
			c.DeployTracker.Fail(deployID, message)
			slog.Warn("deploy failed: ArgoCD reports degraded", "deployId", deployID, "project", projectID, "environment", environment, "message", message)
			return
		}
		// PROGRESSING, OUT_OF_SYNC, UNKNOWN — keep polling
	}

	// Timeout: stop tracking. The image tag is committed — ArgoCD will eventually sync.
	// Readiness is derived from K8s Deployment status, not from this tracker.
	c.DeployTracker.AppendLog(deployID, "Deploy tracking complete — pods may still be starting")
	c.DeployTracker.Succeed(deployID, imageRef, digest)
	slog.Info("deploy tracking complete", "deployId", deployID, "project", projectID, "service", service, "environment", environment, "tag", tag)
}

func buildPhaseToDeployPhase(phase data.BuildPhase) deploy.Phase {
	switch phase {
	case data.BuildPhaseQueued:
		return deploy.PhaseQueued
	case data.BuildPhaseCloning:
		return deploy.PhaseCloning
	case data.BuildPhaseBuilding:
		return deploy.PhaseBuilding
	case data.BuildPhasePushing:
		return deploy.PhasePushing
	case data.BuildPhaseSucceeded:
		return deploy.PhaseSucceeded
	case data.BuildPhaseFailed:
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
func (c *Client) Rollback(ctx context.Context, deploymentID DeploymentID) (bool, error) {
	projectID := deploymentID.Project
	service := deploymentID.Service
	environment := deploymentID.Environment

	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return false, err
	}

	deployment, err := c.platform.Deployment(ctx, deploymentID)

	if err != nil {
		return false, err
	}

	tag, _ := imageParts(deployment.Image)

	if err := c.Packager.UpdateImageTag(ctx, ws, projectID, environment, service, tag, "", "rollback"); err != nil {
		return false, fmt.Errorf("failed to rollback: %w", err)
	}

	// Trigger ArgoCD sync
	if _, err := c.Deployer.SyncDeployment(ctx, ws, projectID, environment); err != nil {
		slog.Warn("failed to trigger sync after rollback", "project", projectID, "environment", environment, "error", err)
	}

	return true, nil
}

// Domain represents a domain hostname with its type, DNS status, and TLS status.
type Domain struct {
	Hostname  string
	Type      string // "PLATFORM" or "CUSTOM"
	DnsStatus string // "VALID", "PENDING", "MISCONFIGURED", or "ERROR"
	TlsStatus string // "NONE", "PROVISIONING", "ACTIVE", or "ERROR"
}

// DnsCheck holds the result of a live DNS verification.
type DnsCheck struct {
	Hostname       string
	Status         string // "VALID", "PENDING", "MISCONFIGURED", "ERROR"
	CnameTarget    string // actual CNAME target found, empty if none
	ExpectedTarget string // platform's domain target
	Message        string // human-readable explanation
	TlsStatus      string // "NONE", "PROVISIONING", "ACTIVE", "ERROR" (custom domains only)
}

// PlatformConfig returns platform-level configuration for domain management.
func (c *Client) PlatformConfig() (workloadDomain, domainTarget, ipAddress string) {
	return c.Config.WorkloadDomain, c.Config.DomainTarget, c.Config.IPAddress
}

// IsPlatformDomain checks if a hostname is a platform-generated domain.
func (c *Client) IsPlatformDomain(hostname string) bool {
	return strings.HasSuffix(hostname, "."+c.Config.WorkloadDomain)
}

// CheckDns performs a live DNS check for a custom domain.
// It verifies that the domain has a CNAME record pointing to the platform's domain target.
// For custom domains, also looks up TLS certificate status from the deployer.
func (c *Client) CheckDns(ctx context.Context, hostname string) DnsCheck {
	result := DnsCheck{
		Hostname:       hostname,
		ExpectedTarget: c.Config.DomainTarget,
	}

	if c.IsPlatformDomain(hostname) {
		result.Status = "VALID"
		result.Message = "Platform domain"
		return result
	}

	// Look up TLS cert status for custom domains.
	cdStatus, err := c.Deployer.CustomDomainStatus(ctx, hostname)
	if err != nil {
		slog.Debug("failed to check TLS status", "hostname", hostname, "error", err)
	} else {
		result.TlsStatus = cdStatus.TLSStatus
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
			expected := strings.TrimSuffix(c.Config.DomainTarget, ".")
			if strings.EqualFold(cname, expected) {
				result.Status = "VALID"
				result.Message = "CNAME record verified"
				return result
			}
			result.Status = "MISCONFIGURED"
			result.Message = fmt.Sprintf("CNAME record points to %s, expected %s", cname, c.Config.DomainTarget)
			return result
		}
	}

	// No CNAME found. Check if the domain resolves at all (A record).
	addrs, lookupErr := resolver.LookupHost(lookupCtx, hostname)
	if lookupErr != nil || len(addrs) == 0 {
		result.Status = "PENDING"
		// Apex domains (e.g. "example.com") can't use CNAME — suggest A record instead.
		parts := strings.Split(hostname, ".")
		if len(parts) <= 2 && c.Config.IPAddress != "" {
			result.Message = fmt.Sprintf("No DNS record found. Add an A record pointing to %s", c.Config.IPAddress)
		} else {
			result.Message = "No DNS record found. Add a CNAME record pointing to " + c.Config.DomainTarget
		}
		return result
	}

	// Domain resolves via A record. Check if it points to our LB.
	if c.Config.IPAddress != "" {
		for _, addr := range addrs {
			if addr == c.Config.IPAddress {
				result.Status = "VALID"
				result.Message = fmt.Sprintf("A record points to platform load balancer (%s)", c.Config.IPAddress)
				return result
			}
		}
	}
	result.Status = "MISCONFIGURED"
	result.Message = fmt.Sprintf("Domain resolves to %s but expected CNAME to %s or A record to %s", addrs[0], c.Config.DomainTarget, c.Config.IPAddress)
	return result
}

// buildDomain constructs a Domain struct with type, DNS status, and TLS status.
func (c *Client) buildDomain(ctx context.Context, hostname string) Domain {
	domainType := "CUSTOM"

	if c.IsPlatformDomain(hostname) {
		domainType = "PLATFORM"
	}

	check := c.CheckDns(ctx, hostname)

	return Domain{
		Hostname:  hostname,
		Type:      domainType,
		DnsStatus: check.Status,
		TlsStatus: check.TlsStatus,
	}
}

// GenerateDomain creates a platform domain for a service in an environment.
// Format: {service}-{env}-{randomSuffix}.{workloadDomain}.
func (c *Client) GenerateDomain(ctx context.Context, svc platform.ServiceID) (*Domain, error) {
	projectID := svc.Project
	service := svc.Name
	environment := svc.Environment
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	hostname, err := c.Packager.GeneratePlatformDomain(ctx, ws, projectID, environment, service)
	if err != nil {
		return nil, fmt.Errorf("failed to generate platform domain: %w", err)
	}

	if _, err := c.Deployer.SyncDeployment(ctx, ws, projectID, environment); err != nil {
		slog.Warn("failed to trigger sync after domain add", "project", projectID, "environment", environment, "error", err)
	}

	d := c.buildDomain(ctx, hostname)
	return &d, nil
}

// repositoryPattern matches valid GitHub owner/repo format.
// Allows alphanumeric, hyphens, underscores, and dots (GitHub's rules).
var repositoryPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`)

// validateRepository checks that a repository string is a valid owner/repo format.
func validateRepository(repository string) (owner, repo string, err error) {
	if !repositoryPattern.MatchString(repository) {
		return "", "", fmt.Errorf("repository must be in owner/repo format (e.g. \"acme/myapp\")")
	}
	parts := strings.SplitN(repository, "/", 2)
	return parts[0], parts[1], nil
}

// resolveRepositoryURL validates a repository string, verifies it's accessible
// through the given GitHub App installation, and returns the HTTPS clone URL.
// The URL is constructed server-side from the verified owner/repo, never from user input.
func (c *Client) resolveRepositoryURL(ctx context.Context, installationID int64, repository string) (string, error) {
	owner, repo, err := validateRepository(repository)
	if err != nil {
		return "", err
	}

	if c.GitHubApp == nil {
		return "", fmt.Errorf("github app not configured")
	}

	token, err := c.GitHubApp.InstallationToken(ctx, installationID)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate with GitHub: %w", err)
	}

	client := gh.NewClient(nil).WithAuthToken(token)
	ghRepo, _, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return "", fmt.Errorf("repository %q not accessible via this GitHub App installation: %w", repository, err)
	}

	// Use the clone URL from GitHub's API response, not user input.
	cloneURL := ghRepo.GetCloneURL()
	if cloneURL == "" {
		cloneURL = "https://github.com/" + owner + "/" + repo
	}
	return cloneURL, nil
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
func (c *Client) AddCustomDomain(ctx context.Context, svc platform.ServiceID, hostname string) (*Domain, error) {
	projectID := svc.Project
	service := svc.Name
	environment := svc.Environment
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := validateHostname(hostname); err != nil {
		return nil, fmt.Errorf("invalid hostname: %w", err)
	}

	// Reject platform domains — those should use GenerateDomain.
	if c.IsPlatformDomain(hostname) {
		return nil, fmt.Errorf("cannot add a platform domain as a custom domain — use Generate Domain instead")
	}

	if err := c.Packager.AddDomain(ctx, ws, projectID, environment, service, hostname); err != nil {
		return nil, fmt.Errorf("failed to add custom domain: %w", err)
	}

	// Provision TLS certificate
	tlsStatus, provErr := c.Deployer.ProvisionCustomDomain(ctx, hostname)
	if provErr != nil {
		slog.Warn("failed to provision TLS certificate", "hostname", hostname, "error", provErr)
		tlsStatus = "ERROR"
	}

	// Trigger ArgoCD sync
	if _, err := c.Deployer.SyncDeployment(ctx, ws, projectID, environment); err != nil {
		slog.Warn("failed to trigger sync after domain add", "project", projectID, "environment", environment, "error", err)
	}

	domainType := "CUSTOM"
	check := c.CheckDns(ctx, hostname)
	d := &Domain{
		Hostname:  hostname,
		Type:      domainType,
		DnsStatus: check.Status,
		TlsStatus: tlsStatus,
	}
	return d, nil
}

// RemoveDomain removes a domain from a service in an environment.
func (c *Client) RemoveDomain(ctx context.Context, svc platform.ServiceID, hostname string) (bool, error) {
	projectID := svc.Project
	service := svc.Name
	environment := svc.Environment
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return false, err
	}

	if err := c.Packager.RemoveDomain(ctx, ws, projectID, environment, service, hostname); err != nil {
		return false, fmt.Errorf("failed to remove domain: %w", err)
	}

	// Delete TLS certificate for custom domains
	if !c.IsPlatformDomain(hostname) {
		if delErr := c.Deployer.DeleteCustomDomain(ctx, hostname); delErr != nil {
			slog.Warn("failed to delete TLS certificate", "hostname", hostname, "error", delErr)
		}
	}

	// Trigger ArgoCD sync
	if _, err := c.Deployer.SyncDeployment(ctx, ws, projectID, environment); err != nil {
		slog.Warn("failed to trigger sync after domain remove", "project", projectID, "environment", environment, "error", err)
	}

	return true, nil
}

func imageParts(image string) (tag, digest string) {
	if i := strings.Index(image, "@sha256:"); i != -1 {
		digest = image[i+1:]
		image = image[:i]
	}

	// Find tag — last colon must be AFTER the last slash (port-safe for "host:5000/repo")
	slashIdx := strings.LastIndex(image, "/")
	colonIdx := strings.LastIndex(image, ":")

	if colonIdx > slashIdx {
		tag = image[colonIdx+1:]
	}

	return tag, digest
}
