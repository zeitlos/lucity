package conductor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"slices"
	"strings"

	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/usererror"
	"github.com/zeitlos/lucity/services/conductor/internal/agentrunner"
)

type Agent struct {
	ID             AgentID
	Harness        values.Harness
	Instructions   string
	Connectors     []AgentConnector
	AllowedHosts   []string
	PrivateNetwork bool
}

type AgentConnector struct {
	Connector   Connector
	Approval    values.Approval
	HiddenTools []string
}

// AgentConnectorSpec grants an agent a connector.
type AgentConnectorSpec struct {
	Connector   ConnectorID
	Approval    values.Approval
	HiddenTools []string
}

type AgentStatus string

const (
	AgentStarting           AgentStatus = "starting"
	AgentIdle               AgentStatus = "idle"
	AgentWorking            AgentStatus = "working"
	AgentWaitingForApproval AgentStatus = "waitingForApproval"
	AgentFailed             AgentStatus = "failed"
)

type AgentModel struct {
	Provider ModelProvider
	Name     string
}

const agentStorage = "10Gi"

var agentResources = values.Resources{
	Requests: values.ResourceList{CPU: "250m", Memory: "1Gi"},
	Limits:   values.ResourceList{Memory: "2Gi"},
}

func (c *Client) Agents(ctx context.Context, environment EnvironmentID) ([]Agent, error) {
	if !c.config.AgentsEnabled {
		return []Agent{}, nil
	}

	agents, err := c.deployer.Agents().Agents(ctx, environment)

	if err != nil {
		return nil, err
	}

	if len(agents) == 0 {
		return []Agent{}, nil
	}

	ws, err := c.workspaceRelease(ctx, environment.Workspace)

	if err != nil {
		return nil, err
	}

	return toAgents(environment, agents, ws), nil
}

func (c *Client) Agent(ctx context.Context, id AgentID) (*Agent, error) {
	agent, ws, err := c.loadAgent(ctx, id)

	if err != nil {
		return nil, err
	}

	return new(toAgent(id, agent, ws)), nil
}

func (c *Client) loadAgent(ctx context.Context, id AgentID) (values.Agent, *values.Workspace, error) {
	if !c.config.AgentsEnabled {
		return values.Agent{}, nil, usererror.Errorf("agent %q not found", id.Name)
	}

	agents, err := c.deployer.Agents().Agents(ctx, id.EnvironmentID())

	if err != nil {
		return values.Agent{}, nil, err
	}

	agent, ok := agents[id.Name]

	if !ok {
		return values.Agent{}, nil, usererror.Errorf("agent %q not found", id.Name)
	}

	ws, err := c.workspaceRelease(ctx, id.Workspace)

	if err != nil {
		return values.Agent{}, nil, err
	}

	return agent, ws, nil
}

func (c *Client) workspaceAgents(ctx context.Context, workspace string, ws *values.Workspace) ([]Agent, error) {
	byEnvironment := map[EnvironmentID]bool{}

	for key := range ws.Agents {
		if id, ok := agentIDFromGrantKey(workspace, key); ok {
			byEnvironment[id.EnvironmentID()] = true
		}
	}

	var result []Agent

	for environment := range byEnvironment {
		agents, err := c.deployer.Agents().Agents(ctx, environment)

		if err != nil {
			return nil, err
		}

		result = append(result, toAgents(environment, agents, ws)...)
	}

	slices.SortFunc(result, func(a, b Agent) int { return strings.Compare(a.ID.String(), b.ID.String()) })

	return result, nil
}

func agentIDFromGrantKey(workspace, key string) (AgentID, bool) {
	parts := strings.Split(key, "/")

	if len(parts) != 3 {
		return AgentID{}, false
	}

	return AgentID{Workspace: workspace, Project: parts[0], Environment: parts[1], Name: parts[2]}, true
}

func toAgents(environment EnvironmentID, agents map[string]values.Agent, ws *values.Workspace) []Agent {
	result := make([]Agent, 0, len(agents))

	for name, agent := range agents {
		id := AgentID{Workspace: environment.Workspace, Project: environment.Project, Environment: environment.Name, Name: name}
		result = append(result, toAgent(id, agent, ws))
	}

	slices.SortFunc(result, func(a, b Agent) int { return strings.Compare(a.ID.Name, b.ID.Name) })

	return result
}

