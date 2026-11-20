package resources

import (
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/zeitlos/lucity/pkg/deployer/values"
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
	DefaultStorageQuota = resource.MustParse("512Gi")
)

const (
	DefaultBackupRetention = "30d"
	DefaultBackupSchedule  = "0 0 2 * * 0"
)

// Request returns the K8s request value for the given limit: the limit scaled
// down by a fixed ratio, so workloads are burstable.
func Request(limit resource.Quantity) *resource.Quantity {
	milli := limit.MilliValue()
	scaled := int64(float64(milli) * burstableRequestRatio)

	return resource.NewMilliQuantity(scaled, limit.Format)
}
