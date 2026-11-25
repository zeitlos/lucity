package conductor

import (
	"context"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/zeitlos/lucity/pkg/agentrunner"
	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/chat"
	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/mattermost"
	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/pkg/usererror"
)

type AgentSession struct {
	ID        AgentSessionID
	Title     string
	StartedBy AgentActor
	StartedAt time.Time
	Model     AgentModel
	Effort    string
	Status    string
	Chat      *AgentChatLink
}

type AgentSessionEvent struct {
	agentrunner.Event
	Author    AgentActor
	DecidedBy *WorkspaceMember
	Via       *AgentChatLink
}

// AgentActor is who started a session or wrote a message: a member of the
// workspace, or one of the agent's triggers. Both are nil once the member left
// or the trigger was deleted.
type AgentActor struct {
	Member  *WorkspaceMember
	Trigger *AgentTrigger
}

// AgentChatLink is the chat a session is mirrored to, or a message came
// through, with a link to the thread there when the chat app has one.
type AgentChatLink struct {
	ChatApp ChatApp
	Channel ChatChannel
	URL     string
}

// AgentModelSpec picks a model of a workspace's model provider.
type AgentModelSpec struct {
	Provider ModelProviderID
	Name     string
}

const (
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

// actors resolves who started a session or wrote a message.
type actors struct {
	people   *members
	triggers []AgentTrigger
}

func (c *Client) actors(id AgentID, agent values.Agent) *actors {
	return &actors{people: c.members(id.Workspace), triggers: toAgentTriggers(id, agent)}
}

func (a *actors) lookup(ctx context.Context, actor string) AgentActor {
	if trigger, ok := agentrunner.ActorTrigger(actor); ok {
		index := slices.IndexFunc(a.triggers, func(t AgentTrigger) bool { return t.ID.Trigger == trigger })

		if index < 0 {
			return AgentActor{}
		}

		return AgentActor{Trigger: &a.triggers[index]}
	}

	return AgentActor{Member: a.people.lookup(ctx, actor)}
}

func (c *Client) toAgentSession(ctx context.Context, agent AgentID, session agentrunner.Session, ws *values.Workspace, people *actors) AgentSession {
	return AgentSession{
		ID:        AgentSessionID{Agent: agent, Session: session.ID},
		Title:     session.Title,
		StartedBy: people.lookup(ctx, session.StartedBy),
		StartedAt: session.StartedAt,
		Model:     sessionModel(agent.Workspace, session.Model, ws),
		Effort:    session.Effort,
		Status:    session.Status,
		Chat:      agentChatLink(agent, session.Chat, ws),
	}
}

func (c *Client) toAgentSessionEvent(ctx context.Context, agent AgentID, event agentrunner.Event, ws *values.Workspace, people *actors) AgentSessionEvent {
	converted := AgentSessionEvent{Event: event, Author: people.lookup(ctx, event.Author), Via: agentChatLink(agent, event.Via, ws)}

	if event.Approval != nil {
		converted.DecidedBy = people.people.lookup(ctx, event.Approval.DecidedBy)
	}

	return converted
}

// agentChatLink resolves a chat the runner reported. The runner is untrusted,
// so the chat app is looked up by name in the agent's own workspace, and a
// chat naming none of its chat apps is dropped.
func agentChatLink(agent AgentID, link *agentrunner.Chat, ws *values.Workspace) *AgentChatLink {
	if link == nil {
		return nil
	}

	app, ok := ws.ChatApps[link.ChatApp]

	if !ok {
		return nil
	}

	channel := ChatChannel{ID: link.Channel, Name: link.Channel}

	if known := ws.Agents[values.AgentGrantKey(agent)].Chat; known != nil && known.ChatApp == link.ChatApp && known.Channel == link.Channel {
		channel.Name = known.Name
	}

	return &AgentChatLink{
		ChatApp: toChatApp(ChatAppID{Workspace: agent.Workspace, Name: link.ChatApp}, app),
		Channel: channel,
		URL:     threadURL(app, link.Channel, link.Thread),
	}
}

// threadURL links to a thread in a chat app: a Mattermost post, or a message
// or topic in a Telegram supergroup. Other Telegram chats have no links.
func threadURL(app values.ChatApp, channel, thread string) string {
	if thread == "" {
		return ""
	}

	switch app.Kind {
	case chat.Mattermost:
		return mattermost.New(app.URL, "", nil).Permalink(thread)
	case chat.Telegram:
		if id, ok := strings.CutPrefix(channel, "-100"); ok {
			return "https://t.me/c/" + id + "/" + url.PathEscape(thread)
		}
	}

	return ""
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

	current, ws, err := c.loadAgent(ctx, agent)

	if err != nil {
		return nil, err
	}

	people := c.actors(agent, current)
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
	current, ws, err := c.loadAgent(ctx, agent)

	if err != nil {
		return nil, err
	}

	return new(c.toAgentSession(ctx, agent, *session, ws, c.actors(agent, current))), nil
}

func (c *Client) AgentSessionEvents(ctx context.Context, id AgentSessionID) ([]AgentSessionEvent, error) {
	session, err := c.agentRunner.Session(ctx, id)

	if err != nil {
		return nil, err
	}

	current, ws, err := c.loadAgent(ctx, id.Agent)

	if err != nil {
		return nil, err
	}

	people := c.actors(id.Agent, current)
	events := make([]AgentSessionEvent, 0, len(session.Events))

	for _, event := range session.Events {
		events = append(events, c.toAgentSessionEvent(ctx, id.Agent, event, ws, people))
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
		effort = agentrunner.DefaultEffort

		if status, err := c.agentRunner.Status(ctx, agent); err == nil && status.LastEffort != "" {
			effort = status.LastEffort
		}
	}

	session, err := c.agentRunner.StartSession(ctx, agent, agentrunner.StartRequest{
		Message:   message,
		Model:     &runnerModel,
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

	name, provider, ok := values.DefaultModel(ws, harness)

	if !ok {
		return agentrunner.Model{}, usererror.New("the workspace has no model provider this agent's harness can use")
	}

	return agentrunner.Model{Provider: name, API: string(provider.API), Name: provider.Models[0]}, nil
}

func sessionRunnerModel(agent AgentID, harness values.Harness, ws *values.Workspace, spec AgentModelSpec) (agentrunner.Model, error) {
	provider, ok := ws.ModelProviders[spec.Provider.Name]

	if spec.Provider.Workspace != agent.Workspace || !ok {
		return agentrunner.Model{}, usererror.Errorf("model provider %q not found", spec.Provider.Name)
	}

	if !values.HarnessSupports(harness, provider.API) {
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

	session, err := c.agentRunner.SendMessage(ctx, id, message, actor, nil)

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

	current, ws, err := c.loadAgent(ctx, id.Agent)

	if err != nil {
		return nil, err
	}

	events, err := c.agentRunner.Events(ctx, id, after)

	if err != nil {
		return nil, err
	}

	people := c.actors(id.Agent, current)
	out := make(chan AgentSessionEvent, 64)

	go func() {
		defer close(out)

		for event := range events {
			select {
			case out <- c.toAgentSessionEvent(ctx, id.Agent, event, ws, people):
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
