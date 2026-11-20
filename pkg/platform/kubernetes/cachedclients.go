package kubernetes

import (
	"context"
	"sort"

	apps "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v2"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	autoscalinginformers "k8s.io/client-go/informers/autoscaling/v2"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	appsclient "k8s.io/client-go/kubernetes/typed/apps/v1"
	autoscalingclient "k8s.io/client-go/kubernetes/typed/autoscaling/v2"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

// CachedClients wraps a clientset and a dynamic client so that List calls for
// the resources behind an environment are answered from informers. Every other
// call goes to the apiserver.
type CachedClients struct {
	kubernetes *cachedClientset
	dynamic    *cachedDynamic
	caches     []*resourceCache
}

func NewCachedClients(clientset kubernetes.Interface, dynamicClient dynamic.Interface) (*CachedClients, error) {
	indexers := cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}

	deployments := &resourceCache{label: serviceLabel, informer: appsinformers.NewFilteredDeploymentInformer(clientset, meta.NamespaceAll, 0, indexers, withLabel(serviceLabel))}
	replicaSets := &resourceCache{label: serviceLabel, informer: appsinformers.NewFilteredReplicaSetInformer(clientset, meta.NamespaceAll, 0, indexers, withLabel(serviceLabel))}
	statefulSets := &resourceCache{label: keyValueStoreLabel, informer: appsinformers.NewFilteredStatefulSetInformer(clientset, meta.NamespaceAll, 0, indexers, withLabel(keyValueStoreLabel))}
	pods := &resourceCache{label: serviceLabel, informer: coreinformers.NewFilteredPodInformer(clientset, meta.NamespaceAll, 0, indexers, withLabel(serviceLabel))}
	volumeClaims := &resourceCache{label: volumeLabel, informer: coreinformers.NewFilteredPersistentVolumeClaimInformer(clientset, meta.NamespaceAll, 0, indexers, withLabel(volumeLabel))}
	autoscalers := &resourceCache{label: serviceLabel, informer: autoscalinginformers.NewFilteredHorizontalPodAutoscalerInformer(clientset, meta.NamespaceAll, 0, indexers, withLabel(serviceLabel))}
	routes := &resourceCache{label: serviceLabel, informer: dynamicinformer.NewFilteredDynamicInformer(dynamicClient, httpRouteGVR, meta.NamespaceAll, 0, indexers, withLabel(serviceLabel)).Informer()}
	clusters := &resourceCache{label: databaseLabel, informer: dynamicinformer.NewFilteredDynamicInformer(dynamicClient, cnpgClusterGVR, meta.NamespaceAll, 0, indexers, withLabel(databaseLabel)).Informer()}

	caches := []*resourceCache{deployments, replicaSets, statefulSets, pods, volumeClaims, autoscalers, routes, clusters}

	for _, cached := range caches {
		if err := cached.informer.SetTransform(stripManagedFields); err != nil {
			return nil, err
		}
	}

	return &CachedClients{
		kubernetes: &cachedClientset{
			Interface: clientset,
			apps: &cachedApps{
				AppsV1Interface: clientset.AppsV1(),
				deployments:     deployments,
				replicaSets:     replicaSets,
				statefulSets:    statefulSets,
			},
			core: &cachedCore{
				CoreV1Interface: clientset.CoreV1(),
				pods:            pods,
				volumeClaims:    volumeClaims,
			},
			autoscaling: &cachedAutoscaling{
				AutoscalingV2Interface: clientset.AutoscalingV2(),
				autoscalers:            autoscalers,
			},
		},
		dynamic: &cachedDynamic{
			Interface: dynamicClient,
			caches: map[schema.GroupVersionResource]*resourceCache{
				httpRouteGVR:   routes,
				cnpgClusterGVR: clusters,
			},
		},
		caches: caches,
	}, nil
}

// Start runs the informers until ctx is done. Lists go to the apiserver until
// their informer has synced.
func (c *CachedClients) Start(ctx context.Context) {
	for _, cached := range c.caches {
		go cached.informer.Run(ctx.Done())
	}
}

func (c *CachedClients) Kubernetes() kubernetes.Interface {
	return c.kubernetes
}

func (c *CachedClients) Dynamic() dynamic.Interface {
	return c.dynamic
}

// OnChange calls notify with the namespace of every object the informers add,
// change or delete.
func (c *CachedClients) OnChange(notify func(namespace string)) error {
	handler := namespaceHandler(notify)

	for _, cached := range c.caches {
		if _, err := cached.informer.AddEventHandler(handler); err != nil {
			return err
		}
	}

	return nil
}

// resourceCache answers a List from its informer when the informer holds every
// object the List asks for: it has synced, the label selector requires the
// label the informer selects on, and there is no field selector.
type resourceCache struct {
	label    string
	informer cache.SharedIndexInformer
}

