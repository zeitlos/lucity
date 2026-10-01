package bridge

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
)

var agentVariables = []string{
	"PATH",
	"HOME",
	"LANG",
	"TZ",
	"ANTHROPIC_API_KEY",
	"ANTHROPIC_BASE_URL",
	"KUBERNETES_SERVICE_HOST",
	"KUBERNETES_SERVICE_PORT",
}

func LoadMCPServers(path string) ([]json.RawMessage, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var servers []json.RawMessage
	if err := json.Unmarshal([]byte(os.ExpandEnv(string(raw))), &servers); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return servers, nil
}

func agentEnvironment() []string {
	var environment []string
	for _, variable := range os.Environ() {
		name, _, _ := strings.Cut(variable, "=")
		if slices.Contains(agentVariables, name) || strings.HasPrefix(name, "CLAUDE_CODE_") {
			environment = append(environment, variable)
		}
	}
	return environment
}
