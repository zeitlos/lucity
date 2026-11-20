// Package mcpserver serves the cockpit's workspaces and abuse marking as
// MCP tools over stdio.
package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/pkg/to"
	"github.com/zeitlos/lucity/services/cockpit/internal/cockpit"
	"github.com/zeitlos/lucity/services/cockpit/internal/signals"
)

// Clients returns the cockpit client of an environment, dev or prod.
type Clients func(environment string) (*cockpit.Client, error)

type server struct {
	clients Clients
}

const instructions = `Lucity Cockpit is the operator view of a Lucity platform: every workspace (tenant) with its billing, members, environments, running pods and abuse signals, and the one change an operator makes, flagging a workspace as abuse.

Every tool takes environment, dev or prod. Prod holds real, paying customers.

Signals: shared-commit (a commit deployed by 3 or more workspaces), image (a denylisted image) and fork (a fork of a known panel repository) are strong. young-account (the owner's GitHub account is younger than 90 days) is weak.

mark_abuse flags workspaces. Stagehand suspends them on its next pass, every 2 minutes or so: workloads scale to zero, routes go away and deploys are blocked. Until then a workspace reads abuse true and suspended false. clear_abuse removes the flag, and stagehand resumes the workspace when its billing is in good standing. Confirm with the user before marking or clearing on prod.

Every call loads fresh from the cluster, identity provider and billing, which takes a few seconds.`

func Serve(ctx context.Context, clients Clients) error {
	m := mcp.NewServer(&mcp.Implementation{Name: "cockpit", Title: "Lucity Cockpit", Version: "0.1.0"}, &mcp.ServerOptions{
		Instructions: instructions,
	})

	s := &server{clients: clients}

	mcp.AddTool(m, &mcp.Tool{
		Name:        "list_workspaces",
		Description: "List workspaces newest first, one compact row each: plan, owner, signals, pods, memory requested, 7-day egress, abuse and suspension. Filter by query, status, strongSignal, plans, the owner's GitHub account age and the creation date. matched counts every match whatever the limit, so a how-many question needs only limit 1.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, s.listWorkspaces)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "get_workspace",
		Description: "Get one workspace in full: billing, members with their GitHub accounts, signals, and every environment with its services, images, commits and repositories.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, s.getWorkspace)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "get_environment_logs",
		Description: "Get the last log lines of every running pod in an environment. The environment ID is workspace/project/environment, as get_workspace lists it.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, s.environmentLogs)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "mark_abuse",
		Description: "Flag workspaces as abuse with an optional note. Stagehand suspends them on its next pass. Returns one result per workspace.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: to.Ptr(true), IdempotentHint: true},
	}, s.markAbuse)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "clear_abuse",
		Description: "Remove the abuse flag from workspaces. Stagehand resumes those in good standing on its next pass. Returns one result per workspace.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: to.Ptr(false), IdempotentHint: true},
	}, s.clearAbuse)

	return m.Run(ctx, &mcp.StdioTransport{})
}

type listInput struct {
	Environment   string   `json:"environment" jsonschema:"dev or prod"`
	Query         string   `json:"query,omitempty" jsonschema:"case-insensitive text in the workspace ID, a member's name, email or GitHub login, or a service's image, repository or commit"`
	Status        string   `json:"status,omitempty" jsonschema:"active, suspending or suspended"`
	StrongSignal  bool     `json:"strongSignal,omitempty" jsonschema:"only workspaces with a strong abuse signal"`
	Plans         []string `json:"plans,omitempty" jsonschema:"any of none (no subscription), trial, hobby and pro"`
	OwnerMinDays  *int     `json:"ownerMinDays,omitempty" jsonschema:"the owner's GitHub account is at least this many days old"`
	OwnerMaxDays  *int     `json:"ownerMaxDays,omitempty" jsonschema:"the owner's GitHub account is at most this many days old"`
	CreatedAfter  string   `json:"createdAfter,omitempty" jsonschema:"created on or after this date, YYYY-MM-DD in local time or RFC 3339"`
	CreatedBefore string   `json:"createdBefore,omitempty" jsonschema:"created before this date, YYYY-MM-DD in local time or RFC 3339"`
	Limit         int      `json:"limit,omitempty" jsonschema:"at most this many rows, default 50, max 100"`
}

