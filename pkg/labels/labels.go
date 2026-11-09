package labels

import "fmt"

// Label key constants for Kubernetes resource discovery.
const (
	Prefix = "lucity.dev/"

	Workspace           = Prefix + "workspace"
	Project             = Prefix + "project"
	Environment         = Prefix + "environment"
	Ephemeral           = Prefix + "ephemeral"
	Service             = Prefix + "service"
	Release             = Prefix + "release"
	BuiltBy             = Prefix + "built-by"
	ManagedBy           = "app.kubernetes.io/managed-by"
	ResourceType        = Prefix + "resource-type"
	GitHubInstallation  = Prefix + "github-installation"
	ObjectStorageBucket = Prefix + "object-storage-bucket"
	CustomDomain        = Prefix + "custom-domain"
	Agent               = Prefix + "agent"
	AgentKey            = Prefix + "agent-key"
	ChatApp             = Prefix + "chat-app"
	AgentChannels       = Prefix + "agent-channels"
)

// Annotation keys for readable names. A resource without one is named by its id.
const (
	Name            = Prefix + "name"
	ProjectName     = Prefix + "project-name"
	EnvironmentName = Prefix + "environment-name"
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
