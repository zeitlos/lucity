package values

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"gopkg.in/yaml.v3"

	"github.com/zeitlos/lucity/pkg/chat"
	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/pkg/usererror"
)

// Workspace is the in-memory model of a lucity-workspace release's values:
// configuration that belongs to a workspace rather than to one of its
// environments. Like Env, mutators shape it and ValidateWorkspace checks the
// result.
type Workspace struct {
	Workspace      string                   `yaml:"workspace"`
	Gateway        WorkspaceGateway         `yaml:"gateway"`
	Switchboard    WorkspaceSwitchboard     `yaml:"switchboard"`
	ModelProviders map[string]ModelProvider `yaml:"modelProviders"`
	Connectors     map[string]Connector     `yaml:"connectors"`
	ChatApps       map[string]ChatApp       `yaml:"chatApps"`
	Agents         map[string]AgentGrant    `yaml:"agents"`
}

type WorkspaceGateway struct {
	Name     string `yaml:"name"`
	Listener string `yaml:"listener"`
}

// WorkspaceSwitchboard is where switchboard reads chat apps and agent
// channels from: its own namespace, so it never sees the gateway's
// credentials.
type WorkspaceSwitchboard struct {
	Namespace string `yaml:"namespace"`
}

type ModelAPI string

const (
	ModelAPIAnthropic        ModelAPI = "anthropic"
	ModelAPIOpenAI           ModelAPI = "openai"
	ModelAPIOpenAICompatible ModelAPI = "openaiCompatible"
)

type CredentialStatus string

const (
	CredentialValid       CredentialStatus = "valid"
	CredentialRejected    CredentialStatus = "rejected"
	CredentialUnreachable CredentialStatus = "unreachable"
)

type ModelProvider struct {
	API        ModelAPI         `yaml:"api"`
	URL        string           `yaml:"url"`
	Credential string           `yaml:"credential"`
	Status     CredentialStatus `yaml:"status"`
	CheckedAt  string           `yaml:"checkedAt,omitempty"`
	Models     []string         `yaml:"models,omitempty"`
}

type Connector struct {
	URL        string           `yaml:"url"`
	Credential string           `yaml:"credential,omitempty"`
	Status     CredentialStatus `yaml:"status"`
	CheckedAt  string           `yaml:"checkedAt,omitempty"`
	Tools      []ConnectorTool  `yaml:"tools,omitempty"`
}

type ConnectorTool struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

// ChatApp is a bot in a chat app that agents talk through: a Telegram bot, or
// a bot account on a Mattermost server at URL. Account is the bot's handle.
type ChatApp struct {
	Kind       string           `yaml:"kind"`
	URL        string           `yaml:"url,omitempty"`
	Credential string           `yaml:"credential"`
	Account    string           `yaml:"account"`
	Status     CredentialStatus `yaml:"status"`
	CheckedAt  string           `yaml:"checkedAt,omitempty"`
}

// AgentChannel is a place in a chat app an agent talks in: a Telegram chat or
// a Mattermost channel, by the chat app's own ID. Name is how it read when it
// was added.
type AgentChannel struct {
	ChatApp string `yaml:"chatApp"`
	Channel string `yaml:"channel"`
	Name    string `yaml:"name"`
}

// AgentGrant is an agent's entry in its workspace release: the hash of its
// gateway key, the namespace the key is accepted from, the connectors it may
// call and the chat channels it talks in. Keyed by project/environment/agent.
type AgentGrant struct {
	KeyHash     string              `yaml:"keyHash"`
	Namespace   string              `yaml:"namespace"`
	Connectors  []string            `yaml:"connectors,omitempty"`
	HiddenTools map[string][]string `yaml:"hiddenTools,omitempty"`
	Channels    []AgentChannel      `yaml:"channels,omitempty"`
}

const (
	maxModels             = 1000
	maxModelNameLen       = 256
	maxTools              = 1000
	maxToolDescriptionLen = 1024
	maxCredentialLen      = 4096
	maxUpstreamURLLen     = 512
	maxChatNameLen        = 256
	maxChannels           = 64
)

