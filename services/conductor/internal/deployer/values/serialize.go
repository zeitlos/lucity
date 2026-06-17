package values

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

func New() *Env {
	return &Env{
		CommonLabels:      map[string]string{},
		CommonAnnotations: map[string]string{},
		Services:          map[string]Service{},
		SharedVariables:   map[string]string{},
		Config:            map[string]map[string]string{},
		Databases: Databases{
			Postgres: map[string]Postgres{},
		},
	}
}

func Parse(data []byte) (*Env, error) {
	env := New()

	if err := yaml.Unmarshal(data, env); err != nil {
		return nil, fmt.Errorf("parse values: %w", err)
	}

	return env, nil
}

func Marshal(env *Env) ([]byte, error) {
	var buf bytes.Buffer

	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)

	if err := enc.Encode(env); err != nil {
		return nil, fmt.Errorf("marshal values: %w", err)
	}

	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("close yaml encoder: %w", err)
	}

	return buf.Bytes(), nil
}
