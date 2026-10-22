package kubernetes

import (
	"context"
	"sort"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	workspaceIndex = "workspace"
	projectIndex   = "project"
)

func newNamespaceInformer(client kubernetes.Interface) cache.SharedIndexInformer {
	indexers := cache.Indexers{
		workspaceIndex: indexByWorkspace,
		projectIndex:   indexByProject,
	}

	return coreinformers.NewFilteredNamespaceInformer(client, 0, indexers, func(options *meta.ListOptions) {
		options.LabelSelector = workspaceLabel
	})
}

// Start runs the namespace cache until ctx is done. Reads use live LISTs until
// the cache has synced, so callers never have to wait for it.
func (c *Client) Start(ctx context.Context) {
	go c.namespaces.Run(ctx.Done())
}

func (c *Client) namespacesMatching(ctx context.Context, selector labels.Selector, index, key string) ([]core.Namespace, error) {
	if cached, ok := c.cachedNamespaces(selector, index, key); ok {
		return cached, nil
	}

	return c.liveNamespaces(ctx, selector)
}

func (c *Client) cachedNamespaces(selector labels.Selector, index, key string) ([]core.Namespace, bool) {
	if !c.namespaces.HasSynced() {
		return nil, false
	}

	var objects []any

	if index == "" {
		objects = c.namespaces.GetIndexer().List()
	} else {
		indexed, err := c.namespaces.GetIndexer().ByIndex(index, key)

		if err != nil {
			return nil, false
		}

		objects = indexed
	}

	result := make([]core.Namespace, 0, len(objects))

	for _, object := range objects {
		namespace, ok := object.(*core.Namespace)

		if !ok || !selector.Matches(labels.Set(namespace.Labels)) {
			continue
		}

		result = append(result, *namespace)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, true
}

func (c *Client) liveNamespaces(ctx context.Context, selector labels.Selector) ([]core.Namespace, error) {
	list, err := c.kubernetes.CoreV1().Namespaces().List(ctx, meta.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func indexByWorkspace(object any) ([]string, error) {
	namespace, ok := object.(*core.Namespace)

	if !ok {
		return nil, nil
	}

	return []string{namespace.Labels[workspaceLabel]}, nil
}

func indexByProject(object any) ([]string, error) {
	namespace, ok := object.(*core.Namespace)

	if !ok || namespace.Labels[projectLabel] == "" {
		return nil, nil
	}

	return []string{projectKey(namespace.Labels[workspaceLabel], namespace.Labels[projectLabel])}, nil
}

func projectKey(workspace, project string) string {
	return workspace + "/" + project
}
