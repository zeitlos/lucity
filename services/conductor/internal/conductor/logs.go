package conductor

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/pkg/tenant"
	inprocdeployer "github.com/zeitlos/lucity/services/conductor/internal/inproc/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

// ServiceLogEntry is re-exported from the inproc deployer so resolvers
// don't have to know about that package.
type ServiceLogEntry = inprocdeployer.ServiceLogEntry

// ServiceLogs returns a channel of log entries from running pods.
// The channel is closed when all pod streams end or ctx is cancelled.
func (c *Client) ServiceLogs(ctx context.Context, service platform.ServiceID, tailLines *int) (<-chan ServiceLogEntry, error) {
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	tail := 1000
	if tailLines != nil {
		tail = *tailLines
	}
	out, err := c.Deployer.ServiceLogs(ctx, ws, service.Project, service.Environment, service.Name, tail)
	if err != nil {
		return nil, fmt.Errorf("failed to open service logs stream: %w", err)
	}
	return out, nil
}
