package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/zeitlos/lucity/services/switchboard/acp"
	"github.com/zeitlos/lucity/services/switchboard/telegram"
)

type turn struct {
	ctx       context.Context
	cancel    context.CancelFunc
	finished  chan struct{}
	sessionID string
	chatID    int64
	threadID  int64
	thread    string
	statusID  int64

	mu      sync.Mutex
	steps   []string
	pending strings.Builder
	last    string
	changed bool
}

func (t *turn) record(update acp.Update) {
	t.mu.Lock()
	defer t.mu.Unlock()
	switch update.Kind {
	case "agent_message_chunk":
		t.pending.WriteString(update.Text())
	case "tool_call":
		if strings.TrimSpace(t.pending.String()) != "" {
			t.last = t.pending.String()
		}
		t.pending.Reset()
		t.steps = append(t.steps, update.Title)
		t.changed = true
	}
}

func (t *turn) answer() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if text := strings.TrimSpace(t.pending.String()); text != "" {
		return text
	}
	return strings.TrimSpace(t.last)
}

func (t *turn) progress() (string, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.changed {
		return "", false
	}
	t.changed = false
	return fmt.Sprintf("Working · %s · %s", count(len(t.steps)), truncate(t.steps[len(t.steps)-1], 160)), true
}

func (t *turn) stepCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.steps)
}

func prompt(message *telegram.Message, topic, text, quoted string) string {
	var b strings.Builder
	sent := time.Unix(message.Date, 0).Format("2006-01-02 15:04 MST")
	fmt.Fprintf(&b, "Telegram message from %s at %s", actor(message.From), sent)
	if topic != "" {
		fmt.Fprintf(&b, " in the topic %q", topic)
	}
	if quoted != "" {
		b.WriteString(", about this alert:\n")
		for line := range strings.SplitSeq(quoted, "\n") {
			b.WriteString("> ")
			b.WriteString(line)
			b.WriteString("\n")
		}
	} else {
		b.WriteString(":\n")
	}
	b.WriteString("\n")
	b.WriteString(text)
	return b.String()
}

func topicOf(message *telegram.Message) string {
	if reply := message.ReplyToMessage; reply != nil && reply.ForumTopicCreated != nil {
		return reply.ForumTopicCreated.Name
	}
	return ""
}

func originThread(message *telegram.Message) int64 {
	if message.IsTopicMessage {
		return message.MessageThreadID
	}
	return 0
}

func alertTopicName(alert *telegram.Message) string {
	title, _, _ := strings.Cut(alert.Text, "\n")
	title = strings.TrimPrefix(strings.TrimPrefix(title, "FIRING: "), "RESOLVED: ")
	return truncate(strings.TrimSpace(title), 100) + " · " + time.Unix(alert.Date, 0).Format("15:04")
}

func questionTopicName(question string) string {
	title, _, _ := strings.Cut(question, "\n")
	return truncate(strings.TrimSpace(title), 100)
}

func topicLink(chatID, threadID int64) string {
	internal := strings.TrimPrefix(strconv.FormatInt(chatID, 10), "-100")
	return "https://t.me/c/" + internal + "/" + strconv.FormatInt(threadID, 10)
}

func literal(call acp.ToolCall) string {
	var input struct {
		Command string `json:"command"`
	}
	if json.Unmarshal(call.RawInput, &input) == nil && input.Command != "" {
		return input.Command
	}
	if raw := strings.TrimSpace(string(call.RawInput)); raw != "" && raw != "null" && raw != "{}" {
		return raw
	}
	return call.Title
}

func approvalCard(call acp.ToolCall, command, outcome string) string {
	var b strings.Builder
	b.WriteString("<b>Needs your approval</b>")
	if call.Title != "" && call.Title != command {
		b.WriteString(" · ")
		b.WriteString(html.EscapeString(truncate(call.Title, 200)))
	}
	b.WriteString("\n<pre>")
	b.WriteString(html.EscapeString(truncate(command, 3000)))
	b.WriteString("</pre>")
	if outcome != "" {
		b.WriteString("\n")
		b.WriteString(html.EscapeString(outcome))
	}
	return b.String()
}

func finalStatus(stopReason string, err error, steps int) string {
	switch {
	case err != nil:
		return "Failed · " + truncate(err.Error(), 300)
	case stopReason == "end_turn":
		return "Done · " + count(steps)
	case stopReason == "cancelled":
		return "Stopped · " + count(steps)
	default:
		return "Ended (" + stopReason + ") · " + count(steps)
	}
}

func stopButton(sessionID string) []telegram.Button {
	return []telegram.Button{{Text: "Stop", CallbackData: "stop:" + sessionID}}
}

func count(steps int) string {
	if steps == 1 {
		return "1 step"
	}
	return strconv.Itoa(steps) + " steps"
}

func chunks(text string, size int) []string {
	var parts []string
	for len(text) > size {
		cut := strings.LastIndex(text[:size], "\n")
		if cut <= 0 {
			cut = size
			for cut > 0 && !utf8.RuneStart(text[cut]) {
				cut--
			}
		}
		parts = append(parts, text[:cut])
		text = strings.TrimLeft(text[cut:], "\n")
	}
	if strings.TrimSpace(text) != "" {
		parts = append(parts, text)
	}
	return parts
}

func truncate(text string, size int) string {
	if len(text) <= size {
		return text
	}
	cut := size
	for cut > 0 && !utf8.RuneStart(text[cut]) {
		cut--
	}
	return text[:cut] + "…"
}

func actor(user *telegram.User) string {
	if user == nil {
		return "unknown"
	}
	if user.Username != "" {
		return "@" + user.Username
	}
	return "user " + strconv.FormatInt(user.ID, 10)
}

func actorID(user *telegram.User) string {
	if user == nil {
		return ""
	}
	return "telegram:" + strconv.FormatInt(user.ID, 10)
}

func senderID(message *telegram.Message) int64 {
	if message.From == nil {
		return 0
	}
	return message.From.ID
}
