package runner

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/zeitlos/lucity/pkg/acp"
	"github.com/zeitlos/lucity/services/agent/internal/harness"
)

const failedAfter = 3

type Runner struct {
	harnessName     string
	harness         harness.Harness
	commandOverride []string
	connectors      []string
	workdir         string
	approvalTimeout time.Duration
	store           *store

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}

	// starting serializes everything that creates, resumes or routes harness
	// sessions, so the harness only restarts while no turn is running.
	starting sync.Mutex

	mu        sync.Mutex
	client    *acp.Client
	ready     chan struct{}
	routing   harness.Routing
	failure   error
	failures  int
	sessions  map[string]*session
	byHarness map[string]*session
}

func New(harnessName string, h harness.Harness, commandOverride, connectors []string, workdir, sessionsDir string, approvalTimeout time.Duration) (*Runner, error) {
	store, err := openStore(sessionsDir)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Runner{
		harnessName:     harnessName,
		harness:         h,
		commandOverride: commandOverride,
		connectors:      connectors,
		workdir:         workdir,
		approvalTimeout: approvalTimeout,
		store:           store,
		ctx:             ctx,
		cancel:          cancel,
		done:            make(chan struct{}),
		ready:           make(chan struct{}),
		routing:         store.routing(),
		sessions:        map[string]*session{},
		byHarness:       map[string]*session{},
	}, nil
}

func (r *Runner) Label() string {
	return "runner"
}

func (r *Runner) Start() error {
	defer close(r.done)
	for r.ctx.Err() == nil {
		r.mu.Lock()
		command, err := r.harness.Command(r.routing)
		r.mu.Unlock()
		if len(r.commandOverride) > 0 {
			command.Path, command.Args = r.commandOverride[0], r.commandOverride[1:]
		}

		var client *acp.Client
		if err == nil {
			client, err = acp.Start(r.ctx, "lucity-agent", command, r)
		}
		if err != nil {
			failures := r.fail(err)
			slog.Error("harness failed to start", "harness", r.harnessName, "error", err, "failures", failures)
			sleep(r.ctx, backoff(failures))
			continue
		}

		r.mu.Lock()
		r.client = client
		r.failure = nil
		r.failures = 0
		close(r.ready)
		r.mu.Unlock()
		slog.Info("harness started", "harness", r.harnessName, "command", command.Path)

		<-client.Done()

		r.mu.Lock()
		restarted := r.client != client
		if !restarted {
			r.client = nil
			r.ready = make(chan struct{})
		}
		r.mu.Unlock()
		if r.ctx.Err() == nil && !restarted {
			failures := r.fail(fmt.Errorf("harness exited: %v", client.Err()))
			slog.Error("harness exited, restarting", "harness", r.harnessName, "error", client.Err(), "failures", failures)
			sleep(r.ctx, backoff(failures))
		}
	}
	return nil
}

