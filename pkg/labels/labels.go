package labels

import "fmt"

// Label key constants for Kubernetes resource discovery.
const (
	Prefix = "lucity.dev/"

	Workspace    = Prefix + "workspace"
	Project      = Prefix + "project"
	Environment  = Prefix + "environment"
	Ephemeral    = Prefix + "ephemeral"
	Service      = Prefix + "service"
	BuiltBy      = Prefix + "built-by"
	ManagedBy    = Prefix + "managed-by"
	ResourceType = Prefix + "resource-type"
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

// Well-known namespace.
const LucityNamespace = "lucity-system"

// Selector returns a Kubernetes label selector string for the given key-value pair.
func Selector(key, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}

// NamespaceFor derives the K8s namespace from workspace, project, and environment.
// "acme" + "api" + "production" → "acme-api-production"
func NamespaceFor(workspace, project, environment string) string {
	return workspace + "-" + project + "-" + environment
}
