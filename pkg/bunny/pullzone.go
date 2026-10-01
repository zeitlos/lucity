package bunny

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
)

const (
	ActionForceSSL         = 0
	ActionSetRequestHeader = 6
	TriggerURL             = 0
	MatchAny               = 0
)

type PullZone struct {
	ID        int64      `json:"Id"`
	Name      string     `json:"Name"`
	OriginURL string     `json:"OriginUrl"`
	Hostnames []Hostname `json:"Hostnames"`
	EdgeRules []EdgeRule `json:"EdgeRules"`
}

type Hostname struct {
	Value            string `json:"Value"`
	ForceSSL         bool   `json:"ForceSSL"`
	IsSystemHostname bool   `json:"IsSystemHostname"`
	HasCertificate   bool   `json:"HasCertificate"`
}

type EdgeRule struct {
	GUID                string            `json:"Guid,omitempty"`
	ActionType          int               `json:"ActionType"`
	ActionParameter1    string            `json:"ActionParameter1"`
	ActionParameter2    string            `json:"ActionParameter2"`
	Triggers            []EdgeRuleTrigger `json:"Triggers"`
	TriggerMatchingType int               `json:"TriggerMatchingType"`
	Description         string            `json:"Description"`
	Enabled             bool              `json:"Enabled"`
}

type EdgeRuleTrigger struct {
	Type                int      `json:"Type"`
	PatternMatches      []string `json:"PatternMatches"`
	PatternMatchingType int      `json:"PatternMatchingType"`
}

type PullZoneSpec struct {
	Name                 string `json:"Name"`
	OriginURL            string `json:"OriginUrl"`
	OriginType           int    `json:"OriginType"`
	AWSSigningEnabled    bool   `json:"AWSSigningEnabled,omitempty"`
	AWSSigningKey        string `json:"AWSSigningKey,omitempty"`
	AWSSigningSecret     string `json:"AWSSigningSecret,omitempty"`
	AWSSigningRegionName string `json:"AWSSigningRegionName,omitempty"`

	AddHostHeader                   *bool  `json:"AddHostHeader,omitempty"`
	VerifyOriginSSL                 *bool  `json:"VerifyOriginSSL,omitempty"`
	EnableWebSockets                *bool  `json:"EnableWebSockets,omitempty"`
	MaxWebSocketConnections         *int   `json:"MaxWebSocketConnections,omitempty"`
	MonthlyBandwidthLimit           *int64 `json:"MonthlyBandwidthLimit,omitempty"`
	EnableSmartCache                *bool  `json:"EnableSmartCache,omitempty"`
	CacheControlMaxAgeOverride      *int64 `json:"CacheControlMaxAgeOverride,omitempty"`
	EnableHostnameVary              *bool  `json:"EnableHostnameVary,omitempty"`
	DisableCookies                  *bool  `json:"DisableCookies,omitempty"`
	EnableAccessControlOriginHeader *bool  `json:"EnableAccessControlOriginHeader,omitempty"`
	EnableTLS1                      *bool  `json:"EnableTLS1,omitempty"`
	EnableTLS1_1                    *bool  `json:"EnableTLS1_1,omitempty"`
}

func (c *Client) PullZones(ctx context.Context) ([]PullZone, error) {
	var zones []PullZone

	for page := 1; ; page++ {
		var result struct {
			Items        []PullZone `json:"Items"`
			HasMoreItems bool       `json:"HasMoreItems"`
		}

		if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/pullzone?page=%d&perPage=100", page), nil, &result); err != nil {
			return nil, fmt.Errorf("list pull zones: %w", err)
		}

		zones = append(zones, result.Items...)

		if !result.HasMoreItems {
			return zones, nil
		}
	}
}

func (c *Client) PullZoneByName(ctx context.Context, name string) (*PullZone, error) {
	var result struct {
		Items []PullZone `json:"Items"`
	}

	if err := c.do(ctx, http.MethodGet, "/pullzone?perPage=100&search="+url.QueryEscape(name), nil, &result); err != nil {
		return nil, fmt.Errorf("search pull zones: %w", err)
	}

	for _, zone := range result.Items {
		if zone.Name == name {
			return &zone, nil
		}
	}

	return nil, nil
}

func (c *Client) CreatePullZone(ctx context.Context, spec PullZoneSpec) (*PullZone, error) {
	var zone PullZone

	if err := c.do(ctx, http.MethodPost, "/pullzone", spec, &zone); err != nil {
		return nil, fmt.Errorf("create pull zone: %w", err)
	}

	return &zone, nil
}

func (c *Client) DeletePullZone(ctx context.Context, pullZoneID int64) error {
	if err := c.do(ctx, http.MethodDelete, fmt.Sprintf("/pullzone/%d", pullZoneID), nil, nil); err != nil {
		return fmt.Errorf("delete pull zone: %w", err)
	}

	return nil
}

func (c *Client) AddHostname(ctx context.Context, pullZoneID int64, hostname string) error {
	if err := c.do(ctx, http.MethodPost, fmt.Sprintf("/pullzone/%d/addHostname", pullZoneID), map[string]string{"Hostname": hostname}, nil); err != nil {
		return fmt.Errorf("add hostname: %w", err)
	}

	return nil
}

func (c *Client) RemoveHostname(ctx context.Context, pullZoneID int64, hostname string) error {
	if err := c.do(ctx, http.MethodDelete, fmt.Sprintf("/pullzone/%d/removeHostname", pullZoneID), map[string]string{"Hostname": hostname}, nil); err != nil {
		return fmt.Errorf("remove hostname: %w", err)
	}

	return nil
}

func (c *Client) UploadCertificate(ctx context.Context, pullZoneID int64, hostname string, certificate, key []byte) error {
	body := map[string]string{
		"Hostname":       hostname,
		"Certificate":    base64.StdEncoding.EncodeToString(certificate),
		"CertificateKey": base64.StdEncoding.EncodeToString(key),
	}

	if err := c.do(ctx, http.MethodPost, fmt.Sprintf("/pullzone/%d/addCertificate", pullZoneID), body, nil); err != nil {
		return fmt.Errorf("upload certificate: %w", err)
	}

	return nil
}

func (c *Client) IssueCertificate(ctx context.Context, hostname string) error {
	if err := c.do(ctx, http.MethodGet, "/pullzone/loadFreeCertificate?hostname="+url.QueryEscape(hostname), nil, nil); err != nil {
		return fmt.Errorf("issue certificate: %w", err)
	}

	return nil
}

func (c *Client) AddEdgeRule(ctx context.Context, pullZoneID int64, rule EdgeRule) error {
	if err := c.do(ctx, http.MethodPost, fmt.Sprintf("/pullzone/%d/edgerules/addOrUpdate", pullZoneID), rule, nil); err != nil {
		return fmt.Errorf("add edge rule: %w", err)
	}

	return nil
}
