// Package bridge connects the chat apps of every workspace to the runners of
// the agents they are bound to. It runs one bot per chat app, turns chat
// messages from linked members into session messages, and relays what the
// sessions do back into the chat they started in.
package bridge

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"sync"
	"time"

	"github.com/zeitlos/lucity/pkg/agentrunner"
	"github.com/zeitlos/lucity/pkg/chat"
	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/services/switchboard/internal/bindings"
	"github.com/zeitlos/lucity/services/switchboard/internal/people"
)

const (
	callTimeout       = 15 * time.Second
	discoverEvery     = 15 * time.Second
	followIdle        = 10 * time.Minute
	restartAtMost     = time.Minute
	decisionAllowOnce = "allow_once"
	decisionDeny      = "deny"
)

// thread is where a session lives in a chat app: a chat or channel, and the
// topic or thread in it. An empty thread is the chat itself.
type thread struct {
	place  bindings.Place
	thread string
}

func (t thread) origin() agentrunner.Origin {
	return agentrunner.Origin{ChatApp: t.place.ChatApp.Name, Channel: t.place.Channel, Thread: t.thread}
}

// bot is one chat app's bot: it receives until its context ends, and posts
// what the bridge relays in the chat app's own format.
type bot interface {
	run(ctx context.Context) error
	say(ctx context.Context, at thread, text string)
	status(ctx context.Context, at thread, id, text string) string
	clear(ctx context.Context, at thread, id string)
	ask(ctx context.Context, at thread, request approvalRequest) string
	decided(ctx context.Context, at thread, id string, request approvalRequest, outcome string)
	typing(ctx context.Context, at thread)
}

type approvalRequest struct {
	session  string
	approval string
	agent    string
	title    string
	command  string
}

type running struct {
	app    bindings.ChatApp
	bot    bot
	cancel context.CancelFunc
}

type sessionKey struct {
	agent  platform.AgentID
	origin agentrunner.Origin
}

type hintKey struct {
	app    platform.ChatAppID
	sender string
}

type Bridge struct {
	runner    *agentrunner.Client
	people    *people.Directory
	watcher   *bindings.Watcher
	dashboard string

	telegramHTTP   *http.Client
	mattermostHTTP *http.Client

	ctx    context.Context
	cancel context.CancelFunc

	mu        sync.Mutex
	snapshot  bindings.Snapshot
	bots      map[platform.ChatAppID]*running
	sessions  map[sessionKey]string
	followers map[platform.AgentSessionID]bool
	hinted    map[hintKey]bool
}

func New(runner *agentrunner.Client, people *people.Directory, watcher *bindings.Watcher, dashboardURL string) *Bridge {
	ctx, cancel := context.WithCancel(context.Background())

	return &Bridge{
		runner:         runner,
		people:         people,
		watcher:        watcher,
		dashboard:      dashboardURL,
		telegramHTTP:   &http.Client{Timeout: 70 * time.Second},
		mattermostHTTP: &http.Client{},
		ctx:            ctx,
		cancel:         cancel,
		bots:           map[platform.ChatAppID]*running{},
		sessions:       map[sessionKey]string{},
		followers:      map[platform.AgentSessionID]bool{},
		hinted:         map[hintKey]bool{},
	}
}

func (b *Bridge) Label() string {
	return "switchboard"
}

func (b *Bridge) Start() error {
	go func() {
		if err := b.watcher.Run(b.ctx); err != nil && b.ctx.Err() == nil {
			slog.Error("bindings watch stopped", "error", err)
		}
	}()

	discover := time.NewTicker(discoverEvery)
	defer discover.Stop()

	for {
		select {
		case <-b.ctx.Done():
			return nil
		case <-b.watcher.Changed():
			b.reconcile()
		case <-discover.C:
			b.discover()
		}
	}
}

func (b *Bridge) Shutdown(context.Context) error {
	b.cancel()

	return nil
}

