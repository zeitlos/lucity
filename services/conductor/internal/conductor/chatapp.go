package conductor

import (
	"context"
	"crypto/rand"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/zeitlos/lucity/pkg/chat"
	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/usererror"
	"github.com/zeitlos/lucity/services/conductor/internal/upstream"
)

type ChatApp struct {
	ID             ChatAppID
	Kind           string
	URL            string
	Account        string
	CredentialHint string
	Status         values.CredentialStatus
	CheckedAt      *time.Time
}

// ChatChannel is a chat or channel in a chat app, by the chat app's own ID.
type ChatChannel struct {
	ID   string
	Name string
}

type AgentChannel struct {
	ChatApp ChatApp
	Channel ChatChannel
}

// AgentChannelSpec adds an agent to a chat or channel of a chat app.
type AgentChannelSpec struct {
	ChatApp ChatAppID
	Channel string
}

type ChatAccount struct {
	Kind   string
	Server string
	Handle string
}

// ChatAccountLink is a started link: the Telegram deep link to open, or the
// code to send the Mattermost bot, and when it stops working.
type ChatAccountLink struct {
	URL       string
	Code      string
	ExpiresAt time.Time
}

func (c *Client) ChatApps(ctx context.Context) ([]ChatApp, error) {
	workspace, ws, err := c.callerWorkspaceRelease(ctx)

	if err != nil {
		return nil, err
	}

	apps := make([]ChatApp, 0, len(ws.ChatApps))

	for name, app := range ws.ChatApps {
		apps = append(apps, toChatApp(ChatAppID{Workspace: workspace, Name: name}, app))
	}

	slices.SortFunc(apps, func(a, b ChatApp) int { return strings.Compare(a.ID.Name, b.ID.Name) })

	return apps, nil
}

func (c *Client) ChatApp(ctx context.Context, id ChatAppID) (*ChatApp, error) {
	app, err := c.chatApp(ctx, id)

	if err != nil {
		return nil, err
	}

	return new(toChatApp(id, app)), nil
}

func (c *Client) chatApp(ctx context.Context, id ChatAppID) (values.ChatApp, error) {
	ws, err := c.workspaceRelease(ctx, id.Workspace)

	if err != nil {
		return values.ChatApp{}, err
	}

	app, ok := ws.ChatApps[id.Name]

	if !ok {
		return values.ChatApp{}, usererror.Errorf("chat app %q not found", id.Name)
	}

	return app, nil
}

func toChatApp(id ChatAppID, app values.ChatApp) ChatApp {
	return ChatApp{
		ID:             id,
		Kind:           app.Kind,
		URL:            app.URL,
		Account:        app.Account,
		CredentialHint: credentialHint(app.Credential),
		Status:         app.Status,
		CheckedAt:      parseCheckedAt(app.CheckedAt),
	}
}

func (c *Client) CreateChatApp(ctx context.Context, name, kind, rawURL, token string) (*ChatApp, error) {
	workspace, ws, err := c.callerWorkspaceRelease(ctx)

	if err != nil {
		return nil, err
	}

	id := ChatAppID{Workspace: workspace, Name: name}

	if _, ok := ws.ChatApps[name]; ok {
		return nil, usererror.Errorf("a chat app named %q already exists", name)
	}

	serverURL := ""

	if kind == chat.Mattermost {
		serverURL, err = normalizeUpstreamURL(rawURL)

		if err != nil {
			return nil, err
		}
	}

	app, err := c.checkChatApp(ctx, kind, serverURL, token)

	if err != nil {
		return nil, err
	}

	if err := c.deployer.Workspaces().SetChatApp(ctx, id, app); err != nil {
		return nil, err
	}

	return new(toChatApp(id, app)), nil
}

func (c *Client) ReplaceChatAppToken(ctx context.Context, id ChatAppID, token string) (*ChatApp, error) {
	current, err := c.chatApp(ctx, id)

	if err != nil {
		return nil, err
	}

	app, err := c.checkChatApp(ctx, current.Kind, current.URL, token)

	if err != nil {
		return nil, err
	}

	if err := c.deployer.Workspaces().SetChatApp(ctx, id, app); err != nil {
		return nil, err
	}

	return new(toChatApp(id, app)), nil
}

