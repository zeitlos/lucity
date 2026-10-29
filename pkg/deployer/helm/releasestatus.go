package helm

import (
	"context"
	"strconv"
	"strings"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/metadata/metadatainformer"
	"k8s.io/client-go/tools/cache"
)

// Statuses counts the lucity-app releases by the status of their newest
// revision. It reports false until the cache has synced.
func (r *ReleaseCache) Statuses() (map[string]int, bool) {
	if !r.informer.HasSynced() {
		return nil, false
	}

	return newestStatuses(r.informer.GetStore().List()), true
}

// WorkspaceReleaseStatuses watches the metadata of the lucity-workspace
// release secrets and counts the releases by the status of their newest
// revision.
type WorkspaceReleaseStatuses struct {
	informer cache.SharedIndexInformer
}

func NewWorkspaceReleaseStatuses(client metadata.Interface, namespace string) *WorkspaceReleaseStatuses {
	informer := metadatainformer.NewFilteredMetadataInformer(
		client,
		core.SchemeGroupVersion.WithResource("secrets"),
		namespace,
		0,
		cache.Indexers{},
		func(options *meta.ListOptions) {
			options.LabelSelector = "owner=helm"
		},
	)

	return &WorkspaceReleaseStatuses{informer: informer.Informer()}
}

func (w *WorkspaceReleaseStatuses) Start(ctx context.Context) {
	go w.informer.Run(ctx.Done())
}

// Statuses reports false until the informer has synced.
func (w *WorkspaceReleaseStatuses) Statuses() (map[string]int, bool) {
	if !w.informer.HasSynced() {
		return nil, false
	}

	var objects []any

	for _, object := range w.informer.GetStore().List() {
		if secret, ok := object.(*meta.PartialObjectMetadata); ok && strings.HasPrefix(secret.Labels["name"], workspaceReleasePrefix) {
			objects = append(objects, object)
		}
	}

	return newestStatuses(objects), true
}

func newestStatuses(objects []any) map[string]int {
	type revision struct {
		version int
		status  string
	}

	newest := map[string]revision{}

	for _, object := range objects {
		secret, ok := object.(*meta.PartialObjectMetadata)

		if !ok {
			continue
		}

		version, err := strconv.Atoi(secret.Labels["version"])

		if err != nil {
			continue
		}

		key := secret.Namespace + "/" + secret.Labels["name"]

		if current, ok := newest[key]; ok && current.version >= version {
			continue
		}

		newest[key] = revision{version: version, status: secret.Labels["status"]}
	}

	statuses := map[string]int{}

	for _, release := range newest {
		statuses[release.status]++
	}

	return statuses
}
