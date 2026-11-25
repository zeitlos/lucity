package bridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/zeitlos/lucity/pkg/agentrunner"
	"github.com/zeitlos/lucity/pkg/platform"
)

const typingEvery = 4 * time.Second

// relay posts what a session does into its chat, until its turn has ended
// and nothing happened for followIdle, with nothing waiting for an approval.
func (b *Bridge) relay(id platform.AgentSessionID, at thread, after string) {
	ctx, cancel := context.WithCancel(b.ctx)
	defer cancel()

	events, err := b.runner.Events(ctx, id, after)

	if err != nil {
		slog.Warn("bridge: failed to follow a session", "session", id, "error", err)
		return
	}

	chatBot := b.bot(at.place.ChatApp)

	r := relayState{
		bridge:  b,
		id:      id,
		at:      at,
		pending: at.thread == "" && chatBot != nil && chatBot.threads(),
		opening: after == "",
		calls:   map[string]agentrunner.ToolCall{},
		cards:   map[string]string{},
	}

	idle := time.NewTimer(followIdle)
	defer idle.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-idle.C:
			if !r.turn && len(r.cards) == 0 {
				return
			}

			idle.Reset(followIdle)
		case event, ok := <-events:
			if !ok {
				return
			}

			if event.Type == agentrunner.EventMessageDraft {
				continue
			}

			if err := r.handle(ctx, event); err != nil {
				slog.Error("bridge: stopped following a session", "session", id, "error", err)
				return
			}

			b.setCursor(id, event.ID)

			if !idle.Stop() {
				select {
				case <-idle.C:
				default:
				}
			}

			idle.Reset(followIdle)
		}
	}
}

type relayState struct {
	bridge  *Bridge
	id      platform.AgentSessionID
	at      thread
	pending bool
	opening bool
	title   string

	turn   bool
	status string
	steps  int
	typed  time.Time
	calls  map[string]agentrunner.ToolCall
	cards  map[string]string
}

func (r *relayState) handle(ctx context.Context, event agentrunner.Event) error {
	chatBot := r.bridge.bot(r.at.place.ChatApp)

	if chatBot == nil {
		return nil
	}

	r.turn = event.Type != agentrunner.EventTurnEnd

	switch event.Type {
	case agentrunner.EventUserMessage:
		opening := r.opening
		r.opening = false

		if event.Via != nil || r.pending {
			return nil
		}

		if _, ok := agentrunner.ActorTrigger(event.Author); ok {
			if opening {
				return nil
			}

			title, err := r.sessionTitle(ctx)

			if err != nil {
				return err
			}

			chatBot.say(ctx, r.at, title+": new event")

			return nil
		}

		name := r.bridge.memberName(ctx, r.id.Agent.Workspace, event.Author)
		chatBot.say(ctx, r.at, name+" in Lucity: "+event.Text)
	case agentrunner.EventMessage:
		if strings.TrimSpace(event.Text) != "" {
			return r.post(ctx, chatBot, event.Text)
		}
	case agentrunner.EventToolCall:
		call := *event.ToolCall

		if _, seen := r.calls[call.CallID]; !seen {
			r.steps++

			if !r.pending {
				r.status = chatBot.status(ctx, r.at, r.status, "Working · "+steps(r.steps)+" · "+truncate(call.Title, 160))
			}
		}

		r.calls[call.CallID] = call

		if !r.pending && time.Since(r.typed) > typingEvery {
			r.typed = time.Now()
			chatBot.typing(ctx, r.at)
		}
	case agentrunner.EventApproval:
		approval := event.Approval
		request := approvalRequest{
			session:  r.id.Session,
			approval: approval.ApprovalID,
			agent:    r.id.Agent.Name,
			title:    approval.Title,
			command:  command(r.calls[approval.CallID], approval.Title),
		}

		if approval.Status == "pending" {
			r.cards[approval.ApprovalID] = chatBot.ask(ctx, r.at, request)
			return nil
		}

		card, ok := r.cards[approval.ApprovalID]

		if !ok {
			return nil
		}

		delete(r.cards, approval.ApprovalID)
		chatBot.decided(ctx, r.at, card, request, r.outcome(ctx, approval))
	case agentrunner.EventNotice:
		text := event.Notice.Title

		if event.Notice.Description != "" {
			text += ": " + event.Notice.Description
		}

		return r.post(ctx, chatBot, text)
	case agentrunner.EventTurnEnd:
		end := event.TurnEnd

		if r.status != "" {
			switch end.Reason {
			case "completed":
				chatBot.clear(ctx, r.at, r.status)
			case "stopped":
				chatBot.status(ctx, r.at, r.status, "Stopped · "+steps(r.steps))
			default:
				chatBot.status(ctx, r.at, r.status, "Failed · "+steps(r.steps))
			}
		}

		r.status = ""
		r.steps = 0

		if end.Reason == "failed" && end.Error != "" {
			return r.post(ctx, chatBot, "The turn failed: "+end.Error)
		}
	}

	return nil
}