func (r *resourceCache) list(namespace string, options meta.ListOptions) ([]any, bool) {
	if !r.informer.HasSynced() || options.FieldSelector != "" {
		return nil, false
	}

	selector, err := labels.Parse(options.LabelSelector)

	if err != nil || !requires(selector, r.label) {
		return nil, false
	}

	var objects []any

	if namespace == meta.NamespaceAll {
		objects = r.informer.GetIndexer().List()
	} else {
		indexed, err := r.informer.GetIndexer().ByIndex(cache.NamespaceIndex, namespace)

		if err != nil {
			return nil, false
		}

		objects = indexed
	}

	matching := make([]any, 0, len(objects))

	for _, object := range objects {
		accessor, ok := object.(meta.Object)

		if ok && selector.Matches(labels.Set(accessor.GetLabels())) {
			matching = append(matching, object)
		}
	}

	sort.Slice(matching, func(i, j int) bool {
		left := matching[i].(meta.Object)
		right := matching[j].(meta.Object)

		if left.GetNamespace() != right.GetNamespace() {
			return left.GetNamespace() < right.GetNamespace()
		}

		return left.GetName() < right.GetName()
	})

	return matching, true
}

func requires(selector labels.Selector, key string) bool {
	requirements, _ := selector.Requirements()

	for _, requirement := range requirements {
		if requirement.Key() != key {
			continue
		}

		switch requirement.Operator() {
		case selection.Exists, selection.Equals, selection.DoubleEquals, selection.In:
			return true
		}
	}

	return false
}

func items[T any, P interface {
	*T
	runtime.Object
}](objects []any) []T {
	result := make([]T, 0, len(objects))

	for _, object := range objects {
		result = append(result, *object.(P).DeepCopyObject().(P))
	}

	return result
}

func withLabel(label string) func(*meta.ListOptions) {
	return func(options *meta.ListOptions) {
		options.LabelSelector = label
	}
}

func stripManagedFields(object any) (any, error) {
	if accessor, ok := object.(meta.Object); ok {
		accessor.SetManagedFields(nil)
	}

	return object, nil
}

func namespaceHandler(notify func(namespace string)) cache.ResourceEventHandler {
	handle := func(object any) {
		if tombstone, ok := object.(cache.DeletedFinalStateUnknown); ok {
			object = tombstone.Obj
		}

		if accessor, ok := object.(meta.Object); ok {
			notify(accessor.GetNamespace())
		}
	}

	return cache.ResourceEventHandlerFuncs{
		AddFunc: handle,
		UpdateFunc: func(previous, object any) {
			before, beforeOK := previous.(meta.Object)
			after, afterOK := object.(meta.Object)

			if beforeOK && afterOK && before.GetResourceVersion() == after.GetResourceVersion() {
				return
			}

			handle(object)
		},
		DeleteFunc: handle,
	}
}

type cachedClientset struct {
	kubernetes.Interface
	apps        *cachedApps
	core        *cachedCore
	autoscaling *cachedAutoscaling
}

func (c *cachedClientset) AppsV1() appsclient.AppsV1Interface {
	return c.apps
}

func (c *cachedClientset) CoreV1() coreclient.CoreV1Interface {
	return c.core
}

func (c *cachedClientset) AutoscalingV2() autoscalingclient.AutoscalingV2Interface {
	return c.autoscaling
}

type cachedApps struct {
	appsclient.AppsV1Interface
	deployments  *resourceCache
	replicaSets  *resourceCache
	statefulSets *resourceCache
}

func (a *cachedApps) Deployments(namespace string) appsclient.DeploymentInterface {
	return &cachedDeployments{DeploymentInterface: a.AppsV1Interface.Deployments(namespace), cache: a.deployments, namespace: namespace}
}

func (a *cachedApps) ReplicaSets(namespace string) appsclient.ReplicaSetInterface {
	return &cachedReplicaSets{ReplicaSetInterface: a.AppsV1Interface.ReplicaSets(namespace), cache: a.replicaSets, namespace: namespace}
}

func (a *cachedApps) StatefulSets(namespace string) appsclient.StatefulSetInterface {
	return &cachedStatefulSets{StatefulSetInterface: a.AppsV1Interface.StatefulSets(namespace), cache: a.statefulSets, namespace: namespace}
}

type cachedDeployments struct {
	appsclient.DeploymentInterface
	cache     *resourceCache
	namespace string
}

func (d *cachedDeployments) List(ctx context.Context, options meta.ListOptions) (*apps.DeploymentList, error) {
	if objects, ok := d.cache.list(d.namespace, options); ok {
		return &apps.DeploymentList{Items: items[apps.Deployment](objects)}, nil
	}

	return d.DeploymentInterface.List(ctx, options)
}

type cachedReplicaSets struct {
	appsclient.ReplicaSetInterface
	cache     *resourceCache
	namespace string
}

