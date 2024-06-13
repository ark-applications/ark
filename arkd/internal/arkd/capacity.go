package arkd

import (
	"context"
	"runtime"
)

type SystemMetrics struct {
	// the available system cpu count
	TotalCpu int `json:"total_cpu"`
	// TotalMem     uint64  `json:"total_mem"`

	// the total count of tasks (includes running and suspended)
	TotalTasks int `json:"total_tasks"`
	// allocated cpu's
	AllocatedCpu float64 `json:"allocated_cpu"`
	// allocated memory
	AllocatedMem int `json:"allocated_mem"`
	// available cpu's (total - allocated)
	AvailableCpu float64 `json:"available_cpu"`
}

func GetSystemMetrics(ctx context.Context, ts *TaskStore) SystemMetrics {
	aggTaskMetrics := ts.AggMetrics(ctx)

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	numCpu := runtime.NumCPU()

	return SystemMetrics{
		TotalCpu:     numCpu,
		TotalTasks:   aggTaskMetrics.TotalTasks,
		AllocatedMem: aggTaskMetrics.AllocatedMem,
		AllocatedCpu: aggTaskMetrics.AllocatedCpu,
		AvailableCpu: float64(numCpu) - aggTaskMetrics.AllocatedCpu,
	}
}
