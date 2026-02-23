package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/packager"
)

type Variable struct {
	Key   string
	Value string
}

type DatabaseRef struct {
	Database string
	Key      string
}

type ServiceRef struct {
	Service string
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

	for k, ref := range resp.DatabaseRefs {
		displayKey := cnpgKeyDisplayNames[ref.Key]
		if displayKey == "" {
			displayKey = ref.Key
		}
		dbName := strings.ToUpper(ref.Database[:1]) + ref.Database[1:]
		result = append(result, ServiceVariable{
			Key:         k,
			Value:       fmt.Sprintf("${{%s.%s}}", dbName, displayKey),
			DatabaseRef: &DatabaseRef{Database: ref.Database, Key: ref.Key},
		})
	}

	for k, ref := range resp.ServiceRefs {
		svcName := strings.ToUpper(ref.Service[:1]) + ref.Service[1:]
		result = append(result, ServiceVariable{
			Key:        k,
			Value:      fmt.Sprintf("${{%s.URL}}", svcName),
			ServiceRef: &ServiceRef{Service: ref.Service},
		})
	}

	return result, nil
}

func (c *Client) SetServiceVariables(ctx context.Context, projectID, environment, service string, vars []Variable, sharedRefs []string, dbRefs map[string]DatabaseRef, svcRefs map[string]ServiceRef) (bool, error) {
	ctx = auth.OutgoingContext(ctx)

	m := make(map[string]string, len(vars))
	for _, v := range vars {
		m[v.Key] = v.Value
	}

	protoDBRefs := make(map[string]*packager.DatabaseRef, len(dbRefs))
	for k, ref := range dbRefs {
		protoDBRefs[k] = &packager.DatabaseRef{Database: ref.Database, Key: ref.Key}
	}

	protoSvcRefs := make(map[string]*packager.ServiceRef, len(svcRefs))
	for k, ref := range svcRefs {
		protoSvcRefs[k] = &packager.ServiceRef{Service: ref.Service}
	}

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	_, err := c.Packager.SetServiceVariables(callCtx, &packager.SetServiceVariablesRequest{
		Project:      projectID,
		Environment:  environment,
		Service:      service,
		Variables:    m,
		SharedRefs:   sharedRefs,
		DatabaseRefs: protoDBRefs,
		ServiceRefs:  protoSvcRefs,
	})
	if err != nil {
		return false, fmt.Errorf("failed to set service variables: %w", err)
	}
	return true, nil
}
