package build

import (
	"sync"

	"github.com/zeitlos/lucity/pkg/builder"
)

// BuildState holds the current state of a build.
type BuildState struct {
	ID       string
	Phase    builder.BuildPhase
	ImageRef string
	Digest   string
	Error    string
}

// Tracker manages in-memory build state for async builds.
type Tracker struct {
	mu     sync.RWMutex
	builds map[string]*BuildState
}

// NewTracker creates a new build state tracker.
func NewTracker() *Tracker {
	return &Tracker{
		builds: make(map[string]*BuildState),
	}
}

// Create registers a new build with QUEUED phase.
func (t *Tracker) Create(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.builds[id] = &BuildState{
		ID:    id,
		Phase: builder.BuildPhase_BUILD_PHASE_QUEUED,
	}
}

// Get returns the current state of a build, or nil if not found.
func (t *Tracker) Get(id string) *BuildState {
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
func (t *Tracker) Update(id string, phase builder.BuildPhase) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if s := t.builds[id]; s != nil {
		s.Phase = phase
	}
}

// Succeed marks a build as succeeded with the image ref and digest.
func (t *Tracker) Succeed(id, imageRef, digest string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if s := t.builds[id]; s != nil {
		s.Phase = builder.BuildPhase_BUILD_PHASE_SUCCEEDED
		s.ImageRef = imageRef
		s.Digest = digest
	}
}

// Fail marks a build as failed with an error message.
func (t *Tracker) Fail(id, errMsg string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if s := t.builds[id]; s != nil {
		s.Phase = builder.BuildPhase_BUILD_PHASE_FAILED
		s.Error = errMsg
	}
}
