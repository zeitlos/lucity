package helm

import (
	"context"
	"strconv"
	"sync"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/metadata/metadatainformer"
	"k8s.io/client-go/tools/cache"

	"github.com/zeitlos/lucity/pkg/deployer/values"
)

// ReleaseCache watches the metadata of the Helm release secrets, never their
// payloads, so the reconciler can tell whether a release changed since it last
// confirmed the release runs the embedded chart version.
type ReleaseCache struct {
	informer cache.SharedIndexInformer

	mu         sync.Mutex
	verified   map[string]types.UID
	placements map[string]placedRelease
}

type placedRelease struct {
	uid       types.UID
	placement values.Placement
}

func NewReleaseCache(client metadata.Interface) *ReleaseCache {
	informer := metadatainformer.NewFilteredMetadataInformer(
		client,
		core.SchemeGroupVersion.WithResource("secrets"),
		meta.NamespaceAll,
		0,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
		func(options *meta.ListOptions) {
			options.LabelSelector = "owner=helm,name=" + releaseName
		},
	)

	return &ReleaseCache{
		informer:   informer.Informer(),
		verified:   map[string]types.UID{},
		placements: map[string]placedRelease{},
	}
}

func (r *ReleaseCache) Start(ctx context.Context) {
	go r.informer.Run(ctx.Done())
}

func (r *ReleaseCache) newest(namespace string) (uid types.UID, revision int, synced bool) {
	if !r.informer.HasSynced() {
		return "", 0, false
	}

	objects, err := r.informer.GetIndexer().ByIndex(cache.NamespaceIndex, namespace)

	if err != nil {
		return "", 0, false
	}

	for _, object := range objects {
		secret, ok := object.(*meta.PartialObjectMetadata)

		if !ok {
			continue
		}

		version, err := strconv.Atoi(secret.Labels["version"])

		if err != nil || version <= revision {
			continue
		}

		uid = secret.UID
		revision = version
	}

	return uid, revision, true
}

func (r *ReleaseCache) isVerified(namespace string, uid types.UID) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.verified[namespace] == uid
}

func (r *ReleaseCache) markVerified(namespace string, uid types.UID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.verified[namespace] = uid
}

func (r *ReleaseCache) placement(namespace string, uid types.UID) (values.Placement, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	placed, ok := r.placements[namespace]

	if !ok || placed.uid != uid {
		return values.Placement{}, false
	}

	return placed.placement, true
}

func (r *ReleaseCache) markPlacement(namespace string, uid types.UID, placement values.Placement) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.placements[namespace] = placedRelease{uid: uid, placement: placement}
}
