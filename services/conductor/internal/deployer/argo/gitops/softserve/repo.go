package softserve

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer/argo/gitops"
)

const httpCloneUser = "admin"

// maxDeploymentHistory is the maximum number of deployment history entries to return.
const maxDeploymentHistory = 20

type Repo struct {
	slug      string
	workspace string // optional; set via SetWorkspace by callers that know it
	dir       string
	git       *git.Repository

	httpCloneURL *url.URL
	httpToken    string

	cleanup func()
}

// SetWorkspace records the workspace that owns this repo. Used by
// ProjectName to strip the workspace prefix from the slug. Safe to
// leave unset when the workspace is not known to the caller; in that
// case ProjectName returns the slug minus the repoSuffix only.
func (r *Repo) SetWorkspace(workspace string) {
	r.workspace = workspace
}

func cloneRepo(httpCloneURL, httpToken string) (*Repo, error) {
	parsed, err := url.Parse(httpCloneURL)

	if err != nil {
		return nil, fmt.Errorf("invalid repo clone url: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("invalid git repo url scheme, only http(s) supported: %s", httpCloneURL)
	}

	if !strings.EqualFold(path.Ext(parsed.Path), ".git") {
		return nil, fmt.Errorf("invalid git repo url, missing .git extension: %s", httpCloneURL)
	}

	_, slug := path.Split(parsed.Path)
	slug = strings.TrimSuffix(slug, path.Ext(parsed.Path))

	if slug == "" {
		return nil, fmt.Errorf("invalid git repo url, missing repo slug: %s", httpCloneURL)
	}

	tmpDir, err := os.MkdirTemp("", slug)

	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	cleanup := func() { os.RemoveAll(tmpDir) }

	opts := &git.CloneOptions{
		URL: httpCloneURL,
		Auth: &githttp.BasicAuth{
			Username: httpCloneUser,
			Password: httpToken,
		},
	}
	gitRepo, err := git.PlainClone(tmpDir, false, opts)

	if errors.Is(err, transport.ErrEmptyRemoteRepository) {
		// Soft-serve creates repos with no commits; PlainClone can't materialize
		// a worktree without a HEAD. Init locally and wire up the remote so the
		// first modifyRepo push populates the remote.
		gitRepo, err = git.PlainInit(tmpDir, false)

		if err != nil {
			cleanup()
			return nil, fmt.Errorf("failed to init %s: %w", httpCloneURL, err)
		}

		if _, err := gitRepo.CreateRemote(&gitconfig.RemoteConfig{
			Name: "origin",
			URLs: []string{httpCloneURL},
		}); err != nil {
			cleanup()
			return nil, fmt.Errorf("failed to add origin remote: %w", err)
		}
	} else if err != nil {
		cleanup()
		return nil, fmt.Errorf("failed to clone %s: %w", httpCloneURL, err)
	}

	return &Repo{
		slug: slug,
		dir:  tmpDir,
		git:  gitRepo,

		httpCloneURL: parsed,
		httpToken:    httpToken,

		cleanup: cleanup,
	}, nil
}

// Cleanup removes the cloned repository from the filesystem.
func (r *Repo) Cleanup() {
	r.cleanup()
}

// isInitialized checks whether repo has been initialized.
func (r *Repo) isInitialized() bool {
	_, err := os.Stat(filepath.Join(r.dir, "base", "Chart.yaml"))

	return err == nil
}

// initialize repo with lucity-app chart, base values file and default development environment.
func (r *Repo) initialize(ctx context.Context) error {
	project := r.ProjectName()
	commitMsg := fmt.Sprintf("init: %s", project)

	files := map[string]string{
		"base/Chart.yaml":                      baseChartYAML(project),
		"base/values.yaml":                     baseValuesYAML(project),
		"environments/development/values.yaml": environmentValuesYAML,
	}

	return r.modifyRepo(commitMsg, false, func(dir string) error {
		for path, content := range files {
			fullPath := filepath.Join(dir, path)
			dir := filepath.Dir(fullPath)

			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("failed to create dir %s: %w", dir, err)
			}

			if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
				return fmt.Errorf("failed to write %s: %w", path, err)
			}
		}

		return nil
	})
}

