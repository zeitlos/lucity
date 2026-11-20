package bridge

import (
	"context"
	"errors"
	"html"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/zeitlos/lucity/pkg/chat"
	"github.com/zeitlos/lucity/pkg/telegram"
	"github.com/zeitlos/lucity/services/switchboard/internal/bindings"
)

const telegramChunk = 3500

type telegramBot struct {
	bridge *Bridge
	app    bindings.ChatApp
	client *telegram.Client
	me     telegram.User
}

func newTelegramBot(b *Bridge, app bindings.ChatApp) *telegramBot {
	return &telegramBot{bridge: b, app: app, client: telegram.New(app.Token, b.telegramHTTP)}
}

func (t *telegramBot) run(ctx context.Context) error {
	me, err := t.client.Me(ctx)

	if err != nil {
		return err
	}

	t.me = me
	offset := int64(0)

	for ctx.Err() == nil {
		updates, err := t.client.Updates(ctx, offset)

		if err != nil {
			var apiErr *telegram.Error

			if errors.As(err, &apiErr) && apiErr.RetryAfter > 0 {
				select {
				case <-ctx.Done():
				case <-time.After(apiErr.RetryAfter):
				}

				continue
			}

			return err
		}

		for _, update := range updates {
			offset = update.UpdateID + 1
			t.handle(ctx, update)
		}
	}

	return nil
}

func (t *telegramBot) handle(ctx context.Context, update telegram.Update) {
	if callback := update.CallbackQuery; callback != nil && callback.Message != nil {
		t.handleCallback(ctx, callback)
		return
	}

	message := update.Message

	if message == nil || message.From == nil || message.From.IsBot || strings.TrimSpace(message.Text) == "" {
		return
	}

	at := t.thread(message)
	sender := t.sender(message.From)
	private := message.Chat.Type == "private"

	if command, args, ok := t.command(message.Text); ok {
		switch command {
		case "start":
			if private && args != "" {
				t.say(ctx, at, t.bridge.link(ctx, args, sender))
			} else if private {
				t.say(ctx, at, "This is the bot of a Lucity workspace. To link your account, open Linked accounts in Lucity's user menu. To add a group to an agent, send /connect there.")
			}
		case "connect":
			t.send(ctx, at, telegram.Outgoing{
				Text:    "Pick the agent this chat talks to in Lucity.",
				Buttons: []telegram.Button{{Text: "Add to an agent", URL: t.bridge.connectURL(t.app.ID, at.place.Channel)}},
			})
		case "new":
			if args == "" {
				t.say(ctx, at, "Send /new with your message to start a new session.")
			} else if !t.bridge.converse(ctx, t.app, at, sender, args, true, true) {
				t.unbound(ctx, at, private)
			}
		}

		return
	}

	text := strings.TrimSpace(strings.ReplaceAll(message.Text, "@"+t.me.Username, ""))

	if text == "" {
		return
	}

	if !t.bridge.converse(ctx, t.app, at, sender, text, true, false) {
		t.unbound(ctx, at, private)
	}
}

func (t *telegramBot) unbound(ctx context.Context, at thread, private bool) {
	if private {
		t.say(ctx, at, "This chat doesn't talk to an agent yet. Send /connect to add it to one.")
	}
}

func (t *telegramBot) handleCallback(ctx context.Context, callback *telegram.CallbackQuery) {
	decision, rest, _ := strings.Cut(callback.Data, ":")
	session, approval, ok := strings.Cut(rest, ":")

	reply := "This button doesn't work anymore."

	if ok && (decision == "y" || decision == "n") {
		reply, _ = t.bridge.answer(ctx, t.app, t.thread(callback.Message), t.sender(&callback.From), session, approval, decision == "y")
	}

	if err := t.client.Answer(ctx, callback.ID, reply); err != nil {
		slog.Debug("telegram: failed to answer a callback", "error", err)
	}
}

// command splits a bot command, and reports false for text that isn't one or
// a command meant for another bot.
func (t *telegramBot) command(text string) (string, string, bool) {
	if !strings.HasPrefix(text, "/") {
		return "", "", false
	}

	word, args, _ := strings.Cut(strings.TrimPrefix(text, "/"), " ")
	command, bot, addressed := strings.Cut(word, "@")

	if addressed && !strings.EqualFold(bot, t.me.Username) {
		return "", "", false
	}

	return strings.ToLower(command), strings.TrimSpace(args), true
}

