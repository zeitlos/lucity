package graphql

import (
	"fmt"
	"log/slog"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/services/conductor/internal/api/graphql/model"
	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"
	"github.com/zeitlos/lucity/services/conductor/internal/conductor"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/deployjob"
	"github.com/zeitlos/lucity/services/conductor/internal/hostname"
	"github.com/zeitlos/lucity/services/conductor/internal/metrics"
	"github.com/zeitlos/lucity/services/conductor/internal/planner"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
	"github.com/zeitlos/lucity/services/conductor/internal/scanjob"
	"github.com/zeitlos/lucity/services/conductor/internal/vulnerabilities"
)

func convertMetricWindow(window model.MetricWindow) (metrics.Window, error) {
	switch window {
	case model.MetricWindowLast1h:
		return metrics.Window1h, nil
	case model.MetricWindowLast6h:
		return metrics.Window6h, nil
	case model.MetricWindowLast24h:
		return metrics.Window24h, nil
	case model.MetricWindowLast7d:
		return metrics.Window7d, nil
	case model.MetricWindowLast30d:
		return metrics.Window30d, nil
	default:
		return "", fmt.Errorf("unknown metric window %q", window)
	}
}

func convertMetricSeries(series []metrics.Series) []model.MetricSeries {
	out := make([]model.MetricSeries, 0, len(series))

	for _, s := range series {
		var replica *string
		if s.Replica != "" {
			value := s.Replica
			replica = &value
		}

		out = append(out, model.MetricSeries{
			Metric:  metricKindToModel(s.Kind),
			Unit:    metricUnit(s.Kind),
			Replica: replica,
			Points:  convertMetricPoints(s.Points),
		})
	}

	return out
}

func metricKindToModel(kind metrics.Kind) model.ResourceMetric {
	switch kind {
	case metrics.KindCPUUsage:
		return model.ResourceMetricCPUUsage
	case metrics.KindMemoryUsage:
		return model.ResourceMetricMemoryUsage
	default:
		return model.ResourceMetricStorageUsed
	}
}

func metricUnit(kind metrics.Kind) model.MetricUnit {
	if kind == metrics.KindCPUUsage {
		return model.MetricUnitCores
	}
	return model.MetricUnitBytes
}

func serviceMetricKinds(requested []model.ResourceMetric) []metrics.Kind {
	kinds := make([]metrics.Kind, 0, len(requested))

	for _, metric := range requested {
		if metric == model.ResourceMetricCPUUsage || metric == model.ResourceMetricMemoryUsage {
			kinds = append(kinds, metricModelToKind(metric))
		}
	}

	return kinds
}

func metricModelToKind(metric model.ResourceMetric) metrics.Kind {
	switch metric {
	case model.ResourceMetricCPUUsage:
		return metrics.KindCPUUsage
	case model.ResourceMetricMemoryUsage:
		return metrics.KindMemoryUsage
	default:
		return metrics.KindStorageUsed
	}
}

func convertMetricPoints(points []metrics.Point) []model.MetricPoint {
	out := make([]model.MetricPoint, 0, len(points))

	for _, p := range points {
		out = append(out, model.MetricPoint{
			Timestamp: p.Timestamp,
			Value:     p.Value,
		})
	}

	return out
}

