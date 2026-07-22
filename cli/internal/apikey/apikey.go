package apikey

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const prefix = "lucity_"

// Key is a decoded workspace API key.
type Key struct {
	ClientID  string
	Secret    string
	OrgID     string
	Workspace string
}

// Parse decodes a lucity_ API key into its parts.
func Parse(raw string) (*Key, error) {
	if !strings.HasPrefix(raw, prefix) {
		return nil, fmt.Errorf("invalid API key")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(raw, prefix))
	if err != nil {
		return nil, fmt.Errorf("invalid API key encoding: %w", err)
	}
	parts := strings.SplitN(string(decoded), ":", 4)
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid API key format")
	}
	return &Key{ClientID: parts[0], Secret: parts[1], OrgID: parts[2], Workspace: parts[3]}, nil
}

// Exchange trades the key's client credentials for a short-lived organization
// access token audienced to the given resource.
func (k *Key) Exchange(ctx context.Context, httpClient *http.Client, issuer, audience string) (token string, expiresIn int, err error) {
	form := url.Values{
		"grant_type":      {"client_credentials"},
		"resource":        {audience},
		"organization_id": {k.OrgID},
		"scope":           {"admin member deployer"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(issuer, "/")+"/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", 0, err
	}
	req.SetBasicAuth(k.ClientID, k.Secret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("API key exchange failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("API key exchange returned %d: %s", resp.StatusCode, string(b))
	}

	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", 0, err
	}
	return out.AccessToken, out.ExpiresIn, nil
}
