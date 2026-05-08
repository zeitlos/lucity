package packager

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"

	"github.com/zeitlos/lucity/services/conductor/internal/data"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer/argo/eject"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer/argo/gitops"
)

// DeployerService is the narrow slice of the inproc deployer that
// the packager calls back into. Defined here to break the import
// cycle between inproc/packager and inproc/deployer.
type DeployerService interface {
	SyncDeployment(ctx context.Context, workspace, project, environment string) (data.DeploymentStatus, error)
}

type Server struct {
	gitops   gitops.Forge
	deployer DeployerService

	workloadDomain string
}

// NewServer creates a packager server with the given GitOps provider.
// The deployer reference can be nil at construction time and wired
// later via SetDeployer to break the inproc cross-package cycle.
func NewServer(forge gitops.Forge, deployerSvc DeployerService, workloadDomain string) *Server {
	return &Server{
		gitops:         forge,
		deployer:       deployerSvc,
		workloadDomain: workloadDomain,
	}
}

// SetDeployer wires the deployer after construction.
func (s *Server) SetDeployer(d DeployerService) {
	s.deployer = d
}

// syncEnvironment triggers an ArgoCD sync for a single environment.
// Best-effort: logs on failure but never returns an error.
func (s *Server) syncEnvironment(ctx context.Context, workspace, project, environment string) {
	if _, err := s.deployer.SyncDeployment(ctx, workspace, project, environment); err != nil {
		slog.Warn("failed to trigger sync", "project", project, "environment", environment, "error", err)
		return
	}
	slog.Info("triggered ArgoCD sync", "project", project, "environment", environment)
}

// syncAllEnvironments triggers an ArgoCD sync for every environment in a project.
// Used after base-level changes (services, databases, chart) that affect all environments.
func (s *Server) syncAllEnvironments(ctx context.Context, workspace, project string, environments []string) {
	for _, env := range environments {
		s.syncEnvironment(ctx, workspace, project, env)
	}
}

// InitProject creates a new GitOps repository for a project.
func (s *Server) InitProject(ctx context.Context, workspace, project, displayName string) (gitopsRepoURL string, err error) {
	slog.Info("InitProject called", "project", project)
	repoURL, err := s.gitops.CreateRepo(ctx, project, workspace, displayName)
	if err != nil {
		return "", fmt.Errorf("failed to init project: %w", err)
	}
	return repoURL, nil
}

