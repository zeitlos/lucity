package values

import (
	"fmt"
	"slices"
	"strings"
)

type Service struct {
	Image                ImageRef               `yaml:"image"`
	Port                 int                    `yaml:"port"`
	Replicas             int                    `yaml:"replicas,omitempty"`
	Autoscaling          *Autoscaling           `yaml:"autoscaling,omitempty"`
	Resources            Resources              `yaml:"resources,omitempty"`
	Domains              []Domain               `yaml:"domains,omitempty"`
	CustomStartCommand   string                 `yaml:"customStartCommand,omitempty"`
	Env                  map[string]string      `yaml:"env,omitempty"`
	SharedRefs           []string               `yaml:"sharedRefs,omitempty"`
	DatabaseRefs         map[string]DatabaseRef `yaml:"databaseRefs,omitempty"`
	ServiceRefs          map[string]ServiceRef  `yaml:"serviceRefs,omitempty"`
	SourceURL            string                 `yaml:"sourceUrl,omitempty"`
	ContextPath          string                 `yaml:"contextPath,omitempty"`
	Branch               string                 `yaml:"branch,omitempty"`
	GitHubInstallationID int64                  `yaml:"githubInstallationId,omitempty"`
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

type ServiceRef struct {
	Service string `yaml:"service"`
}

type Domain struct {
	Host     string `yaml:"host"`
	Verified bool   `yaml:"verified"`
}

type ServiceSpec struct {
	Image                string
	Port                 int
	SourceURL            string
	ContextPath          string
	Branch               string
	GitHubInstallationID int64
	StartCommand         string
}

func CreateService(env *Env, name string, spec ServiceSpec) error {
	if !isValidName(name) {
		return fmt.Errorf("invalid service name %q", name)
	}

	if _, ok := env.Services[name]; ok {
		return fmt.Errorf("service %q already exists", name)
	}

	repository, tag := splitImageRef(spec.Image)

	if env.Services == nil {
		env.Services = map[string]Service{}
	}

	env.Services[name] = Service{
		Image: ImageRef{
			Repository: repository,
			Tag:        tag,
		},
		Port:                 spec.Port,
		SourceURL:            spec.SourceURL,
		ContextPath:          spec.ContextPath,
		Branch:               spec.Branch,
		GitHubInstallationID: spec.GitHubInstallationID,
		CustomStartCommand:   spec.StartCommand,
	}

	return nil
}

func RemoveService(env *Env, name string) error {
	if _, ok := env.Services[name]; !ok {
		return fmt.Errorf("service %q not found", name)
	}

	delete(env.Services, name)

	return nil
}

func SetServiceImage(env *Env, name, ref, digest string) error {
	return mutateService(env, name, func(s *Service) {
		repository, tag := splitImageRef(ref)
		s.Image.Repository = repository
		s.Image.Tag = tag
		s.Image.Digest = digest
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

func SetServiceResources(env *Env, name string, requests, limits ResourceList) error {
	return mutateService(env, name, func(s *Service) {
		s.Resources = Resources{
			Requests: requests,
			Limits:   limits,
		}
	})
}

func SetServiceCommand(env *Env, name, command string) error {
	return mutateService(env, name, func(s *Service) {
		s.CustomStartCommand = command
	})
}

func SetServiceBranch(env *Env, name, branch string) error {
	return mutateService(env, name, func(s *Service) {
		s.Branch = branch
	})
}

func SetServicePort(env *Env, name string, port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("port must be in (0, 65535]")
	}

	return mutateService(env, name, func(s *Service) {
		s.Port = port
	})
}

func SetServiceVariables(env *Env, name string, vars map[string]string) error {
	for k := range vars {
		if !isValidVarName(k) {
			return fmt.Errorf("invalid variable name %q", k)
		}
	}

	return mutateService(env, name, func(s *Service) {
		s.Env = cloneStringMap(vars)
	})
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

func splitImageRef(ref string) (repository, tag string) {
	slashIdx := strings.LastIndex(ref, "/")
	colonIdx := strings.LastIndex(ref, ":")

	if colonIdx > slashIdx {
		return ref[:colonIdx], ref[colonIdx+1:]
	}

	return ref, ""
}
