package deploy

import "sync"

// Phase represents a stage in the unified deploy pipeline.
type Phase string

const (
	PhaseQueued    Phase = "QUEUED"
	PhaseCloning   Phase = "CLONING"
	PhaseBuilding  Phase = "BUILDING"
	PhasePushing   Phase = "PUSHING"
	PhaseDeploying Phase = "DEPLOYING"
	PhaseSucceeded Phase = "SUCCEEDED"
	PhaseFailed    Phase = "FAILED"
)

const maxLogLines = 5000

// State holds the current state of a deploy operation.
type State struct {
	ID          string
	Phase       Phase
	BuildID     string
	ImageRef    string
	Digest      string
	Error       string
	Project     string
	Service     string
	Environment string

	// Rollout health status, populated during DEPLOYING phase and after SUCCEEDED.
	RolloutHealth  string // SYNCED, PROGRESSING, DEGRADED, OUT_OF_SYNC, UNKNOWN
	RolloutMessage string // detailed reason, e.g. "ImagePullBackOff on web-abc123"

	logs      []string
	listeners []chan string
	done      chan struct{} // closed when deploy reaches terminal phase
}

// serviceKey builds a lookup key for project+service+environment.
func serviceKey(project, service, environment string) string {
	return project + "/" + service + "/" + environment
}

// Tracker manages in-memory deploy state for async build+deploy operations.
type Tracker struct {
	mu      sync.RWMutex
	deploys map[string]*State
	// byService maps project/service/environment to the latest deploy ID.
	byService map[string]string
}

// NewTracker creates a new deploy state tracker.
func NewTracker() *Tracker {
	return &Tracker{
		deploys:   make(map[string]*State),
		byService: make(map[string]string),
	}
}

// Create registers a new deploy with QUEUED phase.
func (t *Tracker) Create(id, buildID, project, service, environment string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.deploys[id] = &State{
		ID:          id,
		Phase:       PhaseQueued,
		BuildID:     buildID,
		Project:     project,
		Service:     service,
		Environment: environment,
		done:        make(chan struct{}),
	}
	t.byService[serviceKey(project, service, environment)] = id
}

// ActiveForService returns the in-flight deploy for a project/service/environment, or nil.
func (t *Tracker) ActiveForService(project, service, environment string) *State {
	t.mu.RLock()
	defer t.mu.RUnlock()
	id, ok := t.byService[serviceKey(project, service, environment)]
	if !ok {
		return nil
	}
	s := t.deploys[id]
	if s == nil {
		return nil
	}
	// Only return if still in-flight.
	if s.Phase == PhaseSucceeded || s.Phase == PhaseFailed {
		return nil
	}
	cp := t.copyState(s)
	return cp
}

// Get returns the current state of a deploy, or nil if not found.
func (t *Tracker) Get(id string) *State {
	t.mu.RLock()
	defer t.mu.RUnlock()
	s := t.deploys[id]
	if s == nil {
		return nil
	}
	return t.copyState(s)
}

// Update sets the phase of a deploy.
func (t *Tracker) Update(id string, phase Phase) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if s := t.deploys[id]; s != nil {
		s.Phase = phase
	}
}

// Succeed marks a deploy as succeeded with the image ref and digest.
func (t *Tracker) Succeed(id, imageRef, digest string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if s := t.deploys[id]; s != nil {
		s.Phase = PhaseSucceeded
		s.ImageRef = imageRef
		s.Digest = digest
		t.closeDone(s)
	}
}

// Fail marks a deploy as failed with an error message.
func (t *Tracker) Fail(id, errMsg string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if s := t.deploys[id]; s != nil {
		s.Phase = PhaseFailed
		s.Error = errMsg
		t.closeDone(s)
	}
}

// UpdateRolloutHealth updates the rollout health status for a deploy.
func (t *Tracker) UpdateRolloutHealth(id, health, message string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if s := t.deploys[id]; s != nil {
		s.RolloutHealth = health
		s.RolloutMessage = message
	}
}

// AppendLog adds a log line to a deploy and notifies all subscribers.
func (t *Tracker) AppendLog(id, line string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	s := t.deploys[id]
	if s == nil || len(s.logs) >= maxLogLines {
		return
	}
	s.logs = append(s.logs, line)
	// Notify listeners (non-blocking).
	for _, ch := range s.listeners {
		select {
		case ch <- line:
		default:
		}
	}
}

// Subscribe returns a channel that receives new log lines for a deploy
// and an unsubscribe function. Existing lines are NOT replayed — call
// LogLines first if you need the backlog.
func (t *Tracker) Subscribe(id string) (<-chan string, func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	s := t.deploys[id]
	if s == nil {
		ch := make(chan string)
		close(ch)
		return ch, func() {}
	}
	ch := make(chan string, 64)
	s.listeners = append(s.listeners, ch)
	unsub := func() {
		t.mu.Lock()
		defer t.mu.Unlock()
		for i, l := range s.listeners {
			if l == ch {
				s.listeners = append(s.listeners[:i], s.listeners[i+1:]...)
				break
			}
		}
	}
	return ch, unsub
}

// LogLines returns log lines starting from offset.
func (t *Tracker) LogLines(id string, offset int) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	s := t.deploys[id]
	if s == nil || offset >= len(s.logs) {
		return nil
	}
	lines := make([]string, len(s.logs)-offset)
	copy(lines, s.logs[offset:])
	return lines
}

// Done returns a channel that's closed when the deploy reaches a terminal phase.
func (t *Tracker) Done(id string) <-chan struct{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	s := t.deploys[id]
	if s == nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	return s.done
}

// closeDone closes the done channel and all listener channels for a deploy.
// Must be called with t.mu held.
func (t *Tracker) closeDone(s *State) {
	select {
	case <-s.done:
		// already closed
	default:
		close(s.done)
	}
	for _, ch := range s.listeners {
		close(ch)
	}
	s.listeners = nil
}

// copyState returns a shallow copy of State (without internal fields).
func (t *Tracker) copyState(s *State) *State {
	cp := *s
	cp.logs = nil
	cp.listeners = nil
	cp.done = nil
	return &cp
}
