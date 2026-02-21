package grpc

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/packager"
	"github.com/zeitlos/lucity/services/packager/gitops"
)

type Server struct {
	packager.UnimplementedPackagerServiceServer
	// sharedProvider is used for non-GitHub backends (e.g., Soft-serve)
	// where the provider does not depend on per-request auth context.
	sharedProvider gitops.Provider
}

// NewServer creates a server that uses the GitHub provider (per-request OAuth token).
func NewServer() *Server {
	return &Server{}
}

// NewServerWithProvider creates a server that uses a shared provider instance.
func NewServerWithProvider(provider gitops.Provider) *Server {
	return &Server{sharedProvider: provider}
}

// provider returns the GitOps provider for the current request.
// If a shared provider is configured, it's returned directly.
// Otherwise, creates a GitHub provider from the JWT claims.
func (s *Server) provider(ctx context.Context) (gitops.Provider, error) {
	if s.sharedProvider != nil {
		return s.sharedProvider, nil
	}

	claims := auth.FromContext(ctx)
	if claims == nil {
		return nil, fmt.Errorf("unauthenticated")
	}
	if claims.GitHubToken == "" {
		return nil, fmt.Errorf("no github token in session")
	}
	return gitops.NewGitHubProvider(claims.GitHubToken), nil
}

func (s *Server) InitProject(ctx context.Context, req *packager.InitProjectRequest) (*packager.InitProjectResponse, error) {
	slog.Info("InitProject called", "project", req.Project, "source_url", req.SourceUrl)

	p, err := s.provider(ctx)
	if err != nil {
		return nil, err
	}

	repoURL, err := p.CreateRepo(ctx, req.Project, req.SourceUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to init project: %w", err)
	}

	return &packager.InitProjectResponse{
		GitopsRepoUrl: repoURL,
	}, nil
}

