package graphql

import (
	"fmt"
	"log/slog"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/services/conductor/internal/api/graphql/model"
	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"
	"github.com/zeitlos/lucity/services/conductor/internal/conductor"
	"github.com/zeitlos/lucity/services/conductor/internal/hostname"
	"github.com/zeitlos/lucity/services/conductor/internal/planner"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

func convertService(service platform.Service) model.Service {
	result := model.Service{
		ID:                service.ID,
		Name:              service.Name,
		Status:            convertServiceStatus(service.Status),
		Replicas:          convertReplicaCount(service.Replicas),
		Port:              service.Port,
		PlatformEndpoints: service.Endpoints,
		SourceURL:         service.SourceURL,
		ContextPath:       service.ContextPath,
		Resources:         convertResources(service.Resources),
		Command:           service.Command,
		CreatedAt:         service.CreatedAt,
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
		ID:            deployment.ID,
		Image:         deployment.Image.String(),
		Commit:        deployment.Commit,
		CommitMessage: deployment.CommitMessage,
		Ref:           deployment.Ref,
		SourceURL:     deployment.SourceURL,
		ContextPath:   deployment.ContextPath,
		Command:       deployment.Command,
		BuildID:       deployment.BuildID,
		Status:        convertDeploymentStatus(deployment.Status),
		Replicas:      convertReplicaCount(deployment.Replicas),
		Resources:     convertResources(deployment.Resources),
		CreatedAt:     deployment.CreatedAt,
	}

	if deployment.Image.Digest != "" {
		result.ImageDigest = &deployment.Image.Digest
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
		ID:           e.ID,
		Name:         e.Name,
		ResourceTier: convertResourceTier(e.ResourceTier),
	}

	return result
}

func convertDetectedService(s planner.Plan) model.DetectedService {
	// Providers convention is [language, ..., framework] — e.g. ["node", "next"].
	// Single-entry slices carry only the language; framework stays empty.
	var language, framework string
	if len(s.Providers) > 0 {
		language = s.Providers[0]
	}
	if len(s.Providers) > 1 {
		framework = s.Providers[len(s.Providers)-1]
	}

	return model.DetectedService{
		Name:         s.Name,
		Language:     language,
		Framework:    framework,
		StartCommand: s.StartCommand,
		// SuggestedPort: buildjob.DetectedService no longer carries port info.
		// Left at zero until the new detector returns a port or we drop the field.
	}
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
		AccountLogin:     i.AccountLogin,
		AccountAvatarURL: i.AccountAvatarURL,
		AccountType:      convertGitHubAccountType(i.AccountType),
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
		result[i] = model.WorkspaceMembership{
			Workspace: m.Workspace,
			Role:      convertWorkspaceRole(m.Role),
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
	result := &model.WorkspaceMember{
		ID:    m.ID,
		Email: m.Email,
		Role:  convertWorkspaceRole(m.Role),
	}
	if m.Name != "" {
		result.Name = &m.Name
	}
	return result
}

func convertEnvironmentResources(r conductor.EnvironmentResources) model.EnvironmentResources {
	return model.EnvironmentResources{
		Tier: convertResourceTier(r.Tier),
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
		p := convertPlan(*s.Plan)
		plan = &p
	}
	return &model.BillingSubscription{
		Plan:              plan,
		Status:            convertSubscriptionStatus(s.Status),
		CurrentPeriodEnd:  s.CurrentPeriodEnd,
		CreditAmountCents: s.CreditAmountCents,
		CreditExpiry:      s.CreditExpiry,
		HasPaymentMethod:  s.HasPaymentMethod,
	}
}

func convertBuild(build conductor.Build) model.Build {
	return model.Build{
		ID:         build.ID,
		Status:     convertBuildStatus(build.Status),
		StartedAt:  build.StartedAt,
		FinishedAt: build.FinishedAt,
	}
}

func convertDatabase(d conductor.Database) model.Database {
	return model.Database{
		ID:        d.ID,
		Name:      d.Name,
		Version:   d.Version,
		Instances: d.Instances,
		Status:    convertDatabaseStatus(d.Status),
		Size:      d.Size.String(),
		CreatedAt: d.CreatedAt,
		Public:    d.PublicHost != "",
	}
}

func convertDatabaseCredentials(c conductor.DatabaseCredentials) model.DatabaseCredentials {
	return model.DatabaseCredentials{
		Type:     convertEndpointType(c.Type),
		Host:     c.Host,
		Port:     c.Port,
		Dbname:   c.DBName,
		User:     c.User,
		Password: c.Password,
		URI:      c.URI,
	}
}

func convertKeyValueStore(s conductor.KeyValueStore) model.KeyValueStore {
	return model.KeyValueStore{
		ID:        s.ID,
		Name:      s.Name,
		Version:   s.Version,
		Status:    convertDatabaseStatus(s.Status),
		Size:      s.Size.String(),
		CreatedAt: s.CreatedAt,
	}
}

func convertKeyValueStoreCredentials(c conductor.KeyValueStoreCredentials) model.KeyValueStoreCredentials {
	return model.KeyValueStoreCredentials{
		Type:     convertEndpointType(c.Type),
		Host:     c.Host,
		Port:     c.Port,
		Password: c.Password,
		URI:      c.URI,
	}
}

// --- Enum converters ----------------------------------------------------
//
// Each GraphQL enum gets a typed switch instead of a free-form string cast.
// Adding a new value to either the source or the model enum will fail a
// compile check if it's referenced here, or surface as a runtime warning
// if a value flows through without a case. Either way: nothing silent.

func convertServiceStatus(status platform.ServiceStatus) model.ServiceStatus {
	switch status {
	case platform.ServiceHealthy:
		return model.ServiceStatusHealthy
	case platform.ServiceDegraded:
		return model.ServiceStatusDegraded
	case platform.ServiceDeploying:
		return model.ServiceStatusDeploying
	case platform.ServiceFailed:
		return model.ServiceStatusFailed
	case platform.ServiceStopped:
		return model.ServiceStatusStopped
	case platform.ServiceBuilding:
		return model.ServiceStatusBuilding
	}

	slog.Warn("unknown service status", "status", status)

	return model.ServiceStatusFailed
}

func convertDeploymentStatus(status platform.DeploymentStatus) model.DeploymentStatus {
	switch status {
	case platform.DeploymentDeploying:
		return model.DeploymentStatusDeploying
	case platform.DeploymentActive:
		return model.DeploymentStatusActive
	case platform.DeploymentSuperseded:
		return model.DeploymentStatusSuperseded
	case platform.DeploymentFailed:
		return model.DeploymentStatusFailed
	}

	slog.Warn("unknown deployment status", "status", status)

	return model.DeploymentStatusFailed
}

func convertBuildStatus(status buildjob.Status) model.BuildStatus {
	switch status {
	case buildjob.StatusQueued:
		return model.BuildStatusQueued
	case buildjob.StatusRunning:
		return model.BuildStatusRunning
	case buildjob.StatusSucceeded:
		return model.BuildStatusSucceeded
	case buildjob.StatusFailed:
		return model.BuildStatusFailed
	case buildjob.StatusCancelling:
		return model.BuildStatusRunning
	case buildjob.StatusCancelled:
		return model.BuildStatusCancelled
	}

	slog.Warn("unknown build status", "status", status)

	return model.BuildStatusFailed
}

func convertDatabaseStatus(status platform.DatabaseStatus) model.DatabaseStatus {
	switch status {
	case platform.DatabaseHealthy:
		return model.DatabaseStatusHealthy
	case platform.DatabaseDegraded:
		return model.DatabaseStatusDegraded
	case platform.DatabaseFailed:
		return model.DatabaseStatusFailed
	case platform.DatabasePending:
		return model.DatabaseStatusPending
	case platform.DatabaseStopped:
		return model.DatabaseStatusStopped
	}

	slog.Warn("unknown database status", "status", status)

	return model.DatabaseStatusFailed
}

func convertProtocol(protocol platform.Protocol) model.Protocol {
	switch protocol {
	case platform.ProtocolHTTP:
		return model.ProtocolHTTP
	case platform.ProtocolHTTPS:
		return model.ProtocolHTTPS
	case platform.ProtocolTCP:
		return model.ProtocolTCP
	}

	slog.Warn("unknown protocol", "protocol", protocol)

	return model.ProtocolTCP
}

func convertEndpoint(endpoint conductor.Endpoint) model.Endpoint {
	records := make([]model.DNSRecord, 0, len(endpoint.RequiredDNSRecords))

	for _, record := range endpoint.RequiredDNSRecords {
		records = append(records, convertDNSRecord(record))
	}

	return model.Endpoint{
		Host:     endpoint.Host,
		Port:     endpoint.Port,
		Protocol: convertProtocol(endpoint.Protocol),
		Type:     convertEndpointType(endpoint.Type),
		DNS: &model.DNSState{
			Status:          convertDNSStatus(endpoint.DNSStatus),
			RequiredRecords: records,
		},
		TLS: convertTLSStatus(endpoint.TLSStatus),
	}
}

func convertEndpointType(t conductor.EndpointType) model.EndpointType {
	switch t {
	case conductor.InternalEndpoint:
		return model.EndpointTypeInternal
	case conductor.PlatformEndpoint:
		return model.EndpointTypePlatform
	case conductor.CustomDomainEndpoint:
		return model.EndpointTypeCustom
	}

	slog.Warn("unknown endpoint type", "type", t)

	return model.EndpointTypeInternal
}

func convertDNSStatus(status hostname.DNSStatus) model.DNSStatus {
	switch status {
	case hostname.DNSValid:
		return model.DNSStatusValid
	case hostname.DNSPending:
		return model.DNSStatusPending
	case hostname.DNSMisconfigured:
		return model.DNSStatusMisconfigured
	case hostname.DNSError:
		return model.DNSStatusError
	}

	slog.Warn("unknown dns status", "status", status)

	return model.DNSStatusError
}

func convertTLSStatus(status hostname.TLSStatus) model.TLSStatus {
	switch status {
	case hostname.TLSNone:
		return model.TLSStatusNone
	case hostname.TLSProvisioning:
		return model.TLSStatusProvisioning
	case hostname.TLSActive:
		return model.TLSStatusActive
	case hostname.TLSError:
		return model.TLSStatusError
	}

	slog.Warn("unknown tls status", "status", status)

	return model.TLSStatusError
}

func convertDNSRecord(record hostname.DNSRecord) model.DNSRecord {
	return model.DNSRecord{
		Type:  convertDNSRecordType(record.Type),
		Host:  record.Host,
		Value: record.Value,
	}
}

func convertDNSRecordType(t hostname.RecordType) model.DNSRecordType {
	switch t {
	case hostname.TXT:
		return model.DNSRecordTypeTxt
	case hostname.CNAME:
		return model.DNSRecordTypeCname
	case hostname.A:
		return model.DNSRecordTypeA
	}

	slog.Warn("unknown dns record type", "type", t)

	return model.DNSRecordTypeTxt
}

func convertResourceTier(tier platform.ResourceTier) model.ResourceTier {
	switch tier {
	case platform.EcoTier:
		return model.ResourceTierEco
	case platform.ProductionTier:
		return model.ResourceTierProduction
	}

	slog.Warn("unknown resource tier", "tier", tier)

	return model.ResourceTierEco
}

func parseResourceTier(tier model.ResourceTier) (platform.ResourceTier, error) {
	switch tier {
	case model.ResourceTierEco:
		return platform.EcoTier, nil
	case model.ResourceTierProduction:
		return platform.ProductionTier, nil
	}

	return "", fmt.Errorf("invalid resource tier %q", tier)
}

func convertGitHubAccountType(accountType string) model.GitHubAccountType {
	switch accountType {
	case "ORGANIZATION":
		return model.GitHubAccountTypeOrganization
	case "USER":
		return model.GitHubAccountTypeUser
	}

	slog.Warn("unknown github account type", "type", accountType)

	return model.GitHubAccountTypeUser
}

func convertSubscriptionStatus(status string) model.SubscriptionStatus {
	switch status {
	case "ACTIVE":
		return model.SubscriptionStatusActive
	case "PAST_DUE":
		return model.SubscriptionStatusPastDue
	case "CANCELED":
		return model.SubscriptionStatusCanceled
	case "INCOMPLETE":
		return model.SubscriptionStatusIncomplete
	case "TRIALING":
		return model.SubscriptionStatusTrialing
	}

	slog.Warn("unknown subscription status", "status", status)

	return model.SubscriptionStatusIncomplete
}

func convertPlan(plan string) model.Plan {
	switch plan {
	case "HOBBY":
		return model.PlanHobby
	case "PRO":
		return model.PlanPro
	}

	slog.Warn("unknown plan", "plan", plan)

	return model.PlanHobby
}

func convertWorkspaceRole(role auth.WorkspaceRole) model.WorkspaceRole {
	switch role {
	case auth.WorkspaceRoleUser:
		return model.WorkspaceRoleUser
	case auth.WorkspaceRoleAdmin:
		return model.WorkspaceRoleAdmin
	}

	slog.Warn("unknown workspace role", "role", role)

	return model.WorkspaceRoleUser
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