var (
	keyHashRe  = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	toolNameRe = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,128}$`)
	channelRe  = regexp.MustCompile(`^-?[A-Za-z0-9_]{1,64}$`)
	urlPathRe  = regexp.MustCompile(`^(/[A-Za-z0-9._~!$&'()*+,;=:@%-]*)*$`)

	internalSuffixes = []string{".local", ".localhost", ".internal", ".svc", ".home.arpa", ".lan"}
)

func NewWorkspace(workspace string) *Workspace {
	return &Workspace{
		Workspace:      workspace,
		ModelProviders: map[string]ModelProvider{},
		Connectors:     map[string]Connector{},
		ChatApps:       map[string]ChatApp{},
		Agents:         map[string]AgentGrant{},
	}
}

func ParseWorkspace(data []byte) (*Workspace, error) {
	ws := NewWorkspace("")

	if err := yaml.Unmarshal(data, ws); err != nil {
		return nil, fmt.Errorf("parse workspace values: %w", err)
	}

	return ws, nil
}

func MarshalWorkspace(ws *Workspace) ([]byte, error) {
	var buf bytes.Buffer

	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)

	if err := enc.Encode(ws); err != nil {
		return nil, fmt.Errorf("marshal workspace values: %w", err)
	}

	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("close yaml encoder: %w", err)
	}

	return buf.Bytes(), nil
}

// AgentGrantKey is the key of an agent in its workspace release.
func AgentGrantKey(id platform.AgentID) string {
	return id.Project + "/" + id.Environment + "/" + id.Name
}

func SetModelProvider(ws *Workspace, name string, provider ModelProvider) error {
	if ws.ModelProviders == nil {
		ws.ModelProviders = map[string]ModelProvider{}
	}

	ws.ModelProviders[name] = provider

	return nil
}

func DeleteModelProvider(ws *Workspace, name string) error {
	if _, ok := ws.ModelProviders[name]; !ok {
		return usererror.Errorf("model provider %q not found", name)
	}

	delete(ws.ModelProviders, name)

	return nil
}

func SetConnector(ws *Workspace, name string, connector Connector) error {
	if ws.Connectors == nil {
		ws.Connectors = map[string]Connector{}
	}

	ws.Connectors[name] = connector

	return nil
}

// DeleteConnector removes a connector and every grant of it.
func DeleteConnector(ws *Workspace, name string) error {
	if _, ok := ws.Connectors[name]; !ok {
		return usererror.Errorf("connector %q not found", name)
	}

	delete(ws.Connectors, name)

	for key, grant := range ws.Agents {
		grant.Connectors = slices.DeleteFunc(slices.Clone(grant.Connectors), func(c string) bool { return c == name })
		delete(grant.HiddenTools, name)
		ws.Agents[key] = grant
	}

	return nil
}

func SetChatApp(ws *Workspace, name string, app ChatApp) error {
	if ws.ChatApps == nil {
		ws.ChatApps = map[string]ChatApp{}
	}

	ws.ChatApps[name] = app

	return nil
}

// DeleteChatApp removes a chat app and every agent channel in it.
func DeleteChatApp(ws *Workspace, name string) error {
	if _, ok := ws.ChatApps[name]; !ok {
		return usererror.Errorf("chat app %q not found", name)
	}

	delete(ws.ChatApps, name)

	for key, grant := range ws.Agents {
		grant.Channels = slices.DeleteFunc(slices.Clone(grant.Channels), func(c AgentChannel) bool { return c.ChatApp == name })
		ws.Agents[key] = grant
	}

	return nil
}

// IsValidChannel reports whether channel can be a chat app's ID of a chat or
// channel: a Telegram chat ID, or a Mattermost channel ID.
func IsValidChannel(channel string) bool {
	return channelRe.MatchString(channel)
}

// SetAgentGrant sets an agent's key and connectors. Its channels stay as they
// are: only SetAgentChannels changes them, so neither overwrites the other.
func SetAgentGrant(ws *Workspace, id platform.AgentID, grant AgentGrant) error {
	if ws.Agents == nil {
		ws.Agents = map[string]AgentGrant{}
	}

	grant.Channels = ws.Agents[AgentGrantKey(id)].Channels
	ws.Agents[AgentGrantKey(id)] = grant

	return nil
}

// SetAgentChannels replaces the chat channels an agent talks in.
func SetAgentChannels(ws *Workspace, id platform.AgentID, channels []AgentChannel) error {
	grant, ok := ws.Agents[AgentGrantKey(id)]

	if !ok {
		return usererror.Errorf("agent %q not found", id.Name)
	}

	grant.Channels = channels
	ws.Agents[AgentGrantKey(id)] = grant

	return nil
}

// DeleteAgentGrant removes an agent's entry. An agent without one is already
// gone, so that is not an error.
func DeleteAgentGrant(ws *Workspace, id platform.AgentID) error {
	delete(ws.Agents, AgentGrantKey(id))

	return nil
}

