package harness

import (
	"encoding/json"
	"path/filepath"

	"github.com/zeitlos/lucity/pkg/acp"
	"github.com/zeitlos/lucity/services/agent/internal/config"
)

var claudeBuiltinTools = []string{
	"Bash", "BashOutput", "KillShell", "Read", "Edit", "MultiEdit", "Write", "Glob", "Grep", "LS",
	"NotebookEdit", "WebFetch", "WebSearch", "TodoWrite", "Task", "Agent", "Skill", "ToolSearch",
}

type claudeCode struct {
	agent config.Agent
	paths Paths
}

type claudeSettings struct {
	Permissions                     claudePermissions `json:"permissions"`
	AllowManagedPermissionRulesOnly bool              `json:"allowManagedPermissionRulesOnly"`
}

type claudePermissions struct {
	Allow                        []string `json:"allow"`
	Ask                          []string `json:"ask"`
	Deny                         []string `json:"deny"`
	DisableBypassPermissionsMode string   `json:"disableBypassPermissionsMode"`
}

func (c *claudeCode) Configure() error {
	permissions := claudePermissions{
		Allow:                        append([]string{}, claudeBuiltinTools...),
		Ask:                          []string{},
		Deny:                         []string{"AskUserQuestion"},
		DisableBypassPermissionsMode: "disable",
	}
	for _, connector := range c.agent.Connectors {
		rule := "mcp__" + connector.Name
		if connector.Approval == config.Allow {
			permissions.Allow = append(permissions.Allow, rule)
		} else {
			permissions.Ask = append(permissions.Ask, rule)
		}
	}

	settings := claudeSettings{Permissions: permissions, AllowManagedPermissionRulesOnly: true}
	if err := writeJSON(filepath.Join(c.paths.Policy, "claude-code", "managed-settings.json"), settings); err != nil {
		return err
	}
	return writeInstructions(c.paths.Workdir, "CLAUDE.md", c.agent.Instructions)
}

func (c *claudeCode) Command(Routing) (acp.Command, error) {
	return acp.Command{
		Path: "claude-agent-acp",
		Dir:  c.paths.Workdir,
		Env: environment(c.paths.Home,
			"ANTHROPIC_AUTH_TOKEN="+c.agent.Gateway.Key,
			"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1",
			"DISABLE_AUTOUPDATER=1",
		),
	}, nil
}

func (c *claudeCode) MCPServers() []json.RawMessage {
	return httpMCPServers(c.agent)
}

func (c *claudeCode) SessionMeta(model Model) map[string]any {
	if model.Provider == "" {
		return nil
	}
	env := map[string]string{
		"ANTHROPIC_BASE_URL":   c.agent.ModelURL(model.Provider, config.Anthropic),
		"ANTHROPIC_AUTH_TOKEN": c.agent.Gateway.Key,
	}
	return map[string]any{"claudeCode": map[string]any{"options": map[string]any{"env": env, "promptSuggestions": false}}}
}

func (c *claudeCode) ModelValue(model Model) (string, error) {
	if err := requireAPI(c.agent, model, config.Anthropic); err != nil {
		return "", err
	}
	return model.Name, nil
}

func (c *claudeCode) Covers(Routing, Model) bool {
	return true
}

func (c *claudeCode) Route(routing Routing, _ Model) Routing {
	return routing
}
