package gitops

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	gh "github.com/google/go-github/v68/github"
	"gopkg.in/yaml.v3"
)

// GitHubProvider implements Provider using GitHub as the git backend.
// Uses the user's OAuth token for all GitHub API operations.
type GitHubProvider struct {
	client *gh.Client
	token  string // OAuth token for HTTPS git auth
	owner  string // GitHub user or org that owns GitOps repos
}

// NewGitHubProvider creates a Provider backed by GitHub repositories.
// The token is a user OAuth access token from the GitHub App.
// The owner is the GitHub login (user or org) under which repos are managed.
func NewGitHubProvider(token, owner string) *GitHubProvider {
	return &GitHubProvider{
		client: gh.NewClient(nil).WithAuthToken(token),
		token:  token,
		owner:  owner,
	}
}

// repoName returns the GitOps repo name for a project.
func (p *GitHubProvider) repoName(project string) string {
	return project + RepoSuffix
}

// CreateRepo creates a GitOps repo under the provider's owner on GitHub.
func (p *GitHubProvider) CreateRepo(ctx context.Context, project string) (string, error) {
	repoName := p.repoName(project)
	repoOpts := &gh.Repository{
		Name:     gh.Ptr(repoName),
		Private:  gh.Ptr(true),
		AutoInit: gh.Ptr(false),
	}

	// Try org endpoint first; fall back to user endpoint for personal accounts
	repo, _, err := p.client.Repositories.Create(ctx, p.owner, repoOpts)
	if err != nil {
		// Personal accounts can't use POST /orgs/{org}/repos — use POST /user/repos
		repo, _, err = p.client.Repositories.Create(ctx, "", repoOpts)
	}
	if err != nil {
		return "", fmt.Errorf("failed to create gitops repo %s/%s: %w", p.owner, repoName, err)
	}

	slog.Info("created gitops repo", "repo", repo.GetFullName())

	// Clone, populate, commit, and push
	if err := p.initRepoContents(repo.GetCloneURL(), project); err != nil {
		return "", fmt.Errorf("failed to initialize repo contents: %w", err)
	}

	return repo.GetCloneURL(), nil
}

