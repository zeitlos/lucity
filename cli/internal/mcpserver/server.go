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
	"github.com/zeitlos/lucity/cli/internal/ids"
	"github.com/zeitlos/lucity/cli/internal/selfupdate"
	"github.com/zeitlos/lucity/cli/internal/session"
)

type server struct {
	manager *session.Manager
}

func Serve(ctx context.Context, manager *session.Manager, version string) error {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	s := &server{manager: manager}

	var latest string
	if !selfupdate.Disabled() {
		cached, stale := selfupdate.Cached()
		latest = cached
		if stale && !selfupdate.AutoUpdates(version) {
			go selfupdate.Refresh(ctx)
		}
	}

	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "lucity",
		Title:   "Lucity",
		Version: version,
	}, &mcp.ServerOptions{
		Instructions: s.instructions(version, latest),
		Logger:       logger,
	})

	s.registerAccount(mcpServer)
	s.registerProject(mcpServer)
	s.registerService(mcpServer)
	s.registerResource(mcpServer)
	s.registerDeploy(mcpServer)

	err := mcpServer.Run(ctx, &mcp.StdioTransport{})
	if _, stale := selfupdate.Cached(); stale && selfupdate.AutoUpdates(version) {
		selfupdate.StartBackground()
	}
	return err
}

func (s *server) instructions(version, latest string) string {
	workspace := s.manager.Workspace()
	var b strings.Builder

	b.WriteString("Lucity is a deploy platform: it builds source repositories or prebuilt images into services running on Kubernetes, with managed PostgreSQL databases, Redis-compatible key-value stores, S3-compatible buckets, and persistent volumes.\n\n")

	if workspace == "" {
		b.WriteString("WARNING: no active workspace is set. Tool calls will fail until one is selected. Run 'lucity workspace <id>' (list memberships with 'lucity account').\n\n")
	} else {
		b.WriteString("Active workspace: " + workspace + " (it follows 'lucity workspace <id>' from the next call on; get_account shows the current one)\n")
	}
	b.WriteString("Platform URL: " + s.manager.APIURL() + "\n\n")

	if notice := selfupdate.Notice(version, latest); notice != "" {
		fmt.Fprintf(&b, "UPDATE AVAILABLE: %s. Tell the user once, early on, and that a new session picks up the new version, since tools may have changed.\n\n", notice)
	}

	b.WriteString("All IDs are '/'-separated composite strings whose first segment is the workspace: project=workspace/project, environment=workspace/project/environment, service=workspace/project/environment/service. Pass these IDs verbatim. The workspace segment may be left out (project/environment/service), and a project may be given as its bare id. Every resource also has a name, which is what people read: it can contain spaces and can be renamed, so never pass a name where an id is expected. Create tools take a name and an optional id; leave the id out to derive it from the name, and use the id the tool returns.\n\n")

	b.WriteString("Canonical flow: create_project -> add_service -> set_variables/configure_service -> get_deploy_status -> get_logs. add_service already starts a repository service's first build (an image service rolls out without one), so poll get_deploy_status with the release_id it returns instead of calling deploy. Deploys are asynchronous: poll get_deploy_status and read get_logs when a phase fails.\n\n")

	b.WriteString("IMPORTANT — config changes do NOT need a rebuild. set_variables, configure_service (resources, port, scaling), start-command changes, and volume mounts each roll the service out automatically with its CURRENT image, applied in seconds. Only call deploy when the SOURCE CODE changed, or a variable the app reads during its build (RAILPACK_* builder settings, or values a frontend framework inlines into its bundle); deploy rebuilds from scratch, which takes minutes. Builds only see literal values set on the service itself, never shared variables or refs. So to fix an OOM, a missing runtime env var, or a wrong start command: apply the change and poll get_deploy_status — do not deploy again.\n\n")

	b.WriteString("Secret values never enter this conversation. list_variables shows keys and refs but never values, and check_variables tells you whether a variable holds the value you expect. set_variables with generate stores a random app secret (SECRET_KEY_BASE, a JWT secret) that you never see. Database, key-value store, and bucket credentials reach services only as refs, and runtime logs come back with known secrets redacted. Never ask the user to paste a secret here: they add third-party keys in the Lucity dashboard (the service's Variables tab), and you confirm the keys exist with list_variables. Bulk database dump imports need credentials, so the user runs them from their own terminal.\n\n")

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
	return ids.Scoped(workspace, id, segments)
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
		if lower == "unauthenticated" {
			return fmt.Errorf("%s failed: the platform did not accept your Lucity sign-in — run `lucity login` in a terminal, then retry", operation)
		}
		if strings.Contains(lower, "github token") {
			return fmt.Errorf("%s failed: your GitHub session for Lucity has expired — run `lucity login` in a terminal, then retry. (This is separate from get_account, which stays valid.)", operation)
		}
		if strings.Contains(lower, "another operation (install/upgrade/rollback) is in progress") {
			return fmt.Errorf("%s failed: %s — the environment is still applying another change, such as a build's rollout or an earlier config change; retry in a few seconds", operation, message)
		}
		if strings.Contains(lower, `" not found`) {
			return fmt.Errorf("%s failed: %s — check the id and its format (list_projects / get_project to discover valid ids); the id's first segment must be your active workspace", operation, message)
		}
		return fmt.Errorf("%s failed: %s", operation, message)
	}
	return fmt.Errorf("%s failed: %w", operation, err)
}
