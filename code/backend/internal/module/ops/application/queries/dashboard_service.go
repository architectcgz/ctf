package queries

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	opsports "ctf-platform/internal/module/ops/ports"
)

type DashboardService struct {
	runtimeQuery opsports.RuntimeQuery
	runtime      opsports.RuntimeStatsProvider
	state        opsports.DashboardStateStore
	config       *config.Config
	logger       *zap.Logger
}

func NewDashboardService(
	runtimeQuery opsports.RuntimeQuery,
	runtimeStats opsports.RuntimeStatsProvider,
	state opsports.DashboardStateStore,
	cfg *config.Config,
	logger *zap.Logger,
) *DashboardService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &DashboardService{
		runtimeQuery: runtimeQuery,
		runtime:      runtimeStats,
		state:        state,
		config:       cfg,
		logger:       logger,
	}
}

func (s *DashboardService) GetDashboardStats(ctx context.Context) (*dto.DashboardStats, error) {
	cached, err := s.getFromCache(ctx)
	if err == nil && cached != nil {
		return cached, nil
	}
	if err != nil {
		s.logger.Error("读取仪表盘缓存失败，降级到实时查询", zap.Error(err))
	}

	stats := &dto.DashboardStats{
		ContainerStats: []dto.ContainerStat{},
		Alerts:         []dto.ResourceAlert{},
	}

	onlineUsers, err := s.countOnlineUsers(ctx)
	if err != nil {
		s.logger.Error("统计在线用户失败", zap.Error(err))
		onlineUsers = -1
	}
	stats.OnlineUsers = onlineUsers

	activeContainers := int64(0)
	if s.runtimeQuery != nil {
		activeContainers, err = s.runtimeQuery.CountRunning(ctx)
		if err != nil {
			return nil, fmt.Errorf("统计活跃容器失败: %w", err)
		}
	}
	stats.ActiveContainers = activeContainers

	containerStats, err := s.getContainerStats(ctx)
	if err != nil {
		s.logger.Error("获取容器统计失败", zap.Error(err))
	} else {
		stats.ContainerStats = containerStats
		stats.CPUUsage, stats.MemoryUsage = s.calculateAverageUsage(containerStats)
		stats.Alerts = s.checkAlerts(containerStats)
	}

	s.saveToCache(ctx, stats)
	return stats, nil
}

func (s *DashboardService) getContainerStats(ctx context.Context) ([]dto.ContainerStat, error) {
	if s.runtime == nil {
		return []dto.ContainerStat{}, nil
	}
	stats, err := s.runtime.ListManagedContainerStats(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.ContainerStat, 0, len(stats))
	for _, stat := range stats {
		result = append(result, dto.ContainerStat{
			ContainerID:   stat.ContainerID,
			ContainerName: stat.ContainerName,
			CPUPercent:    stat.CPUPercent,
			MemoryPercent: stat.MemoryPercent,
			MemoryUsage:   stat.MemoryUsage,
			MemoryLimit:   stat.MemoryLimit,
		})
	}
	return result, nil
}

func (s *DashboardService) calculateAverageUsage(stats []dto.ContainerStat) (float64, float64) {
	if len(stats) == 0 {
		return 0, 0
	}
	var totalCPU, totalMem float64
	for _, stat := range stats {
		totalCPU += stat.CPUPercent
		totalMem += stat.MemoryPercent
	}
	return totalCPU / float64(len(stats)), totalMem / float64(len(stats))
}

func (s *DashboardService) checkAlerts(stats []dto.ContainerStat) []dto.ResourceAlert {
	alerts := []dto.ResourceAlert{}
	threshold := s.config.Dashboard.AlertThreshold
	for _, stat := range stats {
		if stat.CPUPercent > threshold {
			alerts = append(alerts, dto.ResourceAlert{
				ContainerID: stat.ContainerID,
				Type:        "cpu",
				Value:       stat.CPUPercent,
				Threshold:   threshold,
				Message:     fmt.Sprintf("容器 %s CPU 使用率过高: %.2f%%", stat.ContainerName, stat.CPUPercent),
			})
		}
		if stat.MemoryPercent > threshold {
			alerts = append(alerts, dto.ResourceAlert{
				ContainerID: stat.ContainerID,
				Type:        "memory",
				Value:       stat.MemoryPercent,
				Threshold:   threshold,
				Message:     fmt.Sprintf("容器 %s 内存使用率过高: %.2f%%", stat.ContainerName, stat.MemoryPercent),
			})
		}
	}
	return alerts
}

