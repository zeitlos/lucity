package values

import "fmt"

type Env struct {
	Suspended         bool               `yaml:"suspended"`
	CommonLabels      map[string]string  `yaml:"commonLabels,omitempty"`
	CommonAnnotations map[string]string  `yaml:"commonAnnotations,omitempty"`
	ImagePullSecrets  []PullSecret       `yaml:"imagePullSecrets,omitempty"`
	Services          map[string]Service `yaml:"services"`
	SharedVariables   map[string]string  `yaml:"sharedVariables"`
	Databases         Databases          `yaml:"databases"`
	Gateway           Gateway            `yaml:"gateway"`
}

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
	for k := range vars {
		if !isValidVarName(k) {
			return fmt.Errorf("invalid variable name %q", k)
		}
	}

	env.SharedVariables = cloneStringMap(vars)

	return nil
}

func cloneStringMap(in map[string]string) map[string]string {
	if in == nil {
		return map[string]string{}
	}

	out := make(map[string]string, len(in))

	for k, v := range in {
		out[k] = v
	}

	return out
}
