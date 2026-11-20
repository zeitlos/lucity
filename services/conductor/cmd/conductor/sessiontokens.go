package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/pkg/oidc"
)

const (
	sessionTTL    = 14 * 24 * time.Hour
	tokenCacheTTL = 5 * time.Minute
	accountKey    = "@account"
)

var (
	errNoSession           = errors.New("session not found")
	errWorkspaceRoleNotSet = errors.New("workspace role not yet propagated")

	roleScopes         = []string{"admin", "member", "deployer"}
	orgTokenScopes     = []string{"admin", "member"}
	accountTokenScopes = []string{"openid", "profile", "email", "identities", "urn:logto:scope:organizations", "urn:logto:scope:organization_roles"}
	loginScopes        = []string{"openid", "profile", "email", "offline_access", "identities", "urn:logto:scope:organizations", "urn:logto:scope:organization_roles", "admin", "member"}
)

type sessionData struct {
	SID          string `json:"sid"`
	Sub          string `json:"sub"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Picture      string `json:"picture"`
	RefreshToken string `json:"rt"`
}

type cachedToken struct {
	token string
	exp   time.Time
}

type cachedSession struct {
	mu      sync.Mutex
	tokens  map[string]cachedToken
	checked time.Time
	used    time.Time
}

type sessionTokens struct {
	provider *oidc.Provider
	logto    *logto.Client

	mu       sync.Mutex
	sessions map[string]*cachedSession
	swept    time.Time
}

func newSessionTokens(provider *oidc.Provider, logtoClient *logto.Client) *sessionTokens {
	return &sessionTokens{provider: provider, logto: logtoClient, sessions: map[string]*cachedSession{}}
}

func (s *sessionTokens) session(sid string) *cachedSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if now.Sub(s.swept) > tokenCacheTTL {
		for id, cached := range s.sessions {
			if now.Sub(cached.used) > tokenCacheTTL {
				delete(s.sessions, id)
			}
		}
		s.swept = now
	}

	cached, ok := s.sessions[sid]
	if !ok {
		cached = &cachedSession{tokens: map[string]cachedToken{}}
		s.sessions[sid] = cached
	}
	cached.used = now
	return cached
}

func (s *sessionTokens) forget(sid string) {
	s.mu.Lock()
	delete(s.sessions, sid)
	s.mu.Unlock()
}

// check confirms with the provider, at most every tokenCacheTTL, that the
// session's refresh token is still valid.
func (s *sessionTokens) check(ctx context.Context, data sessionData) error {
	cached := s.session(data.SID)
	cached.mu.Lock()
	recent := time.Since(cached.checked) < tokenCacheTTL
	cached.mu.Unlock()
	if recent {
		return nil
	}
	_, err := s.accountToken(ctx, data)
	return err
}

func (s *sessionTokens) workspaceToken(ctx context.Context, data sessionData, workspace string) (string, error) {
	return s.token(ctx, data, workspace, func(ctx context.Context) (*oidc.Tokens, error) {
		org, err := s.logto.OrganizationByName(ctx, workspace)
		if err != nil {
			return nil, err
		}
		tokens, err := s.provider.Refresh(ctx, data.RefreshToken, s.provider.Audience, org.ID, orgTokenScopes)
		if err != nil {
			return nil, err
		}
		if !accessTokenHasRole(tokens.AccessToken) {
			return nil, errWorkspaceRoleNotSet
		}
		return tokens, nil
	})
}

func (s *sessionTokens) accountToken(ctx context.Context, data sessionData) (string, error) {
	return s.token(ctx, data, accountKey, func(ctx context.Context) (*oidc.Tokens, error) {
		return s.provider.Refresh(ctx, data.RefreshToken, "", "", accountTokenScopes)
	})
}

func (s *sessionTokens) token(ctx context.Context, data sessionData, key string, mint func(context.Context) (*oidc.Tokens, error)) (string, error) {
	cached := s.session(data.SID)
	cached.mu.Lock()
	defer cached.mu.Unlock()

	if t, ok := cached.tokens[key]; ok && time.Now().Before(t.exp) {
		return t.token, nil
	}

	tokens, err := mint(ctx)
	if errors.Is(err, oidc.ErrInvalidGrant) {
		return "", fmt.Errorf("%w: %w", errNoSession, err)
	}
	if err != nil {
		return "", err
	}

	now := time.Now()
	exp := now.Add(time.Duration(tokens.ExpiresIn)*time.Second - time.Minute)
	if limit := now.Add(tokenCacheTTL); limit.Before(exp) {
		exp = limit
	}
	cached.tokens[key] = cachedToken{token: tokens.AccessToken, exp: exp}
	cached.checked = now
	return tokens.AccessToken, nil
}

func accessTokenHasRole(accessToken string) bool {
	parts := strings.Split(accessToken, ".")
	if len(parts) < 2 {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	var claims struct {
		Scope string `json:"scope"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return false
	}
	return hasRoleScope(claims.Scope)
}

func hasRoleScope(scope string) bool {
	for _, granted := range strings.Fields(scope) {
		for _, role := range roleScopes {
			if granted == role {
				return true
			}
		}
	}
	return false
}

func generateSessionID() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
