package values

import "fmt"

type Env struct {
	Suspended       bool                         `yaml:"suspended"`
	Workspace       string                       `yaml:"workspace"`
	Project         string                       `yaml:"project"`
	Environment     string                       `yaml:"environment"`
	Services        map[string]Service           `yaml:"services"`
	CronJobs        map[string]CronJob           `yaml:"cronJobs"`
	SharedVariables map[string]string            `yaml:"sharedVariables"`
	Config          map[string]map[string]string `yaml:"config"`
	Databases       Databases                    `yaml:"databases"`
	Gateway         Gateway                      `yaml:"gateway"`
}

type Gateway struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

type CronJob struct {
	Schedule  string    `yaml:"schedule"`
	Image     ImageRef  `yaml:"image"`
	Command   []string  `yaml:"command,omitempty"`
	Resources Resources `yaml:"resources,omitempty"`
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
