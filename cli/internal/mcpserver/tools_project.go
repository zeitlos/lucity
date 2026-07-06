package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *server) registerAccount(m *mcp.Server) {
	mcp.AddTool(m, &mcp.Tool{
		Name:        "get_account",
		Description: "Show the signed-in identity, the workspaces you belong to, the active workspace, and the platform URL.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: ptr(false)},
	}, s.getAccount)
}

type emptyInput struct{}

func (s *server) getAccount(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	identity, err := s.manager.Client().Me(ctx)
	if err != nil {
		return nil, nil, err
	}

	return jsonResult(map[string]any{
		"name":            identity.Name,
		"email":           identity.Email,
		"workspaces":      identity.Workspaces,
		"activeWorkspace": s.manager.Workspace(),
		"platformUrl":     s.manager.APIURL(),
	})
}

func (s *server) registerProject(m *mcp.Server) {
	mcp.AddTool(m, &mcp.Tool{
		Name:        "list_projects",
		Description: "List all projects in the active workspace with their environments.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: ptr(false)},
	}, s.listProjects)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "get_project",
		Description: "Get a project with its environments and everything in them: services (status, replicas, endpoints, resources), databases, key-value stores, buckets, and volumes.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: ptr(false)},
	}, s.getProject)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "detect_services",
		Description: "Inspect a source repository and detect deployable services (language, framework, context path, start command, suggested port). Pass the URL of a GitHub repository accessible to the workspace's GitHub installation, e.g. https://github.com/owner/repo.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: ptr(false)},
	}, s.detectServices)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "list_github_repos",
		Description: "Find deployable GitHub repositories and check the Lucity GitHub app installation. Without account: lists the connected GitHub installations. With account: lists that account's repositories. If an expected account is missing, the user must install the Lucity GitHub app via the dashboard project-creation flow.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: ptr(false)},
	}, s.listGithubRepos)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "create_project",
		Description: "Create a project. A 'development' environment is created automatically.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.createProject)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "create_environment",
		Description: "Create an environment in a project. Tier ECO is cost-optimized; PRODUCTION reserves more resources.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.createEnvironment)
}

func (s *server) listProjects(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	if _, err := s.requireWorkspace(); err != nil {
		return nil, nil, err
	}

	const query = `query { projects { id name environments { id name } } }`

	var out struct {
		Projects []any `json:"projects"`
	}
	if err := s.query(ctx, "list_projects", query, nil, &out); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{"projects": out.Projects})
}

type getProjectInput struct {
	Project string `json:"project" jsonschema:"project id (workspace/project) or a bare project name resolved against the active workspace"`
}

func (s *server) getProject(ctx context.Context, _ *mcp.CallToolRequest, input getProjectInput) (*mcp.CallToolResult, any, error) {
	id, err := s.projectID(input.Project)
	if err != nil {
		return nil, nil, err
	}

	const query = `query($id: ProjectID!) {
  project(id: $id) {
    id name
    environments {
      id name resourceTier
      services { id name status replicas { desired ready } port command resources { cpu memory } endpoints { host protocol type } lastDeployedAt }
      databases { id name status version size public }
      keyValueStores { id name status }
      buckets { id name status public }
      volumes { id name size mount { service path } }
    }
  }
}`

	var out struct {
		Project any `json:"project"`
	}
	if err := s.query(ctx, "get_project", query, map[string]any{"id": id}, &out); err != nil {
		if splitErr := s.getProjectSplit(ctx, id, &out); splitErr == nil {
			return jsonResult(out.Project)
		}
		return nil, nil, err
	}
	return jsonResult(out.Project)
}

func (s *server) getProjectSplit(ctx context.Context, id string, out *struct {
	Project any `json:"project"`
}) error {
	const coreQuery = `query($id: ProjectID!) {
  project(id: $id) {
    id name
    environments {
      id name resourceTier
      services { id name status replicas { desired ready } port command resources { cpu memory } endpoints { host protocol type } lastDeployedAt }
    }
  }
}`

	const dataQuery = `query($id: ProjectID!) {
  project(id: $id) {
    environments {
      id
      databases { id name status version size public }
      keyValueStores { id name status }
      buckets { id name status public }
      volumes { id name size mount { service path } }
    }
  }
}`

	var core struct {
		Project struct {
			ID           string           `json:"id"`
			Name         string           `json:"name"`
			Environments []map[string]any `json:"environments"`
		} `json:"project"`
	}
	if err := s.query(ctx, "get_project", coreQuery, map[string]any{"id": id}, &core); err != nil {
		return err
	}

	var data struct {
		Project struct {
			Environments []map[string]any `json:"environments"`
		} `json:"project"`
	}
	if err := s.query(ctx, "get_project", dataQuery, map[string]any{"id": id}, &data); err != nil {
		return err
	}

	byID := make(map[string]map[string]any, len(data.Project.Environments))
	for _, env := range data.Project.Environments {
		if envID, ok := env["id"].(string); ok {
			byID[envID] = env
		}
	}
	for _, env := range core.Project.Environments {
		envID, _ := env["id"].(string)
		extra := byID[envID]
		if extra == nil {
			continue
		}
		for _, key := range []string{"databases", "keyValueStores", "buckets", "volumes"} {
			env[key] = extra[key]
		}
	}

	out.Project = map[string]any{
		"id":           core.Project.ID,
		"name":         core.Project.Name,
		"environments": core.Project.Environments,
	}
	return nil
}

