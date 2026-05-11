package kubernetes

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *Client) Environments(ctx context.Context, projectID platform.ProjectID) ([]platform.Environment, error) {
	set := labels.Set{
		workspaceLabel: projectID.Workspace,
		projectLabel:   projectID.Name,
	}

	list, err := c.kubernetes.CoreV1().Namespaces().List(ctx, meta.ListOptions{
		LabelSelector: labels.SelectorFromSet(set).String(),
	})

	if err != nil {
		return nil, err
	}

	return toEnvironments(list.Items), nil
}

func (c *Client) Environment(ctx context.Context, id platform.EnvironmentID) (*platform.Environment, error) {
	set := labels.Set{
		workspaceLabel:   id.Workspace,
		projectLabel:     id.Project,
		environmentLabel: id.Name,
	}

	list, err := c.kubernetes.CoreV1().Namespaces().List(ctx, meta.ListOptions{
		LabelSelector: labels.SelectorFromSet(set).String(),
	})

	if err != nil {
		return nil, err
	}

	if len(list.Items) == 0 {
		return nil, fmt.Errorf("environment %q not found", id)
	}

	return new(toEnvironment(list.Items[0])), nil
}

func toEnvironment(namespace core.Namespace) platform.Environment {
	env := platform.Environment{
		ID:        environmentID(namespace),
		Name:      namespace.Labels[environmentLabel],
		CreatedAt: namespace.CreationTimestamp.Time,
	}

	switch namespace.Labels[resourceTierLabel] {
	case resourceTierProd:
		env.ResourceTier = platform.ProductionTier
	case resourceTierEco:
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

func toEnvironments(namespaces []core.Namespace) []platform.Environment {
	result := make([]platform.Environment, 0, len(namespaces))

	for _, namespace := range namespaces {
		result = append(result, toEnvironment(namespace))
	}

	return result
}

func environmentID(namespace core.Namespace) platform.EnvironmentID {
	return platform.EnvironmentID{
		Workspace: namespace.Labels[workspaceLabel],
		Project:   namespace.Labels[projectLabel],
		Name:      namespace.Labels[environmentLabel],
	}
}
