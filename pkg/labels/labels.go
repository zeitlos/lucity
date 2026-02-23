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
	ManagedBy   = Prefix + "managed-by"
)

// OCI image label constants.
const (
	OCISource   = "org.opencontainers.image.source"
	OCIRevision = "org.opencontainers.image.revision"
)

// Values for well-known label values.
const (
	BuiltByBuilder  = "lucity-builder"
	ManagedByLucity = "lucity"
)

// Selector returns a Kubernetes label selector string for the given key-value pair.
func Selector(key, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}

// ShortName returns the project name. Kept for backward compatibility with
// callers that used the old org/name format.
func ShortName(project string) string {
	return project
}

// NamespaceFor derives the K8s namespace from a project and environment name.
// "warm-wren" + "production" → "warm-wren-production"
func NamespaceFor(project, environment string) string {
	return project + "-" + environment
}
