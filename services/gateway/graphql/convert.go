package graphql

import (
	"fmt"
	"strings"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/services/gateway/graphql/model"
	"github.com/zeitlos/lucity/services/gateway/handler"
)

func convertProject(p handler.Project, workloadDomain string) model.Project {
	result := model.Project{
		ID:        p.ID,
		Name:      p.Name,
		CreatedAt: p.CreatedAt,
	}
	for _, e := range p.Environments {
		result.Environments = append(result.Environments, convertEnvironment(e, workloadDomain))
	}
	for _, s := range p.Services {
		result.Services = append(result.Services, convertService(s, workloadDomain))
	}
	for _, d := range p.Databases {
		result.Databases = append(result.Databases, convertDatabase(d))
	}
	for _, d := range p.InitialDeploys {
		op := convertDeploymentOp(d)
		result.InitialDeploys = append(result.InitialDeploys, op)
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

func convertService(s handler.Service, workloadDomain string) model.Service {
	svc := model.Service{
		Name:  s.Name,
		Image: s.Image,
	}
	if s.Port > 0 {
		port := s.Port
		svc.Port = &port
	}
	if s.Framework != "" {
		svc.Framework = &s.Framework
	}
	if s.SourceURL != "" {
		svc.SourceURL = &s.SourceURL
	}
	if s.ContextPath != "" {
		svc.ContextPath = &s.ContextPath
	}
	for _, si := range s.Instances {
		svc.Instances = append(svc.Instances, convertServiceInstance(si, workloadDomain))
	}
	return svc
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

func convertBuild(b handler.Build) model.Build {
	build := model.Build{
		ID:    b.ID,
		Phase: model.BuildPhase(b.Phase),
	}
	if b.ImageRef != "" {
		build.ImageRef = &b.ImageRef
	}
	if b.Digest != "" {
		build.Digest = &b.Digest
	}
	if b.Error != "" {
		build.Error = &b.Error
	}
	return build
}

func convertServiceInstance(si handler.ServiceInstance, workloadDomain string) model.ServiceInstance {
	result := model.ServiceInstance{
		Name:        si.Name,
		Environment: si.Environment,
		ImageTag:    si.ImageTag,
		Ready:       si.Ready,
		Replicas:    si.Replicas,
	}

	// Convert domains with type derived from workload domain suffix
	for _, hostname := range si.Domains {
		domainType := model.DomainTypeCustom
		dnsStatus := model.DNSStatusPending
		if strings.HasSuffix(hostname, "."+workloadDomain) {
			domainType = model.DomainTypePlatform
			dnsStatus = model.DNSStatusValid
		}
		result.Domains = append(result.Domains, model.Domain{
			Hostname:  hostname,
			Type:      domainType,
			DNSStatus: dnsStatus,
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

func convertWorkspace(ws *handler.Workspace, githubAppSlug string) *model.Workspace {
	result := &model.Workspace{
		ID:           ws.ID,
		Name:         ws.Name,
		Personal:     ws.Personal,
		GithubLinked: ws.GithubInstallationID > 0,
	}

	if githubAppSlug != "" && !result.GithubLinked {
		url := fmt.Sprintf("https://github.com/apps/%s/installations/new", githubAppSlug)
		result.GithubInstallURL = &url
	}

	if ws.GithubAccountLogin != "" {
		result.GithubAccountLogin = &ws.GithubAccountLogin
		result.GithubAccountAvatarURL = &ws.GithubAccountAvatarURL
		accountType := model.GitHubAccountType(strings.ToUpper(ws.GithubAccountType))
		result.GithubAccountType = &accountType
		id := fmt.Sprintf("%d", ws.GithubInstallationID)
		result.GithubInstallationID = &id
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
