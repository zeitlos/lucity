package values

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

type Databases struct {
	Postgres map[string]Postgres `yaml:"postgres"`
	Valkey   map[string]Valkey   `yaml:"valkey"`
}

type Postgres struct {
	Instances   int               `yaml:"instances,omitempty"`
	Size        string            `yaml:"size,omitempty"`
	Version     string            `yaml:"version,omitempty"`
	PublicHost  string            `yaml:"publicHost,omitempty"`
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

func ExposeDatabase(env *Env, name, host string) error {
	if !isValidHostname(host) {
		return fmt.Errorf("invalid hostname %q", host)
	}

	return mutateDatabase(env, name, func(p *Postgres) {
		p.PublicHost = host

		if p.Annotations == nil {
			p.Annotations = map[string]string{}
		}

		p.Annotations[annotationDatabaseHost] = host
	})
}

func UnexposeDatabase(env *Env, name string) error {
	return mutateDatabase(env, name, func(p *Postgres) {
		p.PublicHost = ""
		delete(p.Annotations, annotationDatabaseHost)
	})
}

func mutateDatabase(env *Env, name string, mutate func(*Postgres)) error {
	postgres, ok := env.Databases.Postgres[name]

	if !ok {
		return fmt.Errorf("database %q not found", name)
	}

	mutate(&postgres)
	env.Databases.Postgres[name] = postgres

	return nil
}