func (s *DashboardService) countOnlineUsers(ctx context.Context) (int64, error) {
	if s.state == nil {
		return 0, nil
	}
	return s.state.CountOnlineUsers(ctx)
}

func (s *DashboardService) getFromCache(ctx context.Context) (*dto.DashboardStats, error) {
	if s.state == nil {
		return nil, nil
	}
	snapshot, err := s.state.LoadDashboardStats(ctx)
	if err != nil {
		return nil, err
	}
	if snapshot == nil {
		return nil, nil
	}
	return dashboardStatsFromSnapshot(snapshot), nil
}

func (s *DashboardService) saveToCache(ctx context.Context, stats *dto.DashboardStats) {
	if s.state == nil || stats == nil {
		return
	}
	if err := s.state.SaveDashboardStats(ctx, dashboardSnapshotFromStats(stats)); err != nil {
		s.logger.Error("缓存统计数据失败", zap.Error(err))
		return
	}
	s.logger.Debug("仪表盘缓存已更新")
}

func dashboardStatsFromSnapshot(snapshot *opsports.DashboardStatsSnapshot) *dto.DashboardStats {
	if snapshot == nil {
		return nil
	}

	containerStats := make([]dto.ContainerStat, 0, len(snapshot.ContainerStats))
	for _, item := range snapshot.ContainerStats {
		containerStats = append(containerStats, dto.ContainerStat{
			ContainerID:   item.ContainerID,
			ContainerName: item.ContainerName,
			CPUPercent:    item.CPUPercent,
			MemoryPercent: item.MemoryPercent,
			MemoryUsage:   item.MemoryUsage,
			MemoryLimit:   item.MemoryLimit,
		})
	}

	alerts := make([]dto.ResourceAlert, 0, len(snapshot.Alerts))
	for _, item := range snapshot.Alerts {
		alerts = append(alerts, dto.ResourceAlert{
			ContainerID: item.ContainerID,
			Type:        item.Type,
			Value:       item.Value,
			Threshold:   item.Threshold,
			Message:     item.Message,
		})
	}

	return &dto.DashboardStats{
		OnlineUsers:      snapshot.OnlineUsers,
		ActiveContainers: snapshot.ActiveContainers,
		CPUUsage:         snapshot.CPUUsage,
		MemoryUsage:      snapshot.MemoryUsage,
		ContainerStats:   containerStats,
		Alerts:           alerts,
	}
}

func dashboardSnapshotFromStats(stats *dto.DashboardStats) *opsports.DashboardStatsSnapshot {
	if stats == nil {
		return nil
	}

	containerStats := make([]opsports.DashboardContainerStat, 0, len(stats.ContainerStats))
	for _, item := range stats.ContainerStats {
		containerStats = append(containerStats, opsports.DashboardContainerStat{
			ContainerID:   item.ContainerID,
			ContainerName: item.ContainerName,
			CPUPercent:    item.CPUPercent,
			MemoryPercent: item.MemoryPercent,
			MemoryUsage:   item.MemoryUsage,
			MemoryLimit:   item.MemoryLimit,
		})
	}

	alerts := make([]opsports.DashboardResourceAlert, 0, len(stats.Alerts))
	for _, item := range stats.Alerts {
		alerts = append(alerts, opsports.DashboardResourceAlert{
			ContainerID: item.ContainerID,
			Type:        item.Type,
			Value:       item.Value,
			Threshold:   item.Threshold,
			Message:     item.Message,
		})
	}

	return &opsports.DashboardStatsSnapshot{
		OnlineUsers:      stats.OnlineUsers,
		ActiveContainers: stats.ActiveContainers,
		CPUUsage:         stats.CPUUsage,
		MemoryUsage:      stats.MemoryUsage,
		ContainerStats:   containerStats,
		Alerts:           alerts,
	}
}
