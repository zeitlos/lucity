package conductor

import (
	"context"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

// ServiceLogEntry is re-exported from the platform package so resolvers
// don't have to import it directly.
type ServiceLogEntry = platform.LogEntry

// ServiceLogs returns a channel of log entries from running pods. The
// channel closes when all pod streams end or ctx is cancelled.
func (c *Client) ServiceLogs(ctx context.Context, service platform.ServiceID, tailLines *int) (<-chan ServiceLogEntry, error) {
	tail := 1000

	if tailLines != nil {
		tail = *tailLines
	}

	return c.platform.Logs(ctx, service, tail)
}
