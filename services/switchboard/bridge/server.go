package bridge

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"html"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/zeitlos/lucity/services/switchboard/acp"
	"github.com/zeitlos/lucity/services/switchboard/telegram"
)

type Agent struct {
	Command    []string
	Workdir    string
	MCPServers []json.RawMessage
}

type Forum struct {
	ChatID        int64
	AlertsTopicID int64
}

type Option func(*Server)

func WithAllowedUsers(userIDs ...int64) Option {
	return func(s *Server) {
		for _, userID := range userIDs {
			s.allowed[userID] = true
		}
	}
}

func WithApprovalTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.approvalTimeout = timeout
	}
}

type Server struct {
	telegram        *telegram.Client
	forum           Forum
	threads         *Threads
	agentConfig     Agent
	allowed         map[int64]bool
	approvalTimeout time.Duration
	botID           int64

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}

	sessionMu sync.Mutex
	mu        sync.Mutex
	agent     *acp.Client
	ready     chan struct{}
	attached  map[string]bool
	turns     map[string]*turn
	approvals map[string]chan decision
}

func New(client *telegram.Client, forum Forum, threads *Threads, agent Agent, options ...Option) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Server{
		telegram:        client,
		forum:           forum,
		threads:         threads,
		agentConfig:     agent,
		allowed:         map[int64]bool{},
		approvalTimeout: 15 * time.Minute,
		ctx:             ctx,
		cancel:          cancel,
		done:            make(chan struct{}),
		ready:           make(chan struct{}),
		attached:        map[string]bool{},
		turns:           map[string]*turn{},
		approvals:       map[string]chan decision{},
	}
	for _, option := range options {
		option(s)
	}
	return s
}

func (s *Server) Label() string {
	return "switchboard"
}

