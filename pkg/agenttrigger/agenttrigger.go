// Package agenttrigger finds the triggers of agents and fires them. Triggers
// live in the environment release values; the release renders each into a
// ConfigMap labelled with its id, which is where they are discovered.
package agenttrigger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/zeitlos/lucity/pkg/agentrunner"
	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/pkg/platform"
)

var (
	// ErrNotFound means no trigger has the id.
	ErrNotFound = errors.New("trigger not found")
	// ErrNoModel means the agent never had a session and its workspace has
	// no model provider its harness can use.
	ErrNoModel = errors.New("the workspace has no model provider this agent's harness can use")
)

type Trigger struct {
	ID platform.AgentTriggerID
	values.AgentTrigger
}

// Workspaces reads workspace releases, where an agent's chat is set.
type Workspaces interface {
	Workspace(ctx context.Context, workspace string) (*values.Workspace, error)
}

type Client struct {
	kubernetes kubernetes.Interface
	runner     *agentrunner.Client
	workspaces Workspaces
}

func New(kubernetes kubernetes.Interface, runner *agentrunner.Client, workspaces Workspaces) *Client {
	return &Client{kubernetes: kubernetes, runner: runner, workspaces: workspaces}
}

// Triggers returns the triggers of every agent.
func (c *Client) Triggers(ctx context.Context) ([]Trigger, error) {
	return c.list(ctx, labels.AgentTrigger)
}

// Trigger returns the trigger with the id, which is unique across agents.
func (c *Client) Trigger(ctx context.Context, id string) (*Trigger, error) {
	if !platform.IsValidTriggerName(id) {
		return nil, ErrNotFound
	}

	triggers, err := c.list(ctx, labels.AgentTrigger+"="+id)

	if err != nil {
		return nil, err
	}

	if len(triggers) == 0 {
		return nil, ErrNotFound
	}

	return &triggers[0], nil
}

func (c *Client) list(ctx context.Context, selector string) ([]Trigger, error) {
	configMaps, err := c.kubernetes.CoreV1().ConfigMaps(meta.NamespaceAll).List(ctx, meta.ListOptions{LabelSelector: selector})

	if err != nil {
		return nil, fmt.Errorf("listing triggers: %w", err)
	}

	triggers := make([]Trigger, 0, len(configMaps.Items))

	for _, configMap := range configMaps.Items {
		if trigger, ok := fromConfigMap(configMap); ok {
			triggers = append(triggers, trigger)
		}
	}

	return triggers, nil
}

func fromConfigMap(configMap core.ConfigMap) (Trigger, bool) {
	agent := platform.AgentID{
		Workspace:   configMap.Labels[labels.Workspace],
		Project:     configMap.Labels[labels.Project],
		Environment: configMap.Labels[labels.Environment],
		Name:        configMap.Labels[labels.Agent],
	}

	id := configMap.Labels[labels.AgentTrigger]

	if agent.Namespace() != configMap.Namespace {
		return Trigger{}, false
	}

	trigger := Trigger{ID: platform.AgentTriggerID{Agent: agent, Trigger: id}}

	if err := json.Unmarshal([]byte(configMap.Data["trigger.json"]), &trigger.AgentTrigger); err != nil {
		return Trigger{}, false
	}

	return trigger, true
}

// Fire hands an event to the trigger's agent. With a key, it continues the
// newest session the trigger started with that key. Otherwise, and when there
// is none, it starts a session, mirrored to the agent's chat if it has one.
func (c *Client) Fire(ctx context.Context, trigger Trigger, key, event string) error {
	agent := trigger.ID.Agent
	actor := agentrunner.TriggerActor(trigger.ID.Trigger)
	message := Message(trigger.Instructions, event)

	if key != "" {
		sessions, err := c.runner.Sessions(ctx, agent)

		if err != nil {
			return err
		}

		for _, session := range sessions {
			if session.StartedBy == actor && session.Key == key {
				_, err := c.runner.SendMessage(ctx, platform.AgentSessionID{Agent: agent, Session: session.ID}, message, actor, nil)

				return err
			}
		}
	}

	ws, err := c.workspaces.Workspace(ctx, agent.Workspace)

	if err != nil {
		return err
	}

	model, effort, err := c.defaults(ctx, agent, ws)

	if err != nil {
		return err
	}

	request := agentrunner.StartRequest{
		Message:   message,
		Model:     model,
		Effort:    effort,
		StartedBy: actor,
		Title:     trigger.Name,
		Key:       key,
	}

	if chat := ws.Agents[values.AgentGrantKey(agent)].Chat; chat != nil {
		request.Chat = &agentrunner.Chat{ChatApp: chat.ChatApp, Channel: chat.Channel}
	}

	_, err = c.runner.StartSession(ctx, agent, request)

	return err
}

// defaults are the model and effort a trigger's session starts with. Either
// is left empty when the agent has one it used last, which the runner then
// takes. An agent that never had a session starts like a session started in
// the dashboard: with the workspace's default model for its harness, at the
// default effort.
func (c *Client) defaults(ctx context.Context, agent platform.AgentID, ws *values.Workspace) (*agentrunner.Model, string, error) {
	status, err := c.runner.Status(ctx, agent)

	if err != nil {
		return nil, "", err
	}

	effort := ""

	if status.LastEffort == "" {
		effort = agentrunner.DefaultEffort
	}

	if status.LastModel != nil {
		return nil, effort, nil
	}

	name, provider, ok := values.DefaultModel(ws, values.Harness(status.Harness))

	if !ok {
		return nil, "", ErrNoModel
	}

	return &agentrunner.Model{Provider: name, API: string(provider.API), Name: provider.Models[0]}, effort, nil
}

// Message is what an agent reads when a trigger fires: the trigger's
// instructions, then the event as fenced data.
func Message(instructions, event string) string {
	fence := "```"

	for strings.Contains(event, fence) {
		fence += "`"
	}

	return strings.TrimSpace(instructions) + "\n\nThe event, as data:\n" + fence + "\n" + event + "\n" + fence
}

// LastRun is when the trigger last started a session, or the zero time.
func LastRun(sessions []agentrunner.Session, trigger string) time.Time {
	actor := agentrunner.TriggerActor(trigger)

	var last time.Time

	for _, session := range sessions {
		if session.StartedBy == actor && session.StartedAt.After(last) {
			last = session.StartedAt
		}
	}

	return last
}
