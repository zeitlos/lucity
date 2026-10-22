package conductor

import (
	"context"
	"log/slog"

	"github.com/zeitlos/lucity/pkg/hostname"
	"github.com/zeitlos/lucity/pkg/platform"
)

type Endpoint struct {
	Host               string
	Port               int
	Protocol           platform.Protocol
	RedirectTo         string
	RequiredDNSRecords []hostname.DNSRecord
	DNSStatus          hostname.DNSStatus
	TLSStatus          hostname.TLSStatus
	Type               EndpointType
}

type EndpointType = string

const (
	InternalEndpoint     EndpointType = "internal"
	PlatformEndpoint     EndpointType = "platform"
	CustomDomainEndpoint EndpointType = "custom"
)

func (c *Client) ResolveEndpoints(ctx context.Context, serviceID ServiceID, endpoints []platform.Endpoint) ([]Endpoint, error) {
	result := make([]Endpoint, 0, len(endpoints))

	for _, endpoint := range endpoints {
		var err error
		resolved := Endpoint{
			Host:               endpoint.Host,
			Port:               endpoint.Port,
			Protocol:           endpoint.Protocol,
			RedirectTo:         endpoint.RedirectTo,
			RequiredDNSRecords: c.hostname.DNSRecords(serviceID.Workspace, endpoint.Host, endpoint.RedirectTo != ""),
		}

		resolved.DNSStatus, err = c.hostname.DNSStatus(ctx, serviceID.Workspace, endpoint.Host, endpoint.RedirectTo != "")

		if err != nil {
			slog.ErrorContext(ctx, "failed to lookup dns status", "error", err, "service", serviceID, "host", endpoint.Host)
		}

		resolved.TLSStatus, err = c.hostname.TLSStatus(ctx, serviceID.Namespace(), endpoint.Host)

		if err != nil {
			slog.ErrorContext(ctx, "failed to lookup tls status", "error", err, "service", serviceID, "host", endpoint.Host)
		}

		if c.hostname.IsInternal(endpoint.Host) {
			resolved.Type = InternalEndpoint
		} else if c.hostname.IsPlatform(endpoint.Host) {
			resolved.Type = PlatformEndpoint
		} else {
			resolved.Type = CustomDomainEndpoint
		}

		result = append(result, resolved)
	}

	return result, nil
}
