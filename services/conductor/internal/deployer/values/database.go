package values

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

type Databases struct {
	Postgres map[string]Postgres `yaml:"postgres"`
	Redis    map[string]Redis    `yaml:"redis"`
}

type Postgres struct {
	Instances int       `yaml:"instances,omitempty"`
	Size      string    `yaml:"size,omitempty"`
	Version   string    `yaml:"version,omitempty"`
	Resources Resources `yaml:"resources,omitempty"`
}

type Redis struct {
	Image     string    `yaml:"image,omitempty"`
	Resources Resources `yaml:"resources,omitempty"`
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
		return fmt.Errorf("database %q already exists", name)
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
