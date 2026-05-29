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

func (c *Client) Status(ctx context.Context, namespace, workspace, host string) (Status, error) {
	if c.IsPlatform(host) {
		return Status{DNS: DNSValid, TLS: TLSActive}, nil
	}

	dnsState, err := c.dnsState(ctx, workspace, host)

	if err != nil {
		return Status{}, err
	}

	// TODO: query cert-manager Certificate (CRD) in `namespace` to populate
	// TLS state. cert-manager 1.20 with XListenerSet auto-creates a
	// Certificate per listener with the secretRef name; map its Ready
	// condition to TLSActive/Provisioning/Error. Until that lands, surface
	// TLSNone for unverified and a placeholder for verified.
	tlsState := TLSNone

	if dnsState == DNSValid {
		tlsState = TLSProvisioning
	}

	_ = namespace

	return Status{DNS: dnsState, TLS: tlsState}, nil
}

func (c *Client) dnsState(ctx context.Context, workspace, host string) (DNSState, error) {
	lookupCtx, cancel := context.WithTimeout(ctx, dnsLookupTimeout)
	defer cancel()

	resolver := &net.Resolver{}

	txtOK, txtErr := checkTXT(lookupCtx, resolver, verifyRecordPrefix+host, challenge(workspace, host))

	if txtErr != nil {
		return DNSError, txtErr
	}

	routingOK, routingErr := c.checkRouting(lookupCtx, resolver, host)

	if routingErr != nil {
		return DNSError, routingErr
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