// reconcile runs a bot for every chat app, restarting one whose token or
// server changed and stopping one that is gone.
func (b *Bridge) reconcile() {
	snapshot := b.watcher.Snapshot()

	b.mu.Lock()
	defer b.mu.Unlock()

	b.snapshot = snapshot

	for id, current := range b.bots {
		if app, ok := snapshot.ChatApps[id]; !ok || app != current.app {
			current.cancel()
			delete(b.bots, id)
		}
	}

	for id, app := range snapshot.ChatApps {
		if _, ok := b.bots[id]; ok {
			continue
		}

		var chatBot bot

		switch app.Kind {
		case chat.Telegram:
			chatBot = newTelegramBot(b, app)
		case chat.Mattermost:
			chatBot = newMattermostBot(b, app)
		default:
			slog.Warn("bridge: unsupported chat app", "chat_app", id, "kind", app.Kind)
			continue
		}

		ctx, cancel := context.WithCancel(b.ctx)
		b.bots[id] = &running{app: app, bot: chatBot, cancel: cancel}

		go b.keepRunning(ctx, id, chatBot)
	}
}

func (b *Bridge) keepRunning(ctx context.Context, id platform.ChatAppID, chatBot bot) {
	wait := 5 * time.Second

	for ctx.Err() == nil {
		started := time.Now()
		err := chatBot.run(ctx)

		if ctx.Err() != nil {
			return
		}

		if time.Since(started) > restartAtMost {
			wait = 5 * time.Second
		}

		slog.Warn("bridge: chat app stopped, restarting", "chat_app", id, "error", err, "in", wait)

		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}

		wait = min(wait*2, restartAtMost)
	}
}

func (b *Bridge) bot(id platform.ChatAppID) bot {
	b.mu.Lock()
	defer b.mu.Unlock()

	if current, ok := b.bots[id]; ok {
		return current.bot
	}

	return nil
}

func (b *Bridge) binding(place bindings.Place) (bindings.Binding, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	binding, ok := b.snapshot.Bindings[place]

	return binding, ok
}

// connectURL is the dashboard page that adds a chat to an agent, in the chat
// app's workspace.
func (b *Bridge) connectURL(app platform.ChatAppID, channel string) string {
	return b.dashboard + "/" + url.PathEscape(app.Workspace) + "/~/add-channel?" + url.Values{"chatApp": {app.Name}, "channel": {channel}}.Encode()
}

// member returns the member of the chat app's workspace that sender is
// linked to. An unlinked sender is told once how to link.
func (b *Bridge) member(ctx context.Context, app bindings.ChatApp, at thread, sender chat.Account) *people.Member {
	member, err := b.people.Find(ctx, app.ID.Workspace, sender.Kind, sender.Server, sender.ID)

	if err != nil {
		slog.Warn("bridge: failed to look up a chat account", "chat_app", app.ID, "error", err)
		return nil
	}

	if member != nil {
		return member
	}

	key := hintKey{app: app.ID, sender: sender.ID}

	b.mu.Lock()
	hinted := b.hinted[key]
	b.hinted[key] = true
	b.mu.Unlock()

	if !hinted {
		if chatBot := b.bot(app.ID); chatBot != nil {
			chatBot.say(ctx, at, sender.Handle+", link your account to Lucity first: open Linked accounts in Lucity's user menu.")
		}
	}

	return nil
}

