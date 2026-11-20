package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Harness string

const (
	ClaudeCode Harness = "claudeCode"
	Codex      Harness = "codex"
	OpenCode   Harness = "opencode"
)

type API string

const (
	Anthropic        API = "anthropic"
	OpenAI           API = "openai"
	OpenAICompatible API = "openaiCompatible"
)

type Approval string

const (
	Allow Approval = "allow"
	Ask   Approval = "ask"
)

// Agent is the agent's configuration, read from agent.json and instructions.md.
type Agent struct {
	Harness      Harness     `json:"harness"`
	Connectors   []Connector `json:"connectors"`
	Instructions string      `json:"-"`
	Gateway      Gateway     `json:"-"`
}

type Connector struct {
	Name     string   `json:"name"`
	Approval Approval `json:"approval"`
}

// Gateway is the workspace's gateway base URL and the agent's key for it.
type Gateway struct {
	URL string
	Key string
}

func Load(dir string, gateway Gateway) (Agent, error) {
	path := filepath.Join(dir, "agent.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return Agent{}, err
	}
	var agent Agent
	if err := json.Unmarshal(raw, &agent); err != nil {
		return Agent{}, fmt.Errorf("parse %s: %w", path, err)
	}

	instructions, err := os.ReadFile(filepath.Join(dir, "instructions.md"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return Agent{}, err
	}
	agent.Instructions = string(instructions)

	if !slices.Contains([]Harness{ClaudeCode, Codex, OpenCode}, agent.Harness) {
		return Agent{}, fmt.Errorf("unknown harness %q", agent.Harness)
	}
	for i, connector := range agent.Connectors {
		if connector.Approval != Allow {
			agent.Connectors[i].Approval = Ask
		}
	}

	base, err := url.Parse(strings.TrimSuffix(gateway.URL, "/"))
	if err != nil || (base.Scheme != "http" && base.Scheme != "https") {
		return Agent{}, fmt.Errorf("gateway url %q is not an http(s) url", gateway.URL)
	}
	agent.Gateway = Gateway{URL: base.String(), Key: gateway.Key}
	return agent, nil
}

// ConnectorURL is the gateway's MCP endpoint for a connector.
func (a Agent) ConnectorURL(name string) string {
	return a.Gateway.URL + "/connectors/" + url.PathEscape(name) + "/mcp"
}

// ModelURL is the gateway's base URL for a model provider, with /v1 for the OpenAI APIs.
func (a Agent) ModelURL(provider string, api API) string {
	base := a.Gateway.URL + "/models/" + url.PathEscape(provider)
	if api == Anthropic {
		return base
	}
	return base + "/v1"
}