func (r *Runner) Shutdown(ctx context.Context) error {
	r.cancel()
	select {
	case <-r.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *Runner) Status() Status {
	r.mu.Lock()
	client, failure, failures := r.client, r.failure, r.failures
	sessions := r.loaded()
	r.mu.Unlock()

	lastModel, lastEffort := r.store.defaults()
	status := Status{Harness: r.harnessName, Status: StatusIdle, LastModel: lastModel, LastEffort: lastEffort}
	if client == nil {
		status.Status = StatusStarting
		if failures >= failedAfter {
			status.Status = StatusFailed
		}
		if failure != nil {
			status.Error = failure.Error()
		}
		return status
	}
	for _, s := range sessions {
		switch s.status() {
		case StatusWaitingForApproval:
			status.Status = StatusWaitingForApproval
			return status
		case StatusWorking:
			status.Status = StatusWorking
		}
	}
	return status
}

func (r *Runner) Sessions() []SessionStatus {
	records := r.store.sessions()
	slices.SortFunc(records, func(a, b Session) int { return b.StartedAt.Compare(a.StartedAt) })

	r.mu.Lock()
	defer r.mu.Unlock()
	sessions := make([]SessionStatus, 0, len(records))
	for _, record := range records {
		status := StatusIdle
		if s, ok := r.sessions[record.ID]; ok {
			status = s.status()
		}
		sessions = append(sessions, SessionStatus{Session: record, Status: status})
	}
	return sessions
}

func (r *Runner) Session(id string) (SessionDetail, error) {
	s, err := r.session(id)
	if err != nil {
		return SessionDetail{}, err
	}
	return s.detail(), nil
}

func (r *Runner) Create(ctx context.Context, message string, model *Model, effort, startedBy string) (SessionStatus, error) {
	if err := validate(message, effort); err != nil {
		return SessionStatus{}, err
	}
	lastModel, lastEffort := r.store.defaults()
	if model == nil {
		model = lastModel
	}
	if model == nil {
		return SessionStatus{}, InvalidError{errors.New("model is required")}
	}
	if effort == "" {
		effort = lastEffort
	}
	if _, err := r.harness.ModelValue(*model); err != nil {
		return SessionStatus{}, InvalidError{err}
	}

	r.starting.Lock()
	defer r.starting.Unlock()

	client, err := r.route(ctx, *model)
	if err != nil {
		return SessionStatus{}, err
	}
	created, err := client.NewSession(ctx, r.workdir, r.harness.MCPServers(), r.harness.SessionMeta(*model))
	if err != nil {
		return SessionStatus{}, harnessError(err)
	}

	s := newSession(Session{
		ID:               rand.Text(),
		Title:            title(message),
		StartedBy:        startedBy,
		StartedAt:        time.Now().UTC(),
		Model:            *model,
		Effort:           effort,
		HarnessSessionID: created.ID,
	}, nil)
	r.register(s)

	options, err := r.apply(ctx, client, created.ID, created.ConfigOptions, *model, effort)
	if err != nil {
		r.unregister(s)
		return SessionStatus{}, err
	}
	s.mu.Lock()
	s.attached = client
	s.options = options
	record := s.record
	s.mu.Unlock()

	if err := r.store.save(record); err != nil {
		r.unregister(s)
		return SessionStatus{}, err
	}
	if err := r.prompt(s, client, message, startedBy); err != nil {
		return SessionStatus{}, err
	}
	return s.snapshot(), nil
}

func (r *Runner) Send(ctx context.Context, id, message, author string) (SessionStatus, error) {
	if err := validate(message, ""); err != nil {
		return SessionStatus{}, err
	}
	s, err := r.session(id)
	if err != nil {
		return SessionStatus{}, err
	}

	r.starting.Lock()
	defer r.starting.Unlock()

	if s.status() != StatusIdle {
		return SessionStatus{}, ErrBusy
	}
	client, err := r.attach(ctx, s)
	if err != nil {
		return SessionStatus{}, err
	}
	if err := r.prompt(s, client, message, author); err != nil {
		return SessionStatus{}, err
	}
	return s.snapshot(), nil
}

func (r *Runner) SetModel(ctx context.Context, id string, model *Model, effort string) (SessionStatus, error) {
	if effort != "" && !slices.Contains(efforts, effort) {
		return SessionStatus{}, InvalidError{fmt.Errorf("effort must be one of %s", strings.Join(efforts, ", "))}
	}
	if model != nil {
		if _, err := r.harness.ModelValue(*model); err != nil {
			return SessionStatus{}, InvalidError{err}
		}
	}
	s, err := r.session(id)
	if err != nil {
		return SessionStatus{}, err
	}

	r.starting.Lock()
	defer r.starting.Unlock()

	s.mu.Lock()
	previous := s.record
	providerChanged := model != nil && (model.Provider != previous.Model.Provider || model.API != previous.Model.API)
	if providerChanged && s.statusLocked() != StatusIdle {
		s.mu.Unlock()
		return SessionStatus{}, fmt.Errorf("%w: switch the model provider once the session is idle", ErrBusy)
	}
	if model != nil {
		s.record.Model = *model
	}
	if effort != "" {
		s.record.Effort = effort
	}
	if providerChanged {
		s.attached = nil
	}
	record := s.record
	s.mu.Unlock()

	restore := func() {
		s.mu.Lock()
		s.record.Model, s.record.Effort = previous.Model, previous.Effort
		s.mu.Unlock()
	}
	client, err := r.attach(ctx, s)
	if err != nil {
		restore()
		return SessionStatus{}, err
	}
	s.mu.Lock()
	harnessSessionID, options := s.record.HarnessSessionID, s.options
	s.mu.Unlock()
	options, err = r.apply(ctx, client, harnessSessionID, options, record.Model, record.Effort)
	if err != nil {
		restore()
		return SessionStatus{}, err
	}

	s.mu.Lock()
	s.options = options
	record = s.record
	s.mu.Unlock()
	if err := r.store.save(record); err != nil {
		return SessionStatus{}, err
	}
	return s.snapshot(), nil
}

func (r *Runner) Answer(id, approvalID, outcome, decidedBy string) (SessionStatus, error) {
	if outcome != DecisionAllowOnce && outcome != DecisionDeny {
		return SessionStatus{}, InvalidError{fmt.Errorf("decision must be %s or %s", DecisionAllowOnce, DecisionDeny)}
	}
	s, err := r.session(id)
	if err != nil {
		return SessionStatus{}, err
	}
	s.mu.Lock()
	decisions, pending := s.approvals[approvalID]
	asked := slices.ContainsFunc(s.events, func(event Event) bool {
		return event.Approval != nil && event.Approval.ApprovalID == approvalID
	})
	s.mu.Unlock()
	switch {
	case !pending && asked:
		return SessionStatus{}, ErrAnswered
	case !pending:
		return SessionStatus{}, ErrNotFound
	}
	recorded := make(chan struct{})
	select {
	case decisions <- decision{outcome: outcome, decidedBy: decidedBy, recorded: recorded}:
	default:
		return SessionStatus{}, ErrAnswered
	}
	select {
	case <-recorded:
	case <-time.After(5 * time.Second):
	}
	return s.snapshot(), nil
}

func (r *Runner) Stop(id string) (SessionStatus, error) {
	s, err := r.session(id)
	if err != nil {
		return SessionStatus{}, err
	}
	s.mu.Lock()
	if s.turn != nil {
		s.turn.cancel()
	}
	s.mu.Unlock()
	return s.snapshot(), nil
}

func (r *Runner) Events(id string, after int64) ([]Event, error) {
	s, err := r.session(id)
	if err != nil {
		return nil, err
	}
	return s.eventsAfter(after), nil
}

func (r *Runner) Subscribe(id string) (<-chan struct{}, func(), error) {
	s, err := r.session(id)
	if err != nil {
		return nil, nil, err
	}
	waiter, unsubscribe := s.subscribe()
	return waiter, unsubscribe, nil
}

func (r *Runner) Update(harnessSessionID string, update acp.Update) {
	s := r.byHarnessID(harnessSessionID)
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.replaying {
		return
	}
	switch update.Kind {
	case "agent_message_chunk":
		s.appendMessage(update.Text())
	case "tool_call", "tool_call_update":
		if call, changed := s.toolCall(update, r.connectors); changed {
			r.flush(s)
			r.record(s, Event{Type: EventToolCall, ToolCall: &call})
		}
	case "config_option_update":
		s.mergeOptions(update.ConfigOptions)
	}
}

func (r *Runner) Permission(ctx context.Context, request acp.PermissionRequest) string {
	allow, reject := choose(request.Options)
	s := r.byHarnessID(request.SessionID)
	if s == nil || allow == "" {
		return reject
	}

	s.mu.Lock()
	t := s.turn
	if t == nil {
		s.mu.Unlock()
		return reject
	}
	approval := Approval{
		ApprovalID: rand.Text(),
		CallID:     request.ToolCall.ToolCallID,
		Title:      request.ToolCall.Title,
		Status:     ApprovalPending,
	}
	if approval.Title == "" {
		if state, ok := s.tools[approval.CallID]; ok {
			approval.Title = state.call.Title
		}
	}
	decisions := make(chan decision, 1)
	s.approvals[approval.ApprovalID] = decisions
	r.flush(s)
	asked := approval
	r.record(s, Event{Type: EventApproval, Approval: &asked})
	s.mu.Unlock()

	var answer decision
	select {
	case answer = <-decisions:
	case <-time.After(r.approvalTimeout):
		answer = decision{outcome: "timeout"}
	case <-t.ctx.Done():
		answer = decision{outcome: "cancelled"}
	case <-ctx.Done():
		answer = decision{outcome: "cancelled"}
	}

	approval.Status = ApprovalDenied
	if answer.outcome == DecisionAllowOnce {
		approval.Status = ApprovalAllowed
	}
	approval.DecidedBy = answer.decidedBy
	s.mu.Lock()
	if state, ok := s.tools[approval.CallID]; ok && approval.Status == ApprovalDenied {
		state.denied = true
	}
	r.record(s, Event{Type: EventApproval, Approval: &approval})
	delete(s.approvals, approval.ApprovalID)
	s.mu.Unlock()
	if answer.recorded != nil {
		close(answer.recorded)
	}
	slog.Info("tool decision", "session", s.record.ID, "tool", approval.Title, "decision", answer.outcome, "decided_by", answer.decidedBy)

	switch answer.outcome {
	case DecisionAllowOnce:
		return allow
	case "cancelled":
		return ""
	default:
		return reject
	}
}

func (r *Runner) prompt(s *session, client *acp.Client, message, author string) error {
	s.mu.Lock()
	if s.turn != nil {
		s.mu.Unlock()
		return ErrBusy
	}
	ctx, cancel := context.WithCancel(r.ctx)
	t := &turn{ctx: ctx, cancel: cancel}
	s.turn = t
	harnessSessionID := s.record.HarnessSessionID
	r.flush(s)
	r.record(s, Event{Type: EventUserMessage, Text: truncate(message, maxMessage), Author: author})
	s.mu.Unlock()

	go func() {
		stopReason, err := client.Prompt(t.ctx, harnessSessionID, message)
		cancel()
		end := turnEnd(stopReason, err)
		if err != nil {
			slog.Warn("turn failed", "session", s.record.ID, "error", err)
		}
		s.mu.Lock()
		r.flush(s)
		s.turn = nil
		r.record(s, Event{Type: EventTurnEnd, TurnEnd: &end})
		s.mu.Unlock()
	}()
	return nil
}

func (r *Runner) route(ctx context.Context, model Model) (*acp.Client, error) {
	client, err := r.waitForClient(ctx)
	if err != nil {
		return nil, err
	}
	r.mu.Lock()
	routing := r.routing
	r.mu.Unlock()
	if r.harness.Covers(routing, model) {
		return client, nil
	}
	if r.busy() {
		return nil, fmt.Errorf("%w: the harness restarts for model provider %s once no session is working", ErrBusy, model.Provider)
	}

	routing = r.harness.Route(routing, model)
	if err := r.store.saveRouting(routing); err != nil {
		return nil, err
	}
	slog.Info("restarting harness for a model provider", "harness", r.harnessName, "provider", model.Provider, "model", model.Name)
	r.mu.Lock()
	r.routing = routing
	stale := r.client
	if stale != nil {
		r.client = nil
		r.ready = make(chan struct{})
	}
	r.mu.Unlock()
	if stale != nil {
		stale.Close()
	}
	return r.waitForClient(ctx)
}

func (r *Runner) attach(ctx context.Context, s *session) (*acp.Client, error) {
	s.mu.Lock()
	record := s.record
	s.mu.Unlock()
	client, err := r.route(ctx, record.Model)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	if s.attached == client {
		s.mu.Unlock()
		return client, nil
	}
	s.replaying = true
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(r.ctx, 2*time.Minute)
	defer cancel()
	meta := r.harness.SessionMeta(record.Model)
	resumed, err := client.ResumeSession(ctx, record.HarnessSessionID, r.workdir, r.harness.MCPServers(), meta)

	s.mu.Lock()
	s.replaying = false
	s.message.Reset()
	s.mu.Unlock()

	if err != nil {
		if errors.Is(err, acp.ErrExited) || ctx.Err() != nil {
			return nil, harnessError(err)
		}
		slog.Warn("could not resume session, starting a new one", "session", record.ID, "error", err)
		resumed, err = client.NewSession(ctx, r.workdir, r.harness.MCPServers(), meta)
		if err != nil {
			return nil, harnessError(err)
		}
		r.rekey(s, resumed.ID)
		s.mu.Lock()
		record = s.record
		s.mu.Unlock()
		if err := r.store.save(record); err != nil {
			return nil, err
		}
	}

	options, err := r.apply(ctx, client, resumed.ID, resumed.ConfigOptions, record.Model, record.Effort)
	if err != nil {
		slog.Warn("could not restore model and effort", "session", record.ID, "error", err)
	}
	s.mu.Lock()
	s.attached = client
	s.options = options
	s.mu.Unlock()
	slog.Info("session resumed", "session", record.ID)
	return client, nil
}

func (r *Runner) apply(ctx context.Context, client *acp.Client, harnessSessionID string, options []acp.ConfigOption, model Model, effort string) ([]acp.ConfigOption, error) {
	set := func(option acp.ConfigOption, value string) error {
		if option.Current() == value {
			return nil
		}
		updated, err := client.SetConfigOption(ctx, harnessSessionID, option.ID, value)
		if err != nil {
			return harnessError(fmt.Errorf("set %s to %s: %w", option.ID, value, err))
		}
		if len(updated) > 0 {
			options = updated
		}
		return nil
	}

	value, err := r.harness.ModelValue(model)
	if err != nil {
		return options, InvalidError{err}
	}
	option, ok := acp.FindConfigOption(options, "model")
	if !ok {
		return options, InvalidError{errors.New("the harness offers no model choice")}
	}
	if err := set(option, value); err != nil {
		return options, err
	}

	if effort == "" {
		return options, nil
	}
	option, ok = acp.FindConfigOption(options, "thought_level")
	if !ok {
		slog.Info("the model has no effort levels", "model", model.Name, "effort", effort)
		return options, nil
	}
	level, ok := effortValue(option, effort)
	if !ok {
		slog.Warn("the model doesn't offer this effort", "model", model.Name, "effort", effort)
		return options, nil
	}
	return options, set(option, level)
}

func (r *Runner) flush(s *session) {
	if strings.TrimSpace(s.message.String()) == "" {
		s.message.Reset()
		return
	}
	text := s.message.String()
	s.message.Reset()
	r.record(s, Event{Type: EventMessage, Text: text})
}

func (r *Runner) record(s *session, event Event) {
	s.nextID++
	event.ID = s.nextID
	event.Time = time.Now().UTC()
	s.events = append(s.events, event)
	if err := r.store.append(s.record.ID, event); err != nil {
		slog.Error("failed to write event", "session", s.record.ID, "error", err)
	}
	for waiter := range s.waiters {
		select {
		case waiter <- struct{}{}:
		default:
		}
	}
}

func (r *Runner) session(id string) (*session, error) {
	r.mu.Lock()
	s, ok := r.sessions[id]
	r.mu.Unlock()
	if ok {
		return s, nil
	}

	record, ok := r.store.session(id)
	if !ok {
		return nil, ErrNotFound
	}
	events, err := r.store.events(id)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if existing, ok := r.sessions[id]; ok {
		return existing, nil
	}
	s = newSession(record, events)
	r.sessions[id] = s
	r.byHarness[record.HarnessSessionID] = s
	return s, nil
}

func (r *Runner) loaded() []*session {
	sessions := make([]*session, 0, len(r.sessions))
	for _, s := range r.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

func (r *Runner) busy() bool {
	r.mu.Lock()
	sessions := r.loaded()
	r.mu.Unlock()
	return slices.ContainsFunc(sessions, func(s *session) bool { return s.status() != StatusIdle })
}

func (r *Runner) register(s *session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.record.ID] = s
	r.byHarness[s.record.HarnessSessionID] = s
}

func (r *Runner) unregister(s *session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, s.record.ID)
	delete(r.byHarness, s.record.HarnessSessionID)
}

