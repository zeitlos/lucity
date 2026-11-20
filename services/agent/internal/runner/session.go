package runner

import (
	"cmp"
	"context"
	"encoding/json"
	"slices"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/zeitlos/lucity/pkg/acp"
)

type session struct {
	mu        sync.Mutex
	record    Session
	events    []Event
	nextID    int64
	attached  *acp.Client
	replaying bool
	options   []acp.ConfigOption
	turn      *turn
	approvals map[string]chan decision
	tools     map[string]*toolState
	message   strings.Builder
	waiters   map[chan struct{}]struct{}
}

type turn struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type decision struct {
	outcome   string
	decidedBy string
	recorded  chan struct{}
}

type toolState struct {
	call   ToolCall
	denied bool
}

func newSession(record Session, events []Event) *session {
	s := &session{
		record:    record,
		events:    events,
		approvals: map[string]chan decision{},
		tools:     map[string]*toolState{},
		waiters:   map[chan struct{}]struct{}{},
	}
	if len(events) > 0 {
		s.nextID = events[len(events)-1].ID
	}
	return s
}

func (s *session) status() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.statusLocked()
}

func (s *session) statusLocked() string {
	switch {
	case len(s.approvals) > 0:
		return StatusWaitingForApproval
	case s.turn != nil:
		return StatusWorking
	default:
		return StatusIdle
	}
}

func (s *session) snapshot() SessionStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return SessionStatus{Session: s.record, Status: s.statusLocked()}
}

func (s *session) detail() SessionDetail {
	s.mu.Lock()
	defer s.mu.Unlock()
	return SessionDetail{
		SessionStatus: SessionStatus{Session: s.record, Status: s.statusLocked()},
		Events:        append([]Event{}, s.events...),
	}
}

func (s *session) eventsAfter(after int64) []Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	position, _ := slices.BinarySearchFunc(s.events, after+1, func(event Event, id int64) int {
		return cmp.Compare(event.ID, id)
	})
	return slices.Clone(s.events[position:])
}

func (s *session) subscribe() (<-chan struct{}, func()) {
	waiter := make(chan struct{}, 1)
	s.mu.Lock()
	s.waiters[waiter] = struct{}{}
	s.mu.Unlock()
	return waiter, func() {
		s.mu.Lock()
		delete(s.waiters, waiter)
		s.mu.Unlock()
	}
}

func (s *session) appendMessage(text string) {
	if s.message.Len()+len(text) > maxMessage {
		text = truncate(text, maxMessage-s.message.Len())
	}
	s.message.WriteString(text)
}

func (s *session) toolCall(update acp.Update, connectors []string) (ToolCall, bool) {
	state, ok := s.tools[update.ToolCallID]
	if !ok {
		state = &toolState{call: ToolCall{CallID: update.ToolCallID, Status: ToolRunning}}
		s.tools[update.ToolCallID] = state
	}
	previous := state.call
	call := &state.call
	if update.Title != "" {
		call.Title = update.Title
		call.Connector, call.Tool = splitTool(update.Title, connectors)
	}
	if input := compact(update.RawInput); input != "" {
		call.Input = truncate(input, maxTool)
	}
	if output := toolOutput(update.Content); output != "" {
		call.Output = truncate(output, maxTool)
	}
	switch update.Status {
	case "pending", "in_progress":
		call.Status = ToolRunning
	case "completed":
		call.Status = ToolCompleted
	case "failed":
		call.Status = ToolFailed
		if state.denied {
			call.Status = ToolDenied
		}
	}
	changed := !ok || call.Status != previous.Status || call.Title != previous.Title
	return *call, changed
}

func (s *session) mergeOptions(options []acp.ConfigOption) {
	for _, option := range options {
		position := slices.IndexFunc(s.options, func(existing acp.ConfigOption) bool { return existing.ID == option.ID })
		if position < 0 {
			s.options = append(s.options, option)
			continue
		}
		s.options[position] = option
	}
}

func splitTool(title string, connectors []string) (string, string) {
	for _, name := range connectors {
		for _, prefix := range []string{"mcp__" + name + "__", "mcp." + name + ".", name + "_"} {
			if tool, ok := strings.CutPrefix(title, prefix); ok && tool != "" {
				return name, tool
			}
		}
	}
	return "", ""
}

func compact(raw json.RawMessage) string {
	text := strings.TrimSpace(string(raw))
	if text == "" || text == "null" || text == "{}" {
		return ""
	}
	return text
}

func toolOutput(raw json.RawMessage) string {
	var blocks []struct {
		Type    string `json:"type"`
		Content struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if json.Unmarshal(raw, &blocks) != nil {
		return ""
	}
	var texts []string
	for _, block := range blocks {
		if block.Type == "content" && block.Content.Type == "text" && block.Content.Text != "" {
			texts = append(texts, block.Content.Text)
		}
	}
	return strings.Join(texts, "\n")
}

func truncate(text string, size int) string {
	if size <= 0 {
		return ""
	}
	if len(text) <= size {
		return text
	}
	cut := size
	for cut > 0 && !utf8.RuneStart(text[cut]) {
		cut--
	}
	return text[:cut]
}
