package handler

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"

	gh "github.com/google/go-github/v68/github"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/builder"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/pkg/packager"
)

// gRPC call timeouts. Short for quick lookups, long for operations that
// clone repos or touch external systems.
const (
	grpcTimeout     = 10 * time.Second
	grpcLongTimeout = 60 * time.Second
)

type Project struct {
	ID             string
	Name           string
	Environments   []Environment
	Services       []Service
	Databases      []Database
	InitialDeploys []DeployOp
	CreatedAt      time.Time
}

type Environment struct {
	ID         string
	Name       string
	Namespace  string
	Ephemeral  bool
	SyncStatus string
	Services   []ServiceInstance
	Databases  []DatabaseInstance
}

type Service struct {
	Name        string
	Image       string
	Port        int
	Framework   string
	SourceURL   string
	ContextPath string
	Instances   []ServiceInstance
}

type ServiceInstance struct {
	Name        string
	Environment string
	ImageTag    string
	Ready       bool
	Replicas    int
	Domains     []string
	Deployments []Deployment
}

type Deployment struct {
	ID                  string
	ImageTag            string
	Active              bool
	Timestamp           time.Time
	Revision            string
	Message             string
	SourceCommitMessage string
	SourceURL           string // full URL to commit on GitHub
}

func (c *Client) Projects(ctx context.Context) ([]Project, error) {
	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.Packager.ListProjects(callCtx, &packager.ListProjectsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	result := make([]Project, 0, len(resp.Projects))
	for _, p := range resp.Projects {
		proj := projectFromProto(p)
		c.enrichSyncStatus(ctx, &proj)
		c.enrichDatabaseStatus(ctx, &proj)
		c.enrichDeploymentHistory(ctx, &proj)
		c.enrichCommitMessages(ctx, &proj)
		result = append(result, proj)
	}
	return result, nil
}

func (c *Client) Project(ctx context.Context, id string) (*Project, error) {
	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.Packager.GetProject(callCtx, &packager.GetProjectRequest{Project: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	p := projectFromProto(resp.Project)
	c.enrichSyncStatus(ctx, &p)
	c.enrichDatabaseStatus(ctx, &p)
	c.enrichDeploymentHistory(ctx, &p)
	c.enrichCommitMessages(ctx, &p)
	return &p, nil
}

func (c *Client) CreateProject(ctx context.Context, name string) (*Project, error) {
	ctx = auth.OutgoingContext(ctx)

	// 1. Create GitOps repo
	initCtx, initCancel := context.WithTimeout(ctx, grpcLongTimeout)
	defer initCancel()
	resp, err := c.Packager.InitProject(initCtx, &packager.InitProjectRequest{
		Project: name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// 2. Deploy the default development environment via ArgoCD
	ns := labels.NamespaceFor(name, "development")
	deployCtx, deployCancel := context.WithTimeout(ctx, grpcTimeout)
	defer deployCancel()
	_, err = c.Deployer.DeployEnvironment(deployCtx, &deployer.DeployEnvironmentRequest{
		Project:         name,
		Environment:     "development",
		GitopsRepoUrl:   resp.GitopsRepoUrl,
		TargetNamespace: ns,
	})
	if err != nil {
		slog.Warn("failed to deploy development environment", "project", name, "error", err)
	}

	return &Project{
		ID:        name,
		Name:      name,
		CreatedAt: time.Now(),
		Environments: []Environment{
			{
				ID:         name + "/development",
				Name:       "development",
				Namespace:  ns,
				SyncStatus: "PROGRESSING",
			},
		},
	}, nil
}

func (c *Client) DeleteProject(ctx context.Context, id string) (bool, error) {
	ctx = auth.OutgoingContext(ctx)

	// 1. Fetch project to discover all environments
	getCtx, getCancel := context.WithTimeout(ctx, grpcTimeout)
	defer getCancel()
	resp, err := c.Packager.GetProject(getCtx, &packager.GetProjectRequest{Project: id})
	if err != nil {
		return false, fmt.Errorf("failed to get project for deletion: %w", err)
	}

	// 2. Remove ArgoCD Application for each environment (best-effort)
	for _, env := range resp.Project.Environments {
		rmCtx, rmCancel := context.WithTimeout(ctx, grpcTimeout)
		_, err := c.Deployer.RemoveDeployment(rmCtx, &deployer.RemoveDeploymentRequest{
			Project:     id,
			Environment: env,
		})
		rmCancel()
		if err != nil {
			slog.Warn("failed to remove deployment during project deletion",
				"project", id, "environment", env, "error", err)
		}
	}

	// 3. Remove ArgoCD repository credential (best-effort)
	repoCtx, repoCancel := context.WithTimeout(ctx, grpcTimeout)
	defer repoCancel()
	_, err = c.Deployer.DeleteRepository(repoCtx, &deployer.DeleteRepositoryRequest{
		Project: id,
	})
	if err != nil {
		slog.Warn("failed to delete ArgoCD repository credential",
			"project", id, "error", err)
	}

	// 4. Delete OCI images from registry (best-effort)
	imgCtx, imgCancel := context.WithTimeout(ctx, grpcTimeout)
	defer imgCancel()
	_, err = c.Builder.DeleteImages(imgCtx, &builder.DeleteImagesRequest{
		Project: id,
	})
	if err != nil {
		slog.Warn("failed to delete registry images",
			"project", id, "error", err)
	}

	// 5. Delete GitOps repo
	delCtx, delCancel := context.WithTimeout(ctx, grpcTimeout)
	defer delCancel()
	_, err = c.Packager.DeleteProject(delCtx, &packager.DeleteProjectRequest{Project: id})
	if err != nil {
		return false, fmt.Errorf("failed to delete project: %w", err)
	}
	return true, nil
}

func (c *Client) CreateEnvironment(ctx context.Context, projectID, name, fromEnvironment string) (*Environment, error) {
	ctx = auth.OutgoingContext(ctx)

	createCtx, createCancel := context.WithTimeout(ctx, grpcLongTimeout)
	defer createCancel()
	resp, err := c.Packager.CreateEnvironment(createCtx, &packager.CreateEnvironmentRequest{
		Project:         projectID,
		Environment:     name,
		FromEnvironment: fromEnvironment,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	// Deploy the new environment via ArgoCD
	envDeployCtx, envDeployCancel := context.WithTimeout(ctx, grpcTimeout)
	defer envDeployCancel()
	_, err = c.Deployer.DeployEnvironment(envDeployCtx, &deployer.DeployEnvironmentRequest{
		Project:         projectID,
		Environment:     name,
		TargetNamespace: resp.Namespace,
	})
	if err != nil {
		slog.Warn("failed to deploy environment", "project", projectID, "environment", name, "error", err)
	}

	return &Environment{
		ID:         projectID + "/" + name,
		Name:       name,
		Namespace:  resp.Namespace,
		SyncStatus: "PROGRESSING",
	}, nil
}

func (c *Client) DeleteEnvironment(ctx context.Context, projectID, environment string) (bool, error) {
	ctx = auth.OutgoingContext(ctx)

	// Remove ArgoCD Application first (cascade deletes managed resources)
	rmCtx, rmCancel := context.WithTimeout(ctx, grpcTimeout)
	defer rmCancel()
	_, err := c.Deployer.RemoveDeployment(rmCtx, &deployer.RemoveDeploymentRequest{
		Project:     projectID,
		Environment: environment,
	})
	if err != nil {
		slog.Warn("failed to remove deployment", "project", projectID, "environment", environment, "error", err)
	}

	// Then remove from GitOps repo
	delEnvCtx, delEnvCancel := context.WithTimeout(ctx, grpcTimeout)
	defer delEnvCancel()
	_, err = c.Packager.DeleteEnvironment(delEnvCtx, &packager.DeleteEnvironmentRequest{
		Project:     projectID,
		Environment: environment,
	})
	if err != nil {
		return false, fmt.Errorf("failed to delete environment: %w", err)
	}
	return true, nil
}

func (c *Client) Promote(ctx context.Context, projectID, service, fromEnv, toEnv string) (*ServiceInstance, error) {
	ctx = auth.OutgoingContext(ctx)

	promoteCtx, promoteCancel := context.WithTimeout(ctx, grpcLongTimeout)
	defer promoteCancel()
	resp, err := c.Packager.Promote(promoteCtx, &packager.PromoteRequest{
		Project:         projectID,
		Service:         service,
		FromEnvironment: fromEnv,
		ToEnvironment:   toEnv,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to promote: %w", err)
	}

	return &ServiceInstance{
		Name:        service,
		Environment: toEnv,
		ImageTag:    resp.ImageTag,
	}, nil
}

func (c *Client) SyncChart(ctx context.Context, projectID string) (bool, error) {
	ctx = auth.OutgoingContext(ctx)

	callCtx, callCancel := context.WithTimeout(ctx, grpcLongTimeout)
	defer callCancel()
	_, err := c.Packager.SyncChart(callCtx, &packager.SyncChartRequest{Project: projectID})
	if err != nil {
		return false, fmt.Errorf("failed to sync chart: %w", err)
	}
	return true, nil
}

func (c *Client) Service(ctx context.Context, projectID, name string) (*Service, error) {
	proj, err := c.Project(ctx, projectID)
	if err != nil {
		return nil, err
	}

	for _, svc := range proj.Services {
		if svc.Name == name {
			return &svc, nil
		}
	}
	return nil, nil
}

// enrichSyncStatus queries the deployer for each environment's ArgoCD sync status.
// Best-effort: logs warnings on failure and leaves status as "UNKNOWN".
// Calls are made concurrently to avoid serial N+1 latency.
func (c *Client) enrichSyncStatus(ctx context.Context, proj *Project) {
	var wg sync.WaitGroup
	for i := range proj.Environments {
		env := &proj.Environments[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			statusCtx, statusCancel := context.WithTimeout(ctx, grpcTimeout)
			defer statusCancel()
			resp, err := c.Deployer.GetDeploymentStatus(statusCtx, &deployer.GetDeploymentStatusRequest{
				Project:     proj.ID,
				Environment: env.Name,
			})
			if err != nil {
				slog.Debug("failed to get deployment status", "project", proj.ID, "environment", env.Name, "error", err)
				return
			}
			env.SyncStatus = deploymentStatusToString(resp.Status)

			// Derive per-service readiness from environment health.
			ready := resp.Status == deployer.DeploymentStatus_DEPLOYMENT_STATUS_SYNCED
			for j := range env.Services {
				env.Services[j].Ready = ready
			}
		}()
	}
	wg.Wait()
}

// enrichDatabaseStatus queries the deployer for each database's runtime status
// per environment. Best-effort: logs warnings on failure.
func (c *Client) enrichDatabaseStatus(ctx context.Context, proj *Project) {
	if len(proj.Databases) == 0 {
		return
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := range proj.Environments {
		env := &proj.Environments[i]
		for _, db := range proj.Databases {
			wg.Add(1)
			go func(envPtr *Environment, dbInfo Database) {
				defer wg.Done()
				statusCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
				defer cancel()
				resp, err := c.Deployer.DatabaseStatus(statusCtx, &deployer.DatabaseStatusRequest{
					Project:     proj.ID,
					Environment: envPtr.Name,
					Database:    dbInfo.Name,
				})
				inst := DatabaseInstance{
					Name:        dbInfo.Name,
					Environment: envPtr.Name,
					Version:     dbInfo.Version,
					Size:        dbInfo.Size,
					Instances:   dbInfo.Instances,
				}
				if err != nil {
					slog.Debug("failed to get database status", "project", proj.ID, "environment", envPtr.Name, "database", dbInfo.Name, "error", err)
				} else {
					inst.Ready = resp.Ready
					if resp.Instances > 0 {
						inst.Instances = int(resp.Instances)
					}
					if resp.Volume != nil {
						inst.Volume = &Volume{
							Name:          resp.Volume.Name,
							Size:          resp.Volume.Size,
							RequestedSize: resp.Volume.RequestedSize,
							UsedBytes:     resp.Volume.UsedBytes,
							CapacityBytes: resp.Volume.CapacityBytes,
						}
					}
				}
				mu.Lock()
				envPtr.Databases = append(envPtr.Databases, inst)
				mu.Unlock()
			}(env, db)
		}
	}
	wg.Wait()
}

// enrichDeploymentHistory fetches deployment history from the packager for each
// service instance in every environment and attaches it.
// Calls are made concurrently — each goroutine writes to its own ServiceInstance.
func (c *Client) enrichDeploymentHistory(ctx context.Context, proj *Project) {
	var wg sync.WaitGroup
	for i := range proj.Environments {
		env := &proj.Environments[i]
		for j := range env.Services {
			si := &env.Services[j]
			wg.Add(1)
			go func() {
				defer wg.Done()
				histCtx, histCancel := context.WithTimeout(ctx, grpcTimeout)
				defer histCancel()
				resp, err := c.Packager.DeploymentHistory(histCtx, &packager.DeploymentHistoryRequest{
					Project:     proj.ID,
					Environment: env.Name,
					Service:     si.Name,
				})
				if err != nil {
					slog.Debug("failed to get deployment history", "project", proj.ID, "environment", env.Name, "service", si.Name, "error", err)
					return
				}

				for k, e := range resp.Entries {
					deployedAt := e.DeployedAt.AsTime()
					si.Deployments = append(si.Deployments, Deployment{
						ID:        fmt.Sprintf("%s/%s/%s/%d", proj.ID, env.Name, si.Name, k),
						ImageTag:  e.ImageTag,
						Active:    k == 0,
						Timestamp: deployedAt,
						Revision:  e.Revision,
						Message:   fmt.Sprintf("deploy(%s): %s %s", env.Name, si.Name, e.ImageTag),
					})
				}
			}()
		}
	}
	wg.Wait()

	// Also attach to the cross-referenced Service.Instances
	for i, svc := range proj.Services {
		for j := range svc.Instances {
			inst := &proj.Services[i].Instances[j]
			for _, env := range proj.Environments {
				for _, esi := range env.Services {
					if esi.Name == inst.Name && esi.Environment == inst.Environment {
						inst.Deployments = esi.Deployments
						inst.Ready = esi.Ready
					}
				}
			}
		}
	}
}

// shaPattern matches a hex string of 7+ characters (git short SHA).
var shaPattern = regexp.MustCompile(`^[0-9a-f]{7,}$`)

// enrichCommitMessages fetches source commit messages from GitHub for
// deployment entries whose imageTag is a git SHA. Best-effort — failures
// are silently ignored. Also sets SourceURL for each SHA-based deployment.
func (c *Client) enrichCommitMessages(ctx context.Context, proj *Project) {
	claims := auth.FromContext(ctx)
	if claims == nil || claims.GitHubToken == "" {
		return
	}

	// Build service name → sourceURL lookup
	sourceURLs := make(map[string]string, len(proj.Services))
	for _, svc := range proj.Services {
		if svc.SourceURL != "" {
			sourceURLs[svc.Name] = svc.SourceURL
		}
	}

	// Collect unique (owner/repo, sha) pairs that need fetching
	type commitKey struct{ owner, repo, sha string }
	type commitResult struct{ message, url string }

	needed := make(map[commitKey]bool)
	for _, env := range proj.Environments {
		for _, si := range env.Services {
			srcURL := sourceURLs[si.Name]
			if srcURL == "" {
				continue
			}
			owner, repo := parseGitHubRepoURL(srcURL)
			if owner == "" {
				continue
			}
			for _, dep := range si.Deployments {
				if shaPattern.MatchString(dep.ImageTag) {
					needed[commitKey{owner, repo, dep.ImageTag}] = true
				}
			}
		}
	}

	if len(needed) == 0 {
		return
	}

	// Fetch commit messages concurrently
	client := gh.NewClient(nil).WithAuthToken(claims.GitHubToken)
	results := make(map[commitKey]commitResult, len(needed))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for key := range needed {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fetchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			commit, _, err := client.Repositories.GetCommit(fetchCtx, key.owner, key.repo, key.sha, nil)
			if err != nil {
				slog.Debug("failed to fetch commit message", "owner", key.owner, "repo", key.repo, "sha", key.sha, "error", err)
				return
			}

			msg := commit.GetCommit().GetMessage()
			// Use first line only
			if idx := strings.IndexByte(msg, '\n'); idx >= 0 {
				msg = msg[:idx]
			}

			mu.Lock()
			results[key] = commitResult{
				message: msg,
				url:     fmt.Sprintf("https://github.com/%s/%s/commit/%s", key.owner, key.repo, key.sha),
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	// Apply results to all deployments
	for i := range proj.Environments {
		env := &proj.Environments[i]
		for j := range env.Services {
			si := &env.Services[j]
			srcURL := sourceURLs[si.Name]
			if srcURL == "" {
				continue
			}
			owner, repo := parseGitHubRepoURL(srcURL)
			if owner == "" {
				continue
			}
			for k := range si.Deployments {
				dep := &si.Deployments[k]
				if r, ok := results[commitKey{owner, repo, dep.ImageTag}]; ok {
					dep.SourceCommitMessage = r.message
					dep.SourceURL = r.url
				}
			}
		}
	}

	// Also update the cross-referenced Service.Instances
	for i := range proj.Services {
		for j := range proj.Services[i].Instances {
			inst := &proj.Services[i].Instances[j]
			srcURL := sourceURLs[inst.Name]
			if srcURL == "" {
				continue
			}
			owner, repo := parseGitHubRepoURL(srcURL)
			if owner == "" {
				continue
			}
			for k := range inst.Deployments {
				dep := &inst.Deployments[k]
				if r, ok := results[commitKey{owner, repo, dep.ImageTag}]; ok {
					dep.SourceCommitMessage = r.message
					dep.SourceURL = r.url
				}
			}
		}
	}
}

// parseGitHubRepoURL extracts owner and repo name from a GitHub URL.
// Supports "https://github.com/owner/repo" and "https://github.com/owner/repo.git".
func parseGitHubRepoURL(rawURL string) (owner, repo string) {
	// Strip protocol and host
	idx := strings.Index(rawURL, "github.com/")
	if idx < 0 {
		return "", ""
	}
	path := rawURL[idx+len("github.com/"):]
	path = strings.TrimSuffix(path, ".git")
	path = strings.TrimSuffix(path, "/")

	parts := strings.SplitN(path, "/", 3)
	if len(parts) < 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func deploymentStatusToString(status deployer.DeploymentStatus) string {
	switch status {
	case deployer.DeploymentStatus_DEPLOYMENT_STATUS_SYNCED:
		return "SYNCED"
	case deployer.DeploymentStatus_DEPLOYMENT_STATUS_OUT_OF_SYNC:
		return "OUT_OF_SYNC"
	case deployer.DeploymentStatus_DEPLOYMENT_STATUS_PROGRESSING:
		return "PROGRESSING"
	case deployer.DeploymentStatus_DEPLOYMENT_STATUS_DEGRADED:
		return "DEGRADED"
	default:
		return "UNKNOWN"
	}
}

func projectFromProto(p *packager.ProjectInfo) Project {
	createdAt := p.CreatedAt.AsTime()

	proj := Project{
		ID:        p.Name,
		Name:      p.Name,
		CreatedAt: createdAt,
	}

	// Build a lookup of per-env service info from the richer EnvironmentInfos.
	envInfoMap := make(map[string][]*packager.ServiceInstanceInfo, len(p.EnvironmentInfos))
	for _, ei := range p.EnvironmentInfos {
		envInfoMap[ei.Name] = ei.Services
	}

	for _, envName := range p.Environments {
		env := Environment{
			ID:         p.Name + "/" + envName,
			Name:       envName,
			Namespace:  labels.NamespaceFor(p.Name, envName),
			SyncStatus: "UNKNOWN",
		}

		// Populate service instances from environment values
		for _, svc := range envInfoMap[envName] {
			env.Services = append(env.Services, ServiceInstance{
				Name:        svc.Name,
				Environment: envName,
				ImageTag:    svc.ImageTag,
				Domains:     svc.Domains,
				Replicas:    1, // default until we query K8s
			})
		}

		proj.Environments = append(proj.Environments, env)
	}

	for _, svc := range p.Services {
		proj.Services = append(proj.Services, Service{
			Name:        svc.Name,
			Image:       svc.Image,
			Port:        int(svc.Port),
			Framework:   svc.Framework,
			SourceURL:   svc.SourceUrl,
			ContextPath: svc.ContextPath,
		})
	}

	for _, db := range p.Databases {
		proj.Databases = append(proj.Databases, Database{
			Name:      db.Name,
			Version:   db.Version,
			Instances: int(db.Instances),
			Size:      db.Size,
		})
	}

	// Cross-reference: collect all service instances across environments onto each Service
	for i, svc := range proj.Services {
		for _, env := range proj.Environments {
			for _, si := range env.Services {
				if si.Name == svc.Name {
					proj.Services[i].Instances = append(proj.Services[i].Instances, si)
				}
			}
		}
	}

	return proj
}
