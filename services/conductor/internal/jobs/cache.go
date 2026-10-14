package jobs

import (
	"context"
	"sort"
	"sync"
	"time"

	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	batchlisters "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/zeitlos/lucity/pkg/labels"
)

const labelComponent = labels.Prefix + "component"

const awaitTimeout = 2 * time.Second

// Cache holds the build, scan and deploy Jobs. Their pod templates are dropped,
// so the cached Jobs carry metadata, suspension and status only.
type Cache struct {
	client          kubernetes.Interface
	buildNamespace  string
	deployNamespace string
	factories       []informers.SharedInformerFactory
	builds          cache.SharedIndexInformer
	deploys         cache.SharedIndexInformer

	buildLister  batchlisters.JobNamespaceLister
	deployLister batchlisters.JobNamespaceLister

	mu       sync.Mutex
	messages map[types.UID]string
}

func NewCache(client kubernetes.Interface, buildNamespace, deployNamespace string) (*Cache, error) {
	buildFactory := jobFactory(client, buildNamespace, labelComponent+" in (build, scan)")
	deployFactory := jobFactory(client, deployNamespace, labelComponent+"=deploy")

	builds := buildFactory.Batch().V1().Jobs()
	deploys := deployFactory.Batch().V1().Jobs()

	c := &Cache{
		client:          client,
		buildNamespace:  buildNamespace,
		deployNamespace: deployNamespace,
		factories:       []informers.SharedInformerFactory{buildFactory, deployFactory},
		builds:          builds.Informer(),
		deploys:         deploys.Informer(),
		buildLister:     builds.Lister().Jobs(buildNamespace),
		deployLister:    deploys.Lister().Jobs(deployNamespace),
		messages:        map[types.UID]string{},
	}

	_, err := c.builds.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: c.forgetMessage,
	})

	if err != nil {
		return nil, err
	}

	return c, nil
}

func jobFactory(client kubernetes.Interface, namespace, selector string) informers.SharedInformerFactory {
	return informers.NewSharedInformerFactoryWithOptions(client, 0,
		informers.WithNamespace(namespace),
		informers.WithTweakListOptions(func(options *meta.ListOptions) {
			options.LabelSelector = selector
		}),
		informers.WithTransform(stripJob),
	)
}

func (c *Cache) Start(ctx context.Context) {
	for _, factory := range c.factories {
		factory.Start(ctx.Done())
	}
}

func (c *Cache) HasSynced() bool {
	return c.builds.HasSynced() && c.deploys.HasSynced()
}

// Builds returns the build and scan Jobs matching selector. Until the cache
// has synced it lists them live.
func (c *Cache) Builds(ctx context.Context, selector k8slabels.Selector) ([]*batch.Job, error) {
	if c.builds.HasSynced() {
		return c.buildLister.List(selector)
	}

	return c.live(ctx, c.buildNamespace, selector)
}

// Deploys returns the deploy Jobs matching selector. Until the cache has
// synced it lists them live.
func (c *Cache) Deploys(ctx context.Context, selector k8slabels.Selector) ([]*batch.Job, error) {
	if c.deploys.HasSynced() {
		return c.deployLister.List(selector)
	}

	return c.live(ctx, c.deployNamespace, selector)
}

// Await returns once the cache holds a Job that was just created, so reads
// right after the create see it. It gives up after a short timeout.
func (c *Cache) Await(ctx context.Context, job *batch.Job) {
	if c == nil || !c.HasSynced() {
		return
	}

	lister := c.buildLister

	if job.Labels[labelComponent] == "deploy" {
		lister = c.deployLister
	}

	ctx, cancel := context.WithTimeout(ctx, awaitTimeout)
	defer cancel()

	_ = wait.PollUntilContextCancel(ctx, 10*time.Millisecond, true, func(context.Context) (bool, error) {
		cached, err := lister.Get(job.Name)
		return err == nil && cached.UID == job.UID, nil
	})
}

func (c *Cache) live(ctx context.Context, namespace string, selector k8slabels.Selector) ([]*batch.Job, error) {
	list, err := c.client.BatchV1().Jobs(namespace).List(ctx, meta.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return nil, err
	}

	result := make([]*batch.Job, 0, len(list.Items))

	for i := range list.Items {
		result = append(result, &list.Items[i])
	}

	return result, nil
}

// TerminationMessage returns the termination message of a finished Job. It is
// read from the Job's pods once and kept until the Job is deleted.
func (c *Cache) TerminationMessage(ctx context.Context, job *batch.Job) (string, error) {
	c.mu.Lock()
	message, ok := c.messages[job.UID]
	c.mu.Unlock()

	if ok {
		return message, nil
	}

	message, err := TerminationMessage(ctx, c.client, job.Namespace, job.Name)

	if err != nil || message == "" {
		return message, err
	}

	c.mu.Lock()
	c.messages[job.UID] = message
	c.mu.Unlock()

	return message, nil
}

func (c *Cache) forgetMessage(object any) {
	if tombstone, ok := object.(cache.DeletedFinalStateUnknown); ok {
		object = tombstone.Obj
	}

	job, ok := object.(*batch.Job)

	if !ok {
		return
	}

	c.mu.Lock()
	delete(c.messages, job.UID)
	c.mu.Unlock()
}

// SortNewestFirst orders Jobs by creation time, newest first, then by name.
func SortNewestFirst(list []*batch.Job) {
	sort.Slice(list, func(i, j int) bool {
		if !list[i].CreationTimestamp.Equal(&list[j].CreationTimestamp) {
			return list[j].CreationTimestamp.Before(&list[i].CreationTimestamp)
		}

		return list[i].Name < list[j].Name
	})
}

func stripJob(object any) (any, error) {
	if job, ok := object.(*batch.Job); ok {
		job.ManagedFields = nil
		job.Spec.Template = core.PodTemplateSpec{}
	}

	return object, nil
}