// Metadata reads project metadata from repo files.
func (r *Repo) Metadata(ctx context.Context) (*gitops.RepoMeta, error) {
	meta := &gitops.RepoMeta{
		Name:    r.ProjectName(),
		RepoURL: r.httpCloneURL.String(),
	}

	// Read createdAt from the initial git commit.
	meta.CreatedAt = readFirstCommitTime(r.dir)

	// Read base service definitions for enrichment
	var baseDefs map[string]gitops.ServiceDef
	baseInner, err := readSubchartValues(filepath.Join(r.dir, "base", "values.yaml"))
	if err == nil {
		if services, ok := baseInner["services"].(map[string]any); ok {
			defs := parseServiceDefs(services)
			baseDefs = make(map[string]gitops.ServiceDef, len(defs))
			for _, d := range defs {
				baseDefs[d.Name] = d
			}
		}
	}

	// Read databases from base
	if baseInner != nil {
		meta.Databases = parseDatabaseDefs(baseInner)
	}

	// List environments and read per-env service state, enriched with base defs
	envDir := filepath.Join(r.dir, "environments")
	entries, err := os.ReadDir(envDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			envName := entry.Name()
			meta.Environments = append(meta.Environments, envName)

			envMeta := gitops.EnvironmentMeta{Name: envName}
			envInner, readErr := readSubchartValues(filepath.Join(envDir, envName, "values.yaml"))
			if readErr == nil {
				if envSvcs, ok := envInner["services"].(map[string]any); ok {
					instances := parseServiceInstanceMetas(envSvcs)
					// Enrich each instance with definition fields from base
					for i, inst := range instances {
						if def, ok := baseDefs[inst.Name]; ok {
							instances[i].Image = def.Image
							instances[i].Port = def.Port
							instances[i].Framework = def.Framework
							instances[i].SourceURL = def.SourceURL
							instances[i].ContextPath = def.ContextPath
							instances[i].GitHubInstallationID = def.GitHubInstallationID
							instances[i].StartCommand = def.StartCommand
						}
						// Per-env customStartCommand overrides base
						if cmd, ok := envSvcs[inst.Name].(map[string]any); ok {
							if c, ok := cmd["customStartCommand"].(string); ok {
								instances[i].CustomStartCommand = c
							} else if def, ok := baseDefs[inst.Name]; ok {
								instances[i].CustomStartCommand = def.CustomStartCommand
							}
						}
					}
					envMeta.Services = instances
				}
			}
			meta.EnvironmentInfos = append(meta.EnvironmentInfos, envMeta)
		}
	}

	return meta, nil
}

// ProjectName derives the project name from the repo slug by
// stripping the workspace prefix (when set via SetWorkspace) and
// the repoSuffix.
//
// Repo format: {workspace}-{project}-{repoSuffix}
//
// This package no longer reads the workspace from the request
// context; callers that know the workspace must call SetWorkspace
// after constructing the Repo.
func (r *Repo) ProjectName() string {
	name := r.slug
	if r.workspace != "" {
		name = strings.TrimPrefix(name, r.workspace+"-")
	}
	return strings.TrimSuffix(name, repoSuffix)
}

