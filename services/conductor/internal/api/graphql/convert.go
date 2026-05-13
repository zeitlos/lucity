package graphql

import (
	"strings"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/services/conductor/internal/api/graphql/model"
	"github.com/zeitlos/lucity/services/conductor/internal/conductor"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

func convertProject(p conductor.ProjectNew, workloadDomain string) model.Project {
	result := model.Project{
		ID:   p.ID,
		Name: p.Name,
	}

	for _, e := range p.Environments {
		result.Environments = append(result.Environments, convertEnvironment(e, workloadDomain))
	}

	return result
}

func convertEnvironment(e conductor.EnvironmentNew, workloadDomain string) model.Environment {
	result := model.Environment{
		ID:   e.ID,
		Name: e.Name,
	}

	switch e.ResourceTier {
	case platform.EcoTier:
		result.ResourceTier = new(model.ResourceTierEco)
	case platform.ProductionTier:
		result.ResourceTier = new(model.ResourceTierProduction)
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

func convertServiceInstance(si conductor.ServiceInstance, workloadDomain string) model.ServiceInstance {
	scaling := convertScalingConfig(si.Scaling)
	result := model.ServiceInstance{
		ID:       si.ID,
		Name:     si.Name,
		Image:    si.Image,
		ImageTag: si.ImageTag,
		Ready:    si.Ready,
		Replicas: si.Replicas,
		Scaling:  &scaling,
	}

	if si.Port > 0 {
		port := si.Port
		result.Port = &port
	}
	if si.Framework != "" {
		result.Framework = &si.Framework
	}
	if si.SourceURL != "" {
		result.SourceURL = &si.SourceURL
	}
	if si.ContextPath != "" {
		result.ContextPath = &si.ContextPath
	}
	if si.StartCommand != "" {
		result.StartCommand = &si.StartCommand
	}
	if si.CustomStartCommand != "" {
		result.CustomStartCommand = &si.CustomStartCommand
	}
	if si.InitialDeploy != nil {
		d := convertDeploymentOp(*si.InitialDeploy)
		result.InitialDeploy = &d
	}
	if si.Resources != nil {
		result.Resources = &model.ServiceResources{
			CPUMillicores:      si.Resources.CpuMillicores,
			MemoryMb:           si.Resources.MemoryMB,
			CPULimitMillicores: si.Resources.CpuLimitMillicores,
			MemoryLimitMb:      si.Resources.MemoryLimitMB,
		}
	}

	// Convert domains with type derived from workload domain suffix
	for _, hostname := range si.Domains {
		domainType := model.DomainTypeCustom
		dnsStatus := model.DNSStatusPending
		tlsStatus := model.TLSStatusNone
		if strings.HasSuffix(hostname, "."+workloadDomain) {
			domainType = model.DomainTypePlatform
			dnsStatus = model.DNSStatusValid
		}
		result.Domains = append(result.Domains, model.Domain{
			Hostname:  hostname,
			Type:      domainType,
			DNSStatus: dnsStatus,
			TLSStatus: tlsStatus,
		})
	}

	// Convert deployment history
	for _, d := range si.Deployments {
		result.Deployments = append(result.Deployments, convertDeployment(d))
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

func convertDeployment(d conductor.Deployment) model.Deployment {
	dep := model.Deployment{
		ID:       d.ID,
		ImageTag: d.ImageTag,
		Active:   d.Active,
	}
	if !d.Timestamp.IsZero() {
		dep.Timestamp = &d.Timestamp
	}
	if d.Revision != "" {
		dep.Revision = &d.Revision
	}
	if d.Message != "" {
		dep.Message = &d.Message
	}
	if d.SourceCommitMessage != "" {
		dep.SourceCommitMessage = &d.SourceCommitMessage
	}
	if d.SourceURL != "" {
		dep.SourceURL = &d.SourceURL
	}
	return dep
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
		Size:      d.Size,
	}
}

func convertDatabaseInstance(di conductor.DatabaseInstance) model.DatabaseInstance {
	result := model.DatabaseInstance{
		Name:      di.Name,
		Ready:     di.Ready,
		Instances: di.Instances,
		Version:   di.Version,
		Size:      di.Size,
	}
	if di.Volume != nil {
		result.Volume = &model.Volume{
			ID:            di.Volume.ID,
			Name:          di.Volume.Name,
			Size:          di.Volume.Size,
			RequestedSize: di.Volume.RequestedSize,
			UsedBytes:     int(di.Volume.UsedBytes),
			CapacityBytes: int(di.Volume.CapacityBytes),
		}
	}
	return result
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
