package session

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/zeitlos/lucity/cli/internal/api"
	"github.com/zeitlos/lucity/cli/internal/apitoken"
	"github.com/zeitlos/lucity/cli/internal/ciauth"
	"github.com/zeitlos/lucity/cli/internal/config"
	"github.com/zeitlos/lucity/pkg/oidc"
)

var ErrLoggedOut = errors.New("not logged in — run `lucity login`")

const DirectSignIn = "social:github"

var (
	LoginScopes    = []string{"openid", "profile", "email", "offline_access", "identities", "urn:logto:scope:organizations", "urn:logto:scope:organization_roles", "admin", "member"}
	accountScopes  = []string{"openid", "profile", "email", "identities", "urn:logto:scope:organizations", "urn:logto:scope:organization_roles"}
	resourceScopes = []string{"admin", "member"}
)

type cachedToken struct {
	token string
	exp   time.Time
}

type Manager struct {
	mu  sync.Mutex
	cfg *config.Config

	cfgMu      sync.Mutex
	authConfig *api.AuthConfig

	ciMu     sync.Mutex
	ciTried  bool
	ciResult *ciauth.Session
	ciErr    error

	apiTokenMu     sync.Mutex
	apiTokenAccess string
	apiTokenExp    time.Time

	refreshMu sync.Mutex

	tokMu       sync.Mutex
	orgTokens   map[string]cachedToken
	accountTok  cachedToken
	acctAPITok  cachedToken
	orgIDByName map[string]string
}

func Load() (*Manager, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return &Manager{
		cfg:         cfg,
		orgTokens:   map[string]cachedToken{},
		orgIDByName: map[string]string{},
	}, nil
}

func (m *Manager) APIURL() string {
	if url := os.Getenv("LUCITY_API_URL"); url != "" {
		return url
	}
	if m.cfg.APIURL != "" {
		return m.cfg.APIURL
	}
	return api.DefaultBaseURL
}

func (m *Manager) Workspace() string {
	if workspace := os.Getenv("LUCITY_WORKSPACE"); workspace != "" {
		return workspace
	}
	if raw := m.apiToken(); raw != "" {
		if token, err := apitoken.Parse(raw); err == nil {
			return token.Workspace
		}
	}
	if m.cfg.Workspace != "" {
		return m.cfg.Workspace
	}
	if session, _ := m.ciSession(context.Background()); session != nil {
		return session.Workspace
	}
	return ""
}

func (m *Manager) SetWorkspace(workspace string) error {
	return m.updateConfig(func(cfg *config.Config) {
		cfg.Workspace = workspace
	})
}

func (m *Manager) SetLogin(apiURL, refreshToken string) error {
	return m.updateConfig(func(cfg *config.Config) {
		cfg.APIURL = apiURL
		cfg.RefreshToken = refreshToken
	})
}

func (m *Manager) Clear() error {
	err := m.updateConfig(func(cfg *config.Config) {
		cfg.RefreshToken = ""
	})

	m.tokMu.Lock()
	m.orgTokens = map[string]cachedToken{}
	m.accountTok = cachedToken{}
	m.acctAPITok = cachedToken{}
	m.orgIDByName = map[string]string{}
	m.tokMu.Unlock()
	return err
}

func (m *Manager) Client() *api.Client {
	return api.NewClient(m.APIURL(), m.Workspace(), m)
}

func (m *Manager) Prepare(ctx context.Context) error {
	if m.apiToken() != "" {
		return nil
	}
	if m.storedRefreshToken() != "" {
		return nil
	}
	_, err := m.ciSession(ctx)
	return err
}

func (m *Manager) Personal() bool {
	return m.apiToken() == "" && m.storedRefreshToken() != ""
}

func (m *Manager) Token(ctx context.Context) (string, error) {
	if raw := m.apiToken(); raw != "" {
		return m.apiTokenBearer(ctx, raw)
	}

	if m.storedRefreshToken() != "" {
		if workspace := m.Workspace(); workspace != "" {
			return m.orgBearer(ctx, workspace)
		}
		return m.accountBearer(ctx)
	}

	session, err := m.ciSession(ctx)
	if session != nil {
		return session.Token, nil
	}
	if err != nil {
		return "", fmt.Errorf("CI deploy authentication failed: %w", err)
	}
	return "", ErrLoggedOut
}

func (m *Manager) AccountToken(ctx context.Context) (string, error) {
	if m.apiToken() != "" {
		return "", nil
	}
	if m.storedRefreshToken() == "" {
		return "", nil
	}
	return m.accountAPIToken(ctx)
}

