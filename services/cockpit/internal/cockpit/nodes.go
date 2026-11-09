package cockpit

import (
	"context"
	"fmt"
	"sort"
	"time"

	"golang.org/x/sync/errgroup"
	core "k8s.io/api/core/v1"
	storage "k8s.io/api/storage/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	labelNodePool     = "lucity.dev/node-pool"
	labelInstanceType = "node.kubernetes.io/instance-type"
)

// NodePool is the nodes that share a node-pool label. Name is empty for
// nodes without one.
type NodePool struct {
	Name  string
	Nodes []Node
}

// Node is one node's room. Requests count every pod that isn't finished. The
// used values are nil when the metrics have no recent sample.
type Node struct {
	Name              string
	InstanceType      string
	CreatedAt         time.Time
	AttachedVolumes   int
	VolumeLimit       int
	Pods              int
	PodLimit          int
	MemoryAllocatable int64
	MemoryRequested   int64
	MemoryUsed        *int64
	CPUAllocatable    float64
	CPURequested      float64
	CPUUsed           *float64
}

// NodePools returns the nodes grouped by pool, pools and nodes by name.
func (c *Client) NodePools(ctx context.Context) ([]NodePool, error) {
	var nodes *core.NodeList
	var pods *core.PodList
	var attachments *storage.VolumeAttachmentList
	var drivers *storage.CSINodeList
	memoryUsed := map[string]int64{}
	cpuUsed := map[string]float64{}

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		var err error
		nodes, err = c.kubernetes.CoreV1().Nodes().List(ctx, meta.ListOptions{})

		return wrapList("nodes", err)
	})

	group.Go(func() error {
		var err error
		pods, err = c.kubernetes.CoreV1().Pods("").List(ctx, meta.ListOptions{FieldSelector: "status.phase!=Succeeded,status.phase!=Failed"})

		return wrapList("pods", err)
	})

	group.Go(func() error {
		var err error
		attachments, err = c.kubernetes.StorageV1().VolumeAttachments().List(ctx, meta.ListOptions{})

		return wrapList("volume attachments", err)
	})

	group.Go(func() error {
		var err error
		drivers, err = c.kubernetes.StorageV1().CSINodes().List(ctx, meta.ListOptions{})

		return wrapList("CSI nodes", err)
	})

	group.Go(func() error {
		samples, err := c.metrics.QueryVector(ctx, `max by (k8s_node_name) (k8s_node_memory_working_set_bytes)`, time.Now())

		for _, sample := range samples {
			memoryUsed[sample.Labels["k8s_node_name"]] = int64(sample.Value)
		}

		return wrap("node memory", err)
	})

	group.Go(func() error {
		samples, err := c.metrics.QueryVector(ctx, `max by (k8s_node_name) (k8s_node_cpu_usage)`, time.Now())

		for _, sample := range samples {
			cpuUsed[sample.Labels["k8s_node_name"]] = sample.Value
		}

		return wrap("node CPU", err)
	})

	if err := group.Wait(); err != nil {
		return nil, err
	}

	byName := map[string]*Node{}
	pools := map[string]*NodePool{}

	for _, item := range nodes.Items {
		node := &Node{
			Name:              item.Name,
			InstanceType:      item.Labels[labelInstanceType],
			CreatedAt:         item.CreationTimestamp.Time,
			PodLimit:          int(item.Status.Allocatable.Pods().Value()),
			MemoryAllocatable: item.Status.Allocatable.Memory().Value(),
			CPUAllocatable:    item.Status.Allocatable.Cpu().AsApproximateFloat64(),
		}

		if used, ok := memoryUsed[item.Name]; ok {
			node.MemoryUsed = &used
		}

		if used, ok := cpuUsed[item.Name]; ok {
			node.CPUUsed = &used
		}

		byName[item.Name] = node
	}

	for _, pod := range pods.Items {
		node := byName[pod.Spec.NodeName]

		if node == nil {
			continue
		}

		memory, cpu := podRequests(pod)
		node.Pods++
		node.MemoryRequested += memory
		node.CPURequested += cpu
	}

	for _, attachment := range attachments.Items {
		if node := byName[attachment.Spec.NodeName]; node != nil {
			node.AttachedVolumes++
		}
	}

	for _, item := range drivers.Items {
		node := byName[item.Name]

		if node == nil {
			continue
		}

		for _, driver := range item.Spec.Drivers {
			if driver.Allocatable != nil && driver.Allocatable.Count != nil {
				node.VolumeLimit += int(*driver.Allocatable.Count)
			}
		}
	}

	for _, item := range nodes.Items {
		name := item.Labels[labelNodePool]

		if pools[name] == nil {
			pools[name] = &NodePool{Name: name}
		}

		pools[name].Nodes = append(pools[name].Nodes, *byName[item.Name])
	}

	result := make([]NodePool, 0, len(pools))

	for _, pool := range pools {
		sort.Slice(pool.Nodes, func(i, j int) bool {
			return pool.Nodes[i].Name < pool.Nodes[j].Name
		})

		result = append(result, *pool)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

// podRequests is what the scheduler reserves for the pod: the larger of its
// containers together and its largest init container.
func podRequests(pod core.Pod) (int64, float64) {
	var memory int64
	var cpu float64

	for _, container := range pod.Spec.Containers {
		memory += container.Resources.Requests.Memory().Value()
		cpu += container.Resources.Requests.Cpu().AsApproximateFloat64()
	}

	for _, container := range pod.Spec.InitContainers {
		memory = max(memory, container.Resources.Requests.Memory().Value())
		cpu = max(cpu, container.Resources.Requests.Cpu().AsApproximateFloat64())
	}

	return memory, cpu
}

func wrapList(what string, err error) error {
	if err != nil {
		return fmt.Errorf("list %s: %w", what, err)
	}

	return nil
}
