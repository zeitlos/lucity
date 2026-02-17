package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/services/deployer/argocd"
)

type Server struct {
	deployer.UnimplementedDeployerServiceServer
	argo *argocd.Client

	// softServeHTTP is the cluster-internal Soft-serve HTTP URL for ArgoCD to clone from.
	softServeHTTP string
}

func NewServer(argo *argocd.Client, softServeHTTP string) *Server {
	return &Server{
		argo:          argo,
		softServeHTTP: softServeHTTP,
	}
}

func (s *Server) DeployEnvironment(ctx context.Context, req *deployer.DeployEnvironmentRequest) (*deployer.DeployEnvironmentResponse, error) {
	appName := applicationName(req.Project, req.Environment)

	app := argocd.Application{
		Metadata: argocd.Metadata{
			Name: appName,
		},
		Spec: argocd.ApplicationSpec{
			Source: argocd.Source{
				RepoURL:        s.repoURL(req.Project),
				Path:           "base",
				TargetRevision: "HEAD",
				Helm: &argocd.Helm{
					ValueFiles: []string{
						fmt.Sprintf("../environments/%s/values.yaml", req.Environment),
					},
				},
			},
			Destination: argocd.Destination{
				Server:    "https://kubernetes.default.svc",
				Namespace: req.TargetNamespace,
			},
			Project: "default",
			SyncPolicy: &argocd.SyncPolicy{
				Automated: &argocd.Automated{
					Prune:    true,
					SelfHeal: true,
				},
			},
		},
	}

	result, err := s.argo.CreateApplication(ctx, app)
	if err != nil {
		return nil, fmt.Errorf("failed to create ArgoCD application: %w", err)
	}

	slog.Info("created ArgoCD application",
		"app", result.Metadata.Name,
		"project", req.Project,
		"environment", req.Environment,
		"namespace", req.TargetNamespace,
	)

	return &deployer.DeployEnvironmentResponse{
		DeploymentName: result.Metadata.Name,
	}, nil
}

func (s *Server) RemoveDeployment(ctx context.Context, req *deployer.RemoveDeploymentRequest) (*deployer.RemoveDeploymentResponse, error) {
	appName := applicationName(req.Project, req.Environment)

	if err := s.argo.DeleteApplication(ctx, appName, true); err != nil {
		return nil, fmt.Errorf("failed to delete ArgoCD application: %w", err)
	}

	slog.Info("deleted ArgoCD application", "app", appName)
	return &deployer.RemoveDeploymentResponse{}, nil
}

func (s *Server) GetDeploymentStatus(ctx context.Context, req *deployer.GetDeploymentStatusRequest) (*deployer.GetDeploymentStatusResponse, error) {
	appName := applicationName(req.Project, req.Environment)

	app, err := s.argo.Application(ctx, appName)
	if err != nil {
		return nil, fmt.Errorf("failed to get ArgoCD application: %w", err)
	}

	status, message := mapStatus(app.Status)

	return &deployer.GetDeploymentStatusResponse{
		Status:  status,
		Message: message,
	}, nil
}

func (s *Server) SyncDeployment(ctx context.Context, req *deployer.SyncDeploymentRequest) (*deployer.SyncDeploymentResponse, error) {
	appName := applicationName(req.Project, req.Environment)

	app, err := s.argo.SyncApplication(ctx, appName)
	if err != nil {
		return nil, fmt.Errorf("failed to sync ArgoCD application: %w", err)
	}

	slog.Info("triggered sync", "app", appName)

	status, _ := mapStatus(app.Status)
	return &deployer.SyncDeploymentResponse{
		Status: status,
	}, nil
}

// applicationName derives the ArgoCD Application name from project and environment.
// "zeitlos/myapp" + "production" → "myapp-production"
func applicationName(project, environment string) string {
	parts := strings.SplitN(project, "/", 2)
	name := project
	if len(parts) == 2 {
		name = parts[1]
	}
	return name + "-" + environment
}

// repoURL returns the Soft-serve HTTP clone URL for a project's GitOps repo.
func (s *Server) repoURL(project string) string {
	parts := strings.SplitN(project, "/", 2)
	name := project
	if len(parts) == 2 {
		name = parts[1]
	}
	return strings.TrimSuffix(s.softServeHTTP, "/") + "/" + name + "-gitops.git"
}

// mapStatus converts ArgoCD health/sync status to proto DeploymentStatus.
func mapStatus(status *argocd.AppStatus) (deployer.DeploymentStatus, string) {
	if status == nil {
		return deployer.DeploymentStatus_DEPLOYMENT_STATUS_UNKNOWN, "no status available"
	}

	switch status.Health.Status {
	case "Healthy":
		if status.Sync.Status == "Synced" {
			return deployer.DeploymentStatus_DEPLOYMENT_STATUS_SYNCED, "all resources synced and healthy"
		}
		return deployer.DeploymentStatus_DEPLOYMENT_STATUS_OUT_OF_SYNC, "healthy but out of sync"
	case "Progressing":
		return deployer.DeploymentStatus_DEPLOYMENT_STATUS_PROGRESSING, status.Health.Message
	case "Degraded":
		return deployer.DeploymentStatus_DEPLOYMENT_STATUS_DEGRADED, status.Health.Message
	default:
		if status.Sync.Status == "OutOfSync" {
			return deployer.DeploymentStatus_DEPLOYMENT_STATUS_OUT_OF_SYNC, "out of sync"
		}
		return deployer.DeploymentStatus_DEPLOYMENT_STATUS_UNKNOWN, status.Health.Status
	}
}
