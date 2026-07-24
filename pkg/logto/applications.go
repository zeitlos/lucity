package logto

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ApplicationTypeM2M is the Logto application type for machine-to-machine apps.
const ApplicationTypeM2M = "MachineToMachine"

// Application is a Logto application.
type Application struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Secret      string         `json:"secret,omitempty"`
	Type        string         `json:"type"`
	Description string         `json:"description,omitempty"`
	CustomData  map[string]any `json:"customData,omitempty"`
	CreatedAt   int64          `json:"createdAt,omitempty"`
}

// ApplicationSecret is a named, optionally-expiring secret for an application.
type ApplicationSecret struct {
	ApplicationID string `json:"applicationId"`
	Name          string `json:"name"`
	Value         string `json:"value"`
	ExpiresAt     int64  `json:"expiresAt,omitempty"`
	CreatedAt     int64  `json:"createdAt,omitempty"`
}

// CreateApplication creates a machine-to-machine application.
func (c *Client) CreateApplication(ctx context.Context, name, description string, customData map[string]any) (*Application, error) {
	payload := map[string]any{"name": name, "type": ApplicationTypeM2M}
	if description != "" {
		payload["description"] = description
	}
	if customData != nil {
		payload["customData"] = customData
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var app Application
	if err := c.doJSON(ctx, http.MethodPost, "/api/applications", bytes.NewReader(body), &app); err != nil {
		return nil, err
	}
	return &app, nil
}

// DeleteApplication removes an application. A missing application is treated as success.
func (c *Client) DeleteApplication(ctx context.Context, applicationID string) error {
	resp, err := c.doManagement(ctx, http.MethodDelete, "/api/applications/"+applicationID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("logto DELETE application %s returned %d: %s", applicationID, resp.StatusCode, string(b))
	}
	return nil
}

// CreateApplicationSecret mints a new secret for an application. A zero expiresAt
// creates a non-expiring secret.
func (c *Client) CreateApplicationSecret(ctx context.Context, applicationID, name string, expiresAt time.Time) (*ApplicationSecret, error) {
	payload := map[string]any{"name": name}
	if !expiresAt.IsZero() {
		payload["expiresAt"] = expiresAt.UnixMilli()
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var secret ApplicationSecret
	if err := c.doJSON(ctx, http.MethodPost, "/api/applications/"+applicationID+"/secrets", bytes.NewReader(body), &secret); err != nil {
		return nil, err
	}
	return &secret, nil
}

// AddApplicationToOrganization associates an application with an organization.
func (c *Client) AddApplicationToOrganization(ctx context.Context, organizationID, applicationID string) error {
	body, err := json.Marshal(map[string]any{"applicationIds": []string{applicationID}})
	if err != nil {
		return err
	}
	return c.doNoContent(ctx, http.MethodPost, "/api/organizations/"+organizationID+"/applications", bytes.NewReader(body))
}

// AssignApplicationOrganizationRoles grants organization roles to an application
// within an organization.
func (c *Client) AssignApplicationOrganizationRoles(ctx context.Context, organizationID, applicationID string, roleIDs []string) error {
	body, err := json.Marshal(map[string]any{"organizationRoleIds": roleIDs})
	if err != nil {
		return err
	}
	return c.doNoContent(ctx, http.MethodPost, "/api/organizations/"+organizationID+"/applications/"+applicationID+"/roles", bytes.NewReader(body))
}

// OrganizationApplications lists the applications associated with an organization.
func (c *Client) OrganizationApplications(ctx context.Context, organizationID string) ([]Application, error) {
	var apps []Application
	if err := c.doJSON(ctx, http.MethodGet, "/api/organizations/"+organizationID+"/applications", nil, &apps); err != nil {
		return nil, err
	}
	return apps, nil
}

// OrganizationAccessToken exchanges an application's client credentials for an
// organization-scoped access token audienced to the given resource. An empty
// scopes slice omits the scope parameter.
func (c *Client) OrganizationAccessToken(ctx context.Context, clientID, clientSecret, resource, organizationID string, scopes []string) (string, int, error) {
	data := url.Values{
		"grant_type":      {"client_credentials"},
		"resource":        {resource},
		"organization_id": {organizationID},
	}
	if len(scopes) > 0 {
		data.Set("scope", strings.Join(scopes, " "))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/oidc/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", 0, err
	}
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("organization token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("organization token request returned %d: %s", resp.StatusCode, string(b))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", 0, fmt.Errorf("failed to decode organization token response: %w", err)
	}
	return tokenResp.AccessToken, tokenResp.ExpiresIn, nil
}
