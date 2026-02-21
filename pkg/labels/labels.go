package labels

import (
	"fmt"
	"strings"
)

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

// SplitProject splits "org/name" into org and name.
func SplitProject(project string) (org, name string, err error) {
	parts := strings.SplitN(project, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid project name %q: must be org/name", project)
	}
	return parts[0], parts[1], nil
}

// ShortName extracts the short name from a project identifier.
// "zeitlos/myapp" → "myapp", "myapp" → "myapp"
func ShortName(project string) string {
	_, name, err := SplitProject(project)
	if err != nil {
		return project
	}
	return name
}

// NamespaceFor derives the K8s namespace from a project and environment name.
// "zeitlos/myapp" + "production" → "myapp-production"
func NamespaceFor(project, environment string) string {
	return ShortName(project) + "-" + environment
}
