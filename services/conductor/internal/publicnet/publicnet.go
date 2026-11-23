// Package publicnet dials only public addresses, so a URL or registry a user
// supplies can't point the conductor at the cluster or the cloud's metadata
// service.
package publicnet

import (
	"errors"
	"net"
	"net/netip"
	"syscall"
	"time"
)

var ErrNotPublic = errors.New("address is not public")

var blockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("2002::/16"),
}

// Dialer refuses every connection whose resolved address isn't public.
func Dialer(timeout time.Duration) *net.Dialer {
	return &net.Dialer{
		Timeout: timeout,
		Control: func(_, address string, _ syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)

			if err != nil {
				return err
			}

			addr, err := netip.ParseAddr(host)

			if err != nil || !isPublic(addr) {
				return ErrNotPublic
			}

			return nil
		},
	}
}

func isPublic(addr netip.Addr) bool {
	addr = addr.Unmap()

	if !addr.IsGlobalUnicast() || addr.IsPrivate() || addr.IsLoopback() || addr.IsLinkLocalUnicast() {
		return false
	}

	for _, prefix := range blockedPrefixes {
		if prefix.Contains(addr) {
			return false
		}
	}

	return true
}
