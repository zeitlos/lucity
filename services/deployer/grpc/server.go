package grpc

import (
	"context"
	"log/slog"

	"github.com/zeitlos/lucity/pkg/deployer"
)

type Server struct {
	deployer.UnimplementedDeployerServiceServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) DeployEnvironment(ctx context.Context, req *deployer.DeployEnvironmentRequest) (*deployer.DeployEnvironmentResponse, error) {
	slog.Info("DeployEnvironment called",
		"project", req.Project,
		"environment", req.Environment,
		"namespace", req.TargetNamespace,
	)
	return &deployer.DeployEnvironmentResponse{
		DeploymentName: req.Project + "-" + req.Environment,
	}, nil
}

func (s *Server) RemoveDeployment(ctx context.Context, req *deployer.RemoveDeploymentRequest) (*deployer.RemoveDeploymentResponse, error) {
	slog.Info("RemoveDeployment called", "project", req.Project, "environment", req.Environment)
	return &deployer.RemoveDeploymentResponse{}, nil
}

func (s *Server) GetDeploymentStatus(ctx context.Context, req *deployer.GetDeploymentStatusRequest) (*deployer.GetDeploymentStatusResponse, error) {
	slog.Info("GetDeploymentStatus called", "project", req.Project, "environment", req.Environment)
	return &deployer.GetDeploymentStatusResponse{
		Status:  deployer.DeploymentStatus_DEPLOYMENT_STATUS_SYNCED,
		Message: "all resources synced",
	}, nil
}

func (s *Server) SyncDeployment(ctx context.Context, req *deployer.SyncDeploymentRequest) (*deployer.SyncDeploymentResponse, error) {
	slog.Info("SyncDeployment called", "project", req.Project, "environment", req.Environment)
	return &deployer.SyncDeploymentResponse{
		Status: deployer.DeploymentStatus_DEPLOYMENT_STATUS_PROGRESSING,
	}, nil
}
