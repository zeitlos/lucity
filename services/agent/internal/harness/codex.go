package harness

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/zeitlos/lucity/pkg/acp"
	"github.com/zeitlos/lucity/services/agent/internal/config"
)

const codexProviderID = "gateway"

type codex struct {
	agent config.Agent
	paths Paths
}

func (c *codex) Configure() error {
	var b strings.Builder
	fmt.Fprintf(&b, "approval_policy = %s\n", tomlString("on-request"))
	fmt.Fprintf(&b, "sandbox_mode = %s\n", tomlString("danger-full-access"))
	for _, connector := range c.agent.Connectors {
		approval := "prompt"
		if connector.Approval == config.Allow {
			approval = "approve"
		}
		fmt.Fprintf(&b, "\n[mcp_servers.%s]\n", tomlString(connector.Name))
		fmt.Fprintf(&b, "url = %s\n", tomlString(c.agent.ConnectorURL(connector.Name)))
		fmt.Fprintf(&b, "bearer_token_env_var = %s\n", tomlString(keyVariable))
		fmt.Fprintf(&b, "default_tools_approval_mode = %s\n", tomlString(approval))
	}

	if err := write(filepath.Join(c.paths.Policy, "codex", "config.toml"), []byte(b.String())); err != nil {
		return err
	}
	return writeInstructions(c.paths.Workdir, "AGENTS.md", c.agent.Instructions)
}

func (c *codex) Command(routing Routing) (acp.Command, error) {
	var b strings.Builder
	if route, ok := routing.Providers[routing.Active]; ok {
		fmt.Fprintf(&b, "model_provider = %s\n", tomlString(codexProviderID))
		fmt.Fprintf(&b, "model = %s\n", tomlString(routing.Model))
		fmt.Fprintf(&b, "\n[model_providers.%s]\n", codexProviderID)
		fmt.Fprintf(&b, "name = %s\n", tomlString(routing.Active))
		fmt.Fprintf(&b, "base_url = %s\n", tomlString(c.agent.ModelURL(routing.Active, route.API)))
		fmt.Fprintf(&b, "env_key = %s\n", tomlString(keyVariable))
		fmt.Fprintf(&b, "wire_api = %s\n", tomlString("responses"))
	}
	if err := write(filepath.Join(c.paths.Home, ".codex", "config.toml"), []byte(b.String())); err != nil {
		return acp.Command{}, err
	}
	return acp.Command{
		Path: "codex-acp",
		Dir:  c.paths.Workdir,
		Env:  environment(c.paths.Home, keyVariable+"="+c.agent.Gateway.Key),
	}, nil
}

func (c *codex) MCPServers() []json.RawMessage {
	return nil
}

func (c *codex) SessionMeta(Model) map[string]any {
	return nil
}

func (c *codex) ModelValue(model Model) (string, error) {
	if err := requireAPI(c.agent, model, config.OpenAI); err != nil {
		return "", err
	}
	return model.Name, nil
}

func (c *codex) Prepare(Model) error {
	return nil
}

func (c *codex) Covers(routing Routing, model Model) bool {
	return routing.Active == model.Provider && routing.Providers[model.Provider].API == model.API && routing.Model == model.Name
}

func (c *codex) Route(routing Routing, model Model) Routing {
	return withProvider(routing, model)
}

func tomlString(value string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range value {
		switch {
		case r == '"' || r == '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r < 0x20 || r == 0x7f:
			fmt.Fprintf(&b, "\\u%04X", r)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}