// ListProjects returns every project owned by the given workspace.
func (s *Server) ListProjects(ctx context.Context, workspace string) ([]data.ProjectInfo, error) {
	slog.Info("ListProjects called", "workspace", workspace)

	repos, err := s.gitops.Repos(ctx, workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	var infos []data.ProjectInfo
	for _, proj := range repos {
		repo, err := s.gitops.CloneRepo(ctx, proj.HTTPURL)
		if err != nil {
			slog.Warn("failed to clone repo", "repo", proj.Slug, "error", err)
			continue
		}
		repo.SetWorkspace(workspace)
		defer repo.Cleanup()

		meta, err := repo.Metadata(ctx)
		if err != nil {
			slog.Warn("failed to read repo metadata", "repo", proj.Slug, "error", err)
			continue
		}

		infos = append(infos, data.ProjectInfo{
			Name:             meta.Name,
			DisplayName:      proj.DisplayName,
			GitopsRepoURL:    meta.RepoURL,
			Environments:     meta.Environments,
			EnvironmentInfos: envInfosFromMeta(meta.EnvironmentInfos),
			CreatedAt:        meta.CreatedAt,
			Databases:        databaseInfosFromDefs(meta.Databases),
		})
	}
	return infos, nil
}

// GetProject returns a single project by name.
func (s *Server) GetProject(ctx context.Context, workspace, project string) (data.ProjectInfo, error) {
	slog.Info("GetProject called", "project", project)

	entry, err := s.gitops.Repo(ctx, project, workspace)
	if err != nil {
		return data.ProjectInfo{}, fmt.Errorf("failed to get project: %w", err)
	}

	repo, err := s.gitops.CloneRepo(ctx, entry.HTTPURL)
	if err != nil {
		return data.ProjectInfo{}, err
	}
	repo.SetWorkspace(workspace)
	defer repo.Cleanup()

	meta, err := repo.Metadata(ctx)
	if err != nil {
		return data.ProjectInfo{}, err
	}

	return data.ProjectInfo{
		Name:             meta.Name,
		DisplayName:      entry.DisplayName,
		GitopsRepoURL:    meta.RepoURL,
		Environments:     meta.Environments,
		EnvironmentInfos: envInfosFromMeta(meta.EnvironmentInfos),
		CreatedAt:        meta.CreatedAt,
		Databases:        databaseInfosFromDefs(meta.Databases),
	}, nil
}

// DeleteProject removes a project's GitOps repository.
func (s *Server) DeleteProject(ctx context.Context, workspace, project string) error {
	slog.Info("DeleteProject called", "project", project)
	if err := s.gitops.DeleteRepo(ctx, project, workspace); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	return nil
}

// AddService writes a service definition to base values and stamps the
// initial image tag into the target environment.
func (s *Server) AddService(ctx context.Context, workspace, project string, def data.ServiceDef) error {
	slog.Info("AddService called", "project", project, "service", def.Name, "environment", def.Environment, "image", def.Image)

	repo, err := s.cloneRepo(ctx, workspace, project)
	if err != nil {
		return err
	}
	defer repo.Cleanup()

	if err := repo.AddService(ctx, def.Environment, gitops.ServiceDef{
		Name:                 def.Name,
		Image:                def.Image,
		Port:                 def.Port,
		Framework:            def.Framework,
		SourceURL:            def.SourceURL,
		ContextPath:          def.ContextPath,
		GitHubInstallationID: def.GitHubInstallationID,
		ImageTag:             def.ImageTag,
		ImagePullPolicy:      def.ImagePullPolicy,
		CustomStartCommand:   def.CustomStartCommand,
		StartCommand:         def.StartCommand,
	}); err != nil {
		return fmt.Errorf("failed to add service: %w", err)
	}

	s.syncEnvironment(ctx, workspace, project, def.Environment)
	return nil
}

// RemoveService removes a service from an environment's values.
func (s *Server) RemoveService(ctx context.Context, workspace, project, environment, service string) error {
	slog.Info("RemoveService called", "project", project, "service", service, "environment", environment)

	repo, err := s.cloneRepo(ctx, workspace, project)
	if err != nil {
		return err
	}
	defer repo.Cleanup()

	if err := repo.RemoveService(ctx, environment, service); err != nil {
		return fmt.Errorf("failed to remove service: %w", err)
	}
	s.syncEnvironment(ctx, workspace, project, environment)
	return nil
}

// UpdateImageTag stamps a new image tag onto a service in one environment.
func (s *Server) UpdateImageTag(ctx context.Context, workspace, project, environment, service, tag, digest, commitPrefix string) error {
	slog.Info("UpdateImageTag called", "project", project, "environment", environment, "service", service, "tag", tag)

	repo, err := s.cloneRepo(ctx, workspace, project)
	if err != nil {
		return err
	}
	defer repo.Cleanup()

	if err := repo.UpdateImageTag(ctx, environment, service, tag, digest, commitPrefix); err != nil {
		return fmt.Errorf("failed to update image tag: %w", err)
	}
	s.syncEnvironment(ctx, workspace, project, environment)
	return nil
}

// CreateEnvironment creates a new environment in a project's GitOps repo.
// Returns the workload namespace name.
func (s *Server) CreateEnvironment(ctx context.Context, workspace, project, environment, fromEnvironment string) (namespace string, err error) {
	slog.Info("CreateEnvironment called", "project", project, "environment", environment, "from", fromEnvironment)

	repo, err := s.cloneRepo(ctx, workspace, project)
	if err != nil {
		return "", err
	}
	defer repo.Cleanup()

	if err := repo.CreateEnvironment(ctx, environment, fromEnvironment); err != nil {
		return "", fmt.Errorf("failed to create environment: %w", err)
	}

	return gitops.NamespaceFor(workspace, project, environment), nil
}

// DeleteEnvironment removes an environment from a project's GitOps repo.
func (s *Server) DeleteEnvironment(ctx context.Context, workspace, project, environment string) error {
	slog.Info("DeleteEnvironment called", "project", project, "environment", environment)

	repo, err := s.cloneRepo(ctx, workspace, project)
	if err != nil {
		return err
	}
	defer repo.Cleanup()

	if err := repo.DeleteEnvironment(ctx, environment); err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}
	return nil
}

