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
func (p *SoftServeProvider) CreateRepo(ctx context.Context, project, sourceURL string) (string, error) {
	_, name, err := SplitProject(project)
	if err != nil {
		return "", err
	}
	repoName := name + RepoSuffix
	cloneURL := p.repoHTTPURL(repoName)

	// Create the repo via SSH (idempotent: handle "already exists")
	_, err = p.sshCmd("repo", "create", repoName)
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
	if err := p.initRepoContents(cloneURL, project, sourceURL); err != nil {
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
	_, name, err := SplitProject(project)
	if err != nil {
		return nil, err
	}
	repoName := name + RepoSuffix

	meta, err := p.readProjectMeta(repoName)
	if err != nil {
		return nil, err
	}
	meta.RepoURL = p.repoHTTPURL(repoName)

	return meta, nil
}

// DeleteRepo removes a repo from Soft-serve.
func (p *SoftServeProvider) DeleteRepo(ctx context.Context, project string) error {
	_, name, err := SplitProject(project)
	if err != nil {
		return err
	}
	repoName := name + RepoSuffix

	if _, err := p.sshCmd("repo", "delete", repoName); err != nil {
		return fmt.Errorf("failed to delete repo %s: %w", repoName, err)
	}

	slog.Info("deleted softserve repo", "repo", repoName)
	return nil
}

// AddService adds a service to base/values.yaml.
func (p *SoftServeProvider) AddService(ctx context.Context, project string, svc ServiceDef) error {
	return p.modifyRepo(ctx, project, fmt.Sprintf("config: add service %s", svc.Name), func(dir string) error {
		values, err := readLocalValuesYAML(filepath.Join(dir, "base", "values.yaml"))
		if err != nil {
			return err
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

		return writeLocalValuesYAML(filepath.Join(dir, "base", "values.yaml"), values)
	})
}

// RemoveService removes a service from base/values.yaml.
func (p *SoftServeProvider) RemoveService(ctx context.Context, project, service string) error {
	return p.modifyRepo(ctx, project, fmt.Sprintf("config: remove service %s", service), func(dir string) error {
		values, err := readLocalValuesYAML(filepath.Join(dir, "base", "values.yaml"))
		if err != nil {
			return err
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

		return writeLocalValuesYAML(filepath.Join(dir, "base", "values.yaml"), values)
	})
}

// UpdateImageTag updates the image tag for a service in an environment's values.yaml.
func (p *SoftServeProvider) UpdateImageTag(ctx context.Context, project, environment, service, tag, digest string) error {
	return p.modifyRepo(ctx, project, fmt.Sprintf("deploy(%s): %s %s", environment, service, tag), func(dir string) error {
		filePath := filepath.Join(dir, "environments", environment, "values.yaml")
		values, err := readLocalValuesYAML(filePath)
		if err != nil {
			return err
		}

		services, ok := values["services"].(map[string]any)
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
		values["services"] = services

		return writeLocalValuesYAML(filePath, values)
	})
}

// Services reads the services from base/values.yaml.
func (p *SoftServeProvider) Services(ctx context.Context, project string) ([]ServiceDef, error) {
	_, name, err := SplitProject(project)
	if err != nil {
		return nil, err
	}
	repoName := name + RepoSuffix

	dir, cleanup, err := p.cloneRepo(repoName)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	values, err := readLocalValuesYAML(filepath.Join(dir, "base", "values.yaml"))
	if err != nil {
		return nil, err
	}

	services, ok := values["services"].(map[string]any)
	if !ok {
		return nil, nil
	}

	return parseServiceDefs(services), nil
}

// CreateEnvironment creates a new environment directory.
func (p *SoftServeProvider) CreateEnvironment(ctx context.Context, project, environment, fromEnvironment string) error {
	return p.modifyRepo(ctx, project, fmt.Sprintf("env(create): %s", environment), func(dir string) error {
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
	return p.modifyRepo(ctx, project, fmt.Sprintf("env(delete): %s", environment), func(dir string) error {
		envDir := filepath.Join(dir, "environments", environment)
		return os.RemoveAll(envDir)
	})
}

// Promote copies the image tag from one environment to another.
func (p *SoftServeProvider) Promote(ctx context.Context, project, service, fromEnv, toEnv string) (string, error) {
	var promotedTag string

	err := p.modifyRepo(ctx, project,
		fmt.Sprintf("promote(%s): %s %s from %s", toEnv, service, fromEnv, toEnv), func(dir string) error {
			// Read source environment
			srcPath := filepath.Join(dir, "environments", fromEnv, "values.yaml")
			srcValues, err := readLocalValuesYAML(srcPath)
			if err != nil {
				return fmt.Errorf("failed to read source environment %s: %w", fromEnv, err)
			}

			// Extract tag
			services, ok := srcValues["services"].(map[string]any)
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
			dstValues, err := readLocalValuesYAML(dstPath)
			if err != nil {
				return fmt.Errorf("failed to read target environment %s: %w", toEnv, err)
			}

			dstServices, ok := dstValues["services"].(map[string]any)
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
			dstValues["services"] = dstServices

			return writeLocalValuesYAML(dstPath, dstValues)
		})

	return promotedTag, err
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

// cloneRepo clones a Soft-serve repo to a temp directory.
// Returns the directory path and a cleanup function.
func (p *SoftServeProvider) cloneRepo(repoName string) (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "lucity-gitops-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	cleanup := func() { os.RemoveAll(tmpDir) }

	cloneURL := p.repoHTTPURL(repoName)
	_, err = git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL: cloneURL,
		Auth: &githttp.BasicAuth{
			Username: "admin",
			Password: p.token,
		},
		Depth: 1,
	})
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to clone %s: %w", cloneURL, err)
	}

	return tmpDir, cleanup, nil
}

// modifyRepo clones a repo, applies a modification function, commits, and pushes.
func (p *SoftServeProvider) modifyRepo(ctx context.Context, project, commitMsg string, modify func(dir string) error) error {
	_, name, err := SplitProject(project)
	if err != nil {
		return err
	}
	repoName := name + RepoSuffix

	dir, cleanup, err := p.cloneRepo(repoName)
	if err != nil {
		return err
	}
	defer cleanup()

	// Apply the modification
	if err := modify(dir); err != nil {
		return err
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
	if status.IsClean() {
		slog.Debug("no changes to commit", "project", project)
		return nil
	}

	_, err = wt.Commit(commitMsg, &git.CommitOptions{
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
func (p *SoftServeProvider) initRepoContents(cloneURL, project, sourceURL string) error {
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
		"project.yaml":                         projectYAML(project, sourceURL, now),
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
	values, err := readLocalValuesYAML(filepath.Join(dir, "base", "values.yaml"))
	if err == nil {
		if services, ok := values["services"].(map[string]any); ok {
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
			envValues, readErr := readLocalValuesYAML(filepath.Join(envDir, envName, "values.yaml"))
			if readErr == nil {
				if envSvcs, ok := envValues["services"].(map[string]any); ok {
					for svcName, svcRaw := range envSvcs {
						svcMap, ok := svcRaw.(map[string]any)
						if !ok {
							continue
						}
						imageMap, ok := svcMap["image"].(map[string]any)
						if !ok {
							continue
						}
						tag, _ := imageMap["tag"].(string)
						if tag != "" {
							envMeta.Services = append(envMeta.Services, ServiceInstanceMeta{
								Name:     svcName,
								ImageTag: tag,
							})
						}
					}
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

// parseServiceDefs converts a raw YAML services map to ServiceDef slice.
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
		if public, ok := svcMap["public"].(bool); ok {
			def.Public = public
		}
		if framework, ok := svcMap["framework"].(string); ok {
			def.Framework = framework
		}

		result = append(result, def)
	}
	return result
}