type filter struct {
	listInput
	query         string
	createdAfter  time.Time
	createdBefore time.Time
}

func newFilter(input listInput) (filter, error) {
	f := filter{listInput: input, query: strings.ToLower(input.Query)}

	if input.Status != "" && !slices.Contains([]string{"active", "suspending", "suspended"}, input.Status) {
		return f, fmt.Errorf("status must be active, suspending or suspended")
	}

	for _, plan := range input.Plans {
		if !slices.Contains([]string{"none", "trial", "hobby", "pro"}, plan) {
			return f, fmt.Errorf("plans must be none, trial, hobby or pro, not %q", plan)
		}
	}

	if (input.OwnerMinDays != nil && *input.OwnerMinDays < 0) || (input.OwnerMaxDays != nil && *input.OwnerMaxDays < 0) {
		return f, fmt.Errorf("owner account days must not be negative")
	}

	var err error

	if f.createdAfter, err = parseDate(input.CreatedAfter); err != nil {
		return f, fmt.Errorf("createdAfter must be YYYY-MM-DD or RFC 3339")
	}

	if f.createdBefore, err = parseDate(input.CreatedBefore); err != nil {
		return f, fmt.Errorf("createdBefore must be YYYY-MM-DD or RFC 3339")
	}

	return f, nil
}

func (f filter) matches(workspace cockpit.Workspace) bool {
	if f.Status != "" && status(workspace) != f.Status {
		return false
	}

	if f.StrongSignal && !slices.ContainsFunc(workspace.Signals, func(signal signals.Signal) bool { return signal.Strong }) {
		return false
	}

	if len(f.Plans) > 0 && !slices.Contains(f.Plans, plan(workspace.Billing)) {
		return false
	}

	if f.OwnerMinDays != nil || f.OwnerMaxDays != nil {
		days, ok := ownerDays(workspace)

		if !ok || (f.OwnerMinDays != nil && days < *f.OwnerMinDays) || (f.OwnerMaxDays != nil && days > *f.OwnerMaxDays) {
			return false
		}
	}

	if !f.createdAfter.IsZero() && workspace.CreatedAt.Before(f.createdAfter) {
		return false
	}

	if !f.createdBefore.IsZero() && !workspace.CreatedAt.Before(f.createdBefore) {
		return false
	}

	return f.query == "" || strings.Contains(haystack(workspace), f.query)
}

func (s *server) listWorkspaces(ctx context.Context, _ *mcp.CallToolRequest, input listInput) (*mcp.CallToolResult, any, error) {
	f, err := newFilter(input)

	if err != nil {
		return nil, nil, err
	}

	if input.Limit < 0 || input.Limit > 100 {
		return nil, nil, fmt.Errorf("limit must be between 1 and 100")
	}

	limit := input.Limit

	if limit == 0 {
		limit = 50
	}

	snapshot, err := s.snapshot(ctx, input.Environment)

	if err != nil {
		return nil, nil, err
	}

	rows := []row{}
	matched := 0

	for _, workspace := range snapshot.Workspaces {
		if !f.matches(workspace) {
			continue
		}

		matched++

		if len(rows) < limit {
			rows = append(rows, convertRow(workspace))
		}
	}

	return jsonResult(map[string]any{
		"loadedAt":   snapshot.LoadedAt.Local(),
		"matched":    matched,
		"total":      len(snapshot.Workspaces),
		"workspaces": rows,
	})
}

type workspaceInput struct {
	Environment string `json:"environment" jsonschema:"dev or prod"`
	ID          string `json:"id" jsonschema:"workspace ID"`
}

func (s *server) getWorkspace(ctx context.Context, _ *mcp.CallToolRequest, input workspaceInput) (*mcp.CallToolResult, any, error) {
	snapshot, err := s.snapshot(ctx, input.Environment)

	if err != nil {
		return nil, nil, err
	}

	for _, workspace := range snapshot.Workspaces {
		if workspace.ID == input.ID {
			return jsonResult(convertDetail(workspace))
		}
	}

	return nil, nil, fmt.Errorf("no workspace %q in %s", input.ID, input.Environment)
}

