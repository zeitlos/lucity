package labels

import "fmt"

// Label key constants for Kubernetes resource discovery.
const (
	Prefix = "lucity.dev/"

	Project     = Prefix + "project"
	Environment = Prefix + "environment"
	Ephemeral   = Prefix + "ephemeral"
	Service     = Prefix + "service"
	BuiltBy     = Prefix + "built-by"
)

// OCI image label constants.
const (
	OCISource   = "org.opencontainers.image.source"
	OCIRevision = "org.opencontainers.image.revision"
)

// Values for well-known label values.
const (
	BuiltByBuilder = "lucity-builder"
)

// Selector returns a Kubernetes label selector string for the given key-value pair.
func Selector(key, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}
