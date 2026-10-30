package mcpserver

import (
	"time"

	"github.com/zeitlos/lucity/pkg/billing"
	"github.com/zeitlos/lucity/services/cockpit/internal/cockpit"
	"github.com/zeitlos/lucity/services/cockpit/internal/signals"
)

type row struct {
	ID              string     `json:"id"`
	CreatedAt       time.Time  `json:"createdAt"`
	Plan            string     `json:"plan"`
	Owner           string     `json:"owner,omitempty"`
	OwnerAccountAge *time.Time `json:"ownerAccountCreatedAt,omitempty"`
	Members         int        `json:"members"`
	Signals         []signal   `json:"signals"`
	RunningPods     int        `json:"runningPods"`
	PendingPods     int        `json:"pendingPods"`
	MemoryRequested int64      `json:"memoryRequestedBytes"`
	Egress7d        int64      `json:"egress7dBytes"`
	Status          string     `json:"status"`
}

type signal struct {
	Kind   signals.Kind `json:"kind"`
	Strong bool         `json:"strong"`
	Detail string       `json:"detail"`
}

type detail struct {
	row
	AbuseNote    string        `json:"abuseNote,omitempty"`
	Billing      *billingView  `json:"billing"`
	MemberList   []member      `json:"memberList"`
	Environments []environment `json:"environments"`
}

type billingView struct {
	TrialEndsAt      *time.Time `json:"trialEndsAt,omitempty"`
	GoodStanding     bool       `json:"goodStanding"`
	StandingReason   string     `json:"standingReason,omitempty"`
	HasPaymentMethod bool       `json:"hasPaymentMethod"`
}

type member struct {
	Name                 string     `json:"name,omitempty"`
	Email                string     `json:"email,omitempty"`
	Role                 string     `json:"role"`
	GitHubLogin          string     `json:"githubLogin,omitempty"`
	GitHubAccountCreated *time.Time `json:"githubAccountCreatedAt,omitempty"`
}

type environment struct {
	ID              string    `json:"id"`
	RunningPods     int       `json:"runningPods"`
	PendingPods     int       `json:"pendingPods"`
	MemoryRequested int64     `json:"memoryRequestedBytes"`
	Services        []service `json:"services"`
}

type service struct {
	Name       string `json:"name"`
	Image      string `json:"image,omitempty"`
	Commit     string `json:"commit,omitempty"`
	Repository string `json:"repository,omitempty"`
}

type result struct {
	WorkspaceID string `json:"workspaceId"`
	Error       string `json:"error,omitempty"`
}

func convertRow(workspace cockpit.Workspace) row {
	converted := row{
		ID:              workspace.ID,
		CreatedAt:       workspace.CreatedAt.Local(),
		Plan:            plan(workspace.Billing),
		Members:         len(workspace.Members),
		Signals:         []signal{},
		RunningPods:     workspace.RunningPods,
		PendingPods:     workspace.PendingPods,
		MemoryRequested: workspace.MemoryRequested,
		Egress7d:        workspace.Egress7d,
		Status:          status(workspace),
	}

	if owner := owner(workspace.Members); owner != nil {
		converted.Owner = owner.Email

		if owner.Account != nil {
			converted.Owner = owner.Account.Login
			converted.OwnerAccountAge = localTime(owner.Account.CreatedAt)
		}
	}

	for _, s := range workspace.Signals {
		converted.Signals = append(converted.Signals, signal{Kind: s.Kind, Strong: s.Strong, Detail: s.Detail})
	}

	return converted
}

func convertDetail(workspace cockpit.Workspace) detail {
	converted := detail{
		row:          convertRow(workspace),
		AbuseNote:    workspace.AbuseNote,
		MemberList:   []member{},
		Environments: []environment{},
	}

	if workspace.Billing != nil {
		converted.Billing = &billingView{
			TrialEndsAt:      localTime(workspace.Billing.TrialEndsAt),
			GoodStanding:     workspace.Billing.GoodStanding,
			StandingReason:   workspace.Billing.StandingReason,
			HasPaymentMethod: workspace.Billing.HasPaymentMethod,
		}
	}

	for _, m := range workspace.Members {
		converted.MemberList = append(converted.MemberList, convertMember(m))
	}

	for _, e := range workspace.Environments {
		services := []service{}

		for _, s := range e.Services {
			services = append(services, service{Name: s.Name, Image: s.Image, Commit: s.Commit, Repository: s.Repository})
		}

		converted.Environments = append(converted.Environments, environment{
			ID:              e.ID.String(),
			RunningPods:     e.RunningPods,
			PendingPods:     e.PendingPods,
			MemoryRequested: e.MemoryRequested,
			Services:        services,
		})
	}

	return converted
}

func convertMember(m cockpit.Member) member {
	converted := member{Name: m.Name, Email: m.Email, Role: m.Role}

	if m.Account != nil {
		converted.GitHubLogin = m.Account.Login
		converted.GitHubAccountCreated = localTime(m.Account.CreatedAt)
	}

	return converted
}

func convertResults(results []cockpit.Result) []result {
	converted := []result{}

	for _, r := range results {
		item := result{WorkspaceID: r.WorkspaceID}

		if r.Err != nil {
			item.Error = r.Err.Error()
		}

		converted = append(converted, item)
	}

	return converted
}

func plan(b *cockpit.Billing) string {
	if b == nil {
		return "none"
	}

	switch b.Plan {
	case billing.PlanHobby:
		return "hobby"
	case billing.PlanPro:
		return "pro"
	default:
		return "trial"
	}
}

func owner(members []cockpit.Member) *cockpit.Member {
	for i, m := range members {
		if m.Role == "admin" {
			return &members[i]
		}
	}

	if len(members) > 0 {
		return &members[0]
	}

	return nil
}

func localTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}

	local := t.Local()
	return &local
}
