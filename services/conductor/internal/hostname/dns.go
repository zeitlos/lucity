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

	txtOK, err := checkTXT(lookupCtx, resolver, verifyRecordPrefix+host, challenge(workspace, host))

	if err != nil {
		return DNSError, err
	}

	routingOK, err := c.checkRouting(lookupCtx, resolver, host)

	if err != nil {
		return DNSError, err
	}

	if txtOK && routingOK {
		return DNSValid, nil
	}

	if !txtOK && !routingOK {
		return DNSPending, nil
	}

	return DNSMisconfigured, nil
}

func checkTXT(ctx context.Context, resolver *net.Resolver, host, expected string) (bool, error) {
	records, err := resolver.LookupTXT(ctx, host)

	if dnsNoSuchHost(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return slices.Contains(records, expected), nil
}

func (c *Client) checkRouting(ctx context.Context, resolver *net.Resolver, host string) (bool, error) {
	if isApex(host) {
		addrs, err := resolver.LookupHost(ctx, host)

		if dnsNoSuchHost(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return slices.Contains(addrs, c.customApexIP), nil
	}

	cname, err := resolver.LookupCNAME(ctx, host)

	if dnsNoSuchHost(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return strings.EqualFold(strings.TrimSuffix(cname, "."), c.customCNAMETarget), nil
}

func dnsNoSuchHost(err error) bool {
	if err == nil {
		return false
	}

	var dnsErr *net.DNSError

	if errors.As(err, &dnsErr) {
		return dnsErr.IsNotFound
	}

	return false
}
