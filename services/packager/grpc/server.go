package grpc

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/packager"
	"github.com/zeitlos/lucity/pkg/tenant"
	"github.com/zeitlos/lucity/services/packager/eject"
	"github.com/zeitlos/lucity/services/packager/gitops"
)

type Server struct {
	packager.UnimplementedPackagerServiceServer
	gitops   gitops.Forge
	deployer deployer.DeployerServiceClient
	issuer   *auth.Issuer

	workloadDomain string
}

// NewServer creates a packager server with the given GitOps provider.
func NewServer(forge gitops.Forge, deployerClient deployer.DeployerServiceClient, issuer *auth.Issuer, workloadDomain string) *Server {
	return &Server{
		gitops:         forge,
		deployer:       deployerClient,
		issuer:         issuer,
		workloadDomain: workloadDomain,
	}
}

// syncEnvironment triggers an ArgoCD sync for a single environment.
// Best-effort: logs on failure but never returns an error.
func (s *Server) syncEnvironment(ctx context.Context, project, environment string) {
	if s.issuer != nil {
		ctx = auth.WithClaims(ctx, &auth.Claims{
			Subject: "packager",
			Roles:   []auth.Role{auth.RoleUser},
		})
		ctx = auth.WithIssuer(ctx, s.issuer)
		ctx = auth.OutgoingContext(ctx)
	}
	ctx = tenant.OutgoingContext(ctx)
	_, err := s.deployer.SyncDeployment(ctx, &deployer.SyncDeploymentRequest{
		Project:     project,
		Environment: environment,
	})
	if err != nil {
		slog.Warn("failed to trigger sync", "project", project, "environment", environment, "error", err)
		return
	}
	slog.Info("triggered ArgoCD sync", "project", project, "environment", environment)
}

// syncAllEnvironments triggers an ArgoCD sync for every environment in a project.
// Used after base-level changes (services, databases, chart) that affect all environments.
func (s *Server) syncAllEnvironments(ctx context.Context, project string, environments []string) {
	for _, env := range environments {
		s.syncEnvironment(ctx, project, env)
	}
}

func (s *Server) InitProject(ctx context.Context, req *packager.InitProjectRequest) (*packager.InitProjectResponse, error) {
	slog.Info("InitProject called", "project", req.Project)

	repoURL, err := s.gitops.CreateRepo(ctx, req.Project, tenant.FromContext(ctx), req.DisplayName)

	if err != nil {
		return nil, fmt.Errorf("failed to init project: %w", err)
	}

	return &packager.InitProjectResponse{
		GitopsRepoUrl: repoURL,
	}, nil
}

func (s *Server) ListProjects(ctx context.Context, req *packager.ListProjectsRequest) (*packager.ListProjectsResponse, error) {
	slog.Info("ListProjects called")

	workspace := tenant.FromContext(ctx)
	repos, err := s.gitops.Repos(ctx, workspace)

	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	var infos []*packager.ProjectInfo
	for _, proj := range repos {
		repo, err := s.gitops.CloneRepo(ctx, proj.HTTPURL)

		if err != nil {
			slog.Warn("failed to clone repo", "repo", proj.Slug, "error", err)
			continue
		}
		defer repo.Cleanup()

		meta, err := repo.Metadata(ctx)

		if err != nil {
			slog.Warn("failed to read repo metadata", "repo", proj.Slug, "error", err)
			continue
		}

		infos = append(infos, &packager.ProjectInfo{
			Name:             meta.Name,
			DisplayName:      proj.DisplayName,
			GitopsRepoUrl:    meta.RepoURL,
			Environments:     meta.Environments,
			EnvironmentInfos: envInfosFromMeta(meta.EnvironmentInfos),
			CreatedAt:        timestamppb.New(meta.CreatedAt),
			Databases:        databaseInfosFromDefs(meta.Databases),
		})
	}

	return &packager.ListProjectsResponse{Projects: infos}, nil
}

