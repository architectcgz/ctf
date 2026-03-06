package system

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/module/container"
	rediskeys "ctf-platform/internal/pkg/redis"
)


type DashboardService struct {
	containerRepo *container.Repository
	dockerClient  *client.Client
	redis         *redislib.Client
	config        *config.Config
	logger        *zap.Logger
}

func NewDashboardService(
	containerRepo *container.Repository,
	dockerClient *client.Client,
	redis *redislib.Client,
	cfg *config.Config,
	logger *zap.Logger,
) *DashboardService {
	return &DashboardService{
		containerRepo: containerRepo,
		dockerClient:  dockerClient,
		redis:         redis,
		config:        cfg,
		logger:        logger,
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

	stats := &dto.DashboardStats{
		ContainerStats: []dto.ContainerStat{},
		Alerts:         []dto.ResourceAlert{},
	}

	// 统计在线用户数
	onlineUsers, err := s.countOnlineUsers(ctx)
	if err != nil {
		s.logger.Error("统计在线用户失败", zap.Error(err))
	}
	stats.OnlineUsers = onlineUsers

	// 统计活跃容器数
	activeContainers, err := s.containerRepo.CountRunning()
	if err != nil {
		s.logger.Error("统计活跃容器失败", zap.Error(err))
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
	containers, err := s.dockerClient.ContainerList(ctx, containertypes.ListOptions{})
	if err != nil {
		return nil, err
	}

	stats := make([]dto.ContainerStat, 0, len(containers))
	for _, c := range containers {
		func() {
			stat, err := s.dockerClient.ContainerStats(ctx, c.ID, false)
			if err != nil {
				s.logger.Warn("获取容器统计失败", zap.String("container_id", c.ID), zap.Error(err))
				return
			}
			defer stat.Body.Close()

			var v types.StatsJSON
			if err := json.NewDecoder(stat.Body).Decode(&v); err != nil {
				s.logger.Warn("解析容器统计失败", zap.String("container_id", c.ID), zap.Error(err))
				return
			}

			cpuPercent := calculateCPUPercent(&v)
			memPercent := calculateMemoryPercent(&v)

			containerName := c.ID[:12]
			if len(c.Names) > 0 {
				containerName = c.Names[0]
			}

			stats = append(stats, dto.ContainerStat{
				ContainerID:   c.ID[:12],
				ContainerName: containerName,
				CPUPercent:    cpuPercent,
				MemoryPercent: memPercent,
				MemoryUsage:   int64(v.MemoryStats.Usage),
				MemoryLimit:   int64(v.MemoryStats.Limit),
			})
		}()
	}

	return stats, nil
}

// calculateCPUPercent 计算 CPU 使用率
func calculateCPUPercent(stats *types.StatsJSON) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
	if systemDelta > 0 && cpuDelta > 0 {
		return (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	return 0.0
}

// calculateMemoryPercent 计算内存使用率
func calculateMemoryPercent(stats *types.StatsJSON) float64 {
	if stats.MemoryStats.Limit > 0 {
		return float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit) * 100.0
	}
	return 0.0
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
	data, err := json.Marshal(stats)
	if err != nil {
		s.logger.Error("序列化统计数据失败", zap.Error(err))
		return
	}
	if err := s.redis.Set(ctx, s.getCacheKey(), data, s.config.Dashboard.CacheTTL).Err(); err != nil {
		s.logger.Error("缓存统计数据失败", zap.Error(err))
	}
}

