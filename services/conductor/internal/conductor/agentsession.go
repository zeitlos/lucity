package conductor

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/pkg/usererror"
	"github.com/zeitlos/lucity/services/conductor/internal/agentrunner"
)

type AgentSession struct {
	ID        AgentSessionID
	Title     string
	StartedBy *WorkspaceMember
	StartedAt time.Time
	Model     AgentModel
	Effort    string
	Status    string
}

type AgentSessionEvent struct {
	agentrunner.Event
	Author    *WorkspaceMember
	DecidedBy *WorkspaceMember
}

// AgentModelSpec picks a model of a workspace's model provider.
type AgentModelSpec struct {
	Provider ModelProviderID
	Name     string
}

const (
	defaultEffort        = "medium"
	memberRefreshAtLeast = 30 * time.Second
)

type members struct {
	client    *Client
	workspace string

	mu      sync.Mutex
	byID    map[string]WorkspaceMember
	fetched time.Time
}

func (c *Client) members(workspace string) *members {
	return &members{client: c, workspace: workspace}
}

func (m *members) lookup(ctx context.Context, id string) *WorkspaceMember {
	if id == "" {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if member, ok := m.byID[id]; ok {
		return &member
	}

	if !m.fetched.IsZero() && time.Since(m.fetched) < memberRefreshAtLeast {
		return nil
	}

	m.fetched = time.Now()

	details, err := m.client.directory.Workspace(ctx, m.workspace)

	if err != nil {
		return nil
	}

	m.byID = make(map[string]WorkspaceMember, len(details.Members))

	for _, member := range details.Members {
		m.byID[member.ID] = member
	}

	if member, ok := m.byID[id]; ok {
		return &member
	}

	return nil
}

func (c *Client) toAgentSession(ctx context.Context, agent AgentID, session agentrunner.Session, ws *values.Workspace, people *members) AgentSession {
	return AgentSession{
		ID:        AgentSessionID{Agent: agent, Session: session.ID},
		Title:     session.Title,
		StartedBy: people.lookup(ctx, session.StartedBy),
		StartedAt: session.StartedAt,
		Model:     sessionModel(agent.Workspace, session.Model, ws),
		Effort:    session.Effort,
		Status:    session.Status,
	}
}

func (c *Client) toAgentSessionEvent(ctx context.Context, event agentrunner.Event, people *members) AgentSessionEvent {
	converted := AgentSessionEvent{Event: event, Author: people.lookup(ctx, event.Author)}

	if event.Approval != nil {
		converted.DecidedBy = people.lookup(ctx, event.Approval.DecidedBy)
	}

	return converted
}

func sessionModel(workspace string, model agentrunner.Model, ws *values.Workspace) AgentModel {
	id := ModelProviderID{Workspace: workspace, Name: model.Provider}

	if provider, ok := ws.ModelProviders[model.Provider]; ok {
		return AgentModel{Provider: toModelProvider(id, provider), Name: model.Name}
	}

	return AgentModel{
		Provider: ModelProvider{ID: id, API: values.ModelAPI(model.API), Status: values.CredentialUnreachable},
		Name:     model.Name,
	}
}

func (c *Client) AgentSessions(ctx context.Context, agent AgentID) ([]AgentSession, error) {
	sessions, err := c.agentRunner.Sessions(ctx, agent)

	if err != nil {
		return nil, err
	}

	ws, err := c.workspaceRelease(ctx, agent.Workspace)

	if err != nil {
		return nil, err
	}

	people := c.members(agent.Workspace)
	result := make([]AgentSession, 0, len(sessions))

	for _, session := range sessions {
		result = append(result, c.toAgentSession(ctx, agent, session, ws, people))
	}

	slices.SortStableFunc(result, func(a, b AgentSession) int { return b.StartedAt.Compare(a.StartedAt) })

	return result, nil
}

func (c *Client) AgentSession(ctx context.Context, id AgentSessionID) (*AgentSession, error) {
	session, err := c.agentRunner.Session(ctx, id)

	if err != nil {
		return nil, err
	}

	return c.sessionResult(ctx, id.Agent, session)
}

func (c *Client) sessionResult(ctx context.Context, agent AgentID, session *agentrunner.Session) (*AgentSession, error) {
	ws, err := c.workspaceRelease(ctx, agent.Workspace)

	if err != nil {
		return nil, err
	}

	return new(c.toAgentSession(ctx, agent, *session, ws, c.members(agent.Workspace))), nil
}

func (c *Client) AgentSessionEvents(ctx context.Context, id AgentSessionID) ([]AgentSessionEvent, error) {
	session, err := c.agentRunner.Session(ctx, id)

	if err != nil {
		return nil, err
	}

	people := c.members(id.Agent.Workspace)
	events := make([]AgentSessionEvent, 0, len(session.Events))

	for _, event := range session.Events {
		events = append(events, c.toAgentSessionEvent(ctx, event, people))
	}

	return events, nil
}

func (c *Client) StartAgentSession(ctx context.Context, agent AgentID, message string, model *AgentModelSpec, effort string) (*AgentSession, error) {
	actor, err := actorID(ctx)

	if err != nil {
		return nil, err
	}

	current, ws, err := c.loadAgent(ctx, agent)

	if err != nil {
		return nil, err
	}

	var runnerModel agentrunner.Model

	if model != nil {
		runnerModel, err = sessionRunnerModel(agent, current.Harness, ws, *model)

		if err != nil {
			return nil, err
		}
	} else {
		runnerModel, err = c.lastModel(ctx, agent, current.Harness, ws)

		if err != nil {
			return nil, err
		}
	}

	if effort == "" {
		effort = defaultEffort

		if status, err := c.agentRunner.Status(ctx, agent); err == nil && status.LastEffort != "" {
			effort = status.LastEffort
		}
	}

	session, err := c.agentRunner.StartSession(ctx, agent, agentrunner.StartRequest{
		Message:   message,
		Model:     runnerModel,
		Effort:    effort,
		StartedBy: actor,
	})

	if err != nil {
		return nil, err
	}

	return c.sessionResult(ctx, agent, session)
}

func (c *Client) lastModel(ctx context.Context, agent AgentID, harness values.Harness, ws *values.Workspace) (agentrunner.Model, error) {
	if status, err := c.agentRunner.Status(ctx, agent); err == nil && status.LastModel != nil {
		last := AgentModelSpec{Provider: ModelProviderID{Workspace: agent.Workspace, Name: status.LastModel.Provider}, Name: status.LastModel.Name}

		if model, err := sessionRunnerModel(agent, harness, ws, last); err == nil {
			return model, nil
		}
	}

	for _, provider := range modelProviders(agent.Workspace, ws) {
		if provider.Status == values.CredentialValid && harnessSupports(harness, provider.API) && len(provider.Models) > 0 {
			return agentrunner.Model{Provider: provider.ID.Name, API: string(provider.API), Name: provider.Models[0]}, nil
		}
	}

	return agentrunner.Model{}, usererror.New("the workspace has no model provider this agent's harness can use")
}

func sessionRunnerModel(agent AgentID, harness values.Harness, ws *values.Workspace, spec AgentModelSpec) (agentrunner.Model, error) {
	provider, ok := ws.ModelProviders[spec.Provider.Name]

	if spec.Provider.Workspace != agent.Workspace || !ok {
		return agentrunner.Model{}, usererror.Errorf("model provider %q not found", spec.Provider.Name)
	}

	if !harnessSupports(harness, provider.API) {
		return agentrunner.Model{}, usererror.Errorf("the agent's harness can't use model provider %q", spec.Provider.Name)
	}

	if !slices.Contains(provider.Models, spec.Name) {
		return agentrunner.Model{}, usererror.Errorf("model provider %q has no model %q", spec.Provider.Name, spec.Name)
	}

	return agentrunner.Model{Provider: spec.Provider.Name, API: string(provider.API), Name: spec.Name}, nil
}

func (c *Client) SetAgentSessionModel(ctx context.Context, id AgentSessionID, model AgentModelSpec, effort string) (*AgentSession, error) {
	current, ws, err := c.loadAgent(ctx, id.Agent)

	if err != nil {
		return nil, err
	}

	runnerModel, err := sessionRunnerModel(id.Agent, current.Harness, ws, model)

	if err != nil {
		return nil, err
	}

	session, err := c.agentRunner.SetModel(ctx, id, runnerModel, effort)

	if err != nil {
		return nil, err
	}

	return c.sessionResult(ctx, id.Agent, session)
}

func (c *Client) SendAgentMessage(ctx context.Context, id AgentSessionID, message string) (*AgentSession, error) {
	actor, err := actorID(ctx)

	if err != nil {
		return nil, err
	}

	session, err := c.agentRunner.SendMessage(ctx, id, message, actor)

	if err != nil {
		return nil, err
	}

	return c.sessionResult(ctx, id.Agent, session)
}

func (c *Client) AnswerAgentApproval(ctx context.Context, id AgentSessionID, approval, decision string) (*AgentSession, error) {
	actor, err := actorID(ctx)

	if err != nil {
		return nil, err
	}

	if !agentrunner.IsValidApprovalID(approval) {
		return nil, usererror.Errorf("approval %q not found", approval)
	}

	session, err := c.agentRunner.AnswerApproval(ctx, id, approval, decision, actor)

	if err != nil {
		return nil, err
	}

	return c.sessionResult(ctx, id.Agent, session)
}

func (c *Client) StopAgentSession(ctx context.Context, id AgentSessionID) (*AgentSession, error) {
	session, err := c.agentRunner.Stop(ctx, id)

	if err != nil {
		return nil, err
	}

	return c.sessionResult(ctx, id.Agent, session)
}

func (c *Client) SubscribeAgentSessionEvents(ctx context.Context, id AgentSessionID, after string) (<-chan AgentSessionEvent, error) {
	if after != "" && !platform.IsValidSessionName(after) {
		return nil, usererror.Errorf("invalid event %q", after)
	}

	events, err := c.agentRunner.Events(ctx, id, after)

	if err != nil {
		return nil, err
	}

	people := c.members(id.Agent.Workspace)
	out := make(chan AgentSessionEvent, 64)

	go func() {
		defer close(out)

		for event := range events {
			select {
			case out <- c.toAgentSessionEvent(ctx, event, people):
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, nil
}

func actorID(ctx context.Context) (string, error) {
	claims, err := auth.FromContext(ctx)

	if err != nil {
		return "", err
	}

	return claims.Subject, nil
}
