package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/zeitlos/lucity/services/agent/internal/runner"
)

const maxBody = 1 << 20

type Server struct {
	runner *runner.Runner
	server *http.Server
	ctx    context.Context
	cancel context.CancelFunc
}

func New(address string, r *runner.Runner) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Server{runner: r, ctx: ctx, cancel: cancel}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /status", s.status)
	mux.HandleFunc("GET /sessions", s.sessions)
	mux.HandleFunc("POST /sessions", s.create)
	mux.HandleFunc("GET /sessions/{id}", s.session)
	mux.HandleFunc("GET /sessions/{id}/events", s.events)
	mux.HandleFunc("POST /sessions/{id}/messages", s.send)
	mux.HandleFunc("POST /sessions/{id}/model", s.model)
	mux.HandleFunc("POST /sessions/{id}/approvals/{approvalID}", s.answer)
	mux.HandleFunc("POST /sessions/{id}/stop", s.stop)

	s.server = &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return s
}

func (s *Server) Label() string {
	return "api"
}

func (s *Server) Start() error {
	slog.Info("listening", "address", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.cancel()
	return s.server.Shutdown(ctx)
}

func (s *Server) status(w http.ResponseWriter, _ *http.Request) {
	respond(w, http.StatusOK, s.runner.Status())
}

func (s *Server) sessions(w http.ResponseWriter, _ *http.Request) {
	respond(w, http.StatusOK, s.runner.Sessions())
}

func (s *Server) session(w http.ResponseWriter, r *http.Request) {
	session, err := s.runner.Session(r.PathValue("id"))
	if err != nil {
		fail(w, err)
		return
	}
	respond(w, http.StatusOK, session)
}

func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Message   string        `json:"message"`
		Model     *runner.Model `json:"model"`
		Effort    string        `json:"effort"`
		StartedBy string        `json:"startedBy"`
	}
	if !decode(w, r, &request) {
		return
	}
	session, err := s.runner.Create(r.Context(), request.Message, request.Model, request.Effort, request.StartedBy)
	if err != nil {
		fail(w, err)
		return
	}
	respond(w, http.StatusCreated, session)
}

func (s *Server) send(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Message string `json:"message"`
		Author  string `json:"author"`
	}
	if !decode(w, r, &request) {
		return
	}
	session, err := s.runner.Send(r.Context(), r.PathValue("id"), request.Message, request.Author)
	if err != nil {
		fail(w, err)
		return
	}
	respond(w, http.StatusAccepted, session)
}

func (s *Server) model(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Model  *runner.Model `json:"model"`
		Effort string        `json:"effort"`
	}
	if !decode(w, r, &request) {
		return
	}
	session, err := s.runner.SetModel(r.Context(), r.PathValue("id"), request.Model, request.Effort)
	if err != nil {
		fail(w, err)
		return
	}
	respond(w, http.StatusOK, session)
}

func (s *Server) answer(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Decision  string `json:"decision"`
		DecidedBy string `json:"decidedBy"`
	}
	if !decode(w, r, &request) {
		return
	}
	session, err := s.runner.Answer(r.PathValue("id"), r.PathValue("approvalID"), request.Decision, request.DecidedBy)
	if err != nil {
		fail(w, err)
		return
	}
	respond(w, http.StatusOK, session)
}

func (s *Server) stop(w http.ResponseWriter, r *http.Request) {
	session, err := s.runner.Stop(r.PathValue("id"))
	if err != nil {
		fail(w, err)
		return
	}
	respond(w, http.StatusOK, session)
}

func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	after, err := cursor(r)
	if err != nil {
		fail(w, runner.InvalidError{Err: err})
		return
	}
	waiter, unsubscribe, err := s.runner.Subscribe(id)
	if err != nil {
		fail(w, err)
		return
	}
	defer unsubscribe()

	controller := http.NewResponseController(w)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	keepalive := time.NewTicker(20 * time.Second)
	defer keepalive.Stop()
	for {
		events, err := s.runner.Events(id, after)
		if err != nil {
			return
		}
		for _, event := range events {
			payload, err := json.Marshal(event)
			if err != nil {
				return
			}
			if _, err := fmt.Fprintf(w, "id: %d\ndata: %s\n\n", event.ID, payload); err != nil {
				return
			}
			after = event.ID
		}
		if err := controller.Flush(); err != nil {
			return
		}

		select {
		case <-waiter:
		case <-keepalive.C:
			if _, err := fmt.Fprint(w, ": keepalive\n\n"); err != nil {
				return
			}
		case <-r.Context().Done():
			return
		case <-s.ctx.Done():
			return
		}
	}
}

func cursor(r *http.Request) (int64, error) {
	value := r.URL.Query().Get("after")
	if value == "" {
		value = r.Header.Get("Last-Event-ID")
	}
	if value == "" {
		return 0, nil
	}
	after, err := strconv.ParseInt(value, 10, 64)
	if err != nil || after < 0 {
		return 0, fmt.Errorf("after must be an event id")
	}
	return after, nil
}

func decode(w http.ResponseWriter, r *http.Request, value any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBody))
	if err := decoder.Decode(value); err != nil {
		fail(w, runner.InvalidError{Err: fmt.Errorf("invalid request body: %w", err)})
		return false
	}
	return true
}

func respond(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Debug("failed to write response", "error", err)
	}
}

func fail(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	var invalid runner.InvalidError
	switch {
	case errors.Is(err, runner.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, runner.ErrBusy), errors.Is(err, runner.ErrAnswered):
		status = http.StatusConflict
	case errors.Is(err, runner.ErrUnavailable):
		status = http.StatusServiceUnavailable
	case errors.As(err, &invalid):
		status = http.StatusUnprocessableEntity
	default:
		slog.Error("request failed", "error", err)
	}
	respond(w, status, map[string]string{"error": err.Error()})
}