func (r *cachedReplicaSets) List(ctx context.Context, options meta.ListOptions) (*apps.ReplicaSetList, error) {
	if objects, ok := r.cache.list(r.namespace, options); ok {
		return &apps.ReplicaSetList{Items: items[apps.ReplicaSet](objects)}, nil
	}

	return r.ReplicaSetInterface.List(ctx, options)
}

type cachedStatefulSets struct {
	appsclient.StatefulSetInterface
	cache     *resourceCache
	namespace string
}

func (s *cachedStatefulSets) List(ctx context.Context, options meta.ListOptions) (*apps.StatefulSetList, error) {
	if objects, ok := s.cache.list(s.namespace, options); ok {
		return &apps.StatefulSetList{Items: items[apps.StatefulSet](objects)}, nil
	}

	return s.StatefulSetInterface.List(ctx, options)
}

type cachedCore struct {
	coreclient.CoreV1Interface
	pods         *resourceCache
	volumeClaims *resourceCache
}

func (c *cachedCore) Pods(namespace string) coreclient.PodInterface {
	return &cachedPods{PodInterface: c.CoreV1Interface.Pods(namespace), cache: c.pods, namespace: namespace}
}

func (c *cachedCore) PersistentVolumeClaims(namespace string) coreclient.PersistentVolumeClaimInterface {
	return &cachedVolumeClaims{PersistentVolumeClaimInterface: c.CoreV1Interface.PersistentVolumeClaims(namespace), cache: c.volumeClaims, namespace: namespace}
}

type cachedPods struct {
	coreclient.PodInterface
	cache     *resourceCache
	namespace string
}

func (p *cachedPods) List(ctx context.Context, options meta.ListOptions) (*core.PodList, error) {
	if objects, ok := p.cache.list(p.namespace, options); ok {
		return &core.PodList{Items: items[core.Pod](objects)}, nil
	}

	return p.PodInterface.List(ctx, options)
}

type cachedVolumeClaims struct {
	coreclient.PersistentVolumeClaimInterface
	cache     *resourceCache
	namespace string
}

func (v *cachedVolumeClaims) List(ctx context.Context, options meta.ListOptions) (*core.PersistentVolumeClaimList, error) {
	if objects, ok := v.cache.list(v.namespace, options); ok {
		return &core.PersistentVolumeClaimList{Items: items[core.PersistentVolumeClaim](objects)}, nil
	}

	return v.PersistentVolumeClaimInterface.List(ctx, options)
}

type cachedAutoscaling struct {
	autoscalingclient.AutoscalingV2Interface
	autoscalers *resourceCache
}

func (a *cachedAutoscaling) HorizontalPodAutoscalers(namespace string) autoscalingclient.HorizontalPodAutoscalerInterface {
	return &cachedAutoscalers{HorizontalPodAutoscalerInterface: a.AutoscalingV2Interface.HorizontalPodAutoscalers(namespace), cache: a.autoscalers, namespace: namespace}
}

type cachedAutoscalers struct {
	autoscalingclient.HorizontalPodAutoscalerInterface
	cache     *resourceCache
	namespace string
}

func (a *cachedAutoscalers) List(ctx context.Context, options meta.ListOptions) (*autoscaling.HorizontalPodAutoscalerList, error) {
	if objects, ok := a.cache.list(a.namespace, options); ok {
		return &autoscaling.HorizontalPodAutoscalerList{Items: items[autoscaling.HorizontalPodAutoscaler](objects)}, nil
	}

	return a.HorizontalPodAutoscalerInterface.List(ctx, options)
}

type cachedDynamic struct {
	dynamic.Interface
	caches map[schema.GroupVersionResource]*resourceCache
}

func (d *cachedDynamic) Resource(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	client := d.Interface.Resource(resource)
	cached, ok := d.caches[resource]

	if !ok {
		return client
	}

	return &cachedNamespaceableResource{NamespaceableResourceInterface: client, cache: cached}
}

type cachedNamespaceableResource struct {
	dynamic.NamespaceableResourceInterface
	cache *resourceCache
}

func (r *cachedNamespaceableResource) Namespace(namespace string) dynamic.ResourceInterface {
	return &cachedResource{ResourceInterface: r.NamespaceableResourceInterface.Namespace(namespace), cache: r.cache, namespace: namespace}
}

func (r *cachedNamespaceableResource) List(ctx context.Context, options meta.ListOptions) (*unstructured.UnstructuredList, error) {
	resource := &cachedResource{ResourceInterface: r.NamespaceableResourceInterface, cache: r.cache, namespace: meta.NamespaceAll}

	return resource.List(ctx, options)
}

type cachedResource struct {
	dynamic.ResourceInterface
	cache     *resourceCache
	namespace string
}

func (r *cachedResource) List(ctx context.Context, options meta.ListOptions) (*unstructured.UnstructuredList, error) {
	if objects, ok := r.cache.list(r.namespace, options); ok {
		return &unstructured.UnstructuredList{Items: items[unstructured.Unstructured](objects)}, nil
	}

	return r.ResourceInterface.List(ctx, options)
}
