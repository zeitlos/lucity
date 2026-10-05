package conductor

import (
	"context"
	"net/netip"
)

type clientAddressKey struct{}

var sharedAddressSpace = netip.MustParsePrefix("100.64.0.0/10")

// WithClientAddress records the address a request reached the platform from.
func WithClientAddress(ctx context.Context, address netip.Addr) context.Context {
	return context.WithValue(ctx, clientAddressKey{}, address)
}

// ClientAddress returns the public IPv4 address the request reached the
// platform from, and false when there is none, such as for a request from
// a private network.
func (c *Client) ClientAddress(ctx context.Context) (netip.Addr, bool) {
	address, ok := ctx.Value(clientAddressKey{}).(netip.Addr)

	if !ok {
		return netip.Addr{}, false
	}

	address = address.Unmap()

	if !address.Is4() || !isPublicAddress(address) {
		return netip.Addr{}, false
	}

	return address, true
}

func isPublicAddress(address netip.Addr) bool {
	return address.IsGlobalUnicast() && !address.IsPrivate() && !sharedAddressSpace.Contains(address)
}
