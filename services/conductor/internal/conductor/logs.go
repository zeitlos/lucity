package conductor

import (
	"context"
	"time"

	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/services/conductor/internal/servicelogs"
)

type (
	ServiceLogEntry     = servicelogs.Entry
	ServiceLogFilter    = servicelogs.Filter
	ServiceLogVolume    = servicelogs.Volume
	ServiceLogDirection = servicelogs.Direction
)

// ServiceLogInstance is an instance that wrote lines in a range, with its
// state right now.
type ServiceLogInstance struct {
	servicelogs.Instance
	State platform.InstanceState
}

const (
	serviceLogLimit = 1000
	serviceLogSlots = 60
	serviceLogTail  = 24 * time.Hour
)

// ServiceLogs returns the log entries of a service between from and to,
// oldest first.
func (c *Client) ServiceLogs(ctx context.Context, service platform.ServiceID, from, to time.Time, filter ServiceLogFilter, surrounding *int, limit *int, direction ServiceLogDirection) ([]ServiceLogEntry, error) {
	from, err := c.serviceLogsFrom(ctx, service, from)

	if err != nil {
		return nil, err
	}

	lines := serviceLogLimit

	if limit != nil {
		lines = *limit
	}

	around := 0

	if surrounding != nil {
		around = *surrounding
	}

	return c.serviceLogs.Entries(ctx, service, from, to, filter, around, lines, direction)
}

// ServiceLogInstances returns the instances of a service that wrote lines
// between from and to.
func (c *Client) ServiceLogInstances(ctx context.Context, service platform.ServiceID, from, to time.Time) ([]ServiceLogInstance, error) {
	from, err := c.serviceLogsFrom(ctx, service, from)

	if err != nil {
		return nil, err
	}

	instances, err := c.serviceLogs.Instances(ctx, service, from, to)

	if err != nil {
		return nil, err
	}

	live, err := c.cachedPlatform.Instances(ctx, service)

	if err != nil {
		return nil, err
	}

	states := make(map[string]platform.InstanceState, len(live))

	for _, instance := range live {
		states[instance.Name] = instance.State
	}

	result := make([]ServiceLogInstance, 0, len(instances))

	for _, instance := range instances {
		state, ok := states[instance.Name]

		if !ok {
			state = platform.InstanceStopped
		}

		result = append(result, ServiceLogInstance{Instance: instance, State: state})
	}

	return result, nil
}

// ServiceLogVolume counts the lines of a service between from and to in
// slots of equal length.
func (c *Client) ServiceLogVolume(ctx context.Context, service platform.ServiceID, from, to time.Time, filter ServiceLogFilter, slots *int) ([]ServiceLogVolume, error) {
	from, err := c.serviceLogsFrom(ctx, service, from)

	if err != nil {
		return nil, err
	}

	count := serviceLogSlots

	if slots != nil {
		count = *slots
	}

	return c.serviceLogs.Volume(ctx, service, from, to, filter, count)
}

// TailServiceLogs sends the last entries of the past day, then every new one
// until ctx is done.
func (c *Client) TailServiceLogs(ctx context.Context, service platform.ServiceID, filter ServiceLogFilter, tailLines *int) (<-chan ServiceLogEntry, error) {
	from, err := c.serviceLogsFrom(ctx, service, time.Now().Add(-serviceLogTail))

	if err != nil {
		return nil, err
	}

	lines := serviceLogLimit

	if tailLines != nil {
		lines = *tailLines
	}

	return c.serviceLogs.Tail(ctx, service, from, filter, lines)
}

// serviceLogsFrom moves from up to the creation of the environment, so a
// recreated environment never shows the logs of the one before it.
func (c *Client) serviceLogsFrom(ctx context.Context, service platform.ServiceID, from time.Time) (time.Time, error) {
	environment, err := c.cachedPlatform.Environment(ctx, service.EnvironmentID())

	if err != nil {
		return time.Time{}, err
	}

	if from.Before(environment.CreatedAt) {
		return environment.CreatedAt, nil
	}

	return from, nil
}
