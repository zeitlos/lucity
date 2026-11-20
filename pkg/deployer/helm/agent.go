package helm

import (
	"context"
	"maps"

	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/platform"
)

type agentClient struct {
	client *Client
}

func (a *agentClient) Agents(ctx context.Context, env platform.EnvironmentID) (map[string]values.Agent, error) {
	current, err := a.client.loadEnv(ctx, env)

	if err != nil {
		return nil, err
	}

	return maps.Clone(current.Agents), nil
}

func (a *agentClient) Create(ctx context.Context, env platform.EnvironmentID, name string, agent values.Agent) (deployer.RevisionID, error) {
	if a.client.agents.Image == "" {
		return "", errAgentsDisabled
	}

	return a.client.applyEnv(ctx, env, func(e *values.Env) error {
		return values.CreateAgent(e, name, agent)
	})
}

func (a *agentClient) Update(ctx context.Context, id platform.AgentID, harness values.Harness, instructions string) (deployer.RevisionID, error) {
	return a.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.UpdateAgent(e, id.Name, harness, instructions)
	})
}

func (a *agentClient) SetConnectors(ctx context.Context, id platform.AgentID, connectors map[string]values.AgentConnector) (deployer.RevisionID, error) {
	return a.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.SetAgentConnectors(e, id.Name, connectors)
	})
}

func (a *agentClient) SetNetwork(ctx context.Context, id platform.AgentID, network values.AgentNetwork) (deployer.RevisionID, error) {
	return a.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.SetAgentNetwork(e, id.Name, network)
	})
}

func (a *agentClient) Delete(ctx context.Context, id platform.AgentID) error {
	_, err := a.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.DeleteAgent(e, id.Name)
	})

	return err
}

var _ deployer.AgentClient = (*agentClient)(nil)
