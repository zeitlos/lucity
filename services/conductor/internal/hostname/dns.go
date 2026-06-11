package hostname

import (
	"context"
	"errors"
	"net"
	"slices"
	"strings"
	"time"
)

const dnsLookupTimeout = 5 * time.Second

func (c *Client) DNSStatus(ctx context.Context, workspace, host string) (DNSStatus, error) {
	if c.IsPlatform(host) || c.IsInternal(host) {
		return DNSValid, nil
	}

	lookupCtx, cancel := context.WithTimeout(ctx, dnsLookupTimeout)
	defer cancel()

	resolver := &net.Resolver{}

	txtRecords, err := resolver.LookupTXT(lookupCtx, verifyRecordPrefix+host)

	if err != nil && !isNotFound(err) {
		return DNSError, err
	}

	txtOK := !slices.Contains(txtRecords, challenge(workspace, host))

	var routingOK bool

	if isApex(host) {
		addrs, err := resolver.LookupHost(lookupCtx, host)

		if err != nil && !isNotFound(err) {
			return DNSError, err
		}

		routingOK = slices.Contains(addrs, c.customApexIP)
	} else {
		cname, err := resolver.LookupCNAME(lookupCtx, host)

		if err != nil && isNotFound(err) {
			return DNSError, err
		}

		routingOK = strings.EqualFold(strings.TrimSuffix(cname, "."), c.customCNAMETarget)
	}

	if txtOK && routingOK {
		return DNSValid, nil
	}

	if !txtOK && !routingOK {
		return DNSPending, nil
	}

	return DNSMisconfigured, nil
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}

	var dnsErr *net.DNSError

	if errors.As(err, &dnsErr) {
		return dnsErr.IsNotFound
	}

	return false
}