func toAgent(id AgentID, agent values.Agent, ws *values.Workspace) Agent {
	grant := ws.Agents[values.AgentGrantKey(id)]

	connectors := make([]AgentConnector, 0, len(agent.Connectors))

	for name, setting := range agent.Connectors {
		connector, ok := ws.Connectors[name]

		if !ok || !slices.Contains(grant.Connectors, name) {
			continue
		}

		connectors = append(connectors, AgentConnector{
			Connector:   toConnector(ConnectorID{Workspace: id.Workspace, Name: name}, connector),
			Approval:    setting.Approval,
			HiddenTools: slices.Clone(grant.HiddenTools[name]),
		})
	}

	slices.SortFunc(connectors, func(a, b AgentConnector) int { return strings.Compare(a.Connector.ID.Name, b.Connector.ID.Name) })

	return Agent{
		ID:             id,
		Harness:        agent.Harness,
		Instructions:   agent.Instructions,
		Connectors:     connectors,
		AllowedHosts:   slices.Clone(agent.Network.AllowedHosts),
		PrivateNetwork: agent.Network.PrivateNetwork,
	}
}

func (c *Client) CreateAgent(ctx context.Context, environment EnvironmentID, name string, harness values.Harness, instructions string, connectors []AgentConnectorSpec) (*Agent, error) {
	if !c.config.AgentsEnabled {
		return nil, usererror.New("agents are not enabled on this platform")
	}

	id := AgentID{Workspace: environment.Workspace, Project: environment.Project, Environment: environment.Name, Name: name}

	existing, err := c.deployer.Agents().Agents(ctx, environment)

	if err != nil {
		return nil, err
	}

	if _, ok := existing[name]; ok {
		return nil, usererror.Errorf("an agent named %q already exists", name)
	}

	ws, err := c.workspaceRelease(ctx, id.Workspace)

	if err != nil {
		return nil, err
	}

	settings, err := agentConnectorSettings(id, ws, connectors)

	if err != nil {
		return nil, err
	}

	key := "lak_" + randCrockford32(40)

	if err := c.deployer.Workspaces().SetAgentGrant(ctx, id, agentGrant(id, key, connectors)); err != nil {
		return nil, err
	}

	agent := values.Agent{
		Harness:      harness,
		Instructions: instructions,
		Key:          key,
		Connectors:   settings,
		Storage:      values.AgentStorage{Size: agentStorage},
		Resources:    agentResources,
	}

	if _, err := c.deployer.Agents().Create(ctx, environment, name, agent); err != nil {
		if cleanupErr := c.deployer.Workspaces().DeleteAgentGrant(ctx, id); cleanupErr != nil {
			slog.WarnContext(ctx, "agent: failed to remove the grant of an agent that wasn't created", "agent", id, "error", cleanupErr)
		}

		return nil, err
	}

	ws, err = c.workspaceRelease(ctx, id.Workspace)

	if err != nil {
		return nil, err
	}

	return new(toAgent(id, agent, ws)), nil
}

func (c *Client) UpdateAgent(ctx context.Context, id AgentID, harness values.Harness, instructions string) (*Agent, error) {
	if _, err := c.deployer.Agents().Update(ctx, id, harness, instructions); err != nil {
		return nil, err
	}

	return c.Agent(ctx, id)
}

func (c *Client) SetAgentConnectors(ctx context.Context, id AgentID, connectors []AgentConnectorSpec) (*Agent, error) {
	agent, ws, err := c.loadAgent(ctx, id)

	if err != nil {
		return nil, err
	}

	settings, err := agentConnectorSettings(id, ws, connectors)

	if err != nil {
		return nil, err
	}

	if err := c.deployer.Workspaces().SetAgentGrant(ctx, id, agentGrant(id, agent.Key, connectors)); err != nil {
		return nil, err
	}

	if _, err := c.deployer.Agents().SetConnectors(ctx, id, settings); err != nil {
		return nil, err
	}

	return c.Agent(ctx, id)
}

