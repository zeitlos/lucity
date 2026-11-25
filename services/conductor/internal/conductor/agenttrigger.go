package conductor

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"slices"
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

// AgentTrigger starts sessions of an agent on its own, on a schedule or when
// its webhook is called.
type AgentTrigger struct {
	ID           AgentTriggerID
	Name         string
	Source       values.TriggerSource
	Schedule     string
	TimeZone     string
	Key          string
	Instructions string
	Enabled      bool
}

// AgentTriggerSpec is what a member sets on a trigger. Schedule and TimeZone
// are for a schedule, Key for a webhook.
type AgentTriggerSpec struct {
	Name         string
	Source       values.TriggerSource
	Schedule     string
	TimeZone     string
	Key          string
	Instructions string
}

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
