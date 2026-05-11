package kubernetes

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

const (
	workspaceLabel    = "lucity.dev/workspace"
	projectLabel      = "lucity.dev/project"
	environmentLabel  = "lucity.dev/environment"
	resourceTierLabel = "lucity.dev/resource-tier"
)

const (
	resourceTierEco  = "eco"
	resourceTierProd = "production"
)

func (c *Client) Projects(ctx context.Context, workspaceID string) ([]platform.Project, error) {
	req, err := labels.NewRequirement(projectLabel, selection.Exists, nil)

	if err != nil {
		return nil, err
	}

	set := labels.Set{
		workspaceLabel: workspaceID,
	}

	selector := labels.SelectorFromSet(set).Add(*req)

	list, err := c.kubernetes.CoreV1().Namespaces().List(ctx, meta.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return nil, err
	}

	byProject := make(map[string][]core.Namespace)

	for _, namespace := range list.Items {
		key := namespace.Labels[projectLabel]
		byProject[key] = append(byProject[key], namespace)
	}

	projects := make([]platform.Project, 0, len(byProject))

	for _, namespaces := range byProject {
		projects = append(projects, toProject(namespaces[0], toEnvironments(namespaces)))
	}

	return projects, nil
}

func (c *Client) Project(ctx context.Context, id platform.ProjectID) (*platform.Project, error) {
	set := labels.Set{
		workspaceLabel: id.Workspace,
		projectLabel:   id.Name,
	}

	list, err := c.kubernetes.CoreV1().Namespaces().List(ctx, meta.ListOptions{
		LabelSelector: labels.SelectorFromSet(set).String(),
	})

	if err != nil {
		return nil, err
	}

	if len(list.Items) < 1 {
		return nil, fmt.Errorf("project with id %q not found", id)
	}

	return new(toProject(list.Items[0], toEnvironments(list.Items))), nil
}

func toProject(namespace core.Namespace, environments []platform.Environment) platform.Project {
	return platform.Project{
		ID:           projectID(namespace),
		Name:         namespace.Labels[projectLabel],
		Environments: environments,
	}
}

func projectID(namespace core.Namespace) platform.ProjectID {
	return platform.ProjectID{
		Workspace: namespace.Labels[workspaceLabel],
		Name:      namespace.Labels[projectLabel],
	}
}