type detectServicesInput struct {
	RepositoryURL string `json:"repository_url" jsonschema:"URL of a GitHub repository accessible to the workspace's GitHub installation, e.g. https://github.com/owner/repo"`
}

func (s *server) detectServices(ctx context.Context, _ *mcp.CallToolRequest, input detectServicesInput) (*mcp.CallToolResult, any, error) {
	if _, err := s.requireWorkspace(); err != nil {
		return nil, nil, err
	}

	const query = `query($url: String!) {
  detectServices(repositoryUrl: $url) { name language framework contextPath startCommand suggestedPort }
}`

	var out struct {
		DetectServices []any `json:"detectServices"`
	}
	if err := s.query(ctx, "detect_services", query, map[string]any{"url": input.RepositoryURL}, &out); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{"services": out.DetectServices})
}

type listGithubReposInput struct {
	Account string `json:"account,omitempty" jsonschema:"optional GitHub account login; omit to list connected installations, provide to list that account's repositories"`
}

func (s *server) listGithubRepos(ctx context.Context, _ *mcp.CallToolRequest, input listGithubReposInput) (*mcp.CallToolResult, any, error) {
	if _, err := s.requireWorkspace(); err != nil {
		return nil, nil, err
	}

	if input.Account == "" {
		const query = `query { githubSources { accountLogin accountAvatarUrl accountType } }`
		var out struct {
			GithubSources []any `json:"githubSources"`
		}
		if err := s.query(ctx, "list_github_repos", query, nil, &out); err != nil {
			return nil, nil, err
		}
		return jsonResult(map[string]any{
			"installations": out.GithubSources,
			"note":          "pass one account login to list its repositories; if the account you want is missing, install the Lucity GitHub app via the dashboard project-creation flow",
		})
	}

	const query = `query($account: String!) {
  githubRepositories(account: $account) { name fullName htmlUrl defaultBranch private }
}`
	var out struct {
		GithubRepositories []any `json:"githubRepositories"`
	}
	if err := s.query(ctx, "list_github_repos", query, map[string]any{"account": input.Account}, &out); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{"repositories": out.GithubRepositories})
}

type createProjectInput struct {
	Name string `json:"name" jsonschema:"human-readable project name (1-128 chars)"`
	ID   string `json:"id,omitempty" jsonschema:"optional URL-safe slug (2-16 chars, lowercase alphanumeric and hyphens); derived from the name if omitted"`
}

func (s *server) createProject(ctx context.Context, _ *mcp.CallToolRequest, input createProjectInput) (*mcp.CallToolResult, any, error) {
	if _, err := s.requireWorkspace(); err != nil {
		return nil, nil, err
	}

	inputVar := map[string]any{"name": input.Name}
	if input.ID != "" {
		inputVar["id"] = input.ID
	}

	const mutation = `mutation($input: CreateProjectInput!) {
  createProject(input: $input) { id name environments { id name } }
}`

	var out struct {
		CreateProject any `json:"createProject"`
	}
	if err := s.query(ctx, "create_project", mutation, map[string]any{"input": inputVar}, &out); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{
		"project": out.CreateProject,
		"note":    "a 'development' environment was created automatically; add services to it with add_service",
	})
}

type createEnvironmentInput struct {
	Project string `json:"project" jsonschema:"project id (workspace/project) or a bare project name"`
	Name    string `json:"name" jsonschema:"environment name (2-16 chars, lowercase alphanumeric and hyphens)"`
	Tier    string `json:"tier,omitempty" jsonschema:"resource tier ECO or PRODUCTION; defaults to the platform default"`
}

func (s *server) createEnvironment(ctx context.Context, _ *mcp.CallToolRequest, input createEnvironmentInput) (*mcp.CallToolResult, any, error) {
	projectID, err := s.projectID(input.Project)
	if err != nil {
		return nil, nil, err
	}

	inputVar := map[string]any{"project": projectID, "name": input.Name}
	if input.Tier != "" {
		inputVar["tier"] = input.Tier
	}

	const mutation = `mutation($input: CreateEnvironmentInput!) {
  createEnvironment(input: $input) { id name resourceTier }
}`

	var out struct {
		CreateEnvironment any `json:"createEnvironment"`
	}
	if err := s.query(ctx, "create_environment", mutation, map[string]any{"input": inputVar}, &out); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{"environment": out.CreateEnvironment})
}