type logsInput struct {
	Environment   string `json:"environment" jsonschema:"dev or prod"`
	EnvironmentID string `json:"environmentId" jsonschema:"workspace/project/environment"`
	Lines         int    `json:"lines,omitempty" jsonschema:"last lines per pod, default 100, max 1000"`
}

func (s *server) environmentLogs(ctx context.Context, _ *mcp.CallToolRequest, input logsInput) (*mcp.CallToolResult, any, error) {
	if input.Lines < 0 || input.Lines > 1000 {
		return nil, nil, fmt.Errorf("lines must be between 1 and 1000")
	}

	lines := input.Lines

	if lines == 0 {
		lines = 100
	}

	environmentID, err := platform.ParseEnvironmentID(input.EnvironmentID)

	if err != nil {
		return nil, nil, fmt.Errorf("invalid environment ID %q", input.EnvironmentID)
	}

	client, err := s.clients(input.Environment)

	if err != nil {
		return nil, nil, err
	}

	logs, err := client.EnvironmentLogs(ctx, environmentID, int64(lines))

	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: strings.Join(logs, "\n")}}}, nil, nil
}

type markInput struct {
	Environment  string   `json:"environment" jsonschema:"dev or prod"`
	WorkspaceIDs []string `json:"workspaceIds" jsonschema:"workspaces to flag"`
	Note         string   `json:"note,omitempty" jsonschema:"why, kept on the workspace, max 500 characters"`
}

func (s *server) markAbuse(ctx context.Context, _ *mcp.CallToolRequest, input markInput) (*mcp.CallToolResult, any, error) {
	if len(input.WorkspaceIDs) == 0 {
		return nil, nil, fmt.Errorf("workspaceIds is empty")
	}

	if len([]rune(input.Note)) > 500 {
		return nil, nil, fmt.Errorf("note is longer than 500 characters")
	}

	client, err := s.clients(input.Environment)

	if err != nil {
		return nil, nil, err
	}

	return jsonResult(convertResults(client.MarkAbuse(ctx, input.WorkspaceIDs, strings.TrimSpace(input.Note))))
}

type clearInput struct {
	Environment  string   `json:"environment" jsonschema:"dev or prod"`
	WorkspaceIDs []string `json:"workspaceIds" jsonschema:"workspaces to clear"`
}

func (s *server) clearAbuse(ctx context.Context, _ *mcp.CallToolRequest, input clearInput) (*mcp.CallToolResult, any, error) {
	if len(input.WorkspaceIDs) == 0 {
		return nil, nil, fmt.Errorf("workspaceIds is empty")
	}

	client, err := s.clients(input.Environment)

	if err != nil {
		return nil, nil, err
	}

	return jsonResult(convertResults(client.ClearAbuse(ctx, input.WorkspaceIDs)))
}

func (s *server) snapshot(ctx context.Context, environment string) (cockpit.Snapshot, error) {
	client, err := s.clients(environment)

	if err != nil {
		return cockpit.Snapshot{}, err
	}

	return client.Workspaces(ctx)
}

func status(workspace cockpit.Workspace) string {
	switch {
	case workspace.Suspended:
		return "suspended"
	case workspace.Abuse:
		return "suspending"
	default:
		return "active"
	}
}

func ownerDays(workspace cockpit.Workspace) (int, bool) {
	owner := owner(workspace.Members)

	if owner == nil || owner.Account == nil {
		return 0, false
	}

	return int(time.Since(owner.Account.CreatedAt).Hours() / 24), true
}

func parseDate(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}

	if date, err := time.ParseInLocation(time.DateOnly, value, time.Local); err == nil {
		return date, nil
	}

	return time.Parse(time.RFC3339, value)
}

func haystack(workspace cockpit.Workspace) string {
	parts := []string{workspace.ID}

	for _, member := range workspace.Members {
		parts = append(parts, member.Name, member.Email)

		if member.Account != nil {
			parts = append(parts, member.Account.Login)
		}
	}

	for _, environment := range workspace.Environments {
		for _, service := range environment.Services {
			parts = append(parts, service.Image, service.Commit, service.Repository)
		}
	}

	return strings.ToLower(strings.Join(parts, "\n"))
}

func jsonResult(v any) (*mcp.CallToolResult, any, error) {
	data, err := json.Marshal(v)

	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(data)}}}, nil, nil
}
