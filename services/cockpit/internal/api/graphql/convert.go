package graphql

import (
	"time"

	"github.com/zeitlos/lucity/pkg/billing"
	"github.com/zeitlos/lucity/services/cockpit/internal/api/graphql/model"
	"github.com/zeitlos/lucity/services/cockpit/internal/cockpit"
	"github.com/zeitlos/lucity/services/cockpit/internal/signals"
)

func convertSnapshot(snapshot cockpit.Snapshot) *model.WorkspaceList {
	result := &model.WorkspaceList{
		LoadedAt:   snapshot.LoadedAt,
		Workspaces: make([]model.Workspace, 0, len(snapshot.Workspaces)),
	}

	for _, workspace := range snapshot.Workspaces {
		result.Workspaces = append(result.Workspaces, convertWorkspace(workspace))
	}

	return result
}

func convertWorkspace(workspace cockpit.Workspace) model.Workspace {
	result := model.Workspace{
		ID:              workspace.ID,
		CreatedAt:       workspace.CreatedAt,
		Abuse:           workspace.Abuse,
		AbuseNote:       optional(workspace.AbuseNote),
		Suspended:       workspace.Suspended,
		Members:         []model.Member{},
		Signals:         []model.Signal{},
		RunningPods:     workspace.RunningPods,
		PendingPods:     workspace.PendingPods,
		MemoryRequested: int(workspace.MemoryRequested),
		Egress7d:        int(workspace.Egress7d),
		Environments:    []model.Environment{},
	}

	if workspace.Billing != nil {
		result.Billing = convertBilling(*workspace.Billing)
	}

	for _, member := range workspace.Members {
		result.Members = append(result.Members, convertMember(member))
	}

	for _, signal := range workspace.Signals {
		result.Signals = append(result.Signals, model.Signal{
			Kind:   convertSignalKind(signal.Kind),
			Strong: signal.Strong,
			Detail: signal.Detail,
		})
	}

	for _, environment := range workspace.Environments {
		result.Environments = append(result.Environments, convertEnvironment(environment))
	}

	return result
}

func convertBilling(b cockpit.Billing) *model.Billing {
	return &model.Billing{
		Plan:             convertPlan(b.Plan),
		TrialEndsAt:      optionalTime(b.TrialEndsAt),
		GoodStanding:     b.GoodStanding,
		StandingReason:   optional(b.StandingReason),
		HasPaymentMethod: b.HasPaymentMethod,
	}
}

func convertMember(member cockpit.Member) model.Member {
	result := model.Member{
		Name:  optional(member.Name),
		Email: optional(member.Email),
		Role:  optional(member.Role),
	}

	if member.Account != nil {
		result.Account = &model.Account{Login: member.Account.Login, CreatedAt: member.Account.CreatedAt}
	}

	return result
}

func convertEnvironment(environment cockpit.Environment) model.Environment {
	result := model.Environment{
		ID:          environment.ID.String(),
		Project:     environment.ID.Project,
		Name:        environment.ID.Name,
		RunningPods: environment.RunningPods,
		PendingPods: environment.PendingPods,
		Services:    []model.Service{},
	}

	for _, service := range environment.Services {
		result.Services = append(result.Services, model.Service{
			Name:       service.Name,
			Image:      optional(service.Image),
			Commit:     optional(service.Commit),
			Repository: optional(service.Repository),
		})
	}

	return result
}

func convertResults(results []cockpit.Result) []model.WorkspaceResult {
	converted := make([]model.WorkspaceResult, 0, len(results))

	for _, result := range results {
		item := model.WorkspaceResult{WorkspaceID: result.WorkspaceID}

		if result.Err != nil {
			item.Error = optional(result.Err.Error())
		}

		converted = append(converted, item)
	}

	return converted
}

func convertSignupHours(hours []cockpit.SignupHour) []model.SignupHour {
	converted := make([]model.SignupHour, 0, len(hours))

	for _, hour := range hours {
		converted = append(converted, model.SignupHour{Start: hour.Start, Count: hour.Count})
	}

	return converted
}

func convertWorkspaceStandings(standings []cockpit.WorkspaceStanding) []model.WorkspaceStanding {
	converted := make([]model.WorkspaceStanding, 0, len(standings))

	for _, standing := range standings {
		converted = append(converted, model.WorkspaceStanding{
			At:      standing.At,
			Good:    standing.Good,
			Billing: standing.Billing,
			Abuse:   standing.Abuse,
		})
	}

	return converted
}

func convertWorkspacePlans(plans []cockpit.WorkspacePlan) []model.WorkspacePlan {
	converted := make([]model.WorkspacePlan, 0, len(plans))

	for _, plan := range plans {
		converted = append(converted, model.WorkspacePlan{
			At:    plan.At,
			None:  plan.None,
			Trial: plan.Trial,
			Hobby: plan.Hobby,
			Pro:   plan.Pro,
		})
	}

	return converted
}

func convertPlan(plan billing.Plan) model.Plan {
	switch plan {
	case billing.PlanHobby:
		return model.PlanHobby
	case billing.PlanPro:
		return model.PlanPro
	}
	return model.PlanNone
}

func convertSignalKind(kind signals.Kind) model.SignalKind {
	switch kind {
	case signals.SharedCommit:
		return model.SignalKindSharedCommit
	case signals.Image:
		return model.SignalKindImage
	case signals.Fork:
		return model.SignalKindFork
	}
	return model.SignalKindYoungAccount
}

func optional(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func optionalTime(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}
