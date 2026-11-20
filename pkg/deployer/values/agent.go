package values

import (
	"fmt"
	"maps"
	"net"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/zeitlos/lucity/pkg/usererror"
)

// AgentRuntime is what every agent of an environment shares: the runner
// image and the gateway its calls go through.
type AgentRuntime struct {
	Image   string       `yaml:"image,omitempty"`
	Gateway AgentGateway `yaml:"gateway"`
}

type AgentGateway struct {
	// URL is the gateway's base URL for the environment's workspace.
	URL       string `yaml:"url,omitempty"`
	Name      string `yaml:"name,omitempty"`
	Namespace string `yaml:"namespace,omitempty"`
}

type Agent struct {
	Harness      Harness                   `yaml:"harness"`
	Instructions string                    `yaml:"instructions"`
	Key          string                    `yaml:"key"`
	Connectors   map[string]AgentConnector `yaml:"connectors,omitempty"`
	Network      AgentNetwork              `yaml:"network"`
	Storage      AgentStorage              `yaml:"storage"`
	Resources    Resources                 `yaml:"resources,omitempty"`
}

type Harness string

const (
	HarnessClaudeCode Harness = "claudeCode"
	HarnessCodex      Harness = "codex"
	HarnessOpenCode   Harness = "opencode"
)

type Approval string

const (
	ApprovalAllow Approval = "allow"
	ApprovalAsk   Approval = "ask"
)

type AgentConnector struct {
	Approval Approval `yaml:"approval"`
}

type AgentNetwork struct {
	AllowedHosts   []string `yaml:"allowedHosts,omitempty"`
	PrivateNetwork bool     `yaml:"privateNetwork"`
}

type AgentStorage struct {
	Size string `yaml:"size"`
}

const (
	maxInstructionsLen = 64 * 1024
	maxAllowedHosts    = 64
	maxAgentConnectors = 64
)

var (
	agentKeyRe     = regexp.MustCompile(`^[A-Za-z0-9_-]{32,128}$`)
	minAgentVolume = resource.MustParse("1Gi")
	maxAgentVolume = resource.MustParse("100Gi")
)

// AgentServiceName is the name of the service whose resources would collide
// with an agent's: both render as lucity-app-agent-<name>.
func AgentServiceName(agent string) string {
	return "agent-" + agent
}

func CreateAgent(env *Env, name string, agent Agent) error {
	if _, ok := env.Agents[name]; ok {
		return usererror.Errorf("agent %q already exists", name)
	}

	if env.Agents == nil {
		env.Agents = map[string]Agent{}
	}

	env.Agents[name] = agent

	return nil
}

func UpdateAgent(env *Env, name string, harness Harness, instructions string) error {
	return mutateAgent(env, name, func(a *Agent) {
		a.Harness = harness
		a.Instructions = instructions
	})
}

func SetAgentConnectors(env *Env, name string, connectors map[string]AgentConnector) error {
	return mutateAgent(env, name, func(a *Agent) {
		a.Connectors = maps.Clone(connectors)
	})
}

func SetAgentNetwork(env *Env, name string, network AgentNetwork) error {
	return mutateAgent(env, name, func(a *Agent) {
		a.Network = AgentNetwork{AllowedHosts: slices.Clone(network.AllowedHosts), PrivateNetwork: network.PrivateNetwork}
	})
}

func DeleteAgent(env *Env, name string) error {
	if _, ok := env.Agents[name]; !ok {
		return usererror.Errorf("agent %q not found", name)
	}

	delete(env.Agents, name)

	return nil
}

func mutateAgent(env *Env, name string, mutate func(*Agent)) error {
	agent, ok := env.Agents[name]

	if !ok {
		return usererror.Errorf("agent %q not found", name)
	}

	mutate(&agent)
	env.Agents[name] = agent

	return nil
}

func StripAgents(env *Env) {
	env.Agents = nil
	env.AgentRuntime = AgentRuntime{}
}

func validateAgents(env *Env) error {
	if len(env.Agents) == 0 {
		return nil
	}

	runtime := env.AgentRuntime

	if runtime.Image == "" || len(runtime.Image) > maxImageLen || !imageRepoRe.MatchString(runtime.Image) {
		return fmt.Errorf("invalid agent image %q", runtime.Image)
	}

	if !strings.HasPrefix(runtime.Gateway.URL, "http://") && !strings.HasPrefix(runtime.Gateway.URL, "https://") {
		return fmt.Errorf("invalid agent gateway url %q", runtime.Gateway.URL)
	}

	if !isValidName(runtime.Gateway.Name) || !isValidName(runtime.Gateway.Namespace) {
		return fmt.Errorf("invalid agent gateway %s/%s", runtime.Gateway.Namespace, runtime.Gateway.Name)
	}

	for name, agent := range env.Agents {
		if !isValidName(name) || len(AgentServiceName(name)) > maxNameLen {
			return usererror.Errorf("invalid agent name %q", name)
		}

		if _, ok := env.Services[AgentServiceName(name)]; ok {
			return usererror.Errorf("agent %q can't share an environment with a service named %q", name, AgentServiceName(name))
		}

		switch agent.Harness {
		case HarnessClaudeCode, HarnessCodex, HarnessOpenCode:
		default:
			return usererror.Errorf("agent %q: unsupported harness %q", name, agent.Harness)
		}

		if len(agent.Instructions) > maxInstructionsLen || !utf8.ValidString(agent.Instructions) {
			return usererror.Errorf("agent %q: instructions must be valid text of at most %d bytes", name, maxInstructionsLen)
		}

		if !agentKeyRe.MatchString(agent.Key) {
			return fmt.Errorf("agent %q: invalid gateway key", name)
		}

		if len(agent.Connectors) > maxAgentConnectors {
			return usererror.Errorf("agent %q: at most %d connectors", name, maxAgentConnectors)
		}

		for connector, grant := range agent.Connectors {
			if !isValidName(connector) {
				return usererror.Errorf("agent %q: invalid connector name %q", name, connector)
			}

			if grant.Approval != ApprovalAllow && grant.Approval != ApprovalAsk {
				return usererror.Errorf("agent %q: invalid approval %q for connector %q", name, grant.Approval, connector)
			}
		}

		if err := validateAllowedHosts(agent.Network.AllowedHosts); err != nil {
			return usererror.Errorf("agent %q: %v", name, err)
		}

		size, err := resource.ParseQuantity(agent.Storage.Size)

		if err != nil || size.Cmp(minAgentVolume) < 0 || size.Cmp(maxAgentVolume) > 0 {
			return usererror.Errorf("agent %q: storage must be between %s and %s", name, minAgentVolume.String(), maxAgentVolume.String())
		}
	}

	return nil
}

func validateAllowedHosts(hosts []string) error {
	if len(hosts) > maxAllowedHosts {
		return fmt.Errorf("at most %d allowed hosts", maxAllowedHosts)
	}

	seen := make(map[string]bool, len(hosts))

	for _, host := range hosts {
		domain := strings.TrimPrefix(host, "*.")

		if host == "*" || net.ParseIP(strings.Trim(domain, "[]")) != nil || !isValidHostname(domain) {
			return fmt.Errorf("invalid allowed host %q: use a hostname, or *.example.com for its subdomains", host)
		}

		if seen[host] {
			return fmt.Errorf("allowed host %q is listed twice", host)
		}

		seen[host] = true
	}

	return nil
}
