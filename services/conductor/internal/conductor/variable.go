package conductor

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type Variable = platform.Variable
type VariableID = platform.VariableID

type ServiceVariable struct {
	Key   string
	Value string
	Ref   *platform.VariableID
}

func (c *Client) AvailableVariables(ctx context.Context, environment platform.EnvironmentID) ([]Variable, error) {
	return c.platform.AvailableVariables(ctx, environment)
}

func (c *Client) SetSharedVariables(ctx context.Context, environment platform.EnvironmentID, vars map[string]string) (bool, error) {
	if _, err := c.deployer.Environments().SetVariables(ctx, environment, vars); err != nil {
		return false, fmt.Errorf("write shared variables: %w", err)
	}

	return true, nil
}

func (c *Client) ServiceVariables(ctx context.Context, service platform.ServiceID) ([]ServiceVariable, error) {
	spec, err := c.deployer.Services().Variables(ctx, service)

	if err != nil {
		return nil, err
	}

	result := make([]ServiceVariable, 0, len(spec.Literals)+len(spec.Refs))

	for key, value := range spec.Literals {
		result = append(result, ServiceVariable{Key: key, Value: value})
	}

	for key, ref := range spec.Refs {
		id := platform.VariableID{
			Workspace:   service.Workspace,
			Project:     service.Project,
			Environment: service.Environment,
			Secret:      ref.Secret,
			Name:        ref.Key,
		}

		result = append(result, ServiceVariable{Key: key, Ref: &id})
	}

	return result, nil
}

func (c *Client) SetServiceVariables(ctx context.Context, service platform.ServiceID, literals map[string]string, refs map[string]platform.VariableID) (bool, error) {
	specRefs := make(map[string]deployer.VariableRef, len(refs))

	for key, id := range refs {
		if id.EnvironmentID() != service.EnvironmentID() {
			return false, fmt.Errorf("variable %q: reference must be in the same environment", key)
		}

		specRefs[key] = deployer.VariableRef{Secret: id.Secret, Key: id.Name}
	}

	if _, err := c.deployer.Services().SetVariables(ctx, service, deployer.ServiceVariablesSpec{
		Literals: literals,
		Refs:     specRefs,
	}); err != nil {
		return false, fmt.Errorf("write service variables: %w", err)
	}

	return true, nil
}