func (s *Server) ListProjects(ctx context.Context, req *packager.ListProjectsRequest) (*packager.ListProjectsResponse, error) {
	slog.Info("ListProjects called")

	p, err := s.provider(ctx)
	if err != nil {
		return nil, err
	}

	projects, err := p.Repos(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	var infos []*packager.ProjectInfo
	for _, proj := range projects {
		infos = append(infos, &packager.ProjectInfo{
			Name:             proj.Name,
			SourceUrl:        proj.SourceURL,
			GitopsRepoUrl:    proj.RepoURL,
			Environments:     proj.Environments,
			EnvironmentInfos: envInfosFromMeta(proj.EnvironmentInfos),
			CreatedAt:        timestamppb.New(proj.CreatedAt),
			Services:         serviceInfosFromDefs(proj.Services),
		})
	}

	return &packager.ListProjectsResponse{Projects: infos}, nil
}

func (s *Server) GetProject(ctx context.Context, req *packager.GetProjectRequest) (*packager.GetProjectResponse, error) {
	slog.Info("GetProject called", "project", req.Project)

	p, err := s.provider(ctx)
	if err != nil {
		return nil, err
	}

	proj, err := p.Repo(ctx, req.Project)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	svcs, err := p.Services(ctx, req.Project)
	if err != nil {
		slog.Warn("failed to read services", "project", req.Project, "error", err)
	}

	// If EnvironmentInfos is not populated (e.g., GitHub provider), read per-env data.
	envInfos := proj.EnvironmentInfos
	if len(envInfos) == 0 && len(proj.Environments) > 0 {
		for _, envName := range proj.Environments {
			envMeta := gitops.EnvironmentMeta{Name: envName}
			envSvcs, envErr := p.EnvironmentServices(ctx, req.Project, envName)
			if envErr != nil {
				slog.Debug("failed to read environment services", "project", req.Project, "environment", envName, "error", envErr)
			} else {
				envMeta.Services = envSvcs
			}
			envInfos = append(envInfos, envMeta)
		}
	}

	return &packager.GetProjectResponse{
		Project: &packager.ProjectInfo{
			Name:             proj.Name,
			SourceUrl:        proj.SourceURL,
			GitopsRepoUrl:    proj.RepoURL,
			Environments:     proj.Environments,
			EnvironmentInfos: envInfosFromMeta(envInfos),
			CreatedAt:        timestamppb.New(proj.CreatedAt),
			Services:         serviceInfosFromDefs(svcs),
		},
	}, nil
}

func (s *Server) DeleteProject(ctx context.Context, req *packager.DeleteProjectRequest) (*packager.DeleteProjectResponse, error) {
	slog.Info("DeleteProject called", "project", req.Project)

	p, err := s.provider(ctx)
	if err != nil {
		return nil, err
	}

	if err := p.DeleteRepo(ctx, req.Project); err != nil {
		return nil, fmt.Errorf("failed to delete project: %w", err)
	}

	return &packager.DeleteProjectResponse{}, nil
}

func (s *Server) AddService(ctx context.Context, req *packager.AddServiceRequest) (*packager.AddServiceResponse, error) {
	slog.Info("AddService called", "project", req.Project, "service", req.Service, "image", req.Image)

	p, err := s.provider(ctx)
	if err != nil {
		return nil, err
	}

	if err := p.AddService(ctx, req.Project, gitops.ServiceDef{
		Name:      req.Service,
		Image:     req.Image,
		Port:      int(req.Port),
		Public:    req.Public,
		Framework: req.Framework,
	}); err != nil {
		return nil, fmt.Errorf("failed to add service: %w", err)
	}

	return &packager.AddServiceResponse{}, nil
}

func (s *Server) RemoveService(ctx context.Context, req *packager.RemoveServiceRequest) (*packager.RemoveServiceResponse, error) {
	slog.Info("RemoveService called", "project", req.Project, "service", req.Service)

	p, err := s.provider(ctx)
	if err != nil {
		return nil, err
	}

	if err := p.RemoveService(ctx, req.Project, req.Service); err != nil {
		return nil, fmt.Errorf("failed to remove service: %w", err)
	}

	return &packager.RemoveServiceResponse{}, nil
}

func (s *Server) UpdateImageTag(ctx context.Context, req *packager.UpdateImageTagRequest) (*packager.UpdateImageTagResponse, error) {
	slog.Info("UpdateImageTag called", "project", req.Project, "environment", req.Environment, "service", req.Service, "tag", req.Tag)

	p, err := s.provider(ctx)
	if err != nil {
		return nil, err
	}

	if err := p.UpdateImageTag(ctx, req.Project, req.Environment, req.Service, req.Tag, req.Digest); err != nil {
		return nil, fmt.Errorf("failed to update image tag: %w", err)
	}

	return &packager.UpdateImageTagResponse{}, nil
}

func (s *Server) CreateEnvironment(ctx context.Context, req *packager.CreateEnvironmentRequest) (*packager.CreateEnvironmentResponse, error) {
	slog.Info("CreateEnvironment called", "project", req.Project, "environment", req.Environment)

	p, err := s.provider(ctx)
	if err != nil {
		return nil, err
	}

	if err := p.CreateEnvironment(ctx, req.Project, req.Environment, req.FromEnvironment); err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	return &packager.CreateEnvironmentResponse{
		Namespace: gitops.NamespaceFor(req.Project, req.Environment),
	}, nil
}

func (s *Server) DeleteEnvironment(ctx context.Context, req *packager.DeleteEnvironmentRequest) (*packager.DeleteEnvironmentResponse, error) {
	slog.Info("DeleteEnvironment called", "project", req.Project, "environment", req.Environment)

	p, err := s.provider(ctx)
	if err != nil {
		return nil, err
	}

	if err := p.DeleteEnvironment(ctx, req.Project, req.Environment); err != nil {
		return nil, fmt.Errorf("failed to delete environment: %w", err)
	}

	return &packager.DeleteEnvironmentResponse{}, nil
}

func (s *Server) Promote(ctx context.Context, req *packager.PromoteRequest) (*packager.PromoteResponse, error) {
	slog.Info("Promote called", "project", req.Project, "service", req.Service, "from", req.FromEnvironment, "to", req.ToEnvironment)

	p, err := s.provider(ctx)
	if err != nil {
		return nil, err
	}

	imageTag, err := p.Promote(ctx, req.Project, req.Service, req.FromEnvironment, req.ToEnvironment)
	if err != nil {
		return nil, fmt.Errorf("failed to promote: %w", err)
	}

	return &packager.PromoteResponse{
		ImageTag: imageTag,
	}, nil
}

func (s *Server) DeploymentHistory(ctx context.Context, req *packager.DeploymentHistoryRequest) (*packager.DeploymentHistoryResponse, error) {
	slog.Info("DeploymentHistory called", "project", req.Project, "environment", req.Environment, "service", req.Service)

	p, err := s.provider(ctx)
	if err != nil {
		return nil, err
	}

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

func (s *Server) UpdateServiceConfig(ctx context.Context, req *packager.UpdateServiceConfigRequest) (*packager.UpdateServiceConfigResponse, error) {
	slog.Info("UpdateServiceConfig called", "project", req.Project, "service", req.Service)

	p, err := s.provider(ctx)
	if err != nil {
		return nil, err
	}

	if err := p.UpdateServiceConfig(ctx, req.Project, req.Service, req.Public); err != nil {
		return nil, fmt.Errorf("failed to update service config: %w", err)
	}

	return &packager.UpdateServiceConfigResponse{}, nil
}

func (s *Server) SetServiceDomain(ctx context.Context, req *packager.SetServiceDomainRequest) (*packager.SetServiceDomainResponse, error) {
	slog.Info("SetServiceDomain called", "project", req.Project, "environment", req.Environment, "service", req.Service, "host", req.Host)

	p, err := s.provider(ctx)
	if err != nil {
		return nil, err
	}

	if err := p.SetServiceDomain(ctx, req.Project, req.Environment, req.Service, req.Host); err != nil {
		return nil, fmt.Errorf("failed to set service domain: %w", err)
	}

	return &packager.SetServiceDomainResponse{}, nil
}

func (s *Server) Eject(ctx context.Context, req *packager.EjectRequest) (*packager.EjectResponse, error) {
	slog.Info("Eject called", "project", req.Project)
	return &packager.EjectResponse{
		Archive: []byte("mock-ejected-archive"),
	}, nil
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
				Name:     s.Name,
				ImageTag: s.ImageTag,
				Host:     s.Host,
			})
		}
		infos[i] = &packager.EnvironmentInfo{
			Name:     m.Name,
			Services: svcs,
		}
	}
	return infos
}

func serviceInfosFromDefs(defs []gitops.ServiceDef) []*packager.ServiceInfo {
	if len(defs) == 0 {
		return nil
	}
	infos := make([]*packager.ServiceInfo, len(defs))
	for i, d := range defs {
		infos[i] = &packager.ServiceInfo{
			Name:      d.Name,
			Image:     d.Image,
			Port:      int32(d.Port),
			Public:    d.Public,
			Framework: d.Framework,
		}
	}
	return infos
}
