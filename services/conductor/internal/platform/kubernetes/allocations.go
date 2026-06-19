package kubernetes

import (
	"context"
	"log/slog"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

// ResourceAllocations enumerates every Lucity-managed namespace and returns
// the actual resource usage (summed pod requests + PVC sizes) plus the
// declared tier label. Cashier consumes this for metering.
func (c *Client) ResourceAllocations(ctx context.Context) ([]platform.ResourceAllocation, error) {
	nsList, err := c.kubernetes.CoreV1().Namespaces().List(ctx, meta.ListOptions{
		LabelSelector: resourceTierLabel,
	})

	if err != nil {
		return nil, err
	}

	out := make([]platform.ResourceAllocation, 0, len(nsList.Items))

	for _, ns := range nsList.Items {
		id := environmentID(ns.Labels)

		if id.Workspace == "" || id.Project == "" || id.Name == "" {
			continue
		}

		cpu, mem, disk := c.namespaceAllocations(ctx, ns.Name)

		out = append(out, platform.ResourceAllocation{
			EnvironmentID: id,
			Namespace:     ns.Name,
			Tier:          tierFromLabel(ns.Labels[resourceTierLabel]),
			CPUMillicores: cpu,
			MemoryMB:      mem,
			DiskMB:        disk,
		})
	}

	return out, nil
}

// namespaceAllocations sums running-pod container requests and PVC storage
// requests. Reflects deployed reality, not the ResourceQuota ceiling.
func (c *Client) namespaceAllocations(ctx context.Context, namespace string) (cpu, mem, disk int) {
	pods, err := c.kubernetes.CoreV1().Pods(namespace).List(ctx, meta.ListOptions{
		FieldSelector: "status.phase=Running",
	})

	if err != nil {
		slog.Warn("list pods for allocations failed", "namespace", namespace, "error", err)
		return 0, 0, 0
	}

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			if req, ok := container.Resources.Requests[core.ResourceCPU]; ok {
				cpu += int(req.MilliValue())
			}

			if req, ok := container.Resources.Requests[core.ResourceMemory]; ok {
				mem += int(req.Value() / (1024 * 1024))
			}
		}
	}

	pvcs, err := c.kubernetes.CoreV1().PersistentVolumeClaims(namespace).List(ctx, meta.ListOptions{})

	if err != nil {
		slog.Warn("list pvcs for allocations failed", "namespace", namespace, "error", err)
		return cpu, mem, 0
	}

	for _, pvc := range pvcs.Items {
		if req, ok := pvc.Spec.Resources.Requests[core.ResourceStorage]; ok {
			disk += int(req.Value() / (1024 * 1024))
		}
	}

	return cpu, mem, disk
}

func tierFromLabel(label string) platform.ResourceTier {
	switch label {
	case resourceTierProd:
		return platform.ProductionTier
	default:
		return platform.EcoTier
	}
}
