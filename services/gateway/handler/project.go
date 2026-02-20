package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/packager"
)

type Project struct {
	ID           string
	Name         string
	SourceURL    string
	Environments []Environment
	Services     []Service
	CreatedAt    time.Time
}

type Environment struct {
	ID         string
	Name       string
	Namespace  string
	Ephemeral  bool
	SyncStatus string
	Services   []ServiceInstance
}

type Service struct {
	Name      string
	Image     string
	Port      int
	Public    bool
	Framework string
	Instances []ServiceInstance
}

type ServiceInstance struct {
	Name        string
	Environment string
	ImageTag    string
	Ready       bool
	Replicas    int
}

type Deployment struct {
	ID        string
	ImageTag  string
	Active    bool
	Timestamp time.Time
}

func (c *Client) Projects(ctx context.Context) ([]Project, error) {
	ctx = auth.OutgoingContext(ctx)

	resp, err := c.Packager.ListProjects(ctx, &packager.ListProjectsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	result := make([]Project, 0, len(resp.Projects))
	for _, p := range resp.Projects {
		proj := projectFromProto(p)
		c.enrichSyncStatus(ctx, &proj)
		result = append(result, proj)
	}
	return result, nil
}

func (c *Client) Project(ctx context.Context, id string) (*Project, error) {
	ctx = auth.OutgoingContext(ctx)

	resp, err := c.Packager.GetProject(ctx, &packager.GetProjectRequest{Project: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	p := projectFromProto(resp.Project)
	c.enrichSyncStatus(ctx, &p)
	return &p, nil
}

func (c *Client) CreateProject(ctx context.Context, name, sourceURL string) (*Project, error) {
	ctx = auth.OutgoingContext(ctx)

	resp, err := c.Packager.InitProject(ctx, &packager.InitProjectRequest{
		Project:   name,
		SourceUrl: sourceURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Deploy the default development environment via ArgoCD
	ns := namespaceFor(name, "development")
	_, err = c.Deployer.DeployEnvironment(ctx, &deployer.DeployEnvironmentRequest{
		Project:         name,
		Environment:     "development",
		GitopsRepoUrl:   resp.GitopsRepoUrl,
		TargetNamespace: ns,
	})
	if err != nil {
		slog.Warn("failed to deploy development environment", "project", name, "error", err)
	}

	return &Project{
		ID:        name,
		Name:      name,
		SourceURL: sourceURL,
		CreatedAt: time.Now(),
		Environments: []Environment{
			{
				ID:         name + "/development",
				Name:       "development",
				Namespace:  ns,
				SyncStatus: "PROGRESSING",
			},
		},
	}, nil
}

func (c *Client) DeleteProject(ctx context.Context, id string) (bool, error) {
	ctx = auth.OutgoingContext(ctx)

	_, err := c.Packager.DeleteProject(ctx, &packager.DeleteProjectRequest{Project: id})
	if err != nil {
		return false, fmt.Errorf("failed to delete project: %w", err)
	}
	return true, nil
}

func (c *Client) CreateEnvironment(ctx context.Context, projectID, name, fromEnvironment string) (*Environment, error) {
	ctx = auth.OutgoingContext(ctx)

	resp, err := c.Packager.CreateEnvironment(ctx, &packager.CreateEnvironmentRequest{
		Project:         projectID,
		Environment:     name,
		FromEnvironment: fromEnvironment,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	// Deploy the new environment via ArgoCD
	_, err = c.Deployer.DeployEnvironment(ctx, &deployer.DeployEnvironmentRequest{
		Project:         projectID,
		Environment:     name,
		TargetNamespace: resp.Namespace,
	})
	if err != nil {
		slog.Warn("failed to deploy environment", "project", projectID, "environment", name, "error", err)
	}

	return &Environment{
		ID:         projectID + "/" + name,
		Name:       name,
		Namespace:  resp.Namespace,
		SyncStatus: "PROGRESSING",
	}, nil
}

func (c *Client) DeleteEnvironment(ctx context.Context, projectID, environment string) (bool, error) {
	ctx = auth.OutgoingContext(ctx)

	// Remove ArgoCD Application first (cascade deletes managed resources)
	_, err := c.Deployer.RemoveDeployment(ctx, &deployer.RemoveDeploymentRequest{
		Project:     projectID,
		Environment: environment,
	})
	if err != nil {
		slog.Warn("failed to remove deployment", "project", projectID, "environment", environment, "error", err)
	}

	// Then remove from GitOps repo
	_, err = c.Packager.DeleteEnvironment(ctx, &packager.DeleteEnvironmentRequest{
		Project:     projectID,
		Environment: environment,
	})
	if err != nil {
		return false, fmt.Errorf("failed to delete environment: %w", err)
	}
	return true, nil
}

func (c *Client) Promote(ctx context.Context, projectID, service, fromEnv, toEnv string) (*ServiceInstance, error) {
	ctx = auth.OutgoingContext(ctx)

	resp, err := c.Packager.Promote(ctx, &packager.PromoteRequest{
		Project:         projectID,
		Service:         service,
		FromEnvironment: fromEnv,
		ToEnvironment:   toEnv,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to promote: %w", err)
	}

	return &ServiceInstance{
		Name:        service,
		Environment: toEnv,
		ImageTag:    resp.ImageTag,
	}, nil
}

func (c *Client) Service(ctx context.Context, projectID, name string) (*Service, error) {
	proj, err := c.Project(ctx, projectID)
	if err != nil {
		return nil, err
	}

	for _, svc := range proj.Services {
		if svc.Name == name {
			return &svc, nil
		}
	}
	return nil, nil
}

// namespaceFor derives the K8s namespace from project and environment.
// "zeitlos/myapp" + "production" → "myapp-production"
func namespaceFor(project, environment string) string {
	parts := strings.SplitN(project, "/", 2)
	name := project
	if len(parts) == 2 {
		name = parts[1]
	}
	return name + "-" + environment
}

// enrichSyncStatus queries the deployer for each environment's ArgoCD sync status.
// Best-effort: logs warnings on failure and leaves status as "UNKNOWN".
func (c *Client) enrichSyncStatus(ctx context.Context, proj *Project) {
	for i := range proj.Environments {
		env := &proj.Environments[i]
		resp, err := c.Deployer.GetDeploymentStatus(ctx, &deployer.GetDeploymentStatusRequest{
			Project:     proj.ID,
			Environment: env.Name,
		})
		if err != nil {
			slog.Debug("failed to get deployment status", "project", proj.ID, "environment", env.Name, "error", err)
			continue
		}
		env.SyncStatus = deploymentStatusToString(resp.Status)
	}
}

func deploymentStatusToString(status deployer.DeploymentStatus) string {
	switch status {
	case deployer.DeploymentStatus_DEPLOYMENT_STATUS_SYNCED:
		return "SYNCED"
	case deployer.DeploymentStatus_DEPLOYMENT_STATUS_OUT_OF_SYNC:
		return "OUT_OF_SYNC"
	case deployer.DeploymentStatus_DEPLOYMENT_STATUS_PROGRESSING:
		return "PROGRESSING"
	case deployer.DeploymentStatus_DEPLOYMENT_STATUS_DEGRADED:
		return "DEGRADED"
	default:
		return "UNKNOWN"
	}
}

func projectFromProto(p *packager.ProjectInfo) Project {
	createdAt, _ := time.Parse(time.RFC3339, p.CreatedAt)

	proj := Project{
		ID:        p.Name,
		Name:      p.Name,
		SourceURL: p.SourceUrl,
		CreatedAt: createdAt,
	}

	// Build a lookup of per-env service info from the richer EnvironmentInfos.
	envInfoMap := make(map[string][]*packager.ServiceInstanceInfo, len(p.EnvironmentInfos))
	for _, ei := range p.EnvironmentInfos {
		envInfoMap[ei.Name] = ei.Services
	}

	for _, envName := range p.Environments {
		env := Environment{
			ID:         p.Name + "/" + envName,
			Name:       envName,
			Namespace:  namespaceFor(p.Name, envName),
			SyncStatus: "UNKNOWN",
		}

		// Populate service instances from environment values
		for _, svc := range envInfoMap[envName] {
			env.Services = append(env.Services, ServiceInstance{
				Name:        svc.Name,
				Environment: envName,
				ImageTag:    svc.ImageTag,
				Replicas:    1, // default until we query K8s
			})
		}

		proj.Environments = append(proj.Environments, env)
	}

	for _, svc := range p.Services {
		proj.Services = append(proj.Services, Service{
			Name:      svc.Name,
			Image:     svc.Image,
			Port:      int(svc.Port),
			Public:    svc.Public,
			Framework: svc.Framework,
		})
	}

	// Cross-reference: collect all service instances across environments onto each Service
	for i, svc := range proj.Services {
		for _, env := range proj.Environments {
			for _, si := range env.Services {
				if si.Name == svc.Name {
					proj.Services[i].Instances = append(proj.Services[i].Instances, si)
				}
			}
		}
	}

	return proj
}
