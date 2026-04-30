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
	UpdateStatusAndReleasePort(ctx context.Context, id int64, status string) error
	FindExpired(ctx context.Context) ([]*model.Instance, error)
	ListRecoverableActiveInstances(ctx context.Context) ([]*model.Instance, error)
	RequeueLostRuntime(ctx context.Context, id int64) (bool, error)
	ListActiveContainerIDs(ctx context.Context) ([]string, error)
}

type runtimeMaintenanceEngine interface {
	ListManagedContainers(ctx context.Context) ([]runtimeports.ManagedContainer, error)
	InspectManagedContainer(ctx context.Context, containerID string) (*runtimeports.ManagedContainerState, error)
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
	ctx = normalizeContext(ctx)
	instances, err := s.repo.FindExpired(ctx)
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
		if err := s.repo.UpdateStatusAndReleasePort(ctx, instance.ID, model.InstanceStatusExpired); err != nil {
			s.logger.Warn("更新过期实例状态并释放端口失败", zap.Int64("instance_id", instance.ID), zap.Int("host_port", instance.HostPort), zap.Error(err))
		}
	}

	return nil
}

// ReconcileLostActiveRuntimes finds active instance records whose Docker runtime disappeared and
// requeues them for the existing provisioning scheduler.
func (s *RuntimeMaintenanceService) ReconcileLostActiveRuntimes(ctx context.Context) error {
	ctx = normalizeContext(ctx)
	if s.engine == nil {
		s.logger.Debug("跳过运行时丢失恢复，Docker 引擎未启用")
		return nil
	}

	instances, err := s.repo.ListRecoverableActiveInstances(ctx)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, instance := range instances {
		if instance == nil {
			continue
		}
		lost, reason, err := s.isInstanceRuntimeLost(ctx, instance, now)
		if err != nil {
			return err
		}
		if !lost {
			continue
		}

		requeued, err := s.repo.RequeueLostRuntime(ctx, instance.ID)
		if err != nil {
			return err
		}
		if requeued {
			s.logger.Warn("实例运行时丢失，已重新入队",
				zap.Int64("instance_id", instance.ID),
				zap.String("status", instance.Status),
				zap.String("reason", reason),
				zap.String("container_id", instance.ContainerID))
		}
	}
	return nil
}

func (s *RuntimeMaintenanceService) isInstanceRuntimeLost(ctx context.Context, instance *model.Instance, now time.Time) (bool, string, error) {
	if instance.Status == model.InstanceStatusCreating && now.Sub(instance.UpdatedAt) < s.runtimeCreateTimeout() {
		return false, "", nil
	}

	containerIDs := collectInstanceContainerIDs(instance)
	if len(containerIDs) == 0 {
		return true, "missing_runtime_identity", nil
	}

	for _, containerID := range containerIDs {
		state, err := s.engine.InspectManagedContainer(ctx, containerID)
		if err != nil {
			return false, "", err
		}
		if state == nil || !state.Exists {
			return true, "container_missing", nil
		}
		if !state.Running {
			return true, "container_not_running", nil
		}
	}
	return false, "", nil
}

func (s *RuntimeMaintenanceService) runtimeCreateTimeout() time.Duration {
	if s == nil || s.config == nil || s.config.CreateTimeout <= 0 {
		return 30 * time.Second
	}
	return s.config.CreateTimeout
}

// CleanupOrphans 清理未被实例记录持有的受管孤儿容器。
func (s *RuntimeMaintenanceService) CleanupOrphans(ctx context.Context) error {
	ctx = normalizeContext(ctx)
	if s.engine == nil {
		s.logger.Debug("跳过孤儿容器清理，Docker 引擎未启用")
		return nil
	}
	if s.cleaner == nil {
		s.logger.Debug("跳过孤儿容器清理，运行时清理服务未启用")
		return nil
	}

	managedContainers, err := s.engine.ListManagedContainers(ctx)
	if err != nil {
		return err
	}
	activeContainerIDs, err := s.repo.ListActiveContainerIDs(ctx)
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

func collectInstanceContainerIDs(instance *model.Instance) []string {
	if instance == nil {
		return nil
	}
	ids := make([]string, 0, 1)
	seen := make(map[string]struct{})
	add := func(containerID string) {
		if containerID == "" {
			return
		}
		if _, exists := seen[containerID]; exists {
			return
		}
		seen[containerID] = struct{}{}
		ids = append(ids, containerID)
	}

	add(instance.ContainerID)
	details, err := model.DecodeInstanceRuntimeDetails(instance.RuntimeDetails)
	if err != nil {
		return ids
	}
	for _, container := range details.Containers {
		add(container.ContainerID)
	}
	return ids
}
