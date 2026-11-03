package kubernetes

import (
	"context"
	"sync"

	"github.com/zeitlos/lucity/pkg/jobs"
	lucitylabels "github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/pkg/platform"

	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/metadata/metadatainformer"
	"k8s.io/client-go/tools/cache"
)

var certificateGVR = schema.GroupVersionResource{
	Group:    "cert-manager.io",
	Version:  "v1",
	Resource: "certificates",
}

// Changes tells subscribers when the objects behind an environment change. It
// listens to the cached clients and the jobs cache, and watches the metadata
// of certificates and bucket secrets, which are read outside the platform
// client.
type Changes struct {
	informers []cache.SharedIndexInformer

	mu          sync.Mutex
	subscribers map[string]map[chan struct{}]struct{}
}

var _ platform.Changes = (*Changes)(nil)

func NewChanges(cached *CachedClients, jobsCache *jobs.Cache, client metadata.Interface) (*Changes, error) {
	c := &Changes{
		informers: []cache.SharedIndexInformer{
			metadatainformer.NewFilteredMetadataInformer(client, certificateGVR, meta.NamespaceAll, 0, cache.Indexers{}, nil).Informer(),
			metadatainformer.NewFilteredMetadataInformer(client, core.SchemeGroupVersion.WithResource("secrets"), meta.NamespaceAll, 0, cache.Indexers{}, withLabel(lucitylabels.ObjectStorageBucket)).Informer(),
		},
		subscribers: map[string]map[chan struct{}]struct{}{},
	}

	handler := namespaceHandler(c.notify)

	for _, informer := range c.informers {
		if err := informer.SetTransform(stripManagedFields); err != nil {
			return nil, err
		}

		if _, err := informer.AddEventHandler(handler); err != nil {
			return nil, err
		}
	}

	if err := cached.OnChange(c.notify); err != nil {
		return nil, err
	}

	err := jobsCache.OnChange(func(job *batch.Job) {
		c.notify(environmentID(job.Labels).Namespace())
	})

	if err != nil {
		return nil, err
	}

	return c, nil
}

// Start runs the metadata watches until ctx is done.
func (c *Changes) Start(ctx context.Context) {
	for _, informer := range c.informers {
		go informer.Run(ctx.Done())
	}
}

func (c *Changes) Watch(ctx context.Context, environment platform.EnvironmentID) <-chan struct{} {
	namespace := environment.Namespace()
	signal := make(chan struct{}, 1)

	c.mu.Lock()

	if c.subscribers[namespace] == nil {
		c.subscribers[namespace] = map[chan struct{}]struct{}{}
	}

	c.subscribers[namespace][signal] = struct{}{}
	c.mu.Unlock()

	go func() {
		<-ctx.Done()

		c.mu.Lock()
		defer c.mu.Unlock()

		delete(c.subscribers[namespace], signal)

		if len(c.subscribers[namespace]) == 0 {
			delete(c.subscribers, namespace)
		}
	}()

	return signal
}

func (c *Changes) notify(namespace string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for signal := range c.subscribers[namespace] {
		select {
		case signal <- struct{}{}:
		default:
		}
	}
}