// Repos lists all GitOps repos (projects) accessible to the user.
func (p *GitHubProvider) Repos(ctx context.Context) ([]ProjectMeta, error) {
	var projects []ProjectMeta
	opts := &gh.RepositoryListOptions{
		ListOptions: gh.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := p.client.Repositories.List(ctx, "", opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repos: %w", err)
		}

		for _, r := range repos {
			if !strings.HasSuffix(r.GetName(), RepoSuffix) {
				continue
			}

			meta, err := p.readProjectMeta(ctx, r.GetOwner().GetLogin(), r.GetName())
			if err != nil {
				slog.Warn("skipping repo", "repo", r.GetFullName(), "error", err)
				continue
			}
			meta.RepoURL = r.GetCloneURL()

			projects = append(projects, *meta)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return projects, nil
}

// Repo reads a single project's metadata from its GitOps repo.
func (p *GitHubProvider) Repo(ctx context.Context, project string) (*ProjectMeta, error) {
	repoName := p.repoName(project)

	meta, err := p.readProjectMeta(ctx, p.owner, repoName)
	if err != nil {
		return nil, err
	}

	ghRepo, _, err := p.client.Repositories.Get(ctx, p.owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo info: %w", err)
	}
	meta.RepoURL = ghRepo.GetCloneURL()

	return meta, nil
}

// DeleteRepo removes a project's GitOps repository from GitHub.
func (p *GitHubProvider) DeleteRepo(ctx context.Context, project string) error {
	repoName := p.repoName(project)

	_, err := p.client.Repositories.Delete(ctx, p.owner, repoName)
	if err != nil {
		return fmt.Errorf("failed to delete repo %s/%s: %w", p.owner, repoName, err)
	}

	slog.Info("deleted gitops repo", "owner", p.owner, "repo", repoName)
	return nil
}

// AddService adds a service definition to the project's base/values.yaml.
func (p *GitHubProvider) AddService(ctx context.Context, project string, svc ServiceDef) error {
	repoName := p.repoName(project)

	inner, sha, err := p.readSubchartValuesGH(ctx, p.owner, repoName, "base/values.yaml")
	if err != nil {
		return fmt.Errorf("failed to read base/values.yaml: %w", err)
	}

	services, ok := inner["services"].(map[string]any)
	if !ok {
		services = make(map[string]any)
	}

	svcEntry := map[string]any{
		"image": map[string]any{
			"repository": svc.Image,
			"tag":        "latest",
		},
		"port":     svc.Port,
		"replicas": 1,
	}
	if svc.Framework != "" {
		svcEntry["framework"] = svc.Framework
	}
	if svc.SourceURL != "" {
		svcEntry["sourceUrl"] = svc.SourceURL
	}
	if svc.ContextPath != "" {
		svcEntry["contextPath"] = svc.ContextPath
	}
	services[svc.Name] = svcEntry
	inner["services"] = services

	if err := p.writeSubchartValuesGH(ctx, p.owner, repoName, "base/values.yaml", inner, sha,
		fmt.Sprintf("config: add service %s", svc.Name)); err != nil {
		return fmt.Errorf("failed to update base/values.yaml: %w", err)
	}

	slog.Info("added service to gitops repo", "project", project, "service", svc.Name)
	return nil
}

// RemoveService removes a service definition from the project's base/values.yaml.
func (p *GitHubProvider) RemoveService(ctx context.Context, project, service string) error {
	repoName := p.repoName(project)

	inner, sha, err := p.readSubchartValuesGH(ctx, p.owner, repoName, "base/values.yaml")
	if err != nil {
		return fmt.Errorf("failed to read base/values.yaml: %w", err)
	}

	services, ok := inner["services"].(map[string]any)
	if !ok {
		return fmt.Errorf("no services found in base/values.yaml")
	}

	if _, exists := services[service]; !exists {
		return fmt.Errorf("service %q not found", service)
	}

	delete(services, service)
	inner["services"] = services

	if err := p.writeSubchartValuesGH(ctx, p.owner, repoName, "base/values.yaml", inner, sha,
		fmt.Sprintf("config: remove service %s", service)); err != nil {
		return fmt.Errorf("failed to update base/values.yaml: %w", err)
	}

	slog.Info("removed service from gitops repo", "project", project, "service", service)
	return nil
}

// UpdateImageTag updates the image tag for a service in an environment's values.yaml.
func (p *GitHubProvider) UpdateImageTag(ctx context.Context, project, environment, service, tag, digest, commitPrefix string) error {
	if commitPrefix == "" {
		commitPrefix = "deploy"
	}
	repoName := p.repoName(project)

	filePath := fmt.Sprintf("environments/%s/values.yaml", environment)
	inner, sha, err := p.readSubchartValuesGH(ctx, p.owner, repoName, filePath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	// Ensure services map exists
	services, ok := inner["services"].(map[string]any)
	if !ok {
		services = make(map[string]any)
	}

	// Set image tag for this service
	svcEntry, ok := services[service].(map[string]any)
	if !ok {
		svcEntry = make(map[string]any)
	}
	imageEntry, ok := svcEntry["image"].(map[string]any)
	if !ok {
		imageEntry = make(map[string]any)
	}
	imageEntry["tag"] = tag
	svcEntry["image"] = imageEntry
	services[service] = svcEntry
	inner["services"] = services

	commitMsg := fmt.Sprintf("%s(%s): %s %s", commitPrefix, environment, service, tag)
	if err := p.writeSubchartValuesGH(ctx, p.owner, repoName, filePath, inner, sha, commitMsg); err != nil {
		return fmt.Errorf("failed to update %s: %w", filePath, err)
	}

	slog.Info("updated image tag in gitops repo",
		"project", project, "environment", environment, "service", service, "tag", tag)
	return nil
}

// Services reads the services defined in the project's base/values.yaml.
func (p *GitHubProvider) Services(ctx context.Context, project string) ([]ServiceDef, error) {
	repoName := p.repoName(project)

	inner, _, err := p.readSubchartValuesGH(ctx, p.owner, repoName, "base/values.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read base/values.yaml: %w", err)
	}

	services, ok := inner["services"].(map[string]any)
	if !ok {
		return nil, nil
	}

	return parseServiceDefs(services), nil
}

// readValuesYAML reads and parses a YAML file from the GitOps repo.
// Returns the parsed map, file SHA (for updates), and any error.
func (p *GitHubProvider) readValuesYAML(ctx context.Context, owner, repoName, filePath string) (map[string]any, string, error) {
	content, _, _, err := p.client.Repositories.GetContents(ctx, owner, repoName, filePath, nil)
	if err != nil {
		return nil, "", err
	}

	raw, err := content.GetContent()
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode %s: %w", filePath, err)
	}

	var values map[string]any
	if err := yaml.Unmarshal([]byte(raw), &values); err != nil {
		return nil, "", fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	if values == nil {
		values = make(map[string]any)
	}

	return values, content.GetSHA(), nil
}

// writeValuesYAML marshals a map to YAML and commits it to the GitOps repo.
func (p *GitHubProvider) writeValuesYAML(ctx context.Context, owner, repoName, filePath string, values map[string]any, sha, commitMsg string) error {
	updated, err := yaml.Marshal(values)
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", filePath, err)
	}

	_, _, err = p.client.Repositories.UpdateFile(ctx, owner, repoName, filePath, &gh.RepositoryContentFileOptions{
		Message: gh.Ptr(commitMsg),
		Content: updated,
		SHA:     gh.Ptr(sha),
		Author: &gh.CommitAuthor{
			Name:  gh.Ptr("Lucity"),
			Email: gh.Ptr("lucity@localhost"),
		},
	})
	return err
}

// readSubchartValuesGH reads the lucity-app subchart values via the GitHub API.
func (p *GitHubProvider) readSubchartValuesGH(ctx context.Context, owner, repoName, filePath string) (map[string]any, string, error) {
	values, sha, err := p.readValuesYAML(ctx, owner, repoName, filePath)
	if err != nil {
		return nil, "", err
	}
	inner, ok := values[subchartKey].(map[string]any)
	if !ok {
		inner = make(map[string]any)
	}
	return inner, sha, nil
}

// writeSubchartValuesGH writes values nested under the subchart key via the GitHub API.
func (p *GitHubProvider) writeSubchartValuesGH(ctx context.Context, owner, repoName, filePath string, inner map[string]any, sha, commitMsg string) error {
	return p.writeValuesYAML(ctx, owner, repoName, filePath, map[string]any{subchartKey: inner}, sha, commitMsg)
}

// readProjectMeta fetches and parses project.yaml + environments from a repo.
func (p *GitHubProvider) readProjectMeta(ctx context.Context, owner, repoName string) (*ProjectMeta, error) {
	content, _, _, err := p.client.Repositories.GetContents(ctx, owner, repoName, "project.yaml", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get project.yaml from %s/%s: %w", owner, repoName, err)
	}

	raw, err := content.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode project.yaml: %w", err)
	}

	meta, err := parseProjectYAML([]byte(raw))
	if err != nil {
		return nil, fmt.Errorf("failed to parse project.yaml: %w", err)
	}

	// List environments
	_, dirContents, _, err := p.client.Repositories.GetContents(ctx, owner, repoName, "environments", nil)
	if err == nil {
		for _, entry := range dirContents {
			if entry.GetType() == "dir" {
				meta.Environments = append(meta.Environments, entry.GetName())
			}
		}
	}

	return meta, nil
}