func convertService(service platform.Service) model.Service {
	result := model.Service{
		ID:              service.ID,
		Name:            service.Name,
		Status:          convertServiceStatus(service.Status),
		Replicas:        convertReplicaCount(service.Replicas),
		Port:            service.Port,
		PlatformService: service,
		SourceURL:       service.SourceURL,
		AutoDeploy:      service.AutoDeploy,
		CiDeploy:        service.CIDeploy,
		ContextPath:     service.ContextPath,
		Resources:       convertResources(service.Resources),
		Command:         service.Command,
		CreatedAt:       service.CreatedAt,
	}

	if service.Branch != "" {
		result.Branch = &service.Branch
	}

	if service.Autoscaling != nil {
		result.Autoscaling = convertAutoscaling(*service.Autoscaling)
	}

	if service.HealthCheck != nil {
		result.HealthCheck = convertHealthCheck(*service.HealthCheck)
	}

	if service.SecurityContext.RunAsUser != nil {
		user := int(*service.SecurityContext.RunAsUser)
		result.User = &user
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

	if deployment.Rollout != nil {
		result.Rollout = convertRollout(*deployment.Rollout)
	}

	return result
}

func convertRollout(rollout platform.Rollout) *model.Rollout {
	result := model.Rollout{
		Status:    convertRolloutStatus(rollout.Status),
		Restarts:  rollout.Restarts,
		StartedAt: rollout.StartedAt,
	}

	if rollout.Reason != platform.RolloutReasonNone {
		reason := convertRolloutReason(rollout.Reason)
		result.Reason = &reason
	}

	if rollout.Message != "" {
		result.Message = &rollout.Message
	}

	return &result
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

// toInt64Ptr widens a nullable GraphQL Int (*int) to the *int64 the domain uses.
func toInt64Ptr(v *int) *int64 {
	if v == nil {
		return nil
	}

	n := int64(*v)
	return &n
}

func convertHealthCheck(healthCheck platform.HealthCheck) *model.HealthCheck {
	return &model.HealthCheck{
		Path:                    healthCheck.Path,
		Port:                    healthCheck.Port,
		InitialDelaySeconds:     healthCheck.InitialDelaySeconds,
		PeriodSeconds:           healthCheck.PeriodSeconds,
		TimeoutSeconds:          healthCheck.TimeoutSeconds,
		FailureThreshold:        healthCheck.FailureThreshold,
		StartupFailureThreshold: healthCheck.StartupFailureThreshold,
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
		ContextPath:  s.ContextPath,
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

func convertRelease(release conductor.Release) model.Release {
	result := model.Release{
		ID:        release.ID,
		Status:    convertReleaseStatus(release.Status),
		Trigger:   convertReleaseTrigger(release.Trigger),
		CreatedAt: release.CreatedAt,
	}

	if release.Source != nil {
		source := convertGitSource(*release.Source)
		result.Source = &source
	}

	if release.Build != nil {
		build := convertBuild(*release.Build)
		result.Build = &build
	}

	if release.Deploy != nil {
		deploy := convertDeploy(*release.Deploy, release.Build)
		result.Deploy = &deploy
	}

	if release.Scan != nil {
		scan := convertScan(*release.Scan)
		result.Scan = &scan
	}

	if release.Deployment != nil {
		deployment := convertDeployment(*release.Deployment)
		result.Deployment = &deployment
	}

	return result
}

func convertScan(scan conductor.Scan) model.Scan {
	return model.Scan{
		ID:            scan.ID,
		Status:        convertScanStatus(scan),
		FindingsCount: scan.FindingsCount,
		VerifiedCount: scan.VerifiedCount,
		StartedAt:     scan.StartedAt,
		FinishedAt:    scan.FinishedAt,
	}
}

func convertScanStatus(scan conductor.Scan) model.ScanStatus {
	switch scan.Status {
	case scanjob.StatusQueued:
		return model.ScanStatusQueued
	case scanjob.StatusRunning:
		return model.ScanStatusRunning
	case scanjob.StatusSucceeded:
		if scan.FindingsCount != nil && *scan.FindingsCount > 0 {
			return model.ScanStatusFindings
		}

		return model.ScanStatusClean
	case scanjob.StatusFailed:
		return model.ScanStatusFailed
	}

	slog.Warn("unknown scan status", "status", scan.Status)

	return model.ScanStatusFailed
}

func optional(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}

func convertSecretScanReport(report conductor.SecretScanReport) model.SecretScanReport {
	findings := make([]model.SecretFinding, 0, len(report.Findings))

	for _, finding := range report.Findings {
		findings = append(findings, model.SecretFinding{
			Rule:     finding.Rule,
			File:     finding.File,
			Line:     finding.Line,
			Commit:   finding.Commit,
			Secret:   finding.Secret,
			Author:   optional(finding.Author),
			URL:      optional(finding.URL),
			Verified: finding.Verified,
		})
	}

	return model.SecretScanReport{
		Commit:    report.Commit,
		ScannedAt: report.ScannedAt,
		Findings:  findings,
	}
}

func convertVulnerabilityReport(report conductor.VulnerabilityReport) model.VulnerabilityReport {
	vulnerabilities := make([]model.Vulnerability, 0, len(report.Vulnerabilities))

	for _, vulnerability := range report.Vulnerabilities {
		packages := make([]model.VulnerablePackage, 0, len(vulnerability.Packages))

		for _, pkg := range vulnerability.Packages {
			packages = append(packages, model.VulnerablePackage{
				Name:             pkg.Name,
				InstalledVersion: pkg.InstalledVersion,
				FixedVersion:     optional(pkg.FixedVersion),
				Path:             optional(pkg.Path),
			})
		}

		vulnerabilities = append(vulnerabilities, model.Vulnerability{
			ID:          vulnerability.ID,
			Severity:    convertVulnerabilitySeverity(vulnerability.Severity),
			Source:      convertVulnerabilitySource(vulnerability.Source),
			Title:       optional(vulnerability.Title),
			Description: optional(vulnerability.Description),
			Reference:   optional(vulnerability.Reference),
			Packages:    packages,
		})
	}

	return model.VulnerabilityReport{
		Image: report.Image,
		Summary: &model.VulnerabilitySummary{
			Critical: report.Summary.Critical,
			High:     report.Summary.High,
			Medium:   report.Summary.Medium,
			Low:      report.Summary.Low,
			Unknown:  report.Summary.Unknown,
			Total:    report.Summary.Total,
		},
		Vulnerabilities: vulnerabilities,
	}
}

func convertVulnerabilitySeverity(severity conductor.VulnerabilitySeverity) model.VulnerabilitySeverity {
	switch severity {
	case vulnerabilities.SeverityCritical:
		return model.VulnerabilitySeverityCritical
	case vulnerabilities.SeverityHigh:
		return model.VulnerabilitySeverityHigh
	case vulnerabilities.SeverityMedium:
		return model.VulnerabilitySeverityMedium
	case vulnerabilities.SeverityLow:
		return model.VulnerabilitySeverityLow
	default:
		return model.VulnerabilitySeverityUnknown
	}
}

func convertVulnerabilitySource(source vulnerabilities.Source) model.VulnerabilitySource {
	switch source {
	case vulnerabilities.SourceOperatingSystem:
		return model.VulnerabilitySourceOperatingSystem
	case vulnerabilities.SourceApplication:
		return model.VulnerabilitySourceApplication
	default:
		return model.VulnerabilitySourceUnknown
	}
}

func convertDeploy(deploy conductor.Deploy, build *conductor.Build) model.Deploy {
	return model.Deploy{
		ID:         deploy.ID,
		Status:     convertDeployStatus(deploy.Status, build),
		StartedAt:  deploy.StartedAt,
		FinishedAt: deploy.FinishedAt,
	}
}

func convertDeployStatus(status deployjob.Status, build *conductor.Build) model.DeployStatus {
	if build != nil {
		switch build.Status {
		case buildjob.StatusFailed, buildjob.StatusCancelled, buildjob.StatusCancelling:
			return model.DeployStatusSkipped
		case buildjob.StatusQueued, buildjob.StatusRunning:
			if status == deployjob.StatusFailed {
				return model.DeployStatusFailed
			}

			return model.DeployStatusQueued
		}
	}

	switch status {
	case deployjob.StatusQueued:
		return model.DeployStatusQueued
	case deployjob.StatusRunning:
		return model.DeployStatusRunning
	case deployjob.StatusSucceeded:
		return model.DeployStatusSucceeded
	case deployjob.StatusFailed:
		return model.DeployStatusFailed
	}

	slog.Warn("unknown deploy status", "status", status)

	return model.DeployStatusFailed
}

func convertGitSource(source conductor.GitSource) model.GitSource {
	commit := model.Commit{
		Sha:     source.Commit.SHA,
		Message: source.Commit.Message,
	}

	if source.Commit.URL != "" {
		commit.URL = &source.Commit.URL
	}

	return model.GitSource{
		Provider:    convertSourceProvider(source.Provider),
		Repository:  source.Repository,
		URL:         source.URL,
		Ref:         source.Ref,
		ContextPath: source.ContextPath,
		Commit:      &commit,
	}
}

func convertReleaseTrigger(trigger conductor.ReleaseTrigger) *model.ReleaseTrigger {
	result := &model.ReleaseTrigger{
		Kind: convertReleaseTriggerKind(trigger.Kind),
	}

	if trigger.Actor != "" {
		result.Actor = &trigger.Actor
	}

	return result
}

func convertReleaseStatus(status conductor.ReleaseStatus) model.ReleaseStatus {
	switch status {
	case conductor.ReleaseQueued:
		return model.ReleaseStatusQueued
	case conductor.ReleaseBuilding:
		return model.ReleaseStatusBuilding
	case conductor.ReleaseDeploying:
		return model.ReleaseStatusDeploying
	case conductor.ReleaseLive:
		return model.ReleaseStatusLive
	case conductor.ReleaseFailed:
		return model.ReleaseStatusFailed
	case conductor.ReleaseCancelled:
		return model.ReleaseStatusCancelled
	case conductor.ReleaseSuperseded:
		return model.ReleaseStatusSuperseded
	}

	slog.Warn("unknown release status", "status", status)

	return model.ReleaseStatusFailed
}

func convertSourceProvider(provider conductor.SourceProvider) model.SourceProvider {
	switch provider {
	case conductor.ProviderGitHub:
		return model.SourceProviderGithub
	case conductor.ProviderGitLab:
		return model.SourceProviderGitlab
	case conductor.ProviderBitbucket:
		return model.SourceProviderBitbucket
	}

	return model.SourceProviderGithub
}

func convertReleaseTriggerKind(kind deployer.TriggerKind) model.ReleaseTriggerKind {
	switch kind {
	case deployer.TriggerManual:
		return model.ReleaseTriggerKindManual
	case deployer.TriggerRollback:
		return model.ReleaseTriggerKindRollback
	case deployer.TriggerPromotion:
		return model.ReleaseTriggerKindPromotion
	case deployer.TriggerPush:
		return model.ReleaseTriggerKindPush
	}

	return model.ReleaseTriggerKindManual
}

func convertDatabase(d conductor.Database) model.Database {
	return model.Database{
		ID:        d.ID,
		Name:      d.Name,
		Version:   d.Version,
		Instances: d.Instances,
		Status:    convertDatabaseStatus(d.Status),
		Size:      d.Size.String(),
		Resources: convertResources(d.Resources),
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

func convertVolume(v conductor.Volume) model.Volume {
	volume := model.Volume{
		ID:   v.ID,
		Name: v.Name,
		Size: v.Size.String(),
	}

	if v.Mount != nil {
		volume.Mount = &model.Mount{
			Service: v.Mount.Service,
			Path:    v.Mount.Path,
		}
	}

	return volume
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

func convertBucket(b conductor.Bucket) model.Bucket {
	return model.Bucket{
		ID:             b.ID,
		Name:           b.Name,
		Region:         b.Region,
		Endpoint:       b.Endpoint,
		PublicEndpoint: b.PublicEndpoint,
		Status:         model.BucketStatusReady,
		SizeBytes:      int(b.SizeBytes),
		ObjectCount:    int(b.ObjectCount),
		Public:         b.Public,
		CreatedAt:      b.CreatedAt,
	}
}

func convertBucketCredentials(c conductor.BucketCredentials) model.BucketCredentials {
	return model.BucketCredentials{
		Endpoint:        c.Endpoint,
		Region:          c.Region,
		Bucket:          c.Bucket,
		AccessKeyID:     c.AccessKeyID,
		SecretAccessKey: c.SecretAccessKey,
	}
}

func convertBucketObjectListing(l conductor.BucketObjectListing) model.BucketObjectListing {
	folders := make([]model.BucketFolder, 0, len(l.Folders))

	for _, folder := range l.Folders {
		folders = append(folders, model.BucketFolder{Prefix: folder.Prefix})
	}

	objects := make([]model.BucketObject, 0, len(l.Objects))

	for _, object := range l.Objects {
		objects = append(objects, model.BucketObject{
			Key:          object.Key,
			Size:         int(object.Size),
			LastModified: object.LastModified,
		})
	}

	return model.BucketObjectListing{
		Prefix:  l.Prefix,
		Folders: folders,
		Objects: objects,
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

func convertRolloutStatus(status platform.RolloutStatus) model.RolloutStatus {
	switch status {
	case platform.RolloutProgressing:
		return model.RolloutStatusProgressing
	case platform.RolloutReady:
		return model.RolloutStatusReady
	case platform.RolloutDegraded:
		return model.RolloutStatusDegraded
	case platform.RolloutFailed:
		return model.RolloutStatusFailed
	case platform.RolloutSuperseded:
		return model.RolloutStatusSuperseded
	}

	slog.Warn("unknown rollout status", "status", status)

	return model.RolloutStatusFailed
}

func convertRolloutReason(reason platform.RolloutReason) model.RolloutReason {
	switch reason {
	case platform.RolloutReasonCrashLoop:
		return model.RolloutReasonCrashLoop
	case platform.RolloutReasonOOMKilled:
		return model.RolloutReasonOomKilled
	case platform.RolloutReasonImagePullFailed:
		return model.RolloutReasonImagePullFailed
	case platform.RolloutReasonConfigError:
		return model.RolloutReasonConfigError
	case platform.RolloutReasonQuotaExceeded:
		return model.RolloutReasonQuotaExceeded
	case platform.RolloutReasonUnschedulable:
		return model.RolloutReasonUnschedulable
	case platform.RolloutReasonNotReady:
		return model.RolloutReasonNotReady
	case platform.RolloutReasonDeadlineExceeded:
		return model.RolloutReasonDeadlineExceeded
	}

	slog.Warn("unknown rollout reason", "reason", reason)

	return model.RolloutReasonNotReady
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
	case platform.DatabaseUpdating:
		return model.DatabaseStatusUpdating
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

func convertModelWorkspaceRole(r model.WorkspaceRole) auth.WorkspaceRole {
	switch r {
	case model.WorkspaceRoleAdmin:
		return auth.WorkspaceRoleAdmin
	}
	return auth.WorkspaceRoleUser
}

func convertAPIToken(t *conductor.APIToken) *model.APIToken {
	return &model.APIToken{
		ID:        t.ID,
		Name:      t.Name,
		Role:      convertWorkspaceRole(t.Role),
		CreatedAt: t.CreatedAt,
	}
}

func convertCreatedAPIToken(c *conductor.CreatedAPIToken) *model.CreatedAPIToken {
	return &model.CreatedAPIToken{
		APIToken: convertAPIToken(&c.APIToken),
		Token:    c.Token,
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

func convertVariable(variable conductor.Variable) model.Variable {
	return model.Variable{
		ID:     variable.ID,
		Key:    variable.Name,
		Source: convertVariableSource(variable.Source),
	}
}

func convertVariableSource(source platform.VariableSource) model.VariableSource {
	switch {
	case source.Database != nil:
		return model.DatabaseSource{ID: *source.Database, Name: source.Database.Name}
	case source.KeyValueStore != nil:
		return model.KeyValueStoreSource{ID: *source.KeyValueStore, Name: source.KeyValueStore.Name}
	case source.Bucket != nil:
		return model.BucketSource{ID: *source.Bucket, Name: source.Bucket.Name}
	default:
		return model.SharedSource{}
	}
}

func convertSharedVariable(variable conductor.SharedVariable) model.SharedVariable {
	return model.SharedVariable{
		Key:   variable.Key,
		Value: variable.Value,
	}
}

func convertServiceVariable(variable conductor.ServiceVariable) model.ServiceVariable {
	result := model.ServiceVariable{Key: variable.Key}

	if variable.Ref != nil {
		result.Ref = variable.Ref
	} else {
		value := variable.Value
		result.Value = &value
	}

	return result
}

func convertImageSearchResult(image conductor.ImageSearchResult) model.ImageSearchResult {
	return model.ImageSearchResult{
		Name:        image.Name,
		Description: image.Description,
		StarCount:   image.StarCount,
		PullCount:   int(image.PullCount),
		Official:    image.Official,
	}
}

func convertUsageSummary(summary conductor.UsageSummaryResult) model.UsageSummary {
	return model.UsageSummary{
		ResourceCostCents:   summary.ResourceCostCents,
		CreditsCents:        summary.CreditsCents,
		EstimatedTotalCents: summary.EstimatedTotalCents,
	}
}

func convertServiceLogEntry(entry platform.LogEntry) *model.ServiceLogEntry {
	return &model.ServiceLogEntry{
		Line: entry.Line,
		Pod:  entry.Pod,
	}
}
