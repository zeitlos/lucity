package edge

import (
	"context"
	"time"
)

const addressesTTL = time.Hour

func (c *Client) EdgeAddresses(ctx context.Context) ([]string, error) {
	c.addressesMu.Lock()
	defer c.addressesMu.Unlock()

	if c.addresses != nil && time.Since(c.addressesAt) < addressesTTL {
		return c.addresses, nil
	}

	addresses, err := c.api.EdgeServers(ctx)

	if err != nil {
		if c.addresses != nil {
			return c.addresses, nil
		}

		return nil, err
	}

	c.addresses = addresses
	c.addressesAt = time.Now()

	return addresses, nil
}
