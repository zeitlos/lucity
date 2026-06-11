package conductor

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	gh "github.com/google/go-github/v68/github"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/tenant"
	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/planner"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

func randCrockford32(n int) string {
	const alphabet = "0123456789abcdefghjkmnpqrstvwxyz"

	b := make([]byte, n)
	_, _ = rand.Read(b)

	for i, v := range b {
		b[i] = alphabet[v&31]
	}

	return string(b)
}

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

func (c *Client) AddService(ctx context.Context, environment platform.EnvironmentID, name string, repository, contextPath string, installationID *int64, externalImage string) (*Service, error) {
	ws := environment.Workspace
	projectID := environment.Project
	envName := environment.Name
	id := platform.ServiceID{
		Workspace:   ws,
		Project:     projectID,
		Environment: envName,
		Name:        name,
	}

	var err error

	serviceName := name
	spec := deployer.ServiceSpec{
		ContextPath: contextPath,
		Port:        8080,
	}

	if repository != "" {
		if installationID == nil {
			return nil, fmt.Errorf("installationId is required when repository is set")
		}

		spec.GitHubInstallationID = *installationID
		spec.SourceURL, err = c.resolveRepositoryURL(ctx, *installationID, repository)

		if err != nil {
			return nil, err
		}

		spec.Image = c.Config.RegistryPullURL + "/" + c.imageRepository(id)

		// ctx, err = c.withInstallationTokenForID(ctx, *installationID)

		// if err != nil {
		// 	return nil, fmt.Errorf("github auth: %w", err)
		// }
	} else if externalImage != "" {
		spec.Image = externalImage

		if serviceName == "" {
			serviceName = deriveServiceName(externalImage)
		}
	} else {
		return nil, errors.New("either repository or external image must be set to create a new service")
	}

	if _, err := c.deployer.Services().Create(ctx, environment, name, spec); err != nil {
		return nil, fmt.Errorf("create service: %w", err)
	}

	service := &platform.Service{
		ID:   id,
		Name: name,
	}

	if spec.SourceURL != "" {
		commit, err := c.source.CommitSHA(ctx, spec.SourceURL, "")

		if err != nil {
			slog.Warn("initial commit lookup failed", "project", projectID, "service", name, "error", err)
			return service, nil
		}

		token, err := c.source.Token(ctx, spec.SourceURL)

		if err != nil {
			slog.Warn("source token mint failed", "project", projectID, "service", name, "error", err)
			return service, nil
		}

		imageName := ws + "/" + projectID + "/" + name

		build, err := c.buildjob.Start(ctx, buildjob.StartOptions{
			Workspace:        ws,
			RepoURL:          spec.SourceURL,
			Commit:           commit,
			ContextPath:      contextPath,
			TargetImageNames: []string{imageName},
			Token:            token,
		})

		if err != nil {
			slog.Warn("initial build start failed", "project", projectID, "service", name, "error", err)
			return service, nil
		}

		claims, _ := auth.FromContext(ctx)

		go c.runDeploy(claims, service.ID, build.ID)
	}

	return service, nil
}

// deriveServiceName extracts a service name from an image reference.
// e.g., "nginx:1.25" → "nginx", "ghcr.io/foo/my-app:v1" → "my-app"
func deriveServiceName(imageRef string) string {
	name := imageRef

	if i := strings.LastIndex(name, ":"); i >= 0 {
		if j := strings.LastIndex(name, "/"); i > j {
			name = name[:i]
		}
	}

	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}

	return name
}

func (c *Client) RemoveService(ctx context.Context, svc platform.ServiceID) (bool, error) {
	if err := c.deployer.Services().Delete(ctx, svc); err != nil {
		return false, fmt.Errorf("delete service: %w", err)
	}

	return true, nil
}

func (c *Client) SetCustomStartCommand(ctx context.Context, svc platform.ServiceID, command string) (*Service, error) {
	if _, err := c.deployer.Services().SetCommand(ctx, svc, command); err != nil {
		return nil, fmt.Errorf("set command: %w", err)
	}

	return c.Service(ctx, svc)
}

func (c *Client) SetServicePort(ctx context.Context, svc platform.ServiceID, port int) (*Service, error) {
	if _, err := c.deployer.Services().SetPort(ctx, svc, port); err != nil {
		return nil, fmt.Errorf("set port: %w", err)
	}

	return c.Service(ctx, svc)
}

// withInstallationTokenForID mints a GitHub App installation token for the given
// installation ID and attaches it to the context for downstream gRPC calls.
func (c *Client) withInstallationTokenForID(ctx context.Context, installationID int64) (context.Context, error) {
	if installationID == 0 {
		return ctx, nil
	}

	token, err := c.GitHubApp.InstallationToken(ctx, installationID)

	if err != nil {
		return ctx, fmt.Errorf("mint installation token: %w", err)
	}

	return auth.WithGitHubToken(ctx, token), nil
}

func (c *Client) imageRepository(id ServiceID) string {
	return id.Workspace + "/" + id.Project + "/" + id.Name
}

const maxBuildDuration = 30 * time.Minute

