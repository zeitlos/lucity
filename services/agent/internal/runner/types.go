package runner

import (
	"errors"
	"time"

	"github.com/zeitlos/lucity/services/agent/internal/harness"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrBusy        = errors.New("the session is already working on a message")
	ErrUnavailable = errors.New("the harness is not running")
	ErrAnswered    = errors.New("the approval was already answered")
)

type InvalidError struct {
	Err error
}

func (e InvalidError) Error() string {
	return e.Err.Error()
}

func (e InvalidError) Unwrap() error {
	return e.Err
}

const (
	StatusStarting           = "starting"
	StatusIdle               = "idle"
	StatusWorking            = "working"
	StatusWaitingForApproval = "waitingForApproval"
	StatusFailed             = "failed"
)

const (
	EventUserMessage = "userMessage"
	EventMessage     = "message"
	EventToolCall    = "toolCall"
	EventApproval    = "approval"
	EventTurnEnd     = "turnEnd"
)

const (
	ToolRunning   = "running"
	ToolCompleted = "completed"
	ToolFailed    = "failed"
	ToolDenied    = "denied"
)

const (
	ApprovalPending = "pending"
	ApprovalAllowed = "allowed"
	ApprovalDenied  = "denied"
)

const (
	TurnCompleted = "completed"
	TurnStopped   = "stopped"
	TurnFailed    = "failed"
)

const (
	DecisionAllowOnce = "allow_once"
	DecisionDeny      = "deny"
)

const (
	maxMessage = 256 << 10
	maxTool    = 64 << 10
)

var efforts = []string{"low", "medium", "high", "max"}

type Model = harness.Model

type Session struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	StartedBy        string    `json:"startedBy,omitempty"`
	StartedAt        time.Time `json:"startedAt"`
	Model            Model     `json:"model"`
	Effort           string    `json:"effort,omitempty"`
	HarnessSessionID string    `json:"harnessSessionId,omitempty"`
}

type SessionStatus struct {
	Session
	Status string `json:"status"`
}

type SessionDetail struct {
	SessionStatus
	Events []Event `json:"events"`
}

type Status struct {
	Harness    string `json:"harness"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
	LastModel  *Model `json:"lastModel,omitempty"`
	LastEffort string `json:"lastEffort,omitempty"`
}

type Event struct {
	ID       int64     `json:"id,string"`
	Time     time.Time `json:"time"`
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	Author   string    `json:"author,omitempty"`
	ToolCall *ToolCall `json:"toolCall,omitempty"`
	Approval *Approval `json:"approval,omitempty"`
	TurnEnd  *TurnEnd  `json:"turnEnd,omitempty"`
}

type ToolCall struct {
	CallID    string `json:"callId"`
	Connector string `json:"connector,omitempty"`
	Tool      string `json:"tool,omitempty"`
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
