package agentrunner

import (
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/zeitlos/lucity/pkg/platform"
)

type Model struct {
	Provider string `json:"provider"`
	API      string `json:"api"`
	Name     string `json:"name"`
}

type Status struct {
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
	LastModel  *Model `json:"lastModel,omitempty"`
	LastEffort string `json:"lastEffort,omitempty"`
}

type Session struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	StartedBy string    `json:"startedBy"`
	StartedAt time.Time `json:"startedAt"`
	Model     Model     `json:"model"`
	Effort    string    `json:"effort"`
	Status    string    `json:"status"`
	Events    []Event   `json:"events,omitempty"`
}

type Event struct {
	ID       string    `json:"id"`
	Time     time.Time `json:"time"`
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	Author   string    `json:"author,omitempty"`
	ToolCall *ToolCall `json:"toolCall,omitempty"`
	Approval *Approval `json:"approval,omitempty"`
	Notice   *Notice   `json:"notice,omitempty"`
	TurnEnd  *TurnEnd  `json:"turnEnd,omitempty"`
}

type Notice struct {
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type ToolCall struct {
	CallID    string `json:"callId"`
	Connector string `json:"connector,omitempty"`
	Tool      string `json:"tool"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	Input     string `json:"input,omitempty"`
	Output    string `json:"output,omitempty"`
}

type Approval struct {
	ApprovalID string `json:"approvalId"`
	CallID     string `json:"callId,omitempty"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	DecidedBy  string `json:"decidedBy,omitempty"`
}

type TurnEnd struct {
	Reason string `json:"reason"`
	Error  string `json:"error,omitempty"`
}

const (
	StatusIdle               = "idle"
	StatusWorking            = "working"
	StatusWaitingForApproval = "waitingForApproval"
	StatusStarting           = "starting"
	StatusFailed             = "failed"

	EventUserMessage = "userMessage"
	EventMessage     = "message"
	EventToolCall    = "toolCall"
	EventApproval    = "approval"
	EventNotice      = "notice"
	EventTurnEnd     = "turnEnd"
)

const (
	maxSessions   = 500
	maxEvents     = 10000
	maxTitleLen   = 200
	maxTextLen    = 256 * 1024
	maxToolIOLen  = 64 * 1024
	maxErrorLen   = 2 * 1024
	maxNameLen    = 256
	maxRefLen     = 128
	maxUserIDLen  = 128
	maxAPIMessage = 300
)

var (
	sessionStatuses  = []string{StatusIdle, StatusWorking, StatusWaitingForApproval}
	runnerStatuses   = []string{StatusIdle, StatusWorking, StatusWaitingForApproval, StatusStarting, StatusFailed}
	efforts          = []string{"low", "medium", "high", "max"}
	modelAPIs        = []string{"anthropic", "openai", "openaiCompatible"}
	toolCallStatuses = []string{"running", "completed", "failed", "denied"}
	approvalStatuses = []string{"pending", "allowed", "denied"}
	noticeSeverities = []string{"info", "warning", "error"}
	turnEndReasons   = []string{"completed", "stopped", "failed"}
)

var approvalIDRe = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,128}$`)

// IsValidApprovalID reports whether id can name an approval in a runner URL.
func IsValidApprovalID(id string) bool {
	return approvalIDRe.MatchString(id)
}

func oneOf(value string, allowed []string) bool {
	for _, a := range allowed {
		if value == a {
			return true
		}
	}

	return false
}

func cut(s string, maxBytes int) string {
	s = strings.ToValidUTF8(s, "�")

	if len(s) <= maxBytes {
		return s
	}

	s = s[:maxBytes]

	for !utf8.ValidString(s) {
		s = s[:len(s)-1]
	}

	return s
}

func isRef(s string, maxLen int) bool {
	if s == "" || len(s) > maxLen || !utf8.ValidString(s) {
		return false
	}

	for _, r := range s {
		if r <= 0x20 || r == 0x7f {
			return false
		}
	}

	return true
}

func sanitizeStatus(status *Status) bool {
	if !oneOf(status.Status, runnerStatuses) {
		return false
	}

	status.Error = cut(status.Error, maxErrorLen)

	if status.LastModel != nil && !sanitizeModel(status.LastModel) {
		status.LastModel = nil
	}

	if !oneOf(status.LastEffort, efforts) {
		status.LastEffort = ""
	}

	return true
}

func sanitizeModel(model *Model) bool {
	return isRef(model.Provider, maxNameLen) && oneOf(model.API, modelAPIs) && isRef(model.Name, maxNameLen)
}

func sanitizeSession(session *Session) bool {
	if !platform.IsValidSessionName(session.ID) || !oneOf(session.Status, sessionStatuses) || !oneOf(session.Effort, efforts) || !sanitizeModel(&session.Model) {
		return false
	}

	session.Title = cut(strings.TrimSpace(session.Title), maxTitleLen)

	if !isRef(session.StartedBy, maxUserIDLen) {
		session.StartedBy = ""
	}

	events := session.Events[:0]

	for _, event := range session.Events {
		if len(events) >= maxEvents {
			break
		}

		if sanitizeEvent(&event) {
			events = append(events, event)
		}
	}

	session.Events = events

	return true
}

func sanitizeEvent(event *Event) bool {
	if !platform.IsValidSessionName(event.ID) {
		return false
	}

	if !isRef(event.Author, maxUserIDLen) {
		event.Author = ""
	}

	switch event.Type {
	case EventUserMessage, EventMessage:
		event.Text = cut(event.Text, maxTextLen)
		event.ToolCall, event.Approval, event.Notice, event.TurnEnd = nil, nil, nil, nil

		return true
	case EventNotice:
		notice := event.Notice

		if notice == nil || strings.TrimSpace(notice.Title) == "" {
			return false
		}

		if !oneOf(notice.Severity, noticeSeverities) {
			notice.Severity = "info"
		}

		notice.Title = cut(notice.Title, maxErrorLen)
		notice.Description = cut(notice.Description, maxErrorLen)
		event.Text, event.ToolCall, event.Approval, event.TurnEnd = "", nil, nil, nil

		return true
	case EventToolCall:
		call := event.ToolCall

		if call == nil || !isRef(call.CallID, maxRefLen) || !isRef(call.Tool, maxNameLen) || !oneOf(call.Status, toolCallStatuses) {
			return false
		}

		if !isRef(call.Connector, maxNameLen) {
			call.Connector = ""
		}

		call.Title = cut(call.Title, maxTitleLen)
		call.Input = cut(call.Input, maxToolIOLen)
		call.Output = cut(call.Output, maxToolIOLen)
		event.Text, event.Approval, event.Notice, event.TurnEnd = "", nil, nil, nil

		return true
	case EventApproval:
		approval := event.Approval

		if approval == nil || !IsValidApprovalID(approval.ApprovalID) || !oneOf(approval.Status, approvalStatuses) {
			return false
		}

		if !isRef(approval.CallID, maxRefLen) {
			approval.CallID = ""
		}

		if !isRef(approval.DecidedBy, maxUserIDLen) {
			approval.DecidedBy = ""
		}

		approval.Title = cut(approval.Title, maxTitleLen)
		event.Text, event.ToolCall, event.Notice, event.TurnEnd = "", nil, nil, nil

		return true
	case EventTurnEnd:
		end := event.TurnEnd

		if end == nil || !oneOf(end.Reason, turnEndReasons) {
			return false
		}

		end.Error = cut(end.Error, maxErrorLen)
		event.Text, event.ToolCall, event.Approval, event.Notice = "", nil, nil, nil

		return true
	default:
		return false
	}
}
