package cockpit

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/services/cockpit/internal/rybbit"
)

type UserSessions struct {
	URL      string
	Sessions []rybbit.Session
}

// UserSessions returns what the web analytics recorded of one user's visits
// to the platform's own sites.
func (c *Client) UserSessions(ctx context.Context, userID string) (UserSessions, error) {
	if c.analytics == nil {
		return UserSessions{}, c.missingAnalytics()
	}

	sessions, err := c.analytics.Sessions(ctx, userID)

	if err != nil {
		return UserSessions{}, err
	}

	return UserSessions{URL: c.analytics.UserURL(userID), Sessions: sessions}, nil
}

// SessionReplay returns a session's recording as a JSON array of rrweb events.
func (c *Client) SessionReplay(ctx context.Context, sessionID string) (string, error) {
	if c.analytics == nil {
		return "", c.missingAnalytics()
	}

	events, err := c.analytics.Replay(ctx, sessionID)

	if err != nil {
		return "", err
	}

	return string(events), nil
}

func (c *Client) missingAnalytics() error {
	return fmt.Errorf("no Rybbit API key for %s, set rybbit.secrets.apiKey in deployments/lucity-%s/infra-secrets.yaml", c.environment, c.environment)
}
