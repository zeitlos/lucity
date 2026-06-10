package conductor

import (
	"context"
	"errors"
)

// Eject produces a zip archive of the project ready to run on a standalone
// Kubernetes cluster.
//
// TODO(step-8): not yet implemented for the helm-backed flow. The legacy
// path generated the archive from the Soft-serve GitOps repo; the new flow
// will render the helm chart + values into a self-contained tree (chart/,
// environments/, argocd/, README.md). Scheduled for step 8 of the refactor.
func (c *Client) Eject(ctx context.Context, projectID string) ([]byte, error) {
	return nil, errors.New("eject not yet implemented for the helm backend")
}
