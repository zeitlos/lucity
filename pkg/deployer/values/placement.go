package values

import (
	"bytes"
	"encoding/json"
)

// Placement holds the scheduling constraints applied to every workload pod
// of an environment.
type Placement struct {
	NodeSelector map[string]string `yaml:"nodeSelector,omitempty" json:"nodeSelector,omitempty"`
	Tolerations  []map[string]any  `yaml:"tolerations,omitempty" json:"tolerations,omitempty"`
	Affinity     map[string]any    `yaml:"affinity,omitempty" json:"affinity,omitempty"`
}

// Decode parses a placement from its JSON form, so envconfig can read it
// from a single variable.
func (p *Placement) Decode(value string) error {
	if value == "" {
		return nil
	}

	return json.Unmarshal([]byte(value), p)
}

// Equal compares two placements by their JSON form, so values read back from
// a release compare equal to the same values decoded from configuration.
func (p Placement) Equal(other Placement) bool {
	a, errA := json.Marshal(p)
	b, errB := json.Marshal(other)

	return errA == nil && errB == nil && bytes.Equal(a, b)
}

func (p Placement) IsZero() bool {
	return len(p.NodeSelector) == 0 && len(p.Tolerations) == 0 && len(p.Affinity) == 0
}

func SetPlacement(env *Env, placement Placement) error {
	env.Placement = placement
	return nil
}

func StripPlacement(env *Env) {
	env.Placement = Placement{}
}
