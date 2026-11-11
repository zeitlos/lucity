package values

import "gopkg.in/yaml.v3"

// StartCommand is the argv a service starts with, run as is without a shell.
// Values written before it became a list hold a shell string, which still
// runs through /bin/sh -c.
type StartCommand []string

func (c *StartCommand) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		var line string

		if err := node.Decode(&line); err != nil {
			return err
		}

		if line != "" {
			*c = StartCommand{"/bin/sh", "-c", line}
		}

		return nil
	}

	var words []string

	if err := node.Decode(&words); err != nil {
		return err
	}

	*c = words

	return nil
}
