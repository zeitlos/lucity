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
}

// NewGitHubProvider creates a Provider backed by GitHub repositories.
// The token is a user OAuth access token from the GitHub App.
func NewGitHubProvider(token string) *GitHubProvider {
	return &GitHubProvider{
		client: gh.NewClient(nil).WithAuthToken(token),
		token:  token,
	}
}

// CreateRepo creates a GitOps repo in the project's org on GitHub.
// Project name is org-scoped: "zeitlos/myapp" → creates repo "zeitlos/myapp-gitops".
func (p *GitHubProvider) CreateRepo(ctx context.Context, project, sourceURL string) (string, error) {
	org, name, err := SplitProject(project)
	if err != nil {
		return "", err
	}

	repoName := name + RepoSuffix
	repoOpts := &gh.Repository{
		Name:     gh.Ptr(repoName),
		Private:  gh.Ptr(true),
		AutoInit: gh.Ptr(false),
	}

	// Try org endpoint first; fall back to user endpoint for personal accounts
	repo, _, err := p.client.Repositories.Create(ctx, org, repoOpts)
	if err != nil {
		// Personal accounts can't use POST /orgs/{org}/repos — use POST /user/repos
		repo, _, err = p.client.Repositories.Create(ctx, "", repoOpts)
	}
	if err != nil {
		return "", fmt.Errorf("failed to create gitops repo %s/%s: %w", org, repoName, err)
	}

	slog.Info("created gitops repo", "repo", repo.GetFullName())

	// Clone, populate, commit, and push
	if err := p.initRepoContents(repo.GetCloneURL(), project, sourceURL); err != nil {
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
	org, name, err := SplitProject(project)
	if err != nil {
		return nil, err
	}

	repoName := name + RepoSuffix

	meta, err := p.readProjectMeta(ctx, org, repoName)
	if err != nil {
		return nil, err
	}

	ghRepo, _, err := p.client.Repositories.Get(ctx, org, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo info: %w", err)
	}
	meta.RepoURL = ghRepo.GetCloneURL()

	return meta, nil
}

// DeleteRepo removes a project's GitOps repository from GitHub.
func (p *GitHubProvider) DeleteRepo(ctx context.Context, project string) error {
	org, name, err := SplitProject(project)
	if err != nil {
		return err
	}

	repoName := name + RepoSuffix

	_, err = p.client.Repositories.Delete(ctx, org, repoName)
	if err != nil {
		return fmt.Errorf("failed to delete repo %s/%s: %w", org, repoName, err)
	}

	slog.Info("deleted gitops repo", "org", org, "repo", repoName)
	return nil
}

// AddService adds a service definition to the project's base/values.yaml.
func (p *GitHubProvider) AddService(ctx context.Context, project string, svc ServiceDef) error {
	org, name, err := SplitProject(project)
	if err != nil {
		return err
	}
	repoName := name + RepoSuffix

	inner, sha, err := p.readSubchartValuesGH(ctx, org, repoName, "base/values.yaml")
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
		"public":   svc.Public,
	}
	if svc.Framework != "" {
		svcEntry["framework"] = svc.Framework
	}
	services[svc.Name] = svcEntry
	inner["services"] = services

	if err := p.writeSubchartValuesGH(ctx, org, repoName, "base/values.yaml", inner, sha,
		fmt.Sprintf("config: add service %s", svc.Name)); err != nil {
		return fmt.Errorf("failed to update base/values.yaml: %w", err)
	}

	slog.Info("added service to gitops repo", "project", project, "service", svc.Name)
	return nil
}

// RemoveService removes a service definition from the project's base/values.yaml.
func (p *GitHubProvider) RemoveService(ctx context.Context, project, service string) error {
	org, name, err := SplitProject(project)
	if err != nil {
		return err
	}
	repoName := name + RepoSuffix

	inner, sha, err := p.readSubchartValuesGH(ctx, org, repoName, "base/values.yaml")
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

	if err := p.writeSubchartValuesGH(ctx, org, repoName, "base/values.yaml", inner, sha,
		fmt.Sprintf("config: remove service %s", service)); err != nil {
		return fmt.Errorf("failed to update base/values.yaml: %w", err)
	}

	slog.Info("removed service from gitops repo", "project", project, "service", service)
	return nil
}

// UpdateImageTag updates the image tag for a service in an environment's values.yaml.
func (p *GitHubProvider) UpdateImageTag(ctx context.Context, project, environment, service, tag, digest string) error {
	org, name, err := SplitProject(project)
	if err != nil {
		return err
	}
	repoName := name + RepoSuffix

	filePath := fmt.Sprintf("environments/%s/values.yaml", environment)
	inner, sha, err := p.readSubchartValuesGH(ctx, org, repoName, filePath)
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

	commitMsg := fmt.Sprintf("deploy(%s): %s %s", environment, service, tag)
	if err := p.writeSubchartValuesGH(ctx, org, repoName, filePath, inner, sha, commitMsg); err != nil {
		return fmt.Errorf("failed to update %s: %w", filePath, err)
	}

	slog.Info("updated image tag in gitops repo",
		"project", project, "environment", environment, "service", service, "tag", tag)
	return nil
}