// Promote copies a service's image tag from one environment to another.
func (s *Server) Promote(ctx context.Context, workspace, project, service, fromEnvironment, toEnvironment string) (imageTag string, err error) {
	slog.Info("Promote called", "project", project, "service", service, "from", fromEnvironment, "to", toEnvironment)

	repo, err := s.cloneRepo(ctx, workspace, project)
	if err != nil {
		return "", err
	}
	defer repo.Cleanup()

	imageTag, err = repo.Promote(ctx, service, fromEnvironment, toEnvironment)
	if err != nil {
		return "", fmt.Errorf("failed to promote: %w", err)
	}

	s.syncEnvironment(ctx, workspace, project, toEnvironment)
	return imageTag, nil
}

// DeploymentHistory returns a service's deployment history from git log.
func (s *Server) DeploymentHistory(ctx context.Context, workspace, project, environment, service string) ([]data.DeploymentEntry, error) {
	slog.Info("DeploymentHistory called", "project", project, "environment", environment, "service", service)

	repo, err := s.cloneRepo(ctx, workspace, project)
	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	entries, err := repo.DeploymentHistory(ctx, environment, service)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment history: %w", err)
	}

	out := make([]data.DeploymentEntry, 0, len(entries))
	for _, e := range entries {
		out = append(out, data.DeploymentEntry{
			ImageTag:   e.ImageTag,
			Revision:   e.Revision,
			DeployedAt: e.Timestamp,
			Author:     e.Author,
		})
	}
	return out, nil
}

// GeneratePlatformDomain generates a unique *.{workloadDomain} hostname,
// adds it to a service in an environment, and returns it.
func (s *Server) GeneratePlatformDomain(ctx context.Context, workspace, project, environment, service string) (hostname string, err error) {
	slog.Info("GeneratePlatformDomain called", "project", project, "environment", environment, "service", service)

	hostname, err = s.generatePlatformDomain(ctx, service, environment)
	if err != nil {
		return "", fmt.Errorf("failed to generate unique platform domain: %w", err)
	}

	repo, err := s.cloneRepo(ctx, workspace, project)
	if err != nil {
		return "", err
	}
	defer repo.Cleanup()

	if err := repo.AddDomain(ctx, environment, service, hostname); err != nil {
		return "", fmt.Errorf("failed to add platform domain: %w", err)
	}

	s.syncEnvironment(ctx, workspace, project, environment)
	return hostname, nil
}

// AddDomain adds a domain hostname to a service in an environment.
func (s *Server) AddDomain(ctx context.Context, workspace, project, environment, service, hostname string) error {
	slog.Info("AddDomain called", "project", project, "environment", environment, "service", service, "hostname", hostname)

	repo, err := s.cloneRepo(ctx, workspace, project)
	if err != nil {
		return err
	}
	defer repo.Cleanup()

	if err := repo.AddDomain(ctx, environment, service, hostname); err != nil {
		return fmt.Errorf("failed to add domain: %w", err)
	}
	s.syncEnvironment(ctx, workspace, project, environment)
	return nil
}

// RemoveDomain removes a domain hostname from a service in an environment.
func (s *Server) RemoveDomain(ctx context.Context, workspace, project, environment, service, hostname string) error {
	slog.Info("RemoveDomain called", "project", project, "environment", environment, "service", service, "hostname", hostname)

	repo, err := s.cloneRepo(ctx, workspace, project)
	if err != nil {
		return err
	}
	defer repo.Cleanup()

	if err := repo.RemoveDomain(ctx, environment, service, hostname); err != nil {
		return fmt.Errorf("failed to remove domain: %w", err)
	}
	s.syncEnvironment(ctx, workspace, project, environment)
	return nil
}

