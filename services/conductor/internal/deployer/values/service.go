package values

import (
	"fmt"
	"maps"
	"slices"
	"strconv"

	"github.com/zeitlos/lucity/services/conductor/internal/image"
)

type Service struct {
	Image          ImageRef             `yaml:"image"`
	Port           int                  `yaml:"port,omitempty"`
	Replicas       int                  `yaml:"replicas,omitempty"`
	Autoscaling    *Autoscaling         `yaml:"autoscaling,omitempty"`
	Resources      Resources            `yaml:"resources,omitempty"`
	Domains        []Domain             `yaml:"domains,omitempty"`
	Command        string               `yaml:"command,omitempty"`
	HealthCheck    *HealthCheck         `yaml:"healthCheck,omitempty"`
	RunAsUser      *int64               `yaml:"runAsUser,omitempty"`
	RunAsGroup     *int64               `yaml:"runAsGroup,omitempty"`
	FsGroup        *int64               `yaml:"fsGroup,omitempty"`
	Env            map[string]string    `yaml:"env,omitempty"`
	Refs           map[string]SecretRef `yaml:"refs,omitempty"`
	VolumeMounts   map[string]string    `yaml:"volumeMounts,omitempty"`
	Labels         map[string]string    `yaml:"labels,omitempty"`
	Annotations    map[string]string    `yaml:"annotations,omitempty"`
	PodLabels      map[string]string    `yaml:"podLabels,omitempty"`
	PodAnnotations map[string]string    `yaml:"podAnnotations,omitempty"`
}

