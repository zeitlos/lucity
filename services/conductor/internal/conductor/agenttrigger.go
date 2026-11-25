package conductor

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/zeitlos/lucity/pkg/agentrunner"
	"github.com/zeitlos/lucity/pkg/agenttrigger"
	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/pkg/usererror"
)

type AgentTriggerID = platform.AgentTriggerID

// AgentTrigger starts sessions of an agent on its own, on a schedule, when
// its webhook is called or on an event in a GitHub repository.
type AgentTrigger struct {
	ID           AgentTriggerID
	Name         string
	Source       values.TriggerSource
	Schedule     string
	TimeZone     string
	Key          string
	Repository   string
	Events       []string
	Branches     []string
	OnlyFailures bool
	Instructions string
	Enabled      bool
}

// AgentTriggerSpec is what a member sets on a trigger. Schedule and TimeZone
// are for a schedule, Key for a webhook, Repository, Events, Branches and
// OnlyFailures for GitHub.
type AgentTriggerSpec struct {
	Name         string
	Source       values.TriggerSource
	Schedule     string
	TimeZone     string
	Key          string
	Repository   string
	Events       []string
	Branches     []string
	OnlyFailures bool
	Instructions string
}

// GitHubTriggerEvent is an event from a GitHub repository, as the GitHub App
// webhook delivered it.
type GitHubTriggerEvent struct {
	Type           string
	Action         string
	Repository     string
	InstallationID int64
	HeadRepository string
	Branch         string
	Commit         string
	PullRequest    int
	Issue          int
	Title          string
	Workflow       string
	Conclusion     string
	URL            string
	Sender         string
	Association    string
	Comment        string
	Tag            string
}

// GitHubTriggerEvents are the events a GitHub trigger can start on.
var GitHubTriggerEvents = []string{
	"pull_request.opened",
	"pull_request.reopened",
	"pull_request.synchronize",
	"pull_request.closed",
	"workflow_run.completed",
	"issues.opened",
	"issue_comment.created",
	"release.published",
}

const (
	maxTriggerBranches = 16
	maxTriggerComment  = 4000
)

var (
	trustedAssociations = []string{"OWNER", "MEMBER", "COLLABORATOR"}
	failedConclusions   = []string{"failure", "timed_out", "startup_failure"}
)

var (
	// ErrTriggerNotFound means no enabled webhook trigger has the id.
	ErrTriggerNotFound = errors.New("trigger not found")
	// ErrTriggerUnauthorized means the webhook's token didn't match.
	ErrTriggerUnauthorized = errors.New("wrong token")
	// ErrTriggerInvalid means the webhook's body isn't a JSON object.
	ErrTriggerInvalid = errors.New("the body must be a JSON object")
	// ErrTriggerBusy means the session the event belongs to is working.
	ErrTriggerBusy = errors.New("the session is busy")
)

func toAgentTriggers(id AgentID, agent values.Agent) []AgentTrigger {
	triggers := make([]AgentTrigger, 0, len(agent.Triggers))

	for name, trigger := range agent.Triggers {
		triggers = append(triggers, toAgentTrigger(AgentTriggerID{Agent: id, Trigger: name}, trigger))
	}

	slices.SortFunc(triggers, func(a, b AgentTrigger) int { return strings.Compare(a.Name, b.Name) })

	return triggers
}

func toAgentTrigger(id AgentTriggerID, trigger values.AgentTrigger) AgentTrigger {
	return AgentTrigger{
		ID:           id,
		Name:         trigger.Name,
		Source:       trigger.Source,
		Schedule:     trigger.Schedule,
		TimeZone:     trigger.TimeZone,
		Key:          trigger.Key,
		Repository:   trigger.Repository,
		Events:       trigger.Events,
		Branches:     trigger.Branches,
		OnlyFailures: trigger.OnlyFailures,
		Instructions: trigger.Instructions,
		Enabled:      trigger.Enabled,
	}
}

// CreateAgentTrigger adds a trigger to an agent. A webhook trigger comes with
// its token, which is only returned here.
func (c *Client) CreateAgentTrigger(ctx context.Context, agent AgentID, spec AgentTriggerSpec) (*AgentTrigger, string, error) {
	if _, _, err := c.loadAgent(ctx, agent); err != nil {
		return nil, "", err
	}

	trigger, err := triggerValues(spec)

	if err != nil {
		return nil, "", err
	}

	token := ""

	if trigger.Source == values.TriggerWebhook {
		token, trigger.TokenHash = newTriggerToken()
	}

	if err := c.authorizeTriggerRepository(ctx, &trigger); err != nil {
		return nil, "", err
	}

	trigger.Enabled = true
	id := AgentTriggerID{Agent: agent, Trigger: randCrockford32(12)}

	if _, err := c.deployer.Agents().SetTrigger(ctx, id, trigger); err != nil {
		return nil, "", err
	}

	return new(toAgentTrigger(id, trigger)), token, nil
}

