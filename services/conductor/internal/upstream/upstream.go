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
	"net/http"
	"time"

	"github.com/zeitlos/lucity/services/conductor/internal/publicnet"
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

type Client struct {
	http *http.Client
}

func New() *Client {
	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           publicnet.Dialer(5 * time.Second).DialContext,
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
	case errors.Is(err, publicnet.ErrNotPublic):
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
