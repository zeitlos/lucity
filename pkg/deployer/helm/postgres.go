package helm

import (
	"fmt"
	"math"
	"strconv"

	"k8s.io/apimachinery/pkg/api/resource"
)

func postgresParameters(cpuLimit, memoryLimit resource.Quantity) map[string]string {
	memMB := memoryLimit.Value() / (1024 * 1024)

	cores := int(math.Ceil(float64(cpuLimit.MilliValue()) / 1000.0))
	if cores < 1 {
		cores = 1
	}

	return map[string]string{
		"shared_buffers":                   fmt.Sprintf("%dMB", memMB/4),
		"effective_cache_size":             fmt.Sprintf("%dMB", memMB*3/4),
		"maintenance_work_mem":             fmt.Sprintf("%dMB", clampInt64(memMB/16, 64, 2048)),
		"work_mem":                         fmt.Sprintf("%dMB", clampInt64(memMB/256, 4, 1024)),
		"max_worker_processes":             strconv.Itoa(max(8, cores*2)),
		"max_parallel_workers":             strconv.Itoa(max(2, cores)),
		"max_parallel_workers_per_gather":  strconv.FormatInt(clampInt64(int64(cores)/2, 1, 4), 10),
		"max_parallel_maintenance_workers": strconv.FormatInt(clampInt64(int64(cores)/2, 1, 4), 10),
	}
}

func clampInt64(v, lo, hi int64) int64 {
	if v < lo {
		return lo
	}

	if v > hi {
		return hi
	}

	return v
}
