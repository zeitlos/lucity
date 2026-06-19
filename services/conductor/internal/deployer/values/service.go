package values

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/zeitlos/lucity/services/conductor/internal/image"
)

type Service struct {
	Image              ImageRef               `yaml:"image"`
	Port               int                    `yaml:"port,omitempty"`
	Replicas           int                    `yaml:"replicas,omitempty"`
	Autoscaling        *Autoscaling           `yaml:"autoscaling,omitempty"`
	Resources          Resources              `yaml:"resources,omitempty"`
	Domains            []Domain               `yaml:"domains,omitempty"`
	CustomStartCommand string                 `yaml:"customStartCommand,omitempty"`
	Env                map[string]string      `yaml:"env,omitempty"`
	SharedRefs         []string               `yaml:"sharedRefs,omitempty"`
	DatabaseRefs       map[string]DatabaseRef `yaml:"databaseRefs,omitempty"`
	Labels             map[string]string      `yaml:"labels,omitempty"`
	Annotations        map[string]string      `yaml:"annotations,omitempty"`
	PodLabels          map[string]string      `yaml:"podLabels,omitempty"`
	PodAnnotations     map[string]string      `yaml:"podAnnotations,omitempty"`
}

type ImageRef struct {
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag,omitempty"`
	Digest     string `yaml:"digest,omitempty"`
	PullPolicy string `yaml:"pullPolicy,omitempty"`
}

type Autoscaling struct {
	Enabled     bool `yaml:"enabled"`
	MinReplicas int  `yaml:"minReplicas"`
	MaxReplicas int  `yaml:"maxReplicas"`
	TargetCPU   int  `yaml:"targetCPU"`
}

type Resources struct {
	Requests ResourceList `yaml:"requests,omitempty"`
	Limits   ResourceList `yaml:"limits,omitempty"`
}

type ResourceList struct {
	CPU    string `yaml:"cpu,omitempty"`
	Memory string `yaml:"memory,omitempty"`
}

type DatabaseRef struct {
	Database string `yaml:"database"`
	Key      string `yaml:"key"`
}

type Domain struct {
	Host     string `yaml:"host"`
	Verified bool   `yaml:"verified"`
}

type ServiceSpec struct {
	Image                string
	SourceURL            string
	ContextPath          string
	GitHubInstallationID int64
	Port                 int
	Resources            Resources
}

func CreateService(env *Env, name string, spec ServiceSpec) error {
	if !isValidName(name) {
		return fmt.Errorf("invalid service name %q", name)
	}

	if _, ok := env.Services[name]; ok {
		// To keep this function idempotent, don't return an error if the service already exists.
		return nil
	}

	ref, err := image.Parse(spec.Image)

	if err != nil {
		return fmt.Errorf("invalid image %q: %w", spec.Image, err)
	}

	if env.Services == nil {
		env.Services = map[string]Service{}
	}

	labels := map[string]string{labelService: name}
	podLabels := map[string]string{labelService: name}
	annotations := map[string]string{}
	podAnnotations := map[string]string{}

	if spec.GitHubInstallationID != 0 {
		labels[labelGitHubInstallation] = strconv.FormatInt(spec.GitHubInstallationID, 10)
	}

	if spec.SourceURL != "" {
		annotations[annotationSourceRepo] = spec.SourceURL
		podAnnotations[annotationSourceRepo] = spec.SourceURL
	}

	if spec.ContextPath != "" {
		annotations[annotationSourceContext] = spec.ContextPath
		podAnnotations[annotationSourceContext] = spec.ContextPath
	}

	env.Services[name] = Service{
		Image: ImageRef{
			Repository: ref.Repository,
			Tag:        ref.Tag,
			Digest:     ref.Digest,
		},
		Port:           spec.Port,
		Labels:         labels,
		Annotations:    annotations,
		PodLabels:      podLabels,
		PodAnnotations: podAnnotations,
		Resources:      spec.Resources,
	}

	return nil
}

func DeleteService(env *Env, name string) error {
	if _, ok := env.Services[name]; !ok {
		return fmt.Errorf("service %q not found", name)
	}

	delete(env.Services, name)

	return nil
}