// ValidateWorkspace checks every invariant of a workspace release.
func ValidateWorkspace(ws *Workspace, workspace string) error {
	if ws.Workspace != workspace || !isValidName(workspace) {
		return fmt.Errorf("workspace release for %q names workspace %q", workspace, ws.Workspace)
	}

	if !isValidName(ws.Gateway.Name) || !isValidName(ws.Gateway.Listener) {
		return fmt.Errorf("invalid agent gateway %q, listener %q", ws.Gateway.Name, ws.Gateway.Listener)
	}

	for name, provider := range ws.ModelProviders {
		if !isValidName(name) {
			return usererror.Errorf("invalid model provider name %q", name)
		}

		switch provider.API {
		case ModelAPIAnthropic, ModelAPIOpenAI, ModelAPIOpenAICompatible:
		default:
			return usererror.Errorf("model provider %q: unsupported api %q", name, provider.API)
		}

		if err := ValidateUpstreamURL(provider.URL); err != nil {
			return usererror.Errorf("model provider %q: %v", name, err)
		}

		if provider.Credential == "" {
			return usererror.Errorf("model provider %q needs a key", name)
		}

		if err := validateCredential(provider.Credential); err != nil {
			return usererror.Errorf("model provider %q: %v", name, err)
		}

		if err := validateCheck(provider.Status, provider.CheckedAt); err != nil {
			return fmt.Errorf("model provider %q: %w", name, err)
		}

		if len(provider.Models) > maxModels {
			return fmt.Errorf("model provider %q lists more than %d models", name, maxModels)
		}

		for _, model := range provider.Models {
			if !isPrintable(model, maxModelNameLen) {
				return fmt.Errorf("model provider %q: invalid model name %q", name, model)
			}
		}
	}

	for name, connector := range ws.Connectors {
		if !isValidName(name) {
			return usererror.Errorf("invalid connector name %q", name)
		}

		if err := ValidateUpstreamURL(connector.URL); err != nil {
			return usererror.Errorf("connector %q: %v", name, err)
		}

		if connector.Credential != "" {
			if err := validateCredential(connector.Credential); err != nil {
				return usererror.Errorf("connector %q: %v", name, err)
			}
		}

		if err := validateCheck(connector.Status, connector.CheckedAt); err != nil {
			return fmt.Errorf("connector %q: %w", name, err)
		}

		if len(connector.Tools) > maxTools {
			return fmt.Errorf("connector %q lists more than %d tools", name, maxTools)
		}

		for _, tool := range connector.Tools {
			if !IsValidToolName(tool.Name) {
				return fmt.Errorf("connector %q: invalid tool name %q", name, tool.Name)
			}

			if utf8.RuneCountInString(tool.Description) > maxToolDescriptionLen || !utf8.ValidString(tool.Description) {
				return fmt.Errorf("connector %q: invalid description for tool %q", name, tool.Name)
			}
		}
	}

	if len(ws.ChatApps) > 0 && !isValidName(ws.Switchboard.Namespace) {
		return fmt.Errorf("chat apps need the switchboard namespace, got %q", ws.Switchboard.Namespace)
	}

	for name, app := range ws.ChatApps {
		if !isValidName(name) {
			return usererror.Errorf("invalid chat app name %q", name)
		}

		switch app.Kind {
		case chat.Telegram:
			if app.URL != "" {
				return usererror.Errorf("chat app %q: a Telegram bot has no server URL", name)
			}
		case chat.Mattermost:
			if err := ValidateUpstreamURL(app.URL); err != nil {
				return usererror.Errorf("chat app %q: %v", name, err)
			}
		default:
			return usererror.Errorf("chat app %q: unsupported kind %q", name, app.Kind)
		}

		if app.Credential == "" {
			return usererror.Errorf("chat app %q needs a token", name)
		}

		if err := validateCredential(app.Credential); err != nil {
			return usererror.Errorf("chat app %q: %v", name, err)
		}

		if !isPrintable(app.Account, maxChatNameLen) {
			return fmt.Errorf("chat app %q: invalid account %q", name, app.Account)
		}

		if err := validateCheck(app.Status, app.CheckedAt); err != nil {
			return fmt.Errorf("chat app %q: %w", name, err)
		}
	}

	channelOwners := map[AgentChannel]string{}

	for key, grant := range ws.Agents {
		parts := strings.Split(key, "/")

		if len(parts) != 3 || !isValidName(parts[0]) || !isValidName(parts[1]) || !isValidName(parts[2]) {
			return fmt.Errorf("invalid agent %q", key)
		}

		environment := platform.EnvironmentID{Workspace: workspace, Project: parts[0], Name: parts[1]}

		if grant.Namespace != environment.Namespace() {
			return fmt.Errorf("agent %q: namespace %q is not its environment's", key, grant.Namespace)
		}

		if !keyHashRe.MatchString(grant.KeyHash) {
			return fmt.Errorf("agent %q: invalid key hash", key)
		}

		seen := map[string]bool{}

		for _, connector := range grant.Connectors {
			if _, ok := ws.Connectors[connector]; !ok {
				return usererror.Errorf("agent %q: unknown connector %q", parts[2], connector)
			}

			if seen[connector] {
				return usererror.Errorf("agent %q: connector %q is granted twice", parts[2], connector)
			}

			seen[connector] = true
		}

		for connector, tools := range grant.HiddenTools {
			if !seen[connector] {
				return usererror.Errorf("agent %q hides tools of connector %q, which it isn't granted", parts[2], connector)
			}

			for _, tool := range tools {
				if !IsValidToolName(tool) {
					return usererror.Errorf("agent %q: invalid tool name %q", parts[2], tool)
				}
			}
		}

		if len(grant.Channels) > maxChannels {
			return usererror.Errorf("agent %q is in more than %d channels", parts[2], maxChannels)
		}

		for _, channel := range grant.Channels {
			if _, ok := ws.ChatApps[channel.ChatApp]; !ok {
				return usererror.Errorf("agent %q: unknown chat app %q", parts[2], channel.ChatApp)
			}

			if !IsValidChannel(channel.Channel) {
				return usererror.Errorf("agent %q: invalid channel %q", parts[2], channel.Channel)
			}

			if !isPrintable(channel.Name, maxChatNameLen) {
				return usererror.Errorf("agent %q: invalid channel name %q", parts[2], channel.Name)
			}

			place := AgentChannel{ChatApp: channel.ChatApp, Channel: channel.Channel}

			if owner, ok := channelOwners[place]; ok {
				return usererror.Errorf("%q in chat app %q is already a channel of agent %q", channel.Name, channel.ChatApp, owner)
			}

			channelOwners[place] = parts[2]
		}
	}

	return nil
}

