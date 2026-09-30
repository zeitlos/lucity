package resources

import (
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/zeitlos/lucity/services/conductor/internal/deployer/values"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

const (
	// TODO: This static ratio will be replaced by VPA at some point.
	burstableRequestRatio = 0.5

	VerticalPodAutoscalerUpdateMode = values.UpdateModeInPlaceOrRecreate
)

var (
	DefaultCPULimit    = resource.MustParse("500m")
	DefaultMemoryLimit = resource.MustParse("512Mi")

	DefaultCPUQuota     = resource.MustParse("32")
	DefaultMemoryQuota  = resource.MustParse("64Gi")
	DefaultStorageQuota = resource.MustParse("4Ti")

	MaxDatabaseStorage = resource.MustParse("1Ti")
)

const (
	DefaultBackupRetention = "30d"
	DefaultBackupSchedule  = "0 0 2 * * 0"
)

// Request returns the K8s request value for the given limit under the given
// tier. Guaranteed (production) returns the limit unchanged; burstable (eco)
// scales the limit down by a fixed ratio.
func Request(tier platform.ResourceTier, limit resource.Quantity) *resource.Quantity {
	if tier == platform.ProductionTier {
		return new(limit.DeepCopy())
	}

	milli := limit.MilliValue()
	scaled := int64(float64(milli) * burstableRequestRatio)

	return resource.NewMilliQuantity(scaled, limit.Format)
}
