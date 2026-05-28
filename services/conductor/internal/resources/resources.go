package resources

import (
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

const burstableRequestRatio = 0.5

// Request returns the K8s request value for the given limit under the given
// tier. Guaranteed (production) returns the limit unchanged; burstable (eco)
// scales the limit down by a fixed ratio.
func Request(tier platform.ResourceTier, limit resource.Quantity) resource.Quantity {
	if tier == platform.ProductionTier {
		return limit.DeepCopy()
	}

	milli := limit.MilliValue()
	scaled := int64(float64(milli) * burstableRequestRatio)

	return *resource.NewMilliQuantity(scaled, limit.Format)
}