// runDeploy waits for a build to complete, then stamps the new image onto
// the service via the deployer (single helm upgrade applies both the values
// change and the K8s deployment patch — no separate sync step needed).
func (c *Client) runDeploy(claims *auth.Claims, serviceID platform.ServiceID, buildID string) {
	ctx, cancel := context.WithTimeout(auth.NewContext(context.Background(), claims), maxBuildDuration)
	defer cancel()

	log := slog.With(
		"buildId", buildID,
		"project", serviceID.Project,
		"service", serviceID.Name,
		"environment", serviceID.Environment,
	)
	log.InfoContext(ctx, "deploy: waiting for build")

	deadline := time.Now().Add(maxBuildDuration)

	for time.Now().Before(deadline) {
		time.Sleep(2 * time.Second)

		job, err := c.buildjob.Get(ctx, buildID)

		if err != nil {
			log.ErrorContext(ctx, "deploy: build poll failed", "error", err)
			return
		}

		switch job.Status {
		case buildjob.StatusSucceeded:
			built, err := job.ImageRef(c.imageRepository(serviceID))

			if err != nil {
				log.ErrorContext(ctx, "deploy: failed to get image ref", "error", err)
				return
			}

			ref := c.Config.RegistryPullURL + "/" + built.Context().RepositoryStr() + tagOrDigest(built)

			log.InfoContext(ctx, "deploy: build succeeded, applying image", "ref", ref)

			if _, err := c.deployer.Services().SetImage(ctx, serviceID, ref, ""); err != nil {
				log.ErrorContext(ctx, "deploy: set image failed", "error", err)
				return
			}

			log.InfoContext(ctx, "deploy: complete")

			return

		case buildjob.StatusFailed, buildjob.StatusCancelled:
			log.WarnContext(ctx, "deploy: build did not succeed", "status", string(job.Status))
			return
		}
	}

	log.ErrorContext(ctx, "deploy: timed out waiting for build", "timeout", maxBuildDuration)
}

func (c *Client) Rollback(ctx context.Context, deploymentID DeploymentID) (bool, error) {
	deployment, err := c.platform.Deployment(ctx, deploymentID)

	if err != nil {
		return false, fmt.Errorf("read deployment: %w", err)
	}

	serviceID := platform.ServiceID{
		Workspace:   deploymentID.Workspace,
		Project:     deploymentID.Project,
		Environment: deploymentID.Environment,
		Name:        deploymentID.Service,
	}

	if _, err := c.deployer.Services().SetImage(ctx, serviceID, deployment.Image, ""); err != nil {
		return false, fmt.Errorf("rollback set image: %w", err)
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

// IsPlatformDomain checks if a hostname is a platform-generated domain.
func (c *Client) IsPlatformDomain(hostname string) bool {
	return strings.HasSuffix(hostname, "."+c.Config.WorkloadDomain)
}

// GenerateDomain creates a platform domain ({service}-{env}-{rand}.{workloadDomain})
// for a service and immediately marks it verified — platform DNS is under
// our control, so there's no challenge to run. Returns the updated service.
func (c *Client) GenerateDomain(ctx context.Context, serviceID platform.ServiceID) (*Service, error) {
	hostname := fmt.Sprintf("%s-%s-%s.%s",
		serviceID.Name,
		serviceID.Environment,
		randCrockford32(5),
		c.Config.WorkloadDomain,
	)

	if _, err := c.deployer.Services().AddDomain(ctx, serviceID, hostname); err != nil {
		return nil, fmt.Errorf("add platform domain: %w", err)
	}

	if _, err := c.deployer.Services().VerifyDomain(ctx, serviceID, hostname, true); err != nil {
		return nil, fmt.Errorf("verify platform domain: %w", err)
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
		return "", fmt.Errorf("authenticate with GitHub: %w", err)
	}

	client := gh.NewClient(nil).WithAuthToken(token)
	ghRepo, _, err := client.Repositories.Get(ctx, owner, repo)

	if err != nil {
		return "", fmt.Errorf("repository %q not accessible via this GitHub App installation: %w", repository, err)
	}

	cloneURL := ghRepo.GetCloneURL()

	if cloneURL == "" {
		cloneURL = "https://github.com/" + owner + "/" + repo
	}

	return cloneURL, nil
}

// validateHostname checks that a hostname is a valid domain name.
func validateHostname(hostname string) error {
	for _, prefix := range []string{"https://", "http://", "www."} {
		hostname = strings.TrimPrefix(hostname, prefix)
	}

	hostname = strings.TrimSuffix(hostname, ".")
	hostname = strings.TrimRight(hostname, "/")

	if len(hostname) < 4 || len(hostname) > 253 {
		return fmt.Errorf("hostname must be between 4 and 253 characters")
	}

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

func (c *Client) AddCustomDomain(ctx context.Context, serviceID platform.ServiceID, hostname string) (*Service, error) {
	if err := validateHostname(hostname); err != nil {
		return nil, fmt.Errorf("invalid hostname: %w", err)
	}

	if c.hostname.IsPlatform(hostname) || c.hostname.IsInternal(hostname) {
		return nil, fmt.Errorf("invalid domain")
	}

	if _, err := c.deployer.Services().AddDomain(ctx, serviceID, hostname); err != nil {
		return nil, err
	}

	verified, err := c.isDomainVerified(ctx, serviceID.Workspace, hostname)

	if err != nil {
		return nil, err
	}

	if verified {
		if _, err := c.deployer.Services().VerifyDomain(ctx, serviceID, hostname, verified); err != nil {
			// This will be re-tried by the reconcile loop, therefore we don't surface the error.
			slog.WarnContext(ctx, "failed to set domain to verified", "error", err, "service", serviceID, "domain", hostname)
		}
	}

	return c.Service(ctx, serviceID)
}

func (c *Client) RemoveDomain(ctx context.Context, serviceID platform.ServiceID, hostname string) (*Service, error) {
	if _, err := c.deployer.Services().RemoveDomain(ctx, serviceID, hostname); err != nil {
		return nil, err
	}

	return c.Service(ctx, serviceID)
}

func tagOrDigest(ref name.Reference) string {
	switch v := ref.(type) {
	case name.Tag:
		return ":" + v.TagStr()
	case name.Digest:
		return "@" + v.DigestStr()
	}

	return ""
}
