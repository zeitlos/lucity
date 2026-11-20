package cockpit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	gh "github.com/google/go-github/v92/github"

	"github.com/zeitlos/lucity/pkg/logto"
)

const githubConcurrency = 16

type githubCache struct {
	Accounts map[int64]*Account `json:"accounts"`
	Parents  map[string]string  `json:"parents"`
}

func identityAccount(identity logto.UserIdentity) *Account {
	raw, _ := identity.Details["rawData"].(map[string]interface{})
	info, _ := raw["userInfo"].(map[string]interface{})
	login, _ := info["login"].(string)
	created, _ := info["created_at"].(string)
	createdAt, err := time.Parse(time.RFC3339, created)

	if login == "" || err != nil {
		return nil
	}

	return &Account{Login: login, CreatedAt: createdAt}
}

func (c *Client) account(ctx context.Context, id int64) (*Account, error) {
	c.mu.Lock()
	account, ok := c.accounts[id]
	c.mu.Unlock()

	if ok {
		return account, nil
	}

	user, _, err := c.github.Users.GetByID(ctx, id)

	if err != nil && !notFound(err) {
		return nil, fmt.Errorf("github user %d: %w", id, err)
	}

	if user != nil {
		account = &Account{Login: user.GetLogin(), CreatedAt: user.GetCreatedAt().Time}
	}

	c.mu.Lock()
	c.accounts[id] = account
	c.mu.Unlock()

	return account, nil
}

func (c *Client) parent(ctx context.Context, repositoryURL string) (string, error) {
	owner, name, ok := githubRepository(repositoryURL)

	if !ok {
		return "", nil
	}

	key := strings.ToLower(owner + "/" + name)

	c.mu.Lock()
	parent, ok := c.parents[key]
	c.mu.Unlock()

	if ok {
		return parent, nil
	}

	repository, _, err := c.github.Repositories.Get(ctx, owner, name)

	if err != nil && !notFound(err) {
		return "", fmt.Errorf("github repository %s: %w", key, err)
	}

	if repository != nil && repository.GetFork() {
		parent = repository.GetParent().GetFullName()
	}

	c.mu.Lock()
	c.parents[key] = parent
	c.mu.Unlock()

	return parent, nil
}

func (c *Client) accountsByID(ctx context.Context, ids []int64) map[int64]*Account {
	accounts := make([]*Account, len(ids))
	lookupEach(len(ids), func(i int) error {
		var err error
		accounts[i], err = c.account(ctx, ids[i])
		return err
	})

	result := make(map[int64]*Account, len(ids))

	for i, id := range ids {
		result[id] = accounts[i]
	}

	return result
}

func (c *Client) parentsByURL(ctx context.Context, urls []string) map[string]string {
	parents := make([]string, len(urls))
	lookupEach(len(urls), func(i int) error {
		var err error
		parents[i], err = c.parent(ctx, urls[i])
		return err
	})

	result := make(map[string]string, len(urls))

	for i, url := range urls {
		result[url] = parents[i]
	}

	return result
}

func lookupEach(count int, lookup func(i int) error) {
	var group sync.WaitGroup
	var once sync.Once
	slots := make(chan struct{}, githubConcurrency)

	for i := range count {
		slots <- struct{}{}
		group.Go(func() {
			defer func() { <-slots }()

			if err := lookup(i); err != nil {
				once.Do(func() { slog.Warn("github lookups skipped, retried on the next load", "error", err) })
			}
		})
	}

	group.Wait()
}

func (c *Client) loadGitHubCache() {
	raw, err := os.ReadFile(c.githubCachePath)

	if err != nil {
		return
	}

	var cache githubCache

	if err := json.Unmarshal(raw, &cache); err != nil {
		slog.Warn("read github cache", "error", err)
		return
	}

	if cache.Accounts != nil {
		c.accounts = cache.Accounts
	}

	if cache.Parents != nil {
		c.parents = cache.Parents
	}
}

func (c *Client) saveGitHubCache() error {
	c.mu.Lock()
	raw, err := json.Marshal(githubCache{Accounts: c.accounts, Parents: c.parents})
	c.mu.Unlock()

	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(c.githubCachePath), 0o700); err != nil {
		return err
	}

	return os.WriteFile(c.githubCachePath, raw, 0o600)
}

func githubRepository(url string) (string, string, bool) {
	path, ok := strings.CutPrefix(strings.TrimSuffix(url, ".git"), "https://github.com/")

	if !ok {
		return "", "", false
	}

	owner, name, ok := strings.Cut(path, "/")

	return owner, name, ok && owner != "" && name != "" && !strings.Contains(name, "/")
}

func githubID(value string) (int64, bool) {
	id, err := strconv.ParseInt(value, 10, 64)
	return id, err == nil && id > 0
}

func notFound(err error) bool {
	var response *gh.ErrorResponse
	return errors.As(err, &response) && response.Response != nil && response.Response.StatusCode == http.StatusNotFound
}
