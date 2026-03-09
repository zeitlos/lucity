package handler

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/deployer"
)

// ServiceLogEntry represents a single runtime log line from a running pod.
type ServiceLogEntry struct {
	Line string
	Pod  string
}

// ServiceLogs returns a channel of log entries from running pods.
// The channel is closed when the stream ends or the context is cancelled.
func (c *Client) ServiceLogs(ctx context.Context, projectID, service, environment string, tailLines *int) (<-chan ServiceLogEntry, error) {
	ctx = auth.OutgoingContext(ctx)

	req := &deployer.ServiceLogsRequest{
		Project:     projectID,
		Environment: environment,
		Service:     service,
		TailLines:   1000,
	}
	if tailLines != nil {
		req.TailLines = int32(*tailLines)
	}

	stream, err := c.Deployer.ServiceLogs(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to open service logs stream: %w", err)
	}

	out := make(chan ServiceLogEntry, 128)
	go func() {
		defer close(out)
		for {
			entry, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				slog.Warn("service log stream ended", "project", projectID, "service", service, "error", err)
				return
			}
			select {
			case out <- ServiceLogEntry{Line: entry.Line, Pod: entry.Pod}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, nil
}
