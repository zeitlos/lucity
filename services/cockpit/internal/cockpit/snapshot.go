package cockpit

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// Snapshot is the list of workspaces as loaded at one point in time.
type Snapshot struct {
	LoadedAt   time.Time
	Workspaces []Workspace
}

// Workspaces loads every workspace with its billing, members, environments
// and signals, newest first, and keeps the result as the cached snapshot.
func (c *Client) Workspaces(ctx context.Context) (Snapshot, error) {
	workspaces, err := c.load(ctx)

	if err != nil {
		return Snapshot{}, err
	}

	snapshot := Snapshot{LoadedAt: time.Now(), Workspaces: workspaces}

	if err := c.save(snapshot); err != nil {
		slog.Warn("cache workspaces", "error", err)
	}

	if err := c.saveGitHubCache(); err != nil {
		slog.Warn("cache github lookups", "error", err)
	}

	return snapshot, nil
}

// CachedWorkspaces returns the snapshot of the last load, kept across
// restarts, or nil before the first load.
func (c *Client) CachedWorkspaces() (*Snapshot, error) {
	raw, err := os.ReadFile(c.snapshotPath)

	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	var snapshot Snapshot

	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func (c *Client) save(snapshot Snapshot) error {
	raw, err := json.Marshal(snapshot)

	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(c.snapshotPath), 0o700); err != nil {
		return err
	}

	return os.WriteFile(c.snapshotPath, raw, 0o600)
}
