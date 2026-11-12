// Package updatecheck finds out whether a newer release of the CLI is out, asking
// GitHub at most once a day and caching the answer in the user cache directory.
package updatecheck

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	releaseURL = "https://api.github.com/repos/zeitlos/lucity/releases/latest"
	interval   = 24 * time.Hour
	timeout    = 2 * time.Second
)

type state struct {
	Latest    string    `json:"latest,omitempty"`
	CheckedAt time.Time `json:"checkedAt"`
}

// Disabled reports whether the user turned the check off with LUCITY_NO_UPDATE_CHECK.
func Disabled() bool {
	return os.Getenv("LUCITY_NO_UPDATE_CHECK") != ""
}

// Cached returns the latest release from the last check, and whether that check
// is more than a day old.
func Cached() (string, bool) {
	file, err := statePath()
	if err != nil {
		return "", true
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return "", true
	}
	var cached state
	if err := json.Unmarshal(data, &cached); err != nil {
		return "", true
	}
	return cached.Latest, time.Since(cached.CheckedAt) > interval
}

// Latest returns the cached latest release, asking GitHub first when the cache
// is more than a day old.
func Latest(ctx context.Context) string {
	latest, stale := Cached()
	if !stale {
		return latest
	}
	if fetched, err := Refresh(ctx); err == nil {
		return fetched
	}
	return latest
}

// Refresh asks GitHub for the latest release and caches the answer. A failed
// request is cached as a check too, so an offline machine asks once a day.
func Refresh(ctx context.Context) (string, error) {
	previous, _ := Cached()
	latest, err := fetch(ctx)
	if err != nil {
		save(state{Latest: previous, CheckedAt: time.Now()})
		return "", err
	}
	save(state{Latest: latest, CheckedAt: time.Now()})
	return latest, nil
}

// Newer reports whether latest is a newer CalVer release than current. Builds
// without a release version, such as dev builds, are never outdated.
func Newer(current, latest string) bool {
	currentParts, ok := parse(current)
	if !ok {
		return false
	}
	latestParts, ok := parse(latest)
	if !ok {
		return false
	}
	for i, part := range latestParts {
		if i >= len(currentParts) {
			return true
		}
		if part != currentParts[i] {
			return part > currentParts[i]
		}
	}
	return false
}

// UpgradeHint tells how to upgrade the running binary.
func UpgradeHint() string {
	executable, err := os.Executable()
	if err == nil {
		if resolved, err := filepath.EvalSymlinks(executable); err == nil {
			executable = resolved
		}
		if strings.Contains(executable, "/Caskroom/") {
			return "brew upgrade --cask lucity"
		}
	}
	return "see https://lucity.cloud/docs/cli#update"
}

func fetch(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, releaseURL, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "lucity-cli")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("latest release: %s", response.Status)
	}
	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return "", err
	}
	latest := strings.TrimPrefix(release.TagName, "v")
	if _, ok := parse(latest); !ok {
		return "", fmt.Errorf("latest release has an unexpected tag %q", release.TagName)
	}
	return latest, nil
}

func parse(version string) ([]int, bool) {
	version = strings.TrimPrefix(version, "v")
	version, _, _ = strings.Cut(version, "-")
	fields := strings.Split(version, ".")
	parts := make([]int, 0, len(fields))
	for _, field := range fields {
		part, err := strconv.Atoi(field)
		if err != nil {
			return nil, false
		}
		parts = append(parts, part)
	}
	return parts, true
}

func statePath() (string, error) {
	cacheDirectory, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDirectory, "lucity", "update-check.json"), nil
}

func save(current state) error {
	file, err := statePath()
	if err != nil {
		return err
	}
	data, err := json.Marshal(current)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(file), ".update-check-")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	_, writeErr := temporary.Write(data)
	closeErr := temporary.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return err
	}
	return os.Rename(temporary.Name(), file)
}
