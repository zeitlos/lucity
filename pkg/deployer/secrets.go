package deployer

// Names of the per-environment Secrets the platform creates and the chart
// values reference.
const (
	PullSecretName           = "lucity-registry"
	BackupSecretName         = "lucity-backups"
	BackupAccessKeyIDKey     = "accessKeyId"
	BackupSecretAccessKeyKey = "secretAccessKey"
	BackupRegionKey          = "region"
)

// The header the edge CDN adds to every request, and the Secret holding its
// value, which workload routes match on when the edge header is enforced.
const (
	EdgeHeaderName       = "X-Lucity-Edge"
	EdgeHeaderSecretName = "lucity-edge"
	EdgeHeaderSecretKey  = "value"
)
