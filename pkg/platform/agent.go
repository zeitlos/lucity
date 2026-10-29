package platform

import (
	"io"
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/99designs/gqlgen/graphql"

	"github.com/zeitlos/lucity/pkg/usererror"
)

var sessionNameRe = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

type ModelProviderID struct {
	Workspace string
	Name      string
}

func ParseModelProviderID(s string) (ModelProviderID, error) {
	workspace, name, ok := strings.Cut(s, "/")

	if !ok || workspace == "" || name == "" || strings.Contains(name, "/") {
		return ModelProviderID{}, usererror.Errorf("invalid model provider id %q", s)
	}

	return ModelProviderID{Workspace: workspace, Name: name}, nil
}

func (m ModelProviderID) LogValue() slog.Value {
	return slog.StringValue(m.String())
}

func (m ModelProviderID) String() string {
	return m.Workspace + "/" + m.Name
}

func (m ModelProviderID) WorkspaceID() string {
	return m.Workspace
}

func (m *ModelProviderID) UnmarshalGQL(val interface{}) error {
	str, ok := val.(string)

	if !ok {
		return usererror.New("ModelProviderID must be a string")
	}

	parsed, err := ParseModelProviderID(str)

	if err != nil {
		return err
	}

	*m = parsed

	return nil
}

func (m ModelProviderID) MarshalGQL(w io.Writer) {
	graphql.MarshalString(m.String()).MarshalGQL(w)
}

type ConnectorID struct {
	Workspace string
	Name      string
}

func ParseConnectorID(s string) (ConnectorID, error) {
	workspace, name, ok := strings.Cut(s, "/")

	if !ok || workspace == "" || name == "" || strings.Contains(name, "/") {
		return ConnectorID{}, usererror.Errorf("invalid connector id %q", s)
	}

	return ConnectorID{Workspace: workspace, Name: name}, nil
}

func (c ConnectorID) LogValue() slog.Value {
	return slog.StringValue(c.String())
}

func (c ConnectorID) String() string {
	return c.Workspace + "/" + c.Name
}

func (c ConnectorID) WorkspaceID() string {
	return c.Workspace
}

func (c *ConnectorID) UnmarshalGQL(val interface{}) error {
	str, ok := val.(string)

	if !ok {
		return usererror.New("ConnectorID must be a string")
	}

	parsed, err := ParseConnectorID(str)

	if err != nil {
		return err
	}

	*c = parsed

	return nil
}

func (c ConnectorID) MarshalGQL(w io.Writer) {
	graphql.MarshalString(c.String()).MarshalGQL(w)
}

type AgentID struct {
	Workspace   string
	Project     string
	Environment string
	Name        string
}

func ParseAgentID(s string) (AgentID, error) {
	parts := strings.SplitN(s, "/", 4)

	if len(parts) != 4 || slices.Contains(parts, "") || strings.Contains(parts[3], "/") {
		return AgentID{}, usererror.Errorf("invalid agent id %q", s)
	}

	return AgentID{
		Workspace: parts[0], Project: parts[1],
		Environment: parts[2], Name: parts[3],
	}, nil
}

func (a AgentID) LogValue() slog.Value {
	return slog.StringValue(a.String())
}

func (a AgentID) String() string {
	return a.Workspace + "/" + a.Project + "/" + a.Environment + "/" + a.Name
}

func (a AgentID) EnvironmentID() EnvironmentID {
	return EnvironmentID{Workspace: a.Workspace, Project: a.Project, Name: a.Environment}
}

func (a AgentID) Namespace() string {
	return a.EnvironmentID().Namespace()
}

func (a AgentID) WorkspaceID() string {
	return a.Workspace
}

func (a *AgentID) UnmarshalGQL(val interface{}) error {
	str, ok := val.(string)

	if !ok {
		return usererror.New("AgentID must be a string")
	}

	parsed, err := ParseAgentID(str)

	if err != nil {
		return err
	}

	*a = parsed

	return nil
}

func (a AgentID) MarshalGQL(w io.Writer) {
	graphql.MarshalString(a.String()).MarshalGQL(w)
}

// AgentSessionID names one session of an agent. Session is the id the
// agent's runner gave the session.
type AgentSessionID struct {
	Agent   AgentID
	Session string
}

func ParseAgentSessionID(s string) (AgentSessionID, error) {
	agent, session, ok := cutLast(s, "/")

	if !ok || !IsValidSessionName(session) {
		return AgentSessionID{}, usererror.Errorf("invalid agent session id %q", s)
	}

	agentID, err := ParseAgentID(agent)

	if err != nil {
		return AgentSessionID{}, usererror.Errorf("invalid agent session id %q", s)
	}

	return AgentSessionID{Agent: agentID, Session: session}, nil
}

// IsValidSessionName reports whether name is a session id a runner may hand
// out: short, and safe as a URL path segment.
func IsValidSessionName(name string) bool {
	return sessionNameRe.MatchString(name)
}

func cutLast(s, sep string) (before, after string, found bool) {
	i := strings.LastIndex(s, sep)

	if i < 0 {
		return s, "", false
	}

	return s[:i], s[i+len(sep):], true
}

func (a AgentSessionID) LogValue() slog.Value {
	return slog.StringValue(a.String())
}

func (a AgentSessionID) String() string {
	return a.Agent.String() + "/" + a.Session
}

func (a AgentSessionID) WorkspaceID() string {
	return a.Agent.Workspace
}

func (a *AgentSessionID) UnmarshalGQL(val interface{}) error {
	str, ok := val.(string)

	if !ok {
		return usererror.New("AgentSessionID must be a string")
	}

	parsed, err := ParseAgentSessionID(str)

	if err != nil {
		return err
	}

	*a = parsed

	return nil
}

func (a AgentSessionID) MarshalGQL(w io.Writer) {
	graphql.MarshalString(a.String()).MarshalGQL(w)
}

var (
	_ WorkspaceScoped     = ModelProviderID{}
	_ WorkspaceScoped     = ConnectorID{}
	_ WorkspaceScoped     = AgentID{}
	_ WorkspaceScoped     = AgentSessionID{}
	_ graphql.Marshaler   = ModelProviderID{}
	_ graphql.Marshaler   = ConnectorID{}
	_ graphql.Marshaler   = AgentID{}
	_ graphql.Marshaler   = AgentSessionID{}
	_ graphql.Unmarshaler = (*ModelProviderID)(nil)
	_ graphql.Unmarshaler = (*ConnectorID)(nil)
	_ graphql.Unmarshaler = (*AgentID)(nil)
	_ graphql.Unmarshaler = (*AgentSessionID)(nil)
)