// UpdateAgentTrigger changes a trigger. Its source can't change.
func (c *Client) UpdateAgentTrigger(ctx context.Context, id AgentTriggerID, spec AgentTriggerSpec) (*AgentTrigger, error) {
	current, err := c.loadAgentTrigger(ctx, id)

	if err != nil {
		return nil, err
	}

	if spec.Source != current.Source {
		return nil, usererror.New("a trigger's source can't change, create a new trigger instead")
	}

	trigger, err := triggerValues(spec)

	if err != nil {
		return nil, err
	}

	trigger.TokenHash = current.TokenHash
	trigger.Enabled = current.Enabled

	if err := c.authorizeTriggerRepository(ctx, &trigger); err != nil {
		return nil, err
	}

	return c.setAgentTrigger(ctx, id, trigger)
}

// SetAgentTriggerEnabled turns a trigger on or off.
func (c *Client) SetAgentTriggerEnabled(ctx context.Context, id AgentTriggerID, enabled bool) (*AgentTrigger, error) {
	trigger, err := c.loadAgentTrigger(ctx, id)

	if err != nil {
		return nil, err
	}

	trigger.Enabled = enabled

	return c.setAgentTrigger(ctx, id, trigger)
}

// ReplaceAgentTriggerToken gives a webhook trigger a new token. The old one
// stops working.
func (c *Client) ReplaceAgentTriggerToken(ctx context.Context, id AgentTriggerID) (*AgentTrigger, string, error) {
	trigger, err := c.loadAgentTrigger(ctx, id)

	if err != nil {
		return nil, "", err
	}

	if trigger.Source != values.TriggerWebhook {
		return nil, "", usererror.New("only a webhook trigger has a token")
	}

	token, hash := newTriggerToken()
	trigger.TokenHash = hash

	result, err := c.setAgentTrigger(ctx, id, trigger)

	if err != nil {
		return nil, "", err
	}

	return result, token, nil
}

// DeleteAgentTrigger removes a trigger. The sessions it started stay.
func (c *Client) DeleteAgentTrigger(ctx context.Context, id AgentTriggerID) (*Agent, error) {
	if _, err := c.loadAgentTrigger(ctx, id); err != nil {
		return nil, err
	}

	if _, err := c.deployer.Agents().DeleteTrigger(ctx, id); err != nil {
		return nil, err
	}

	return c.Agent(ctx, id.Agent)
}

// AgentTriggerLastRun is when the trigger last started a session.
func (c *Client) AgentTriggerLastRun(ctx context.Context, id AgentTriggerID) (*time.Time, error) {
	sessions, err := c.agentRunner.Sessions(ctx, id.Agent)

	if err != nil {
		return nil, err
	}

	last := agenttrigger.LastRun(sessions, id.Trigger)

	if last.IsZero() {
		return nil, nil
	}

	return &last, nil
}

// AgentTriggerURL is the address a webhook trigger is called at, on the
// dashboard's host.
func (c *Client) AgentTriggerURL(id AgentTriggerID) (string, error) {
	base, err := url.Parse(c.config.DashboardURL)

	if err != nil {
		return "", fmt.Errorf("parsing the dashboard url: %w", err)
	}

	return base.Scheme + "://" + base.Host + "/webhooks/triggers/" + id.Trigger, nil
}

// FireWebhook hands a webhook call to its trigger. The token is checked
// against the trigger's hash, and the event continues the session of the
// same key when the trigger has one.
func (c *Client) FireWebhook(ctx context.Context, triggerID, token string, body []byte) error {
	trigger, err := c.agentTriggers.Trigger(ctx, triggerID)

	if errors.Is(err, agenttrigger.ErrNotFound) {
		return ErrTriggerNotFound
	}

	if err != nil {
		return err
	}

	if trigger.Source != values.TriggerWebhook || !trigger.Enabled {
		return ErrTriggerNotFound
	}

	sum := sha256.Sum256([]byte(token))

	if subtle.ConstantTimeCompare([]byte("sha256:"+hex.EncodeToString(sum[:])), []byte(trigger.TokenHash)) != 1 {
		return ErrTriggerUnauthorized
	}

	key, err := triggerKey(body, trigger.Key)

	if err != nil {
		return err
	}

	err = c.agentTriggers.Fire(ctx, *trigger, key, string(body))

	if errors.Is(err, agentrunner.ErrConflict) {
		return ErrTriggerBusy
	}

	return err
}

func triggerKey(body []byte, field string) (string, error) {
	var fields map[string]json.RawMessage

	if err := json.Unmarshal(body, &fields); err != nil {
		return "", ErrTriggerInvalid
	}

	if field == "" {
		return "", nil
	}

	var key string

	if err := json.Unmarshal(fields[field], &key); err != nil || key == "" {
		return "", fmt.Errorf("%w with a string field %q", ErrTriggerInvalid, field)
	}

	return key, nil
}

