package conductor

import (
	"context"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/usererror"
	"github.com/zeitlos/lucity/services/conductor/internal/upstream"
)

type Connector struct {
	ID             ConnectorID
	URL            string
	CredentialHint string
	Status         values.CredentialStatus
	CheckedAt      *time.Time
	Tools          []values.ConnectorTool
}

func (c *Client) Connectors(ctx context.Context) ([]Connector, error) {
	workspace, ws, err := c.callerWorkspaceRelease(ctx)

	if err != nil {
		return nil, err
	}

	connectors := make([]Connector, 0, len(ws.Connectors))

	for name, connector := range ws.Connectors {
		connectors = append(connectors, toConnector(ConnectorID{Workspace: workspace, Name: name}, connector))
	}

	slices.SortFunc(connectors, func(a, b Connector) int { return strings.Compare(a.ID.Name, b.ID.Name) })

	return connectors, nil
}

func (c *Client) Connector(ctx context.Context, id ConnectorID) (*Connector, error) {
	ws, err := c.workspaceRelease(ctx, id.Workspace)

	if err != nil {
		return nil, err
	}

	connector, ok := ws.Connectors[id.Name]

	if !ok {
		return nil, usererror.Errorf("connector %q not found", id.Name)
	}

	return new(toConnector(id, connector)), nil
}

func toConnector(id ConnectorID, connector values.Connector) Connector {
	return Connector{
		ID:             id,
		URL:            connector.URL,
		CredentialHint: credentialHint(connector.Credential),
		Status:         connector.Status,
		CheckedAt:      parseCheckedAt(connector.CheckedAt),
		Tools:          connector.Tools,
	}
}

func (c *Client) CreateConnector(ctx context.Context, name, rawURL, credential string) (*Connector, error) {
	workspace, ws, err := c.callerWorkspaceRelease(ctx)

	if err != nil {
		return nil, err
	}

	id := ConnectorID{Workspace: workspace, Name: name}

	if _, ok := ws.Connectors[name]; ok {
		return nil, usererror.Errorf("a connector named %q already exists", name)
	}

	connectorURL, err := normalizeUpstreamURL(rawURL)

	if err != nil {
		return nil, err
	}

	connector, err := c.checkConnector(ctx, connectorURL, credential)

	if err != nil {
		return nil, err
	}

	if err := c.deployer.Workspaces().SetConnector(ctx, id, connector); err != nil {
		return nil, err
	}

	return new(toConnector(id, connector)), nil
}

func (c *Client) ReplaceConnectorCredential(ctx context.Context, id ConnectorID, credential string) (*Connector, error) {
	ws, err := c.workspaceRelease(ctx, id.Workspace)

	if err != nil {
		return nil, err
	}

	current, ok := ws.Connectors[id.Name]

	if !ok {
		return nil, usererror.Errorf("connector %q not found", id.Name)
	}

	connector, err := c.checkConnector(ctx, current.URL, credential)

	if err != nil {
		return nil, err
	}

	if err := c.deployer.Workspaces().SetConnector(ctx, id, connector); err != nil {
		return nil, err
	}

	return new(toConnector(id, connector)), nil
}

func (c *Client) RefreshConnector(ctx context.Context, id ConnectorID) (*Connector, error) {
	ws, err := c.workspaceRelease(ctx, id.Workspace)

	if err != nil {
		return nil, err
	}

	connector, ok := ws.Connectors[id.Name]

	if !ok {
		return nil, usererror.Errorf("connector %q not found", id.Name)
	}

	tools, check := c.upstream.Tools(ctx, connector.URL, connector.Credential)

	connector.Status = values.CredentialStatus(check.Status)
	connector.CheckedAt = time.Now().UTC().Format(time.RFC3339)

	if check.Status == upstream.Valid {
		connector.Tools = toConnectorTools(tools)
	}

	if err := c.deployer.Workspaces().SetConnector(ctx, id, connector); err != nil {
		return nil, err
	}

	return new(toConnector(id, connector)), nil
}

// DeleteConnector removes a connector and its grants from the workspace
// release, which cuts every agent off at once, and then drops it from the
// configuration of the agents that had it.
func (c *Client) DeleteConnector(ctx context.Context, id ConnectorID) (bool, error) {
	ws, err := c.workspaceRelease(ctx, id.Workspace)

	if err != nil {
		return false, err
	}

	var granted []AgentID

	for key, grant := range ws.Agents {
		if slices.Contains(grant.Connectors, id.Name) {
			if agentID, ok := agentIDFromGrantKey(id.Workspace, key); ok {
				granted = append(granted, agentID)
			}
		}
	}

	if err := c.deployer.Workspaces().DeleteConnector(ctx, id); err != nil {
		return false, err
	}

	for _, agentID := range granted {
		agents, err := c.deployer.Agents().Agents(ctx, agentID.EnvironmentID())

		if err != nil {
			slog.WarnContext(ctx, "connector: failed to read agent", "agent", agentID, "error", err)
			continue
		}

		agent, ok := agents[agentID.Name]

		if !ok {
			continue
		}

		if _, ok := agent.Connectors[id.Name]; !ok {
			continue
		}

		connectors := map[string]values.AgentConnector{}

		for name, grant := range agent.Connectors {
			if name != id.Name {
				connectors[name] = grant
			}
		}

		if _, err := c.deployer.Agents().SetConnectors(ctx, agentID, connectors); err != nil {
			slog.WarnContext(ctx, "connector: failed to remove connector from agent", "agent", agentID, "error", err)
		}
	}

	return true, nil
}

// ConnectorAgents returns the agents granted a connector.
func (c *Client) ConnectorAgents(ctx context.Context, id ConnectorID) ([]Agent, error) {
	ws, err := c.workspaceRelease(ctx, id.Workspace)

	if err != nil {
		return nil, err
	}

	if _, ok := ws.Connectors[id.Name]; !ok {
		return nil, usererror.Errorf("connector %q not found", id.Name)
	}

	agents, err := c.workspaceAgents(ctx, id.Workspace, ws)

	if err != nil {
		return nil, err
	}

	return slices.DeleteFunc(agents, func(agent Agent) bool {
		return !slices.ContainsFunc(agent.Connectors, func(granted AgentConnector) bool { return granted.Connector.ID == id })
	}), nil
}

func (c *Client) checkConnector(ctx context.Context, connectorURL, credential string) (values.Connector, error) {
	tools, check := c.upstream.Tools(ctx, connectorURL, credential)

	if check.Status != upstream.Valid {
		return values.Connector{}, usererror.Errorf("couldn't connect the connector: %s", check.Reason)
	}

	return values.Connector{
		URL:        connectorURL,
		Credential: credential,
		Status:     values.CredentialValid,
		CheckedAt:  time.Now().UTC().Format(time.RFC3339),
		Tools:      toConnectorTools(tools),
	}, nil
}

func toConnectorTools(tools []upstream.Tool) []values.ConnectorTool {
	result := make([]values.ConnectorTool, 0, len(tools))

	for _, tool := range tools {
		result = append(result, values.ConnectorTool{Name: tool.Name, Description: tool.Description})
	}

	return result
}
