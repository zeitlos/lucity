package edge

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/pkg/bunny"
	"github.com/zeitlos/lucity/pkg/to"
)

const (
	maxWebSocketConnections = 500
	monthlyBandwidthLimit   = 1 << 40
)

func (c *Client) zoneName(workspaceID string) string {
	return c.config.ZonePrefix + "-" + workspaceID
}

func (c *Client) ensureZone(ctx context.Context, name string, zone *bunny.PullZone) (*bunny.PullZone, error) {
	if zone == nil {
		found, err := c.api.PullZoneByName(ctx, name)

		if err != nil {
			return nil, err
		}

		zone = found
	}

	if zone == nil {
		created, err := c.api.CreatePullZone(ctx, c.zoneSpec(name))

		if err != nil {
			return nil, err
		}

		zone = created
	}

	for _, rule := range c.edgeRules() {
		if hasEdgeRule(zone, rule.Description) {
			continue
		}

		if err := c.api.AddEdgeRule(ctx, zone.ID, rule); err != nil {
			return nil, fmt.Errorf("zone %s: %w", name, err)
		}
	}

	return zone, nil
}

func (c *Client) zoneSpec(name string) bunny.PullZoneSpec {
	return bunny.PullZoneSpec{
		Name:                            name,
		OriginURL:                       c.config.OriginURL,
		AddHostHeader:                   to.Ptr(true),
		VerifyOriginSSL:                 to.Ptr(false),
		EnableWebSockets:                to.Ptr(true),
		MaxWebSocketConnections:         to.Ptr(maxWebSocketConnections),
		MonthlyBandwidthLimit:           to.Ptr(int64(monthlyBandwidthLimit)),
		EnableSmartCache:                to.Ptr(false),
		CacheControlMaxAgeOverride:      to.Ptr(int64(0)),
		EnableHostnameVary:              to.Ptr(true),
		DisableCookies:                  to.Ptr(false),
		EnableAccessControlOriginHeader: to.Ptr(false),
		EnableTLS1:                      to.Ptr(false),
		EnableTLS1_1:                    to.Ptr(false),
	}
}

func (c *Client) edgeRules() []bunny.EdgeRule {
	matchAll := []bunny.EdgeRuleTrigger{{Type: bunny.TriggerURL, PatternMatches: []string{"*"}, PatternMatchingType: bunny.MatchAny}}

	return []bunny.EdgeRule{
		{
			ActionType:          bunny.ActionSetRequestHeader,
			ActionParameter1:    HeaderName,
			ActionParameter2:    c.config.HeaderSecret,
			Triggers:            matchAll,
			TriggerMatchingType: bunny.MatchAny,
			Description:         "lucity: edge header",
			Enabled:             true,
		},
		{
			ActionType:          bunny.ActionForceSSL,
			Triggers:            matchAll,
			TriggerMatchingType: bunny.MatchAny,
			Description:         "lucity: force ssl",
			Enabled:             true,
		},
	}
}

func hasEdgeRule(zone *bunny.PullZone, description string) bool {
	for _, rule := range zone.EdgeRules {
		if rule.Description == description {
			return true
		}
	}

	return false
}

func hostnameOn(zone *bunny.PullZone, name string) *bunny.Hostname {
	for i := range zone.Hostnames {
		if zone.Hostnames[i].Value == name {
			return &zone.Hostnames[i]
		}
	}

	return nil
}
