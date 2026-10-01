package bunny

import (
	"context"
	"fmt"
	"net/http"
)

func (c *Client) EdgeServers(ctx context.Context) ([]string, error) {
	var addresses []string

	for _, path := range []string{"/system/edgeserverlist", "/system/edgeserverlist/ipv6"} {
		var list []string

		if err := c.do(ctx, http.MethodGet, path, nil, &list); err != nil {
			return nil, fmt.Errorf("edge server list: %w", err)
		}

		addresses = append(addresses, list...)
	}

	return addresses, nil
}
