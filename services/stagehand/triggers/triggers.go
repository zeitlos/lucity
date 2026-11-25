// Package triggers runs the schedule triggers of agents. A tick fires when the
// trigger hasn't started a session since, which the agent's own session list
// answers, so restarts never fire a tick twice. Ticks missed while stagehand
// was down are skipped.
package triggers

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/zeitlos/lucity/pkg/agentrunner"
	"github.com/zeitlos/lucity/pkg/agenttrigger"
	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/metrics"
)

type Worker struct {
	triggers *agenttrigger.Client
	runner   *agentrunner.Client
	loops    *metrics.Loops
	interval time.Duration
}

func New(triggers *agenttrigger.Client, runner *agentrunner.Client, loops *metrics.Loops, interval time.Duration) *Worker {
	return &Worker{triggers: triggers, runner: runner, loops: loops, interval: interval}
}

func (w *Worker) Run(ctx context.Context) {
	for {
		start := time.Now()
		w.loops.Observe("agent_triggers", start, w.fire(ctx, start))

		select {
		case <-ctx.Done():
			return
		case <-time.After(w.interval):
		}
	}
}

func (w *Worker) fire(ctx context.Context, now time.Time) error {
	triggers, err := w.triggers.Triggers(ctx)

	if err != nil {
		slog.Error("agent triggers skipped", "error", err)
		return err
	}

	for _, trigger := range triggers {
		if trigger.Source != values.TriggerSchedule || !trigger.Enabled {
			continue
		}

		tick, ok, err := w.due(trigger, now)

		if err != nil {
			slog.Error("agent trigger has an invalid schedule", "trigger", trigger.ID, "error", err)
			continue
		}

		if !ok {
			continue
		}

		sessions, err := w.runner.Sessions(ctx, trigger.ID.Agent)

		if err != nil {
			slog.Warn("agent trigger skipped, the agent can't be reached", "trigger", trigger.ID, "error", err)
			continue
		}

		if !agenttrigger.LastRun(sessions, trigger.ID.Trigger).Before(tick) {
			continue
		}

		event := "Scheduled run for " + tick.Format("Monday 2 January 2006, 15:04 MST")

		if err := w.triggers.Fire(ctx, trigger, "", event); err != nil {
			slog.Warn("agent trigger failed", "trigger", trigger.ID, "error", err)
			continue
		}

		slog.Info("agent trigger fired", "trigger", trigger.ID, "tick", tick)
	}

	return nil
}

// due returns the trigger's latest tick within the last two intervals.
func (w *Worker) due(trigger agenttrigger.Trigger, now time.Time) (time.Time, bool, error) {
	location, err := time.LoadLocation(trigger.TimeZone)

	if err != nil {
		return time.Time{}, false, err
	}

	schedule, err := cron.ParseStandard(trigger.Schedule)

	if err != nil {
		return time.Time{}, false, err
	}

	tick := schedule.Next(now.Add(-2 * w.interval).In(location))

	return tick, !tick.After(now), nil
}
