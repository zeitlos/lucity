package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/pkg/oidc"
)

var (
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

type userSession struct {
	mu           sync.Mutex
	refreshToken string
	tokens       map[string]cachedToken
}

type sessionStore struct {
	mu       sync.Mutex
	sessions map[string]*userSession

	provider *oidc.Provider
	logto    *logto.Client
}

func newSessionStore(provider *oidc.Provider, logtoClient *logto.Client) *sessionStore {
	return &sessionStore{
		sessions: map[string]*userSession{},
		provider: provider,
		logto:    logtoClient,
	}
}

func (s *sessionStore) get(data *sessionData) *userSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[data.SID]
	if !ok {
		sess = &userSession{refreshToken: data.RefreshToken, tokens: map[string]cachedToken{}}
		s.sessions[data.SID] = sess
	}
	return sess
}

func (s *sessionStore) drop(sid string) {
	s.mu.Lock()
	delete(s.sessions, sid)
	s.mu.Unlock()
}

func (s *sessionStore) orgToken(ctx context.Context, sess *userSession, workspace string) (token, rotated string, err error) {
	orgID, err := s.orgID(ctx, workspace)
	if err != nil {
		return "", "", err
	}
	return s.token(ctx, sess, workspace, s.provider.Audience, orgID, orgTokenScopes)
}

func (s *sessionStore) accountAPIToken(ctx context.Context, sess *userSession) (token, rotated string, err error) {
	return s.token(ctx, sess, "@acctapi", "", "", accountTokenScopes)
}

func (s *sessionStore) token(ctx context.Context, sess *userSession, cacheKey, resource, orgID string, scopes []string) (token, rotated string, err error) {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	if t, ok := sess.tokens[cacheKey]; ok && time.Now().Before(t.exp) {
		return t.token, "", nil
	}

	tokens, err := s.provider.Refresh(ctx, sess.refreshToken, resource, orgID, scopes)
	if err != nil {
		return "", "", err
	}
	if tokens.RefreshToken != "" && tokens.RefreshToken != sess.refreshToken {
		sess.refreshToken = tokens.RefreshToken
		rotated = tokens.RefreshToken
	}
	leeway := tokens.ExpiresIn - 60
	if leeway < 0 {
		leeway = tokens.ExpiresIn
	}
	sess.tokens[cacheKey] = cachedToken{token: tokens.AccessToken, exp: time.Now().Add(time.Duration(leeway) * time.Second)}
	return tokens.AccessToken, rotated, nil
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