func (r *Runner) rekey(s *session, harnessSessionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(r.byHarness, s.record.HarnessSessionID)
	s.record.HarnessSessionID = harnessSessionID
	r.byHarness[harnessSessionID] = s
}

func (r *Runner) byHarnessID(harnessSessionID string) *session {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.byHarness[harnessSessionID]
}

func (r *Runner) waitForClient(ctx context.Context) (*acp.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	for {
		r.mu.Lock()
		client, ready := r.client, r.ready
		r.mu.Unlock()
		if client != nil {
			return client, nil
		}
		select {
		case <-ready:
		case <-ctx.Done():
			return nil, ErrUnavailable
		case <-r.ctx.Done():
			return nil, ErrUnavailable
		}
	}
}

func (r *Runner) fail(err error) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.failure = err
	r.failures++
	return r.failures
}

func turnEnd(stopReason string, err error) TurnEnd {
	switch {
	case err != nil:
		return TurnEnd{Reason: TurnFailed, Error: truncate(err.Error(), 2048)}
	case stopReason == "end_turn":
		return TurnEnd{Reason: TurnCompleted}
	case stopReason == "cancelled":
		return TurnEnd{Reason: TurnStopped}
	default:
		return TurnEnd{Reason: TurnFailed, Error: "the agent stopped: " + stopReason}
	}
}

