package logto

import (
	"context"
	"errors"

	"github.com/zeitlos/lucity/services/conductor/internal/directory"
)

func (p *Provider) Projects(ctx, tenantID string) ([]directory.Project, error) {
	return nil, errors.New("not implemented")
}

func (p *Provider) ProjectsForUser(ctx context.Context, userID string) ([]directory.Workspace, error) {
	return nil, errors.New("not implemented")
}

func (p *Provider) Project(ctx, id string) (*directory.Project, error) {
	return nil, errors.New("not implemented")
}
