package composition

import (
	"context"

	runtimeapp "ctf-platform/internal/module/runtime/application"
	runtimedomain "ctf-platform/internal/module/runtime/domain"
	runtimeinfra "ctf-platform/internal/module/runtimeinfra"
)

type runtimeManagedContainerStatsReader struct {
	engine *runtimeinfra.Engine
}

func newRuntimeManagedContainerStatsReader(engine *runtimeinfra.Engine) *runtimeManagedContainerStatsReader {
	return &runtimeManagedContainerStatsReader{engine: engine}
}

func (r *runtimeManagedContainerStatsReader) ListManagedContainerStats(ctx context.Context) ([]runtimeapp.ManagedContainerStat, error) {
	if r == nil || r.engine == nil {
		return []runtimeapp.ManagedContainerStat{}, nil
	}

	stats, err := r.engine.ListManagedContainerStats(ctx, runtimedomain.ManagedByFilter())
	if err != nil {
		return nil, err
	}

	result := make([]runtimeapp.ManagedContainerStat, 0, len(stats))
	for _, item := range stats {
		result = append(result, runtimeapp.ManagedContainerStat{
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

type systemRuntimeStatsProvider struct {
	service *runtimeapp.ContainerStatsService
}

func newSystemRuntimeStatsProvider(service *runtimeapp.ContainerStatsService) *systemRuntimeStatsProvider {
	return &systemRuntimeStatsProvider{service: service}
}

func (p *systemRuntimeStatsProvider) ListManagedContainerStats(ctx context.Context) ([]runtimeinfra.ManagedContainerStat, error) {
	if p == nil || p.service == nil {
		return []runtimeinfra.ManagedContainerStat{}, nil
	}

	stats, err := p.service.ListManagedContainerStats(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]runtimeinfra.ManagedContainerStat, 0, len(stats))
	for _, item := range stats {
		result = append(result, runtimeinfra.ManagedContainerStat{
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
