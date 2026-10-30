package conductor

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/platform"
)

// SetServiceScaling applies either fixed replicas or HPA-driven autoscaling.
// The two are mutually exclusive in the values shape: SetAutoscaling sets the
// HPA config; SetReplicas wipes any HPA and pins a fixed count.
func (c *Client) SetServiceScaling(ctx context.Context, serviceID platform.ServiceID, replicas int, autoscaling *AutoscalingConfig) (*Service, error) {
	capabilities, err := c.Capabilities(ctx, serviceID.Workspace)

	if err != nil {
		return nil, err
	}

	current, err := c.platform.Service(ctx, serviceID)

	if err != nil {
		return nil, err
	}

	largest := current.Replicas.Desired

	if current.Autoscaling != nil {
		largest = max(largest, current.Autoscaling.MaxReplicas)
	}

	if autoscaling != nil && autoscaling.Enabled {
		if err := capabilities.Services.checkReplicas(autoscaling.MaxReplicas, largest); err != nil {
			return nil, err
		}

		if _, err := c.deployer.Services().SetAutoscaling(ctx, serviceID, deployer.Autoscaling{
			MinReplicas: autoscaling.MinReplicas,
			MaxReplicas: autoscaling.MaxReplicas,
			TargetCPU:   autoscaling.TargetCPU,
		}); err != nil {
			return nil, fmt.Errorf("set autoscaling: %w", err)
		}
	} else {
		if err := capabilities.Services.checkReplicas(replicas, largest); err != nil {
			return nil, err
		}

		if _, err := c.deployer.Services().SetReplicas(ctx, serviceID, replicas); err != nil {
			return nil, fmt.Errorf("set replicas: %w", err)
		}
	}

	return c.Service(ctx, serviceID)
}