func (s *Server) Start() error {
	defer close(s.done)
	if !s.identify() {
		return nil
	}
	go s.superviseAgent()

	var offset int64
	for s.ctx.Err() == nil {
		updates, err := s.telegram.Updates(s.ctx, offset)
		if err != nil {
			if s.ctx.Err() != nil {
				break
			}
			slog.Warn("telegram poll failed", "error", err)
			sleep(s.ctx, backoff(err))
			continue
		}
		for _, update := range updates {
			offset = update.UpdateID + 1
			s.handle(update)
		}
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.cancel()
	select {
	case <-s.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Server) identify() bool {
	for s.ctx.Err() == nil {
		me, err := s.telegram.Me(s.ctx)
		if err == nil {
			s.botID = me.ID
			slog.Info("telegram bot ready", "bot", me.Username, "chat", s.forum.ChatID, "alerts_topic", s.forum.AlertsTopicID)
			return true
		}
		slog.Warn("telegram getMe failed", "error", err)
		sleep(s.ctx, backoff(err))
	}
	return false
}

func (s *Server) superviseAgent() {
	for s.ctx.Err() == nil {
		agent, err := acp.Start(s.ctx, acp.Command{
			Path: s.agentConfig.Command[0],
			Args: s.agentConfig.Command[1:],
			Dir:  s.agentConfig.Workdir,
			Env:  agentEnvironment(),
		}, s)
		if err != nil {
			slog.Error("agent failed to start", "error", err)
			sleep(s.ctx, 10*time.Second)
			continue
		}

		s.mu.Lock()
		s.agent = agent
		s.attached = map[string]bool{}
		close(s.ready)
		s.mu.Unlock()
		slog.Info("agent started", "command", s.agentConfig.Command)

		<-agent.Done()

		s.mu.Lock()
		s.agent = nil
		s.ready = make(chan struct{})
		s.mu.Unlock()
		if s.ctx.Err() == nil {
			slog.Error("agent exited, restarting", "error", agent.Err())
			sleep(s.ctx, 5*time.Second)
		}
	}
}

func (s *Server) handle(update telegram.Update) {
	switch {
	case update.CallbackQuery != nil:
		s.handleCallback(update.CallbackQuery)
	case update.Message != nil && update.Message.Text != "":
		s.handleMessage(update.Message)
	}
}

func (s *Server) handleMessage(message *telegram.Message) {
	if message.Chat.ID != s.forum.ChatID {
		slog.Warn("ignored message from another chat", "chat", message.Chat.ID)
		return
	}
	if message.From == nil || !s.allowed[message.From.ID] {
		slog.Debug("ignored message from a sender not on the allowlist", "sender", senderID(message), "thread", message.MessageThreadID)
		return
	}
	text := strings.TrimSpace(message.Text)

	if s.inSessionTopic(message) {
		go s.converse(message, message.MessageThreadID, text, "", topicOf(message))
		return
	}
	if question, ok := command(text, "/new"); ok {
		if question != "" {
			go s.open(message, nil, question)
		}
		return
	}
	if alert := s.repliedAlert(message); alert != nil {
		go s.open(message, alert, text)
		return
	}
	slog.Debug("ignored message outside a session topic", "thread", message.MessageThreadID)
}

func (s *Server) inSessionTopic(message *telegram.Message) bool {
	return message.IsTopicMessage && message.MessageThreadID != 0 && message.MessageThreadID != s.forum.AlertsTopicID
}

func (s *Server) repliedAlert(message *telegram.Message) *telegram.Message {
	reply := message.ReplyToMessage
	if reply == nil || reply.From == nil || reply.From.ID != s.botID {
		return nil
	}
	if message.IsTopicMessage && reply.MessageID == message.MessageThreadID {
		return nil
	}
	return reply
}

func (s *Server) open(message *telegram.Message, alert *telegram.Message, text string) {
	chatID := message.Chat.ID
	origin := originThread(message)
	name := questionTopicName(text)
	quoted := ""
	if alert != nil {
		name = alertTopicName(alert)
		quoted = alert.Text
	}

	threadID, err := s.telegram.CreateTopic(s.ctx, chatID, name)
	if err != nil {
		slog.Error("failed to create topic", "error", err)
		s.say(chatID, origin, message.MessageID, "Could not open a topic: "+err.Error()+". The bot has to be an admin with the Manage Topics right.")
		return
	}
	if alert != nil {
		if _, err := s.telegram.Copy(s.ctx, chatID, threadID, alert); err != nil {
			slog.Warn("failed to copy the alert into its topic", "error", err)
		}
	}
	if _, err := s.telegram.Forward(s.ctx, chatID, threadID, message); err != nil {
		slog.Warn("failed to forward the question into its topic", "error", err)
	}

	pointer := telegram.Outgoing{
		ChatID:   chatID,
		ThreadID: origin,
		ReplyTo:  message.MessageID,
		HTML:     true,
		Text:     `Continuing in <a href="` + topicLink(chatID, threadID) + `">` + html.EscapeString(name) + `</a>`,
	}
	if _, err := s.telegram.Send(s.ctx, pointer); err != nil {
		slog.Warn("failed to post the topic link", "error", err)
	}

	s.converse(message, threadID, text, quoted, name)
}

func (s *Server) converse(message *telegram.Message, threadID int64, text, quoted, topic string) {
	chatID := message.Chat.ID
	thread := threadKey(chatID, threadID)

	agent := s.waitForAgent(2 * time.Minute)
	if agent == nil {
		s.say(chatID, threadID, 0, "The agent is not running. Try again in a minute.")
		return
	}

	sessionID, err := s.session(agent, thread)
	if err != nil {
		slog.Error("failed to open session", "thread", thread, "error", err)
		s.say(chatID, threadID, 0, "Could not start an agent session: "+err.Error())
		return
	}

	t, ok := s.begin(sessionID, chatID, threadID, thread)
	if !ok {
		s.say(chatID, threadID, 0, "Still working on the previous message in this topic. Tap Stop there to cancel it.")
		return
	}

	statusID, err := s.telegram.Send(s.ctx, telegram.Outgoing{
		ChatID:   chatID,
		ThreadID: threadID,
		Text:     "Working…",
		Buttons:  stopButton(sessionID),
	})
	if err == nil {
		t.statusID = statusID
	}
	progressDone := make(chan struct{})
	go s.showProgress(t, progressDone)

	slog.Info("turn started", "event", "turn.start", "session", sessionID, "thread", thread, "actor", actorID(message.From))
	stopReason, err := agent.Prompt(t.ctx, sessionID, prompt(message, topic, text, quoted))
	s.end(t)
	<-progressDone

	answer := t.answer()
	if answer == "" && err == nil && stopReason != "cancelled" {
		answer = "The agent finished without an answer."
	}
	delivered := true
	for _, part := range chunks(answer, 3500) {
		if _, sendErr := s.telegram.Send(s.ctx, telegram.Outgoing{ChatID: chatID, ThreadID: threadID, Text: part}); sendErr != nil {
			slog.Error("failed to send answer", "session", sessionID, "error", sendErr)
			delivered = false
			break
		}
	}

	if t.statusID != 0 {
		if err == nil && stopReason == "end_turn" && delivered {
			if deleteErr := s.telegram.Delete(s.ctx, chatID, t.statusID); deleteErr != nil {
				slog.Warn("failed to delete status message", "error", deleteErr)
			}
		} else {
			status := telegram.Outgoing{ChatID: chatID, Text: finalStatus(stopReason, err, t.stepCount())}
			if editErr := s.telegram.Edit(s.ctx, t.statusID, status); editErr != nil {
				slog.Warn("failed to update status message", "error", editErr)
			}
		}
	}
	slog.Info("turn finished", "event", "turn.end", "session", sessionID, "thread", thread,
		"stop_reason", stopReason, "steps", t.stepCount(), "error", errorText(err))
}

func (s *Server) session(agent *acp.Client, thread string) (string, error) {
	s.sessionMu.Lock()
	defer s.sessionMu.Unlock()

	ctx, cancel := context.WithTimeout(s.ctx, 2*time.Minute)
	defer cancel()

	sessionID, known := s.threads.Session(thread)
	if known && s.isAttached(sessionID) {
		return sessionID, nil
	}
	if known {
		err := agent.ResumeSession(ctx, sessionID, s.agentConfig.Workdir, s.agentConfig.MCPServers)
		if err == nil {
			s.markAttached(sessionID)
			return sessionID, nil
		}
		slog.Warn("could not resume session, starting a new one", "session", sessionID, "error", err)
	}

	sessionID, err := agent.NewSession(ctx, s.agentConfig.Workdir, s.agentConfig.MCPServers)
	if err != nil {
		return "", err
	}
	s.markAttached(sessionID)
	return sessionID, s.threads.SetSession(thread, sessionID)
}

func (s *Server) Update(sessionID string, update acp.Update) {
	if update.Kind == "tool_call" {
		slog.Info("tool call", "event", "tool.call", "session", sessionID, "tool", update.Title,
			"kind", update.ToolKind, "input", truncate(string(update.RawInput), 4000))
	}
	if t := s.turn(sessionID); t != nil {
		t.record(update)
	}
}

func (s *Server) Permission(ctx context.Context, request acp.PermissionRequest) string {
	allow, reject := choose(request.Options)
	t := s.turn(request.SessionID)
	if t == nil || allow == "" {
		return reject
	}

	key := randomKey()
	decisions := make(chan decision, 1)
	s.mu.Lock()
	s.approvals[key] = decisions
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.approvals, key)
		s.mu.Unlock()
	}()

	command := literal(request.ToolCall)
	card := telegram.Outgoing{
		ChatID:   t.chatID,
		ThreadID: t.threadID,
		HTML:     true,
		Text:     approvalCard(request.ToolCall, command, ""),
		Buttons:  []telegram.Button{{Text: "Run", CallbackData: "run:" + key}, {Text: "Deny", CallbackData: "deny:" + key}},
	}
	cardID, err := s.telegram.Send(s.ctx, card)
	if err != nil {
		slog.Error("failed to send approval card", "session", request.SessionID, "error", err)
		return reject
	}

	asked := time.Now()
	var answer decision
	select {
	case answer = <-decisions:
	case <-time.After(s.approvalTimeout):
		answer = decision{outcome: "timeout"}
	case <-t.ctx.Done():
		answer = decision{outcome: "cancelled"}
	case <-ctx.Done():
		answer = decision{outcome: "cancelled"}
	}

	card.Buttons = nil
	card.Text = approvalCard(request.ToolCall, command, answer.summary(s.approvalTimeout))
	if err := s.telegram.Edit(s.ctx, cardID, card); err != nil {
		slog.Warn("failed to update approval card", "error", err)
	}
	slog.Info("tool decision", "event", "tool.decision", "session", request.SessionID, "thread", t.thread,
		"actor", answer.actorID, "tool", request.ToolCall.Title, "input", command,
		"decision", answer.outcome, "waited_ms", time.Since(asked).Milliseconds())

	switch answer.outcome {
	case "allow":
		return allow
	case "cancelled":
		return ""
	default:
		return reject
	}
}

