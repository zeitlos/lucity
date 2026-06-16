package hostname

import (
	"context"
	"errors"
	"fmt"
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

	zone := registrableDomain(host)

	txtResolver, err := authoritativeResolver(lookupCtx, zone)

	if err != nil {
		return DNSError, err
	}

	txtRecords, err := txtResolver.LookupTXT(lookupCtx, verifyRecordPrefix+zone)

	if err != nil && !isNotFound(err) {
		return DNSError, err
	}

	txtOK := slices.Contains(txtRecords, challenge(workspace, zone))

	routeResolver := txtResolver

	if host != zone {
		routeResolver, err = authoritativeResolver(lookupCtx, host)

		if err != nil {
			return DNSError, err
		}
	}

	var routingOK bool

	if isApex(host) {
		addrs, err := routeResolver.LookupHost(lookupCtx, host)

		if err != nil && !isNotFound(err) {
			return DNSError, err
		}

		routingOK = slices.Contains(addrs, c.customApexIP)
	} else {
		cname, err := routeResolver.LookupCNAME(lookupCtx, host)

		if err != nil && !isNotFound(err) {
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

func authoritativeResolver(ctx context.Context, host string) (*net.Resolver, error) {
	servers, err := authoritativeServers(ctx, host)

	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{Timeout: dnsLookupTimeout}

	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			var lastErr error

			for _, server := range servers {
				conn, err := dialer.DialContext(ctx, network, server)

				if err == nil {
					return conn, nil
				}

				lastErr = err
			}

			if lastErr == nil {
				lastErr = fmt.Errorf("no authoritative nameservers reachable for %q", host)
			}

			return nil, lastErr
		},
	}, nil
}

func authoritativeServers(ctx context.Context, host string) ([]string, error) {
	sys := net.Resolver{}
	name := strings.TrimSuffix(host, ".")

	for {
		nameservers, err := sys.LookupNS(ctx, name)

		if err == nil && len(nameservers) > 0 {
			var servers []string

			for _, ns := range nameservers {
				ips, err := sys.LookupHost(ctx, strings.TrimSuffix(ns.Host, "."))

				if err != nil {
					continue
				}

				for _, ip := range ips {
					servers = append(servers, net.JoinHostPort(ip, "53"))
				}
			}

			if len(servers) > 0 {
				return servers, nil
			}
		}

		i := strings.IndexByte(name, '.')

		if i < 0 {
			return nil, fmt.Errorf("no authoritative nameservers found for %q", host)
		}

		name = name[i+1:]
	}
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
