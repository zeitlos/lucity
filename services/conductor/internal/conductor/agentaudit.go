package conductor

import (
	"context"
	"time"

	"github.com/zeitlos/lucity/pkg/usererror"
	"github.com/zeitlos/lucity/services/conductor/internal/agentaudit"
)

type AgentAuditEntry = agentaudit.Entry

const (
	defaultAuditWindow = 24 * time.Hour
	maxAuditWindow     = 30 * 24 * time.Hour
)

func (c *Client) AgentAudit(ctx context.Context, agent AgentID, kinds []agentaudit.Kind, since *time.Time) ([]AgentAuditEntry, error) {
	if _, _, err := c.loadAgent(ctx, agent); err != nil {
		return nil, err
	}

	from := time.Now().Add(-defaultAuditWindow)

	if since != nil {
		from = *since
	}

	if time.Since(from) > maxAuditWindow {
		return nil, usererror.New("the audit reaches back at most 30 days")
	}

	entries, err := c.agentAudit.Entries(ctx, agent, kinds, from)

	if agentaudit.Disabled(err) {
		return []AgentAuditEntry{}, nil
	}

	return entries, err
}
