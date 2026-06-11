package values

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

type Databases struct {
	Postgres map[string]Postgres `yaml:"postgres"`
}

type Postgres struct {
	Instances   int               `yaml:"instances,omitempty"`
	Size        string            `yaml:"size,omitempty"`
	Version     string            `yaml:"version,omitempty"`
	Resources   Resources         `yaml:"resources,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type DatabaseSpec struct {
	Version   string
	Instances int
	Size      resource.Quantity
}

func CreateDatabase(env *Env, name string, spec DatabaseSpec) error {
	if !isValidName(name) {
		return fmt.Errorf("invalid database name %q", name)
	}

	if _, ok := env.Databases.Postgres[name]; ok {
		// To keep this function idempotent, don't return an error if the database already exists.
		return nil
	}

	if env.Databases.Postgres == nil {
		env.Databases.Postgres = map[string]Postgres{}
	}

	instances := spec.Instances

	if instances == 0 {
		instances = 1
	}

	env.Databases.Postgres[name] = Postgres{
		Version:   spec.Version,
		Instances: instances,
		Size:      spec.Size.String(),
		Labels:    map[string]string{labelDatabase: name},
	}

	return nil
}

func DeleteDatabase(env *Env, name string) error {
	if _, ok := env.Databases.Postgres[name]; !ok {
		return fmt.Errorf("database %q not found", name)
	}

	delete(env.Databases.Postgres, name)

	return nil
}