// AddService adds a service definition to base/values.yaml and writes the
// initial image tag to the target environment's values.yaml in a single commit.
func (r *Repo) AddService(ctx context.Context, environment string, svc gitops.ServiceDef) error {
	return r.modifyRepo(fmt.Sprintf("config(%s): add service %s", environment, svc.Name), false, func(dir string) error {
		// 1. Write service definition to base/values.yaml
		path := filepath.Join(dir, "base", "values.yaml")
		inner, err := readSubchartValues(path)
		if err != nil {
			return err
		}

		services, ok := inner["services"].(map[string]any)
		if !ok {
			services = make(map[string]any)
		}

		imageMap := map[string]any{
			"repository": svc.Image,
		}
		if svc.ImagePullPolicy != "" {
			imageMap["pullPolicy"] = svc.ImagePullPolicy
		}
		svcEntry := map[string]any{
			"image":    imageMap,
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
		if svc.GitHubInstallationID != 0 {
			svcEntry["githubInstallationId"] = fmt.Sprintf("%d", svc.GitHubInstallationID)
		}
		if svc.CustomStartCommand != "" {
			svcEntry["customStartCommand"] = svc.CustomStartCommand
		}
		if svc.StartCommand != "" {
			svcEntry["startCommand"] = svc.StartCommand
		}
		services[svc.Name] = svcEntry
		inner["services"] = services

		if err := writeSubchartValues(path, inner); err != nil {
			return err
		}

		// 2. Write initial entry to the target environment's values.yaml
		envPath, err := envFilePath(dir, environment)
		if err != nil {
			return err
		}
		envInner, readErr := readSubchartValues(envPath)
		if readErr != nil {
			return fmt.Errorf("failed to read environment values: %w", readErr)
		}
		envSvcs, ok := envInner["services"].(map[string]any)
		if !ok {
			envSvcs = make(map[string]any)
		}
		envSvcMap := map[string]any{}
		if svc.ImageTag != "" {
			envSvcMap["image"] = map[string]any{
				"tag": svc.ImageTag,
			}
		}
		envSvcs[svc.Name] = envSvcMap
		envInner["services"] = envSvcs

		return writeSubchartValues(envPath, envInner)
	})
}

// RemoveService removes a service from an environment's values.yaml.
// If no other environments reference the service, also removes from base.
func (r *Repo) RemoveService(ctx context.Context, environment, service string) error {
	return r.modifyRepo(fmt.Sprintf("config(%s): remove service %s", environment, service), false, func(dir string) error {
		// 1. Remove from target environment
		envPath, err := envFilePath(dir, environment)
		if err != nil {
			return err
		}
		envInner, err := readSubchartValues(envPath)
		if err != nil {
			return err
		}
		envSvcs, ok := envInner["services"].(map[string]any)
		if ok {
			delete(envSvcs, service)
			// Clean up serviceRefs pointing to the removed service
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
				if changed {
					if len(refs) == 0 {
						delete(svcMap, "serviceRefs")
					} else {
						svcMap["serviceRefs"] = serviceRefsToAny(refs)
					}
					envSvcs[svcName] = svcMap
				}
			}
			envInner["services"] = envSvcs
			if err := writeSubchartValues(envPath, envInner); err != nil {
				return err
			}
		}

		// 2. Check if any other environments still reference this service
		envFiles, _ := filepath.Glob(filepath.Join(dir, "environments", "*", "values.yaml"))
		referencedElsewhere := false
		for _, otherPath := range envFiles {
			if otherPath == envPath {
				continue
			}
			otherInner, readErr := readSubchartValues(otherPath)
			if readErr != nil {
				continue
			}
			otherSvcs, ok := otherInner["services"].(map[string]any)
			if !ok {
				continue
			}
			if _, exists := otherSvcs[service]; exists {
				referencedElsewhere = true
				break
			}
		}

		// 3. If orphaned, remove from base too
		if !referencedElsewhere {
			basePath := filepath.Join(dir, "base", "values.yaml")
			baseInner, readErr := readSubchartValues(basePath)
			if readErr == nil {
				baseSvcs, ok := baseInner["services"].(map[string]any)
				if ok {
					delete(baseSvcs, service)
					baseInner["services"] = baseSvcs
					_ = writeSubchartValues(basePath, baseInner)
				}
			}
		}

		return nil
	})
}

// SetCustomStartCommand sets or clears the custom start command for a service
// in an environment's values.yaml (Helm merge overrides base).
func (r *Repo) SetCustomStartCommand(ctx context.Context, environment, service, command string) error {
	commitMsg := fmt.Sprintf("config(%s/%s): set custom start command", environment, service)

	return r.modifyRepo(commitMsg, false, func(dir string) error {
		envPath := filepath.Join(dir, "environments", environment, "values.yaml")
		inner, err := readSubchartValues(envPath)
		if err != nil {
			return err
		}

		services, ok := inner["services"].(map[string]any)
		if !ok {
			services = make(map[string]any)
		}

		svcMap, ok := services[service].(map[string]any)
		if !ok {
			svcMap = make(map[string]any)
		}

		if command == "" {
			delete(svcMap, "customStartCommand")
		} else {
			svcMap["customStartCommand"] = command
		}
		services[service] = svcMap
		inner["services"] = services

		return writeSubchartValues(envPath, inner)
	})
}

