package kubernetes

import (
	"context"
	"log/slog"

	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/pkg/usererror"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const sharedVariablesLabel = "lucity.dev/shared-variables"

func (c *Client) Environments(ctx context.Context, projectID platform.ProjectID) ([]platform.Environment, error) {
	selector := labels.SelectorFromSet(labels.Set{
		workspaceLabel: projectID.Workspace,
		projectLabel:   projectID.Name,
	})

	namespaces, err := c.namespacesMatching(ctx, selector, projectIndex, projectKey(projectID.Workspace, projectID.Name))

	if err != nil {
		return nil, err
	}

	return toEnvironments(namespaces), nil
}

func (c *Client) Environment(ctx context.Context, id platform.EnvironmentID) (*platform.Environment, error) {
	selector := labels.SelectorFromSet(labels.Set{
		workspaceLabel:   id.Workspace,
		projectLabel:     id.Project,
		environmentLabel: id.Name,
	})

	namespaces, ok := c.cachedNamespaces(selector, projectIndex, projectKey(id.Workspace, id.Project))

	if !ok || len(namespaces) == 0 {
		live, err := c.liveNamespaces(ctx, selector)

		if err != nil {
			return nil, err
		}

		namespaces = live
	}

	if len(namespaces) == 0 {
		return nil, usererror.Errorf("environment %q not found", id)
	}

	return new(toEnvironment(namespaces[0])), nil
}

func (c *Client) EnvironmentsByWorkspace(ctx context.Context, workspaceID string) ([]platform.Environment, error) {
	selector := labels.SelectorFromSet(labels.Set{
		workspaceLabel: workspaceID,
	})

	namespaces, err := c.namespacesMatching(ctx, selector, workspaceIndex, workspaceID)

	if err != nil {
		return nil, err
	}

	return toEnvironments(namespaces), nil
}

func toEnvironments(namespaces []core.Namespace) []platform.Environment {
	result := make([]platform.Environment, 0, len(namespaces))

	for _, namespace := range namespaces {
		result = append(result, toEnvironment(namespace))
	}

	return result
}

func toEnvironment(namespace core.Namespace) platform.Environment {
	env := platform.Environment{
		ID:        environmentID(namespace.Labels),
		Name:      namespace.Labels[environmentLabel],
		CreatedAt: namespace.CreationTimestamp.Time,
	}

	switch namespace.Labels[resourceTierLabel] {
	case string(tierFromLabel(resourceTierProd)):
		env.ResourceTier = platform.ProductionTier
	case string(tierFromLabel(resourceTierEco)):
		env.ResourceTier = platform.EcoTier
	default:
		slog.Warn("invalid value for label",
			"label", resourceTierLabel,
			"value", namespace.Labels[resourceTierLabel],
			"namespace", namespace.Name,
		)

		env.ResourceTier = platform.EcoTier
	}

	return env
}

func environmentID(labels map[string]string) platform.EnvironmentID {
	return platform.EnvironmentID{
		Workspace: labels[workspaceLabel],
		Project:   labels[projectLabel],
		Name:      labels[environmentLabel],
	}
}
