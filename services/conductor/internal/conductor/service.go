package conductor

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strings"

	containername "github.com/google/go-containerregistry/pkg/name"
	gh "github.com/google/go-github/v68/github"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/metrics"
	"github.com/zeitlos/lucity/services/conductor/internal/planner"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
	"github.com/zeitlos/lucity/services/conductor/internal/resources"
)

var (
	minServiceCPU        = resource.MustParse("250m")
	minServiceMemory     = resource.MustParse("256Mi")
	maxServiceCPU        = resource.MustParse("4")
	maxServiceMemory     = resource.MustParse("16Gi")
	defaultServiceCPU    = resources.DefaultCPULimit
	defaultServiceMemory = resources.DefaultMemoryLimit
)

func (c *Client) ServiceUsage(ctx context.Context, id platform.ServiceID, kinds []metrics.Kind, window metrics.Window, perReplica bool) ([]metrics.Series, error) {
	return c.metrics.ServiceUsage(ctx, id.Namespace(), id.Name, kinds, window, perReplica)
}

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

const defaultAutoDeployEnvironment = "development"

func (c *Client) Services(ctx context.Context, environmentID EnvironmentID) ([]Service, error) {
	return c.platform.Services(ctx, environmentID)
}

func (c *Client) Service(ctx context.Context, id ServiceID) (*Service, error) {
	return c.platform.Service(ctx, id)
}

func (c *Client) DetectServices(ctx context.Context, repositoryURL string) ([]Plan, error) {
	parsed, err := url.Parse(repositoryURL)

	if err != nil {
		return nil, fmt.Errorf("parse repository url %q: %w", repositoryURL, err)
	}

	repository := strings.TrimSuffix(strings.Trim(parsed.Path, "/"), ".git")

	if !repositoryPattern.MatchString(repository) {
		return nil, fmt.Errorf("invalid repository url %q: expected owner/repo path", repositoryURL)
	}

	if _, err := c.installationForRepo(ctx, repository); err != nil {
		return nil, err
	}

	commit, err := c.source.Commit(ctx, repositoryURL, "")

	if err != nil {
		return nil, err
	}

	token, err := c.source.Token(ctx, repositoryURL)

	if err != nil {
		return nil, err
	}

	return c.planner.Plan(ctx, repositoryURL, commit.SHA, token)
}

func (c *Client) RepositoryBranches(ctx context.Context, repositoryURL string) ([]string, error) {
	parsed, err := url.Parse(repositoryURL)

	if err != nil {
		return nil, fmt.Errorf("parse repository url %q: %w", repositoryURL, err)
	}

	repository := strings.TrimSuffix(strings.Trim(parsed.Path, "/"), ".git")

	if !repositoryPattern.MatchString(repository) {
		return nil, fmt.Errorf("invalid repository url %q: expected owner/repo path", repositoryURL)
	}

	if _, err := c.installationForRepo(ctx, repository); err != nil {
		return nil, err
	}

	return c.source.Branches(ctx, repositoryURL)
}

