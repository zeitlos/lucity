package bridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"
)

type Threads struct {
	path     string
	mu       sync.Mutex
	sessions map[string]string
}

func OpenThreads(path string) (*Threads, error) {
	threads := &Threads{path: path, sessions: map[string]string{}}
	raw, err := os.ReadFile(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return threads, nil
	case err != nil:
		return nil, err
	}
	if err := json.Unmarshal(raw, &threads.sessions); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return threads, nil
}

func (t *Threads) Session(thread string) (string, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	sessionID, ok := t.sessions[thread]
	return sessionID, ok
}

func (t *Threads) SetSession(thread, sessionID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.sessions[thread] = sessionID

	raw, err := json.Marshal(t.sessions)
	if err != nil {
		return err
	}
	temporary := t.path + ".tmp"
	if err := os.WriteFile(temporary, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(temporary, t.path)
}

func threadKey(chatID, threadID int64) string {
	return fmt.Sprintf("%d:%d", chatID, threadID)
}