// UpdateImageTag updates the image tag for a service in an environment's values.yaml.
func (r *Repo) UpdateImageTag(ctx context.Context, environment, service, tag, digest, commitPrefix string) error {
	if commitPrefix == "" {
		commitPrefix = "deploy"
	}
	return r.modifyRepo(fmt.Sprintf("%s(%s): %s %s", commitPrefix, environment, service, tag), true, func(dir string) error {
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

// CreateEnvironment creates a new environment directory.
// When duplicating from another environment, strips all domains.
func (r *Repo) CreateEnvironment(ctx context.Context, environment, fromEnvironment string) error {
	if fromEnvironment == "" {
		err := r.modifyRepo(fmt.Sprintf("env(create): %s", environment), false, func(dir string) error {
			envDir, err := envDirPath(dir, environment)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(envDir, 0o755); err != nil {
				return fmt.Errorf("failed to create environment dir: %w", err)
			}
			return os.WriteFile(filepath.Join(envDir, "values.yaml"), []byte(environmentValuesYAML), 0o644)
		})
		return err
	}

	commitMsg := fmt.Sprintf("env(create): %s from %s", environment, fromEnvironment)

	err := r.modifyRepo(commitMsg, false, func(dir string) error {
		envDir, err := envDirPath(dir, environment)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(envDir, 0o755); err != nil {
			return fmt.Errorf("failed to create environment dir: %w", err)
		}

		srcPath, err := envFilePath(dir, fromEnvironment)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("failed to read source environment %s: %w", fromEnvironment, err)
		}

		dstPath := filepath.Join(envDir, "values.yaml")

		inner, err := readSubchartValuesFromBytes(content)
		if err != nil {
			// Fallback: write raw content if we can't parse
			return os.WriteFile(dstPath, content, 0o644)
		}

		services, ok := inner["services"].(map[string]any)
		if !ok {
			return os.WriteFile(dstPath, content, 0o644)
		}

		for svcName, svcRaw := range services {
			svcMap, ok := svcRaw.(map[string]any)
			if !ok {
				continue
			}

			rawDomains, hasDomains := svcMap["domains"].([]any)
			if !hasDomains || len(rawDomains) == 0 {
				continue
			}

			delete(svcMap, "domains")

			services[svcName] = svcMap
		}

		inner["services"] = services
		return writeSubchartValues(dstPath, inner)
	})

	return err
}

// DeleteEnvironment removes an environment directory.
func (r *Repo) DeleteEnvironment(ctx context.Context, environment string) error {
	return r.modifyRepo(fmt.Sprintf("env(delete): %s", environment), false, func(dir string) error {
		envDir, err := envDirPath(dir, environment)
		if err != nil {
			return err
		}
		return os.RemoveAll(envDir)
	})
}

// Promote copies the image tag from one environment to another.
func (r *Repo) Promote(ctx context.Context, service, fromEnv, toEnv string) (string, error) {
	var promotedTag string
	commitMsg := fmt.Sprintf("promote(%s): %s %s from %s", toEnv, service, fromEnv, toEnv)

	err := r.modifyRepo(commitMsg, true, func(dir string) error {
		// Read source environment
		srcPath, err := envFilePath(dir, fromEnv)
		if err != nil {
			return err
		}
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
		dstPath, err := envFilePath(dir, toEnv)
		if err != nil {
			return err
		}
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
func (r *Repo) DeploymentHistory(ctx context.Context, environment, service string) ([]gitops.DeploymentEntry, error) {
	repo, err := git.PlainOpen(r.dir)

	if err != nil {
		return nil, fmt.Errorf("failed to open repo: %w", err)
	}

	commits, err := repo.Log(&git.LogOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to read git log: %w", err)
	}
	defer commits.Close()

	var entries []gitops.DeploymentEntry

	err = commits.ForEach(func(c *object.Commit) error {
		if len(entries) >= maxDeploymentHistory {
			return fmt.Errorf("stop") // break iteration
		}

		tag, ok := parseDeployCommit(c.Message, environment, service)
		if !ok {
			return nil
		}

		entries = append(entries, gitops.DeploymentEntry{
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

// AddDomain adds a domain hostname to a service in an environment.
func (r *Repo) AddDomain(ctx context.Context, environment, service, hostname string) error {
	commitMsg := fmt.Sprintf("config(%s): add domain %s for %s", environment, hostname, service)

	return r.modifyRepo(commitMsg, false, func(dir string) error {
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

		// Read existing domains and append (dedup)
		var domains []string
		if raw, ok := svcEntry["domains"].([]any); ok {
			for _, d := range raw {
				if s, ok := d.(string); ok {
					domains = append(domains, s)
				}
			}
		}
		for _, d := range domains {
			if d == hostname {
				return nil // already exists
			}
		}
		domains = append(domains, hostname)

		// Convert to []any for YAML marshaling
		domainsAny := make([]any, len(domains))
		for i, d := range domains {
			domainsAny[i] = d
		}
		svcEntry["domains"] = domainsAny
		services[service] = svcEntry
		inner["services"] = services

		return writeSubchartValues(filePath, inner)
	})
}

// RemoveDomain removes a domain hostname from a service in an environment.
func (r *Repo) RemoveDomain(ctx context.Context, environment, service, hostname string) error {
	commitMsg := fmt.Sprintf("config(%s): remove domain %s for %s", environment, hostname, service)

	return r.modifyRepo(commitMsg, false, func(dir string) error {
		filePath := filepath.Join(dir, "environments", environment, "values.yaml")
		inner, err := readSubchartValues(filePath)
		if err != nil {
			return err
		}

		services, ok := inner["services"].(map[string]any)
		if !ok {
			return nil
		}

		svcEntry, ok := services[service].(map[string]any)
		if !ok {
			return nil
		}

		raw, ok := svcEntry["domains"].([]any)
		if !ok {
			return nil
		}

		var filtered []any
		for _, d := range raw {
			if s, ok := d.(string); ok && s != hostname {
				filtered = append(filtered, d)
			}
		}

		if len(filtered) == 0 {
			delete(svcEntry, "domains")
		} else {
			svcEntry["domains"] = filtered
		}
		services[service] = svcEntry
		inner["services"] = services

		return writeSubchartValues(filePath, inner)
	})
}

// Domains returns all domains configured within this repo.
func (r *Repo) Domains(ctx context.Context) ([]string, error) {
	var allDomains []string

	envDir := filepath.Join(r.dir, "environments")
	entries, readErr := os.ReadDir(envDir)

	if readErr == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			envInner, readErr := readSubchartValues(filepath.Join(envDir, entry.Name(), "values.yaml"))
			if readErr != nil {
				continue
			}
			if svcs, ok := envInner["services"].(map[string]any); ok {
				for _, meta := range parseServiceInstanceMetas(svcs) {
					allDomains = append(allDomains, meta.Domains...)
				}
			}
		}
	}

	return allDomains, nil
}

// EnvironmentServices reads per-environment service state from the environment's values.yaml.
func (r *Repo) EnvironmentServices(ctx context.Context, environment string) ([]gitops.ServiceInstanceMeta, error) {
	filePath := filepath.Join(r.dir, "environments", environment, "values.yaml")
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

// AddDatabase adds a PostgreSQL database definition to base/values.yaml.
func (r *Repo) AddDatabase(ctx context.Context, db gitops.DatabaseDef) error {
	return r.modifyRepo(fmt.Sprintf("config: add database %s", db.Name), false, func(dir string) error {
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
func (r *Repo) RemoveDatabase(ctx context.Context, name string) error {
	return r.modifyRepo(fmt.Sprintf("config: remove database %s", name), false, func(dir string) error {
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
func (r *Repo) Databases(ctx context.Context) ([]gitops.DatabaseDef, error) {
	inner, err := readSubchartValues(filepath.Join(r.dir, "base", "values.yaml"))

	if err != nil {
		return nil, err
	}

	return parseDatabaseDefs(inner), nil
}

// SetResources writes resource requests/limits to an environment's values.yaml.
func (r *Repo) SetResources(ctx context.Context, environment, tier string, cpuMillicores, memoryMB, diskMB int) error {
	commitMsg := fmt.Sprintf("config(%s): set resources tier=%s cpu=%dm mem=%dMi disk=%dMi", environment, tier, cpuMillicores, memoryMB, diskMB)

	return r.modifyRepo(commitMsg, false, func(dir string) error {
		path := filepath.Join(dir, "environments", environment, "values.yaml")
		inner, err := readSubchartValues(path)
		if err != nil {
			return err
		}

		cpuStr := fmt.Sprintf("%dm", cpuMillicores)
		memStr := fmt.Sprintf("%dMi", memoryMB)
		storageStr := fmt.Sprintf("%dMi", diskMB)

		resources := map[string]any{
			"tier": strings.ToLower(tier),
		}

		if strings.EqualFold(tier, "production") {
			// Guaranteed QoS: requests = limits
			resources["requests"] = map[string]any{
				"cpu":    cpuStr,
				"memory": memStr,
			}
			resources["limits"] = map[string]any{
				"cpu":    cpuStr,
				"memory": memStr,
			}
		} else {
			// Burstable QoS: requests < limits
			// Use half of the allocation as requests, full as limits
			reqCPU := cpuMillicores / 2
			if reqCPU < 100 {
				reqCPU = 100
			}
			reqMem := memoryMB / 2
			if reqMem < 128 {
				reqMem = 128
			}
			resources["requests"] = map[string]any{
				"cpu":    fmt.Sprintf("%dm", reqCPU),
				"memory": fmt.Sprintf("%dMi", reqMem),
			}
			resources["limits"] = map[string]any{
				"cpu":    cpuStr,
				"memory": memStr,
			}
		}

		if diskMB > 0 {
			resources["storage"] = storageStr
		}

		inner["resources"] = resources
		return writeSubchartValues(path, inner)
	})
}

// SetServiceScaling writes replica count and autoscaling config for a service.
func (r *Repo) SetServiceScaling(ctx context.Context, environment, service string, replicas int, autoscaling *gitops.AutoscalingConfig) error {
	var commitMsg string
	if autoscaling != nil && autoscaling.Enabled {
		commitMsg = fmt.Sprintf("scale(%s): %s autoscaling min=%d max=%d cpu=%d%%", environment, service, autoscaling.MinReplicas, autoscaling.MaxReplicas, autoscaling.TargetCPU)
	} else {
		commitMsg = fmt.Sprintf("scale(%s): %s replicas=%d", environment, service, replicas)
	}

	return r.modifyRepo(commitMsg, false, func(dir string) error {
		path := filepath.Join(dir, "environments", environment, "values.yaml")
		inner, err := readSubchartValues(path)
		if err != nil {
			return err
		}

		services, _ := inner["services"].(map[string]any)
		if services == nil {
			services = make(map[string]any)
		}
		svcMap, _ := services[service].(map[string]any)
		if svcMap == nil {
			svcMap = make(map[string]any)
		}

		svcMap["replicas"] = replicas

		if autoscaling != nil && autoscaling.Enabled {
			svcMap["autoscaling"] = map[string]any{
				"enabled":     true,
				"minReplicas": autoscaling.MinReplicas,
				"maxReplicas": autoscaling.MaxReplicas,
				"targetCPU":   autoscaling.TargetCPU,
			}
		} else {
			delete(svcMap, "autoscaling")
		}

		services[service] = svcMap
		inner["services"] = services
		return writeSubchartValues(path, inner)
	})
}

// SetSuspended writes or removes the suspended flag in an environment's values.yaml.
func (r *Repo) SetSuspended(ctx context.Context, environment string, suspended bool) error {
	action := "suspend"
	if !suspended {
		action = "resume"
	}
	commitMsg := fmt.Sprintf("%s(%s): workspace %sed", action, environment, action)

	return r.modifyRepo(commitMsg, false, func(dir string) error {
		path := filepath.Join(dir, "environments", environment, "values.yaml")

		// Skip environments that don't exist in the GitOps repo (e.g. namespace
		// exists in K8s but was never fully initialized).
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			slog.Warn("skipping environment without GitOps values", "repo", r.slug, "environment", environment)
			return nil
		}

		inner, err := readSubchartValues(path)
		if err != nil {
			return err
		}

		if suspended {
			inner["suspended"] = true
		} else {
			delete(inner, "suspended")
		}
		return writeSubchartValues(path, inner)
	})
}

// Files returns raw file contents from the GitOps repo, keyed by relative path.
// Reads all files except .git/ and chart/.
func (r *Repo) Files(ctx context.Context) (map[string][]byte, error) {
	files := make(map[string][]byte)

	err := filepath.WalkDir(r.dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(r.dir, path)
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

// ServiceVariables returns all variables and shared refs for a service in an environment.
func (r *Repo) ServiceVariables(ctx context.Context, environment, service string) (map[string]string, []string, map[string]gitops.DatabaseRef, map[string]gitops.ServiceRef, error) {
	filePath := filepath.Join(r.dir, "environments", environment, "values.yaml")
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
func (r *Repo) SetServiceVariables(ctx context.Context, environment, service string, vars map[string]string, sharedRefs []string, databaseRefs map[string]gitops.DatabaseRef, serviceRefs map[string]gitops.ServiceRef) error {
	commitMsg := fmt.Sprintf("config(%s): update variables for %s", environment, service)

	return r.modifyRepo(commitMsg, false, func(dir string) error {
		filePath, err := envFilePath(dir, environment)
		if err != nil {
			return err
		}

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

// SharedVariables returns all shared variables for an environment.
func (r *Repo) SharedVariables(ctx context.Context, environment string) (map[string]string, error) {
	filePath, err := envFilePath(r.dir, environment)
	if err != nil {
		return nil, err
	}
	inner, err := readSubchartValues(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	return parseStringMap(inner, "sharedVariables"), nil
}

// SetSharedVariables replaces all shared variables for an environment.
func (r *Repo) SetSharedVariables(ctx context.Context, environment string, vars map[string]string) error {
	return r.modifyRepo(fmt.Sprintf("config(%s): update shared variables", environment), false, func(dir string) error {
		filePath, err := envFilePath(dir, environment)
		if err != nil {
			return err
		}
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

// readFirstCommitTime returns the author date of the oldest commit in the repo.
// Falls back to time.Time{} if the history can't be read.
func readFirstCommitTime(dir string) time.Time {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return time.Time{}
	}
	iter, err := repo.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})
	if err != nil {
		return time.Time{}
	}
	var oldest time.Time
	_ = iter.ForEach(func(c *object.Commit) error {
		oldest = c.Author.When
		return nil
	})
	return oldest
}

// modifyRepo applies a modification function, commits, and pushes.
func (r *Repo) modifyRepo(commitMsg string, forceCommit bool, modify func(dir string) error) error {
	// Apply the modification
	if err := modify(r.dir); err != nil {
		return err
	}

	// Keep the embedded chart in sync on every write.
	// If the chart hasn't changed, git won't see a diff.
	if err := writeEmbeddedChart(r.dir); err != nil {
		return fmt.Errorf("failed to sync embedded chart: %w", err)
	}

	wt, err := r.git.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Add all changes
	if err := wt.AddWithOptions(&git.AddOptions{All: true}); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	// Check if there are actual changes
	status, err := wt.Status()

	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	if status.IsClean() && !forceCommit {
		slog.Debug("no changes to commit", "repo", r.slug)
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

	err = r.git.Push(&git.PushOptions{
		Auth: &githttp.BasicAuth{
			Username: "admin",
			Password: r.httpToken,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	return nil
}

// parseDeployCommit checks if a commit message represents a deployment of the given
// service to the given environment. Returns the image tag if matched.
func parseDeployCommit(message, environment, service string) (imageTag string, ok bool) {
	// Match: deploy(<env>): <service> <tag>
	// Match: rollback(<env>): <service> <tag>
	for _, prefix := range []string{"deploy", "rollback"} {
		full := fmt.Sprintf("%s(%s): %s ", prefix, environment, service)
		if strings.HasPrefix(message, full) {
			tag := strings.TrimSpace(message[len(full):])
			if tag != "" {
				return tag, true
			}
		}
	}

	// Match: promote(<env>): <service> ...
	promotePrefix := fmt.Sprintf("promote(%s): %s ", environment, service)
	if strings.HasPrefix(message, promotePrefix) {
		// Softserve format: promote(<toEnv>): <service> <fromEnv> <toEnv>
		// The tag isn't in the message — mark as promoted.
		rest := strings.TrimSpace(message[len(promotePrefix):])
		parts := strings.Fields(rest)
		if len(parts) >= 1 {
			return fmt.Sprintf("promoted from %s", parts[0]), true
		}
		return "promoted", true
	}

	return "", false
}

// safePath joins path segments and verifies the result stays within the base directory.
// Returns an error if path traversal is detected (e.g., "../" in segments).
func safePath(base string, segments ...string) (string, error) {
	p := filepath.Join(append([]string{base}, segments...)...)
	clean := filepath.Clean(p)
	baseClean := filepath.Clean(base)
	if !strings.HasPrefix(clean+string(filepath.Separator), baseClean+string(filepath.Separator)) {
		return "", fmt.Errorf("path traversal detected: resolved path escapes base directory")
	}
	return clean, nil
}

// envFilePath returns the path to an environment's values.yaml, with path traversal protection.
func envFilePath(dir, environment string) (string, error) {
	return safePath(dir, "environments", environment, "values.yaml")
}

// envDirPath returns the path to an environment directory, with path traversal protection.
func envDirPath(dir, environment string) (string, error) {
	return safePath(dir, "environments", environment)
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