// Eject builds a tar.gz archive of the project's complete configuration
// for independent operation.
func (s *Server) Eject(ctx context.Context, workspace, project string) ([]byte, error) {
	slog.Info("eject started", "project", project)

	repo, err := s.cloneRepo(ctx, workspace, project)
	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	repoFiles, err := repo.Files(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read repo files: %w", err)
	}

	archive, err := eject.Build(ctx, workspace, repoFiles, project)
	if err != nil {
		return nil, fmt.Errorf("failed to build ejection archive: %w", err)
	}

	slog.Info("eject completed", "project", project, "size", len(archive))
	return archive, nil
}

// SharedVariables returns all shared variables for an environment.
func (s *Server) SharedVariables(ctx context.Context, workspace, project, environment string) (map[string]string, error) {
	slog.Info("SharedVariables called", "project", project, "environment", environment)

	repo, err := s.cloneRepo(ctx, workspace, project)
	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	vars, err := repo.SharedVariables(ctx, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared variables: %w", err)
	}
	return vars, nil
}

// SetSharedVariables replaces all shared variables for an environment.
func (s *Server) SetSharedVariables(ctx context.Context, workspace, project, environment string, vars map[string]string) error {
	slog.Info("SetSharedVariables called", "project", project, "environment", environment)

	repo, err := s.cloneRepo(ctx, workspace, project)
	if err != nil {
		return err
	}
	defer repo.Cleanup()

	if err := repo.SetSharedVariables(ctx, environment, vars); err != nil {
		return fmt.Errorf("failed to set shared variables: %w", err)
	}
	s.syncEnvironment(ctx, workspace, project, environment)
	return nil
}

// ServiceVariables returns a service's variables, shared refs, and
// database/service refs in one round-trip.
type ServiceVariables struct {
	Variables    map[string]string
	SharedRefs   []string
	DatabaseRefs map[string]data.DatabaseRef
	ServiceRefs  map[string]data.ServiceRef
}

func (s *Server) ServiceVariables(ctx context.Context, workspace, project, environment, service string) (ServiceVariables, error) {
	slog.Info("ServiceVariables called", "project", project, "environment", environment, "service", service)

	repo, err := s.cloneRepo(ctx, workspace, project)
	if err != nil {
		return ServiceVariables{}, err
	}
	defer repo.Cleanup()

	vars, refs, dbRefs, svcRefs, err := repo.ServiceVariables(ctx, environment, service)
	if err != nil {
		return ServiceVariables{}, fmt.Errorf("failed to get service variables: %w", err)
	}

	return ServiceVariables{
		Variables:    vars,
		SharedRefs:   refs,
		DatabaseRefs: databaseRefsFromGitops(dbRefs),
		ServiceRefs:  serviceRefsFromGitops(svcRefs),
	}, nil
}

// SetServiceVariables replaces a service's environment variables in one environment.
func (s *Server) SetServiceVariables(ctx context.Context, workspace, project, environment, service string, vars map[string]string, sharedRefs []string, dbRefs map[string]data.DatabaseRef, svcRefs map[string]data.ServiceRef) error {
	slog.Info("SetServiceVariables called", "project", project, "environment", environment, "service", service)

	repo, err := s.cloneRepo(ctx, workspace, project)
	if err != nil {
		return err
	}
	defer repo.Cleanup()

	if err := repo.SetServiceVariables(ctx, environment, service, vars, sharedRefs, databaseRefsToGitops(dbRefs), serviceRefsToGitops(svcRefs)); err != nil {
		return fmt.Errorf("failed to set service variables: %w", err)
	}
	s.syncEnvironment(ctx, workspace, project, environment)
	return nil
}

// AddDatabase adds a PostgreSQL database to a project's base values.
func (s *Server) AddDatabase(ctx context.Context, workspace, project string, db data.DatabaseInfo) error {
	slog.Info("AddDatabase called", "project", project, "database", db.Name)

	if db.Version == "" {
		db.Version = "16"
	}
	if db.Instances == 0 {
		db.Instances = 1
	}
	if db.Size == "" {
		db.Size = "10Gi"
	}

	repo, err := s.cloneRepo(ctx, workspace, project)
	if err != nil {
		return err
	}
	defer repo.Cleanup()

	if err := repo.AddDatabase(ctx, gitops.DatabaseDef{
		Name:      db.Name,
		Version:   db.Version,
		Instances: db.Instances,
		Size:      db.Size,
	}); err != nil {
		return fmt.Errorf("failed to add database: %w", err)
	}

	meta, err := repo.Metadata(ctx)
	if err != nil {
		slog.Warn("failed to read repo metadata for sync", "project", project, "error", err)
	}
	s.syncAllEnvironments(ctx, workspace, project, meta.Environments)
	return nil
}

