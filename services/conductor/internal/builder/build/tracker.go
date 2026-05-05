package build

import (
	"sync"

	"github.com/zeitlos/lucity/pkg/builder"
)

const maxLogLines = 5000

// BuildState holds the current state of a build.
type BuildState struct {
	ID       string
	Phase    builder.BuildPhase
	ImageRef string
	Digest   string
	Error    string
	Logs     []string
}

// Tracker tracks build state. Implementations may store state in-memory
// or read it from external systems like Kubernetes.
type Tracker interface {
	Create(id string)
	Get(id string) *BuildState
	Update(id string, phase builder.BuildPhase)
	Succeed(id, imageRef, digest string)
	Fail(id, errMsg string)
	AppendLog(id, line string)
	LogLines(id string, offset int) []string
	LogCount(id string) int
	IsTerminal(id string) bool
}

// InMemoryTracker manages build state in-memory for async builds.
type InMemoryTracker struct {
	mu     sync.RWMutex
	builds map[string]*BuildState
}

// NewInMemoryTracker creates a new in-memory build state tracker.
func NewInMemoryTracker() *InMemoryTracker {
	return &InMemoryTracker{
		builds: make(map[string]*BuildState),
	}
}

// Create registers a new build with QUEUED phase.
func (t *InMemoryTracker) Create(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.builds[id] = &BuildState{
		ID:    id,
		Phase: builder.BuildPhase_BUILD_PHASE_QUEUED,
	}
}

// Get returns the current state of a build, or nil if not found.
func (t *InMemoryTracker) Get(id string) *BuildState {
	t.mu.RLock()
	defer t.mu.RUnlock()
	s := t.builds[id]
	if s == nil {
		return nil
	}
	// Return a copy to avoid races.
	cp := *s
	return &cp
}

// Update sets the phase of a build.
func (t *InMemoryTracker) Update(id string, phase builder.BuildPhase) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if s := t.builds[id]; s != nil {
		s.Phase = phase
	}
}

// Succeed marks a build as succeeded with the image ref and digest.
func (t *InMemoryTracker) Succeed(id, imageRef, digest string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if s := t.builds[id]; s != nil {
		s.Phase = builder.BuildPhase_BUILD_PHASE_SUCCEEDED
		s.ImageRef = imageRef
		s.Digest = digest
	}
}

// Fail marks a build as failed with an error message.
func (t *InMemoryTracker) Fail(id, errMsg string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if s := t.builds[id]; s != nil {
		s.Phase = builder.BuildPhase_BUILD_PHASE_FAILED
		s.Error = errMsg
	}
}

// AppendLog adds a log line to a build. Capped at maxLogLines.
func (t *InMemoryTracker) AppendLog(id, line string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	s := t.builds[id]
	if s == nil || len(s.Logs) >= maxLogLines {
		return
	}
	s.Logs = append(s.Logs, line)
}

// LogLines returns log lines starting from offset, or nil if build not found.
func (t *InMemoryTracker) LogLines(id string, offset int) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	s := t.builds[id]
	if s == nil || offset >= len(s.Logs) {
		return nil
	}
	// Return a copy of the slice segment.
	lines := make([]string, len(s.Logs)-offset)
	copy(lines, s.Logs[offset:])
	return lines
}

// LogCount returns the number of log lines for a build.
func (t *InMemoryTracker) LogCount(id string) int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if s := t.builds[id]; s != nil {
		return len(s.Logs)
	}
	return 0
}

// IsTerminal returns true if the build is in a terminal phase (SUCCEEDED or FAILED).
func (t *InMemoryTracker) IsTerminal(id string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	s := t.builds[id]
	if s == nil {
		return true // not found = treat as done
	}
	return s.Phase == builder.BuildPhase_BUILD_PHASE_SUCCEEDED || s.Phase == builder.BuildPhase_BUILD_PHASE_FAILED
}
