package infrastructure

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"

	runtimedomain "ctf-platform/internal/module/runtime/domain"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

func (e *Engine) InspectImageSize(ctx context.Context, imageRef string) (int64, error) {
	inspect, _, err := e.cli.ImageInspectWithRaw(ctx, imageRef)
	if err != nil {
		return 0, err
	}
	return inspect.Size, nil
}

func (e *Engine) RemoveImage(ctx context.Context, imageRef string) error {
	if imageRef == "" {
		return nil
	}
	_, err := e.cli.ImageRemove(ctx, imageRef, image.RemoveOptions{})
	return err
}

func (e *Engine) ListManagedContainerStats(ctx context.Context) ([]runtimeports.ManagedContainerStat, error) {
	containers, err := e.cli.ContainerList(ctx, containertypes.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", runtimedomain.ProjectFilter()),
			filters.Arg("label", runtimedomain.ManagedByFilter()),
		),
	})
	if err != nil {
		return nil, err
	}

	return collectManagedContainerStats(ctx, containers, func(ctx context.Context, containerSummary types.Container) (runtimeports.ManagedContainerStat, error) {
		stat, err := e.cli.ContainerStats(ctx, containerSummary.ID, false)
		if err != nil {
			return runtimeports.ManagedContainerStat{}, err
		}
		defer stat.Body.Close()

		var payload types.StatsJSON
		if err := json.NewDecoder(stat.Body).Decode(&payload); err != nil {
			return runtimeports.ManagedContainerStat{}, err
		}

		containerName := shortContainerID(containerSummary.ID)
		if len(containerSummary.Names) > 0 {
			containerName = strings.TrimPrefix(containerSummary.Names[0], "/")
		}

		return runtimeports.ManagedContainerStat{
			ContainerID:   shortContainerID(containerSummary.ID),
			ContainerName: containerName,
			CPUPercent:    calculateCPUPercent(&payload),
			MemoryPercent: calculateMemoryPercent(&payload),
			MemoryUsage:   int64(payload.MemoryStats.Usage),
			MemoryLimit:   int64(payload.MemoryStats.Limit),
		}, nil
	}), nil
}

func collectManagedContainerStats(
	ctx context.Context,
	containers []types.Container,
	fetch func(context.Context, types.Container) (runtimeports.ManagedContainerStat, error),
) []runtimeports.ManagedContainerStat {
	if len(containers) == 0 {
		return []runtimeports.ManagedContainerStat{}
	}

	stats := make([]runtimeports.ManagedContainerStat, len(containers))
	ok := make([]bool, len(containers))
	var (
		wg  sync.WaitGroup
		sem = make(chan struct{}, 8)
	)

	for idx, item := range containers {
		wg.Add(1)
		sem <- struct{}{}
		go func(index int, containerSummary types.Container) {
			defer wg.Done()
			defer func() { <-sem }()

			stat, err := fetch(ctx, containerSummary)
			if err != nil {
				return
			}
			stats[index] = stat
			ok[index] = true
		}(idx, item)
	}
	wg.Wait()

	result := make([]runtimeports.ManagedContainerStat, 0, len(containers))
	for idx, item := range stats {
		if !ok[idx] {
			continue
		}
		result = append(result, item)
	}
	return result
}

func shortContainerID(id string) string {
	if len(id) <= 12 {
		return id
	}
	return id[:12]
}

func calculateCPUPercent(stats *types.StatsJSON) float64 {
	if stats.PreCPUStats.CPUUsage.TotalUsage == 0 {
		return 0
	}
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
	if systemDelta <= 0 || cpuDelta < 0 {
		return 0
	}
	numCPU := float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
	if numCPU == 0 {
		numCPU = 1
	}
	return (cpuDelta / systemDelta) * numCPU * 100.0
}

func calculateMemoryPercent(stats *types.StatsJSON) float64 {
	if stats.MemoryStats.Limit > 0 {
		return float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit) * 100.0
	}
	return 0.0
}