func (c *Client) loadAgentTrigger(ctx context.Context, id AgentTriggerID) (values.AgentTrigger, error) {
	agent, _, err := c.loadAgent(ctx, id.Agent)

	if err != nil {
		return values.AgentTrigger{}, err
	}

	trigger, ok := agent.Triggers[id.Trigger]

	if !ok {
		return values.AgentTrigger{}, usererror.Errorf("trigger %q not found", id.Trigger)
	}

	return trigger, nil
}

func (c *Client) setAgentTrigger(ctx context.Context, id AgentTriggerID, trigger values.AgentTrigger) (*AgentTrigger, error) {
	if _, err := c.deployer.Agents().SetTrigger(ctx, id, trigger); err != nil {
		return nil, err
	}

	return new(toAgentTrigger(id, trigger)), nil
}

func triggerValues(spec AgentTriggerSpec) (values.AgentTrigger, error) {
	trigger := values.AgentTrigger{
		Name:         strings.TrimSpace(spec.Name),
		Source:       spec.Source,
		Instructions: spec.Instructions,
	}

	switch spec.Source {
	case values.TriggerSchedule:
		if _, err := cron.ParseStandard(spec.Schedule); err != nil {
			return values.AgentTrigger{}, usererror.Errorf("invalid schedule: %v", err)
		}

		if _, err := time.LoadLocation(spec.TimeZone); err != nil || spec.TimeZone == "" {
			return values.AgentTrigger{}, usererror.Errorf("unknown time zone %q", spec.TimeZone)
		}

		trigger.Schedule = strings.TrimSpace(spec.Schedule)
		trigger.TimeZone = spec.TimeZone
	case values.TriggerWebhook:
		trigger.Key = strings.TrimSpace(spec.Key)
	case values.TriggerGitHub:
		repository := strings.TrimSpace(spec.Repository)

		if _, _, err := validateRepository(repository); err != nil {
			return values.AgentTrigger{}, err
		}

		events, err := triggerEvents(spec.Events)

		if err != nil {
			return values.AgentTrigger{}, err
		}

		branches, err := triggerBranches(spec.Branches)

		if err != nil {
			return values.AgentTrigger{}, err
		}

		trigger.Repository = repository
		trigger.Events = events
		trigger.Branches = branches
		trigger.OnlyFailures = spec.OnlyFailures && slices.Contains(events, "workflow_run.completed")
	default:
		return values.AgentTrigger{}, usererror.Errorf("unsupported source %q", spec.Source)
	}

	return trigger, nil
}

func newTriggerToken() (string, string) {
	token := "ltk_" + randCrockford32(40)
	sum := sha256.Sum256([]byte(token))

	return token, fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}

func triggerEvents(events []string) ([]string, error) {
	result := make([]string, 0, len(events))

	for _, event := range events {
		if !slices.Contains(GitHubTriggerEvents, event) {
			return nil, usererror.Errorf("unsupported event %q", event)
		}

		if !slices.Contains(result, event) {
			result = append(result, event)
		}
	}

	if len(result) == 0 {
		return nil, usererror.New("pick at least one event")
	}

	return result, nil
}

func triggerBranches(branches []string) ([]string, error) {
	result := make([]string, 0, len(branches))

	for _, branch := range branches {
		branch = strings.TrimSpace(branch)

		if branch == "" || slices.Contains(result, branch) {
			continue
		}

		if _, err := path.Match(branch, ""); err != nil {
			return nil, usererror.Errorf("invalid branch pattern %q", branch)
		}

		result = append(result, branch)
	}

	if len(result) > maxTriggerBranches {
		return nil, usererror.Errorf("at most %d branch patterns", maxTriggerBranches)
	}

	return result, nil
}

// authorizeTriggerRepository checks that the member can read a GitHub
// trigger's repository and records the App installation that covers it,
// which every event must then come from.
func (c *Client) authorizeTriggerRepository(ctx context.Context, trigger *values.AgentTrigger) error {
	if trigger.Source != values.TriggerGitHub {
		return nil
	}

	installationID, err := c.authorizeRepository(ctx, trigger.Repository)

	if err != nil {
		return err
	}

	trigger.InstallationID = strconv.FormatInt(installationID, 10)

	return nil
}