func (s *Server) handleCallback(callback *telegram.CallbackQuery) {
	if !s.allowed[callback.From.ID] {
		slog.Warn("ignored button press from a user not on the allowlist", "sender", callback.From.ID)
		s.answer(callback.ID, "Not allowed")
		return
	}

	action, key, _ := strings.Cut(callback.Data, ":")
	switch action {
	case "run", "deny":
		s.mu.Lock()
		decisions, ok := s.approvals[key]
		s.mu.Unlock()
		if !ok {
			s.answer(callback.ID, "This request has expired")
			return
		}
		outcome := "allow"
		if action == "deny" {
			outcome = "deny"
		}
		select {
		case decisions <- decision{outcome: outcome, actor: actor(&callback.From), actorID: actorID(&callback.From), at: time.Now()}:
		default:
		}
		s.answer(callback.ID, "")
	case "stop":
		if t := s.turn(key); t != nil {
			t.cancel()
		}
		s.answer(callback.ID, "Stopping")
	}
}

func (s *Server) showProgress(t *turn, done chan<- struct{}) {
	defer close(done)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-t.finished:
			return
		case <-ticker.C:
			if err := s.telegram.Typing(s.ctx, t.chatID, t.threadID); err != nil {
				slog.Debug("failed to send typing action", "error", err)
			}
			text, changed := t.progress()
			if !changed || t.statusID == 0 {
				continue
			}
			status := telegram.Outgoing{ChatID: t.chatID, Text: text, Buttons: stopButton(t.sessionID)}
			if err := s.telegram.Edit(s.ctx, t.statusID, status); err != nil {
				slog.Debug("failed to update status message", "error", err)
			}
		}
	}
}

