package conductor

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
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

type ServiceVariable struct {
	Key         string
	Value       string
	FromShared  bool
	DatabaseRef *DatabaseRef
}

// cnpgKeyDisplayNames maps CNPG secret keys to human-readable display names
// for the rendered dashboard placeholder value (e.g. "${{Mydb.DATABASE_URL}}").
var cnpgKeyDisplayNames = map[string]string{
	"uri":      "DATABASE_URL",
	"host":     "PGHOST",
	"port":     "PGPORT",
	"dbname":   "PGDATABASE",
	"user":     "PGUSER",
	"password": "PGPASSWORD",
}

func (c *Client) SharedVariables(ctx context.Context, environment platform.EnvironmentID) ([]Variable, error) {
	vars, err := c.deployer.Environments().Variables(ctx, environment)

	if err != nil {
		return nil, fmt.Errorf("read shared variables: %w", err)
	}

	result := make([]Variable, 0, len(vars))

	for k, v := range vars {
		result = append(result, Variable{Key: k, Value: v})
	}

	return result, nil
}

func (c *Client) SetSharedVariables(ctx context.Context, environment platform.EnvironmentID, vars []Variable) (bool, error) {
	m := make(map[string]string, len(vars))

	for _, v := range vars {
		m[v.Key] = v.Value
	}

	if _, err := c.deployer.Environments().SetVariables(ctx, environment, m); err != nil {
		return false, fmt.Errorf("write shared variables: %w", err)
	}

	return true, nil
}

// ServiceVariables returns the variable rows the dashboard should render
// for a single service:
//   - literals (with FromShared set when the user marked the row as
//     shared-sourced; the literal value wins at pod-runtime over any
//     envFrom-supplied shared value of the same name)
//   - sharedRefs that have no literal counterpart, with their resolved
//     shared value looked up from the env's shared bag
//   - database refs rendered with a placeholder value like
//     "${{MyDB.DATABASE_URL}}" the dashboard treats as a token
func (c *Client) ServiceVariables(ctx context.Context, service platform.ServiceID) ([]ServiceVariable, error) {
	spec, err := c.deployer.Services().Variables(ctx, service)

	if err != nil {
		return nil, err
	}

	sharedSet := make(map[string]bool, len(spec.SharedRefs))

	for _, k := range spec.SharedRefs {
		sharedSet[k] = true
	}

	result := make([]ServiceVariable, 0, len(spec.Literals)+len(spec.DatabaseRefs)+len(spec.SharedRefs))

	for k, v := range spec.Literals {
		result = append(result, ServiceVariable{
			Key:        k,
			Value:      v,
			FromShared: sharedSet[k],
		})
	}

	// Emit a row for each sharedRef without a literal counterpart. The
	// shared value resolves at read time (via the env's shared bag) so
	// the row shows the user what the pod will actually receive.
	missingShared := make([]string, 0)

	for _, k := range spec.SharedRefs {
		if _, hasLiteral := spec.Literals[k]; !hasLiteral {
			missingShared = append(missingShared, k)
		}
	}

	if len(missingShared) > 0 {
		sharedBag, err := c.deployer.Environments().Variables(ctx, service.EnvironmentID())

		if err != nil {
			return nil, fmt.Errorf("resolve shared variables for refs: %w", err)
		}

		for _, k := range missingShared {
			result = append(result, ServiceVariable{
				Key:        k,
				Value:      sharedBag[k],
				FromShared: true,
			})
		}
	}

	for k, ref := range spec.DatabaseRefs {
		displayKey := cnpgKeyDisplayNames[ref.Key]

		if displayKey == "" {
			displayKey = ref.Key
		}

		dbName := capitalize(ref.Database)

		result = append(result, ServiceVariable{
			Key:   k,
			Value: fmt.Sprintf("${{%s.%s}}", dbName, displayKey),
			DatabaseRef: &DatabaseRef{
				Database: platform.DatabaseID{
					Workspace:   service.Workspace,
					Project:     service.Project,
					Environment: service.Environment,
					Name:        ref.Database,
				},
				Key: ref.Key,
			},
		})
	}

	return result, nil
}

func (c *Client) SetServiceVariables(ctx context.Context, service platform.ServiceID, vars []Variable, sharedRefs []string, dbRefs map[string]DatabaseRef) (bool, error) {
	literals := make(map[string]string, len(vars))

	for _, v := range vars {
		literals[v.Key] = v.Value
	}

	specDBRefs := make(map[string]deployer.DatabaseRef, len(dbRefs))

	for k, ref := range dbRefs {
		specDBRefs[k] = deployer.DatabaseRef{
			Database: ref.Database.Name,
			Key:      ref.Key,
		}
	}

	if _, err := c.deployer.Services().SetVariables(ctx, service, deployer.ServiceVariablesSpec{
		Literals:     literals,
		DatabaseRefs: specDBRefs,
		SharedRefs:   sharedRefs,
	}); err != nil {
		return false, fmt.Errorf("write service variables: %w", err)
	}

	return true, nil
}

func capitalize(s string) string {
	if s == "" {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}
