package application

import (
	"context"
	"encoding/json"
	"fmt"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	rediskeys "ctf-platform/internal/pkg/redis"
)

type DashboardService struct {
	runtimeQuery RuntimeQuery
	runtime      RuntimeStatsProvider
	redis        *redislib.Client
	config       *config.Config
	logger       *zap.Logger
}

func NewDashboardService(
	runtimeQuery RuntimeQuery,
	runtimeStats RuntimeStatsProvider,
	redis *redislib.Client,
	cfg *config.Config,
	logger *zap.Logger,
) *DashboardService {
	return &DashboardService{
		runtimeQuery: runtimeQuery,
		runtime:      runtimeStats,
		redis:        redis,
		config:       cfg,
		logger:       logger,
	}
}

// getCacheKey 动态构建缓存 Key
func (s *DashboardService) getCacheKey() string {
	return fmt.Sprintf("%s:stats", s.config.Dashboard.RedisKeyPrefix)
}

// GetDashboardStats 获取仪表盘统计数据
func (s *DashboardService) GetDashboardStats(ctx context.Context) (*dto.DashboardStats, error) {
	// 尝试从缓存获取
	cached, err := s.getFromCache(ctx)
	if err == nil && cached != nil {
		return cached, nil
	}
	if err != nil && err != redislib.Nil {
		s.logger.Error("读取仪表盘缓存失败，降级到实时查询", zap.Error(err))
	}

	stats := &dto.DashboardStats{
		ContainerStats: []dto.ContainerStat{},
		Alerts:         []dto.ResourceAlert{},
	}

	// 统计在线用户数
	onlineUsers, err := s.countOnlineUsers(ctx)
	if err != nil {
		s.logger.Error("统计在线用户失败", zap.Error(err))
		onlineUsers = -1
	}
	stats.OnlineUsers = onlineUsers

	// 统计活跃容器数
	activeContainers := int64(0)
	if s.runtimeQuery != nil {
		activeContainers, err = s.runtimeQuery.CountRunning()
		if err != nil {
			return nil, fmt.Errorf("统计活跃容器失败: %w", err)
		}
	}
	stats.ActiveContainers = activeContainers

	// 获取容器资源使用情况
	containerStats, err := s.getContainerStats(ctx)
	if err != nil {
		s.logger.Error("获取容器统计失败", zap.Error(err))
	} else {
		stats.ContainerStats = containerStats
		stats.CPUUsage, stats.MemoryUsage = s.calculateAverageUsage(containerStats)
		stats.Alerts = s.checkAlerts(containerStats)
	}

	// 缓存结果
	s.saveToCache(ctx, stats)

	return stats, nil
}

// getContainerStats 获取所有容器的资源统计
func (s *DashboardService) getContainerStats(ctx context.Context) ([]dto.ContainerStat, error) {
	if s.runtime == nil {
		return []dto.ContainerStat{}, nil
	}
	stats, err := s.runtime.ListManagedContainerStats(ctx)
	if err != nil {
		if err == redislib.Nil {
			s.logger.Debug("仪表盘缓存未命中")
			return nil, err
		}
		s.logger.Warn("读取仪表盘缓存失败", zap.Error(err))
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

// calculateAverageUsage 计算平均资源使用率
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

// checkAlerts 检查资源告警
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

// countOnlineUsers 统计在线用户数
// 通过统计 Redis 中存在的 Refresh Token 数量来估算在线用户
func (s *DashboardService) countOnlineUsers(ctx context.Context) (int64, error) {
	if s.redis == nil {
		return 0, nil
	}
	pattern := rediskeys.Namespace + ":token:*"
	var cursor uint64
	count := int64(0)
	for {
		keys, nextCursor, err := s.redis.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return 0, err
		}
		count += int64(len(keys))
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return count, nil
}

// getFromCache 从缓存获取统计数据
func (s *DashboardService) getFromCache(ctx context.Context) (*dto.DashboardStats, error) {
	if s.redis == nil {
		return nil, redislib.Nil
	}
	data, err := s.redis.Get(ctx, s.getCacheKey()).Bytes()
	if err != nil {
		return nil, err
	}
	var stats dto.DashboardStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// saveToCache 保存统计数据到缓存
func (s *DashboardService) saveToCache(ctx context.Context, stats *dto.DashboardStats) {
	if s.redis == nil {
		return
	}
	data, err := json.Marshal(stats)
	if err != nil {
		s.logger.Error("序列化统计数据失败", zap.Error(err))
		return
	}
	if err := s.redis.Set(ctx, s.getCacheKey(), data, s.config.Dashboard.CacheTTL).Err(); err != nil {
		s.logger.Error("缓存统计数据失败", zap.Error(err))
		return
	}
	s.logger.Debug("仪表盘缓存已更新")
}