// FireGitHub hands a GitHub event to every enabled GitHub trigger it matches:
// same repository and installation, a listed event and a matching branch.
// Events from forks never fire, and issues and comments only from the
// repository's owners, members and collaborators. A comment fires only when
// it mentions the agent, so the agent's own comments don't start it again.
// Events of one pull request or issue continue one session. It returns how
// many triggers took the event.
func (c *Client) FireGitHub(ctx context.Context, event GitHubTriggerEvent) (int, error) {
	if !strings.EqualFold(event.HeadRepository, event.Repository) {
		return 0, nil
	}

	if (event.Type == "issues" || event.Type == "issue_comment") && !slices.Contains(trustedAssociations, event.Association) {
		return 0, nil
	}

	triggers, err := c.agentTriggers.Triggers(ctx)

	if err != nil {
		return 0, err
	}

	name := event.Type + "." + event.Action
	installationID := strconv.FormatInt(event.InstallationID, 10)
	body, err := gitHubTriggerBody(name, event)

	if err != nil {
		return 0, err
	}

	key := ""

	switch {
	case event.PullRequest > 0:
		key = "pr:" + strconv.Itoa(event.PullRequest)
	case event.Issue > 0:
		key = "issue:" + strconv.Itoa(event.Issue)
	}

	fired := 0
	suspended := map[string]bool{}

	for _, trigger := range triggers {
		if trigger.Source != values.TriggerGitHub || !trigger.Enabled {
			continue
		}

		if !strings.EqualFold(trigger.Repository, event.Repository) || trigger.InstallationID != installationID {
			continue
		}

		if !slices.Contains(trigger.Events, name) || !matchesBranch(trigger.Branches, event.Branch) {
			continue
		}

		if trigger.OnlyFailures && event.Type == "workflow_run" && !slices.Contains(failedConclusions, event.Conclusion) {
			continue
		}

		if event.Type == "issue_comment" && !mentions(event.Comment, trigger.ID.Agent.Name) {
			continue
		}

		workspace := trigger.ID.Agent.Workspace
		workspaceSuspended, checked := suspended[workspace]

		if !checked {
			workspaceSuspended = c.workspaceSuspended(ctx, workspace)
			suspended[workspace] = workspaceSuspended
		}

		if workspaceSuspended {
			continue
		}

		err := c.agentTriggers.Fire(ctx, trigger, key, body)

		if errors.Is(err, agentrunner.ErrConflict) {
			slog.InfoContext(ctx, "github trigger skipped, the session is busy", "trigger", trigger.ID, "key", key)
			continue
		}

		if err != nil {
			slog.WarnContext(ctx, "github trigger failed", "trigger", trigger.ID, "error", err)
			continue
		}

		fired++
	}

	return fired, nil
}

func (c *Client) workspaceSuspended(ctx context.Context, workspace string) bool {
	details, err := c.Workspace(ctx, workspace)

	if err != nil {
		slog.WarnContext(ctx, "github trigger: failed to read workspace, firing anyway", "workspace", workspace, "error", err)
		return false
	}

	return details.Suspended
}

func matchesBranch(patterns []string, branch string) bool {
	if len(patterns) == 0 || branch == "" {
		return true
	}

	for _, pattern := range patterns {
		if ok, _ := path.Match(pattern, branch); ok {
			return true
		}
	}

	return false
}

func gitHubTriggerBody(name string, event GitHubTriggerEvent) (string, error) {
	body, err := json.MarshalIndent(struct {
		Event       string `json:"event"`
		Repository  string `json:"repository"`
		PullRequest int    `json:"pullRequest,omitempty"`
		Issue       int    `json:"issue,omitempty"`
		Title       string `json:"title,omitempty"`
		Tag         string `json:"tag,omitempty"`
		Branch      string `json:"branch,omitempty"`
		Commit      string `json:"commit,omitempty"`
		Workflow    string `json:"workflow,omitempty"`
		Conclusion  string `json:"conclusion,omitempty"`
		URL         string `json:"url,omitempty"`
		Sender      string `json:"sender,omitempty"`
		Comment     string `json:"comment,omitempty"`
	}{
		Event:       name,
		Repository:  event.Repository,
		PullRequest: event.PullRequest,
		Issue:       event.Issue,
		Title:       event.Title,
		Tag:         event.Tag,
		Branch:      event.Branch,
		Commit:      event.Commit,
		Workflow:    event.Workflow,
		Conclusion:  event.Conclusion,
		URL:         event.URL,
		Sender:      event.Sender,
		Comment:     truncate(event.Comment, maxTriggerComment),
	}, "", "  ")

	if err != nil {
		return "", err
	}

	return string(body), nil
}

// mentions reports whether text mentions @name as a whole handle.
func mentions(text, name string) bool {
	pattern := `(?i)(^|[^A-Za-z0-9_-])@` + regexp.QuoteMeta(name) + `($|[^A-Za-z0-9_-])`

	return regexp.MustCompile(pattern).MatchString(text)
}

func truncate(text string, limit int) string {
	runes := []rune(text)

	if len(runes) <= limit {
		return text
	}

	return string(runes[:limit]) + "…"
}
