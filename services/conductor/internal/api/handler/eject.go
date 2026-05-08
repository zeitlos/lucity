package handler

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/pkg/tenant"
)

// Eject produces a zip archive of the ejected project via the packager.
func (c *Client) Eject(ctx context.Context, projectID string) ([]byte, error) {
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	callCtx, cancel := context.WithTimeout(ctx, grpcLongTimeout)
	defer cancel()

	archive, err := c.Packager.Eject(callCtx, ws, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to eject project: %w", err)
	}
	return archive, nil
}
