package graphql

import (
	"github.com/zeitlos/lucity/services/gateway/graphql/model"
	"github.com/zeitlos/lucity/services/gateway/handler"
)

func convertProject(p handler.Project) model.Project {
	result := model.Project{
		ID:        p.ID,
		Name:      p.Name,
		SourceURL: p.SourceURL,
		CreatedAt: p.CreatedAt,
	}
	for _, e := range p.Environments {
		result.Environments = append(result.Environments, convertEnvironment(e))
	}
	for _, s := range p.Services {
		result.Services = append(result.Services, convertService(s))
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
		Name:   s.Name,
		Image:  s.Image,
		Public: s.Public,
	}
	if s.Port > 0 {
		port := s.Port
		svc.Port = &port
	}
	if s.Framework != "" {
		svc.Framework = &s.Framework
	}
	for _, si := range s.Instances {
		svc.Instances = append(svc.Instances, convertServiceInstance(si))
	}
	return svc
}

func convertDetectedService(s handler.DetectedService) model.DetectedService {
	return model.DetectedService{
		Name:          s.Name,
		Provider:      s.Provider,
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
	if si.ImageTag != "" {
		result.Deployment = &model.Deployment{
			ID:       si.ImageTag,
			ImageTag: si.ImageTag,
			Active:   true,
		}
	}
	return result
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
