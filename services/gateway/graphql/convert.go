package graphql

import (
	"github.com/zeitlos/lucity/services/gateway/graphql/model"
	"github.com/zeitlos/lucity/services/gateway/handler"
)

func convertProject(p handler.Project) model.Project {
	result := model.Project{
		ID:        p.ID,
		Name:      p.Name,
		CreatedAt: p.CreatedAt,
	}
	for _, e := range p.Environments {
		result.Environments = append(result.Environments, convertEnvironment(e))
	}
	for _, s := range p.Services {
		result.Services = append(result.Services, convertService(s))
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

func convertEnvironment(e handler.Environment) model.Environment {
	result := model.Environment{
		ID:         e.ID,
		Name:       e.Name,
		Namespace:  e.Namespace,
		Ephemeral:  e.Ephemeral,
		SyncStatus: model.SyncStatus(e.SyncStatus),
	}
	for _, si := range e.Services {
		result.Services = append(result.Services, convertServiceInstance(si))
	}
	return result
}

func convertService(s handler.Service) model.Service {
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
		svc.Instances = append(svc.Instances, convertServiceInstance(si))
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

func convertServiceInstance(si handler.ServiceInstance) model.ServiceInstance {
	result := model.ServiceInstance{
		Name:        si.Name,
		Environment: si.Environment,
		ImageTag:    si.ImageTag,
		Ready:       si.Ready,
		Replicas:    si.Replicas,
	}
	if si.Host != "" {
		result.Host = &si.Host
	}

	// Convert deployment history
	for _, d := range si.Deployments {
		result.Deployments = append(result.Deployments, convertDeployment(d))
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

func convertUser(u *handler.User) *model.User {
	if u == nil {
		return nil
	}
	user := &model.User{
		Login:     u.Login,
		AvatarURL: u.AvatarURL,
	}
	if u.Name != "" {
		user.Name = &u.Name
	}
	if u.Email != "" {
		user.Email = &u.Email
	}
	return user
}

func convertDatabase(d handler.Database) model.Database {
	return model.Database{
		Name:      d.Name,
		Version:   d.Version,
		Instances: d.Instances,
		Size:      d.Size,
	}
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