// converse sends text from sender to the session of at, or starts one when
// fresh is true or none exists and mayStart is. It reports whether the
// message went to an agent.
func (b *Bridge) converse(ctx context.Context, app bindings.ChatApp, at thread, sender chat.Account, text string, mayStart, fresh bool) bool {
	binding, ok := b.binding(at.place)

	if !ok {
		return false
	}

	member := b.member(ctx, app, at, sender)

	if member == nil {
		return true
	}

	chatBot := b.bot(app.ID)

	if chatBot == nil {
		return true
	}

	origin := at.origin()
	sessionName := ""

	if !fresh {
		sessionName = b.findSession(ctx, binding.Agent, origin)
	}

	if sessionName == "" {
		if !mayStart && !fresh {
			return false
		}

		callCtx, cancel := context.WithTimeout(ctx, callTimeout)
		session, err := b.runner.StartSession(callCtx, binding.Agent, agentrunner.StartRequest{Message: text, StartedBy: member.ID, Origin: &origin})
		cancel()

		if err != nil {
			chatBot.say(ctx, at, "Couldn't start a session with "+binding.Agent.Name+": "+err.Error())
			return true
		}

		b.mu.Lock()
		b.sessions[sessionKey{agent: binding.Agent, origin: origin}] = session.ID
		b.mu.Unlock()

		b.follow(platform.AgentSessionID{Agent: binding.Agent, Session: session.ID}, at, "")

		return true
	}

	id := platform.AgentSessionID{Agent: binding.Agent, Session: sessionName}
	b.followCurrentTurn(ctx, id, at)

	callCtx, cancel := context.WithTimeout(ctx, callTimeout)
	_, err := b.runner.SendMessage(callCtx, id, text, member.ID, &origin)
	cancel()

	if err != nil {
		chatBot.say(ctx, at, "Couldn't send that to "+binding.Agent.Name+": "+err.Error())
	}

	return true
}

// findSession returns the newest session of agent that started at origin.
func (b *Bridge) findSession(ctx context.Context, agent platform.AgentID, origin agentrunner.Origin) string {
	key := sessionKey{agent: agent, origin: origin}

	b.mu.Lock()
	cached, ok := b.sessions[key]
	b.mu.Unlock()

	if ok {
		return cached
	}

	b.refreshSessions(ctx, agent)

	b.mu.Lock()
	defer b.mu.Unlock()

	return b.sessions[key]
}

func (b *Bridge) refreshSessions(ctx context.Context, agent platform.AgentID) []agentrunner.Session {
	callCtx, cancel := context.WithTimeout(ctx, callTimeout)
	sessions, err := b.runner.Sessions(callCtx, agent)
	cancel()

	if err != nil {
		slog.Warn("bridge: failed to list sessions", "agent", agent, "error", err)
		return nil
	}

	slices.SortFunc(sessions, func(x, y agentrunner.Session) int { return x.StartedAt.Compare(y.StartedAt) })

	b.mu.Lock()
	defer b.mu.Unlock()

	for _, session := range sessions {
		if session.Origin != nil {
			b.sessions[sessionKey{agent: agent, origin: *session.Origin}] = session.ID
		}
	}

	return sessions
}

// discover follows the chat sessions that are working or waiting for an
// approval, such as one someone wrote to from the dashboard.
func (b *Bridge) discover() {
	b.mu.Lock()
	snapshot := b.snapshot
	b.mu.Unlock()

	for _, agent := range snapshot.Agents() {
		for _, session := range b.refreshSessions(b.ctx, agent) {
			if session.Origin == nil || session.Status == agentrunner.StatusIdle {
				continue
			}

			place := bindings.Place{ChatApp: platform.ChatAppID{Workspace: agent.Workspace, Name: session.Origin.ChatApp}, Channel: session.Origin.Channel}

			if binding, ok := snapshot.Bindings[place]; !ok || binding.Agent != agent {
				continue
			}

			b.followCurrentTurn(b.ctx, platform.AgentSessionID{Agent: agent, Session: session.ID}, thread{place: place, thread: session.Origin.Thread})
		}
	}
}

