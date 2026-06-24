package conductor

import (
	"context"
	"errors"
	"fmt"

	"github.com/zeitlos/lucity/services/conductor/internal/eject"
)

func (c *Client) Eject(ctx context.Context, projectID ProjectID) ([]byte, error) {
	if c.config.ChartFS == nil {
		return nil, errors.New("chart not available")
	}

	environments, err := c.platform.Environments(ctx, projectID)

	if err != nil {
		return nil, fmt.Errorf("list environments: %w", err)
	}

	var envValues []eject.EnvValues

	for _, environment := range environments {
		vals, err := c.deployer.Environments().Export(ctx, environment.ID)

		if err != nil {
			return nil, fmt.Errorf("export environment %q: %w", environment.ID, err)
		}

		if vals == nil {
			continue
		}

		envValues = append(envValues, eject.EnvValues{Name: environment.ID.Name, Values: vals})
	}

	return eject.Build(c.config.ChartFS, eject.Project{Name: projectID.Name, ID: projectID.String()}, envValues)
}
