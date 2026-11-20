// Package selfupdate keeps the CLI on the latest release. It asks GitHub at most
// once a day and replaces the running binary with a newer, signed release.
package selfupdate

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

	"golang.org/x/sys/unix"
)

const (
	interval     = 24 * time.Hour
	checkTimeout = 2 * time.Second
)

var (
	releaseAPI  = "https://api.github.com/repos/zeitlos/lucity/releases/latest"
	downloadURL = "https://github.com/zeitlos/lucity/releases/download"
)

type state struct {
	Latest      string    `json:"latest,omitempty"`
	CheckedAt   time.Time `json:"checkedAt"`
	FailedTo    string    `json:"failedTo,omitempty"`
	UpdatedFrom string    `json:"updatedFrom,omitempty"`
	UpdatedTo   string    `json:"updatedTo,omitempty"`
}

// Disabled reports whether the user turned checking for releases off with
// LUCITY_NO_UPDATE_CHECK, which turns automatic updates off too.
func Disabled() bool {
	return os.Getenv("LUCITY_NO_UPDATE_CHECK") != ""
}

// AutoUpdates reports whether the CLI at version current may replace itself:
// a release build, outside CI, not opted out, in a directory the user can write to.
func AutoUpdates(current string) bool {
	if Disabled() || os.Getenv("LUCITY_NO_AUTO_UPDATE") != "" || os.Getenv("CI") != "" || publicKey == "" {
		return false
	}
	if _, ok := parse(current); !ok {
		return false
	}
	return Writable()
}

// Writable reports whether the user can replace the running binary.
func Writable() bool {
	executable, err := executablePath()
	if err != nil {
		return false
	}
	return unix.Access(filepath.Dir(executable), unix.W_OK) == nil
}

// Cached returns the latest release from the last check, and whether that check
// is more than a day old.
func Cached() (string, bool) {
	cached := load()
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
	latest, err := fetchLatest(ctx)
	current := load()
	current.CheckedAt = time.Now()
	if err == nil {
		current.Latest = latest
	}
	save(current)
	return latest, err
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

// Notice describes what the user should do about a newer release, or returns ""
// when there is nothing to do, including when the CLI is about to update itself.
func Notice(current, latest string) string {
	if !Newer(current, latest) {
		return ""
	}
	available := fmt.Sprintf("lucity %s is available (you have %s)", latest, current)
	if AutoUpdates(current) {
		if load().FailedTo == latest {
			return available + ", but updating automatically failed: run 'lucity update' to see why"
		}
		return ""
	}
	if !Writable() {
		return available + ": run 'sudo lucity update'"
	}
	return available + ": run 'lucity update'"
}

// MarkUpdated remembers an automatic update, for the updated version to mention once.
func MarkUpdated(from, to string) {
	current := load()
	current.UpdatedFrom = from
	current.UpdatedTo = to
	save(current)
}

// Updated returns the version current replaced in an automatic update, once.
func Updated(current string) (string, bool) {
	previous := load()
	if previous.UpdatedTo == "" || previous.UpdatedTo != current {
		return "", false
	}
	from := previous.UpdatedFrom
	previous.UpdatedFrom = ""
	previous.UpdatedTo = ""
	save(previous)
	return from, true
}

func fetchLatest(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, releaseAPI, nil)
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

func cacheDirectory() (string, error) {
	directory, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "lucity"), nil
}

func load() state {
	var current state
	directory, err := cacheDirectory()
	if err != nil {
		return current
	}
	data, err := os.ReadFile(filepath.Join(directory, "update-check.json"))
	if err != nil {
		return current
	}
	json.Unmarshal(data, &current)
	return current
}

func save(current state) error {
	directory, err := cacheDirectory()
	if err != nil {
		return err
	}
	data, err := json.Marshal(current)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(directory, ".update-check-")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	_, writeErr := temporary.Write(data)
	closeErr := temporary.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return err
	}
	return os.Rename(temporary.Name(), filepath.Join(directory, "update-check.json"))
}
