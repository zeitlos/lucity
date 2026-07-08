package session

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/zeitlos/lucity/cli/internal/api"
	"github.com/zeitlos/lucity/cli/internal/config"
)

const refreshLeeway = 48 * time.Hour

var ErrLoggedOut = errors.New("not logged in — run `lucity login`")

type Manager struct {
	mu  sync.Mutex
	cfg *config.Config
}

func Load() (*Manager, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return &Manager{cfg: cfg}, nil
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
	return m.cfg.Workspace
}

func (m *Manager) SetWorkspace(workspace string) error {
	m.cfg.Workspace = workspace
	return config.Save(m.cfg)
}

func (m *Manager) SetSession(apiURL string, s *api.Session) error {
	m.cfg.APIURL = apiURL
	m.cfg.Token = s.Token
	m.cfg.RefreshToken = s.RefreshToken
	m.cfg.LogtoToken = s.LogtoToken
	m.cfg.ExpiresAt = s.ExpiresAt
	return config.Save(m.cfg)
}

func (m *Manager) Clear() error {
	m.cfg.Token = ""
	m.cfg.RefreshToken = ""
	m.cfg.LogtoToken = ""
	m.cfg.ExpiresAt = time.Time{}
	return config.Save(m.cfg)
}

func (m *Manager) Client() *api.Client {
	return api.NewClient(m.APIURL(), m.Workspace(), m)
}

func (m *Manager) CookieTokens(context.Context) (logtoToken, refreshToken string) {
	if token := os.Getenv("LUCITY_LOGTO_TOKEN"); token != "" {
		return token, os.Getenv("LUCITY_REFRESH_TOKEN")
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cfg.LogtoToken, m.cfg.RefreshToken
}

func (m *Manager) PersistRotatedTokens(sessionToken, logtoToken, refreshToken string) {
	if os.Getenv("LUCITY_TOKEN") != "" || os.Getenv("LUCITY_LOGTO_TOKEN") != "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	changed := false
	if sessionToken != "" && sessionToken != m.cfg.Token {
		m.cfg.Token = sessionToken
		changed = true
	}
	if logtoToken != "" && logtoToken != m.cfg.LogtoToken {
		m.cfg.LogtoToken = logtoToken
		changed = true
	}
	if refreshToken != "" && refreshToken != m.cfg.RefreshToken {
		m.cfg.RefreshToken = refreshToken
		changed = true
	}
	if changed {
		_ = config.Save(m.cfg)
	}
}

func (m *Manager) Token(ctx context.Context) (string, error) {
	if token := os.Getenv("LUCITY_TOKEN"); token != "" {
		return token, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cfg.Token == "" {
		return "", ErrLoggedOut
	}
	if m.cfg.ExpiresAt.IsZero() || time.Until(m.cfg.ExpiresAt) > refreshLeeway {
		return m.cfg.Token, nil
	}
	if time.Now().After(m.cfg.ExpiresAt) {
		if m.cfg.RefreshToken == "" {
			return "", fmt.Errorf("session expired — run `lucity login`")
		}
	}

	refreshed, err := api.Refresh(ctx, &http.Client{Timeout: 30 * time.Second}, m.APIURL(), m.cfg.Token, m.cfg.RefreshToken)
	if err != nil {
		if time.Now().Before(m.cfg.ExpiresAt) {
			return m.cfg.Token, nil
		}
		return "", fmt.Errorf("session expired and refresh failed (%w) — run `lucity login`", err)
	}

	m.cfg.Token = refreshed.Token
	if refreshed.RefreshToken != "" {
		m.cfg.RefreshToken = refreshed.RefreshToken
	}
	if refreshed.LogtoToken != "" {
		m.cfg.LogtoToken = refreshed.LogtoToken
	}
	m.cfg.ExpiresAt = refreshed.ExpiresAt
	if err := config.Save(m.cfg); err != nil {
		return "", err
	}
	return m.cfg.Token, nil
}
