package harness

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"slices"

	"github.com/zeitlos/lucity/pkg/acp"
	"github.com/zeitlos/lucity/services/agent/internal/config"
)

var openCodePackages = map[config.API]string{
	config.Anthropic:        "@ai-sdk/anthropic",
	config.OpenAI:           "@ai-sdk/openai",
	config.OpenAICompatible: "@ai-sdk/openai-compatible",
}

type openCode struct {
	agent config.Agent
	paths Paths
}

type openCodePolicy struct {
	Schema     string       `json:"$schema"`
	Autoupdate bool         `json:"autoupdate"`
	Share      string       `json:"share"`
	Permission orderedRules `json:"permission"`
}

type openCodeProviders struct {
	EnabledProviders []string                    `json:"enabled_providers"`
	Provider         map[string]openCodeProvider `json:"provider"`
}

type openCodeProvider struct {
	NPM     string              `json:"npm"`
	Name    string              `json:"name"`
	Options openCodeOptions     `json:"options"`
	Models  map[string]struct{} `json:"models"`
}

type openCodeOptions struct {
	BaseURL string            `json:"baseURL"`
	APIKey  string            `json:"apiKey"`
	Headers map[string]string `json:"headers,omitempty"`
}

type rule struct {
	pattern string
	action  string
}

type orderedRules []rule

func (r orderedRules) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')
	for i, rule := range r {
		if i > 0 {
			b.WriteByte(',')
		}
		pattern, err := json.Marshal(rule.pattern)
		if err != nil {
			return nil, err
		}
		action, err := json.Marshal(rule.action)
		if err != nil {
			return nil, err
		}
		b.Write(pattern)
		b.WriteByte(':')
		b.Write(action)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

func (o *openCode) Configure() error {
	policy := openCodePolicy{
		Schema:     "https://opencode.ai/config.json",
		Share:      "disabled",
		Permission: orderedRules{},
	}
	for _, connector := range o.agent.Connectors {
		action := "ask"
		if connector.Approval == config.Allow {
			action = "allow"
		}
		policy.Permission = append(policy.Permission, rule{pattern: connector.Name + "_*", action: action})
	}

	if err := writeJSON(filepath.Join(o.paths.Policy, "opencode", "opencode.json"), policy); err != nil {
		return err
	}
	return writeInstructions(o.paths.Workdir, "AGENTS.md", o.agent.Instructions)
}

func (o *openCode) Command(routing Routing) (acp.Command, error) {
	reference := "{env:" + keyVariable + "}"
	providers := openCodeProviders{EnabledProviders: []string{}, Provider: map[string]openCodeProvider{}}
	for name, route := range routing.Providers {
		options := openCodeOptions{BaseURL: o.agent.ModelURL(name, route.API), APIKey: reference}
		if route.API == config.Anthropic {
			options.BaseURL += "/v1"
			options.APIKey = "gateway"
			options.Headers = map[string]string{"Authorization": "Bearer " + reference}
		}
		models := map[string]struct{}{}
		for _, model := range route.Models {
			models[model] = struct{}{}
		}
		providers.EnabledProviders = append(providers.EnabledProviders, name)
		providers.Provider[name] = openCodeProvider{NPM: openCodePackages[route.API], Name: name, Options: options, Models: models}
	}
	slices.Sort(providers.EnabledProviders)

	variables := []string{
		keyVariable + "=" + o.agent.Gateway.Key,
		"OPENCODE_DISABLE_AUTOUPDATE=true",
		"OPENCODE_DISABLE_MODELS_FETCH=true",
		"OPENCODE_DISABLE_DEFAULT_PLUGINS=true",
		"OPENCODE_DISABLE_LSP_DOWNLOAD=true",
		"OPENCODE_DISABLE_SHARE=true",
	}
	if content, err := json.Marshal(providers); err == nil {
		variables = append(variables, "OPENCODE_CONFIG_CONTENT="+string(content))
	}
	return acp.Command{
		Path: "opencode",
		Args: []string{"acp"},
		Dir:  o.paths.Workdir,
		Env:  environment(o.paths.Home, variables...),
	}, nil
}

func (o *openCode) MCPServers() []json.RawMessage {
	return httpMCPServers(o.agent)
}

func (o *openCode) SessionMeta(Model) map[string]any {
	return nil
}

func (o *openCode) ModelValue(model Model) (string, error) {
	if err := requireAPI(o.agent, model, config.Anthropic, config.OpenAI, config.OpenAICompatible); err != nil {
		return "", err
	}
	return model.Provider + "/" + model.Name, nil
}

func (o *openCode) Covers(routing Routing, model Model) bool {
	route, ok := routing.Providers[model.Provider]
	return ok && route.API == model.API && slices.Contains(route.Models, model.Name)
}

func (o *openCode) Route(routing Routing, model Model) Routing {
	return withProvider(routing, model)
}
