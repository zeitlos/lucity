package bridge

import (
	"context"
	"encoding/json"
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

	r := relayState{
		bridge: b,
		id:     id,
		at:     at,
		origin: at.origin(),
		calls:  map[string]agentrunner.ToolCall{},
		cards:  map[string]string{},
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

			r.handle(ctx, event)

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
	bridge *Bridge
	id     platform.AgentSessionID
	at     thread
	origin agentrunner.Origin

	turn   bool
	status string
	steps  int
	typed  time.Time
	calls  map[string]agentrunner.ToolCall
	cards  map[string]string
}

func (r *relayState) handle(ctx context.Context, event agentrunner.Event) {
	chatBot := r.bridge.bot(r.at.place.ChatApp)

	if chatBot == nil {
		return
	}

	r.turn = event.Type != agentrunner.EventTurnEnd

	switch event.Type {
	case agentrunner.EventUserMessage:
		if event.Via != nil {
			return
		}

		name := r.bridge.memberName(ctx, r.id.Agent.Workspace, event.Author)
		chatBot.say(ctx, r.at, name+" in Lucity: "+event.Text)
	case agentrunner.EventMessage:
		if strings.TrimSpace(event.Text) != "" {
			chatBot.say(ctx, r.at, event.Text)
		}
	case agentrunner.EventToolCall:
		call := *event.ToolCall

		if _, seen := r.calls[call.CallID]; !seen {
			r.steps++
			r.status = chatBot.status(ctx, r.at, r.status, "Working · "+steps(r.steps)+" · "+truncate(call.Title, 160))
		}

		r.calls[call.CallID] = call

		if time.Since(r.typed) > typingEvery {
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
			return
		}

		card, ok := r.cards[approval.ApprovalID]

		if !ok {
			return
		}

		delete(r.cards, approval.ApprovalID)
		chatBot.decided(ctx, r.at, card, request, r.outcome(ctx, approval))
	case agentrunner.EventNotice:
		text := event.Notice.Title

		if event.Notice.Description != "" {
			text += ": " + event.Notice.Description
		}

		chatBot.say(ctx, r.at, text)
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

		if end.Reason == "failed" && end.Error != "" {
			chatBot.say(ctx, r.at, "The turn failed: "+end.Error)
		}

		r.status = ""
		r.steps = 0
	}
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