// post says text in the session's thread, or starts the thread with it when
// the session has none yet.
func (r *relayState) post(ctx context.Context, chatBot bot, text string) error {
	if !r.pending {
		chatBot.say(ctx, r.at, text)
		return nil
	}

	thread := chatBot.root(ctx, r.at, text)

	if thread == "" {
		return errors.New("the thread's first post failed")
	}

	return r.adopt(ctx, thread)
}

// adopt makes thread the session's thread, and tells the runner, so replies
// there and later relays find the session.
func (r *relayState) adopt(ctx context.Context, thread string) error {
	link := r.at.chat()
	link.Thread = thread

	callCtx, cancel := context.WithTimeout(ctx, callTimeout)
	_, err := r.bridge.runner.SetChat(callCtx, r.id, link)
	cancel()

	if err != nil {
		return fmt.Errorf("saving the session's thread: %w", err)
	}

	r.pending = false
	r.at.thread = thread

	r.bridge.mu.Lock()
	r.bridge.sessions[sessionKey{agent: r.id.Agent, chat: link}] = r.id.Session
	r.bridge.mu.Unlock()

	return nil
}

// sessionTitle is the session's title: for a trigger's session, the trigger's
// name.
func (r *relayState) sessionTitle(ctx context.Context) (string, error) {
	if r.title != "" {
		return r.title, nil
	}

	callCtx, cancel := context.WithTimeout(ctx, callTimeout)
	session, err := r.bridge.runner.Session(callCtx, r.id)
	cancel()

	if err != nil {
		return "", fmt.Errorf("reading the session's title: %w", err)
	}

	r.title = session.Title

	return r.title, nil
}

func (r *relayState) outcome(ctx context.Context, approval *agentrunner.Approval) string {
	if approval.DecidedBy == "" {
		return "Denied, nobody answered in time."
	}

	name := r.bridge.memberName(ctx, r.id.Agent.Workspace, approval.DecidedBy)

	if approval.Status == "allowed" {
		return "Allowed once by " + name + "."
	}

	return "Denied by " + name + "."
}

// command is what an approval would run: a shell command when the tool call
// has one, its input otherwise.
func command(call agentrunner.ToolCall, fallback string) string {
	var input struct {
		Command string `json:"command"`
	}

	if json.Unmarshal([]byte(call.Input), &input) == nil && input.Command != "" {
		return input.Command
	}

	if raw := strings.TrimSpace(call.Input); raw != "" && raw != "null" && raw != "{}" {
		return raw
	}

	return fallback
}

func steps(n int) string {
	if n == 1 {
		return "1 step"
	}

	return strconv.Itoa(n) + " steps"
}

func errorText(err error) string {
	return truncate(fmt.Sprint(err), 300)
}