func (s *Server) GetProject(ctx context.Context, req *packager.GetProjectRequest) (*packager.GetProjectResponse, error) {
	slog.Info("GetProject called", "project", req.Project)

	entry, err := s.gitops.Repo(ctx, req.Project, tenant.FromContext(ctx))

	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	repo, err := s.gitops.CloneRepo(ctx, entry.HTTPURL)

	if err != nil {
		return nil, err
	}

	meta, err := repo.Metadata(ctx)

	if err != nil {
		return nil, err
	}

	return &packager.GetProjectResponse{
		Project: &packager.ProjectInfo{
			Name:             meta.Name,
			DisplayName:      entry.DisplayName,
			GitopsRepoUrl:    meta.RepoURL,
			Environments:     meta.Environments,
			EnvironmentInfos: envInfosFromMeta(meta.EnvironmentInfos),
			CreatedAt:        timestamppb.New(meta.CreatedAt),
			Databases:        databaseInfosFromDefs(meta.Databases),
		},
	}, nil
}

func (s *Server) DeleteProject(ctx context.Context, req *packager.DeleteProjectRequest) (*packager.DeleteProjectResponse, error) {
	slog.Info("DeleteProject called", "project", req.Project)

	if err := s.gitops.DeleteRepo(ctx, req.Project, tenant.FromContext(ctx)); err != nil {
		return nil, fmt.Errorf("failed to delete project: %w", err)
	}

	return &packager.DeleteProjectResponse{}, nil
}

func (s *Server) AddService(ctx context.Context, req *packager.AddServiceRequest) (*packager.AddServiceResponse, error) {
	slog.Info("AddService called", "project", req.Project, "service", req.Service, "environment", req.Environment, "image", req.Image)

	repo, err := s.cloneRepo(ctx, req.Project)

	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	if err := repo.AddService(ctx, req.Environment, gitops.ServiceDef{
		Name:                 req.Service,
		Image:                req.Image,
		Port:                 int(req.Port),
		Framework:            req.Framework,
		SourceURL:            req.SourceUrl,
		ContextPath:          req.ContextPath,
		GitHubInstallationID: req.GithubInstallationId,
		ImageTag:             req.ImageTag,
		ImagePullPolicy:      req.ImagePullPolicy,
		CustomStartCommand:   req.CustomStartCommand,
		StartCommand:         req.StartCommand,
	}); err != nil {
		return nil, fmt.Errorf("failed to add service: %w", err)
	}

	s.syncEnvironment(ctx, req.Project, req.Environment)
	return &packager.AddServiceResponse{}, nil
}

func (s *Server) RemoveService(ctx context.Context, req *packager.RemoveServiceRequest) (*packager.RemoveServiceResponse, error) {
	slog.Info("RemoveService called", "project", req.Project, "service", req.Service, "environment", req.Environment)

	repo, err := s.cloneRepo(ctx, req.Project)

	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	if err := repo.RemoveService(ctx, req.Environment, req.Service); err != nil {
		return nil, fmt.Errorf("failed to remove service: %w", err)
	}

	s.syncEnvironment(ctx, req.Project, req.Environment)
	return &packager.RemoveServiceResponse{}, nil
}

func (s *Server) UpdateImageTag(ctx context.Context, req *packager.UpdateImageTagRequest) (*packager.UpdateImageTagResponse, error) {
	slog.Info("UpdateImageTag called", "project", req.Project, "environment", req.Environment, "service", req.Service, "tag", req.Tag)

	repo, err := s.cloneRepo(ctx, req.Project)

	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	if err := repo.UpdateImageTag(ctx, req.Environment, req.Service, req.Tag, req.Digest, req.CommitPrefix); err != nil {
		return nil, fmt.Errorf("failed to update image tag: %w", err)
	}

	s.syncEnvironment(ctx, req.Project, req.Environment)
	return &packager.UpdateImageTagResponse{}, nil
}

func (s *Server) CreateEnvironment(ctx context.Context, req *packager.CreateEnvironmentRequest) (*packager.CreateEnvironmentResponse, error) {
	slog.Info("CreateEnvironment called", "project", req.Project, "environment", req.Environment, "from", req.FromEnvironment)

	repo, err := s.cloneRepo(ctx, req.Project)

	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	if err := repo.CreateEnvironment(ctx, req.Environment, req.FromEnvironment); err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	ws := tenant.FromContext(ctx)
	return &packager.CreateEnvironmentResponse{
		Namespace: gitops.NamespaceFor(ws, req.Project, req.Environment),
	}, nil
}

