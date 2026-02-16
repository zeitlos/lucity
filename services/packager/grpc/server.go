package grpc

import (
	"context"
	"log/slog"

	"github.com/zeitlos/lucity/pkg/packager"
)

type Server struct {
	packager.UnimplementedPackagerServiceServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) InitProject(ctx context.Context, req *packager.InitProjectRequest) (*packager.InitProjectResponse, error) {
	slog.Info("InitProject called", "project", req.Project, "source_url", req.SourceUrl)
	return &packager.InitProjectResponse{
		GitopsRepoUrl: "ssh://git@localhost:23231/" + req.Project + ".git",
	}, nil
}

func (s *Server) AddService(ctx context.Context, req *packager.AddServiceRequest) (*packager.AddServiceResponse, error) {
	slog.Info("AddService called", "project", req.Project, "service", req.Service, "image", req.Image)
	return &packager.AddServiceResponse{}, nil
}

func (s *Server) RemoveService(ctx context.Context, req *packager.RemoveServiceRequest) (*packager.RemoveServiceResponse, error) {
	slog.Info("RemoveService called", "project", req.Project, "service", req.Service)
	return &packager.RemoveServiceResponse{}, nil
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
