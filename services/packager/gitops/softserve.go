package gitops

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
)

// SoftServeProvider implements Provider using Soft-serve as the git backend.
// Repo management is done via SSH commands; file operations via git clone/push.
type SoftServeProvider struct {
	sshAddr  string     // e.g., "localhost:23231"
	httpAddr string     // e.g., "http://localhost:23232"
	sshKey   ssh.Signer // admin SSH key for repo management commands
	token    string     // HTTP access token for git push
}

// NewSoftServeProvider creates a Provider backed by Soft-serve.
func NewSoftServeProvider(sshAddr, httpAddr string, sshKey ssh.Signer, token string) *SoftServeProvider {
	return &SoftServeProvider{
		sshAddr:  sshAddr,
		httpAddr: httpAddr,
		sshKey:   sshKey,
		token:    token,
	}
}

// CreateRepo creates a GitOps repo on Soft-serve and populates it.
func (p *SoftServeProvider) CreateRepo(ctx context.Context, project string) (string, error) {
	repoName := project + RepoSuffix
	cloneURL := p.repoHTTPURL(repoName)

	// Create the repo via SSH (idempotent: handle "already exists")
	_, err := p.sshCmd("repo", "create", repoName)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return "", fmt.Errorf("failed to create repo %s: %w", repoName, err)
		}

		slog.Info("repo already exists, checking state", "repo", repoName)

		if p.repoHasContent(repoName) {
			slog.Info("repo already initialized", "repo", repoName)
			return cloneURL, nil
		}

		slog.Info("repo exists but empty, re-initializing", "repo", repoName)
	}

	// Make it private (idempotent)
	if _, err := p.sshCmd("repo", "private", repoName, "true"); err != nil {
		slog.Warn("failed to set repo private", "repo", repoName, "error", err)
	}

	slog.Info("initializing softserve repo", "repo", repoName, "url", cloneURL)

	// Initialize with directory structure and files
	if err := p.initRepoContents(cloneURL, project); err != nil {
		return "", fmt.Errorf("failed to initialize repo contents: %w", err)
	}

	return cloneURL, nil
}

// repoHasContent checks whether a Soft-serve repo has been initialized with project.yaml.
func (p *SoftServeProvider) repoHasContent(repoName string) bool {
	dir, cleanup, err := p.cloneRepo(repoName)
	if err != nil {
		return false
	}
	defer cleanup()

	_, err = os.Stat(filepath.Join(dir, "project.yaml"))
	return err == nil
}

// Repos lists all GitOps repos on Soft-serve.
func (p *SoftServeProvider) Repos(ctx context.Context) ([]ProjectMeta, error) {
	output, err := p.sshCmd("repo", "list")
	if err != nil {
		return nil, fmt.Errorf("failed to list repos: %w", err)
	}

	var projects []ProjectMeta
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		repoName := strings.TrimSpace(line)
		if repoName == "" || !strings.HasSuffix(repoName, RepoSuffix) {
			continue
		}

		meta, err := p.readProjectMeta(repoName)
		if err != nil {
			slog.Warn("skipping repo with unreadable metadata",
				"repo", repoName, "error", err,
				"hint", "retry creating the project to recover")
			continue
		}
		meta.RepoURL = p.repoHTTPURL(repoName)
		projects = append(projects, *meta)
	}

	return projects, nil
}

// Repo reads a single project's metadata.
func (p *SoftServeProvider) Repo(ctx context.Context, project string) (*ProjectMeta, error) {
	repoName := project + RepoSuffix

	meta, err := p.readProjectMeta(repoName)
	if err != nil {
		return nil, err
	}
	meta.RepoURL = p.repoHTTPURL(repoName)

	return meta, nil
}

// DeleteRepo removes a repo from Soft-serve.
func (p *SoftServeProvider) DeleteRepo(ctx context.Context, project string) error {
	repoName := project + RepoSuffix

	if _, err := p.sshCmd("repo", "delete", repoName); err != nil {
		return fmt.Errorf("failed to delete repo %s: %w", repoName, err)
	}

	slog.Info("deleted softserve repo", "repo", repoName)
	return nil
}