func (s *Server) DeleteEnvironment(ctx context.Context, req *packager.DeleteEnvironmentRequest) (*packager.DeleteEnvironmentResponse, error) {
	slog.Info("DeleteEnvironment called", "project", req.Project, "environment", req.Environment)

	repo, err := s.cloneRepo(ctx, req.Project)

	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	if err := repo.DeleteEnvironment(ctx, req.Environment); err != nil {
		return nil, fmt.Errorf("failed to delete environment: %w", err)
	}

	return &packager.DeleteEnvironmentResponse{}, nil
}

func (s *Server) Promote(ctx context.Context, req *packager.PromoteRequest) (*packager.PromoteResponse, error) {
	slog.Info("Promote called", "project", req.Project, "service", req.Service, "from", req.FromEnvironment, "to", req.ToEnvironment)

	repo, err := s.cloneRepo(ctx, req.Project)

	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	imageTag, err := repo.Promote(ctx, req.Service, req.FromEnvironment, req.ToEnvironment)

	if err != nil {
		return nil, fmt.Errorf("failed to promote: %w", err)
	}

	s.syncEnvironment(ctx, req.Project, req.ToEnvironment)
	return &packager.PromoteResponse{
		ImageTag: imageTag,
	}, nil
}

func (s *Server) DeploymentHistory(ctx context.Context, req *packager.DeploymentHistoryRequest) (*packager.DeploymentHistoryResponse, error) {
	slog.Info("DeploymentHistory called", "project", req.Project, "environment", req.Environment, "service", req.Service)

	repo, err := s.cloneRepo(ctx, req.Project)

	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	entries, err := repo.DeploymentHistory(ctx, req.Environment, req.Service)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment history: %w", err)
	}

	var protoEntries []*packager.DeploymentHistoryEntry
	for _, e := range entries {
		protoEntries = append(protoEntries, &packager.DeploymentHistoryEntry{
			ImageTag:   e.ImageTag,
			Revision:   e.Revision,
			DeployedAt: timestamppb.New(e.Timestamp),
			Author:     e.Author,
		})
	}

	return &packager.DeploymentHistoryResponse{Entries: protoEntries}, nil
}

func (s *Server) GeneratePlatformDomain(ctx context.Context, req *packager.GeneratePlatformDomainRequest) (*packager.GeneratePlatformDomainResponse, error) {
	slog.Info("GeneratePlatformDomain called", "project", req.Project, "environment", req.Environment, "service", req.Service)

	hostname, err := s.generatePlatformDomain(ctx, req.Service, req.Environment)

	if err != nil {
		return nil, fmt.Errorf("failed to generate unique platform domain: %w", err)
	}

	repo, err := s.cloneRepo(ctx, req.Project)

	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	if err := repo.AddDomain(ctx, req.Environment, req.Service, hostname); err != nil {
		return nil, fmt.Errorf("failed to add platform domain: %w", err)
	}

	s.syncEnvironment(ctx, req.Project, req.Environment)

	return &packager.GeneratePlatformDomainResponse{
		Hostname: hostname,
	}, nil
}

func (s *Server) AddDomain(ctx context.Context, req *packager.AddDomainRequest) (*packager.AddDomainResponse, error) {
	slog.Info("AddDomain called", "project", req.Project, "environment", req.Environment, "service", req.Service, "hostname", req.Hostname)

	repo, err := s.cloneRepo(ctx, req.Project)

	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	if err := repo.AddDomain(ctx, req.Environment, req.Service, req.Hostname); err != nil {
		return nil, fmt.Errorf("failed to add domain: %w", err)
	}

	s.syncEnvironment(ctx, req.Project, req.Environment)

	return &packager.AddDomainResponse{}, nil
}