func (c *Client) AddService(ctx context.Context, environmentID platform.EnvironmentID, name, repository, contextPath string, externalImage string, variables map[string]string, cpu, memory string) (*Service, error) {
	if cpu == "" {
		cpu = defaultServiceCPU.String()
	}

	if memory == "" {
		memory = defaultServiceMemory.String()
	}

	resources, err := validateServiceResources(cpu, memory)

	if err != nil {
		return nil, err
	}

	workspace := environmentID.Workspace
	projectID := environmentID.Project
	id := platform.ServiceID{
		Workspace:   workspace,
		Project:     projectID,
		Environment: environmentID.Name,
		Name:        name,
	}

	environment, err := c.platform.Environment(ctx, environmentID)

	if err != nil {
		return nil, err
	}

	serviceName := name
	spec := deployer.ServiceSpec{
		ContextPath:  contextPath,
		ResourceTier: environment.ResourceTier,
		Env:          variables,
		Resources:    resources,
	}

	if repository != "" {
		spec.Port = 8080
		spec.AutoDeploy = environmentID.Name == defaultAutoDeployEnvironment

		installationID, err := c.installationForRepo(ctx, repository)

		if err != nil {
			return nil, err
		}

		spec.GitHubInstallationID = installationID
		spec.SourceURL, err = c.resolveRepositoryURL(ctx, installationID, repository)

		if err != nil {
			return nil, err
		}

		spec.Image = c.config.RegistryPullURL + "/" + id.ImageRepository()
	} else if externalImage != "" {
		if _, err := containername.ParseReference(externalImage); err != nil {
			return nil, fmt.Errorf("invalid image reference %q: %w", externalImage, err)
		}

		spec.Image = ensureImageTag(externalImage)
		spec.Port = c.imageExposedPort(ctx, spec.Image)

		if serviceName == "" {
			serviceName = deriveServiceName(externalImage)
		}
	} else {
		return nil, errors.New("either repository or external image must be set to create a new service")
	}

	if _, err := c.deployer.Services().Create(ctx, environmentID, serviceName, spec); err != nil {
		return nil, fmt.Errorf("create service: %w", err)
	}

	service := &platform.Service{
		ID:        id,
		Name:      serviceName,
		Variables: variables,
	}

	if spec.SourceURL != "" {
		commit, err := c.source.Commit(ctx, spec.SourceURL, "")

		if err != nil {
			slog.Warn("initial commit lookup failed", "project", projectID, "service", name, "error", err)
			return service, nil
		}

		token, err := c.source.Token(ctx, spec.SourceURL)

		if err != nil {
			slog.Warn("source token mint failed", "project", projectID, "service", name, "error", err)
			return service, nil
		}

		claims, _ := auth.FromContext(ctx)
		release := deployer.NewRelease(deployer.TriggerManual, actorFromClaims(claims))

		imageName := id.ImageRepository()

		build, err := c.buildjob.Start(ctx, buildjob.StartOptions{
			Service:          id,
			RepoURL:          spec.SourceURL,
			Commit:           commit.SHA,
			CommitMessage:    commit.Message,
			ContextPath:      contextPath,
			TargetImageNames: []string{imageName},
			Token:            token,
			BuildVars:        service.Variables,
			ReleaseID:        release.ID,
		})

		if err != nil {
			slog.Warn("initial build start failed", "project", projectID, "service", name, "error", err)
			return service, nil
		}

		if _, err := c.startDeploy(ctx, service.ID, build.ID.Name, commit.Message, release); err != nil {
			slog.Warn("initial deploy start failed", "project", projectID, "service", name, "error", err)
		}

		c.startScan(ctx, service.ID, build.ID.Name, spec.SourceURL, commit.SHA, token, release.ID)
	}

	return service, nil
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

func (c *Client) SetServiceBranch(ctx context.Context, svc platform.ServiceID, branch string) (*Service, error) {
	if _, err := c.deployer.Services().SetBranch(ctx, svc, branch); err != nil {
		return nil, fmt.Errorf("set branch: %w", err)
	}

	return c.Service(ctx, svc)
}

func (c *Client) SetAutoDeploy(ctx context.Context, svc platform.ServiceID, enabled bool) (*Service, error) {
	if _, err := c.deployer.Services().SetAutoDeploy(ctx, svc, enabled); err != nil {
		return nil, fmt.Errorf("set autodeploy: %w", err)
	}

	return c.Service(ctx, svc)
}

func (c *Client) SetCIDeploy(ctx context.Context, svc platform.ServiceID, enabled bool) (*Service, error) {
	if _, err := c.deployer.Services().SetCIDeploy(ctx, svc, enabled); err != nil {
		return nil, fmt.Errorf("set ci-deploy: %w", err)
	}

	return c.Service(ctx, svc)
}

func (c *Client) SetServiceResources(ctx context.Context, service platform.ServiceID, cpu, memory string) (*Service, error) {
	spec, err := validateServiceResources(cpu, memory)

	if err != nil {
		return nil, err
	}

	environment, err := c.platform.Environment(ctx, service.EnvironmentID())

	if err != nil {
		return nil, err
	}

	if _, err := c.deployer.Services().SetResources(ctx, service, environment.ResourceTier, spec); err != nil {
		return nil, fmt.Errorf("set resources: %w", err)
	}

	return c.Service(ctx, service)
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

	provenance := deployer.ImageProvenance{
		Commit:        deployment.Commit,
		CommitMessage: deployment.CommitMessage,
		BuildID:       deployment.BuildID,
	}

	claims, _ := auth.FromContext(ctx)
	release := deployer.NewRelease(deployer.TriggerRollback, actorFromClaims(claims))

	if _, err := c.deployer.Services().SetImage(ctx, serviceID, deployment.Image, provenance, release); err != nil {
		return false, fmt.Errorf("rollback set image: %w", err)
	}

	return true, nil
}

func (c *Client) GenerateDomain(ctx context.Context, serviceID platform.ServiceID) (*Service, error) {
	hostname := fmt.Sprintf("%s-%s-%s.%s",
		serviceID.Name,
		serviceID.Environment,
		randCrockford32(5),
		c.config.WorkloadDomain,
	)

	if _, err := c.deployer.Services().AddDomain(ctx, serviceID, hostname); err != nil {
		return nil, fmt.Errorf("add platform domain: %w", err)
	}

	if _, err := c.deployer.Services().VerifyDomain(ctx, serviceID, hostname, true); err != nil {
		return nil, fmt.Errorf("verify platform domain: %w", err)
	}

	return c.Service(ctx, serviceID)
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

func (c *Client) ReconcileServices(ctx context.Context) error {
	workspaces, err := c.directory.Workspaces(ctx)

	if err != nil {
		return err
	}

	for _, workspace := range workspaces {
		environments, err := c.platform.EnvironmentsByWorkspace(ctx, workspace.ID)

		if err != nil {
			slog.Warn("reconcile services: failed to list projects", "error", err, "workspace", workspace.ID)
			continue
		}

		for _, environment := range environments {
			if _, err := c.deployer.Environments().Reconcile(ctx, environment.ID); err != nil {
				slog.Warn("reconcile services: failed to reconcile", "error", err, "environment", environment)
			}
		}
	}

	return nil
}

func validateServiceResources(cpu, memory string) (deployer.Resources, error) {
	return validateResources(cpu, memory, minServiceCPU, maxServiceCPU, minServiceMemory, maxServiceMemory)
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

var repositoryPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`)

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

	token, err := c.gitHubApp.InstallationToken(ctx, installationID)

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

func ensureImageTag(ref string) string {
	lastComponent := ref[strings.LastIndex(ref, "/")+1:]

	if strings.ContainsAny(lastComponent, ":@") {
		return ref
	}

	return ref + ":latest"
}
