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
	"gopkg.in/yaml.v3"

	gh "github.com/zeitlos/lucity/pkg/github"
)

const repoSuffix = "-gitops"

// GitHubProvider implements Provider using GitHub as the git backend.
type GitHubProvider struct {
	app            *gh.App
	installationID int64
}

// NewGitHubProvider creates a Provider backed by GitHub repositories.
func NewGitHubProvider(app *gh.App, installationID int64) *GitHubProvider {
	return &GitHubProvider{
		app:            app,
		installationID: installationID,
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

	// Create the repo on GitHub
	repo, err := p.app.CreateRepository(ctx, p.installationID, org, repoName, true)
	if err != nil {
		return "", fmt.Errorf("failed to create gitops repo: %w", err)
	}

	slog.Info("created gitops repo", "repo", repo.FullName)

	// Get an installation token for git auth
	client, err := p.app.InstallationClient(ctx, p.installationID)
	if err != nil {
		return "", fmt.Errorf("failed to create installation client: %w", err)
	}

	// Get the installation token for HTTPS git auth
	token, _, err := client.Apps.CreateInstallationToken(ctx, p.installationID, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create installation token: %w", err)
	}

	// Clone, populate, commit, and push
	if err := p.initRepoContents(repo.CloneURL, token.GetToken(), project, sourceURL); err != nil {
		return "", fmt.Errorf("failed to initialize repo contents: %w", err)
	}

	return repo.CloneURL, nil
}

// Repos lists all GitOps repos (projects) accessible via the installation.
func (p *GitHubProvider) Repos(ctx context.Context) ([]ProjectMeta, error) {
	repos, err := p.app.Repositories(ctx, p.installationID)
	if err != nil {
		return nil, fmt.Errorf("failed to list repos: %w", err)
	}

	client, err := p.app.InstallationClient(ctx, p.installationID)
	if err != nil {
		return nil, fmt.Errorf("failed to create installation client: %w", err)
	}

	var projects []ProjectMeta
	for _, r := range repos {
		if !strings.HasSuffix(r.Name, repoSuffix) {
			continue
		}

		// Fetch project.yaml from repo root
		content, _, _, err := client.Repositories.GetContents(ctx, r.Owner, r.Name, "project.yaml", nil)
		if err != nil {
			slog.Warn("skipping repo without project.yaml", "repo", r.FullName, "error", err)
			continue
		}

		raw, err := content.GetContent()
		if err != nil {
			slog.Warn("failed to decode project.yaml", "repo", r.FullName, "error", err)
			continue
		}

		meta, err := parseProjectYAML([]byte(raw))
		if err != nil {
			slog.Warn("failed to parse project.yaml", "repo", r.FullName, "error", err)
			continue
		}
		meta.RepoURL = r.CloneURL

		// List environments by checking environments/ directory
		_, dirContents, _, err := client.Repositories.GetContents(ctx, r.Owner, r.Name, "environments", nil)
		if err == nil {
			for _, entry := range dirContents {
				if entry.GetType() == "dir" {
					meta.Environments = append(meta.Environments, entry.GetName())
				}
			}
		}

		projects = append(projects, *meta)
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

	client, err := p.app.InstallationClient(ctx, p.installationID)
	if err != nil {
		return nil, fmt.Errorf("failed to create installation client: %w", err)
	}

	content, _, _, err := client.Repositories.GetContents(ctx, org, repoName, "project.yaml", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get project.yaml from %s/%s: %w", org, repoName, err)
	}

	raw, err := content.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode project.yaml: %w", err)
	}

	meta, err := parseProjectYAML([]byte(raw))
	if err != nil {
		return nil, fmt.Errorf("failed to parse project.yaml: %w", err)
	}

	// Get the repo info for clone URL
	ghRepo, _, err := client.Repositories.Get(ctx, org, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo info: %w", err)
	}
	meta.RepoURL = ghRepo.GetCloneURL()

	// List environments
	_, dirContents, _, err := client.Repositories.GetContents(ctx, org, repoName, "environments", nil)
	if err == nil {
		for _, entry := range dirContents {
			if entry.GetType() == "dir" {
				meta.Environments = append(meta.Environments, entry.GetName())
			}
		}
	}

	return meta, nil
}

// DeleteRepo removes a project's GitOps repository from GitHub.
func (p *GitHubProvider) DeleteRepo(ctx context.Context, project string) error {
	org, name, err := splitProject(project)
	if err != nil {
		return err
	}

	repoName := name + repoSuffix

	client, err := p.app.InstallationClient(ctx, p.installationID)
	if err != nil {
		return fmt.Errorf("failed to create installation client: %w", err)
	}

	_, err = client.Repositories.Delete(ctx, org, repoName)
	if err != nil {
		return fmt.Errorf("failed to delete repo %s/%s: %w", org, repoName, err)
	}

	slog.Info("deleted gitops repo", "org", org, "repo", repoName)
	return nil
}

// initRepoContents clones the empty repo, creates the GitOps directory structure,
// commits, and pushes.
func (p *GitHubProvider) initRepoContents(cloneURL, token, project, sourceURL string) error {
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
		"project.yaml":                       projectYAML(project, sourceURL, now),
		"base/Chart.yaml":                    baseChartYAML(project),
		"base/values.yaml":                   baseValuesYAML,
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
			Password: token,
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
