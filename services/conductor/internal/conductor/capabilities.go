package conductor

import "context"

type Capabilities struct {
	Agents bool
}

// Capabilities reports what a workspace can use. Every guard checks the same
// answer, so a client never offers what the conductor would refuse.
func (c *Client) Capabilities(ctx context.Context, workspaceID string) Capabilities {
	return Capabilities{Agents: c.config.AgentsEnabled}
}
