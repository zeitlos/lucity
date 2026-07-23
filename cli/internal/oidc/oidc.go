// Package oidc is a minimal OpenID Connect client for the identity provider,
// enough for the CLI to run a native Authorization-Code + PKCE login and to
// trade the resulting refresh token for account, organization, and Account-API
// access tokens directly against the IdP.
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

// Scopes requested at sign-in. offline_access yields a refresh token; the
// organization scopes let the refresh token mint per-workspace org tokens; the
// resource scopes (admin/member/deployer) are granted per the caller's org role.
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
	"deployer",
}

// AccountScopes are requested for the Account-API token (no resource): enough to
// read the profile, organization memberships, and linked social identities.
var AccountScopes = []string{
	"openid",
	"profile",
	"email",
	"identities",
	"urn:logto:scope:organizations",
	"urn:logto:scope:organization_roles",
}

// ResourceScopes are requested for API resource access tokens; the IdP returns
// the subset granted to the caller's organization role.
var ResourceScopes = []string{"admin", "member", "deployer"}

// directSignIn skips the connector picker and goes straight to GitHub, the only
// configured sign-in method.
const directSignIn = "social:github"

// Provider addresses a single IdP tenant. Endpoint is the issuer without its
// /oidc suffix (as advertised by the conductor's /auth/config endpoint).
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

// AuthCodeURL builds the browser authorization URL for the PKCE login.
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
	}
	return p.authEndpoint() + "?" + query.Encode()
}

// Exchange trades an authorization code for tokens. It requests no resource, so
// the returned access token targets the IdP itself (Account API / userinfo); the
// caller keeps the refresh token to mint resource tokens later.
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

// Refresh mints an access token from a refresh token. Passing a resource and
// organization ID yields an organization access token; passing neither yields
// an Account-API token. The IdP rotates the refresh token, so callers must
// persist Tokens.RefreshToken when it changes.
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

// UserInfo reads the profile, organization memberships, and roles from the
// IdP userinfo endpoint using an Account-API access token.
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