// ValidateUpstreamURL accepts an https URL of a public hostname: no IP
// address, no internal name, no credentials, query or fragment.
func ValidateUpstreamURL(raw string) error {
	if len(raw) > maxUpstreamURLLen {
		return fmt.Errorf("url exceeds %d characters", maxUpstreamURLLen)
	}

	if strings.ContainsFunc(raw, func(r rune) bool { return r <= 0x20 || r >= 0x7f }) {
		return fmt.Errorf("url contains an invalid character")
	}

	parsed, err := url.Parse(raw)

	if err != nil || parsed.Scheme != "https" || parsed.Opaque != "" {
		return fmt.Errorf("url must be an https URL")
	}

	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.ForceQuery {
		return fmt.Errorf("url can't carry credentials, a query or a fragment")
	}

	hostname := parsed.Hostname()

	if net.ParseIP(hostname) != nil || !isValidHostname(hostname) {
		return fmt.Errorf("url needs a public hostname")
	}

	for _, suffix := range internalSuffixes {
		if strings.HasSuffix(hostname, suffix) {
			return fmt.Errorf("url needs a public hostname")
		}
	}

	if port := parsed.Port(); port != "" {
		number, err := strconv.Atoi(port)

		if err != nil || number < 1 || number > 65535 {
			return fmt.Errorf("invalid port %q", port)
		}
	} else if strings.HasSuffix(parsed.Host, ":") {
		return fmt.Errorf("invalid port")
	}

	if !urlPathRe.MatchString(parsed.EscapedPath()) {
		return fmt.Errorf("invalid path")
	}

	return nil
}

func IsValidToolName(name string) bool {
	return toolNameRe.MatchString(name)
}

func validateCredential(credential string) error {
	if len(credential) > maxCredentialLen {
		return fmt.Errorf("credential exceeds %d characters", maxCredentialLen)
	}

	for _, r := range credential {
		if r < 0x20 || r > 0x7e {
			return fmt.Errorf("credential contains an invalid character")
		}
	}

	return nil
}

func validateCheck(status CredentialStatus, checkedAt string) error {
	switch status {
	case CredentialValid, CredentialRejected, CredentialUnreachable:
	default:
		return fmt.Errorf("invalid status %q", status)
	}

	if checkedAt == "" {
		return nil
	}

	if _, err := time.Parse(time.RFC3339, checkedAt); err != nil {
		return fmt.Errorf("invalid check time %q", checkedAt)
	}

	return nil
}

func isPrintable(s string, maxLen int) bool {
	if s == "" || len(s) > maxLen || !utf8.ValidString(s) {
		return false
	}

	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}

	return true
}
