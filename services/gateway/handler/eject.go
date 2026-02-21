package handler

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/packager"
)

// Eject produces a zip archive of the ejected project via the packager service.
func (c *Client) Eject(ctx context.Context, projectID string) ([]byte, error) {
	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcLongTimeout)
	defer cancel()

	resp, err := c.Packager.Eject(callCtx, &packager.EjectRequest{
		Project: projectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to eject project: %w", err)
	}

	return resp.Archive, nil
}