func (s *Server) RemoveDomain(ctx context.Context, req *packager.RemoveDomainRequest) (*packager.RemoveDomainResponse, error) {
	slog.Info("RemoveDomain called", "project", req.Project, "environment", req.Environment, "service", req.Service, "hostname", req.Hostname)

	repo, err := s.cloneRepo(ctx, req.Project)

	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	if err := repo.RemoveDomain(ctx, req.Environment, req.Service, req.Hostname); err != nil {
		return nil, fmt.Errorf("failed to remove domain: %w", err)
	}

	s.syncEnvironment(ctx, req.Project, req.Environment)
	return &packager.RemoveDomainResponse{}, nil
}

func (s *Server) Eject(ctx context.Context, req *packager.EjectRequest) (*packager.EjectResponse, error) {
	slog.Info("eject started", "project", req.Project)

	repo, err := s.cloneRepo(ctx, req.Project)

	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	repoFiles, err := repo.Files(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to read repo files: %w", err)
	}

	archive, err := eject.Build(ctx, repoFiles, req.Project)
	if err != nil {
		return nil, fmt.Errorf("failed to build ejection archive: %w", err)
	}

	slog.Info("eject completed", "project", req.Project, "size", len(archive))
	return &packager.EjectResponse{Archive: archive}, nil
}

func (s *Server) SharedVariables(ctx context.Context, req *packager.SharedVariablesRequest) (*packager.SharedVariablesResponse, error) {
	slog.Info("SharedVariables called", "project", req.Project, "environment", req.Environment)

	repo, err := s.cloneRepo(ctx, req.Project)

	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	vars, err := repo.SharedVariables(ctx, req.Environment)

	if err != nil {
		return nil, fmt.Errorf("failed to get shared variables: %w", err)
	}

	return &packager.SharedVariablesResponse{Variables: vars}, nil
}

func (s *Server) SetSharedVariables(ctx context.Context, req *packager.SetSharedVariablesRequest) (*packager.SetSharedVariablesResponse, error) {
	slog.Info("SetSharedVariables called", "project", req.Project, "environment", req.Environment)

	repo, err := s.cloneRepo(ctx, req.Project)

	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	if err := repo.SetSharedVariables(ctx, req.Environment, req.Variables); err != nil {
		return nil, fmt.Errorf("failed to set shared variables: %w", err)
	}

	s.syncEnvironment(ctx, req.Project, req.Environment)
	return &packager.SetSharedVariablesResponse{}, nil
}

func (s *Server) ServiceVariables(ctx context.Context, req *packager.ServiceVariablesRequest) (*packager.ServiceVariablesResponse, error) {
	slog.Info("ServiceVariables called", "project", req.Project, "environment", req.Environment, "service", req.Service)

	repo, err := s.cloneRepo(ctx, req.Project)

	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	vars, refs, databaseRefs, serviceRefs, err := repo.ServiceVariables(ctx, req.Environment, req.Service)

	if err != nil {
		return nil, fmt.Errorf("failed to get service variables: %w", err)
	}

	return &packager.ServiceVariablesResponse{
		Variables:    vars,
		SharedRefs:   refs,
		DatabaseRefs: databaseRefsToProto(databaseRefs),
		ServiceRefs:  serviceRefsToProto(serviceRefs),
	}, nil
}

func (s *Server) SetServiceVariables(ctx context.Context, req *packager.SetServiceVariablesRequest) (*packager.SetServiceVariablesResponse, error) {
	slog.Info("SetServiceVariables called", "project", req.Project, "environment", req.Environment, "service", req.Service)

	repo, err := s.cloneRepo(ctx, req.Project)

	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	if err := repo.SetServiceVariables(ctx, req.Environment, req.Service, req.Variables, req.SharedRefs, databaseRefsFromProto(req.DatabaseRefs), serviceRefsFromProto(req.ServiceRefs)); err != nil {
		return nil, fmt.Errorf("failed to set service variables: %w", err)
	}

	s.syncEnvironment(ctx, req.Project, req.Environment)
	return &packager.SetServiceVariablesResponse{}, nil
}

