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
	cp := *s
	return &cp
}

// Get returns the current state of a deploy, or nil if not found.
func (t *Tracker) Get(id string) *State {
	t.mu.RLock()
	defer t.mu.RUnlock()
	s := t.deploys[id]
	if s == nil {
		return nil
	}
	// Return a copy to avoid races.
	cp := *s
	return &cp
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
	}
}

// Fail marks a deploy as failed with an error message.
func (t *Tracker) Fail(id, errMsg string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if s := t.deploys[id]; s != nil {
		s.Phase = PhaseFailed
		s.Error = errMsg
	}
}