func (m *Manager) Identity(ctx context.Context) (*api.Identity, error) {
	accountToken, err := m.accountAPIToken(ctx)
	if err != nil {
		return nil, err
	}
	provider, err := m.provider(ctx)
	if err != nil {
		return nil, err
	}
	info, err := provider.UserInfo(ctx, accountToken)
	if err != nil {
		return nil, err
	}
	return identityFromUserInfo(info), nil
}

func (m *Manager) BootstrapWorkspaces(ctx context.Context) error {
	const query = `query { workspaces { id } }`
	return m.Client().GraphQL(ctx, query, nil, nil)
}

func (m *Manager) apiToken() string {
	return os.Getenv("LUCITY_API_TOKEN")
}

func (m *Manager) storedRefreshToken() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cfg.RefreshToken
}

func (m *Manager) authConfigCached(ctx context.Context) (*api.AuthConfig, error) {
	m.cfgMu.Lock()
	defer m.cfgMu.Unlock()

	if m.authConfig != nil {
		return m.authConfig, nil
	}
	cfg, err := api.Config(ctx, &http.Client{Timeout: 30 * time.Second}, m.APIURL())
	if err != nil {
		return nil, fmt.Errorf("fetch auth config: %w", err)
	}
	m.authConfig = cfg
	return cfg, nil
}

