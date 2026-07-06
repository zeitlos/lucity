package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/zeitlos/lucity/cli/internal/api"
	"github.com/zeitlos/lucity/cli/internal/session"
)

type server struct {
	manager *session.Manager
}

func Serve(ctx context.Context, manager *session.Manager, version string) error {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	s := &server{manager: manager}

	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "lucity",
		Title:   "Lucity",
		Version: version,
	}, &mcp.ServerOptions{
		Instructions: s.instructions(),
		Logger:       logger,
	})

	s.registerAccount(mcpServer)
	s.registerProject(mcpServer)
	s.registerService(mcpServer)
	s.registerResource(mcpServer)
	s.registerDeploy(mcpServer)

	return mcpServer.Run(ctx, &mcp.StdioTransport{})
}

func (s *server) instructions() string {
	workspace := s.manager.Workspace()
	var b strings.Builder

	b.WriteString("Lucity is a deploy platform: it builds source repositories or prebuilt images into services running on Kubernetes, with managed PostgreSQL databases, Redis-compatible key-value stores, S3-compatible buckets, and persistent volumes.\n\n")

	if workspace == "" {
		b.WriteString("WARNING: no active workspace is set. Tool calls will fail until one is selected. Run 'lucity workspace <id>' (list memberships with 'lucity account').\n\n")
	} else {
		b.WriteString("Active workspace: " + workspace + "\n")
	}
	b.WriteString("Platform URL: " + s.manager.APIURL() + "\n\n")

	b.WriteString("All IDs are '/'-separated composite strings whose first segment is the workspace: project=workspace/project, environment=workspace/project/environment, service=workspace/project/environment/service. Pass these IDs verbatim; single-segment values (a bare project or service name) are resolved against the active workspace.\n\n")

	b.WriteString("Canonical flow: create_project -> add_service -> set_variables/configure_service -> deploy -> get_deploy_status -> get_logs. Deploys are asynchronous: poll get_deploy_status and read get_logs when a phase fails.\n\n")

	b.WriteString("IMPORTANT — config changes do NOT need a rebuild. set_variables, configure_service (resources, port, scaling), start-command changes, and volume mounts each roll the service out automatically with its CURRENT image, applied in seconds. Only call deploy when the SOURCE CODE changed (deploy rebuilds from scratch, which takes minutes). So to fix an OOM, a missing env var, or a wrong start command: apply the change and poll get_deploy_status — do not deploy again.\n\n")

	b.WriteString("There is deliberately no delete tool. Removing projects, services, databases, or other resources happens in the Lucity dashboard.")

	return b.String()
}

func jsonResult(v any) (*mcp.CallToolResult, any, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}

func ptr[T any](v T) *T { return &v }

func (s *server) requireWorkspace() (string, error) {
	workspace := s.manager.Workspace()
	if workspace == "" {
		return "", errors.New("no active workspace — run 'lucity workspace <id>' (list with 'lucity account')")
	}
	return workspace, nil
}

func (s *server) scopedID(id string, segments int) (string, error) {
	workspace, err := s.requireWorkspace()
	if err != nil {
		return "", err
	}

	id = strings.Trim(strings.TrimSpace(id), "/")
	if id == "" {
		return "", errors.New("id must not be empty")
	}

	parts := strings.Split(id, "/")

	if len(parts) == segments-1 {
		return workspace + "/" + id, nil
	}

	if len(parts) == segments {
		if parts[0] != workspace {
			return "", fmt.Errorf("id belongs to workspace %s but your active workspace is %s — run 'lucity workspace %s'", parts[0], workspace, parts[0])
		}
		return id, nil
	}

	return "", fmt.Errorf("invalid id %q: expected %d '/'-separated segments (or %d relative to the active workspace)", id, segments, segments-1)
}

func (s *server) projectID(id string) (string, error)     { return s.scopedID(id, 2) }
func (s *server) environmentID(id string) (string, error) { return s.scopedID(id, 3) }
func (s *server) serviceID(id string) (string, error)     { return s.scopedID(id, 4) }
func (s *server) resourceID(id string) (string, error)    { return s.scopedID(id, 4) }

func environmentOfService(serviceID string) string {
	parts := strings.Split(serviceID, "/")
	if len(parts) < 4 {
		return serviceID
	}
	return strings.Join(parts[:3], "/")
}

func (s *server) query(ctx context.Context, operation, query string, variables map[string]any, out any) error {
	if err := s.manager.Client().GraphQL(ctx, query, variables, out); err != nil {
		return wrapGraphQL(operation, err)
	}
	return nil
}

func wrapGraphQL(operation string, err error) error {
	var requestErr *api.RequestError
	if errors.As(err, &requestErr) {
		message := requestErr.Error()
		lower := strings.ToLower(message)
		if strings.Contains(lower, "logto access token") || strings.Contains(lower, "github token") || strings.Contains(lower, "invalid_grant") {
			return fmt.Errorf("%s failed: your GitHub session for Lucity has expired — run `lucity login` in a terminal, then retry. (This is separate from get_account, which stays valid.)", operation)
		}
		if strings.Contains(lower, "not found") {
			return fmt.Errorf("%s failed: %s — check the id and its format (list_projects / get_project to discover valid ids); the id's first segment must be your active workspace", operation, message)
		}
		return fmt.Errorf("%s failed: %s", operation, message)
	}
	return fmt.Errorf("%s failed: %w", operation, err)
}
