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

const sharedVariablesLabel = "lucity.dev/shared-variables"

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

	result := make([]platform.Environment, 0, len(list.Items))

	for _, namespace := range list.Items {
		env := toEnvironment(namespace)
		env.Variables, _ = c.sharedVariables(ctx, namespace.Name)
		result = append(result, env)
	}

	return result, nil
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

	env := toEnvironment(list.Items[0])
	env.Variables, _ = c.sharedVariables(ctx, list.Items[0].Name)

	return &env, nil
}

func (c *Client) EnvironmentsByWorkspace(ctx context.Context, workspaceID string) ([]platform.Environment, error) {
	set := labels.Set{
		workspaceLabel: workspaceID,
	}

	list, err := c.kubernetes.CoreV1().Namespaces().List(ctx, meta.ListOptions{
		LabelSelector: labels.SelectorFromSet(set).String(),
	})

	if err != nil {
		return nil, err
	}

	result := make([]platform.Environment, 0, len(list.Items))

	for _, namespace := range list.Items {
		env := toEnvironment(namespace)
		env.Variables, _ = c.sharedVariables(ctx, namespace.Name)
		result = append(result, env)
	}

	return result, nil
}

// sharedVariables returns the contents of the namespace's shared-variables
// Secret. Returns nil if no such Secret exists (release not yet installed,
// or no shared vars configured).
func (c *Client) sharedVariables(ctx context.Context, namespace string) (map[string]string, error) {
	list, err := c.kubernetes.CoreV1().Secrets(namespace).List(ctx, meta.ListOptions{
		LabelSelector: sharedVariablesLabel + "=true",
	})

	if err != nil {
		return nil, err
	}

	if len(list.Items) == 0 {
		return nil, nil
	}

	variables := make(map[string]string, len(list.Items[0].Data))

	for name, value := range list.Items[0].Data {
		variables[name] = string(value)
	}

	return variables, nil
}

// toEnvironments builds Environment values from namespaces without
// populating Variables (avoids one Secret fetch per env when only the
// summary view is needed, e.g. project listings).
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

func environmentID(labels map[string]string) platform.EnvironmentID {
	return platform.EnvironmentID{
		Workspace: labels[workspaceLabel],
		Project:   labels[projectLabel],
		Name:      labels[environmentLabel],
	}
}
