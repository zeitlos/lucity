package helm

import (
	"errors"

	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/pkg/resources"
)

func validateResources(res deployer.Resources) error {
	if res.CPU.Value() == 0 {
		return errors.New("cpu limit must be provided")
	}

	if res.Memory.Value() == 0 {
		return errors.New("memory limit must be provided")
	}

	return nil
}

func deriveRequestsAndLimtis(res deployer.Resources, tier platform.ResourceTier) values.Resources {
	cpuLimit := res.CPU
	memoryLimit := res.Memory

	limits := values.ResourceList{
		CPU:    cpuLimit.String(),
		Memory: memoryLimit.String(),
	}

	requests := values.ResourceList{
		CPU:    resources.Request(tier, cpuLimit).String(),
		Memory: resources.Request(tier, memoryLimit).String(),
	}

	return values.Resources{
		Requests: requests,
		Limits:   limits,
	}
}
