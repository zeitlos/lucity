package graphql

import (
	"context"

	"github.com/zeitlos/lucity/services/conductor/internal/api/graphql/model"
)

// CreateAPIKey is the resolver for the createApiKey field.
func (r *mutationResolver) CreateAPIKey(ctx context.Context, input model.CreateAPIKeyInput) (*model.CreatedAPIKey, error) {
	created, err := r.Conductor.CreateAPIKey(ctx, input.Name, convertModelWorkspaceRole(input.Role))
	if err != nil {
		return nil, err
	}
	return convertCreatedAPIKey(created), nil
}

// RevokeAPIKey is the resolver for the revokeApiKey field.
func (r *mutationResolver) RevokeAPIKey(ctx context.Context, id string) (bool, error) {
	if err := r.Conductor.RevokeAPIKey(ctx, id); err != nil {
		return false, err
	}
	return true, nil
}

// APIKeys is the resolver for the apiKeys field.
func (r *queryResolver) APIKeys(ctx context.Context) ([]model.APIKey, error) {
	keys, err := r.Conductor.APIKeys(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]model.APIKey, 0, len(keys))
	for i := range keys {
		out = append(out, *convertAPIKey(&keys[i]))
	}
	return out, nil
}
