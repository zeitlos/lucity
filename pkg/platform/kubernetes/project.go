package kubernetes

import (
	"context"
	"sort"

	pkglabels "github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/pkg/usererror"

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
		projects = append(projects, toProject(namespaces))
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].ID.Name < projects[j].ID.Name
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

	return new(toProject(namespaces)), nil
}

func toProject(namespaces []core.Namespace) platform.Project {
	project := platform.Project{
		ID:           projectID(namespaces[0]),
		Name:         namespaces[0].Labels[projectLabel],
		Environments: toEnvironments(namespaces),
	}

	for _, namespace := range namespaces {
		if name := namespace.Annotations[pkglabels.ProjectName]; name != "" {
			project.Name = name
			break
		}
	}

	return project
}

func projectID(namespace core.Namespace) platform.ProjectID {
	return platform.ProjectID{
		Workspace: namespace.Labels[workspaceLabel],
		Name:      namespace.Labels[projectLabel],
	}
}
