package conductor

import (
	"context"
	"errors"
)

// Eject produces a zip archive of the project ready to run on a standalone
// Kubernetes cluster.
//
// TODO(step-8): not yet implemented for the helm-backed flow.
func (c *Client) Eject(ctx context.Context, projectID string) ([]byte, error) {
	return nil, errors.New("eject not yet implemented for the helm backend")
}