func (s *Server) begin(sessionID string, chatID, threadID int64, thread string) (*turn, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, busy := s.turns[sessionID]; busy {
		return nil, false
	}
	ctx, cancel := context.WithCancel(s.ctx)
	t := &turn{
		ctx:       ctx,
		cancel:    cancel,
		finished:  make(chan struct{}),
		sessionID: sessionID,
		chatID:    chatID,
		threadID:  threadID,
		thread:    thread,
	}
	s.turns[sessionID] = t
	return t, true
}

func (s *Server) end(t *turn) {
	s.mu.Lock()
	delete(s.turns, t.sessionID)
	s.mu.Unlock()
	t.cancel()
	close(t.finished)
}

func (s *Server) turn(sessionID string) *turn {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.turns[sessionID]
}

func (s *Server) waitForAgent(timeout time.Duration) *acp.Client {
	deadline := time.After(timeout)
	for {
		s.mu.Lock()
		agent, ready := s.agent, s.ready
		s.mu.Unlock()
		if agent != nil {
			return agent
		}
		select {
		case <-ready:
		case <-deadline:
			return nil
		case <-s.ctx.Done():
			return nil
		}
	}
}

func (s *Server) isAttached(sessionID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.attached[sessionID]
}

func (s *Server) markAttached(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attached[sessionID] = true
}

func (s *Server) say(chatID, threadID, replyTo int64, text string) {
	if _, err := s.telegram.Send(s.ctx, telegram.Outgoing{ChatID: chatID, ThreadID: threadID, ReplyTo: replyTo, Text: text}); err != nil {
		slog.Error("failed to send message", "error", err)
	}
}

func (s *Server) answer(callbackID, text string) {
	if err := s.telegram.Answer(s.ctx, callbackID, text); err != nil {
		slog.Debug("failed to answer button press", "error", err)
	}
}

type decision struct {
	outcome string
	actor   string
	actorID string
	at      time.Time
}

func (d decision) summary(timeout time.Duration) string {
	switch d.outcome {
	case "allow":
		return "Approved by " + d.actor + " at " + d.at.Format("15:04")
	case "deny":
		return "Denied by " + d.actor + " at " + d.at.Format("15:04")
	case "timeout":
		return "No answer within " + timeout.String() + ", denied"
	default:
		return "Cancelled"
	}
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

func command(text, name string) (string, bool) {
	rest, ok := strings.CutPrefix(text, name)
	if !ok {
		return "", false
	}
	if strings.HasPrefix(rest, "@") {
		_, rest, _ = strings.Cut(rest, " ")
	} else if rest != "" && !strings.HasPrefix(rest, " ") && !strings.HasPrefix(rest, "\n") {
		return "", false
	}
	return strings.TrimSpace(rest), true
}

func backoff(err error) time.Duration {
	var apiErr *telegram.Error
	if errors.As(err, &apiErr) && apiErr.RetryAfter > 0 {
		return apiErr.RetryAfter
	}
	return 5 * time.Second
}

func sleep(ctx context.Context, duration time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(duration):
	}
}

func randomKey() string {
	key := make([]byte, 8)
	_, _ = rand.Read(key)
	return hex.EncodeToString(key)
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