// AddService adds a service to base/values.yaml.
func (p *SoftServeProvider) AddService(ctx context.Context, project string, svc ServiceDef) error {
	return p.modifyRepo(ctx, project, fmt.Sprintf("config: add service %s", svc.Name), false, func(dir string) error {
		path := filepath.Join(dir, "base", "values.yaml")
		inner, err := readSubchartValues(path)
		if err != nil {
			return err
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

		return writeSubchartValues(path, inner)
	})
}

// RemoveService removes a service from base/values.yaml.
func (p *SoftServeProvider) RemoveService(ctx context.Context, project, service string) error {
	return p.modifyRepo(ctx, project, fmt.Sprintf("config: remove service %s", service), false, func(dir string) error {
		path := filepath.Join(dir, "base", "values.yaml")
		inner, err := readSubchartValues(path)
		if err != nil {
			return err
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

		if err := writeSubchartValues(path, inner); err != nil {
			return err
		}

		// Clean up serviceRefs referencing the deleted service across all environments.
		envFiles, _ := filepath.Glob(filepath.Join(dir, "environments", "*", "values.yaml"))
		for _, envPath := range envFiles {
			envInner, readErr := readSubchartValues(envPath)
			if readErr != nil {
				continue
			}
			envSvcs, ok := envInner["services"].(map[string]any)
			if !ok {
				continue
			}
			modified := false
			for svcName, svcRaw := range envSvcs {
				svcMap, ok := svcRaw.(map[string]any)
				if !ok {
					continue
				}
				refs := parseServiceRefs(svcMap)
				if refs == nil {
					continue
				}
				changed := false
				for refName, ref := range refs {
					if ref.Service == service {
						delete(refs, refName)
						changed = true
					}
				}
				if !changed {
					continue
				}
				if len(refs) == 0 {
					delete(svcMap, "serviceRefs")
				} else {
					svcMap["serviceRefs"] = serviceRefsToAny(refs)
				}
				envSvcs[svcName] = svcMap
				modified = true
			}
			if modified {
				envInner["services"] = envSvcs
				if writeErr := writeSubchartValues(envPath, envInner); writeErr != nil {
					return writeErr
				}
			}
		}

		return nil
	})
}

// UpdateImageTag updates the image tag for a service in an environment's values.yaml.
func (p *SoftServeProvider) UpdateImageTag(ctx context.Context, project, environment, service, tag, digest, commitPrefix string) error {
	if commitPrefix == "" {
		commitPrefix = "deploy"
	}
	return p.modifyRepo(ctx, project, fmt.Sprintf("%s(%s): %s %s", commitPrefix, environment, service, tag), true, func(dir string) error {
		filePath := filepath.Join(dir, "environments", environment, "values.yaml")
		inner, err := readSubchartValues(filePath)
		if err != nil {
			return err
		}

		services, ok := inner["services"].(map[string]any)
		if !ok {
			services = make(map[string]any)
		}

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

		return writeSubchartValues(filePath, inner)
	})
}

// Services reads the services from base/values.yaml.
func (p *SoftServeProvider) Services(ctx context.Context, project string) ([]ServiceDef, error) {
	repoName := project + RepoSuffix

	dir, cleanup, err := p.cloneRepo(repoName)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	inner, err := readSubchartValues(filepath.Join(dir, "base", "values.yaml"))
	if err != nil {
		return nil, err
	}

	services, ok := inner["services"].(map[string]any)
	if !ok {
		return nil, nil
	}

	return parseServiceDefs(services), nil
}

// CreateEnvironment creates a new environment directory.
func (p *SoftServeProvider) CreateEnvironment(ctx context.Context, project, environment, fromEnvironment string) error {
	return p.modifyRepo(ctx, project, fmt.Sprintf("env(create): %s", environment), false, func(dir string) error {
		envDir := filepath.Join(dir, "environments", environment)
		if err := os.MkdirAll(envDir, 0o755); err != nil {
			return fmt.Errorf("failed to create environment dir: %w", err)
		}

		var content []byte
		if fromEnvironment != "" {
			srcPath := filepath.Join(dir, "environments", fromEnvironment, "values.yaml")
			var err error
			content, err = os.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("failed to read source environment %s: %w", fromEnvironment, err)
			}
		} else {
			content = []byte(environmentValuesYAML)
		}

		return os.WriteFile(filepath.Join(envDir, "values.yaml"), content, 0o644)
	})
}

// DeleteEnvironment removes an environment directory.
func (p *SoftServeProvider) DeleteEnvironment(ctx context.Context, project, environment string) error {
	return p.modifyRepo(ctx, project, fmt.Sprintf("env(delete): %s", environment), false, func(dir string) error {
		envDir := filepath.Join(dir, "environments", environment)
		return os.RemoveAll(envDir)
	})
}

// Promote copies the image tag from one environment to another.
func (p *SoftServeProvider) Promote(ctx context.Context, project, service, fromEnv, toEnv string) (string, error) {
	var promotedTag string

	err := p.modifyRepo(ctx, project,
		fmt.Sprintf("promote(%s): %s %s from %s", toEnv, service, fromEnv, toEnv), true, func(dir string) error {
			// Read source environment
			srcPath := filepath.Join(dir, "environments", fromEnv, "values.yaml")
			srcInner, err := readSubchartValues(srcPath)
			if err != nil {
				return fmt.Errorf("failed to read source environment %s: %w", fromEnv, err)
			}

			// Extract tag
			services, ok := srcInner["services"].(map[string]any)
			if !ok {
				return fmt.Errorf("no services in %s", fromEnv)
			}
			svcEntry, ok := services[service].(map[string]any)
			if !ok {
				return fmt.Errorf("service %q not found in %s", service, fromEnv)
			}
			imageEntry, ok := svcEntry["image"].(map[string]any)
			if !ok {
				return fmt.Errorf("no image entry for service %q in %s", service, fromEnv)
			}
			tag, ok := imageEntry["tag"].(string)
			if !ok || tag == "" {
				return fmt.Errorf("no image tag for service %q in %s", service, fromEnv)
			}
			promotedTag = tag

			// Write to target environment
			dstPath := filepath.Join(dir, "environments", toEnv, "values.yaml")
			dstInner, err := readSubchartValues(dstPath)
			if err != nil {
				return fmt.Errorf("failed to read target environment %s: %w", toEnv, err)
			}

			dstServices, ok := dstInner["services"].(map[string]any)
			if !ok {
				dstServices = make(map[string]any)
			}
			dstSvc, ok := dstServices[service].(map[string]any)
			if !ok {
				dstSvc = make(map[string]any)
			}
			dstImg, ok := dstSvc["image"].(map[string]any)
			if !ok {
				dstImg = make(map[string]any)
			}
			dstImg["tag"] = tag
			dstSvc["image"] = dstImg
			dstServices[service] = dstSvc
			dstInner["services"] = dstServices

			return writeSubchartValues(dstPath, dstInner)
		})

	return promotedTag, err
}

// DeploymentHistory returns deployment history for a service in an environment
// by parsing the GitOps repo's git log for matching commit messages.
func (p *SoftServeProvider) DeploymentHistory(ctx context.Context, project, environment, service string) ([]DeploymentEntry, error) {
	repoName := project + RepoSuffix

	dir, cleanup, err := p.cloneRepoWithDepth(repoName, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to clone repo for history: %w", err)
	}
	defer cleanup()

	repo, err := git.PlainOpen(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to open repo: %w", err)
	}

	commits, err := repo.Log(&git.LogOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to read git log: %w", err)
	}
	defer commits.Close()

	var entries []DeploymentEntry
	err = commits.ForEach(func(c *object.Commit) error {
		if len(entries) >= maxDeploymentHistory {
			return fmt.Errorf("stop") // break iteration
		}

		tag, ok := parseDeployCommit(c.Message, environment, service)
		if !ok {
			return nil
		}

		entries = append(entries, DeploymentEntry{
			ImageTag:  tag,
			Revision:  c.Hash.String(),
			Timestamp: c.Author.When,
			Author:    c.Author.Name,
		})
		return nil
	})
	// The "stop" error is our break signal, not a real error
	if err != nil && err.Error() != "stop" {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return entries, nil
}

// SetServiceDomain sets or removes the domain hostname for a service in an environment.
func (p *SoftServeProvider) SetServiceDomain(ctx context.Context, project, environment, service, host string) error {
	commitMsg := fmt.Sprintf("config(%s): set domain for %s", environment, service)
	if host == "" {
		commitMsg = fmt.Sprintf("config(%s): remove domain for %s", environment, service)
	}

	return p.modifyRepo(ctx, project, commitMsg, false, func(dir string) error {
		filePath := filepath.Join(dir, "environments", environment, "values.yaml")
		inner, err := readSubchartValues(filePath)
		if err != nil {
			return err
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

		return writeSubchartValues(filePath, inner)
	})
}

// EnvironmentServices reads per-environment service state from the environment's values.yaml.
func (p *SoftServeProvider) EnvironmentServices(ctx context.Context, project, environment string) ([]ServiceInstanceMeta, error) {
	repoName := project + RepoSuffix

	dir, cleanup, err := p.cloneRepo(repoName)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	filePath := filepath.Join(dir, "environments", environment, "values.yaml")
	inner, err := readSubchartValues(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	services, ok := inner["services"].(map[string]any)
	if !ok {
		return nil, nil
	}

	return parseServiceInstanceMetas(services), nil
}

// sshCmd executes a command on the Soft-serve SSH server.
func (p *SoftServeProvider) sshCmd(args ...string) (string, error) {
	sshConfig := &ssh.ClientConfig{
		User: "admin",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(p.sshKey),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", p.sshAddr, sshConfig)
	if err != nil {
		return "", fmt.Errorf("failed to connect to soft-serve: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create ssh session: %w", err)
	}
	defer session.Close()

	cmd := strings.Join(args, " ")
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(cmd); err != nil {
		return "", fmt.Errorf("ssh command %q failed: %w (stderr: %s)", cmd, err, stderr.String())
	}

	return stdout.String(), nil
}

// repoHTTPURL returns the HTTP clone URL for a repo.
func (p *SoftServeProvider) repoHTTPURL(repoName string) string {
	return strings.TrimSuffix(p.httpAddr, "/") + "/" + repoName + ".git"
}

// cloneRepo clones a Soft-serve repo to a temp directory (shallow, depth=1).
// Returns the directory path and a cleanup function.
func (p *SoftServeProvider) cloneRepo(repoName string) (string, func(), error) {
	return p.cloneRepoWithDepth(repoName, 1)
}

// cloneRepoWithDepth clones a Soft-serve repo with the given depth.
// Use depth=0 for a full clone (needed for git log).
func (p *SoftServeProvider) cloneRepoWithDepth(repoName string, depth int) (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "lucity-gitops-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	cleanup := func() { os.RemoveAll(tmpDir) }

	cloneURL := p.repoHTTPURL(repoName)
	opts := &git.CloneOptions{
		URL: cloneURL,
		Auth: &githttp.BasicAuth{
			Username: "admin",
			Password: p.token,
		},
	}
	if depth > 0 {
		opts.Depth = depth
	}

	_, err = git.PlainClone(tmpDir, false, opts)
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to clone %s: %w", cloneURL, err)
	}

	return tmpDir, cleanup, nil
}

// modifyRepo clones a repo, applies a modification function, commits, and pushes.
func (p *SoftServeProvider) modifyRepo(ctx context.Context, project, commitMsg string, forceCommit bool, modify func(dir string) error) error {
	repoName := project + RepoSuffix

	dir, cleanup, err := p.cloneRepo(repoName)
	if err != nil {
		return err
	}
	defer cleanup()

	// Apply the modification
	if err := modify(dir); err != nil {
		return err
	}

	// Keep the embedded chart in sync on every write.
	// If the chart hasn't changed, git won't see a diff.
	if err := writeEmbeddedChart(dir); err != nil {
		return fmt.Errorf("failed to sync embedded chart: %w", err)
	}

	// Open the repo, add all changes, commit, and push
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("failed to open repo: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Add all changes
	if err := addAll(wt, dir); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	// Check if there are actual changes
	status, err := wt.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}
	if status.IsClean() && !forceCommit {
		slog.Debug("no changes to commit", "project", project)
		return nil
	}

	_, err = wt.Commit(commitMsg, &git.CommitOptions{
		AllowEmptyCommits: status.IsClean(),
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
			Username: "admin",
			Password: p.token,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	return nil
}

// initRepoContents initializes a new GitOps repo with the standard directory structure.
func (p *SoftServeProvider) initRepoContents(cloneURL, project string) error {
	tmpDir, err := os.MkdirTemp("", "lucity-gitops-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

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

	files := map[string]string{
		"project.yaml":                         projectYAML(project, now),
		"base/Chart.yaml":                      baseChartYAML(project),
		"base/values.yaml":                     baseValuesYAML,
		"environments/development/values.yaml": environmentValuesYAML,
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
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
			Username: "admin",
			Password: p.token,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	slog.Info("initialized softserve gitops repo", "project", project)
	return nil
}

// readProjectMeta reads project metadata by cloning the repo.
func (p *SoftServeProvider) readProjectMeta(repoName string) (*ProjectMeta, error) {
	dir, cleanup, err := p.cloneRepo(repoName)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	data, err := os.ReadFile(filepath.Join(dir, "project.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to read project.yaml: %w", err)
	}

	meta, err := parseProjectYAML(data)
	if err != nil {
		return nil, err
	}

	// Read services from base
	baseInner, err := readSubchartValues(filepath.Join(dir, "base", "values.yaml"))
	if err == nil {
		if services, ok := baseInner["services"].(map[string]any); ok {
			meta.Services = parseServiceDefs(services)
		}
	}

	// List environments and read per-env service image tags
	envDir := filepath.Join(dir, "environments")
	entries, err := os.ReadDir(envDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			envName := entry.Name()
			meta.Environments = append(meta.Environments, envName)

			envMeta := EnvironmentMeta{Name: envName}
			envInner, readErr := readSubchartValues(filepath.Join(envDir, envName, "values.yaml"))
			if readErr == nil {
				if envSvcs, ok := envInner["services"].(map[string]any); ok {
					envMeta.Services = parseServiceInstanceMetas(envSvcs)
				}
			}
			meta.EnvironmentInfos = append(meta.EnvironmentInfos, envMeta)
		}
	}

	return meta, nil
}

// addAll stages all changes in the working tree.
func addAll(wt *git.Worktree, dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		_, err = wt.Add(rel)
		return err
	})
}

// subchartKey is the Helm dependency name used in GitOps repos.
// Values must be scoped under this key for Helm to pass them to the subchart.
const subchartKey = "lucity-app"

// readSubchartValues reads the lucity-app subchart values from a local values.yaml.
func readSubchartValues(path string) (map[string]any, error) {
	values, err := readLocalValuesYAML(path)
	if err != nil {
		return nil, err
	}
	inner, ok := values[subchartKey].(map[string]any)
	if !ok {
		inner = make(map[string]any)
	}
	return inner, nil
}

// writeSubchartValues writes values nested under the subchart key.
func writeSubchartValues(path string, inner map[string]any) error {
	return writeLocalValuesYAML(path, map[string]any{subchartKey: inner})
}

// readLocalValuesYAML reads and parses a local YAML file.
func readLocalValuesYAML(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var values map[string]any
	if err := yaml.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	if values == nil {
		values = make(map[string]any)
	}

	return values, nil
}

// writeLocalValuesYAML marshals values and writes them to a local file.
func writeLocalValuesYAML(path string, values map[string]any) error {
	data, err := yaml.Marshal(values)
	if err != nil {
		return fmt.Errorf("failed to marshal values: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// RepoFiles returns raw file contents from the GitOps repo, keyed by relative path.
// Clones the repo and reads all files except .git/ and chart/.
func (p *SoftServeProvider) RepoFiles(ctx context.Context, project string) (map[string][]byte, error) {
	repoName := project + RepoSuffix

	dir, cleanup, err := p.cloneRepo(repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to clone repo for eject: %w", err)
	}
	defer cleanup()

	files := make(map[string][]byte)
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		// Skip .git and chart directories.
		if d.IsDir() {
			if rel == ".git" || rel == "chart" {
				return filepath.SkipDir
			}
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		files[rel] = data
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk repo directory: %w", err)
	}

	return files, nil
}

// SharedVariables returns all shared variables for an environment.
func (p *SoftServeProvider) SharedVariables(ctx context.Context, project, environment string) (map[string]string, error) {
	repoName := project + RepoSuffix

	dir, cleanup, err := p.cloneRepo(repoName)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	filePath := filepath.Join(dir, "environments", environment, "values.yaml")
	inner, err := readSubchartValues(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	return parseStringMap(inner, "sharedVariables"), nil
}

// SetSharedVariables replaces all shared variables for an environment.
func (p *SoftServeProvider) SetSharedVariables(ctx context.Context, project, environment string, vars map[string]string) error {
	return p.modifyRepo(ctx, project, fmt.Sprintf("config(%s): update shared variables", environment), false, func(dir string) error {
		filePath := filepath.Join(dir, "environments", environment, "values.yaml")
		inner, err := readSubchartValues(filePath)
		if err != nil {
			return err
		}

		if len(vars) > 0 {
			inner["sharedVariables"] = stringMapToAny(vars)
		} else {
			delete(inner, "sharedVariables")
		}

		// Propagate to services with sharedRefs
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
			var validRefs []any
			for _, refKey := range refs {
				if val, ok := vars[refKey]; ok {
					env[refKey] = val
					validRefs = append(validRefs, refKey)
				} else {
					delete(env, refKey)
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

		return writeSubchartValues(filePath, inner)
	})
}

// ServiceVariables returns all variables and shared refs for a service in an environment.
func (p *SoftServeProvider) ServiceVariables(ctx context.Context, project, environment, service string) (map[string]string, []string, map[string]DatabaseRef, map[string]ServiceRef, error) {
	repoName := project + RepoSuffix

	dir, cleanup, err := p.cloneRepo(repoName)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer cleanup()

	filePath := filepath.Join(dir, "environments", environment, "values.yaml")
	inner, err := readSubchartValues(filePath)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	services, _ := inner["services"].(map[string]any)
	svcMap, _ := services[service].(map[string]any)
	if svcMap == nil {
		return nil, nil, nil, nil, nil
	}

	vars := parseStringMap(svcMap, "env")
	refs := parseStringSlice(svcMap, "sharedRefs")
	databaseRefs := parseDatabaseRefs(svcMap)
	serviceRefs := parseServiceRefs(svcMap)
	return vars, refs, databaseRefs, serviceRefs, nil
}

// SetServiceVariables replaces all variables for a service in an environment.
func (p *SoftServeProvider) SetServiceVariables(ctx context.Context, project, environment, service string, vars map[string]string, sharedRefs []string, databaseRefs map[string]DatabaseRef, serviceRefs map[string]ServiceRef) error {
	return p.modifyRepo(ctx, project, fmt.Sprintf("config(%s): update variables for %s", environment, service), false, func(dir string) error {
		filePath := filepath.Join(dir, "environments", environment, "values.yaml")
		inner, err := readSubchartValues(filePath)
		if err != nil {
			return err
		}

		env := make(map[string]any, len(vars)+len(sharedRefs))
		for k, v := range vars {
			env[k] = v
		}

		sharedVars := parseStringMap(inner, "sharedVariables")
		var validRefs []any
		for _, refKey := range sharedRefs {
			if val, ok := sharedVars[refKey]; ok {
				env[refKey] = val
				validRefs = append(validRefs, refKey)
			}
		}

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
		if len(databaseRefs) > 0 {
			svcMap["databaseRefs"] = databaseRefsToAny(databaseRefs)
		} else {
			delete(svcMap, "databaseRefs")
		}
		if len(serviceRefs) > 0 {
			svcMap["serviceRefs"] = serviceRefsToAny(serviceRefs)
		} else {
			delete(svcMap, "serviceRefs")
		}
		services[service] = svcMap
		inner["services"] = services

		return writeSubchartValues(filePath, inner)
	})
}

// SyncChart updates the embedded lucity-app chart in the GitOps repo.
func (p *SoftServeProvider) SyncChart(ctx context.Context, project string) error {
	return p.modifyRepo(ctx, project, "chart(sync): update lucity-app chart", false, func(dir string) error {
		return writeEmbeddedChart(dir)
	})
}

// AddDatabase adds a PostgreSQL database definition to base/values.yaml.
func (p *SoftServeProvider) AddDatabase(ctx context.Context, project string, db DatabaseDef) error {
	return p.modifyRepo(ctx, project, fmt.Sprintf("config: add database %s", db.Name), false, func(dir string) error {
		path := filepath.Join(dir, "base", "values.yaml")
		inner, err := readSubchartValues(path)
		if err != nil {
			return err
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

		return writeSubchartValues(path, inner)
	})
}

// RemoveDatabase removes a database definition from base/values.yaml and cleans
// up databaseRefs that reference it across all environment values files.
func (p *SoftServeProvider) RemoveDatabase(ctx context.Context, project, name string) error {
	return p.modifyRepo(ctx, project, fmt.Sprintf("config: remove database %s", name), false, func(dir string) error {
		path := filepath.Join(dir, "base", "values.yaml")
		inner, err := readSubchartValues(path)
		if err != nil {
			return err
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

		if err := writeSubchartValues(path, inner); err != nil {
			return err
		}

		// Clean up databaseRefs referencing the deleted database across all environments.
		envFiles, _ := filepath.Glob(filepath.Join(dir, "environments", "*", "values.yaml"))
		for _, envPath := range envFiles {
			envInner, readErr := readSubchartValues(envPath)
			if readErr != nil {
				continue
			}
			envSvcs, ok := envInner["services"].(map[string]any)
			if !ok {
				continue
			}
			modified := false
			for svcName, svcRaw := range envSvcs {
				svcMap, ok := svcRaw.(map[string]any)
				if !ok {
					continue
				}
				refs := parseDatabaseRefs(svcMap)
				if refs == nil {
					continue
				}
				changed := false
				for refName, ref := range refs {
					if ref.Database == name {
						delete(refs, refName)
						changed = true
					}
				}
				if !changed {
					continue
				}
				if len(refs) == 0 {
					delete(svcMap, "databaseRefs")
				} else {
					svcMap["databaseRefs"] = databaseRefsToAny(refs)
				}
				envSvcs[svcName] = svcMap
				modified = true
			}
			if modified {
				envInner["services"] = envSvcs
				if writeErr := writeSubchartValues(envPath, envInner); writeErr != nil {
					return writeErr
				}
			}
		}

		return nil
	})
}

// Databases reads the database definitions from base/values.yaml.
func (p *SoftServeProvider) Databases(ctx context.Context, project string) ([]DatabaseDef, error) {
	repoName := project + RepoSuffix

	dir, cleanup, err := p.cloneRepo(repoName)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	inner, err := readSubchartValues(filepath.Join(dir, "base", "values.yaml"))
	if err != nil {
		return nil, err
	}

	return parseDatabaseDefs(inner), nil
}

func parseServiceDefs(services map[string]any) []ServiceDef {
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
		if framework, ok := svcMap["framework"].(string); ok {
			def.Framework = framework
		}
		if sourceURL, ok := svcMap["sourceUrl"].(string); ok {
			def.SourceURL = sourceURL
		}
		if contextPath, ok := svcMap["contextPath"].(string); ok {
			def.ContextPath = contextPath
		}

		result = append(result, def)
	}
	return result
}
