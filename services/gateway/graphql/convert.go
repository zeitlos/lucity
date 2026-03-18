package graphql

import (
	"strings"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/services/gateway/graphql/model"
	"github.com/zeitlos/lucity/services/gateway/handler"
)

func convertProject(p handler.Project, workloadDomain string) model.Project {
	result := model.Project{
		ID:   p.ID,
		Name: p.Name,
	}
	if !p.CreatedAt.IsZero() {
		result.CreatedAt = &p.CreatedAt
	}
	for _, e := range p.Environments {
		result.Environments = append(result.Environments, convertEnvironment(e, workloadDomain))
	}
	for _, d := range p.Databases {
		result.Databases = append(result.Databases, convertDatabase(d))
	}
	return result
}

func convertEnvironment(e handler.Environment, workloadDomain string) model.Environment {
	result := model.Environment{
		ID:         e.ID,
		Name:       e.Name,
		Namespace:  e.Namespace,
		Ephemeral:  e.Ephemeral,
		SyncStatus: model.SyncStatus(e.SyncStatus),
	}
	for _, si := range e.Services {
		result.Services = append(result.Services, convertServiceInstance(si, workloadDomain))
	}
	for _, di := range e.Databases {
		result.Databases = append(result.Databases, convertDatabaseInstance(di))
	}
	return result
}

func convertDetectedService(s handler.DetectedService) model.DetectedService {
	return model.DetectedService{
		Name:          s.Name,
		Language:      s.Provider,
		Framework:     s.Framework,
		StartCommand:  s.StartCommand,
		SuggestedPort: s.SuggestedPort,
	}
}

func convertScalingConfig(sc handler.ScalingConfig) model.ScalingConfig {
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

func convertServiceInstance(si handler.ServiceInstance, workloadDomain string) model.ServiceInstance {
	scaling := convertScalingConfig(si.Scaling)
	result := model.ServiceInstance{
		ID:          si.ID,
		Name:        si.Name,
		Environment: si.Environment,
		Image:       si.Image,
		ImageTag:    si.ImageTag,
		Ready:       si.Ready,
		Replicas:    si.Replicas,
		Scaling:     &scaling,
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

func convertDomain(d handler.Domain) *model.Domain {
	return &model.Domain{
		Hostname:  d.Hostname,
		Type:      model.DomainType(d.Type),
		DNSStatus: model.DNSStatus(d.DnsStatus),
		TLSStatus: model.TLSStatus(d.TlsStatus),
	}
}

func convertDnsCheck(d handler.DnsCheck) *model.DNSCheck {
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

func convertDeployment(d handler.Deployment) model.Deployment {
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

func convertDeploymentOp(d handler.DeployOp) model.DeployRun {
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

func convertGitHubRepository(r handler.GitHubRepository) model.GitHubRepository {
	return model.GitHubRepository{
		ID:            r.ID,
		Name:          r.Name,
		FullName:      r.FullName,
		HTMLURL:       r.HTMLURL,
		DefaultBranch: r.DefaultBranch,
		Private:       r.Private,
	}
}

func convertGitHubInstallation(i handler.GitHubInstallation) model.GitHubInstallation {
	return model.GitHubInstallation{
		ID:               i.ID,
		AccountLogin:     i.AccountLogin,
		AccountAvatarURL: i.AccountAvatarURL,
		AccountType:      model.GitHubAccountType(i.AccountType),
	}
}

func convertUser(u *handler.User) *model.User {
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

func convertWorkspace(ws *handler.Workspace) *model.Workspace {
	result := &model.Workspace{
		ID:        ws.ID,
		Name:      ws.Name,
		Personal:  ws.Personal,
		Suspended: ws.Suspended,
	}

	for _, m := range ws.Members {
		result.Members = append(result.Members, *convertWorkspaceMember(&m))
	}

	return result
}

func convertWorkspaceMember(m *handler.WorkspaceMember) *model.WorkspaceMember {
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

func convertEnvironmentResources(r handler.EnvironmentResources) model.EnvironmentResources {
	return model.EnvironmentResources{
		Tier: model.ResourceTier(r.Tier),
		Allocation: &model.ResourceAllocation{
			CPUMillicores: r.CpuMillicores,
			MemoryMb:      r.MemoryMB,
			DiskMb:        r.DiskMB,
		},
	}
}

func convertBillingSubscription(s *handler.BillingSubscription) *model.BillingSubscription {
	return &model.BillingSubscription{
		Plan:              model.Plan(s.Plan),
		Status:            model.SubscriptionStatus(s.Status),
		CurrentPeriodEnd:  s.CurrentPeriodEnd,
		CreditAmountCents: s.CreditAmountCents,
		CreditExpiry:      s.CreditExpiry,
		HasPaymentMethod:  s.HasPaymentMethod,
	}
}

func convertDatabase(d handler.Database) model.Database {
	return model.Database{
		Name:      d.Name,
		Version:   d.Version,
		Instances: d.Instances,
		Size:      d.Size,
	}
}

func convertDatabaseInstance(di handler.DatabaseInstance) model.DatabaseInstance {
	result := model.DatabaseInstance{
		Name:        di.Name,
		Environment: di.Environment,
		Ready:       di.Ready,
		Instances:   di.Instances,
		Version:     di.Version,
		Size:        di.Size,
	}
	if di.Volume != nil {
		result.Volume = &model.Volume{
			Name:          di.Volume.Name,
			Size:          di.Volume.Size,
			RequestedSize: di.Volume.RequestedSize,
			UsedBytes:     int(di.Volume.UsedBytes),
			CapacityBytes: int(di.Volume.CapacityBytes),
		}
	}
	return result
}

func convertDatabaseTable(t handler.DatabaseTable) model.DatabaseTable {
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

func convertDatabaseTableData(d *handler.DatabaseTableData) *model.DatabaseTableData {
	return &model.DatabaseTableData{
		Columns:            d.Columns,
		Rows:               d.Rows,
		TotalEstimatedRows: d.TotalEstimatedRows,
	}
}

func convertQueryResult(r *handler.QueryResult) *model.QueryResult {
	return &model.QueryResult{
		Columns:      r.Columns,
		Rows:         r.Rows,
		AffectedRows: r.AffectedRows,
	}
}
