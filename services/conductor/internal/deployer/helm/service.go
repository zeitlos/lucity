package helm

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer/values"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
	"github.com/zeitlos/lucity/services/conductor/internal/resources"
)

type serviceClient struct {
	client *Client
}

func (s *serviceClient) Create(ctx context.Context, env platform.EnvironmentID, name string, spec deployer.ServiceSpec) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, env, func(e *values.Env) error {

		spec := values.ServiceSpec{
			Image:                spec.Image,
			SourceURL:            spec.SourceURL,
			ContextPath:          spec.ContextPath,
			GitHubInstallationID: spec.GitHubInstallationID,
			Port:                 spec.Port,
			Resources:            deriveRequestsAndLimtis(spec.Resources, spec.ResourceTier),
		}

		return values.CreateService(e, name, spec)
	})
}

func (s *serviceClient) Delete(ctx context.Context, id platform.ServiceID) error {
	_, err := s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.DeleteService(e, id.Name)
	})

	return err
}

func (s *serviceClient) SetImage(ctx context.Context, id platform.ServiceID, ref, digest string) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.SetServiceImage(e, id.Name, ref, digest)
	})
}

func (s *serviceClient) SetReplicas(ctx context.Context, id platform.ServiceID, replicas int) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.SetServiceReplicas(e, id.Name, replicas)
	})
}

func (s *serviceClient) SetAutoscaling(ctx context.Context, id platform.ServiceID, config deployer.Autoscaling) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.SetServiceAutoscaling(e, id.Name, values.Autoscaling{
			MinReplicas: config.MinReplicas,
			MaxReplicas: config.MaxReplicas,
			TargetCPU:   config.TargetCPU,
		})
	})
}

func (s *serviceClient) SetResources(ctx context.Context, id platform.ServiceID, tier platform.ResourceTier, res deployer.Resources) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.SetServiceResources(e, id.Name, deriveRequestsAndLimtis(res, tier))
	})
}

func (s *serviceClient) SetCommand(ctx context.Context, id platform.ServiceID, command string) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.SetServiceCommand(e, id.Name, command)
	})
}

func (s *serviceClient) SetBranch(ctx context.Context, id platform.ServiceID, branch string) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.SetServiceBranch(e, id.Name, branch)
	})
}

func (s *serviceClient) SetPort(ctx context.Context, id platform.ServiceID, port int) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.SetServicePort(e, id.Name, port)
	})
}

func (s *serviceClient) Variables(ctx context.Context, id platform.ServiceID) (deployer.ServiceVariablesSpec, error) {
	env, err := s.client.loadEnv(ctx, id.EnvironmentID())

	if err != nil {
		return deployer.ServiceVariablesSpec{}, err
	}

	svc, ok := env.Services[id.Name]

	if !ok {
		return deployer.ServiceVariablesSpec{}, fmt.Errorf("service %q not found", id.Name)
	}

	dbRefs := make(map[string]deployer.DatabaseRef, len(svc.DatabaseRefs))

	for k, ref := range svc.DatabaseRefs {
		dbRefs[k] = deployer.DatabaseRef{
			Database: ref.Database,
			Key:      ref.Key,
		}
	}

	return deployer.ServiceVariablesSpec{
		Literals:     cloneStringMap(svc.Env),
		DatabaseRefs: dbRefs,
		SharedRefs:   append([]string(nil), svc.SharedRefs...),
	}, nil
}

func (s *serviceClient) SetVariables(ctx context.Context, id platform.ServiceID, spec deployer.ServiceVariablesSpec) (deployer.RevisionID, error) {
	dbRefs := make(map[string]values.DatabaseRef, len(spec.DatabaseRefs))

	for k, ref := range spec.DatabaseRefs {
		dbRefs[k] = values.DatabaseRef{
			Database: ref.Database,
			Key:      ref.Key,
		}
	}

	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.SetServiceVariables(e, id.Name, spec.Literals, dbRefs, spec.SharedRefs)
	})
}

func cloneStringMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}

	out := make(map[string]string, len(in))

	for k, v := range in {
		out[k] = v
	}

	return out
}

func (s *serviceClient) AddDomain(ctx context.Context, id platform.ServiceID, host string) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.AddServiceDomain(e, id.Name, host)
	})
}

func (s *serviceClient) RemoveDomain(ctx context.Context, id platform.ServiceID, host string) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.RemoveServiceDomain(e, id.Name, host)
	})
}

func (s *serviceClient) VerifyDomain(ctx context.Context, id platform.ServiceID, host string, verified bool) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.VerifyServiceDomain(e, id.Name, host, verified)
	})
}

func (s *serviceClient) Mount(ctx context.Context, id platform.ServiceID, volume platform.VolumeID, mountPath string) (deployer.RevisionID, error) {
	return "", fmt.Errorf("Mount: chart does not support volumes yet")
}

func (s *serviceClient) Unmount(ctx context.Context, id platform.ServiceID, volume platform.VolumeID) (deployer.RevisionID, error) {
	return "", fmt.Errorf("Unmount: chart does not support volumes yet")
}

func deriveRequestsAndLimtis(res deployer.Resources, tier platform.ResourceTier) values.Resources {
	cpuLimit := res.CPU
	memoryLimit := res.CPU

	if cpuLimit.Value() == 0 {
		cpuLimit = resources.DefaultCPULimit
	}

	if memoryLimit.Value() == 0 {
		memoryLimit = resources.DefaultMemoryLimit
	}

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

var _ deployer.ServiceClient = (*serviceClient)(nil)
