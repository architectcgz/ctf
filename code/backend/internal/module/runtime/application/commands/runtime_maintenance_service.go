package commands

import (
	"context"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type runtimeMaintenanceRepository interface {
	UpdateStatusAndReleasePort(id int64, status string) error
	UpdateStatusAndReleasePortWithContext(ctx context.Context, id int64, status string) error
	FindExpired() ([]*model.Instance, error)
	ListActiveContainerIDs() ([]string, error)
}

type runtimeMaintenanceEngine interface {
	ListManagedContainers(ctx context.Context) ([]runtimeports.ManagedContainer, error)
}

type runtimeMaintenanceCleaner interface {
	runtimeports.RuntimeCleaner
	RemoveContainer(ctx context.Context, containerID string) error
}

// RuntimeMaintenanceService 收口后台定时任务驱动的运行时维护能力。
type RuntimeMaintenanceService struct {
	repo    runtimeMaintenanceRepository
	engine  runtimeMaintenanceEngine
	cleaner runtimeMaintenanceCleaner
	config  *config.ContainerConfig
	logger  *zap.Logger
}

// NewRuntimeMaintenanceService 创建运行时维护服务。
func NewRuntimeMaintenanceService(repo runtimeMaintenanceRepository, engine runtimeMaintenanceEngine, cleaner runtimeMaintenanceCleaner, cfg *config.ContainerConfig, logger *zap.Logger) *RuntimeMaintenanceService {
	if logger == nil {
		logger = zap.NewNop()
	}
	if isNilCommandDependency(repo) {
		repo = nil
	}
	if isNilCommandDependency(engine) {
		engine = nil
	}
	if isNilCommandDependency(cleaner) {
		cleaner = nil
	}
	if cfg == nil {
		cfg = &config.ContainerConfig{}
	}
	return &RuntimeMaintenanceService{
		repo:    repo,
		engine:  engine,
		cleaner: cleaner,
		config:  cfg,
		logger:  logger,
	}
}

// CleanExpiredInstances 清理已过期实例的运行时资源并释放端口占用。
func (s *RuntimeMaintenanceService) CleanExpiredInstances(ctx context.Context) error {
	instances, err := s.repo.FindExpired()
	if err != nil {
		return err
	}

	for _, instance := range instances {
		s.logger.Info("清理过期实例", zap.Int64("instance_id", instance.ID))

		if s.cleaner != nil {
			if err := s.cleaner.CleanupRuntime(normalizeContext(ctx), instance); err != nil {
				s.logger.Warn("清理过期实例运行时失败", zap.Int64("instance_id", instance.ID), zap.Error(err))
				continue
			}
		}
		if err := s.repo.UpdateStatusAndReleasePortWithContext(normalizeContext(ctx), instance.ID, model.InstanceStatusExpired); err != nil {
			s.logger.Warn("更新过期实例状态并释放端口失败", zap.Int64("instance_id", instance.ID), zap.Int("host_port", instance.HostPort), zap.Error(err))
		}
	}

	return nil
}

// CleanupOrphans 清理未被实例记录持有的受管孤儿容器。
func (s *RuntimeMaintenanceService) CleanupOrphans(ctx context.Context) error {
	if s.engine == nil {
		s.logger.Debug("跳过孤儿容器清理，Docker 引擎未启用")
		return nil
	}
	if s.cleaner == nil {
		s.logger.Debug("跳过孤儿容器清理，运行时清理服务未启用")
		return nil
	}

	managedContainers, err := s.engine.ListManagedContainers(normalizeContext(ctx))
	if err != nil {
		return err
	}
	activeContainerIDs, err := s.repo.ListActiveContainerIDs()
	if err != nil {
		return err
	}

	activeSet := make(map[string]struct{}, len(activeContainerIDs))
	for _, containerID := range activeContainerIDs {
		activeSet[containerID] = struct{}{}
	}

	for _, orphan := range selectOrphanContainers(managedContainers, activeSet, s.config.OrphanGracePeriod) {
		if err := s.cleaner.RemoveContainer(ctx, orphan.ID); err != nil {
			s.logger.Warn("删除孤儿容器失败",
				zap.String("container_id", orphan.ID),
				zap.String("container_name", orphan.Name),
				zap.Error(err))
			continue
		}
		s.logger.Warn("已清理孤儿容器",
			zap.String("container_id", orphan.ID),
			zap.String("container_name", orphan.Name),
			zap.Time("created_at", orphan.CreatedAt))
	}

	return nil
}

func selectOrphanContainers(managedContainers []runtimeports.ManagedContainer, activeContainerIDs map[string]struct{}, gracePeriod time.Duration) []runtimeports.ManagedContainer {
	now := time.Now()
	orphanContainers := make([]runtimeports.ManagedContainer, 0, len(managedContainers))
	for _, container := range managedContainers {
		if _, exists := activeContainerIDs[container.ID]; exists {
			continue
		}
		if !container.CreatedAt.IsZero() && now.Sub(container.CreatedAt) < gracePeriod {
			continue
		}
		orphanContainers = append(orphanContainers, container)
	}
	return orphanContainers
}
