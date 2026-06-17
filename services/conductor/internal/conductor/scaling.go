package conductor

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

// SetServiceScaling applies either fixed replicas or HPA-driven autoscaling.
// The two are mutually exclusive in the values shape: SetAutoscaling sets the
// HPA config; SetReplicas wipes any HPA and pins a fixed count.
func (c *Client) SetServiceScaling(ctx context.Context, serviceID platform.ServiceID, replicas int, autoscaling *AutoscalingConfig) (*Service, error) {
	if autoscaling != nil && autoscaling.Enabled {
		if _, err := c.deployer.Services().SetAutoscaling(ctx, serviceID, deployer.Autoscaling{
			MinReplicas: autoscaling.MinReplicas,
			MaxReplicas: autoscaling.MaxReplicas,
			TargetCPU:   autoscaling.TargetCPU,
		}); err != nil {
			return nil, fmt.Errorf("set autoscaling: %w", err)
		}
	} else {
		if _, err := c.deployer.Services().SetReplicas(ctx, serviceID, replicas); err != nil {
			return nil, fmt.Errorf("set replicas: %w", err)
		}
	}

	return c.Service(ctx, serviceID)
}