// RemoveDatabase removes a database from a project's base values.
func (s *Server) RemoveDatabase(ctx context.Context, workspace, project, name string) error {
	slog.Info("RemoveDatabase called", "project", project, "database", name)

	repo, err := s.cloneRepo(ctx, workspace, project)
	if err != nil {
		return err
	}
	defer repo.Cleanup()

	if err := repo.RemoveDatabase(ctx, name); err != nil {
		return fmt.Errorf("failed to remove database: %w", err)
	}

	meta, err := repo.Metadata(ctx)
	if err != nil {
		slog.Warn("failed to read repo metadata for sync", "project", project, "error", err)
	}
	s.syncAllEnvironments(ctx, workspace, project, meta.Environments)
	return nil
}

// SetResources writes resource requests/limits to an environment's values.
func (s *Server) SetResources(ctx context.Context, workspace, project, environment, tier string, cpuMillicores, memoryMB, diskMB int) error {
	repo, err := s.cloneRepo(ctx, workspace, project)
	if err != nil {
		return err
	}
	defer repo.Cleanup()

	if err := repo.SetResources(ctx, environment, tier, cpuMillicores, memoryMB, diskMB); err != nil {
		return fmt.Errorf("failed to set resources: %w", err)
	}
	s.syncEnvironment(ctx, workspace, project, environment)
	return nil
}

// SetServiceScaling writes replica count + autoscaling config for a service.
func (s *Server) SetServiceScaling(ctx context.Context, workspace, project, environment, service string, replicas int, autoscaling *data.AutoscalingConfig) error {
	var as *gitops.AutoscalingConfig
	if autoscaling != nil && autoscaling.Enabled {
		as = &gitops.AutoscalingConfig{
			Enabled:     true,
			MinReplicas: autoscaling.MinReplicas,
			MaxReplicas: autoscaling.MaxReplicas,
			TargetCPU:   autoscaling.TargetCPU,
		}
	}

	repo, err := s.cloneRepo(ctx, workspace, project)
	if err != nil {
		return err
	}
	defer repo.Cleanup()

	if err := repo.SetServiceScaling(ctx, environment, service, replicas, as); err != nil {
		return fmt.Errorf("failed to set service scaling: %w", err)
	}
	s.syncEnvironment(ctx, workspace, project, environment)
	return nil
}

// SetCustomStartCommand sets or clears the custom start command for a service.
func (s *Server) SetCustomStartCommand(ctx context.Context, workspace, project, environment, service, command string) error {
	slog.Info("SetCustomStartCommand called", "project", project, "service", service, "environment", environment, "command", command)

	repo, err := s.cloneRepo(ctx, workspace, project)
	if err != nil {
		return err
	}
	defer repo.Cleanup()

	if err := repo.SetCustomStartCommand(ctx, environment, service, command); err != nil {
		return fmt.Errorf("failed to set custom start command: %w", err)
	}
	s.syncEnvironment(ctx, workspace, project, environment)
	return nil
}

// SetSuspended writes or removes the suspended flag in an environment's values.
func (s *Server) SetSuspended(ctx context.Context, workspace, project, environment string, suspended bool) error {
	repo, err := s.cloneRepo(ctx, workspace, project)
	if err != nil {
		return err
	}
	defer repo.Cleanup()

	if err := repo.SetSuspended(ctx, environment, suspended); err != nil {
		return fmt.Errorf("failed to set suspended: %w", err)
	}
	s.syncEnvironment(ctx, workspace, project, environment)
	return nil
}