// initRepoContents clones the empty repo, creates the GitOps directory structure,
// commits, and pushes.
func (p *GitHubProvider) initRepoContents(cloneURL, project string) error {
	tmpDir, err := os.MkdirTemp("", "lucity-gitops-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Init a new repo and add the remote
	repo, err := git.PlainInit(tmpDir, false)
	if err != nil {
		return fmt.Errorf("failed to init repo: %w", err)
	}

	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{cloneURL},
	})
	if err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	now := time.Now().UTC()

	// Create directory structure and files
	files := map[string]string{
		"project.yaml":                         projectYAML(project, now),
		"base/Chart.yaml":                      baseChartYAML(project),
		"base/values.yaml":                     baseValuesYAML,
		"environments/development/values.yaml": environmentValuesYAML,
	}

	for path, content := range files {
		fullPath := tmpDir + "/" + path
		dir := fullPath[:strings.LastIndex(fullPath, "/")]
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("failed to write %s: %w", path, err)
		}
		if _, err := wt.Add(path); err != nil {
			return fmt.Errorf("failed to stage %s: %w", path, err)
		}
	}

	// Write the embedded lucity-app chart so ArgoCD can resolve the dependency
	if err := writeEmbeddedChart(tmpDir); err != nil {
		return fmt.Errorf("failed to write embedded chart: %w", err)
	}
	if _, err := wt.Add("chart"); err != nil {
		return fmt.Errorf("failed to stage chart: %w", err)
	}

	_, err = wt.Commit(fmt.Sprintf("init: %s", project), &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Lucity",
			Email: "lucity@localhost",
			When:  now,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	err = repo.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth: &githttp.BasicAuth{
			Username: "x-access-token",
			Password: p.token,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	slog.Info("initialized gitops repo", "project", project)
	return nil
}

// CreateEnvironment creates a new environment directory in the GitOps repo.
func (p *GitHubProvider) CreateEnvironment(ctx context.Context, project, environment, fromEnvironment string) error {
	repoName := p.repoName(project)

	var content []byte
	if fromEnvironment != "" {
		// Copy values from the source environment
		srcPath := fmt.Sprintf("environments/%s/values.yaml", fromEnvironment)
		values, _, err := p.readValuesYAML(ctx, p.owner, repoName, srcPath)
		if err != nil {
			return fmt.Errorf("failed to read source environment %s: %w", fromEnvironment, err)
		}
		content, err = yaml.Marshal(values)
		if err != nil {
			return fmt.Errorf("failed to marshal values: %w", err)
		}
	} else {
		content = []byte(environmentValuesYAML)
	}

	filePath := fmt.Sprintf("environments/%s/values.yaml", environment)
	_, _, err := p.client.Repositories.CreateFile(ctx, p.owner, repoName, filePath, &gh.RepositoryContentFileOptions{
		Message: gh.Ptr(fmt.Sprintf("env(create): %s", environment)),
		Content: content,
		Author: &gh.CommitAuthor{
			Name:  gh.Ptr("Lucity"),
			Email: gh.Ptr("lucity@localhost"),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create environment %s: %w", environment, err)
	}

	slog.Info("created environment in gitops repo", "project", project, "environment", environment)
	return nil
}

// DeleteEnvironment removes an environment directory from the GitOps repo.
func (p *GitHubProvider) DeleteEnvironment(ctx context.Context, project, environment string) error {
	repoName := p.repoName(project)

	filePath := fmt.Sprintf("environments/%s/values.yaml", environment)
	content, _, _, err := p.client.Repositories.GetContents(ctx, p.owner, repoName, filePath, nil)
	if err != nil {
		return fmt.Errorf("failed to get environment file %s: %w", filePath, err)
	}

	_, _, err = p.client.Repositories.DeleteFile(ctx, p.owner, repoName, filePath, &gh.RepositoryContentFileOptions{
		Message: gh.Ptr(fmt.Sprintf("env(delete): %s", environment)),
		SHA:     gh.Ptr(content.GetSHA()),
		Author: &gh.CommitAuthor{
			Name:  gh.Ptr("Lucity"),
			Email: gh.Ptr("lucity@localhost"),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete environment %s: %w", environment, err)
	}

	slog.Info("deleted environment from gitops repo", "project", project, "environment", environment)
	return nil
}

// Promote copies the image tag for a service from one environment to another.
func (p *GitHubProvider) Promote(ctx context.Context, project, service, fromEnv, toEnv string) (string, error) {
	repoName := p.repoName(project)

	// Read the source environment's values
	srcPath := fmt.Sprintf("environments/%s/values.yaml", fromEnv)
	srcInner, _, err := p.readSubchartValuesGH(ctx, p.owner, repoName, srcPath)
	if err != nil {
		return "", fmt.Errorf("failed to read source environment %s: %w", fromEnv, err)
	}

	// Extract the image tag for the service
	services, ok := srcInner["services"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("no services in %s", srcPath)
	}
	svcEntry, ok := services[service].(map[string]any)
	if !ok {
		return "", fmt.Errorf("service %q not found in %s", service, srcPath)
	}
	imageEntry, ok := svcEntry["image"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("no image entry for service %q in %s", service, srcPath)
	}
	tag, ok := imageEntry["tag"].(string)
	if !ok || tag == "" {
		return "", fmt.Errorf("no image tag for service %q in %s", service, srcPath)
	}

	// Write the tag to the target environment (UpdateImageTag already handles subchart scoping)
	if err := p.UpdateImageTag(ctx, project, toEnv, service, tag, "", "promote"); err != nil {
		return "", fmt.Errorf("failed to promote to %s: %w", toEnv, err)
	}

	slog.Info("promoted service", "project", project, "service", service,
		"from", fromEnv, "to", toEnv, "tag", tag)
	return tag, nil
}

// DeploymentHistory returns deployment history for a service in an environment
// by querying the GitHub Commits API and parsing commit messages.
func (p *GitHubProvider) DeploymentHistory(ctx context.Context, project, environment, service string) ([]DeploymentEntry, error) {
	repoName := p.repoName(project)

	// Filter commits by the environment's values file path
	filePath := fmt.Sprintf("environments/%s/values.yaml", environment)
	commits, _, err := p.client.Repositories.ListCommits(ctx, p.owner, repoName, &gh.CommitsListOptions{
		Path:        filePath,
		ListOptions: gh.ListOptions{PerPage: 50},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list commits: %w", err)
	}

	var entries []DeploymentEntry
	for _, c := range commits {
		if len(entries) >= maxDeploymentHistory {
			break
		}

		msg := c.GetCommit().GetMessage()
		tag, ok := parseDeployCommit(msg, environment, service)
		if !ok {
			continue
		}

		entries = append(entries, DeploymentEntry{
			ImageTag:  tag,
			Revision:  c.GetSHA(),
			Timestamp: c.GetCommit().GetAuthor().GetDate().Time,
			Author:    c.GetCommit().GetAuthor().GetName(),
		})
	}

	return entries, nil
}

// SetServiceDomain sets or removes the domain hostname for a service in an environment.
func (p *GitHubProvider) SetServiceDomain(ctx context.Context, project, environment, service, host string) error {
	repoName := p.repoName(project)

	filePath := fmt.Sprintf("environments/%s/values.yaml", environment)
	inner, sha, err := p.readSubchartValuesGH(ctx, p.owner, repoName, filePath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	services, ok := inner["services"].(map[string]any)
	if !ok {
		services = make(map[string]any)
	}

	svcEntry, ok := services[service].(map[string]any)
	if !ok {
		svcEntry = make(map[string]any)
	}

	if host == "" {
		delete(svcEntry, "host")
	} else {
		svcEntry["host"] = host
	}
	services[service] = svcEntry
	inner["services"] = services

	commitMsg := fmt.Sprintf("config(%s): set domain for %s", environment, service)
	if host == "" {
		commitMsg = fmt.Sprintf("config(%s): remove domain for %s", environment, service)
	}
	if err := p.writeSubchartValuesGH(ctx, p.owner, repoName, filePath, inner, sha, commitMsg); err != nil {
		return fmt.Errorf("failed to update %s: %w", filePath, err)
	}

	slog.Info("set service domain", "project", project, "environment", environment, "service", service, "host", host)
	return nil
}

// EnvironmentServices reads per-environment service state from the environment's values.yaml.
func (p *GitHubProvider) EnvironmentServices(ctx context.Context, project, environment string) ([]ServiceInstanceMeta, error) {
	repoName := p.repoName(project)

	filePath := fmt.Sprintf("environments/%s/values.yaml", environment)
	inner, _, err := p.readSubchartValuesGH(ctx, p.owner, repoName, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	services, ok := inner["services"].(map[string]any)
	if !ok {
		return nil, nil
	}

	return parseServiceInstanceMetas(services), nil
}

// parseServiceInstanceMetas extracts per-environment service state from a raw YAML services map.
func parseServiceInstanceMetas(services map[string]any) []ServiceInstanceMeta {
	var result []ServiceInstanceMeta
	for svcName, svcRaw := range services {
		svcMap, ok := svcRaw.(map[string]any)
		if !ok {
			continue
		}

		meta := ServiceInstanceMeta{Name: svcName}
		if imageMap, ok := svcMap["image"].(map[string]any); ok {
			if tag, ok := imageMap["tag"].(string); ok {
				meta.ImageTag = tag
			}
		}
		if host, ok := svcMap["host"].(string); ok {
			meta.Host = host
		}

		result = append(result, meta)
	}
	return result
}

// RepoFiles returns raw file contents from the GitOps repo, keyed by relative path.
// Reads project.yaml, base/, and environments/ — skips chart/ since the embedded
// version is used during ejection.
func (p *GitHubProvider) RepoFiles(ctx context.Context, project string) (map[string][]byte, error) {
	repoName := p.repoName(project)

	files := make(map[string][]byte)

	// Read project.yaml
	if err := p.readFileToMap(ctx, files, p.owner, repoName, "project.yaml"); err != nil {
		slog.Warn("failed to read project.yaml during eject", "error", err)
	}

	// Read base/ directory
	if err := p.readFileToMap(ctx, files, p.owner, repoName, "base/Chart.yaml"); err != nil {
		slog.Warn("failed to read base/Chart.yaml during eject", "error", err)
	}
	if err := p.readFileToMap(ctx, files, p.owner, repoName, "base/values.yaml"); err != nil {
		slog.Warn("failed to read base/values.yaml during eject", "error", err)
	}

	// List environments and read each values.yaml
	_, dirContents, _, err := p.client.Repositories.GetContents(ctx, p.owner, repoName, "environments", nil)
	if err == nil {
		for _, entry := range dirContents {
			if entry.GetType() == "dir" {
				envPath := "environments/" + entry.GetName() + "/values.yaml"
				if err := p.readFileToMap(ctx, files, p.owner, repoName, envPath); err != nil {
					slog.Warn("failed to read environment values during eject",
						"environment", entry.GetName(), "error", err)
				}
			}
		}
	}

	return files, nil
}

// readFileToMap reads a single file from GitHub and adds it to the map.
func (p *GitHubProvider) readFileToMap(ctx context.Context, files map[string][]byte, owner, repoName, path string) error {
	content, _, _, err := p.client.Repositories.GetContents(ctx, owner, repoName, path, nil)
	if err != nil {
		return err
	}
	raw, err := content.GetContent()
	if err != nil {
		return fmt.Errorf("failed to decode %s: %w", path, err)
	}
	files[path] = []byte(raw)
	return nil
}

// SharedVariables returns all shared variables for an environment.
func (p *GitHubProvider) SharedVariables(ctx context.Context, project, environment string) (map[string]string, error) {
	repoName := p.repoName(project)

	filePath := fmt.Sprintf("environments/%s/values.yaml", environment)
	inner, _, err := p.readSubchartValuesGH(ctx, p.owner, repoName, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	return parseStringMap(inner, "sharedVariables"), nil
}

// SetSharedVariables replaces all shared variables for an environment.
// Propagates value changes to services that reference shared vars via sharedRefs.
func (p *GitHubProvider) SetSharedVariables(ctx context.Context, project, environment string, vars map[string]string) error {
	repoName := p.repoName(project)

	filePath := fmt.Sprintf("environments/%s/values.yaml", environment)
	inner, sha, err := p.readSubchartValuesGH(ctx, p.owner, repoName, filePath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	// Replace shared variables
	if len(vars) > 0 {
		inner["sharedVariables"] = stringMapToAny(vars)
	} else {
		delete(inner, "sharedVariables")
	}

	// Propagate to services that have sharedRefs
	services, _ := inner["services"].(map[string]any)
	for svcName, svcRaw := range services {
		svcMap, ok := svcRaw.(map[string]any)
		if !ok {
			continue
		}
		refs := parseStringSlice(svcMap, "sharedRefs")
		if len(refs) == 0 {
			continue
		}
		env, _ := svcMap["env"].(map[string]any)
		if env == nil {
			env = make(map[string]any)
		}
		for _, refKey := range refs {
			if val, ok := vars[refKey]; ok {
				env[refKey] = val
			} else {
				// Shared var was removed — remove from env and refs
				delete(env, refKey)
			}
		}
		// Clean up refs that no longer exist in shared vars
		var validRefs []any
		for _, refKey := range refs {
			if _, ok := vars[refKey]; ok {
				validRefs = append(validRefs, refKey)
			}
		}
		if len(env) > 0 {
			svcMap["env"] = env
		} else {
			delete(svcMap, "env")
		}
		if len(validRefs) > 0 {
			svcMap["sharedRefs"] = validRefs
		} else {
			delete(svcMap, "sharedRefs")
		}
		services[svcName] = svcMap
	}
	if len(services) > 0 {
		inner["services"] = services
	}

	if err := p.writeSubchartValuesGH(ctx, p.owner, repoName, filePath, inner, sha,
		fmt.Sprintf("config(%s): update shared variables", environment)); err != nil {
		return fmt.Errorf("failed to update %s: %w", filePath, err)
	}

	slog.Info("updated shared variables", "project", project, "environment", environment, "count", len(vars))
	return nil
}

// ServiceVariables returns all variables and shared refs for a service in an environment.
func (p *GitHubProvider) ServiceVariables(ctx context.Context, project, environment, service string) (map[string]string, []string, error) {
	repoName := p.repoName(project)

	filePath := fmt.Sprintf("environments/%s/values.yaml", environment)
	inner, _, err := p.readSubchartValuesGH(ctx, p.owner, repoName, filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	services, _ := inner["services"].(map[string]any)
	svcMap, _ := services[service].(map[string]any)
	if svcMap == nil {
		return nil, nil, nil
	}

	vars := parseStringMap(svcMap, "env")
	refs := parseStringSlice(svcMap, "sharedRefs")
	return vars, refs, nil
}

// SetServiceVariables replaces all variables for a service in an environment.
// Direct values come from vars. Keys in sharedRefs are resolved from the
// environment's shared variables and merged into the service's env.
func (p *GitHubProvider) SetServiceVariables(ctx context.Context, project, environment, service string, vars map[string]string, sharedRefs []string) error {
	repoName := p.repoName(project)

	filePath := fmt.Sprintf("environments/%s/values.yaml", environment)
	inner, sha, err := p.readSubchartValuesGH(ctx, p.owner, repoName, filePath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	// Build the merged env map: direct values + resolved shared refs
	env := make(map[string]any, len(vars)+len(sharedRefs))
	for k, v := range vars {
		env[k] = v
	}

	// Resolve shared refs
	sharedVars := parseStringMap(inner, "sharedVariables")
	var validRefs []any
	for _, refKey := range sharedRefs {
		if val, ok := sharedVars[refKey]; ok {
			env[refKey] = val
			validRefs = append(validRefs, refKey)
		} else {
			slog.Warn("shared variable not found for ref", "key", refKey, "environment", environment)
		}
	}

	// Update the service entry
	services, _ := inner["services"].(map[string]any)
	if services == nil {
		services = make(map[string]any)
	}
	svcMap, _ := services[service].(map[string]any)
	if svcMap == nil {
		svcMap = make(map[string]any)
	}

	if len(env) > 0 {
		svcMap["env"] = env
	} else {
		delete(svcMap, "env")
	}
	if len(validRefs) > 0 {
		svcMap["sharedRefs"] = validRefs
	} else {
		delete(svcMap, "sharedRefs")
	}
	services[service] = svcMap
	inner["services"] = services

	if err := p.writeSubchartValuesGH(ctx, p.owner, repoName, filePath, inner, sha,
		fmt.Sprintf("config(%s): update variables for %s", environment, service)); err != nil {
		return fmt.Errorf("failed to update %s: %w", filePath, err)
	}

	slog.Info("updated service variables", "project", project, "environment", environment, "service", service)
	return nil
}

// parseStringMap extracts a map[string]string from a nested YAML map.
func parseStringMap(m map[string]any, key string) map[string]string {
	raw, ok := m[key].(map[string]any)
	if !ok {
		return nil
	}
	result := make(map[string]string, len(raw))
	for k, v := range raw {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

// parseStringSlice extracts a []string from a nested YAML list.
func parseStringSlice(m map[string]any, key string) []string {
	raw, ok := m[key].([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// stringMapToAny converts map[string]string to map[string]any for YAML marshaling.
func stringMapToAny(m map[string]string) map[string]any {
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// projectYAMLData matches the structure of project.yaml for parsing.
type projectYAMLData struct {
	Name      string `yaml:"name"`
	CreatedAt string `yaml:"created_at"`
}

func parseProjectYAML(data []byte) (*ProjectMeta, error) {
	var d projectYAMLData
	if err := yaml.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project.yaml: %w", err)
	}

	createdAt, _ := time.Parse(time.RFC3339, d.CreatedAt)

	return &ProjectMeta{
		Name:      d.Name,
		CreatedAt: createdAt,
	}, nil
}

// AddDatabase adds a PostgreSQL database definition to the project's base/values.yaml.
func (p *GitHubProvider) AddDatabase(ctx context.Context, project string, db DatabaseDef) error {
	repoName := p.repoName(project)

	inner, sha, err := p.readSubchartValuesGH(ctx, p.owner, repoName, "base/values.yaml")
	if err != nil {
		return fmt.Errorf("failed to read base/values.yaml: %w", err)
	}

	databases, ok := inner["databases"].(map[string]any)
	if !ok {
		databases = make(map[string]any)
	}
	postgres, ok := databases["postgres"].(map[string]any)
	if !ok {
		postgres = make(map[string]any)
	}

	postgres[db.Name] = map[string]any{
		"instances": db.Instances,
		"size":      db.Size,
		"version":   db.Version,
	}
	databases["postgres"] = postgres
	inner["databases"] = databases

	if err := p.writeSubchartValuesGH(ctx, p.owner, repoName, "base/values.yaml", inner, sha,
		fmt.Sprintf("config: add database %s", db.Name)); err != nil {
		return fmt.Errorf("failed to update base/values.yaml: %w", err)
	}

	slog.Info("added database to gitops repo", "project", project, "database", db.Name)
	return nil
}

// RemoveDatabase removes a database definition from the project's base/values.yaml.
func (p *GitHubProvider) RemoveDatabase(ctx context.Context, project, name string) error {
	repoName := p.repoName(project)

	inner, sha, err := p.readSubchartValuesGH(ctx, p.owner, repoName, "base/values.yaml")
	if err != nil {
		return fmt.Errorf("failed to read base/values.yaml: %w", err)
	}

	databases, ok := inner["databases"].(map[string]any)
	if !ok {
		return fmt.Errorf("no databases found in base/values.yaml")
	}
	postgres, ok := databases["postgres"].(map[string]any)
	if !ok {
		return fmt.Errorf("no postgres databases found")
	}
	if _, exists := postgres[name]; !exists {
		return fmt.Errorf("database %q not found", name)
	}

	delete(postgres, name)
	databases["postgres"] = postgres
	inner["databases"] = databases

	if err := p.writeSubchartValuesGH(ctx, p.owner, repoName, "base/values.yaml", inner, sha,
		fmt.Sprintf("config: remove database %s", name)); err != nil {
		return fmt.Errorf("failed to update base/values.yaml: %w", err)
	}

	slog.Info("removed database from gitops repo", "project", project, "database", name)
	return nil
}

// Databases reads the database definitions from the project's base/values.yaml.
func (p *GitHubProvider) Databases(ctx context.Context, project string) ([]DatabaseDef, error) {
	repoName := p.repoName(project)

	inner, _, err := p.readSubchartValuesGH(ctx, p.owner, repoName, "base/values.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read base/values.yaml: %w", err)
	}

	return parseDatabaseDefs(inner), nil
}

// SyncChart updates the embedded lucity-app chart in the GitOps repo.
// Clones the repo, overwrites the chart/ directory, and pushes if changed.
func (p *GitHubProvider) SyncChart(ctx context.Context, project string) error {
	repoName := p.repoName(project)

	ghRepo, _, err := p.client.Repositories.Get(ctx, p.owner, repoName)
	if err != nil {
		return fmt.Errorf("failed to get repo info: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "lucity-gitops-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	repo, err := git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL: ghRepo.GetCloneURL(),
		Auth: &githttp.BasicAuth{
			Username: "x-access-token",
			Password: p.token,
		},
		Depth: 1,
	})
	if err != nil {
		return fmt.Errorf("failed to clone: %w", err)
	}

	if err := writeEmbeddedChart(tmpDir); err != nil {
		return fmt.Errorf("failed to write embedded chart: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	if err := addAll(wt, tmpDir); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}
	if status.IsClean() {
		slog.Debug("chart already up to date", "project", project)
		return nil
	}

	_, err = wt.Commit("chart(sync): update lucity-app chart", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Lucity",
			Email: "lucity@localhost",
			When:  time.Now().UTC(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	err = repo.Push(&git.PushOptions{
		Auth: &githttp.BasicAuth{
			Username: "x-access-token",
			Password: p.token,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	slog.Info("synced chart in gitops repo", "project", project)
	return nil
}

// parseDatabaseDefs extracts database definitions from the inner values map.
func parseDatabaseDefs(inner map[string]any) []DatabaseDef {
	databases, ok := inner["databases"].(map[string]any)
	if !ok {
		return nil
	}
	postgres, ok := databases["postgres"].(map[string]any)
	if !ok {
		return nil
	}

	var result []DatabaseDef
	for dbName, dbRaw := range postgres {
		dbMap, ok := dbRaw.(map[string]any)
		if !ok {
			continue
		}
		def := DatabaseDef{Name: dbName}
		if v, ok := dbMap["version"].(string); ok {
			def.Version = v
		}
		if v, ok := dbMap["instances"].(int); ok {
			def.Instances = v
		}
		if v, ok := dbMap["size"].(string); ok {
			def.Size = v
		}
		result = append(result, def)
	}
	return result
}