func (m *Manager) provider(ctx context.Context) (*oidc.Provider, error) {
	cfg, err := m.authConfigCached(ctx)
	if err != nil {
		return nil, err
	}
	return &oidc.Provider{
		Endpoint:     cfg.Endpoint,
		ClientID:     cfg.CliClientID,
		Audience:     cfg.Audience,
		DirectSignIn: DirectSignIn,
		Scopes:       LoginScopes,
		HTTP:         &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (m *Manager) Provider(ctx context.Context) (*oidc.Provider, error) {
	return m.provider(ctx)
}

func (m *Manager) apiTokenBearer(ctx context.Context, raw string) (string, error) {
	m.apiTokenMu.Lock()
	defer m.apiTokenMu.Unlock()

	if m.apiTokenAccess != "" && time.Now().Before(m.apiTokenExp) {
		return m.apiTokenAccess, nil
	}

	parsed, err := apitoken.Parse(raw)
	if err != nil {
		return "", err
	}

	cfg, err := m.authConfigCached(ctx)
	if err != nil {
		return "", err
	}

	token, expiresIn, err := parsed.Exchange(ctx, &http.Client{Timeout: 30 * time.Second}, cfg.Issuer, cfg.Audience)
	if err != nil {
		return "", err
	}

	m.apiTokenAccess = token
	m.apiTokenExp = tokenExpiry(expiresIn)
	return token, nil
}

func (m *Manager) orgBearer(ctx context.Context, workspace string) (string, error) {
	m.tokMu.Lock()
	if t, ok := m.orgTokens[workspace]; ok && time.Now().Before(t.exp) {
		m.tokMu.Unlock()
		return t.token, nil
	}
	m.tokMu.Unlock()

	cfg, err := m.authConfigCached(ctx)
	if err != nil {
		return "", err
	}
	orgID, err := m.orgID(ctx, workspace)
	if err != nil {
		return "", err
	}

	token, expiresIn, err := m.refreshGrant(ctx, cfg.Audience, orgID, resourceScopes)
	if err != nil {
		return "", err
	}

	m.tokMu.Lock()
	m.orgTokens[workspace] = cachedToken{token: token, exp: tokenExpiry(expiresIn)}
	m.tokMu.Unlock()
	return token, nil
}

func (m *Manager) accountBearer(ctx context.Context) (string, error) {
	m.tokMu.Lock()
	if m.accountTok.token != "" && time.Now().Before(m.accountTok.exp) {
		token := m.accountTok.token
		m.tokMu.Unlock()
		return token, nil
	}
	m.tokMu.Unlock()

	cfg, err := m.authConfigCached(ctx)
	if err != nil {
		return "", err
	}

	token, expiresIn, err := m.refreshGrant(ctx, cfg.Audience, "", resourceScopes)
	if err != nil {
		return "", err
	}

	m.tokMu.Lock()
	m.accountTok = cachedToken{token: token, exp: tokenExpiry(expiresIn)}
	m.tokMu.Unlock()
	return token, nil
}

func (m *Manager) accountAPIToken(ctx context.Context) (string, error) {
	m.tokMu.Lock()
	if m.acctAPITok.token != "" && time.Now().Before(m.acctAPITok.exp) {
		token := m.acctAPITok.token
		m.tokMu.Unlock()
		return token, nil
	}
	m.tokMu.Unlock()

	token, expiresIn, err := m.refreshGrant(ctx, "", "", accountScopes)
	if err != nil {
		return "", err
	}

	m.tokMu.Lock()
	m.acctAPITok = cachedToken{token: token, exp: tokenExpiry(expiresIn)}
	m.tokMu.Unlock()
	return token, nil
}

func (m *Manager) orgID(ctx context.Context, workspace string) (string, error) {
	m.tokMu.Lock()
	id, ok := m.orgIDByName[workspace]
	m.tokMu.Unlock()
	if ok {
		return id, nil
	}

	if err := m.loadOrgIDs(ctx); err != nil {
		return "", err
	}

	m.tokMu.Lock()
	id, ok = m.orgIDByName[workspace]
	m.tokMu.Unlock()
	if !ok {
		return "", fmt.Errorf("you are not a member of workspace %q — run `lucity account` to list memberships", workspace)
	}
	return id, nil
}

func (m *Manager) loadOrgIDs(ctx context.Context) error {
	accountToken, err := m.accountAPIToken(ctx)
	if err != nil {
		return err
	}
	provider, err := m.provider(ctx)
	if err != nil {
		return err
	}
	info, err := provider.UserInfo(ctx, accountToken)
	if err != nil {
		return err
	}

	m.tokMu.Lock()
	for _, org := range info.OrganizationData {
		m.orgIDByName[org.Name] = org.ID
	}
	m.tokMu.Unlock()
	return nil
}

func (m *Manager) updateConfig(change func(*config.Config)) error {
	unlock, err := config.Lock()
	if err != nil {
		return err
	}
	defer unlock()

	stored, err := config.Load()
	if err != nil {
		return err
	}
	change(stored)
	if err := config.Save(stored); err != nil {
		return err
	}

	m.mu.Lock()
	change(m.cfg)
	m.mu.Unlock()
	return nil
}

func (m *Manager) refreshGrant(ctx context.Context, resource, organizationID string, scopes []string) (string, int, error) {
	m.refreshMu.Lock()
	defer m.refreshMu.Unlock()

	unlock, err := config.Lock()
	if err != nil {
		return "", 0, err
	}
	defer unlock()

	stored, err := config.Load()
	if err != nil {
		return "", 0, err
	}
	m.mu.Lock()
	m.cfg.RefreshToken = stored.RefreshToken
	m.mu.Unlock()
	if stored.RefreshToken == "" {
		return "", 0, ErrLoggedOut
	}

	provider, err := m.provider(ctx)
	if err != nil {
		return "", 0, err
	}

	tokens, err := provider.Refresh(ctx, stored.RefreshToken, resource, organizationID, scopes)
	if err != nil {
		return "", 0, err
	}

	if tokens.RefreshToken != "" && tokens.RefreshToken != stored.RefreshToken {
		m.mu.Lock()
		m.cfg.RefreshToken = tokens.RefreshToken
		m.mu.Unlock()
		stored.RefreshToken = tokens.RefreshToken
		if err := config.Save(stored); err != nil {
			return "", 0, fmt.Errorf("store the rotated refresh token: %w", err)
		}
	}

	if tokens.AccessToken == "" {
		return "", 0, errors.New("the identity provider returned an empty access token")
	}
	return tokens.AccessToken, tokens.ExpiresIn, nil
}

func (m *Manager) ciSession(ctx context.Context) (*ciauth.Session, error) {
	if !ciauth.Available() {
		if ciauth.Detected() {
			return nil, ciauth.ErrNoIDToken
		}
		return nil, nil
	}

	m.ciMu.Lock()
	defer m.ciMu.Unlock()

	if m.ciTried {
		return m.ciResult, m.ciErr
	}

	m.ciTried = true
	m.ciResult, m.ciErr = ciauth.Exchange(ctx, &http.Client{Timeout: 30 * time.Second}, m.APIURL())
	return m.ciResult, m.ciErr
}

func identityFromUserInfo(info *oidc.UserInfo) *api.Identity {
	roleByOrg := map[string]string{}
	for _, entry := range info.OrganizationRoles {
		orgID, roleName, ok := strings.Cut(entry, ":")
		if !ok {
			continue
		}
		if roleName == "admin" {
			roleByOrg[orgID] = "admin"
		} else if _, exists := roleByOrg[orgID]; !exists {
			roleByOrg[orgID] = "user"
		}
	}

	workspaces := make([]api.WorkspaceMembership, 0, len(info.OrganizationData))
	for _, org := range info.OrganizationData {
		role := roleByOrg[org.ID]
		if role == "" {
			role = "user"
		}
		workspaces = append(workspaces, api.WorkspaceMembership{Workspace: org.Name, Role: role})
	}

	return &api.Identity{
		ID:         info.Subject,
		Name:       info.Name,
		Email:      info.Email,
		AvatarURL:  info.Picture,
		Workspaces: workspaces,
	}
}

func tokenExpiry(expiresIn int) time.Time {
	leeway := expiresIn - 60
	if leeway < 0 {
		leeway = expiresIn
	}
	return time.Now().Add(time.Duration(leeway) * time.Second)
}
