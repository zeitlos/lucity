package bridge

import (
	"context"
	"log/slog"
	"strings"

	"github.com/zeitlos/lucity/pkg/chat"
	"github.com/zeitlos/lucity/pkg/mattermost"
	"github.com/zeitlos/lucity/services/switchboard/internal/bindings"
)

const mattermostChunk = 15000

type mattermostBot struct {
	bridge *Bridge
	app    bindings.ChatApp
	client *mattermost.Client
	me     mattermost.User
}

func newMattermostBot(b *Bridge, app bindings.ChatApp) *mattermostBot {
	return &mattermostBot{bridge: b, app: app, client: mattermost.New(app.URL, app.Token, b.mattermostHTTP)}
}

func (m *mattermostBot) run(ctx context.Context) error {
	callCtx, cancel := context.WithTimeout(ctx, callTimeout)
	me, err := m.client.Me(callCtx)
	cancel()

	if err != nil {
		return err
	}

	m.me = me

	return m.client.Listen(ctx, func(posted mattermost.Posted) {
		m.handle(ctx, posted)
	})
}

// handle routes a post. In a direct message it can only link an account. In
// a channel, a post that mentions the bot starts a session in its thread,
// and every reply in that thread continues it.
func (m *mattermostBot) handle(ctx context.Context, posted mattermost.Posted) {
	post := posted.Post

	if post.UserID == m.me.ID || post.Type != "" || strings.TrimSpace(post.Message) == "" {
		return
	}

	root := post.RootID

	if root == "" {
		root = post.ID
	}

	at := thread{place: bindings.Place{ChatApp: m.app.ID, Channel: post.ChannelID}, thread: root}

	if posted.ChannelType == mattermost.ChannelDirect {
		m.direct(ctx, at, post)
		return
	}

	if posted.ChannelType != mattermost.ChannelOpen && posted.ChannelType != mattermost.ChannelPrivate {
		return
	}

	mention := "@" + m.me.Username
	mentioned := strings.Contains(post.Message, mention)
	text := strings.TrimSpace(strings.ReplaceAll(post.Message, mention, ""))
	sender := chat.Account{Kind: chat.Mattermost, Server: m.app.URL, ID: post.UserID}

	if post.RootID == "" && !mentioned {
		return
	}

	if post.RootID != "" {
		switch strings.ToLower(text) {
		case "allow", "deny":
			if m.bridge.answerPending(ctx, m.app, at, sender, strings.EqualFold(text, "allow")) {
				return
			}
		}
	}

	if text == "" {
		return
	}

	m.bridge.converse(ctx, m.app, at, sender, text, mentioned, false)
}

func (m *mattermostBot) direct(ctx context.Context, at thread, post mattermost.Post) {
	at.thread = ""
	word, code, _ := strings.Cut(strings.TrimSpace(post.Message), " ")

	if !strings.EqualFold(word, "link") || strings.TrimSpace(code) == "" {
		m.say(ctx, at, "This is the bot of a Lucity workspace. To link your account, start a link under Linked accounts in Lucity's user menu, and send me `link <code>` here. To talk to an agent, mention me in a channel it's added to.")
		return
	}

	callCtx, cancel := context.WithTimeout(ctx, callTimeout)
	user, err := m.client.User(callCtx, post.UserID)
	cancel()

	if err != nil {
		m.say(ctx, at, "Linking failed. Try again in a moment.")
		return
	}

	account := chat.Account{Kind: chat.Mattermost, Server: m.app.URL, ID: post.UserID, Handle: "@" + user.Username}
	m.say(ctx, at, m.bridge.link(ctx, strings.TrimSpace(code), account))
}

func (m *mattermostBot) post(ctx context.Context, at thread, message string) string {
	callCtx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	created, err := m.client.CreatePost(callCtx, mattermost.Post{ChannelID: at.place.Channel, RootID: at.thread, Message: message})

	if err != nil {
		slog.Warn("mattermost: failed to post", "chat_app", m.app.ID, "error", err)
		return ""
	}

	return created.ID
}

func (m *mattermostBot) edit(ctx context.Context, id, message string) {
	callCtx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	if err := m.client.EditPost(callCtx, id, message); err != nil {
		slog.Debug("mattermost: failed to edit", "chat_app", m.app.ID, "error", err)
	}
}

func (m *mattermostBot) say(ctx context.Context, at thread, text string) {
	for _, part := range chunks(text, mattermostChunk) {
		m.post(ctx, at, part)
	}
}

func (m *mattermostBot) status(ctx context.Context, at thread, id, text string) string {
	if id == "" {
		return m.post(ctx, at, "_"+text+"_")
	}

	m.edit(ctx, id, "_"+text+"_")

	return id
}

func (m *mattermostBot) clear(ctx context.Context, _ thread, id string) {
	callCtx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	if err := m.client.DeletePost(callCtx, id); err != nil {
		slog.Debug("mattermost: failed to delete", "chat_app", m.app.ID, "error", err)
	}
}

func (m *mattermostBot) ask(ctx context.Context, at thread, request approvalRequest) string {
	return m.post(ctx, at, card(request, request.agent+" wants to run")+"\nReply **allow** or **deny** in this thread.")
}

func (m *mattermostBot) decided(ctx context.Context, _ thread, id string, request approvalRequest, outcome string) {
	m.edit(ctx, id, card(request, request.agent+" wanted to run")+"\n"+outcome)
}

func card(request approvalRequest, heading string) string {
	var b strings.Builder

	b.WriteString("**" + heading + "**")

	if request.title != "" && request.title != request.command {
		b.WriteString(" · " + truncate(request.title, 200))
	}

	b.WriteString("\n```\n" + strings.ReplaceAll(truncate(request.command, 3000), "```", "'''") + "\n```")

	return b.String()
}

func (m *mattermostBot) typing(ctx context.Context, at thread) {
	callCtx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	_ = m.client.Typing(callCtx, m.me.ID, at.place.Channel, at.thread)
}
