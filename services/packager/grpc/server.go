package grpc

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/packager"
	"github.com/zeitlos/lucity/pkg/tenant"
	"github.com/zeitlos/lucity/services/packager/eject"
	"github.com/zeitlos/lucity/services/packager/gitops"
)

type Server struct {
	packager.UnimplementedPackagerServiceServer
	provider gitops.Provider
	deployer deployer.DeployerServiceClient
}

// NewServer creates a packager server with the given GitOps provider.
func NewServer(provider gitops.Provider, deployerClient deployer.DeployerServiceClient) *Server {
	return &Server{provider: provider, deployer: deployerClient}
}

// syncEnvironment triggers an ArgoCD sync for a single environment.
// Best-effort: logs on failure but never returns an error.
func (s *Server) syncEnvironment(ctx context.Context, project, environment string) {
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
func (s *Server) syncAllEnvironments(ctx context.Context, project string) {
	meta, err := s.provider.Repo(ctx, project)
	if err != nil {
		slog.Warn("failed to read project for sync", "project", project, "error", err)
		return
	}
	for _, env := range meta.Environments {
		s.syncEnvironment(ctx, project, env)
	}
}

func (s *Server) InitProject(ctx context.Context, req *packager.InitProjectRequest) (*packager.InitProjectResponse, error) {
	slog.Info("InitProject called", "project", req.Project)

	p := s.provider

	repoURL, err := p.CreateRepo(ctx, req.Project, req.DisplayName)
	if err != nil {
		return nil, fmt.Errorf("failed to init project: %w", err)
	}

	return &packager.InitProjectResponse{
		GitopsRepoUrl: repoURL,
	}, nil
}

func (s *Server) ListProjects(ctx context.Context, req *packager.ListProjectsRequest) (*packager.ListProjectsResponse, error) {
	slog.Info("ListProjects called")

	p := s.provider

	projects, err := p.Repos(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	var infos []*packager.ProjectInfo
	for _, proj := range projects {
		infos = append(infos, &packager.ProjectInfo{
			Name:             proj.Name,
			DisplayName:      proj.DisplayName,
			GitopsRepoUrl:    proj.RepoURL,
			Environments:     proj.Environments,
			EnvironmentInfos: envInfosFromMeta(proj.EnvironmentInfos),
			CreatedAt:        timestamppb.New(proj.CreatedAt),
			Databases:        databaseInfosFromDefs(proj.Databases),
		})
	}

	return &packager.ListProjectsResponse{Projects: infos}, nil
}

func (s *Server) GetProject(ctx context.Context, req *packager.GetProjectRequest) (*packager.GetProjectResponse, error) {
	slog.Info("GetProject called", "project", req.Project)

	proj, err := s.provider.Repo(ctx, req.Project)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &packager.GetProjectResponse{
		Project: &packager.ProjectInfo{
			Name:             proj.Name,
			DisplayName:      proj.DisplayName,
			GitopsRepoUrl:    proj.RepoURL,
			Environments:     proj.Environments,
			EnvironmentInfos: envInfosFromMeta(proj.EnvironmentInfos),
			CreatedAt:        timestamppb.New(proj.CreatedAt),
			Databases:        databaseInfosFromDefs(proj.Databases),
		},
	}, nil
}

func (s *Server) DeleteProject(ctx context.Context, req *packager.DeleteProjectRequest) (*packager.DeleteProjectResponse, error) {
	slog.Info("DeleteProject called", "project", req.Project)

	p := s.provider

	if err := p.DeleteRepo(ctx, req.Project); err != nil {
		return nil, fmt.Errorf("failed to delete project: %w", err)
	}

	return &packager.DeleteProjectResponse{}, nil
}

func (s *Server) AddService(ctx context.Context, req *packager.AddServiceRequest) (*packager.AddServiceResponse, error) {
	slog.Info("AddService called", "project", req.Project, "service", req.Service, "environment", req.Environment, "image", req.Image)

	p := s.provider

	// Sync chart before adding service so new templates (e.g., HTTPRoute parentRefs)
	// are available when ArgoCD renders the Helm chart after deployment.
	if err := p.SyncChart(ctx, req.Project); err != nil {
		slog.Warn("failed to sync chart before adding service", "project", req.Project, "error", err)
	}

	if err := p.AddService(ctx, req.Project, req.Environment, gitops.ServiceDef{
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

	if err := s.provider.RemoveService(ctx, req.Project, req.Environment, req.Service); err != nil {
		return nil, fmt.Errorf("failed to remove service: %w", err)
	}

	s.syncEnvironment(ctx, req.Project, req.Environment)
	return &packager.RemoveServiceResponse{}, nil
}

func (s *Server) UpdateImageTag(ctx context.Context, req *packager.UpdateImageTagRequest) (*packager.UpdateImageTagResponse, error) {
	slog.Info("UpdateImageTag called", "project", req.Project, "environment", req.Environment, "service", req.Service, "tag", req.Tag)

	p := s.provider

	// Sync chart so updated templates (e.g., image.tag guards) reach the GitOps repo.
	if err := p.SyncChart(ctx, req.Project); err != nil {
		slog.Warn("failed to sync chart before updating image tag", "project", req.Project, "error", err)
	}

	if err := p.UpdateImageTag(ctx, req.Project, req.Environment, req.Service, req.Tag, req.Digest, req.CommitPrefix); err != nil {
		return nil, fmt.Errorf("failed to update image tag: %w", err)
	}

	s.syncEnvironment(ctx, req.Project, req.Environment)
	return &packager.UpdateImageTagResponse{}, nil
}

func (s *Server) CreateEnvironment(ctx context.Context, req *packager.CreateEnvironmentRequest) (*packager.CreateEnvironmentResponse, error) {
	slog.Info("CreateEnvironment called", "project", req.Project, "environment", req.Environment, "from", req.FromEnvironment)

	serviceNames, err := s.provider.CreateEnvironment(ctx, req.Project, req.Environment, req.FromEnvironment, req.WorkloadDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	ws := tenant.FromContext(ctx)
	return &packager.CreateEnvironmentResponse{
		Namespace:           gitops.NamespaceFor(ws, req.Project, req.Environment),
		ServicesWithDomains: serviceNames,
	}, nil
}

func (s *Server) DeleteEnvironment(ctx context.Context, req *packager.DeleteEnvironmentRequest) (*packager.DeleteEnvironmentResponse, error) {
	slog.Info("DeleteEnvironment called", "project", req.Project, "environment", req.Environment)

	p := s.provider

	if err := p.DeleteEnvironment(ctx, req.Project, req.Environment); err != nil {
		return nil, fmt.Errorf("failed to delete environment: %w", err)
	}

	return &packager.DeleteEnvironmentResponse{}, nil
}

func (s *Server) Promote(ctx context.Context, req *packager.PromoteRequest) (*packager.PromoteResponse, error) {
	slog.Info("Promote called", "project", req.Project, "service", req.Service, "from", req.FromEnvironment, "to", req.ToEnvironment)

	p := s.provider

	if err := p.SyncChart(ctx, req.Project); err != nil {
		slog.Warn("failed to sync chart before promote", "project", req.Project, "error", err)
	}

	imageTag, err := p.Promote(ctx, req.Project, req.Service, req.FromEnvironment, req.ToEnvironment)
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

	p := s.provider

	entries, err := p.DeploymentHistory(ctx, req.Project, req.Environment, req.Service)
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

func (s *Server) AddDomain(ctx context.Context, req *packager.AddDomainRequest) (*packager.AddDomainResponse, error) {
	slog.Info("AddDomain called", "project", req.Project, "environment", req.Environment, "service", req.Service, "hostname", req.Hostname)

	p := s.provider

	if err := p.AddDomain(ctx, req.Project, req.Environment, req.Service, req.Hostname); err != nil {
		return nil, fmt.Errorf("failed to add domain: %w", err)
	}

	s.syncEnvironment(ctx, req.Project, req.Environment)
	return &packager.AddDomainResponse{}, nil
}

func (s *Server) RemoveDomain(ctx context.Context, req *packager.RemoveDomainRequest) (*packager.RemoveDomainResponse, error) {
	slog.Info("RemoveDomain called", "project", req.Project, "environment", req.Environment, "service", req.Service, "hostname", req.Hostname)

	p := s.provider

	if err := p.RemoveDomain(ctx, req.Project, req.Environment, req.Service, req.Hostname); err != nil {
		return nil, fmt.Errorf("failed to remove domain: %w", err)
	}

	s.syncEnvironment(ctx, req.Project, req.Environment)
	return &packager.RemoveDomainResponse{}, nil
}

func (s *Server) AllDomains(ctx context.Context, req *packager.AllDomainsRequest) (*packager.AllDomainsResponse, error) {
	slog.Info("AllDomains called")

	hostnames, err := s.provider.AllDomains(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list all domains: %w", err)
	}

	return &packager.AllDomainsResponse{Hostnames: hostnames}, nil
}

func (s *Server) Eject(ctx context.Context, req *packager.EjectRequest) (*packager.EjectResponse, error) {
	slog.Info("eject started", "project", req.Project)

	p := s.provider

	archive, err := eject.Build(ctx, p, req.Project)
	if err != nil {
		return nil, fmt.Errorf("failed to build ejection archive: %w", err)
	}

	slog.Info("eject completed", "project", req.Project, "size", len(archive))
	return &packager.EjectResponse{Archive: archive}, nil
}

func (s *Server) SharedVariables(ctx context.Context, req *packager.SharedVariablesRequest) (*packager.SharedVariablesResponse, error) {
	slog.Info("SharedVariables called", "project", req.Project, "environment", req.Environment)

	p := s.provider

	vars, err := p.SharedVariables(ctx, req.Project, req.Environment)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared variables: %w", err)
	}

	return &packager.SharedVariablesResponse{Variables: vars}, nil
}

func (s *Server) SetSharedVariables(ctx context.Context, req *packager.SetSharedVariablesRequest) (*packager.SetSharedVariablesResponse, error) {
	slog.Info("SetSharedVariables called", "project", req.Project, "environment", req.Environment)

	p := s.provider

	if err := p.SetSharedVariables(ctx, req.Project, req.Environment, req.Variables); err != nil {
		return nil, fmt.Errorf("failed to set shared variables: %w", err)
	}

	s.syncEnvironment(ctx, req.Project, req.Environment)
	return &packager.SetSharedVariablesResponse{}, nil
}

func (s *Server) ServiceVariables(ctx context.Context, req *packager.ServiceVariablesRequest) (*packager.ServiceVariablesResponse, error) {
	slog.Info("ServiceVariables called", "project", req.Project, "environment", req.Environment, "service", req.Service)

	p := s.provider

	vars, refs, databaseRefs, serviceRefs, err := p.ServiceVariables(ctx, req.Project, req.Environment, req.Service)
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

	p := s.provider

	if err := p.SetServiceVariables(ctx, req.Project, req.Environment, req.Service, req.Variables, req.SharedRefs, databaseRefsFromProto(req.DatabaseRefs), serviceRefsFromProto(req.ServiceRefs)); err != nil {
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

func (s *Server) SyncChart(ctx context.Context, req *packager.SyncChartRequest) (*packager.SyncChartResponse, error) {
	slog.Info("SyncChart called", "project", req.Project)

	p := s.provider

	if err := p.SyncChart(ctx, req.Project); err != nil {
		return nil, fmt.Errorf("failed to sync chart: %w", err)
	}

	s.syncAllEnvironments(ctx, req.Project)
	return &packager.SyncChartResponse{}, nil
}

func (s *Server) AddDatabase(ctx context.Context, req *packager.AddDatabaseRequest) (*packager.AddDatabaseResponse, error) {
	slog.Info("AddDatabase called", "project", req.Project, "database", req.Name)

	p := s.provider

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

	if err := p.AddDatabase(ctx, req.Project, gitops.DatabaseDef{
		Name:      req.Name,
		Version:   version,
		Instances: instances,
		Size:      size,
	}); err != nil {
		return nil, fmt.Errorf("failed to add database: %w", err)
	}

	s.syncAllEnvironments(ctx, req.Project)
	return &packager.AddDatabaseResponse{}, nil
}

func (s *Server) RemoveDatabase(ctx context.Context, req *packager.RemoveDatabaseRequest) (*packager.RemoveDatabaseResponse, error) {
	slog.Info("RemoveDatabase called", "project", req.Project, "database", req.Name)

	p := s.provider

	if err := p.RemoveDatabase(ctx, req.Project, req.Name); err != nil {
		return nil, fmt.Errorf("failed to remove database: %w", err)
	}

	s.syncAllEnvironments(ctx, req.Project)
	return &packager.RemoveDatabaseResponse{}, nil
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
	if err := s.provider.SetResources(ctx, req.Project, req.Environment, req.Tier, int(req.CpuMillicores), int(req.MemoryMb), int(req.DiskMb)); err != nil {
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

	if err := s.provider.SetServiceScaling(ctx, req.Project, req.Environment, req.Service, int(req.Replicas), autoscaling); err != nil {
		return nil, fmt.Errorf("failed to set service scaling: %w", err)
	}
	s.syncEnvironment(ctx, req.Project, req.Environment)
	return &packager.SetServiceScalingResponse{}, nil
}

func (s *Server) SetCustomStartCommand(ctx context.Context, req *packager.SetCustomStartCommandRequest) (*packager.SetCustomStartCommandResponse, error) {
	slog.Info("SetCustomStartCommand called", "project", req.Project, "service", req.Service, "environment", req.Environment, "command", req.Command)

	if err := s.provider.SetCustomStartCommand(ctx, req.Project, req.Environment, req.Service, req.Command); err != nil {
		return nil, fmt.Errorf("failed to set custom start command: %w", err)
	}
	s.syncEnvironment(ctx, req.Project, req.Environment)
	return &packager.SetCustomStartCommandResponse{}, nil
}

func (s *Server) SetSuspended(ctx context.Context, req *packager.SetSuspendedRequest) (*packager.SetSuspendedResponse, error) {
	if err := s.provider.SetSuspended(ctx, req.Project, req.Environment, req.Suspended); err != nil {
		return nil, fmt.Errorf("failed to set suspended: %w", err)
	}
	s.syncEnvironment(ctx, req.Project, req.Environment)
	return &packager.SetSuspendedResponse{}, nil
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
