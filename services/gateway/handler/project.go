package handler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/builder"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/pkg/packager"
)

// gRPC call timeouts. Short for quick lookups, long for operations that
// clone repos or touch external systems.
const (
	grpcTimeout     = 10 * time.Second
	grpcLongTimeout = 60 * time.Second
)

type Project struct {
	ID             string
	Name           string
	SourceURL      string
	Environments   []Environment
	Services       []Service
	InitialDeploys []DeployOp
	CreatedAt      time.Time
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
	Deployments []Deployment
}

type Deployment struct {
	ID        string
	ImageTag  string
	Active    bool
	Timestamp time.Time
	Revision  string
	Message   string
}

func (c *Client) Projects(ctx context.Context) ([]Project, error) {
	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.Packager.ListProjects(callCtx, &packager.ListProjectsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	result := make([]Project, 0, len(resp.Projects))
	for _, p := range resp.Projects {
		proj := projectFromProto(p)
		c.enrichSyncStatus(ctx, &proj)
		c.enrichDeploymentHistory(ctx, &proj)
		result = append(result, proj)
	}
	return result, nil
}

func (c *Client) Project(ctx context.Context, id string) (*Project, error) {
	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.Packager.GetProject(callCtx, &packager.GetProjectRequest{Project: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	p := projectFromProto(resp.Project)
	c.enrichSyncStatus(ctx, &p)
	c.enrichDeploymentHistory(ctx, &p)
	return &p, nil
}

func (c *Client) CreateProject(ctx context.Context, name, sourceURL string) (*Project, error) {
	ctx = auth.OutgoingContext(ctx)

	// 1. Create GitOps repo
	initCtx, initCancel := context.WithTimeout(ctx, grpcLongTimeout)
	defer initCancel()
	resp, err := c.Packager.InitProject(initCtx, &packager.InitProjectRequest{
		Project:   name,
		SourceUrl: sourceURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// 2. Detect services from source repo (non-fatal on failure)
	var services []Service
	detectCtx, detectCancel := context.WithTimeout(ctx, grpcLongTimeout)
	defer detectCancel()
	detectResp, err := c.Builder.DetectServices(detectCtx, &builder.DetectServicesRequest{
		SourceUrl: sourceURL,
	})
	if err != nil {
		slog.Warn("failed to detect services", "project", name, "error", err)
	} else {
		for _, detected := range detectResp.Services {
			image := deriveImagePath(c.RegistryImagePrefix, name, detected.Name)
			addCtx, addCancel := context.WithTimeout(ctx, grpcTimeout)
			_, addErr := c.Packager.AddService(addCtx, &packager.AddServiceRequest{
				Project:   name,
				Service:   detected.Name,
				Image:     image,
				Port:      detected.SuggestedPort,
				Public:    true,
				Framework: detected.Framework,
			})
			addCancel()
			if addErr != nil {
				slog.Warn("failed to add detected service", "project", name, "service", detected.Name, "error", addErr)
				continue
			}
			services = append(services, Service{
				Name:      detected.Name,
				Image:     image,
				Port:      int(detected.SuggestedPort),
				Public:    true,
				Framework: detected.Framework,
			})
		}
	}

	// 3. Deploy the default development environment via ArgoCD
	ns := labels.NamespaceFor(name, "development")
	deployCtx, deployCancel := context.WithTimeout(ctx, grpcTimeout)
	defer deployCancel()
	_, err = c.Deployer.DeployEnvironment(deployCtx, &deployer.DeployEnvironmentRequest{
		Project:         name,
		Environment:     "development",
		GitopsRepoUrl:   resp.GitopsRepoUrl,
		TargetNamespace: ns,
	})
	if err != nil {
		slog.Warn("failed to deploy development environment", "project", name, "error", err)
	}

	// 4. Auto-deploy each detected service
	var initialDeploys []DeployOp
	for _, svc := range services {
		deployOp, deployErr := c.Deploy(ctx, name, svc.Name, "development", "", "")
		if deployErr != nil {
			slog.Warn("failed to start initial deploy", "project", name, "service", svc.Name, "error", deployErr)
			continue
		}
		initialDeploys = append(initialDeploys, *deployOp)
	}

	return &Project{
		ID:        name,
		Name:      name,
		SourceURL: sourceURL,
		CreatedAt: time.Now(),
		Services:  services,
		Environments: []Environment{
			{
				ID:         name + "/development",
				Name:       "development",
				Namespace:  ns,
				SyncStatus: "PROGRESSING",
			},
		},
		InitialDeploys: initialDeploys,
	}, nil
}

func (c *Client) DeleteProject(ctx context.Context, id string) (bool, error) {
	ctx = auth.OutgoingContext(ctx)

	// 1. Fetch project to discover all environments
	getCtx, getCancel := context.WithTimeout(ctx, grpcTimeout)
	defer getCancel()
	resp, err := c.Packager.GetProject(getCtx, &packager.GetProjectRequest{Project: id})
	if err != nil {
		return false, fmt.Errorf("failed to get project for deletion: %w", err)
	}

	// 2. Remove ArgoCD Application for each environment (best-effort)
	for _, env := range resp.Project.Environments {
		rmCtx, rmCancel := context.WithTimeout(ctx, grpcTimeout)
		_, err := c.Deployer.RemoveDeployment(rmCtx, &deployer.RemoveDeploymentRequest{
			Project:     id,
			Environment: env,
		})
		rmCancel()
		if err != nil {
			slog.Warn("failed to remove deployment during project deletion",
				"project", id, "environment", env, "error", err)
		}
	}

	// 3. Remove ArgoCD repository credential (best-effort)
	repoCtx, repoCancel := context.WithTimeout(ctx, grpcTimeout)
	defer repoCancel()
	_, err = c.Deployer.DeleteRepository(repoCtx, &deployer.DeleteRepositoryRequest{
		Project: id,
	})
	if err != nil {
		slog.Warn("failed to delete ArgoCD repository credential",
			"project", id, "error", err)
	}

	// 4. Delete GitOps repo
	delCtx, delCancel := context.WithTimeout(ctx, grpcTimeout)
	defer delCancel()
	_, err = c.Packager.DeleteProject(delCtx, &packager.DeleteProjectRequest{Project: id})
	if err != nil {
		return false, fmt.Errorf("failed to delete project: %w", err)
	}
	return true, nil
}

func (c *Client) CreateEnvironment(ctx context.Context, projectID, name, fromEnvironment string) (*Environment, error) {
	ctx = auth.OutgoingContext(ctx)

	createCtx, createCancel := context.WithTimeout(ctx, grpcLongTimeout)
	defer createCancel()
	resp, err := c.Packager.CreateEnvironment(createCtx, &packager.CreateEnvironmentRequest{
		Project:         projectID,
		Environment:     name,
		FromEnvironment: fromEnvironment,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	// Deploy the new environment via ArgoCD
	envDeployCtx, envDeployCancel := context.WithTimeout(ctx, grpcTimeout)
	defer envDeployCancel()
	_, err = c.Deployer.DeployEnvironment(envDeployCtx, &deployer.DeployEnvironmentRequest{
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
	rmCtx, rmCancel := context.WithTimeout(ctx, grpcTimeout)
	defer rmCancel()
	_, err := c.Deployer.RemoveDeployment(rmCtx, &deployer.RemoveDeploymentRequest{
		Project:     projectID,
		Environment: environment,
	})
	if err != nil {
		slog.Warn("failed to remove deployment", "project", projectID, "environment", environment, "error", err)
	}

	// Then remove from GitOps repo
	delEnvCtx, delEnvCancel := context.WithTimeout(ctx, grpcTimeout)
	defer delEnvCancel()
	_, err = c.Packager.DeleteEnvironment(delEnvCtx, &packager.DeleteEnvironmentRequest{
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

	promoteCtx, promoteCancel := context.WithTimeout(ctx, grpcLongTimeout)
	defer promoteCancel()
	resp, err := c.Packager.Promote(promoteCtx, &packager.PromoteRequest{
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

// enrichSyncStatus queries the deployer for each environment's ArgoCD sync status.
// Best-effort: logs warnings on failure and leaves status as "UNKNOWN".
// Calls are made concurrently to avoid serial N+1 latency.
func (c *Client) enrichSyncStatus(ctx context.Context, proj *Project) {
	var wg sync.WaitGroup
	for i := range proj.Environments {
		env := &proj.Environments[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			statusCtx, statusCancel := context.WithTimeout(ctx, grpcTimeout)
			defer statusCancel()
			resp, err := c.Deployer.GetDeploymentStatus(statusCtx, &deployer.GetDeploymentStatusRequest{
				Project:     proj.ID,
				Environment: env.Name,
			})
			if err != nil {
				slog.Debug("failed to get deployment status", "project", proj.ID, "environment", env.Name, "error", err)
				return
			}
			env.SyncStatus = deploymentStatusToString(resp.Status)

			// Derive per-service readiness from environment health.
			ready := resp.Status == deployer.DeploymentStatus_DEPLOYMENT_STATUS_SYNCED
			for j := range env.Services {
				env.Services[j].Ready = ready
			}
		}()
	}
	wg.Wait()
}

// enrichDeploymentHistory fetches deployment history from the packager for each
// service instance in every environment and attaches it.
// Calls are made concurrently — each goroutine writes to its own ServiceInstance.
func (c *Client) enrichDeploymentHistory(ctx context.Context, proj *Project) {
	var wg sync.WaitGroup
	for i := range proj.Environments {
		env := &proj.Environments[i]
		for j := range env.Services {
			si := &env.Services[j]
			wg.Add(1)
			go func() {
				defer wg.Done()
				histCtx, histCancel := context.WithTimeout(ctx, grpcTimeout)
				defer histCancel()
				resp, err := c.Packager.DeploymentHistory(histCtx, &packager.DeploymentHistoryRequest{
					Project:     proj.ID,
					Environment: env.Name,
					Service:     si.Name,
				})
				if err != nil {
					slog.Debug("failed to get deployment history", "project", proj.ID, "environment", env.Name, "service", si.Name, "error", err)
					return
				}

				for k, e := range resp.Entries {
					deployedAt := e.DeployedAt.AsTime()
					si.Deployments = append(si.Deployments, Deployment{
						ID:        fmt.Sprintf("%s/%s/%s/%d", proj.ID, env.Name, si.Name, k),
						ImageTag:  e.ImageTag,
						Active:    k == 0,
						Timestamp: deployedAt,
						Revision:  e.Revision,
						Message:   fmt.Sprintf("deploy(%s): %s %s", env.Name, si.Name, e.ImageTag),
					})
				}
			}()
		}
	}
	wg.Wait()

	// Also attach to the cross-referenced Service.Instances
	for i, svc := range proj.Services {
		for j := range svc.Instances {
			inst := &proj.Services[i].Instances[j]
			for _, env := range proj.Environments {
				for _, esi := range env.Services {
					if esi.Name == inst.Name && esi.Environment == inst.Environment {
						inst.Deployments = esi.Deployments
						inst.Ready = esi.Ready
					}
				}
			}
		}
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
	createdAt := p.CreatedAt.AsTime()

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
			Namespace:  labels.NamespaceFor(p.Name, envName),
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
