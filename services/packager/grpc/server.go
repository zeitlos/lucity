package grpc

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/packager"
	"github.com/zeitlos/lucity/services/packager/gitops"
)

type Server struct {
	packager.UnimplementedPackagerServiceServer
}

func NewServer() *Server {
	return &Server{}
}

// provider creates a GitOps provider for the current request using
// the OAuth token from the JWT claims in context.
func (s *Server) provider(ctx context.Context) (gitops.Provider, error) {
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
			Name:         proj.Name,
			SourceUrl:    proj.SourceURL,
			GitopsRepoUrl: proj.RepoURL,
			Environments: proj.Environments,
			CreatedAt:    proj.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Services:     serviceInfosFromDefs(proj.Services),
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

	return &packager.GetProjectResponse{
		Project: &packager.ProjectInfo{
			Name:         proj.Name,
			SourceUrl:    proj.SourceURL,
			GitopsRepoUrl: proj.RepoURL,
			Environments: proj.Environments,
			CreatedAt:    proj.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Services:     serviceInfosFromDefs(svcs),
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
	return &packager.CreateEnvironmentResponse{
		Namespace: req.Project + "-" + req.Environment,
	}, nil
}

func (s *Server) DeleteEnvironment(ctx context.Context, req *packager.DeleteEnvironmentRequest) (*packager.DeleteEnvironmentResponse, error) {
	slog.Info("DeleteEnvironment called", "project", req.Project, "environment", req.Environment)
	return &packager.DeleteEnvironmentResponse{}, nil
}

func (s *Server) Promote(ctx context.Context, req *packager.PromoteRequest) (*packager.PromoteResponse, error) {
	slog.Info("Promote called", "project", req.Project, "service", req.Service, "from", req.FromEnvironment, "to", req.ToEnvironment)
	return &packager.PromoteResponse{
		ImageTag: "promoted-tag",
	}, nil
}

func (s *Server) Eject(ctx context.Context, req *packager.EjectRequest) (*packager.EjectResponse, error) {
	slog.Info("Eject called", "project", req.Project)
	return &packager.EjectResponse{
		Archive: []byte("mock-ejected-archive"),
	}, nil
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