// Services reads the services defined in the project's base/values.yaml.
func (p *GitHubProvider) Services(ctx context.Context, project string) ([]ServiceDef, error) {
	org, name, err := SplitProject(project)
	if err != nil {
		return nil, err
	}
	repoName := name + RepoSuffix

	inner, _, err := p.readSubchartValuesGH(ctx, org, repoName, "base/values.yaml")
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
func (p *GitHubProvider) initRepoContents(cloneURL, project, sourceURL string) error {
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
		"project.yaml":                         projectYAML(project, sourceURL, now),
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

	_, err = wt.Commit(fmt.Sprintf("init: %s from %s", project, sourceURL), &git.CommitOptions{
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
	org, name, err := SplitProject(project)
	if err != nil {
		return err
	}
	repoName := name + RepoSuffix

	var content []byte
	if fromEnvironment != "" {
		// Copy values from the source environment
		srcPath := fmt.Sprintf("environments/%s/values.yaml", fromEnvironment)
		values, _, err := p.readValuesYAML(ctx, org, repoName, srcPath)
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
	_, _, err = p.client.Repositories.CreateFile(ctx, org, repoName, filePath, &gh.RepositoryContentFileOptions{
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
	org, name, err := SplitProject(project)
	if err != nil {
		return err
	}
	repoName := name + RepoSuffix

	filePath := fmt.Sprintf("environments/%s/values.yaml", environment)
	content, _, _, err := p.client.Repositories.GetContents(ctx, org, repoName, filePath, nil)
	if err != nil {
		return fmt.Errorf("failed to get environment file %s: %w", filePath, err)
	}

	_, _, err = p.client.Repositories.DeleteFile(ctx, org, repoName, filePath, &gh.RepositoryContentFileOptions{
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
	org, name, err := SplitProject(project)
	if err != nil {
		return "", err
	}
	repoName := name + RepoSuffix

	// Read the source environment's values
	srcPath := fmt.Sprintf("environments/%s/values.yaml", fromEnv)
	srcInner, _, err := p.readSubchartValuesGH(ctx, org, repoName, srcPath)
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
	if err := p.UpdateImageTag(ctx, project, toEnv, service, tag, ""); err != nil {
		return "", fmt.Errorf("failed to promote to %s: %w", toEnv, err)
	}

	slog.Info("promoted service", "project", project, "service", service,
		"from", fromEnv, "to", toEnv, "tag", tag)
	return tag, nil
}

// DeploymentHistory returns deployment history for a service in an environment
// by querying the GitHub Commits API and parsing commit messages.
func (p *GitHubProvider) DeploymentHistory(ctx context.Context, project, environment, service string) ([]DeploymentEntry, error) {
	org, name, err := SplitProject(project)
	if err != nil {
		return nil, err
	}
	repoName := name + RepoSuffix

	// Filter commits by the environment's values file path
	filePath := fmt.Sprintf("environments/%s/values.yaml", environment)
	commits, _, err := p.client.Repositories.ListCommits(ctx, org, repoName, &gh.CommitsListOptions{
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

// UpdateServiceConfig updates a service's base configuration in base/values.yaml.
func (p *GitHubProvider) UpdateServiceConfig(ctx context.Context, project, service string, public *bool) error {
	org, name, err := SplitProject(project)
	if err != nil {
		return err
	}
	repoName := name + RepoSuffix

	inner, sha, err := p.readSubchartValuesGH(ctx, org, repoName, "base/values.yaml")
	if err != nil {
		return fmt.Errorf("failed to read base/values.yaml: %w", err)
	}

	services, ok := inner["services"].(map[string]any)
	if !ok {
		return fmt.Errorf("no services found in base/values.yaml")
	}

	svcEntry, ok := services[service].(map[string]any)
	if !ok {
		return fmt.Errorf("service %q not found", service)
	}

	if public != nil {
		svcEntry["public"] = *public
	}
	services[service] = svcEntry
	inner["services"] = services

	if err := p.writeSubchartValuesGH(ctx, org, repoName, "base/values.yaml", inner, sha,
		fmt.Sprintf("config(service): update %s", service)); err != nil {
		return fmt.Errorf("failed to update base/values.yaml: %w", err)
	}

	slog.Info("updated service config", "project", project, "service", service)
	return nil
}

// SetServiceDomain sets or removes the domain hostname for a service in an environment.
func (p *GitHubProvider) SetServiceDomain(ctx context.Context, project, environment, service, host string) error {
	org, name, err := SplitProject(project)
	if err != nil {
		return err
	}
	repoName := name + RepoSuffix

	filePath := fmt.Sprintf("environments/%s/values.yaml", environment)
	inner, sha, err := p.readSubchartValuesGH(ctx, org, repoName, filePath)
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
	if err := p.writeSubchartValuesGH(ctx, org, repoName, filePath, inner, sha, commitMsg); err != nil {
		return fmt.Errorf("failed to update %s: %w", filePath, err)
	}

	slog.Info("set service domain", "project", project, "environment", environment, "service", service, "host", host)
	return nil
}

// EnvironmentServices reads per-environment service state from the environment's values.yaml.
func (p *GitHubProvider) EnvironmentServices(ctx context.Context, project, environment string) ([]ServiceInstanceMeta, error) {
	org, name, err := SplitProject(project)
	if err != nil {
		return nil, err
	}
	repoName := name + RepoSuffix

	filePath := fmt.Sprintf("environments/%s/values.yaml", environment)
	inner, _, err := p.readSubchartValuesGH(ctx, org, repoName, filePath)
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

// projectYAMLData matches the structure of project.yaml for parsing.
type projectYAMLData struct {
	Name      string `yaml:"name"`
	SourceURL string `yaml:"source_url"`
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
		SourceURL: d.SourceURL,
		CreatedAt: createdAt,
	}, nil
}
