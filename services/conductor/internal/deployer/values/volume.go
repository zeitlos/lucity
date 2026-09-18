package values

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	minVolumeSize = resource.MustParse("10Gi")
	maxVolumeSize = resource.MustParse("1Ti")
)

type Volume struct {
	Size        string            `yaml:"size,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

func CreateVolume(env *Env, name string, size resource.Quantity) error {
	if _, ok := env.Volumes[name]; ok {
		return nil
	}

	if env.Volumes == nil {
		env.Volumes = map[string]Volume{}
	}

	env.Volumes[name] = Volume{
		Size:   size.String(),
		Labels: map[string]string{labelVolume: name},
	}

	return nil
}

func ExpandVolume(env *Env, name string, size resource.Quantity) error {
	volume, ok := env.Volumes[name]

	if !ok {
		return fmt.Errorf("volume %q not found", name)
	}

	current, err := resource.ParseQuantity(volume.Size)

	if err != nil {
		return fmt.Errorf("parse current size %q: %w", volume.Size, err)
	}

	if size.Cmp(current) < 0 {
		return fmt.Errorf("volume storage can only be increased, not shrunk from %s to %s", current.String(), size.String())
	}

	volume.Size = size.String()
	env.Volumes[name] = volume

	return nil
}

func DeleteVolume(env *Env, name string) error {
	if _, ok := env.Volumes[name]; !ok {
		return fmt.Errorf("volume %q not found", name)
	}

	for serviceName, svc := range env.Services {
		if _, mounted := svc.VolumeMounts[name]; mounted {
			return fmt.Errorf("volume %q is mounted by service %q; unmount it first", name, serviceName)
		}
	}

	delete(env.Volumes, name)

	return nil
}

func MountVolume(env *Env, volumeName, serviceName, path string) error {
	svc, ok := env.Services[serviceName]

	if !ok {
		return fmt.Errorf("service %q not found", serviceName)
	}

	if existing, mounted := svc.VolumeMounts[volumeName]; mounted {
		if existing == path {
			return nil
		}

		return fmt.Errorf("volume %q is already mounted on service %q at %q", volumeName, serviceName, existing)
	}

	return mutateService(env, serviceName, func(s *Service) {
		if s.VolumeMounts == nil {
			s.VolumeMounts = map[string]string{}
		}

		s.VolumeMounts[volumeName] = path
	})
}

func UnmountVolume(env *Env, volumeName string) error {
	if _, ok := env.Volumes[volumeName]; !ok {
		return fmt.Errorf("volume %q not found", volumeName)
	}

	for serviceName, svc := range env.Services {
		if _, mounted := svc.VolumeMounts[volumeName]; !mounted {
			continue
		}

		if err := mutateService(env, serviceName, func(s *Service) {
			delete(s.VolumeMounts, volumeName)
		}); err != nil {
			return err
		}
	}

	return nil
}
