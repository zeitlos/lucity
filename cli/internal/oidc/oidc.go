// Package oidc is a minimal OpenID Connect client for the CLI's native login.
package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var LoginScopes = []string{
	"openid",
	"profile",
	"email",
	"offline_access",
	"identities",
	"urn:logto:scope:organizations",
	"urn:logto:scope:organization_roles",
	"admin",
	"member",
}

var AccountScopes = []string{
	"openid",
	"profile",
	"email",
	"identities",
	"urn:logto:scope:organizations",
	"urn:logto:scope:organization_roles",
}

var ResourceScopes = []string{"admin", "member"}

const directSignIn = "social:github"

type Provider struct {
	Endpoint string
	ClientID string
	Audience string
	HTTP     *http.Client
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserInfo struct {
	Subject           string         `json:"sub"`
	Name              string         `json:"name"`
	Email             string         `json:"email"`
	Picture           string         `json:"picture"`
	OrganizationData  []Organization `json:"organization_data"`
	OrganizationRoles []string       `json:"organization_roles"`
}

func (p *Provider) httpClient() *http.Client {
	if p.HTTP != nil {
		return p.HTTP
	}
	return http.DefaultClient
}

func (p *Provider) authEndpoint() string  { return strings.TrimRight(p.Endpoint, "/") + "/oidc/auth" }
func (p *Provider) tokenEndpoint() string { return strings.TrimRight(p.Endpoint, "/") + "/oidc/token" }
func (p *Provider) userInfoEndpoint() string {
	return strings.TrimRight(p.Endpoint, "/") + "/oidc/me"
}

func (p *Provider) AuthCodeURL(redirectURI, state, challenge string) string {
	query := url.Values{
		"client_id":             {p.ClientID},
		"redirect_uri":          {redirectURI},
		"response_type":         {"code"},
		"scope":                 {strings.Join(LoginScopes, " ")},
		"state":                 {state},
		"resource":              {p.Audience},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
		"direct_sign_in":        {directSignIn},
		"prompt":                {"consent"},
	}
	return p.authEndpoint() + "?" + query.Encode()
}

func (p *Provider) Exchange(ctx context.Context, code, redirectURI, verifier string) (*Tokens, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {p.ClientID},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"code_verifier": {verifier},
	}
	return p.token(ctx, form)
}

func (p *Provider) Refresh(ctx context.Context, refreshToken, resource, organizationID string, scopes []string) (*Tokens, error) {
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {p.ClientID},
		"refresh_token": {refreshToken},
	}
	if resource != "" {
		form.Set("resource", resource)
	}
	if organizationID != "" {
		form.Set("organization_id", organizationID)
	}
	if len(scopes) > 0 {
		form.Set("scope", strings.Join(scopes, " "))
	}
	return p.token(ctx, form)
}

func (p *Provider) token(ctx context.Context, form url.Values) (*Tokens, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tokenEndpoint(), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request returned %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}

	var tokens Tokens
	if err := json.Unmarshal(payload, &tokens); err != nil {
		return nil, err
	}
	return &tokens, nil
}

func (p *Provider) UserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.userInfoEndpoint(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request returned %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}

	var info UserInfo
	if err := json.Unmarshal(payload, &info); err != nil {
		return nil, err
	}
	return &info, nil
}
