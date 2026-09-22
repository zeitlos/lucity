package bunny

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const DNSRecordTypeCNAME = 2

type DNSZone struct {
	ID      int64       `json:"Id"`
	Domain  string      `json:"Domain"`
	Records []DNSRecord `json:"Records"`
}

type DNSRecord struct {
	ID    int64  `json:"Id,omitempty"`
	Type  int    `json:"Type"`
	Name  string `json:"Name"`
	Value string `json:"Value"`
	TTL   int    `json:"Ttl,omitempty"`
}

func (c *Client) DNSZoneByDomain(ctx context.Context, domain string) (*DNSZone, error) {
	var result struct {
		Items []DNSZone `json:"Items"`
	}

	if err := c.do(ctx, http.MethodGet, "/dnszone?search="+url.QueryEscape(domain), nil, &result); err != nil {
		return nil, fmt.Errorf("search dns zones: %w", err)
	}

	for _, zone := range result.Items {
		if zone.Domain == domain {
			return &zone, nil
		}
	}

	return nil, nil
}

func (c *Client) DNSZone(ctx context.Context, zoneID int64) (*DNSZone, error) {
	var zone DNSZone

	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/dnszone/%d", zoneID), nil, &zone); err != nil {
		return nil, fmt.Errorf("get dns zone: %w", err)
	}

	return &zone, nil
}

func (c *Client) CreateDNSRecord(ctx context.Context, zoneID int64, record DNSRecord) error {
	if err := c.do(ctx, http.MethodPut, fmt.Sprintf("/dnszone/%d/records", zoneID), record, nil); err != nil {
		return fmt.Errorf("create dns record: %w", err)
	}

	return nil
}

func (c *Client) DeleteDNSRecord(ctx context.Context, zoneID, recordID int64) error {
	if err := c.do(ctx, http.MethodDelete, fmt.Sprintf("/dnszone/%d/records/%d", zoneID, recordID), nil, nil); err != nil {
		return fmt.Errorf("delete dns record: %w", err)
	}

	return nil
}
