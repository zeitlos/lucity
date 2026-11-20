// Package bindings watches what the workspace releases render into
// switchboard's namespace: a Secret per chat app, and a ConfigMap per agent
// with the chat channels it talks in.
package bindings

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/pkg/platform"
)

// ChatApp is a workspace's bot in a chat app.
type ChatApp struct {
	ID      platform.ChatAppID
	Kind    string
	URL     string
	Token   string
	Account string
}

// Place is a chat or channel of a chat app.
type Place struct {
	ChatApp platform.ChatAppID
	Channel string
}

// Binding is the agent a chat or channel talks to.
type Binding struct {
	Agent platform.AgentID
	Name  string
}

// Snapshot is everything switchboard serves at one moment.
type Snapshot struct {
	ChatApps map[platform.ChatAppID]ChatApp
	Bindings map[Place]Binding
}

// Agents returns every agent with at least one channel.
func (s Snapshot) Agents() []platform.AgentID {
	seen := map[platform.AgentID]bool{}

	var agents []platform.AgentID

	for _, binding := range s.Bindings {
		if !seen[binding.Agent] {
			seen[binding.Agent] = true
			agents = append(agents, binding.Agent)
		}
	}

	return agents
}

type Watcher struct {
	kubernetes kubernetes.Interface
	namespace  string

	mu         sync.Mutex
	secrets    cache.Store
	configMaps cache.Store
	changed    chan struct{}
}

func New(kubernetes kubernetes.Interface, namespace string) *Watcher {
	return &Watcher{kubernetes: kubernetes, namespace: namespace, changed: make(chan struct{}, 1)}
}

// Run watches until ctx is done. Changed fires after every change, coalesced.
func (w *Watcher) Run(ctx context.Context) error {
	secrets := w.informer(ctx, labels.ChatApp, func(factory informers.SharedInformerFactory) cache.SharedIndexInformer {
		return factory.Core().V1().Secrets().Informer()
	})

	configMaps := w.informer(ctx, labels.AgentChannels, func(factory informers.SharedInformerFactory) cache.SharedIndexInformer {
		return factory.Core().V1().ConfigMaps().Informer()
	})

	w.mu.Lock()
	w.secrets = secrets.GetStore()
	w.configMaps = configMaps.GetStore()
	w.mu.Unlock()

	if !cache.WaitForCacheSync(ctx.Done(), secrets.HasSynced, configMaps.HasSynced) {
		return ctx.Err()
	}

	w.notify()

	<-ctx.Done()

	return nil
}

func (w *Watcher) informer(ctx context.Context, label string, pick func(informers.SharedInformerFactory) cache.SharedIndexInformer) cache.SharedIndexInformer {
	factory := informers.NewSharedInformerFactoryWithOptions(w.kubernetes, 10*time.Minute,
		informers.WithNamespace(w.namespace),
		informers.WithTweakListOptions(func(options *meta.ListOptions) {
			options.LabelSelector = labels.Selector(label, "true")
		}),
	)

	informer := pick(factory)

	_, _ = informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(any) { w.notify() },
		UpdateFunc: func(any, any) { w.notify() },
		DeleteFunc: func(any) { w.notify() },
	})

	factory.Start(ctx.Done())

	return informer
}

func (w *Watcher) notify() {
	select {
	case w.changed <- struct{}{}:
	default:
	}
}

// Changed fires after the bindings changed.
func (w *Watcher) Changed() <-chan struct{} {
	return w.changed
}

// Snapshot reads the current chat apps and bindings. An object that doesn't
// parse is skipped and logged, so one bad release never stops the others.
func (w *Watcher) Snapshot() Snapshot {
	w.mu.Lock()
	secrets, configMaps := w.secrets, w.configMaps
	w.mu.Unlock()

	snapshot := Snapshot{ChatApps: map[platform.ChatAppID]ChatApp{}, Bindings: map[Place]Binding{}}

	if secrets == nil || configMaps == nil {
		return snapshot
	}

	for _, object := range secrets.List() {
		secret, ok := object.(*core.Secret)

		if !ok {
			continue
		}

		workspace := secret.Labels[labels.Workspace]
		name := string(secret.Data["name"])

		if workspace == "" || name == "" || len(secret.Data["token"]) == 0 {
			slog.Warn("bindings: skipping an incomplete chat app", "secret", secret.Name)
			continue
		}

		id := platform.ChatAppID{Workspace: workspace, Name: name}

		snapshot.ChatApps[id] = ChatApp{
			ID:      id,
			Kind:    string(secret.Data["kind"]),
			URL:     string(secret.Data["url"]),
			Token:   string(secret.Data["token"]),
			Account: string(secret.Data["account"]),
		}
	}

	for _, object := range configMaps.List() {
		configMap, ok := object.(*core.ConfigMap)

		if !ok {
			continue
		}

		agent, err := platform.ParseAgentID(configMap.Data["agent"])

		if err != nil || agent.Workspace != configMap.Labels[labels.Workspace] {
			slog.Warn("bindings: skipping agent channels with an invalid agent", "configmap", configMap.Name)
			continue
		}

		var channels []struct {
			ChatApp string `json:"chatApp"`
			Channel string `json:"channel"`
			Name    string `json:"name"`
		}

		if err := json.Unmarshal([]byte(configMap.Data["channels"]), &channels); err != nil {
			slog.Warn("bindings: skipping unreadable agent channels", "configmap", configMap.Name, "error", err)
			continue
		}

		for _, channel := range channels {
			place := Place{ChatApp: platform.ChatAppID{Workspace: agent.Workspace, Name: channel.ChatApp}, Channel: channel.Channel}
			snapshot.Bindings[place] = Binding{Agent: agent, Name: channel.Name}
		}
	}

	return snapshot
}
