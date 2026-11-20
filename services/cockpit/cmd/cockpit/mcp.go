package main

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/zeitlos/lucity/services/cockpit/internal/api/mcpserver"
	"github.com/zeitlos/lucity/services/cockpit/internal/cockpit"
)

func serveMCP(ctx context.Context, deployments string) error {
	var mu sync.Mutex
	clients := map[string]*cockpit.Client{}

	return mcpserver.Serve(ctx, func(environment string) (*cockpit.Client, error) {
		if !slices.Contains(knownEnvironments, environment) {
			return nil, fmt.Errorf("environment must be dev or prod, not %q", environment)
		}

		mu.Lock()
		defer mu.Unlock()

		if client, ok := clients[environment]; ok {
			return client, nil
		}

		client, err := connect(ctx, environment, deployments)

		if err != nil {
			return nil, err
		}

		clients[environment] = client

		return client, nil
	})
}