func (c *Client) RefreshChatApp(ctx context.Context, id ChatAppID) (*ChatApp, error) {
	app, err := c.chatApp(ctx, id)

	if err != nil {
		return nil, err
	}

	account, check := c.upstream.ChatBot(ctx, app.Kind, app.URL, app.Credential)

	app.Status = values.CredentialStatus(check.Status)
	app.CheckedAt = time.Now().UTC().Format(time.RFC3339)

	if check.Status == upstream.Valid {
		app.Account = account
	}

	if err := c.deployer.Workspaces().SetChatApp(ctx, id, app); err != nil {
		return nil, err
	}

	return new(toChatApp(id, app)), nil
}

// DeleteChatApp removes a chat app and every agent's channels in it.
func (c *Client) DeleteChatApp(ctx context.Context, id ChatAppID) (bool, error) {
	if err := c.deployer.Workspaces().DeleteChatApp(ctx, id); err != nil {
		return false, err
	}

	return true, nil
}

func (c *Client) checkChatApp(ctx context.Context, kind, serverURL, token string) (values.ChatApp, error) {
	if kind != chat.Telegram && kind != chat.Mattermost {
		return values.ChatApp{}, usererror.Errorf("unsupported chat app %q", kind)
	}

	account, check := c.upstream.ChatBot(ctx, kind, serverURL, token)

	if check.Status != upstream.Valid {
		return values.ChatApp{}, usererror.Errorf("couldn't connect the chat app: %s", check.Reason)
	}

	return values.ChatApp{
		Kind:       kind,
		URL:        serverURL,
		Credential: token,
		Account:    account,
		Status:     values.CredentialValid,
		CheckedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// ChatAppChannels lists the channels a chat app's bot is a member of, which
// only Mattermost can tell.
func (c *Client) ChatAppChannels(ctx context.Context, id ChatAppID) ([]ChatChannel, error) {
	app, err := c.chatApp(ctx, id)

	if err != nil {
		return nil, err
	}

	channels, check := c.upstream.ChatChannels(ctx, app.Kind, app.URL, app.Credential)

	if check.Status != upstream.Valid {
		return nil, usererror.Errorf("couldn't list the channels: %s", check.Reason)
	}

	result := make([]ChatChannel, 0, len(channels))

	for _, channel := range channels {
		result = append(result, ChatChannel(channel))
	}

	slices.SortFunc(result, func(a, b ChatChannel) int { return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)) })

	return result, nil
}

// ChatChannel looks up one chat or channel of a chat app as its bot, which
// only works for one the bot is a member of.
func (c *Client) ChatChannel(ctx context.Context, id ChatAppID, channel string) (*ChatChannel, error) {
	app, err := c.chatApp(ctx, id)

	if err != nil {
		return nil, err
	}

	return c.lookupChatChannel(ctx, app, channel)
}

func (c *Client) lookupChatChannel(ctx context.Context, app values.ChatApp, channel string) (*ChatChannel, error) {
	if !values.IsValidChannel(channel) {
		return nil, usererror.Errorf("invalid channel %q", channel)
	}

	found, check := c.upstream.ChatChannel(ctx, app.Kind, app.URL, app.Credential, channel)

	if check.Status != upstream.Valid {
		return nil, usererror.Errorf("%s can't see that chat: %s", app.Account, check.Reason)
	}

	return new(ChatChannel(found)), nil
}

// ChatAppAgents returns the agents with a channel in a chat app.
func (c *Client) ChatAppAgents(ctx context.Context, id ChatAppID) ([]Agent, error) {
	ws, err := c.workspaceRelease(ctx, id.Workspace)

	if err != nil {
		return nil, err
	}

	if _, ok := ws.ChatApps[id.Name]; !ok {
		return nil, usererror.Errorf("chat app %q not found", id.Name)
	}

	agents, err := c.workspaceAgents(ctx, id.Workspace, ws)

	if err != nil {
		return nil, err
	}

	return slices.DeleteFunc(agents, func(agent Agent) bool {
		return !slices.ContainsFunc(agent.Channels, func(channel AgentChannel) bool { return channel.ChatApp.ID == id })
	}), nil
}

