package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/zeitlos/lucity/pkg/auth"
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
	Services   []DeployedService
}

type Service struct {
	Name   string
	Image  string
	Port   int
	Public bool
}

type DeployedService struct {
	Name     string
	ImageTag string
	Ready    bool
	Replicas int
}

func (c *Client) Projects(ctx context.Context) ([]Project, error) {
	ctx = auth.OutgoingContext(ctx)

	resp, err := c.Packager.ListProjects(ctx, &packager.ListProjectsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	result := make([]Project, 0, len(resp.Projects))
	for _, p := range resp.Projects {
		result = append(result, projectFromProto(p))
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

	return &Project{
		ID:        name,
		Name:      name,
		SourceURL: sourceURL,
		CreatedAt: time.Now(),
		Environments: []Environment{
			{
				ID:         name + "/development",
				Name:       "development",
				Namespace:  name + "-development",
				SyncStatus: "PROGRESSING",
			},
		},
		Services: []Service{
			{
				Name:  "gitops",
				Image: resp.GitopsRepoUrl,
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

	return &Environment{
		ID:         projectID + "/" + name,
		Name:       name,
		Namespace:  resp.Namespace,
		SyncStatus: "PROGRESSING",
	}, nil
}

func (c *Client) DeleteEnvironment(ctx context.Context, projectID, environment string) (bool, error) {
	ctx = auth.OutgoingContext(ctx)

	_, err := c.Packager.DeleteEnvironment(ctx, &packager.DeleteEnvironmentRequest{
		Project:     projectID,
		Environment: environment,
	})
	if err != nil {
		return false, fmt.Errorf("failed to delete environment: %w", err)
	}
	return true, nil
}

func (c *Client) Promote(ctx context.Context, projectID, service, fromEnv, toEnv string) (*DeployedService, error) {
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

	return &DeployedService{
		Name:     service,
		ImageTag: resp.ImageTag,
	}, nil
}

func projectFromProto(p *packager.ProjectInfo) Project {
	createdAt, _ := time.Parse(time.RFC3339, p.CreatedAt)

	proj := Project{
		ID:        p.Name,
		Name:      p.Name,
		SourceURL: p.SourceUrl,
		CreatedAt: createdAt,
	}

	for _, envName := range p.Environments {
		proj.Environments = append(proj.Environments, Environment{
			ID:         p.Name + "/" + envName,
			Name:       envName,
			Namespace:  p.Name + "-" + envName,
			SyncStatus: "UNKNOWN",
		})
	}

	return proj
}
