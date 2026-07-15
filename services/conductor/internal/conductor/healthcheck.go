package conductor

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type HealthCheckConfig struct {
	Path                    string
	Port                    int
	InitialDelaySeconds     int
	PeriodSeconds           int
	TimeoutSeconds          int
	FailureThreshold        int
	StartupFailureThreshold int
}

// SetServiceHealthCheck configures the readiness/startup probe for a service.
// A nil config clears it, reverting to the default TCP probe.
func (c *Client) SetServiceHealthCheck(ctx context.Context, serviceID platform.ServiceID, config *HealthCheckConfig) (*Service, error) {
	var healthCheck *deployer.HealthCheck

	if config != nil {
		healthCheck = &deployer.HealthCheck{
			Path:                    config.Path,
			Port:                    config.Port,
			InitialDelaySeconds:     config.InitialDelaySeconds,
			PeriodSeconds:           config.PeriodSeconds,
			TimeoutSeconds:          config.TimeoutSeconds,
			FailureThreshold:        config.FailureThreshold,
			StartupFailureThreshold: config.StartupFailureThreshold,
		}
	}

	if _, err := c.deployer.Services().SetHealthCheck(ctx, serviceID, healthCheck); err != nil {
		return nil, fmt.Errorf("set health check: %w", err)
	}

	return c.Service(ctx, serviceID)
}
