package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/zeitlos/lucity/pkg/kvstore"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/pkg/oidc"
)

const sessionTTL = 30 * 24 * time.Hour

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
	Token string    `json:"token"`
	Exp   time.Time `json:"exp"`
}

type sessionValue struct {
	RefreshToken string                 `json:"refreshToken"`
	Tokens       map[string]cachedToken `json:"tokens"`
}

type sessionStore struct {
	mint     sync.Mutex
	store    kvstore.Store[sessionValue]
	provider *oidc.Provider
	logto    *logto.Client
}

func newSessionStore(store kvstore.Store[sessionValue], provider *oidc.Provider, logtoClient *logto.Client) *sessionStore {
	return &sessionStore{store: store, provider: provider, logto: logtoClient}
}

func (s *sessionStore) create(ctx context.Context, refreshToken string) (string, error) {
	sid := generateSessionID()
	if err := s.store.Set(ctx, sid, sessionValue{RefreshToken: refreshToken, Tokens: map[string]cachedToken{}}, sessionTTL); err != nil {
		return "", err
	}
	return sid, nil
}

func (s *sessionStore) ensure(ctx context.Context, sid, refreshToken string) {
	if _, ok, _ := s.store.Get(ctx, sid); ok {
		return
	}
	s.mint.Lock()
	defer s.mint.Unlock()
	if _, ok, _ := s.store.Get(ctx, sid); ok {
		return
	}
	_ = s.store.Set(ctx, sid, sessionValue{RefreshToken: refreshToken, Tokens: map[string]cachedToken{}}, sessionTTL)
}

func (s *sessionStore) delete(ctx context.Context, sid string) {
	_ = s.store.Delete(ctx, sid)
}

func (s *sessionStore) refreshToken(ctx context.Context, sid string) (string, bool) {
	val, ok, _ := s.store.Get(ctx, sid)
	if !ok {
		return "", false
	}
	return val.RefreshToken, true
}

func (s *sessionStore) orgToken(ctx context.Context, sid, workspace string) (string, error) {
	orgID, err := s.orgID(ctx, workspace)
	if err != nil {
		return "", err
	}
	return s.token(ctx, sid, workspace, s.provider.Audience, orgID, orgTokenScopes, true)
}

func (s *sessionStore) accountAPIToken(ctx context.Context, sid string) (string, error) {
	return s.token(ctx, sid, "@acctapi", "", "", accountTokenScopes, false)
}

func (s *sessionStore) token(ctx context.Context, sid, cacheKey, resource, orgID string, scopes []string, requireRole bool) (string, error) {
	if val, ok, _ := s.store.Get(ctx, sid); ok {
		if t, ok := val.Tokens[cacheKey]; ok && time.Now().Before(t.Exp) {
			return t.Token, nil
		}
	}

	s.mint.Lock()
	defer s.mint.Unlock()

	val, ok, err := s.store.Get(ctx, sid)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errNoSession
	}
	if t, ok := val.Tokens[cacheKey]; ok && time.Now().Before(t.Exp) {
		return t.Token, nil
	}

	tokens, err := s.provider.Refresh(ctx, val.RefreshToken, resource, orgID, scopes)
	if err != nil {
		return "", err
	}

	// Persist the rotated refresh token even when the grant is otherwise
	// unusable — Logto invalidates the previous one on every exchange.
	if tokens.RefreshToken != "" {
		val.RefreshToken = tokens.RefreshToken
	}

	if requireRole && !accessTokenHasRole(tokens.AccessToken) {
		_ = s.store.Set(ctx, sid, val, sessionTTL)
		return "", errWorkspaceRoleNotSet
	}

	if val.Tokens == nil {
		val.Tokens = map[string]cachedToken{}
	}
	leeway := tokens.ExpiresIn - 60
	if leeway < 0 {
		leeway = tokens.ExpiresIn
	}
	val.Tokens[cacheKey] = cachedToken{Token: tokens.AccessToken, Exp: time.Now().Add(time.Duration(leeway) * time.Second)}
	if err := s.store.Set(ctx, sid, val, sessionTTL); err != nil {
		return "", err
	}
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

func (s *sessionStore) orgID(ctx context.Context, workspace string) (string, error) {
	org, err := s.logto.OrganizationByName(ctx, workspace)
	if err != nil {
		return "", err
	}
	return org.ID, nil
}

func generateSessionID() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
