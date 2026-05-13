package graphql

import (
	"strings"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/services/conductor/internal/api/graphql/model"
	"github.com/zeitlos/lucity/services/conductor/internal/conductor"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

func convertService(service platform.Service) model.Service {
	result := model.Service{
		ID:          service.ID,
		Name:        service.Name,
		Status:      model.ServiceStatus(service.Status),
		Replicas:    convertReplicaCount(service.Replicas),
		Endpoints:   convertEndpoints(service.Endpoints),
		SourceURL:   service.SourceURL,
		ContextPath: service.ContextPath,
		Resources:   convertResources(service.Resources),
		Command:     service.Command,
		CreatedAt:   service.CreatedAt,
	}

	if service.Autoscaling != nil {
		result.Autoscaling = convertAutoscaling(*service.Autoscaling)
	}

	if service.ActiveDeployment != nil {
		deployment := convertDeployment(*service.ActiveDeployment)
		result.ActiveDeployment = &deployment
	}

	if !service.LastDeployedAt.IsZero() {
		result.LastDeployedAt = &service.LastDeployedAt
	}

	return result
}

func convertDeployment(deployment platform.Deployment) model.Deployment {
	result := model.Deployment{
		ID:        deployment.ID,
		Image:     deployment.Image,
		Status:    model.DeploymentStatus(deployment.Status),
		Replicas:  convertReplicaCount(deployment.Replicas),
		Resources: convertResources(deployment.Resources),
		CreatedAt: deployment.CreatedAt,
	}

	if deployment.ImageDigest != "" {
		result.ImageDigest = &deployment.ImageDigest
	}

	if deployment.Commit != "" {
		result.Commit = deployment.Commit
	}

	if deployment.Ref != "" {
		result.Ref = deployment.Ref
	}

	if deployment.SourceURL != "" {
		result.SourceURL = deployment.SourceURL
	}

	if deployment.ContextPath != "" {
		result.ContextPath = deployment.ContextPath
	}

	if deployment.Command != "" {
		result.Command = deployment.Command
	}

	if deployment.BuildID != "" {
		result.BuildID = deployment.BuildID
	}

	if deployment.DeployedBy != "" {
		result.DeployedBy = deployment.DeployedBy
	}

	return result
}

func convertReplicaCount(replicas platform.ReplicaCount) *model.ReplicaCount {
	return &model.ReplicaCount{
		Desired: replicas.Desired,
		Ready:   replicas.Ready,
	}
}

func convertAutoscaling(autoscaling platform.AutoscalingSettings) *model.AutoscalingSettings {
	return &model.AutoscalingSettings{
		MinReplicas: autoscaling.MinReplicas,
		MaxReplicas: autoscaling.MaxReplicas,
		TargetCPU:   autoscaling.TargetCPU,
	}
}

func convertResources(resources platform.Resources) *model.Resources {
	return &model.Resources{
		CPU:    resources.CPU.String(),
		Memory: resources.Memory.String(),
	}
}

func convertEndpoints(endpoints []platform.Endpoint) []model.Endpoint {
	result := make([]model.Endpoint, 0, len(endpoints))

	for _, endpoint := range endpoints {
		result = append(result, model.Endpoint{
			Host:     endpoint.Host,
			Port:     endpoint.Port,
			Protocol: model.Protocol(strings.ToUpper(string(endpoint.Protocol))),
		})
	}

	return result
}

func convertProject(p conductor.Project) model.Project {
	result := model.Project{
		ID:   p.ID,
		Name: p.Name,
	}

	for _, e := range p.Environments {
		result.Environments = append(result.Environments, convertEnvironment(e))
	}

	return result
}

func convertEnvironment(e conductor.Environment) model.Environment {
	result := model.Environment{
		ID:   e.ID,
		Name: e.Name,
	}

	switch e.ResourceTier {
	case platform.EcoTier:
		result.ResourceTier = model.ResourceTierEco
	case platform.ProductionTier:
		result.ResourceTier = model.ResourceTierProduction
	}

	return result
}

func convertDetectedService(s conductor.DetectedService) model.DetectedService {
	return model.DetectedService{
		Name:          s.Name,
		Language:      s.Provider,
		Framework:     s.Framework,
		StartCommand:  s.StartCommand,
		SuggestedPort: s.SuggestedPort,
	}
}

func convertScalingConfig(sc conductor.ScalingConfig) model.ScalingConfig {
	result := model.ScalingConfig{
		Replicas: sc.Replicas,
	}
	if sc.Autoscaling != nil {
		result.Autoscaling = &model.AutoscalingConfig{
			Enabled:     sc.Autoscaling.Enabled,
			MinReplicas: sc.Autoscaling.MinReplicas,
			MaxReplicas: sc.Autoscaling.MaxReplicas,
			TargetCPU:   sc.Autoscaling.TargetCPU,
		}
	}
	return result
}

func convertDomain(d conductor.Domain) *model.Domain {
	return &model.Domain{
		Hostname:  d.Hostname,
		Type:      model.DomainType(d.Type),
		DNSStatus: model.DNSStatus(d.DnsStatus),
		TLSStatus: model.TLSStatus(d.TlsStatus),
	}
}

func convertDnsCheck(d conductor.DnsCheck) *model.DNSCheck {
	result := &model.DNSCheck{
		Hostname:       d.Hostname,
		Status:         model.DNSStatus(d.Status),
		ExpectedTarget: d.ExpectedTarget,
	}
	if d.CnameTarget != "" {
		result.CnameTarget = &d.CnameTarget
	}
	if d.Message != "" {
		result.Message = &d.Message
	}
	if d.TlsStatus != "" {
		tlsStatus := model.TLSStatus(d.TlsStatus)
		result.TLSStatus = &tlsStatus
	}
	return result
}

func convertDeploymentOp(d conductor.DeployOp) model.DeployRun {
	op := model.DeployRun{
		ID:    d.ID,
		Phase: model.DeployPhase(d.Phase),
	}
	if d.BuildID != "" {
		op.BuildID = &d.BuildID
	}
	if d.ImageRef != "" {
		op.ImageRef = &d.ImageRef
	}
	if d.Digest != "" {
		op.Digest = &d.Digest
	}
	if d.Error != "" {
		op.Error = &d.Error
	}
	if !d.StartedAt.IsZero() {
		op.StartedAt = &d.StartedAt
	}
	if d.RolloutHealth != "" {
		health := model.SyncStatus(d.RolloutHealth)
		op.RolloutHealth = &health
	}
	if d.RolloutMessage != "" {
		op.RolloutMessage = &d.RolloutMessage
	}
	return op
}

func convertGitHubRepository(r conductor.GitHubRepository) model.GitHubRepository {
	return model.GitHubRepository{
		ID:            r.ID,
		Name:          r.Name,
		FullName:      r.FullName,
		HTMLURL:       r.HTMLURL,
		DefaultBranch: r.DefaultBranch,
		Private:       r.Private,
	}
}

func convertGitHubInstallation(i conductor.GitHubInstallation) model.GitHubInstallation {
	return model.GitHubInstallation{
		ID:               i.ID,
		AccountLogin:     i.AccountLogin,
		AccountAvatarURL: i.AccountAvatarURL,
		AccountType:      model.GitHubAccountType(i.AccountType),
	}
}

func convertUser(u *conductor.User) *model.User {
	if u == nil {
		return nil
	}
	user := &model.User{
		AvatarURL:  u.AvatarURL,
		Workspaces: convertWorkspaceMemberships(u.Workspaces),
	}
	if u.Name != "" {
		user.Name = &u.Name
	}
	if u.Email != "" {
		user.Email = &u.Email
	}
	return user
}

func convertWorkspaceMemberships(memberships []auth.WorkspaceMembership) []model.WorkspaceMembership {
	result := make([]model.WorkspaceMembership, len(memberships))
	for i, m := range memberships {
		role := model.WorkspaceRoleUser
		if m.Role == auth.WorkspaceRoleAdmin {
			role = model.WorkspaceRoleAdmin
		}
		result[i] = model.WorkspaceMembership{
			Workspace: m.Workspace,
			Role:      role,
		}
	}
	return result
}

func convertWorkspace(ws *conductor.Workspace) *model.Workspace {
	result := &model.Workspace{
		ID:        ws.ID,
		Name:      ws.Name,
		Personal:  ws.Personal,
		Suspended: ws.Suspended,
	}

	return result
}

func convertWorkspaceDetails(ws *conductor.WorkspaceDetails) *model.Workspace {
	result := convertWorkspace(&ws.Workspace)

	for _, m := range ws.Members {
		result.Members = append(result.Members, *convertWorkspaceMember(&m))
	}

	return result
}

func convertWorkspaceMember(m *conductor.WorkspaceMember) *model.WorkspaceMember {
	role := model.WorkspaceRoleUser
	if m.Role == auth.WorkspaceRoleAdmin {
		role = model.WorkspaceRoleAdmin
	}
	result := &model.WorkspaceMember{
		ID:    m.ID,
		Email: m.Email,
		Role:  role,
	}
	if m.Name != "" {
		result.Name = &m.Name
	}
	return result
}

func convertEnvironmentResources(r conductor.EnvironmentResources) model.EnvironmentResources {
	return model.EnvironmentResources{
		Tier: model.ResourceTier(r.Tier),
		Allocation: &model.ResourceAllocation{
			CPUMillicores: r.CpuMillicores,
			MemoryMb:      r.MemoryMB,
			DiskMb:        r.DiskMB,
		},
	}
}

func convertBillingSubscription(s *conductor.BillingSubscription) *model.BillingSubscription {
	var plan *model.Plan
	if s.Plan != nil {
		p := model.Plan(*s.Plan)
		plan = &p
	}
	return &model.BillingSubscription{
		Plan:              plan,
		Status:            model.SubscriptionStatus(s.Status),
		CurrentPeriodEnd:  s.CurrentPeriodEnd,
		CreditAmountCents: s.CreditAmountCents,
		CreditExpiry:      s.CreditExpiry,
		HasPaymentMethod:  s.HasPaymentMethod,
	}
}

func convertDatabase(d conductor.Database) model.Database {
	return model.Database{
		ID:        d.ID,
		Name:      d.Name,
		Version:   d.Version,
		Instances: d.Instances,
		Size:      d.Size.String(),
	}
}

func convertDatabaseTable(t conductor.DatabaseTable) model.DatabaseTable {
	cols := make([]model.DatabaseColumn, 0, len(t.Columns))
	for _, c := range t.Columns {
		cols = append(cols, model.DatabaseColumn{
			Name:       c.Name,
			Type:       c.Type,
			Nullable:   c.Nullable,
			PrimaryKey: c.PrimaryKey,
		})
	}
	return model.DatabaseTable{
		Name:          t.Name,
		Schema:        t.Schema,
		EstimatedRows: t.EstimatedRows,
		Columns:       cols,
	}
}

func convertDatabaseTableData(d *conductor.DatabaseTableData) *model.DatabaseTableData {
	return &model.DatabaseTableData{
		Columns:            d.Columns,
		Rows:               d.Rows,
		TotalEstimatedRows: d.TotalEstimatedRows,
	}
}

func convertQueryResult(r *conductor.QueryResult) *model.QueryResult {
	return &model.QueryResult{
		Columns:      r.Columns,
		Rows:         r.Rows,
		AffectedRows: r.AffectedRows,
	}
}
