package main

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"sync"
)

const environmentHeader = "X-Cockpit-Environment"

var knownEnvironments = []string{"dev", "prod"}

type environments struct {
	ctx         context.Context
	deployments string
	initial     string

	mu       sync.Mutex
	handlers map[string]http.Handler
}

func newEnvironments(ctx context.Context, deployments, initial string) *environments {
	return &environments{
		ctx:         ctx,
		deployments: deployments,
		initial:     initial,
		handlers:    map[string]http.Handler{},
	}
}

func (e *environments) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	environment := r.Header.Get(environmentHeader)

	if environment == "" {
		environment = e.initial
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

	handler.ServeHTTP(w, r)
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