func envInfosFromMeta(metas []gitops.EnvironmentMeta) []*packager.EnvironmentInfo {
	if len(metas) == 0 {
		return nil
	}
	infos := make([]*packager.EnvironmentInfo, len(metas))
	for i, m := range metas {
		var svcs []*packager.ServiceInstanceInfo
		for _, s := range m.Services {
			svcs = append(svcs, &packager.ServiceInstanceInfo{
				Name:                 s.Name,
				ImageTag:             s.ImageTag,
				Domains:              s.Domains,
				Image:                s.Image,
				Port:                 int32(s.Port),
				Framework:            s.Framework,
				SourceUrl:            s.SourceURL,
				ContextPath:          s.ContextPath,
				GithubInstallationId: s.GitHubInstallationID,
				CustomStartCommand:   s.CustomStartCommand,
				StartCommand:         s.StartCommand,
			})
		}
		infos[i] = &packager.EnvironmentInfo{
			Name:     m.Name,
			Services: svcs,
		}
	}
	return infos
}

func (s *Server) AddDatabase(ctx context.Context, req *packager.AddDatabaseRequest) (*packager.AddDatabaseResponse, error) {
	slog.Info("AddDatabase called", "project", req.Project, "database", req.Name)

	version := req.Version
	if version == "" {
		version = "16"
	}
	instances := int(req.Instances)
	if instances == 0 {
		instances = 1
	}
	size := req.Size
	if size == "" {
		size = "10Gi"
	}

	repo, err := s.cloneRepo(ctx, req.Project)

	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	if err := repo.AddDatabase(ctx, gitops.DatabaseDef{
		Name:      req.Name,
		Version:   version,
		Instances: instances,
		Size:      size,
	}); err != nil {
		return nil, fmt.Errorf("failed to add database: %w", err)
	}

	meta, err := repo.Metadata(ctx)

	if err != nil {
		slog.Warn("failed to read repo metadata for sync", "project", req.Project, "error", err)
	}

	s.syncAllEnvironments(ctx, req.Project, meta.Environments)
	return &packager.AddDatabaseResponse{}, nil
}

func (s *Server) RemoveDatabase(ctx context.Context, req *packager.RemoveDatabaseRequest) (*packager.RemoveDatabaseResponse, error) {
	slog.Info("RemoveDatabase called", "project", req.Project, "database", req.Name)

	repo, err := s.cloneRepo(ctx, req.Project)

	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	if err := repo.RemoveDatabase(ctx, req.Name); err != nil {
		return nil, fmt.Errorf("failed to remove database: %w", err)
	}

	meta, err := repo.Metadata(ctx)

	if err != nil {
		slog.Warn("failed to read repo metadata for sync", "project", req.Project, "error", err)
	}

	s.syncAllEnvironments(ctx, req.Project, meta.Environments)
	return &packager.RemoveDatabaseResponse{}, nil
}

func (s *Server) cloneRepo(ctx context.Context, project string) (gitops.Repository, error) {
	entry, err := s.gitops.Repo(ctx, project, tenant.FromContext(ctx))

	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return s.gitops.CloneRepo(ctx, entry.HTTPURL)
}

func databaseInfosFromDefs(defs []gitops.DatabaseDef) []*packager.DatabaseInfo {
	if len(defs) == 0 {
		return nil
	}
	infos := make([]*packager.DatabaseInfo, len(defs))
	for i, d := range defs {
		infos[i] = &packager.DatabaseInfo{
			Name:      d.Name,
			Version:   d.Version,
			Instances: int32(d.Instances),
			Size:      d.Size,
		}
	}
	return infos
}

func databaseRefsToProto(refs map[string]gitops.DatabaseRef) map[string]*packager.DatabaseRef {
	if len(refs) == 0 {
		return nil
	}
	result := make(map[string]*packager.DatabaseRef, len(refs))
	for k, v := range refs {
		result[k] = &packager.DatabaseRef{
			Database: v.Database,
			Key:      v.Key,
		}
	}
	return result
}

