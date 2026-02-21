package handler

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/packager"
)

type Variable struct {
	Key   string
	Value string
}

type ServiceVariable struct {
	Key        string
	Value      string
	FromShared bool
}

func (c *Client) SharedVariables(ctx context.Context, projectID, environment string) ([]Variable, error) {
	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.Packager.SharedVariables(callCtx, &packager.SharedVariablesRequest{
		Project:     projectID,
		Environment: environment,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get shared variables: %w", err)
	}

	var result []Variable
	for k, v := range resp.Variables {
		result = append(result, Variable{Key: k, Value: v})
	}
	return result, nil
}

func (c *Client) SetSharedVariables(ctx context.Context, projectID, environment string, vars []Variable) (bool, error) {
	ctx = auth.OutgoingContext(ctx)

	m := make(map[string]string, len(vars))
	for _, v := range vars {
		m[v.Key] = v.Value
	}

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	_, err := c.Packager.SetSharedVariables(callCtx, &packager.SetSharedVariablesRequest{
		Project:     projectID,
		Environment: environment,
		Variables:   m,
	})
	if err != nil {
		return false, fmt.Errorf("failed to set shared variables: %w", err)
	}
	return true, nil
}

func (c *Client) ServiceVariables(ctx context.Context, projectID, environment, service string) ([]ServiceVariable, error) {
	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.Packager.ServiceVariables(callCtx, &packager.ServiceVariablesRequest{
		Project:     projectID,
		Environment: environment,
		Service:     service,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get service variables: %w", err)
	}

	refSet := make(map[string]bool, len(resp.SharedRefs))
	for _, ref := range resp.SharedRefs {
		refSet[ref] = true
	}

	var result []ServiceVariable
	for k, v := range resp.Variables {
		result = append(result, ServiceVariable{
			Key:        k,
			Value:      v,
			FromShared: refSet[k],
		})
	}
	return result, nil
}

func (c *Client) SetServiceVariables(ctx context.Context, projectID, environment, service string, vars []Variable, sharedRefs []string) (bool, error) {
	ctx = auth.OutgoingContext(ctx)

	m := make(map[string]string, len(vars))
	for _, v := range vars {
		m[v.Key] = v.Value
	}

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	_, err := c.Packager.SetServiceVariables(callCtx, &packager.SetServiceVariablesRequest{
		Project:     projectID,
		Environment: environment,
		Service:     service,
		Variables:   m,
		SharedRefs:  sharedRefs,
	})
	if err != nil {
		return false, fmt.Errorf("failed to set service variables: %w", err)
	}
	return true, nil
}
