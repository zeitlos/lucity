package values

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

type Valkey struct {
	Size        string            `yaml:"size,omitempty"`
	Version     string            `yaml:"version,omitempty"`
	Password    string            `yaml:"password,omitempty"`
	Resources   Resources         `yaml:"resources,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type KeyValueStoreSpec struct {
	Version  string
	Size     resource.Quantity
	Password string
}

func CreateKeyValueStore(env *Env, name string, spec KeyValueStoreSpec) error {
	if !isValidName(name) {
		return fmt.Errorf("invalid key-value store name %q", name)
	}

	if _, ok := env.Databases.Valkey[name]; ok {
		// To keep this function idempotent, don't return an error if the store already exists.
		return nil
	}

	if env.Databases.Valkey == nil {
		env.Databases.Valkey = map[string]Valkey{}
	}

	env.Databases.Valkey[name] = Valkey{
		Version:  spec.Version,
		Size:     spec.Size.String(),
		Password: spec.Password,
		Labels:   map[string]string{labelKeyValueStore: name},
	}

	return nil
}

func DeleteKeyValueStore(env *Env, name string) error {
	if _, ok := env.Databases.Valkey[name]; !ok {
		return fmt.Errorf("key-value store %q not found", name)
	}

	delete(env.Databases.Valkey, name)

	return nil
}
