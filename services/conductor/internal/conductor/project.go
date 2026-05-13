package conductor

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"

	gh "github.com/google/go-github/v68/github"

	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/pkg/tenant"
	"github.com/zeitlos/lucity/services/conductor/internal/data"
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
type ProjectNew = platform.Project
type ProjectID = platform.ProjectID

type Project struct {
	ID           string
	Name         string
	Environments []Environment
	Databases    []Database
	CreatedAt    time.Time
}

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

func (c *Client) Projects(ctx context.Context, workspace string) ([]ProjectNew, error) {
	return c.platform.Projects(ctx, workspace)
}

func (c *Client) Project(ctx context.Context, id platform.ProjectID) (*ProjectNew, error) {
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

func (c *Client) CreateProject(ctx context.Context, ws, slug, displayName string) (*ProjectNew, error) {
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

// enrichSyncStatus queries the deployer for each environment's ArgoCD sync status.
// Best-effort: logs warnings on failure and leaves status as "UNKNOWN".
// Calls are made concurrently to avoid serial N+1 latency.
func (c *Client) enrichSyncStatus(ctx context.Context, ws string, proj *Project) {
	var wg sync.WaitGroup
	for i := range proj.Environments {
		env := &proj.Environments[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			statusCtx, statusCancel := context.WithTimeout(ctx, grpcTimeout)
			defer statusCancel()
			st, _, err := c.Deployer.GetDeploymentStatus(statusCtx, ws, proj.ID, env.Name)
			if err != nil {
				slog.Warn("failed to get deployment status", "project", proj.ID, "environment", env.Name, "error", err)
				return
			}
			env.SyncStatus = string(st)
		}()
	}
	wg.Wait()
}

// enrichServiceStatus queries the deployer for each service's K8s Deployment
// status per environment. Sets Ready and Replicas from actual pod state.
// Best-effort: logs warnings on failure.
func (c *Client) enrichServiceStatus(ctx context.Context, ws string, proj *Project) {
	var wg sync.WaitGroup
	for i := range proj.Environments {
		env := &proj.Environments[i]
		for j := range env.Services {
			si := &env.Services[j]
			wg.Add(1)
			go func() {
				defer wg.Done()
				statusCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
				defer cancel()
				st, err := c.Deployer.ServiceStatus(statusCtx, ws, proj.ID, env.Name, si.Name)
				if err != nil {
					slog.Warn("failed to get service status",
						"project", proj.ID,
						"environment", env.Name,
						"service", si.Name,
						"error", err)
					return
				}
				si.Ready = st.Ready
				si.Replicas = st.Replicas
				if st.Scaling != nil {
					si.Scaling = ScalingConfig{Replicas: st.Scaling.Replicas}
					if st.Scaling.AutoscalingEnabled {
						si.Scaling.Autoscaling = &AutoscalingConfig{
							Enabled:     true,
							MinReplicas: st.Scaling.MinReplicas,
							MaxReplicas: st.Scaling.MaxReplicas,
							TargetCPU:   st.Scaling.TargetCPU,
						}
					}
				}
				if st.Resources != nil {
					si.Resources = &ServiceResources{
						CpuMillicores:      st.Resources.CPUMillicores,
						MemoryMB:           st.Resources.MemoryMB,
						CpuLimitMillicores: st.Resources.CPULimitMillicores,
						MemoryLimitMB:      st.Resources.MemoryLimitMB,
					}
				}
			}()
		}
	}
	wg.Wait()
}

// enrichDatabaseStatus queries the deployer for each database's runtime status
// per environment. Best-effort: logs warnings on failure.
func (c *Client) enrichDatabaseStatus(ctx context.Context, ws string, proj *Project) {
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
				st, err := c.Deployer.DatabaseStatus(statusCtx, ws, proj.ID, envPtr.Name, dbInfo.Name)
				inst := DatabaseInstance{
					Name:      dbInfo.Name,
					Version:   dbInfo.Version,
					Size:      dbInfo.Size,
					Instances: dbInfo.Instances,
				}
				if err != nil {
					slog.Warn("failed to get database status", "project", proj.ID, "environment", envPtr.Name, "database", dbInfo.Name, "error", err)
				} else {
					inst.Ready = st.Ready
					if st.Instances > 0 {
						inst.Instances = st.Instances
					}
					if st.Volume != nil {
						inst.Volume = &Volume{
							Name:          st.Volume.Name,
							Size:          st.Volume.Size,
							RequestedSize: st.Volume.RequestedSize,
							UsedBytes:     st.Volume.UsedBytes,
							CapacityBytes: st.Volume.CapacityBytes,
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
func (c *Client) enrichDeploymentHistory(ctx context.Context, ws string, proj *Project) {
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
				entries, err := c.Packager.DeploymentHistory(histCtx, ws, proj.ID, env.Name, si.Name)
				if err != nil {
					slog.Warn("failed to get deployment history", "project", proj.ID, "environment", env.Name, "service", si.Name, "error", err)
					return
				}

				for k, e := range entries {
					si.Deployments = append(si.Deployments, Deployment{
						// TODO: rebuild as platform.DeploymentID after typed-ID migration
						ImageTag:  e.ImageTag,
						Active:    k == 0,
						Timestamp: e.DeployedAt,
						Revision:  e.Revision,
						Message:   fmt.Sprintf("deploy(%s): %s %s", env.Name, si.Name, e.ImageTag),
					})
					_ = k
				}
			}()
		}
	}
	wg.Wait()
}

// shaPattern matches a hex string of 7+ characters (git short SHA).
var shaPattern = regexp.MustCompile(`^[0-9a-f]{7,}$`)

// enrichCommitMessages fetches source commit messages from GitHub for
// deployment entries whose imageTag is a git SHA. Uses per-service installation
// tokens (from the service's GitHubInstallationID). Best-effort — failures
// are silently ignored. Also sets SourceURL for each SHA-based deployment.
func (c *Client) enrichCommitMessages(ctx context.Context, proj *Project) {
	if c.GitHubApp == nil {
		return
	}

	// Build service name → info lookup from environment services
	type serviceInfo struct {
		sourceURL      string
		installationID int64
	}
	services := make(map[string]serviceInfo)
	for _, env := range proj.Environments {
		for _, si := range env.Services {
			if si.SourceURL != "" && si.GitHubInstallationID != 0 {
				services[si.Name] = serviceInfo{
					sourceURL:      si.SourceURL,
					installationID: si.GitHubInstallationID,
				}
			}
		}
	}

	// Collect unique (owner/repo, sha) pairs that need fetching, grouped by installation
	type commitKey struct{ owner, repo, sha string }
	type commitResult struct{ message, url string }

	// installation ID → set of commit keys
	byInstallation := make(map[int64]map[commitKey]bool)
	for _, env := range proj.Environments {
		for _, si := range env.Services {
			info := services[si.Name]
			if info.sourceURL == "" {
				continue
			}
			owner, repo := parseGitHubRepoURL(info.sourceURL)
			if owner == "" {
				continue
			}
			for _, dep := range si.Deployments {
				if shaPattern.MatchString(dep.ImageTag) {
					if byInstallation[info.installationID] == nil {
						byInstallation[info.installationID] = make(map[commitKey]bool)
					}
					byInstallation[info.installationID][commitKey{owner, repo, dep.ImageTag}] = true
				}
			}
		}
	}

	if len(byInstallation) == 0 {
		return
	}

	// Fetch commit messages concurrently, one token per installation
	results := make(map[commitKey]commitResult)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for instID, keys := range byInstallation {
		ghToken, err := c.installationTokenForService(ctx, instID)
		if err != nil {
			slog.Debug("skipping commit enrichment for installation", "installation_id", instID, "reason", err)
			continue
		}
		client := gh.NewClient(nil).WithAuthToken(ghToken)

		for key := range keys {
			wg.Add(1)
			go func() {
				defer wg.Done()
				fetchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()

				commit, _, err := client.Repositories.GetCommit(fetchCtx, key.owner, key.repo, key.sha, nil)
				if err != nil {
					slog.Warn("failed to fetch commit message", "owner", key.owner, "repo", key.repo, "sha", key.sha, "error", err)
					return
				}

				msg := commit.GetCommit().GetMessage()
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
	}
	wg.Wait()

	// Apply results to all deployments
	for i := range proj.Environments {
		env := &proj.Environments[i]
		for j := range env.Services {
			si := &env.Services[j]
			info := services[si.Name]
			if info.sourceURL == "" {
				continue
			}
			owner, repo := parseGitHubRepoURL(info.sourceURL)
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

func projectFromInfo(ws string, p data.ProjectInfo) Project {
	displayName := p.DisplayName
	if displayName == "" {
		displayName = p.Name // fall back to slug for old projects
	}

	proj := Project{
		ID:        p.Name,
		Name:      displayName,
		CreatedAt: p.CreatedAt,
	}

	// Build a lookup of per-env service info from the richer EnvironmentInfos.
	envInfoMap := make(map[string][]data.ServiceInstanceInfo, len(p.EnvironmentInfos))
	for _, ei := range p.EnvironmentInfos {
		envInfoMap[ei.Name] = ei.Services
	}

	for _, envName := range p.Environments {
		env := Environment{
			ID:         p.Name + "/" + envName,
			Name:       envName,
			Namespace:  labels.NamespaceFor(ws, p.Name, envName),
			SyncStatus: "UNKNOWN",
		}

		// Build enriched ServiceInstances from environment data
		for _, svc := range envInfoMap[envName] {
			env.Services = append(env.Services, ServiceInstance{
				// TODO: rebuild as platform.ServiceID after typed-ID migration
				Name:                 svc.Name,
				Image:                svc.Image,
				Port:                 svc.Port,
				Framework:            svc.Framework,
				SourceURL:            svc.SourceURL,
				ContextPath:          svc.ContextPath,
				GitHubInstallationID: svc.GitHubInstallationID,
				StartCommand:         svc.StartCommand,
				CustomStartCommand:   svc.CustomStartCommand,
				ImageTag:             svc.ImageTag,
				Domains:              svc.Domains,
			})
		}

		proj.Environments = append(proj.Environments, env)
	}

	for _, db := range p.Databases {
		proj.Databases = append(proj.Databases, Database{
			Name:      db.Name,
			Version:   db.Version,
			Instances: db.Instances,
			Size:      db.Size,
		})
	}

	return proj
}
