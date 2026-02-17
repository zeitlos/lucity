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
	for _, ds := range e.Services {
		result.Services = append(result.Services, convertDeployedService(ds))
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
	return svc
}

func convertDeployedService(ds handler.DeployedService) model.DeployedService {
	return model.DeployedService{
		Name:     ds.Name,
		ImageTag: ds.ImageTag,
		Ready:    ds.Ready,
		Replicas: ds.Replicas,
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