// followCurrentTurn follows a session from the start of its current turn, so
// nothing it does next, or did since its last turn ended, is missed.
func (b *Bridge) followCurrentTurn(ctx context.Context, id platform.AgentSessionID, at thread) {
	b.mu.Lock()
	following := b.followers[id]
	b.mu.Unlock()

	if following {
		return
	}

	callCtx, cancel := context.WithTimeout(ctx, callTimeout)
	session, err := b.runner.Session(callCtx, id)
	cancel()

	if err != nil {
		slog.Warn("bridge: failed to read a session", "session", id, "error", err)
		return
	}

	after := ""

	for _, event := range session.Events {
		if event.Type == agentrunner.EventTurnEnd {
			after = event.ID
		}
	}

	b.follow(id, at, after)
}

func (b *Bridge) follow(id platform.AgentSessionID, at thread, after string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.followers[id] {
		return
	}

	b.followers[id] = true

	go func() {
		b.relay(id, at, after)

		b.mu.Lock()
		delete(b.followers, id)
		b.mu.Unlock()
	}()
}

// answer decides an approval for sender. It returns what to tell sender, and
// whether the decision went through.
func (b *Bridge) answer(ctx context.Context, app bindings.ChatApp, at thread, sender chat.Account, session, approval string, allow bool) (string, bool) {
	binding, ok := b.binding(at.place)

	if !ok {
		return "This chat isn't added to an agent anymore.", false
	}

	member := b.member(ctx, app, at, sender)

	if member == nil {
		return "Link your account to Lucity first.", false
	}

	decision := decisionDeny

	if allow {
		decision = decisionAllowOnce
	}

	callCtx, cancel := context.WithTimeout(ctx, callTimeout)
	_, err := b.runner.AnswerApproval(callCtx, platform.AgentSessionID{Agent: binding.Agent, Session: session}, approval, decision, member.ID)
	cancel()

	if err != nil {
		return errorText(err), false
	}

	if allow {
		return "Allowed once", true
	}

	return "Denied", true
}

// answerPending decides the pending approval of at's session for sender. It
// reports false when there is none, so the reply is just a message.
func (b *Bridge) answerPending(ctx context.Context, app bindings.ChatApp, at thread, sender chat.Account, allow bool) bool {
	binding, ok := b.binding(at.place)

	if !ok {
		return false
	}

	sessionName := b.findSession(ctx, binding.Agent, at.origin())

	if sessionName == "" {
		return false
	}

	id := platform.AgentSessionID{Agent: binding.Agent, Session: sessionName}

	callCtx, cancel := context.WithTimeout(ctx, callTimeout)
	session, err := b.runner.Session(callCtx, id)
	cancel()

	if err != nil {
		return false
	}

	pending := ""

	for _, event := range session.Events {
		if event.Type != agentrunner.EventApproval {
			continue
		}

		if event.Approval.Status == "pending" {
			pending = event.Approval.ApprovalID
		} else if event.Approval.ApprovalID == pending {
			pending = ""
		}
	}

	if pending == "" {
		return false
	}

	if reply, ok := b.answer(ctx, app, at, sender, sessionName, pending, allow); !ok {
		if chatBot := b.bot(app.ID); chatBot != nil {
			chatBot.say(ctx, at, reply)
		}
	}

	return true
}

// link finishes a link from the chat account that sent code, and returns the
// reply for it.
func (b *Bridge) link(ctx context.Context, code string, account chat.Account) string {
	name, err := b.people.Link(ctx, code, account)

	switch {
	case errors.Is(err, people.ErrExpired):
		return "That link expired. Start it again in Lucity, under Linked accounts in the user menu."
	case errors.Is(err, people.ErrInvalidCode):
		return "That code isn't valid. Start a link in Lucity, under Linked accounts in the user menu."
	case err != nil:
		slog.Warn("bridge: failed to link a chat account", "error", err)
		return "Linking failed. Try again in a moment."
	}

	return "Linked. " + account.Handle + " now counts as " + name + " when you talk to agents in Lucity."
}

func (b *Bridge) memberName(ctx context.Context, workspace, userID string) string {
	if userID == "" {
		return ""
	}

	member, err := b.people.Member(ctx, workspace, userID)

	if err != nil || member == nil {
		return "a former member"
	}

	return member.Name
}
