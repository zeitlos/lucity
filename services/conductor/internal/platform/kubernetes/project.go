package kubernetes

import (
	"context"
	"sort"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"
	"github.com/zeitlos/lucity/services/conductor/internal/usererror"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
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

	namespaces, err := c.namespacesMatching(ctx, selector, workspaceIndex, workspaceID)

	if err != nil {
		return nil, err
	}

	byProject := make(map[string][]core.Namespace)

	for _, namespace := range namespaces {
		key := namespace.Labels[projectLabel]
		byProject[key] = append(byProject[key], namespace)
	}

	projects := make([]platform.Project, 0, len(byProject))

	for _, namespaces := range byProject {
		projects = append(projects, toProject(namespaces[0], toEnvironments(namespaces)))
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	return projects, nil
}

func (c *Client) Project(ctx context.Context, id platform.ProjectID) (*platform.Project, error) {
	selector := labels.SelectorFromSet(labels.Set{
		workspaceLabel: id.Workspace,
		projectLabel:   id.Name,
	})

	namespaces, ok := c.cachedNamespaces(selector, projectIndex, projectKey(id.Workspace, id.Name))

	if !ok || len(namespaces) == 0 {
		live, err := c.liveNamespaces(ctx, selector)

		if err != nil {
			return nil, err
		}

		namespaces = live
	}

	if len(namespaces) < 1 {
		return nil, usererror.Errorf("project with id %q not found", id)
	}

	return new(toProject(namespaces[0], toEnvironments(namespaces))), nil
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