func serviceRefsToProto(refs map[string]gitops.ServiceRef) map[string]*packager.ServiceRef {
	if len(refs) == 0 {
		return nil
	}
	result := make(map[string]*packager.ServiceRef, len(refs))
	for k, v := range refs {
		result[k] = &packager.ServiceRef{
			Service: v.Service,
		}
	}
	return result
}

func (s *Server) SetResources(ctx context.Context, req *packager.SetResourcesRequest) (*packager.SetResourcesResponse, error) {
	repo, err := s.cloneRepo(ctx, req.Project)

	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	if err := repo.SetResources(ctx, req.Environment, req.Tier, int(req.CpuMillicores), int(req.MemoryMb), int(req.DiskMb)); err != nil {
		return nil, fmt.Errorf("failed to set resources: %w", err)
	}

	s.syncEnvironment(ctx, req.Project, req.Environment)
	return &packager.SetResourcesResponse{}, nil
}

func (s *Server) SetServiceScaling(ctx context.Context, req *packager.SetServiceScalingRequest) (*packager.SetServiceScalingResponse, error) {
	var autoscaling *gitops.AutoscalingConfig
	if req.Autoscaling != nil && req.Autoscaling.Enabled {
		autoscaling = &gitops.AutoscalingConfig{
			Enabled:     true,
			MinReplicas: int(req.Autoscaling.MinReplicas),
			MaxReplicas: int(req.Autoscaling.MaxReplicas),
			TargetCPU:   int(req.Autoscaling.TargetCpu),
		}
	}

	repo, err := s.cloneRepo(ctx, req.Project)

	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	if err := repo.SetServiceScaling(ctx, req.Environment, req.Service, int(req.Replicas), autoscaling); err != nil {
		return nil, fmt.Errorf("failed to set service scaling: %w", err)
	}
	s.syncEnvironment(ctx, req.Project, req.Environment)
	return &packager.SetServiceScalingResponse{}, nil
}

func (s *Server) SetCustomStartCommand(ctx context.Context, req *packager.SetCustomStartCommandRequest) (*packager.SetCustomStartCommandResponse, error) {
	slog.Info("SetCustomStartCommand called", "project", req.Project, "service", req.Service, "environment", req.Environment, "command", req.Command)

	repo, err := s.cloneRepo(ctx, req.Project)

	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	if err := repo.SetCustomStartCommand(ctx, req.Environment, req.Service, req.Command); err != nil {
		return nil, fmt.Errorf("failed to set custom start command: %w", err)
	}

	s.syncEnvironment(ctx, req.Project, req.Environment)
	return &packager.SetCustomStartCommandResponse{}, nil
}

func (s *Server) SetSuspended(ctx context.Context, req *packager.SetSuspendedRequest) (*packager.SetSuspendedResponse, error) {
	repo, err := s.cloneRepo(ctx, req.Project)

	if err != nil {
		return nil, err
	}
	defer repo.Cleanup()

	if err := repo.SetSuspended(ctx, req.Environment, req.Suspended); err != nil {
		return nil, fmt.Errorf("failed to set suspended: %w", err)
	}

	s.syncEnvironment(ctx, req.Project, req.Environment)
	return &packager.SetSuspendedResponse{}, nil
}

func (s *Server) generatePlatformDomain(ctx context.Context, service, environment string) (string, error) {
	// Get repos across all workspaces to ensure hostname is globally unique.
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
			// Hostname does not exist yet.
			return hostname, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique platform domain for service %s in environment %s", service, environment)
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

func databaseRefsFromProto(refs map[string]*packager.DatabaseRef) map[string]gitops.DatabaseRef {
	if len(refs) == 0 {
		return nil
	}
	result := make(map[string]gitops.DatabaseRef, len(refs))
	for k, v := range refs {
		result[k] = gitops.DatabaseRef{
			Database: v.Database,
			Key:      v.Key,
		}
	}
	return result
}

func serviceRefsFromProto(refs map[string]*packager.ServiceRef) map[string]gitops.ServiceRef {
	if len(refs) == 0 {
		return nil
	}
	result := make(map[string]gitops.ServiceRef, len(refs))
	for k, v := range refs {
		result[k] = gitops.ServiceRef{
			Service: v.Service,
		}
	}
	return result
}
