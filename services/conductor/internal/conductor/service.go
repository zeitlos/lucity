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

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/tenant"
	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"
	"github.com/zeitlos/lucity/services/conductor/internal/data"
	"github.com/zeitlos/lucity/services/conductor/internal/planner"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type ServiceID = platform.ServiceID
type DeploymentID = platform.DeploymentID
type Service = platform.Service
type Plan = planner.Plan

func (c *Client) Services(ctx context.Context, environmentID EnvironmentID) ([]Service, error) {
	return c.platform.Services(ctx, environmentID)
}

func (c *Client) Service(ctx context.Context, id ServiceID) (*Service, error) {
	return c.platform.Service(ctx, id)
}

func (c *Client) DetectServices(ctx context.Context, repositoryURL string, installationID int64) ([]Plan, error) {
	if _, err := tenant.FromContext(ctx); err != nil {
		return nil, err
	}

	commit, err := c.source.CommitSHA(ctx, repositoryURL, "")

	if err != nil {
		return nil, err
	}

	token, err := c.source.Token(ctx, repositoryURL)

	if err != nil {
		return nil, err
	}

	return c.planner.Plan(ctx, repositoryURL, commit, token)
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
		commit, err := c.source.CommitSHA(ctx, sourceURL, "")

		if err != nil {
			slog.Warn("failed to resolve initial commit", "project", projectID, "service", name, "error", err)
			return service, nil
		}

		token, err := c.source.Token(ctx, sourceURL)

		if err != nil {
			slog.Warn("failed to mint source token", "project", projectID, "service", name, "error", err)
			return service, nil
		}

		imageName := ws + "/" + projectID + "/" + name

		build, err := c.buildjob.Start(ctx, buildjob.StartOptions{
			Workspace:        ws,
			RepoURL:          sourceURL,
			Commit:           commit,
			ContextPath:      contextPath,
			TargetImageNames: []string{imageName},
			Token:            token,
		})

		if err != nil {
			slog.Warn("failed to start initial build", "project", projectID, "service", name, "error", err)
			return service, nil
		}

		claims, _ := auth.FromContext(ctx)

		go c.runDeploy(claims, platform.ServiceID{
			Workspace:   ws,
			Project:     projectID,
			Environment: envName,
			Name:        name,
		}, build.ID)
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

func (c *Client) SetCustomStartCommand(serviceID context.Context, svc platform.ServiceID, command string) (*Service, error) {
	projectID := svc.Project
	environment := svc.Environment
	service := svc.Name
	ws, err := tenant.FromContext(serviceID)

	if err != nil {
		return nil, err
	}

	if err := c.Packager.SetCustomStartCommand(serviceID, ws, projectID, environment, service, command); err != nil {
		return nil, fmt.Errorf("failed to set custom start command: %w", err)
	}

	return c.Service(serviceID, svc)
}

// withInstallationTokenForID mints a GitHub App installation token for the given
// installation ID and attaches it to the context for downstream gRPC calls.
func (c *Client) withInstallationTokenForID(ctx context.Context, installationID int64) (context.Context, error) {
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

// maxBuildDuration caps the post-build deploy goroutine to prevent leaks
// from hung builds.
const maxBuildDuration = 30 * time.Minute

// runDeploy waits for a build to complete, then stamps the new image tag
// into the GitOps repo and triggers an ArgoCD sync. Build progress is
// observable via buildjob.Get / buildjob.Logs; rollout progress is
// observable via the deployer and K8s. No in-memory tracking.
func (c *Client) runDeploy(claims *auth.Claims, serviceID platform.ServiceID, buildID string) {
	ctx, cancel := context.WithTimeout(auth.NewContext(context.Background(), claims), maxBuildDuration)
	defer cancel()

	ws := serviceID.Workspace
	project := serviceID.Project
	environment := serviceID.Environment
	service := serviceID.Name

	log := slog.With(
		"buildId", buildID,
		"project", project,
		"service", service,
		"environment", environment,
	)
	log.Info("deploy: waiting for build")

	deadline := time.Now().Add(maxBuildDuration)

	for time.Now().Before(deadline) {
		time.Sleep(2 * time.Second)

		job, err := c.buildjob.Get(ctx, buildID)

		if err != nil {
			log.Error("deploy: failed to poll build", "error", err)
			return
		}

		switch job.Status {
		case buildjob.StatusSucceeded:
			if len(job.ImageRefs) == 0 {
				log.Error("deploy: build succeeded but produced no image refs")
				return
			}

			tag, _ := imageParts(job.ImageRefs[0])

			log.Info("deploy: build succeeded, updating gitops", "tag", tag)

			if err := c.Packager.UpdateImageTag(ctx, ws, project, environment, service, tag, "", ""); err != nil {
				log.Error("deploy: failed to update image tag", "error", err)
				return
			}

			if _, err := c.Deployer.SyncDeployment(ctx, ws, project, environment); err != nil {
				log.Warn("deploy: failed to trigger sync (auto-sync will pick it up)", "error", err)
			}

			log.Info("deploy: complete")

			return

		case buildjob.StatusFailed, buildjob.StatusCancelled:
			log.Warn("deploy: build did not succeed", "status", string(job.Status))
			return
		}
	}

	log.Error("deploy: timed out waiting for build", "timeout", maxBuildDuration)
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

// GenerateDomain creates a platform domain for a service in an environment.
// Format: {service}-{env}-{randomSuffix}.{workloadDomain}.
func (c *Client) GenerateDomain(ctx context.Context, serviceID platform.ServiceID) (*Service, error) {
	projectID := serviceID.Project
	service := serviceID.Name
	environment := serviceID.Environment
	ws, err := tenant.FromContext(ctx)

	if err != nil {
		return nil, err
	}

	if _, err := c.Packager.GeneratePlatformDomain(ctx, ws, projectID, environment, service); err != nil {
		return nil, fmt.Errorf("failed to generate platform domain: %w", err)
	}

	if _, err := c.Deployer.SyncDeployment(ctx, ws, projectID, environment); err != nil {
		slog.Warn("failed to trigger sync after domain add", "project", projectID, "environment", environment, "error", err)
	}

	return c.Service(ctx, serviceID)
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
func (c *Client) AddCustomDomain(ctx context.Context, serviceID platform.ServiceID, hostname string) (*Service, error) {
	projectID := serviceID.Project
	service := serviceID.Name
	environment := serviceID.Environment
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
	if _, err = c.Deployer.ProvisionCustomDomain(ctx, hostname); err != nil {
		slog.Warn("failed to provision TLS certificate", "hostname", hostname, "error", err)
		return nil, err
	}

	// Trigger ArgoCD sync
	if _, err := c.Deployer.SyncDeployment(ctx, ws, projectID, environment); err != nil {
		slog.Warn("failed to trigger sync after domain add", "project", projectID, "environment", environment, "error", err)
	}

	return c.Service(ctx, serviceID)
}

// RemoveDomain removes a domain from a service in an environment.
func (c *Client) RemoveDomain(ctx context.Context, serviceID platform.ServiceID, hostname string) (*Service, error) {
	projectID := serviceID.Project
	service := serviceID.Name
	environment := serviceID.Environment
	ws, err := tenant.FromContext(ctx)

	if err != nil {
		return nil, err
	}

	if err := c.Packager.RemoveDomain(ctx, ws, projectID, environment, service, hostname); err != nil {
		return nil, fmt.Errorf("failed to remove domain: %w", err)
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

	return c.Service(ctx, serviceID)
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