func (c *Client) SetAgentNetwork(ctx context.Context, id AgentID, allowedHosts []string, privateNetwork bool) (*Agent, error) {
	hosts := make([]string, 0, len(allowedHosts))

	for _, host := range allowedHosts {
		hosts = append(hosts, strings.ToLower(strings.TrimSpace(host)))
	}

	if _, err := c.deployer.Agents().SetNetwork(ctx, id, values.AgentNetwork{AllowedHosts: hosts, PrivateNetwork: privateNetwork}); err != nil {
		return nil, err
	}

	return c.Agent(ctx, id)
}

// DeleteAgent removes the agent from its environment first, so its pod is
// gone before its key stops working, and treats either part already being
// gone as done.
func (c *Client) DeleteAgent(ctx context.Context, id AgentID) (bool, error) {
	if !c.config.AgentsEnabled {
		return false, usererror.Errorf("agent %q not found", id.Name)
	}

	agents, err := c.deployer.Agents().Agents(ctx, id.EnvironmentID())

	if err != nil {
		return false, err
	}

	if _, ok := agents[id.Name]; ok {
		if err := c.deployer.Agents().Delete(ctx, id); err != nil {
			return false, err
		}
	}

	if err := c.deployer.Workspaces().DeleteAgentGrant(ctx, id); err != nil {
		return false, err
	}

	return true, nil
}

func agentConnectorSettings(id AgentID, ws *values.Workspace, connectors []AgentConnectorSpec) (map[string]values.AgentConnector, error) {
	settings := make(map[string]values.AgentConnector, len(connectors))

	for _, spec := range connectors {
		if spec.Connector.Workspace != id.Workspace {
			return nil, usererror.Errorf("connector %q not found", spec.Connector.Name)
		}

		if _, ok := ws.Connectors[spec.Connector.Name]; !ok {
			return nil, usererror.Errorf("connector %q not found", spec.Connector.Name)
		}

		if _, ok := settings[spec.Connector.Name]; ok {
			return nil, usererror.Errorf("connector %q is listed twice", spec.Connector.Name)
		}

		settings[spec.Connector.Name] = values.AgentConnector{Approval: spec.Approval}
	}

	return settings, nil
}

func agentGrant(id AgentID, key string, connectors []AgentConnectorSpec) values.AgentGrant {
	sum := sha256.Sum256([]byte(key))

	grant := values.AgentGrant{
		KeyHash:   "sha256:" + hex.EncodeToString(sum[:]),
		Namespace: id.Namespace(),
	}

	for _, spec := range connectors {
		grant.Connectors = append(grant.Connectors, spec.Connector.Name)

		hidden := slices.Clone(spec.HiddenTools)
		slices.Sort(hidden)
		hidden = slices.Compact(hidden)

		if len(hidden) > 0 {
			if grant.HiddenTools == nil {
				grant.HiddenTools = map[string][]string{}
			}

			grant.HiddenTools[spec.Connector.Name] = hidden
		}
	}

	slices.Sort(grant.Connectors)

	return grant
}

// AgentModels returns the workspace's models that an agent's harness can
// use, from providers whose last check passed.
func (c *Client) AgentModels(ctx context.Context, id AgentID) ([]AgentModel, error) {
	agent, ws, err := c.loadAgent(ctx, id)

	if err != nil {
		return nil, err
	}

	var models []AgentModel

	for _, provider := range modelProviders(id.Workspace, ws) {
		if provider.Status != values.CredentialValid || !harnessSupports(agent.Harness, provider.API) {
			continue
		}

		for _, model := range provider.Models {
			models = append(models, AgentModel{Provider: provider, Name: model})
		}
	}

	return models, nil
}

func (c *Client) AgentStatus(ctx context.Context, id AgentID) (AgentStatus, error) {
	status, err := c.agentRunner.Status(ctx, id)

	if err == nil {
		if status.Status == agentrunner.StatusFailed {
			slog.WarnContext(ctx, "agent: harness doesn't start", "agent", id, "error", status.Error)
		}

		return AgentStatus(status.Status), nil
	}

	failed, failedErr := c.agentRunner.Failed(ctx, id)

	if failedErr != nil {
		return "", failedErr
	}

	if failed {
		return AgentFailed, nil
	}

	if errors.Is(err, agentrunner.ErrUnreachable) {
		return AgentStarting, nil
	}

	slog.WarnContext(ctx, "agent: runner status failed", "agent", id, "error", err)

	return AgentStarting, nil
}
