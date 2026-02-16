package handler

import (
	"context"
	"fmt"
	"time"
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

// mockProjects holds hardcoded project data for development.
var mockProjects = []Project{
	{
		ID:        "proj-1",
		Name:      "acme-api",
		SourceURL: "https://github.com/acme/api",
		CreatedAt: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
		Environments: []Environment{
			{
				ID: "env-1", Name: "development", Namespace: "acme-api-development",
				SyncStatus: "SYNCED",
				Services: []DeployedService{
					{Name: "api", ImageTag: "a1b2c3d", Ready: true, Replicas: 1},
					{Name: "worker", ImageTag: "a1b2c3d", Ready: true, Replicas: 1},
				},
			},
			{
				ID: "env-2", Name: "staging", Namespace: "acme-api-staging",
				SyncStatus: "SYNCED",
				Services: []DeployedService{
					{Name: "api", ImageTag: "f4e5d6c", Ready: true, Replicas: 2},
					{Name: "worker", ImageTag: "f4e5d6c", Ready: true, Replicas: 1},
				},
			},
			{
				ID: "env-3", Name: "production", Namespace: "acme-api-production",
				SyncStatus: "PROGRESSING",
				Services: []DeployedService{
					{Name: "api", ImageTag: "f4e5d6c", Ready: false, Replicas: 3},
					{Name: "worker", ImageTag: "d7e8f9a", Ready: true, Replicas: 2},
				},
			},
		},
		Services: []Service{
			{Name: "api", Image: "ghcr.io/acme/api", Port: 8080, Public: true},
			{Name: "worker", Image: "ghcr.io/acme/worker", Public: false},
		},
	},
	{
		ID:        "proj-2",
		Name:      "acme-frontend",
		SourceURL: "https://github.com/acme/frontend",
		CreatedAt: time.Date(2025, 2, 1, 14, 30, 0, 0, time.UTC),
		Environments: []Environment{
			{
				ID: "env-4", Name: "development", Namespace: "acme-frontend-development",
				SyncStatus: "SYNCED",
				Services: []DeployedService{
					{Name: "web", ImageTag: "b2c3d4e", Ready: true, Replicas: 1},
				},
			},
			{
				ID: "env-5", Name: "production", Namespace: "acme-frontend-production",
				SyncStatus: "SYNCED",
				Services: []DeployedService{
					{Name: "web", ImageTag: "a1b2c3d", Ready: true, Replicas: 2},
				},
			},
		},
		Services: []Service{
			{Name: "web", Image: "ghcr.io/acme/frontend", Port: 3000, Public: true},
		},
	},
}

func (c *Client) Projects(ctx context.Context) ([]Project, error) {
	return mockProjects, nil
}

func (c *Client) Project(ctx context.Context, id string) (*Project, error) {
	for i := range mockProjects {
		if mockProjects[i].ID == id {
			return &mockProjects[i], nil
		}
	}
	return nil, fmt.Errorf("project %q not found", id)
}

func (c *Client) CreateProject(ctx context.Context, name, sourceURL string) (*Project, error) {
	p := Project{
		ID:        fmt.Sprintf("proj-%d", len(mockProjects)+1),
		Name:      name,
		SourceURL: sourceURL,
		CreatedAt: time.Now(),
		Environments: []Environment{
			{
				ID:         fmt.Sprintf("env-%d", len(mockProjects)*10+1),
				Name:       "development",
				Namespace:  name + "-development",
				SyncStatus: "PROGRESSING",
			},
		},
	}
	mockProjects = append(mockProjects, p)
	return &p, nil
}

func (c *Client) DeleteProject(ctx context.Context, id string) (bool, error) {
	for i, p := range mockProjects {
		if p.ID == id {
			mockProjects = append(mockProjects[:i], mockProjects[i+1:]...)
			return true, nil
		}
	}
	return false, fmt.Errorf("project %q not found", id)
}

func (c *Client) CreateEnvironment(ctx context.Context, projectID, name, fromEnvironment string) (*Environment, error) {
	for i := range mockProjects {
		if mockProjects[i].ID == projectID {
			env := Environment{
				ID:         fmt.Sprintf("env-%d", len(mockProjects[i].Environments)+1),
				Name:       name,
				Namespace:  mockProjects[i].Name + "-" + name,
				SyncStatus: "PROGRESSING",
			}
			mockProjects[i].Environments = append(mockProjects[i].Environments, env)
			return &env, nil
		}
	}
	return nil, fmt.Errorf("project %q not found", projectID)
}

func (c *Client) DeleteEnvironment(ctx context.Context, projectID, environment string) (bool, error) {
	for i := range mockProjects {
		if mockProjects[i].ID == projectID {
			for j, env := range mockProjects[i].Environments {
				if env.Name == environment {
					mockProjects[i].Environments = append(mockProjects[i].Environments[:j], mockProjects[i].Environments[j+1:]...)
					return true, nil
				}
			}
			return false, fmt.Errorf("environment %q not found in project %q", environment, projectID)
		}
	}
	return false, fmt.Errorf("project %q not found", projectID)
}

func (c *Client) Promote(ctx context.Context, projectID, service, fromEnv, toEnv string) (*DeployedService, error) {
	for i := range mockProjects {
		if mockProjects[i].ID != projectID {
			continue
		}
		// Find the image tag in the source environment
		var imageTag string
		for _, env := range mockProjects[i].Environments {
			if env.Name == fromEnv {
				for _, svc := range env.Services {
					if svc.Name == service {
						imageTag = svc.ImageTag
						break
					}
				}
			}
		}
		if imageTag == "" {
			return nil, fmt.Errorf("service %q not found in environment %q", service, fromEnv)
		}
		// Apply the image tag in the target environment
		for j := range mockProjects[i].Environments {
			if mockProjects[i].Environments[j].Name == toEnv {
				for k := range mockProjects[i].Environments[j].Services {
					if mockProjects[i].Environments[j].Services[k].Name == service {
						mockProjects[i].Environments[j].Services[k].ImageTag = imageTag
						return &mockProjects[i].Environments[j].Services[k], nil
					}
				}
			}
		}
		return nil, fmt.Errorf("service %q not found in environment %q", service, toEnv)
	}
	return nil, fmt.Errorf("project %q not found", projectID)
}