// generatePlatformDomain finds a unique *.{workloadDomain} hostname.
// Walks repos across all workspaces — hostnames are globally unique.
func (s *Server) generatePlatformDomain(ctx context.Context, service, environment string) (string, error) {
	repos, err := s.gitops.Repos(ctx, "")
	if err != nil {
		return "", err
	}

	allDomains := make(map[string]bool)
	// TODO: This is very inefficient.
	for _, entry := range repos {
		repo, err := s.gitops.CloneRepo(ctx, entry.HTTPURL)
		if err != nil {
			slog.Warn("skipping repo", "repo", entry.Slug, "error", err)
			continue
		}
		defer repo.Cleanup()

		domains, err := repo.Domains(ctx)
		if err != nil {
			slog.Warn("failed to get domains", "repo", entry.Slug, "error", err)
			continue
		}
		for _, d := range domains {
			allDomains[d] = true
		}
	}

	for i := 0; i < 10; i++ {
		hostname := fmt.Sprintf("%s-%s-%s.%s", service, environment, randCrockford32(5), s.workloadDomain)
		if !allDomains[hostname] {
			return hostname, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique platform domain for service %s in environment %s", service, environment)
}

func (s *Server) cloneRepo(ctx context.Context, workspace, project string) (gitops.Repository, error) {
	entry, err := s.gitops.Repo(ctx, project, workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	repo, err := s.gitops.CloneRepo(ctx, entry.HTTPURL)
	if err != nil {
		return nil, err
	}
	repo.SetWorkspace(workspace)
	return repo, nil
}

func envInfosFromMeta(metas []gitops.EnvironmentMeta) []data.EnvironmentInfo {
	if len(metas) == 0 {
		return nil
	}
	out := make([]data.EnvironmentInfo, len(metas))
	for i, m := range metas {
		var svcs []data.ServiceInstanceInfo
		for _, s := range m.Services {
			svcs = append(svcs, data.ServiceInstanceInfo{
				Name:                 s.Name,
				ImageTag:             s.ImageTag,
				Domains:              s.Domains,
				Image:                s.Image,
				Port:                 s.Port,
				Framework:            s.Framework,
				SourceURL:            s.SourceURL,
				ContextPath:          s.ContextPath,
				GitHubInstallationID: s.GitHubInstallationID,
				CustomStartCommand:   s.CustomStartCommand,
				StartCommand:         s.StartCommand,
			})
		}
		out[i] = data.EnvironmentInfo{
			Name:     m.Name,
			Services: svcs,
		}
	}
	return out
}

func databaseInfosFromDefs(defs []gitops.DatabaseDef) []data.DatabaseInfo {
	if len(defs) == 0 {
		return nil
	}
	out := make([]data.DatabaseInfo, len(defs))
	for i, d := range defs {
		out[i] = data.DatabaseInfo{
			Name:      d.Name,
			Version:   d.Version,
			Instances: d.Instances,
			Size:      d.Size,
		}
	}
	return out
}

func databaseRefsFromGitops(refs map[string]gitops.DatabaseRef) map[string]data.DatabaseRef {
	if len(refs) == 0 {
		return nil
	}
	out := make(map[string]data.DatabaseRef, len(refs))
	for k, v := range refs {
		out[k] = data.DatabaseRef{Database: v.Database, Key: v.Key}
	}
	return out
}

func serviceRefsFromGitops(refs map[string]gitops.ServiceRef) map[string]data.ServiceRef {
	if len(refs) == 0 {
		return nil
	}
	out := make(map[string]data.ServiceRef, len(refs))
	for k, v := range refs {
		out[k] = data.ServiceRef{Service: v.Service}
	}
	return out
}

func databaseRefsToGitops(refs map[string]data.DatabaseRef) map[string]gitops.DatabaseRef {
	if len(refs) == 0 {
		return nil
	}
	out := make(map[string]gitops.DatabaseRef, len(refs))
	for k, v := range refs {
		out[k] = gitops.DatabaseRef{Database: v.Database, Key: v.Key}
	}
	return out
}

func serviceRefsToGitops(refs map[string]data.ServiceRef) map[string]gitops.ServiceRef {
	if len(refs) == 0 {
		return nil
	}
	out := make(map[string]gitops.ServiceRef, len(refs))
	for k, v := range refs {
		out[k] = gitops.ServiceRef{Service: v.Service}
	}
	return out
}

func randCrockford32(n int) string {
	const crockford32 = "0123456789abcdefghjkmnpqrstvwxyz"

	b := make([]byte, n)
	rand.Read(b) // rand.Read never returns an error
	for i, v := range b {
		b[i] = crockford32[v&31]
	}
	return string(b)
}
