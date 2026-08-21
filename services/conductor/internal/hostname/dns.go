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

const (
	dnsLookupTimeout  = 15 * time.Second
	dnsAttemptTimeout = 3 * time.Second
)

func (c *Client) DNSStatus(ctx context.Context, workspace, host string) (DNSStatus, error) {
	if c.IsPlatform(host) || c.IsInternal(host) {
		return DNSValid, nil
	}

	lookupCtx, cancel := context.WithTimeout(ctx, dnsLookupTimeout)
	defer cancel()

	zone := registrableDomain(host)

	zoneServers, err := authoritativeServers(lookupCtx, zone)

	if err != nil {
		return DNSError, err
	}

	txtRecords, err := firstAnswer(lookupCtx, zoneServers, func(ctx context.Context, resolver *net.Resolver) ([]string, error) {
		return resolver.LookupTXT(ctx, verifyRecordPrefix+zone)
	})

	if err != nil && !isNotFound(err) {
		return DNSError, err
	}

	txtOK := slices.Contains(txtRecords, challenge(workspace, zone))

	routeServers := zoneServers

	if host != zone {
		routeServers, err = authoritativeServers(lookupCtx, host)

		if err != nil {
			return DNSError, err
		}
	}

	var routingOK bool

	if isApex(host) {
		addrs, err := firstAnswer(lookupCtx, routeServers, func(ctx context.Context, resolver *net.Resolver) ([]string, error) {
			return resolver.LookupHost(ctx, host)
		})

		if err != nil && !isNotFound(err) {
			return DNSError, err
		}

		routingOK = slices.Contains(addrs, c.customApexIP)
	} else {
		cname, err := firstAnswer(lookupCtx, routeServers, func(ctx context.Context, resolver *net.Resolver) (string, error) {
			return resolver.LookupCNAME(ctx, host)
		})

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

func firstAnswer[T any](ctx context.Context, servers []string, lookup func(context.Context, *net.Resolver) (T, error)) (T, error) {
	var (
		zero    T
		lastErr error = errors.New("no authoritative nameservers to query")
	)

	for _, server := range servers {
		attemptCtx, cancel := context.WithTimeout(ctx, dnsAttemptTimeout)
		answer, err := lookup(attemptCtx, resolverFor(server))

		cancel()

		if err == nil || isNotFound(err) {
			return answer, err
		}

		lastErr = err
	}

	return zero, lastErr
}

func resolverFor(server string) *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			var dialer net.Dialer

			return dialer.DialContext(ctx, network, server)
		},
	}
}

func authoritativeServers(ctx context.Context, host string) ([]string, error) {
	sys := net.Resolver{}

	for name := strings.TrimSuffix(host, "."); name != ""; name = parentDomain(name) {
		nameservers, err := sys.LookupNS(ctx, name)

		if err != nil {
			continue
		}

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

	return nil, fmt.Errorf("no authoritative nameservers found for %q", host)
}

func parentDomain(name string) string {
	i := strings.IndexByte(name, '.')

	if i < 0 {
		return ""
	}

	return name[i+1:]
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
