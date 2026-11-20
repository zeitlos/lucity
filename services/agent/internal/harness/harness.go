package harness

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/zeitlos/lucity/pkg/acp"
	"github.com/zeitlos/lucity/services/agent/internal/config"
)

const keyVariable = "AGENT_GATEWAY_KEY"

// Paths are the directories a harness reads and writes. Policy holds the
// harness's managed config, which the agent itself must not be able to change.
type Paths struct {
	Home    string
	Workdir string
	Policy  string
}

type Model struct {
	Provider string     `json:"provider"`
	API      config.API `json:"api"`
	Name     string     `json:"name"`
}

// Routing is the set of model providers a harness process was started for.
type Routing struct {
	Providers map[string]Route `json:"providers,omitempty"`
	Active    string           `json:"active,omitempty"`
	Model     string           `json:"model,omitempty"`
}

type Route struct {
	API    config.API `json:"api"`
	Models []string   `json:"models,omitempty"`
}

type Harness interface {
	// Configure writes the harness's policy and instruction files.
	Configure() error
	Command(routing Routing) (acp.Command, error)
	MCPServers() []json.RawMessage
	SessionMeta(model Model) map[string]any
	// Prepare readies the harness to create or resume a session on model.
	Prepare(model Model) error
	ModelValue(model Model) (string, error)
	// Covers reports whether a harness started with routing can serve model.
	Covers(routing Routing, model Model) bool
	// Route returns the routing a restarted harness needs to also serve model.
	Route(routing Routing, model Model) Routing
}

func New(agent config.Agent, paths Paths) (Harness, error) {
	switch agent.Harness {
	case config.ClaudeCode:
		return &claudeCode{agent: agent, paths: paths}, nil
	case config.Codex:
		return &codex{agent: agent, paths: paths}, nil
	case config.OpenCode:
		return &openCode{agent: agent, paths: paths}, nil
	}
	return nil, fmt.Errorf("unknown harness %q", agent.Harness)
}

var passedVariables = []string{"PATH", "LANG", "LC_ALL", "TZ", "TMPDIR"}

func environment(home string, extra ...string) []string {
	variables := []string{"HOME=" + home}
	for _, name := range passedVariables {
		if value, ok := os.LookupEnv(name); ok {
			variables = append(variables, name+"="+value)
		}
	}
	return append(variables, extra...)
}

type mcpHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type mcpServer struct {
	Type    string      `json:"type"`
	Name    string      `json:"name"`
	URL     string      `json:"url"`
	Headers []mcpHeader `json:"headers"`
}

func httpMCPServers(agent config.Agent) []json.RawMessage {
	var servers []json.RawMessage
	for _, connector := range agent.Connectors {
		server, err := json.Marshal(mcpServer{
			Type:    "http",
			Name:    connector.Name,
			URL:     agent.ConnectorURL(connector.Name),
			Headers: []mcpHeader{{Name: "Authorization", Value: "Bearer " + agent.Gateway.Key}},
		})
		if err != nil {
			continue
		}
		servers = append(servers, server)
	}
	return servers
}

func requireAPI(agent config.Agent, model Model, apis ...config.API) error {
	if model.Provider == "" || model.Name == "" {
		return errors.New("model needs a provider and a name")
	}
	if !slices.Contains(apis, model.API) {
		return fmt.Errorf("%s can't use %s models", agent.Harness, model.API)
	}
	return nil
}

func withProvider(routing Routing, model Model) Routing {
	providers := maps.Clone(routing.Providers)
	if providers == nil {
		providers = map[string]Route{}
	}
	route := providers[model.Provider]
	if route.API != model.API {
		route = Route{API: model.API}
	}
	if !slices.Contains(route.Models, model.Name) {
		route.Models = append(slices.Clone(route.Models), model.Name)
	}
	providers[model.Provider] = route
	return Routing{Providers: providers, Active: model.Provider, Model: model.Name}
}

func writeInstructions(workdir, name, instructions string) error {
	if strings.TrimSpace(instructions) == "" {
		return nil
	}
	return write(filepath.Join(workdir, name), []byte(instructions))
}

func writeJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return write(path, append(raw, '\n'))
}

func write(path string, content []byte) error {
	existing, err := os.ReadFile(path)
	switch {
	case err == nil && bytes.Equal(existing, content):
		return nil
	case err != nil && !errors.Is(err, fs.ErrNotExist):
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}
