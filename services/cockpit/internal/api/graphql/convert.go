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
		PaymentMethod:   convertPaymentMethod(workspace.PaymentMethod),
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
		Plan:            convertPlan(b.Plan),
		TrialEndsAt:     optionalTime(b.TrialEndsAt),
		GoodStanding:    b.GoodStanding,
		StandingReason:  optional(b.StandingReason),
		CustomerURL:     b.CustomerURL,
		SubscriptionURL: b.SubscriptionURL,
	}
}

func convertMember(member cockpit.Member) model.Member {
	result := model.Member{
		ID:           member.ID,
		Name:         optional(member.Name),
		Email:        optional(member.Email),
		Avatar:       optional(member.Avatar),
		Role:         optional(member.Role),
		SignedUpAt:   member.SignedUpAt,
		LastSignInAt: optionalTime(member.LastSignInAt),
	}

	if member.Account != nil {
		result.Account = &model.Account{Login: member.Account.Login, CreatedAt: member.Account.CreatedAt}
	}

	return result
}

func convertUserSessions(sessions cockpit.UserSessions) *model.UserSessions {
	result := &model.UserSessions{
		URL:      sessions.URL,
		Sessions: make([]model.Session, 0, len(sessions.Sessions)),
	}

	for _, session := range sessions.Sessions {
		result.Sessions = append(result.Sessions, model.Session{
			ID:              session.ID,
			StartedAt:       session.StartedAt,
			Duration:        int(session.Duration.Seconds()),
			EntryPage:       session.EntryPage,
			Pageviews:       session.Pageviews,
			Country:         session.Country,
			City:            session.City,
			Language:        session.Language,
			Browser:         session.Browser,
			OperatingSystem: session.OperatingSystem,
			DeviceType:      session.DeviceType,
			Referrer:        session.Referrer,
			HasReplay:       session.HasReplay,
		})
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

func convertPaymentMethod(method cockpit.PaymentMethod) model.PaymentMethod {
	switch method {
	case cockpit.PaymentMethodOnFile:
		return model.PaymentMethodOnFile
	case cockpit.PaymentMethodRequired:
		return model.PaymentMethodRequired
	}
	return model.PaymentMethodNone
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

func convertTotals(totals cockpit.Totals) *model.PlatformTotals {
	return &model.PlatformTotals{
		RunningPods:            totals.RunningPods,
		PendingPods:            totals.PendingPods,
		FailedPods:             totals.FailedPods,
		RunningContainers:      totals.RunningContainers,
		RunningWorkspaces:      totals.RunningWorkspaces,
		Workspaces:             totals.Workspaces,
		Releases:               totals.Releases,
		FailedReleases:         totals.FailedReleases,
		FiringAlerts:           totals.FiringAlerts,
		RunningPodsTrend:       totals.RunningPodsTrend,
		RunningContainersTrend: totals.RunningContainersTrend,
		RunningWorkspacesTrend: totals.RunningWorkspacesTrend,
		ReleasesTrend:          totals.ReleasesTrend,
		FiringAlertsTrend:      totals.FiringAlertsTrend,
	}
}

func convertRoutines(routines []cockpit.Routine) []model.Routine {
	result := make([]model.Routine, 0, len(routines))

	for _, routine := range routines {
		converted := model.Routine{
			Name:                  routine.Name,
			Hours:                 make([]model.RoutineHour, 0, len(routine.Hours)),
			AverageDurationChange: routine.AverageDurationChange,
			LastPassedAt:          routine.LastPassedAt,
		}

		if routine.AverageDuration != nil {
			seconds := routine.AverageDuration.Seconds()
			converted.AverageDuration = &seconds
		}

		for _, hour := range routine.Hours {
			converted.Hours = append(converted.Hours, model.RoutineHour{Start: hour.Start, Passed: hour.Passed, Failed: hour.Failed})
		}

		result = append(result, converted)
	}

	return result
}

func convertLimitGroups(groups []cockpit.LimitGroup) []model.LimitGroup {
	result := make([]model.LimitGroup, 0, len(groups))

	for _, group := range groups {
		converted := model.LimitGroup{Name: group.Name, Limits: make([]model.Limit, 0, len(group.Limits))}

		for _, limit := range group.Limits {
			history := make([]model.Point, 0, len(limit.History))

			for _, point := range limit.History {
				history = append(history, model.Point{Time: point.Time, Value: point.Value})
			}

			converted.Limits = append(converted.Limits, model.Limit{
				Name:    limit.Name,
				Unit:    convertLimitUnit(limit.Unit),
				Value:   limit.Value,
				Cap:     limit.Cap,
				History: history,
			})
		}

		result = append(result, converted)
	}

	return result
}

func convertLimitUnit(unit cockpit.LimitUnit) model.LimitUnit {
	if unit == cockpit.LimitUnitBytes {
		return model.LimitUnitBytes
	}

	return model.LimitUnitCount
}

func convertNodePools(pools []cockpit.NodePool) []model.NodePool {
	result := make([]model.NodePool, 0, len(pools))

	for _, pool := range pools {
		converted := model.NodePool{Name: pool.Name, Nodes: make([]model.Node, 0, len(pool.Nodes))}

		for _, node := range pool.Nodes {
			item := model.Node{
				Name:              node.Name,
				InstanceType:      node.InstanceType,
				CreatedAt:         node.CreatedAt,
				AttachedVolumes:   node.AttachedVolumes,
				VolumeLimit:       node.VolumeLimit,
				Pods:              node.Pods,
				PodLimit:          node.PodLimit,
				MemoryAllocatable: int(node.MemoryAllocatable),
				MemoryRequested:   int(node.MemoryRequested),
				CPUAllocatable:    node.CPUAllocatable,
				CPURequested:      node.CPURequested,
				CPUUsed:           node.CPUUsed,
			}

			if node.MemoryUsed != nil {
				used := int(*node.MemoryUsed)
				item.MemoryUsed = &used
			}

			converted.Nodes = append(converted.Nodes, item)
		}

		result = append(result, converted)
	}

	return result
}

func convertPlatformLogs(logs cockpit.PlatformLogs) *model.PlatformLogs {
	result := &model.PlatformLogs{
		Hours:       make([]model.LogHour, 0, len(logs.Hours)),
		Errors24h:   logs.Errors24h,
		Errors7d:    logs.Errors7d,
		Warnings24h: logs.Warnings24h,
		Warnings7d:  logs.Warnings7d,
		Messages:    make([]model.LogMessage, 0, len(logs.Messages)),
	}

	for _, hour := range logs.Hours {
		result.Hours = append(result.Hours, model.LogHour{Start: hour.Start, Errors: hour.Errors, Warnings: hour.Warnings})
	}

	for _, message := range logs.Messages {
		level := model.LogLevelWarn

		if message.Level == cockpit.LogLevelError {
			level = model.LogLevelError
		}

		result.Messages = append(result.Messages, model.LogMessage{
			Level:       level,
			Message:     message.Message,
			Service:     message.Service,
			Count24h:    message.Count24h,
			Count7d:     message.Count7d,
			FirstSeenAt: message.FirstSeenAt,
		})
	}

	return result
}

func convertChartVersions(versions []cockpit.ChartVersion) []model.ChartVersion {
	result := make([]model.ChartVersion, 0, len(versions))

	for _, version := range versions {
		result = append(result, model.ChartVersion{Version: version.Version, Releases: version.Releases, Since: version.Since})
	}

	return result
}