func SetServiceImage(env *Env, name string, ref image.Ref, commitMessage string) error {
	return mutateService(env, name, func(s *Service) {
		s.Image.Repository = ref.Repository
		s.Image.Tag = ref.Tag
		s.Image.Digest = ref.Digest

		if s.PodAnnotations == nil {
			s.PodAnnotations = map[string]string{}
		}

		if commitMessage != "" {
			s.PodAnnotations[annotationSourceMessage] = commitMessage
		} else {
			delete(s.PodAnnotations, annotationSourceMessage)
		}
	})
}

func SetServiceReplicas(env *Env, name string, replicas int) error {
	if replicas < 0 {
		return fmt.Errorf("replicas must be non-negative")
	}

	return mutateService(env, name, func(s *Service) {
		s.Replicas = replicas
		s.Autoscaling = nil
	})
}

func SetServiceAutoscaling(env *Env, name string, cfg Autoscaling) error {
	if cfg.MinReplicas < 0 || cfg.MaxReplicas < cfg.MinReplicas {
		return fmt.Errorf("invalid autoscaling range: min=%d max=%d", cfg.MinReplicas, cfg.MaxReplicas)
	}

	if cfg.TargetCPU <= 0 || cfg.TargetCPU > 100 {
		return fmt.Errorf("targetCPU must be in (0, 100]")
	}

	cfg.Enabled = true

	return mutateService(env, name, func(s *Service) {
		s.Autoscaling = &cfg
	})
}

func SetServiceResources(env *Env, name string, resources Resources) error {
	return mutateService(env, name, func(s *Service) {
		s.Resources = resources
	})
}

func SetServiceCommand(env *Env, name, command string) error {
	if err := validateStartCommand(command); err != nil {
		return err
	}

	return mutateService(env, name, func(s *Service) {
		s.CustomStartCommand = command
	})
}

func SetServiceBranch(env *Env, name, branch string) error {
	return mutateService(env, name, func(s *Service) {
		s.Annotations[annotationSourceBranch] = branch
	})
}

func SetServicePort(env *Env, name string, port int) error {
	if !isValidPort(port) {
		return fmt.Errorf("port must be in [0, 65535]")
	}

	return mutateService(env, name, func(s *Service) {
		s.Port = port
	})
}

// SetServiceVariables replaces a service's entire variable surface:
// literals (Env), database refs, and shared-ref UI metadata.
func SetServiceVariables(env *Env, name string, literals map[string]string, dbRefs map[string]DatabaseRef, sharedRefs []string) error {
	for k := range literals {
		if !isValidVarName(k) {
			return fmt.Errorf("invalid variable name %q", k)
		}
	}

	for k := range dbRefs {
		if !isValidVarName(k) {
			return fmt.Errorf("invalid databaseRef env key %q", k)
		}
	}

	for _, k := range sharedRefs {
		if !isValidVarName(k) {
			return fmt.Errorf("invalid sharedRef key %q", k)
		}
	}

	return mutateService(env, name, func(s *Service) {
		s.Env = cloneStringMap(literals)
		s.DatabaseRefs = cloneDatabaseRefs(dbRefs)
		s.SharedRefs = append([]string(nil), sharedRefs...)
	})
}

func cloneDatabaseRefs(in map[string]DatabaseRef) map[string]DatabaseRef {
	if in == nil {
		return nil
	}

	out := make(map[string]DatabaseRef, len(in))

	for k, v := range in {
		out[k] = v
	}

	return out
}

func AddServiceDomain(env *Env, name, host string) error {
	if !isValidHostname(host) {
		return fmt.Errorf("invalid hostname %q", host)
	}

	return mutateService(env, name, func(s *Service) {
		if slices.ContainsFunc(s.Domains, func(d Domain) bool { return d.Host == host }) {
			return
		}

		s.Domains = append(s.Domains, Domain{Host: host, Verified: false})
	})
}

func RemoveServiceDomain(env *Env, name, host string) error {
	return mutateService(env, name, func(s *Service) {
		s.Domains = slices.DeleteFunc(s.Domains, func(d Domain) bool {
			return d.Host == host
		})
	})
}

func VerifyServiceDomain(env *Env, name, host string, verified bool) error {
	return mutateService(env, name, func(s *Service) {
		i := slices.IndexFunc(s.Domains, func(d Domain) bool { return d.Host == host })

		if i >= 0 {
			s.Domains[i].Verified = verified
		}
	})
}

func mutateService(env *Env, name string, mutate func(*Service)) error {
	svc, ok := env.Services[name]

	if !ok {
		return fmt.Errorf("service %q not found", name)
	}

	mutate(&svc)
	env.Services[name] = svc

	return nil
}
