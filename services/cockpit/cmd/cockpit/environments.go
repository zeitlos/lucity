package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

const environmentHeader = "X-Cockpit-Environment"

var knownEnvironments = []string{"dev", "prod"}

type environments struct {
	ctx         context.Context
	deployments string
	statePath   string

	mu       sync.Mutex
	current  string
	saved    string
	handlers map[string]http.Handler
}

func newEnvironments(ctx context.Context, deployments, statePath, initial string) *environments {
	saved := savedEnvironment(statePath)

	if initial == "" {
		initial = saved
	}

	return &environments{
		ctx:         ctx,
		deployments: deployments,
		statePath:   statePath,
		current:     initial,
		saved:       saved,
		handlers:    map[string]http.Handler{},
	}
}

func savedEnvironment(statePath string) string {
	data, err := os.ReadFile(statePath)
	environment := strings.TrimSpace(string(data))

	if err != nil || !slices.Contains(knownEnvironments, environment) {
		return "dev"
	}

	return environment
}

func (e *environments) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	environment := r.Header.Get(environmentHeader)

	if environment == "" {
		environment = e.currentEnvironment()
	}

	if !slices.Contains(knownEnvironments, environment) {
		writeError(w, "unknown environment "+environment)
		return
	}

	handler, err := e.handler(environment)

	if err != nil {
		writeError(w, err.Error())
		return
	}

	e.remember(environment)
	handler.ServeHTTP(w, r)
}

func (e *environments) currentEnvironment() string {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.current
}

func (e *environments) remember(environment string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.current = environment

	if environment == e.saved {
		return
	}

	if err := os.MkdirAll(filepath.Dir(e.statePath), 0o700); err != nil {
		slog.Warn("remember environment", "error", err)
		return
	}

	if err := os.WriteFile(e.statePath, []byte(environment), 0o600); err != nil {
		slog.Warn("remember environment", "error", err)
		return
	}

	e.saved = environment
}

func (e *environments) handler(environment string) (http.Handler, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if handler, ok := e.handlers[environment]; ok {
		return handler, nil
	}

	client, err := connect(e.ctx, environment, e.deployments)

	if err != nil {
		return nil, err
	}

	handler := graphQLHandler(client)
	e.handlers[environment] = handler

	return handler, nil
}

func writeError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"errors": []map[string]string{{"message": message}},
	})
}
