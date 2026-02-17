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

const repoSuffix = "-gitops"

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
	org, name, err := splitProject(project)
	if err != nil {
		return "", err
	}

	repoName := name + repoSuffix
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
			if !strings.HasSuffix(r.GetName(), repoSuffix) {
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
	org, name, err := splitProject(project)
	if err != nil {
		return nil, err
	}

	repoName := name + repoSuffix

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
	org, name, err := splitProject(project)
	if err != nil {
		return err
	}

	repoName := name + repoSuffix

	_, err = p.client.Repositories.Delete(ctx, org, repoName)
	if err != nil {
		return fmt.Errorf("failed to delete repo %s/%s: %w", org, repoName, err)
	}

	slog.Info("deleted gitops repo", "org", org, "repo", repoName)
	return nil
}

// AddService adds a service definition to the project's base/values.yaml.
func (p *GitHubProvider) AddService(ctx context.Context, project string, svc ServiceDef) error {
	org, name, err := splitProject(project)
	if err != nil {
		return err
	}
	repoName := name + repoSuffix

	values, sha, err := p.readValuesYAML(ctx, org, repoName, "base/values.yaml")
	if err != nil {
		return fmt.Errorf("failed to read base/values.yaml: %w", err)
	}

	services, ok := values["services"].(map[string]any)
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
	values["services"] = services

	if err := p.writeValuesYAML(ctx, org, repoName, "base/values.yaml", values, sha,
		fmt.Sprintf("config: add service %s", svc.Name)); err != nil {
		return fmt.Errorf("failed to update base/values.yaml: %w", err)
	}

	slog.Info("added service to gitops repo", "project", project, "service", svc.Name)
	return nil
}

// RemoveService removes a service definition from the project's base/values.yaml.
func (p *GitHubProvider) RemoveService(ctx context.Context, project, service string) error {
	org, name, err := splitProject(project)
	if err != nil {
		return err
	}
	repoName := name + repoSuffix

	values, sha, err := p.readValuesYAML(ctx, org, repoName, "base/values.yaml")
	if err != nil {
		return fmt.Errorf("failed to read base/values.yaml: %w", err)
	}

	services, ok := values["services"].(map[string]any)
	if !ok {
		return fmt.Errorf("no services found in base/values.yaml")
	}

	if _, exists := services[service]; !exists {
		return fmt.Errorf("service %q not found", service)
	}

	delete(services, service)
	values["services"] = services

	if err := p.writeValuesYAML(ctx, org, repoName, "base/values.yaml", values, sha,
		fmt.Sprintf("config: remove service %s", service)); err != nil {
		return fmt.Errorf("failed to update base/values.yaml: %w", err)
	}

	slog.Info("removed service from gitops repo", "project", project, "service", service)
	return nil
}

// UpdateImageTag updates the image tag for a service in an environment's values.yaml.
func (p *GitHubProvider) UpdateImageTag(ctx context.Context, project, environment, service, tag, digest string) error {
	org, name, err := splitProject(project)
	if err != nil {
		return err
	}
	repoName := name + repoSuffix

	filePath := fmt.Sprintf("environments/%s/values.yaml", environment)
	values, sha, err := p.readValuesYAML(ctx, org, repoName, filePath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	// Ensure services map exists
	services, ok := values["services"].(map[string]any)
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
	values["services"] = services

	commitMsg := fmt.Sprintf("deploy(%s): %s %s", environment, service, tag)
	if err := p.writeValuesYAML(ctx, org, repoName, filePath, values, sha, commitMsg); err != nil {
		return fmt.Errorf("failed to update %s: %w", filePath, err)
	}

	slog.Info("updated image tag in gitops repo",
		"project", project, "environment", environment, "service", service, "tag", tag)
	return nil
}

// Services reads the services defined in the project's base/values.yaml.
func (p *GitHubProvider) Services(ctx context.Context, project string) ([]ServiceDef, error) {
	org, name, err := splitProject(project)
	if err != nil {
		return nil, err
	}
	repoName := name + repoSuffix

	values, _, err := p.readValuesYAML(ctx, org, repoName, "base/values.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read base/values.yaml: %w", err)
	}

	services, ok := values["services"].(map[string]any)
	if !ok {
		return nil, nil
	}

	var result []ServiceDef
	for svcName, svcRaw := range services {
		svcMap, ok := svcRaw.(map[string]any)
		if !ok {
			continue
		}

		def := ServiceDef{Name: svcName}

		if imageMap, ok := svcMap["image"].(map[string]any); ok {
			if repo, ok := imageMap["repository"].(string); ok {
				def.Image = repo
			}
		}
		if port, ok := svcMap["port"].(int); ok {
			def.Port = port
		}
		if public, ok := svcMap["public"].(bool); ok {
			def.Public = public
		}
		if framework, ok := svcMap["framework"].(string); ok {
			def.Framework = framework
		}

		result = append(result, def)
	}

	return result, nil
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

// splitProject splits "org/name" into org and name.
func splitProject(project string) (org, name string, err error) {
	parts := strings.SplitN(project, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid project name %q: must be org/name", project)
	}
	return parts[0], parts[1], nil
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
