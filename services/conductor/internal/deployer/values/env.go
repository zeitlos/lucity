package values

import (
	"maps"
)

type Env struct {
	Suspended             bool                  `yaml:"suspended"`
	CommonLabels          map[string]string     `yaml:"commonLabels,omitempty"`
	CommonAnnotations     map[string]string     `yaml:"commonAnnotations,omitempty"`
	ImagePullSecrets      []PullSecret          `yaml:"imagePullSecrets,omitempty"`
	Services              map[string]Service    `yaml:"services"`
	SharedVariables       map[string]string     `yaml:"sharedVariables"`
	SharedVariableLabels  map[string]string     `yaml:"sharedVariableLabels,omitempty"`
	Databases             Databases             `yaml:"databases"`
	Volumes               map[string]Volume     `yaml:"volumes,omitempty"`
	Gateway               Gateway               `yaml:"gateway"`
	VerticalPodAutoscaler VerticalPodAutoscaler `yaml:"verticalPodAutoscaler"`
}

type VerticalPodAutoscaler struct {
	Enabled      bool         `yaml:"enabled"`
	UpdatePolicy UpdatePolicy `yaml:"updatePolicy"`
}

type UpdatePolicy struct {
	UpdateMode UpdateMode `yaml:"updateMode"`
}

type UpdateMode string

const (
	UpdateModeOff               UpdateMode = "Off"
	UpdateModeInitial           UpdateMode = "Initial"
	UpdateModeInPlaceOrRecreate UpdateMode = "InPlaceOrRecreate"
)

type Gateway struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

type PullSecret struct {
	Name string `yaml:"name"`
}

func SetSuspended(env *Env, suspended bool) error {
	env.Suspended = suspended
	return nil
}

func SetEnvironmentVariables(env *Env, vars map[string]string) error {
	env.SharedVariables = maps.Clone(vars)

	return nil
}
