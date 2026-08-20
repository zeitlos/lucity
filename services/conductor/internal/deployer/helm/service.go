package helm

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer/values"
	"github.com/zeitlos/lucity/services/conductor/internal/image"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type serviceClient struct {
	client *Client
}

func (s *serviceClient) Create(ctx context.Context, env platform.EnvironmentID, name string, spec deployer.ServiceSpec) (deployer.RevisionID, error) {
	if err := validateResources(spec.Resources); err != nil {
		return "", err
	}

	return s.client.applyEnv(ctx, env, func(e *values.Env) error {
		spec := values.ServiceSpec{
			Image:                spec.Image,
			SourceURL:            spec.SourceURL,
			ContextPath:          spec.ContextPath,
			GitHubInstallationID: spec.GitHubInstallationID,
			AutoDeploy:           spec.AutoDeploy,
			Port:                 spec.Port,
			Resources:            deriveRequestsAndLimtis(spec.Resources, spec.ResourceTier),
			Env:                  spec.Env,
			RunAsUser:            spec.SecurityContext.RunAsUser,
			RunAsGroup:           spec.SecurityContext.RunAsGroup,
			FsGroup:              spec.SecurityContext.FsGroup,
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

func (s *serviceClient) SetImage(ctx context.Context, id platform.ServiceID, ref image.Ref, provenance deployer.ImageProvenance, release deployer.ReleaseMeta) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.SetServiceImage(e, id.Name, ref, values.ImageProvenance{
			Commit:        provenance.Commit,
			CommitMessage: provenance.CommitMessage,
			BuildID:       provenance.BuildID,
		}, values.ReleaseMeta{
			ID:      release.ID,
			Trigger: string(release.Trigger),
			Actor:   release.Actor,
		})
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
	if err := validateResources(res); err != nil {
		return "", err
	}

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

func (s *serviceClient) SetAutoDeploy(ctx context.Context, id platform.ServiceID, enabled bool) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.SetServiceAutoDeploy(e, id.Name, enabled)
	})
}

func (s *serviceClient) SetCIDeploy(ctx context.Context, id platform.ServiceID, enabled bool) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.SetServiceCIDeploy(e, id.Name, enabled)
	})
}

func (s *serviceClient) SetPort(ctx context.Context, id platform.ServiceID, port int) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.SetServicePort(e, id.Name, port)
	})
}

func (s *serviceClient) SetHealthCheck(ctx context.Context, id platform.ServiceID, healthCheck *deployer.HealthCheck) (deployer.RevisionID, error) {
	var spec *values.HealthCheck

	if healthCheck != nil {
		spec = &values.HealthCheck{
			Path:                    healthCheck.Path,
			Port:                    healthCheck.Port,
			InitialDelaySeconds:     healthCheck.InitialDelaySeconds,
			PeriodSeconds:           healthCheck.PeriodSeconds,
			TimeoutSeconds:          healthCheck.TimeoutSeconds,
			FailureThreshold:        healthCheck.FailureThreshold,
			StartupFailureThreshold: healthCheck.StartupFailureThreshold,
		}
	}

	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.SetServiceHealthCheck(e, id.Name, spec)
	})
}

func (s *serviceClient) SetSecurityContext(ctx context.Context, id platform.ServiceID, sc deployer.SecurityContext) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.SetServiceSecurityContext(e, id.Name, sc.RunAsUser, sc.RunAsGroup, sc.FsGroup)
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

	refs := make(map[string]deployer.VariableRef, len(svc.Refs))

	for k, ref := range svc.Refs {
		refs[k] = deployer.VariableRef{
			Secret: ref.Secret,
			Key:    ref.Key,
		}
	}

	return deployer.ServiceVariablesSpec{
		Literals: cloneStringMap(svc.Env),
		Refs:     refs,
	}, nil
}

func (s *serviceClient) SetVariables(ctx context.Context, id platform.ServiceID, spec deployer.ServiceVariablesSpec) (deployer.RevisionID, error) {
	refs := make(map[string]values.SecretRef, len(spec.Refs))

	for k, ref := range spec.Refs {
		refs[k] = values.SecretRef{
			Secret: ref.Secret,
			Key:    ref.Key,
		}
	}

	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.SetServiceVariables(e, id.Name, spec.Literals, refs)
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

func (s *serviceClient) AddDomain(ctx context.Context, id platform.ServiceID, host string, ownListener bool) (deployer.RevisionID, error) {
	var listenerSet *values.ListenerSet

	if ownListener {
		listenerSet = &values.ListenerSet{
			Enabled: true,
			Certificate: values.ListenerSetCertificate{
				IssuerRef: values.IssuerRef{Kind: "ClusterIssuer", Name: s.client.clusterIssuer},
			},
		}
	}

	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.AddServiceDomain(e, id.Name, host, listenerSet)
	})
}

func (s *serviceClient) RemoveDomain(ctx context.Context, id platform.ServiceID, host string) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.RemoveServiceDomain(e, id.Name, host)
	})
}

func (s *serviceClient) AttachDomain(ctx context.Context, id platform.ServiceID, host string, attached bool) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.AttachServiceDomain(e, id.Name, host, attached)
	})
}

func (s *serviceClient) Mount(ctx context.Context, id platform.ServiceID, volume platform.VolumeID, mountPath string) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.MountVolume(e, volume.Name, id.Name, mountPath)
	})
}

func (s *serviceClient) Unmount(ctx context.Context, volume platform.VolumeID) (deployer.RevisionID, error) {
	return s.client.applyEnv(ctx, volume.EnvironmentID(), func(e *values.Env) error {
		return values.UnmountVolume(e, volume.Name)
	})
}

var _ deployer.ServiceClient = (*serviceClient)(nil)
