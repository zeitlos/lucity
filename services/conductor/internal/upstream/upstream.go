// Package upstream checks the model providers and connectors a workspace
// connects for its agents, by calling them with the credential before it is
// stored. Every address it dials must be public, so a URL can't point the
// conductor at the cluster or the cloud's metadata service.
package upstream

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"syscall"
	"time"
)

type Status string

const (
	Valid       Status = "valid"
	Rejected    Status = "rejected"
	Unreachable Status = "unreachable"
)

// Check is the outcome of calling an upstream with a credential. Reason says
// why it isn't valid, in words fit for the user.
type Check struct {
	Status Status
	Reason string
}

const maxBody = 4 << 20

var errNotPublic = errors.New("address is not public")

var blockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("2002::/16"),
}

type Client struct {
	http *http.Client
}

func New() *Client {
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
		Control: func(_, address string, _ syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)

			if err != nil {
				return err
			}

			addr, err := netip.ParseAddr(host)

			if err != nil || !isPublic(addr) {
				return errNotPublic
			}

			return nil
		},
	}

	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		MaxIdleConns:          10,
		IdleConnTimeout:       30 * time.Second,
	}

	return &Client{
		http: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
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

func (c *Client) do(req *http.Request) (*http.Response, []byte, error) {
	resp, err := c.http.Do(req)

	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))

	if err != nil {
		return nil, nil, err
	}

	if len(body) > maxBody {
		return nil, nil, fmt.Errorf("response exceeds %d bytes", maxBody)
	}

	return resp, body, nil
}

func unreachable(err error) Check {
	reason := "it can't be reached"

	var timeout interface{ Timeout() bool }

	switch {
	case errors.Is(err, errNotPublic):
		reason = "its address is not public"
	case errors.As(err, &timeout) && timeout.Timeout():
		reason = "it didn't answer in time"
	case errors.Is(err, context.Canceled):
		reason = "the check was canceled"
	}

	return Check{Status: Unreachable, Reason: reason}
}

func statusCheck(code int) Check {
	switch {
	case code == http.StatusUnauthorized || code == http.StatusForbidden:
		return Check{Status: Rejected, Reason: fmt.Sprintf("it turned the credential down (HTTP %d)", code)}
	case code >= 300 && code < 400:
		return Check{Status: Unreachable, Reason: fmt.Sprintf("it redirects elsewhere (HTTP %d)", code)}
	default:
		return Check{Status: Unreachable, Reason: fmt.Sprintf("it answered with HTTP %d", code)}
	}
}