type HealthCheck struct {
	Path                    string `yaml:"path"`
	Port                    int    `yaml:"port,omitempty"`
	InitialDelaySeconds     int    `yaml:"initialDelaySeconds,omitempty"`
	PeriodSeconds           int    `yaml:"periodSeconds,omitempty"`
	TimeoutSeconds          int    `yaml:"timeoutSeconds,omitempty"`
	FailureThreshold        int    `yaml:"failureThreshold,omitempty"`
	StartupFailureThreshold int    `yaml:"startupFailureThreshold,omitempty"`
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

type SecretRef struct {
	Secret string `yaml:"secret"`
	Key    string `yaml:"key"`
}

type Domain struct {
	Host        string       `yaml:"host"`
	Attached    bool         `yaml:"attached"`
	RedirectTo  string       `yaml:"redirectTo,omitempty"`
	ListenerSet *ListenerSet `yaml:"listenerSet,omitempty"`
}

type DomainOptions struct {
	RedirectTo  string
	ListenerSet *ListenerSet
}

type ListenerSet struct {
	Enabled     bool                   `yaml:"enabled"`
	Certificate ListenerSetCertificate `yaml:"certificate"`
}

type ListenerSetCertificate struct {
	IssuerRef IssuerRef `yaml:"issuerRef"`
}

type IssuerRef struct {
	Kind string `yaml:"kind"`
	Name string `yaml:"name"`
}

type ServiceSpec struct {
	Image                string
	SourceURL            string
	ContextPath          string
	GitHubInstallationID int64
	AutoDeploy           bool
	Port                 int
	Resources            Resources
	Env                  map[string]string
	RunAsUser            *int64
	RunAsGroup           *int64
	FsGroup              *int64
}

func CreateService(env *Env, name string, spec ServiceSpec) error {
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
		annotations[annotationAutoDeploy] = strconv.FormatBool(spec.AutoDeploy)
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
		Env:            maps.Clone(spec.Env),
		RunAsUser:      spec.RunAsUser,
		RunAsGroup:     spec.RunAsGroup,
		FsGroup:        spec.FsGroup,
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

type ImageProvenance struct {
	Commit        string
	CommitMessage string
	BuildID       string
}

type ReleaseMeta struct {
	ID      string
	Trigger string
	Actor   string
}

func SetServiceImage(env *Env, name string, ref image.Ref, provenance ImageProvenance, release ReleaseMeta) error {
	return mutateService(env, name, func(s *Service) {
		s.Image.Repository = ref.Repository
		s.Image.Tag = ref.Tag
		s.Image.Digest = ref.Digest

		if s.PodAnnotations == nil {
			s.PodAnnotations = map[string]string{}
		}

		setOrDelete(s.PodAnnotations, annotationSourceCommit, provenance.Commit)
		setOrDelete(s.PodAnnotations, annotationSourceMessage, provenance.CommitMessage)
		setOrDelete(s.PodAnnotations, annotationBuildID, provenance.BuildID)

		setOrDelete(s.PodAnnotations, annotationRelease, release.ID)
		setOrDelete(s.PodAnnotations, annotationReleaseTrigger, release.Trigger)
		setOrDelete(s.PodAnnotations, annotationReleaseActor, release.Actor)
	})
}

func setOrDelete(m map[string]string, key, value string) {
	if value != "" {
		m[key] = value
	} else {
		delete(m, key)
	}
}

func SetServiceReplicas(env *Env, name string, replicas int) error {
	return mutateService(env, name, func(s *Service) {
		s.Replicas = replicas
		s.Autoscaling = nil
	})
}

func SetServiceAutoscaling(env *Env, name string, cfg Autoscaling) error {
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
	return mutateService(env, name, func(s *Service) {
		s.Command = command
	})
}

func SetServiceBranch(env *Env, name, branch string) error {
	return mutateService(env, name, func(s *Service) {
		if s.Annotations == nil {
			s.Annotations = map[string]string{}
		}

		setOrDelete(s.Annotations, annotationSourceBranch, branch)
	})
}

func SetServiceAutoDeploy(env *Env, name string, enabled bool) error {
	return mutateService(env, name, func(s *Service) {
		if s.Annotations == nil {
			s.Annotations = map[string]string{}
		}

		s.Annotations[annotationAutoDeploy] = strconv.FormatBool(enabled)
	})
}

func SetServiceCIDeploy(env *Env, name string, enabled bool) error {
	return mutateService(env, name, func(s *Service) {
		if s.Annotations == nil {
			s.Annotations = map[string]string{}
		}

		setOrDelete(s.Annotations, annotationCIDeploy, ciDeployValue(enabled))
	})
}

func ciDeployValue(enabled bool) string {
	if enabled {
		return "true"
	}
	return ""
}

func SetServicePort(env *Env, name string, port int) error {
	return mutateService(env, name, func(s *Service) {
		s.Port = port
	})
}

// SetServiceHealthCheck configures the readiness/startup probe for a service.
// A nil health check clears the config, reverting to the default TCP probe.
func SetServiceHealthCheck(env *Env, name string, healthCheck *HealthCheck) error {
	return mutateService(env, name, func(s *Service) {
		s.HealthCheck = healthCheck
	})
}

// SetServiceSecurityContext sets the run-as user/group and the volume-owning
// group (fsGroup) for a service. A nil field clears that setting, reverting to
// the image default.
func SetServiceSecurityContext(env *Env, name string, runAsUser, runAsGroup, fsGroup *int64) error {
	return mutateService(env, name, func(s *Service) {
		s.RunAsUser = runAsUser
		s.RunAsGroup = runAsGroup
		s.FsGroup = fsGroup
	})
}

// SetServiceVariables replaces a service's entire variable surface:
// literal values and secret-key references.
func SetServiceVariables(env *Env, name string, literals map[string]string, refs map[string]SecretRef) error {
	return mutateService(env, name, func(s *Service) {
		s.Env = maps.Clone(literals)
		s.Refs = maps.Clone(refs)
	})
}

func AddServiceDomain(env *Env, name, host string, options DomainOptions) error {
	return mutateService(env, name, func(s *Service) {
		if i := slices.IndexFunc(s.Domains, func(d Domain) bool { return d.Host == host }); i >= 0 {
			s.Domains[i].RedirectTo = options.RedirectTo
			s.Domains[i].ListenerSet = options.ListenerSet
			return
		}

		s.Domains = append(s.Domains, Domain{Host: host, RedirectTo: options.RedirectTo, ListenerSet: options.ListenerSet})
	})
}

func RemoveServiceDomain(env *Env, name, host string) error {
	svc, ok := env.Services[name]

	if !ok {
		return fmt.Errorf("service %q not found", name)
	}

	if i := slices.IndexFunc(svc.Domains, func(d Domain) bool { return d.RedirectTo == host }); i >= 0 {
		return fmt.Errorf("%s redirects to %s; remove %s first", svc.Domains[i].Host, host, svc.Domains[i].Host)
	}

	return mutateService(env, name, func(s *Service) {
		s.Domains = slices.DeleteFunc(s.Domains, func(d Domain) bool {
			return d.Host == host
		})
	})
}

func AttachServiceDomain(env *Env, name, host string, attached bool) error {
	return mutateService(env, name, func(s *Service) {
		i := slices.IndexFunc(s.Domains, func(d Domain) bool { return d.Host == host })

		if i >= 0 {
			s.Domains[i].Attached = attached
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