// SetAgentChannels replaces the chats an agent talks in. A chat the agent is
// already in keeps the name it was added with. A new one is looked up as the
// bot, so the bot must already be a member of it.
func (c *Client) SetAgentChannels(ctx context.Context, id AgentID, specs []AgentChannelSpec) (*Agent, error) {
	if _, _, err := c.loadAgent(ctx, id); err != nil {
		return nil, err
	}

	ws, err := c.workspaceRelease(ctx, id.Workspace)

	if err != nil {
		return nil, err
	}

	current := ws.Agents[values.AgentGrantKey(id)].Channels
	channels := make([]values.AgentChannel, 0, len(specs))

	for _, spec := range specs {
		if spec.ChatApp.Workspace != id.Workspace {
			return nil, usererror.Errorf("chat app %q not found", spec.ChatApp.Name)
		}

		app, ok := ws.ChatApps[spec.ChatApp.Name]

		if !ok {
			return nil, usererror.Errorf("chat app %q not found", spec.ChatApp.Name)
		}

		index := slices.IndexFunc(current, func(existing values.AgentChannel) bool {
			return existing.ChatApp == spec.ChatApp.Name && existing.Channel == spec.Channel
		})

		if index >= 0 {
			channels = append(channels, current[index])
			continue
		}

		found, err := c.lookupChatChannel(ctx, app, spec.Channel)

		if err != nil {
			return nil, err
		}

		channels = append(channels, values.AgentChannel{ChatApp: spec.ChatApp.Name, Channel: found.ID, Name: found.Name})
	}

	if err := c.deployer.Workspaces().SetAgentChannels(ctx, id, channels); err != nil {
		return nil, err
	}

	return c.Agent(ctx, id)
}

func toAgentChannels(id AgentID, grant values.AgentGrant, ws *values.Workspace) []AgentChannel {
	channels := make([]AgentChannel, 0, len(grant.Channels))

	for _, channel := range grant.Channels {
		app, ok := ws.ChatApps[channel.ChatApp]

		if !ok {
			continue
		}

		channels = append(channels, AgentChannel{
			ChatApp: toChatApp(ChatAppID{Workspace: id.Workspace, Name: channel.ChatApp}, app),
			Channel: ChatChannel{ID: channel.Channel, Name: channel.Name},
		})
	}

	return channels
}

// ChatAccounts returns the chat accounts linked to the caller.
func (c *Client) ChatAccounts(ctx context.Context) ([]ChatAccount, error) {
	userID, err := actorID(ctx)

	if err != nil {
		return nil, err
	}

	user, err := c.logto.User(ctx, userID)

	if err != nil {
		return nil, err
	}

	var accounts []ChatAccount

	for _, account := range chat.Accounts(user.CustomData) {
		accounts = append(accounts, ChatAccount{Kind: account.Kind, Server: account.Server, Handle: account.Handle})
	}

	return accounts, nil
}

// LinkChatAccount starts linking the caller's account in a chat app. It
// finishes when the code reaches the chat app's bot from that account, so a
// link only ever ties an account to the person who started it.
func (c *Client) LinkChatAccount(ctx context.Context, id ChatAppID) (*ChatAccountLink, error) {
	userID, err := actorID(ctx)

	if err != nil {
		return nil, err
	}

	app, err := c.chatApp(ctx, id)

	if err != nil {
		return nil, err
	}

	pending := chat.PendingLink{Nonce: rand.Text(), ExpiresAt: time.Now().UTC().Add(chat.LinkTTL)}

	if err := c.logto.UpdateUserCustomData(ctx, userID, chat.WithPending(pending)); err != nil {
		return nil, err
	}

	link := &ChatAccountLink{Code: chat.LinkCode(userID, pending.Nonce), ExpiresAt: pending.ExpiresAt}

	if app.Kind == chat.Telegram {
		link.URL = "https://t.me/" + url.PathEscape(strings.TrimPrefix(app.Account, "@")) + "?start=" + url.QueryEscape(link.Code)
	}

	return link, nil
}

// RemoveChatAccount unlinks the caller's account of kind on server.
func (c *Client) RemoveChatAccount(ctx context.Context, kind, server string) (bool, error) {
	userID, err := actorID(ctx)

	if err != nil {
		return false, err
	}

	user, err := c.logto.User(ctx, userID)

	if err != nil {
		return false, err
	}

	if err := c.logto.UpdateUserCustomData(ctx, userID, chat.WithoutAccount(user.CustomData, kind, server)); err != nil {
		return false, err
	}

	return true, nil
}
