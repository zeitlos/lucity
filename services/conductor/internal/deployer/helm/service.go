package helm

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer/values"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type serviceClient struct {
	client *Client
}

func (s *serviceClient) Create(ctx context.Context, env platform.EnvironmentID, name string, spec deployer.ServiceSpec) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, env, func(e *values.Env) error {
		return values.CreateService(e, name, toValuesSpec(spec))
	})
}

func (s *serviceClient) Remove(ctx context.Context, id platform.ServiceID) error {
	_, err := s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.RemoveService(e, id.Name)
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

func (s *serviceClient) SetResources(ctx context.Context, id platform.ServiceID, cpu, memory resource.Quantity) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.SetServiceResources(e, id.Name, cpu, memory)
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

func (s *serviceClient) SetVariables(ctx context.Context, id platform.ServiceID, vars map[string]string) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.SetServiceVariables(e, id.Name, vars)
	})
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

func (s *serviceClient) Mount(ctx context.Context, id platform.ServiceID, volume platform.VolumeID, mountPath string) (deployer.RevisionID, error) {
	return "", fmt.Errorf("Mount: chart does not support volumes yet")
}

func (s *serviceClient) Unmount(ctx context.Context, id platform.ServiceID, volume platform.VolumeID) (deployer.RevisionID, error) {
	return "", fmt.Errorf("Unmount: chart does not support volumes yet")
}

var _ deployer.ServiceClient = (*serviceClient)(nil)

func toValuesSpec(s deployer.ServiceSpec) values.ServiceSpec {
	return values.ServiceSpec{
		Image:                s.Image,
		Port:                 s.Port,
		SourceURL:            s.SourceURL,
		ContextPath:          s.ContextPath,
		Branch:               s.Branch,
		GitHubInstallationID: s.GitHubInstallationID,
		StartCommand:         s.StartCommand,
	}
}
