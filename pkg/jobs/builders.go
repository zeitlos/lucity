package jobs

import (
	"slices"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

const builderSelector = "app.kubernetes.io/name=buildkit"

// Builders tracks the BuildKit pods that build Jobs can be placed on. Their
// pod specs are dropped, so the cached pods carry metadata and status only.
type Builders struct {
	factory  informers.SharedInformerFactory
	informer cache.SharedIndexInformer
	lister   corelisters.PodNamespaceLister
}

func NewBuilders(client kubernetes.Interface, namespace string) *Builders {
	factory := informers.NewSharedInformerFactoryWithOptions(client, 0,
		informers.WithNamespace(namespace),
		informers.WithTweakListOptions(func(options *meta.ListOptions) {
			options.LabelSelector = builderSelector
		}),
		informers.WithTransform(stripPod),
	)

	pods := factory.Core().V1().Pods()

	return &Builders{
		factory:  factory,
		informer: pods.Informer(),
		lister:   pods.Lister().Pods(namespace),
	}
}

func (b *Builders) Start(stop <-chan struct{}) {
	b.factory.Start(stop)
}

func (b *Builders) HasSynced() bool {
	return b.informer.HasSynced()
}

// OnChange calls notify after every add, update or delete of a builder pod.
func (b *Builders) OnChange(notify func()) error {
	_, err := b.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(any) { notify() },
		UpdateFunc: func(any, any) { notify() },
		DeleteFunc: func(any) { notify() },
	})

	return err
}

// Ready returns the names of the builder pods that are ready and not shutting
// down, sorted by name.
func (b *Builders) Ready() ([]string, error) {
	pods, err := b.lister.List(k8slabels.Everything())

	if err != nil {
		return nil, err
	}

	var ready []string

	for _, pod := range pods {
		if pod.DeletionTimestamp == nil && podReady(pod) {
			ready = append(ready, pod.Name)
		}
	}

	slices.Sort(ready)

	return ready, nil
}

func podReady(pod *core.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == core.PodReady {
			return condition.Status == core.ConditionTrue
		}
	}

	return false
}

func stripPod(object any) (any, error) {
	if pod, ok := object.(*core.Pod); ok {
		pod.ManagedFields = nil
		pod.Spec = core.PodSpec{}
	}

	return object, nil
}
