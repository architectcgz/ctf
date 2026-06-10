package composition

import (
	"context"

	runtimeports "ctf-platform/internal/module/container_runtime/ports"
	instanceports "ctf-platform/internal/module/instance/ports"
	opsports "ctf-platform/internal/module/ops/ports"
)

type opsRuntimeQueryAdapter struct {
	query instanceports.RunningInstanceCountRepository
}

func newOpsRuntimeQueryAdapter(query instanceports.RunningInstanceCountRepository) opsports.RuntimeQuery {
	return &opsRuntimeQueryAdapter{query: query}
}

func (a *opsRuntimeQueryAdapter) CountRunning(ctx context.Context) (int64, error) {
	if a == nil || a.query == nil {
		return 0, nil
	}
	return a.query.CountRunningInstances(ctx)
}

type opsRuntimeStatsProviderAdapter struct {
	stats runtimeports.ManagedContainerStatsReader
}

func newOpsRuntimeStatsProviderAdapter(stats runtimeports.ManagedContainerStatsReader) opsports.RuntimeStatsProvider {
	return &opsRuntimeStatsProviderAdapter{stats: stats}
}

func (a *opsRuntimeStatsProviderAdapter) ListManagedContainerStats(ctx context.Context) ([]opsports.ManagedContainerStat, error) {
	if a == nil || a.stats == nil {
		return []opsports.ManagedContainerStat{}, nil
	}

	stats, err := a.stats.ListManagedContainerStats(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]opsports.ManagedContainerStat, 0, len(stats))
	for _, item := range stats {
		result = append(result, opsports.ManagedContainerStat{
			ContainerID:   item.ContainerID,
			ContainerName: item.ContainerName,
			CPUPercent:    item.CPUPercent,
			MemoryPercent: item.MemoryPercent,
			MemoryUsage:   item.MemoryUsage,
			MemoryLimit:   item.MemoryLimit,
		})
	}
	return result, nil
}
