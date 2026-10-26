// Package workspaces is the single source of the workspaces and environments
// every stagehand worker acts on: the Logto organizations, and the namespaces
// labelled with one of them. Namespaces left over from deleted workspaces are
// never returned.
package workspaces

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/pkg/platform"
)

type Directory struct {
	logto    *logto.Client
	platform platform.Interface
	maxAge   time.Duration

	mu            sync.Mutex
	fetched       time.Time
	organizations []logto.Organization
}

// New returns a directory that reads the Logto organizations at most once
// per maxAge.
func New(logto *logto.Client, platform platform.Interface, maxAge time.Duration) *Directory {
	return &Directory{logto: logto, platform: platform, maxAge: maxAge}
}

func (d *Directory) Organizations(ctx context.Context) ([]logto.Organization, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.organizations != nil && time.Since(d.fetched) < d.maxAge {
		return d.organizations, nil
	}

	organizations, err := d.logto.Organizations(ctx)

	if err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}

	d.organizations = organizations
	d.fetched = time.Now()

	return organizations, nil
}

func (d *Directory) Known(ctx context.Context) (map[string]bool, error) {
	organizations, err := d.Organizations(ctx)

	if err != nil {
		return nil, err
	}

	known := make(map[string]bool, len(organizations))

	for _, organization := range organizations {
		known[organization.Name] = true
	}

	return known, nil
}

func (d *Directory) Environments(ctx context.Context) ([]platform.Environment, error) {
	organizations, err := d.Organizations(ctx)

	if err != nil {
		return nil, err
	}

	var environments []platform.Environment

	for _, organization := range organizations {
		workspaceEnvironments, err := d.platform.EnvironmentsByWorkspace(ctx, organization.Name)

		if err != nil {
			slog.WarnContext(ctx, "workspaces: failed to list environments", "workspace", organization.Name, "error", err)
			continue
		}

		environments = append(environments, workspaceEnvironments...)
	}

	return environments, nil
}

// MarkSuspended records on the organization whether its workspace is
// suspended, the flag conductor's mutation gate reads. Logto replaces the
// whole custom data on update, so it re-reads the organization first and
// writes only a change.
func (d *Directory) MarkSuspended(ctx context.Context, organization logto.Organization, suspended bool) error {
	current, err := d.logto.Organization(ctx, organization.ID)

	if err != nil {
		return err
	}

	customData := current.CustomData

	if customData == nil {
		customData = map[string]interface{}{}
	}

	if flagged, _ := customData["suspended"].(bool); flagged == suspended {
		return nil
	}

	if suspended {
		customData["suspended"] = true
	} else {
		delete(customData, "suspended")
	}

	if _, err := d.logto.UpdateOrganizationCustomData(ctx, organization.ID, customData); err != nil {
		return fmt.Errorf("mark workspace %s: %w", organization.Name, err)
	}

	return nil
}

func IDs(environments []platform.Environment) []platform.EnvironmentID {
	ids := make([]platform.EnvironmentID, 0, len(environments))

	for _, environment := range environments {
		ids = append(ids, environment.ID)
	}

	return ids
}
