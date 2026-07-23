package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/zeitlos/lucity/pkg/logto"
)

var (
	orgTokenScopes     = []string{"admin", "member"}
	accountTokenScopes = []string{"openid", "profile", "email", "identities", "urn:logto:scope:organizations", "urn:logto:scope:organization_roles"}
	sessionHTTPClient  = &http.Client{Timeout: 30 * time.Second}
)

type sessionData struct {
	SID          string `json:"sid"`
	Sub          string `json:"sub"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Picture      string `json:"picture"`
	RefreshToken string `json:"rt"`
}

func sessionSealKey(secret string) [32]byte {
	return sha256.Sum256([]byte("lucity-session:" + secret))
}

func sealSession(secret string, data sessionData) (string, error) {
	plaintext, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	key := sessionSealKey(secret)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(gcm.Seal(nonce, nonce, plaintext, nil)), nil
}

func openSession(secret, value string) (*sessionData, error) {
	sealed, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, errors.New("invalid session")
	}
	key := sessionSealKey(secret)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, errors.New("invalid session")
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.New("invalid session")
	}
	if len(sealed) < gcm.NonceSize() {
		return nil, errors.New("invalid session")
	}
	plaintext, err := gcm.Open(nil, sealed[:gcm.NonceSize()], sealed[gcm.NonceSize():], nil)
	if err != nil {
		return nil, errors.New("invalid session")
	}
	var data sessionData
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return nil, errors.New("invalid session")
	}
	return &data, nil
}

type cachedToken struct {
	token string
	exp   time.Time
}

type session struct {
	mu           sync.Mutex
	refreshToken string
	tokens       map[string]cachedToken
}

type sessionStore struct {
	mu       sync.Mutex
	sessions map[string]*session

	provider *OIDCProvider
	logto    *logto.Client
	audience string
	http     *http.Client
}

func newSessionStore(provider *OIDCProvider, logtoClient *logto.Client, audience string) *sessionStore {
	client := provider.httpClient
	if client == nil {
		client = sessionHTTPClient
	}
	return &sessionStore{
		sessions: map[string]*session{},
		provider: provider,
		logto:    logtoClient,
		audience: audience,
		http:     client,
	}
}

func (s *sessionStore) get(data *sessionData) *session {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[data.SID]
	if !ok {
		sess = &session{refreshToken: data.RefreshToken, tokens: map[string]cachedToken{}}
		s.sessions[data.SID] = sess
	}
	return sess
}

func (s *sessionStore) drop(sid string) {
	s.mu.Lock()
	delete(s.sessions, sid)
	s.mu.Unlock()
}

func (s *sessionStore) orgToken(ctx context.Context, sess *session, workspace string) (token, rotated string, err error) {
	orgID, err := s.orgID(ctx, workspace)
	if err != nil {
		return "", "", err
	}
	return s.token(ctx, sess, workspace, s.audience, orgID, orgTokenScopes)
}

func (s *sessionStore) accountToken(ctx context.Context, sess *session) (token, rotated string, err error) {
	return s.token(ctx, sess, "@account", s.audience, "", orgTokenScopes)
}

func (s *sessionStore) accountAPIToken(ctx context.Context, sess *session) (token, rotated string, err error) {
	return s.token(ctx, sess, "@acctapi", "", "", accountTokenScopes)
}

func (s *sessionStore) token(ctx context.Context, sess *session, cacheKey, resource, orgID string, scopes []string) (token, rotated string, err error) {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	if t, ok := sess.tokens[cacheKey]; ok && time.Now().Before(t.exp) {
		return t.token, "", nil
	}

	resp, err := s.refresh(ctx, sess.refreshToken, resource, orgID, scopes)
	if err != nil {
		return "", "", err
	}
	if resp.RefreshToken != "" && resp.RefreshToken != sess.refreshToken {
		sess.refreshToken = resp.RefreshToken
		rotated = resp.RefreshToken
	}
	leeway := resp.ExpiresIn - 60
	if leeway < 0 {
		leeway = resp.ExpiresIn
	}
	sess.tokens[cacheKey] = cachedToken{token: resp.AccessToken, exp: time.Now().Add(time.Duration(leeway) * time.Second)}
	return resp.AccessToken, rotated, nil
}

func (s *sessionStore) orgID(ctx context.Context, workspace string) (string, error) {
	org, err := s.logto.OrganizationByName(ctx, workspace)
	if err != nil {
		return "", err
	}
	return org.ID, nil
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (s *sessionStore) refresh(ctx context.Context, refreshToken, resource, orgID string, scopes []string) (*tokenResponse, error) {
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}
	if resource != "" {
		form.Set("resource", resource)
	}
	if orgID != "" {
		form.Set("organization_id", orgID)
	}
	if len(scopes) > 0 {
		form.Set("scope", strings.Join(scopes, " "))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.provider.oauthConfig.Endpoint.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.provider.oauthConfig.ClientID, s.provider.oauthConfig.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh returned %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}

	var out tokenResponse
	if err := json.Unmarshal(payload, &out); err != nil {
		return nil, err
	}
	if out.AccessToken == "" {
		return nil, errors.New("token refresh returned an empty access token")
	}
	return &out, nil
}

func generateSessionID() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
