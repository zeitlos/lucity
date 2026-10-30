package billing

import "k8s.io/apimachinery/pkg/api/resource"

// Unlimited marks a count without a ceiling.
const Unlimited = -1

// Limits is what a plan allows: per instance, per disk and per workspace.
type Limits struct {
	CPU               resource.Quantity
	Memory            resource.Quantity
	Replicas          int
	Volumes           bool
	VolumeSize        resource.Quantity
	DatabaseStorage   resource.Quantity
	KeyValueStoreSize resource.Quantity
	Members           int
	CustomDomains     int
}

// Plans lists every plan from the free tier up, in upgrade order.
var Plans = []Plan{PlanNone, PlanHobby, PlanPro}

// InstanceLimits apply on an instance without billing.
var InstanceLimits = Limits{
	CPU:               resource.MustParse("4"),
	Memory:            resource.MustParse("16Gi"),
	Replicas:          10,
	Volumes:           true,
	VolumeSize:        resource.MustParse("128Gi"),
	DatabaseStorage:   resource.MustParse("128Gi"),
	KeyValueStoreSize: resource.MustParse("128Gi"),
	Members:           Unlimited,
	CustomDomains:     Unlimited,
}

// Limits returns what the plan allows. PlanNone is the free tier.
func (p Plan) Limits() Limits {
	switch p {
	case PlanPro:
		return InstanceLimits
	case PlanHobby:
		return Limits{
			CPU:               resource.MustParse("2"),
			Memory:            resource.MustParse("4Gi"),
			Replicas:          3,
			Volumes:           true,
			VolumeSize:        resource.MustParse("32Gi"),
			DatabaseStorage:   resource.MustParse("32Gi"),
			KeyValueStoreSize: resource.MustParse("32Gi"),
			Members:           1,
			CustomDomains:     1,
		}
	}

	return Limits{
		CPU:               resource.MustParse("1"),
		Memory:            resource.MustParse("2Gi"),
		Replicas:          1,
		DatabaseStorage:   resource.MustParse("10Gi"),
		KeyValueStoreSize: resource.MustParse("1Gi"),
		Members:           1,
	}
}

// Title is the plan's name as people read it.
func (p Plan) Title() string {
	switch p {
	case PlanPro:
		return "Pro"
	case PlanHobby:
		return "Hobby"
	}

	return "Trial"
}