func validate(message, effort string) error {
	if strings.TrimSpace(message) == "" {
		return InvalidError{errors.New("message is empty")}
	}
	if effort != "" && !slices.Contains(efforts, effort) {
		return InvalidError{fmt.Errorf("effort must be one of %s", strings.Join(efforts, ", "))}
	}
	return nil
}

func harnessError(err error) error {
	if errors.Is(err, acp.ErrExited) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	var invalid InvalidError
	if errors.As(err, &invalid) {
		return err
	}
	return InvalidError{err}
}

func effortValue(option acp.ConfigOption, effort string) (string, bool) {
	values := option.Values()
	for _, value := range values {
		if strings.EqualFold(value.Value, effort) {
			return value.Value, true
		}
	}
	if effort == "max" {
		for _, value := range slices.Backward(values) {
			if value.Value != "default" && value.Value != "none" {
				return value.Value, true
			}
		}
	}
	return "", false
}

func choose(options []acp.PermissionOption) (string, string) {
	var allow, reject, rejectAlways string
	for _, option := range options {
		switch {
		case option.Kind == "allow_once" && allow == "":
			allow = option.OptionID
		case option.Kind == "reject_once" && reject == "":
			reject = option.OptionID
		case option.Kind == "reject_always" && rejectAlways == "":
			rejectAlways = option.OptionID
		}
	}
	if reject == "" {
		reject = rejectAlways
	}
	return allow, reject
}

func title(message string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(message), "\n")
	line = strings.TrimSpace(line)
	if utf8.RuneCountInString(line) <= 80 {
		return line
	}
	return string([]rune(line)[:80]) + "…"
}

func backoff(failures int) time.Duration {
	return min(time.Duration(failures)*5*time.Second, time.Minute)
}

func sleep(ctx context.Context, duration time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(duration):
	}
}
