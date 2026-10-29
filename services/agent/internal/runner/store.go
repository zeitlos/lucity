package runner

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/zeitlos/lucity/services/agent/internal/harness"
)

type index struct {
	Sessions   []Session       `json:"sessions"`
	LastModel  *Model          `json:"lastModel,omitempty"`
	LastEffort string          `json:"lastEffort,omitempty"`
	Routing    harness.Routing `json:"routing"`
}

type store struct {
	dir   string
	mu    sync.Mutex
	index index
}

func openStore(dir string) (*store, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	s := &store{dir: dir}
	raw, err := os.ReadFile(s.indexPath())
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return s, nil
	case err != nil:
		return nil, err
	}
	if err := json.Unmarshal(raw, &s.index); err != nil {
		return nil, fmt.Errorf("parse %s: %w", s.indexPath(), err)
	}
	return s, nil
}

func (s *store) sessions() []Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.index.Sessions)
}

func (s *store) session(id string) (Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, session := range s.index.Sessions {
		if session.ID == id {
			return session, true
		}
	}
	return Session{}, false
}

func (s *store) defaults() (*Model, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.index.LastModel, s.index.LastEffort
}

func (s *store) routing() harness.Routing {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.index.Routing
}

func (s *store) saveRouting(routing harness.Routing) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.index.Routing = routing
	return s.write()
}

func (s *store) save(session Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	position := slices.IndexFunc(s.index.Sessions, func(existing Session) bool { return existing.ID == session.ID })
	if position < 0 {
		s.index.Sessions = append(s.index.Sessions, session)
	} else {
		s.index.Sessions[position] = session
	}
	if session.Model.Name != "" {
		model := session.Model
		s.index.LastModel = &model
	}
	if session.Effort != "" {
		s.index.LastEffort = session.Effort
	}
	return s.write()
}

func (s *store) write() error {
	raw, err := json.MarshalIndent(s.index, "", "  ")
	if err != nil {
		return err
	}
	temporary := s.indexPath() + ".tmp"
	if err := os.WriteFile(temporary, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(temporary, s.indexPath())
}

func (s *store) append(sessionID string, event Event) error {
	raw, err := json.Marshal(event)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(s.eventsPath(sessionID), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := file.Write(append(raw, '\n')); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

func (s *store) events(sessionID string) ([]Event, error) {
	file, err := os.Open(s.eventsPath(sessionID))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var events []Event
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 16<<20)
	for scanner.Scan() {
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}
		events = append(events, event)
	}
	return events, scanner.Err()
}

func (s *store) indexPath() string {
	return filepath.Join(s.dir, "index.json")
}

func (s *store) eventsPath(sessionID string) string {
	return filepath.Join(s.dir, sessionID+".jsonl")
}
