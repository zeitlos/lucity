package conductor

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeitlos/lucity/pkg/tenant"
	"github.com/zeitlos/lucity/services/conductor/internal/data"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type Variable struct {
	Key   string
	Value string
}

type DatabaseRef struct {
	Database platform.DatabaseID
	Key      string
}

type ServiceRef struct {
	Service platform.ServiceID
}

type ServiceVariable struct {
	Key         string
	Value       string
	FromShared  bool
	DatabaseRef *DatabaseRef
	ServiceRef  *ServiceRef
}

// cnpgKeyDisplayNames maps CNPG secret keys to human-readable display names.
var cnpgKeyDisplayNames = map[string]string{
	"uri":      "DATABASE_URL",
	"host":     "PGHOST",
	"port":     "PGPORT",
	"dbname":   "PGDATABASE",
	"user":     "PGUSER",
	"password": "PGPASSWORD",
}

func (c *Client) SharedVariables(ctx context.Context, environment platform.EnvironmentID) ([]Variable, error) {
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	vars, err := c.Packager.SharedVariables(callCtx, ws, environment.Project, environment.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared variables: %w", err)
	}

	var result []Variable
	for k, v := range vars {
		result = append(result, Variable{Key: k, Value: v})
	}
	return result, nil
}

func (c *Client) SetSharedVariables(ctx context.Context, environment platform.EnvironmentID, vars []Variable) (bool, error) {
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return false, err
	}

	m := make(map[string]string, len(vars))
	for _, v := range vars {
		m[v.Key] = v.Value
	}

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	if err := c.Packager.SetSharedVariables(callCtx, ws, environment.Project, environment.Name, m); err != nil {
		return false, fmt.Errorf("failed to set shared variables: %w", err)
	}
	return true, nil
}

func (c *Client) ServiceVariables(ctx context.Context, service platform.ServiceID) ([]ServiceVariable, error) {
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.Packager.ServiceVariables(callCtx, ws, service.Project, service.Environment, service.Name)
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

	for k, ref := range resp.DatabaseRefs {
		displayKey := cnpgKeyDisplayNames[ref.Key]
		if displayKey == "" {
			displayKey = ref.Key
		}
		dbName := strings.ToUpper(ref.Database[:1]) + ref.Database[1:]
		// TODO: rebuild DatabaseRef with typed DatabaseID after migration
		result = append(result, ServiceVariable{
			Key:   k,
			Value: fmt.Sprintf("${{%s.%s}}", dbName, displayKey),
		})
	}

	for k, ref := range resp.ServiceRefs {
		svcName := strings.ToUpper(ref.Service[:1]) + ref.Service[1:]
		// TODO: rebuild ServiceRef with typed ServiceID after migration
		result = append(result, ServiceVariable{
			Key:   k,
			Value: fmt.Sprintf("${{%s.URL}}", svcName),
		})
	}

	return result, nil
}

func (c *Client) SetServiceVariables(ctx context.Context, service platform.ServiceID, vars []Variable, sharedRefs []string, dbRefs map[string]DatabaseRef, svcRefs map[string]ServiceRef) (bool, error) {
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return false, err
	}

	m := make(map[string]string, len(vars))
	for _, v := range vars {
		m[v.Key] = v.Value
	}

	dataDBRefs := make(map[string]data.DatabaseRef, len(dbRefs))
	for k, ref := range dbRefs {
		dataDBRefs[k] = data.DatabaseRef{Database: ref.Database.Name, Key: ref.Key}
	}
	dataSvcRefs := make(map[string]data.ServiceRef, len(svcRefs))
	for k, ref := range svcRefs {
		dataSvcRefs[k] = data.ServiceRef{Service: ref.Service.Name}
	}

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	if err := c.Packager.SetServiceVariables(callCtx, ws, service.Project, service.Environment, service.Name, m, sharedRefs, dataDBRefs, dataSvcRefs); err != nil {
		return false, fmt.Errorf("failed to set service variables: %w", err)
	}
	return true, nil
}