func (t *telegramBot) thread(message *telegram.Message) thread {
	at := thread{place: bindings.Place{ChatApp: t.app.ID, Channel: strconv.FormatInt(message.Chat.ID, 10)}}

	if message.IsTopicMessage && message.MessageThreadID != 0 {
		at.thread = strconv.FormatInt(message.MessageThreadID, 10)
	}

	return at
}

func (t *telegramBot) sender(user *telegram.User) chat.Account {
	return chat.Account{Kind: chat.Telegram, ID: strconv.FormatInt(user.ID, 10), Handle: user.Handle()}
}

func (t *telegramBot) target(at thread) (int64, int64) {
	chatID, _ := strconv.ParseInt(at.place.Channel, 10, 64)
	threadID, _ := strconv.ParseInt(at.thread, 10, 64)

	return chatID, threadID
}

func (t *telegramBot) send(ctx context.Context, at thread, message telegram.Outgoing) string {
	message.ChatID, message.ThreadID = t.target(at)

	id, err := t.client.Send(ctx, message)

	if err != nil {
		slog.Warn("telegram: failed to send", "chat_app", t.app.ID, "error", err)
		return ""
	}

	return strconv.FormatInt(id, 10)
}

func (t *telegramBot) say(ctx context.Context, at thread, text string) {
	for _, part := range chunks(text, telegramChunk) {
		t.send(ctx, at, telegram.Outgoing{Text: part})
	}
}

func (t *telegramBot) status(ctx context.Context, at thread, id, text string) string {
	if id == "" {
		return t.send(ctx, at, telegram.Outgoing{Text: text})
	}

	t.edit(ctx, at, id, telegram.Outgoing{Text: text})

	return id
}

func (t *telegramBot) clear(ctx context.Context, at thread, id string) {
	chatID, _ := t.target(at)
	messageID, _ := strconv.ParseInt(id, 10, 64)

	if err := t.client.Delete(ctx, chatID, messageID); err != nil {
		slog.Debug("telegram: failed to delete", "chat_app", t.app.ID, "error", err)
	}
}

func (t *telegramBot) ask(ctx context.Context, at thread, request approvalRequest) string {
	return t.send(ctx, at, telegram.Outgoing{
		Text: t.card(request, request.agent+" wants to run", ""),
		HTML: true,
		Buttons: []telegram.Button{
			{Text: "Allow once", CallbackData: "y:" + request.session + ":" + request.approval},
			{Text: "Deny", CallbackData: "n:" + request.session + ":" + request.approval},
		},
	})
}

func (t *telegramBot) decided(ctx context.Context, at thread, id string, request approvalRequest, outcome string) {
	t.edit(ctx, at, id, telegram.Outgoing{Text: t.card(request, request.agent+" wanted to run", outcome), HTML: true})
}

func (t *telegramBot) card(request approvalRequest, heading, outcome string) string {
	var b strings.Builder

	b.WriteString("<b>" + html.EscapeString(heading) + "</b>")

	if request.title != "" && request.title != request.command {
		b.WriteString(" · " + html.EscapeString(truncate(request.title, 200)))
	}

	b.WriteString("\n<pre>" + html.EscapeString(truncate(request.command, 3000)) + "</pre>")

	if outcome != "" {
		b.WriteString("\n" + html.EscapeString(outcome))
	}

	return b.String()
}

func (t *telegramBot) edit(ctx context.Context, at thread, id string, message telegram.Outgoing) {
	messageID, _ := strconv.ParseInt(id, 10, 64)
	message.ChatID, _ = t.target(at)

	if err := t.client.Edit(ctx, messageID, message); err != nil {
		slog.Debug("telegram: failed to edit", "chat_app", t.app.ID, "error", err)
	}
}

func (t *telegramBot) typing(ctx context.Context, at thread) {
	chatID, threadID := t.target(at)
	_ = t.client.Typing(ctx, chatID, threadID)
}
