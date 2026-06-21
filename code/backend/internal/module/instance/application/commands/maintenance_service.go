package commands

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
	instanceports "ctf-platform/internal/module/instance/ports"
	platformevents "ctf-platform/internal/platform/events"
	"ctf-platform/internal/platform/logctx"
)

type instanceMaintenanceRepository interface {
	UpdateStatusAndReleasePort(ctx context.Context, id int64, status string) error
	FindExpired(ctx context.Context) ([]*instanceentity.Instance, error)
	ListStoppingInstances(ctx context.Context, updatedBefore time.Time, limit int) ([]*instanceentity.Instance, error)
	ListRecoverableActiveInstances(ctx context.Context) ([]*instanceentity.Instance, error)
	FindRunningAWDDefenseWorkspaceByInstanceID(ctx context.Context, instanceID int64) (*instanceports.AWDDefenseWorkspace, error)
	CreateAWDServiceOperation(ctx context.Context, operation *instanceports.AWDServiceOperation) error
	FinishAWDServiceOperation(ctx context.Context, operationID int64, status, errorMessage string, finishedAt time.Time) error
	FinalizeStoppedRuntime(ctx context.Context, id int64) error
	RequeueLostRuntime(ctx context.Context, id int64) (bool, error)
	RequeueLostRuntimesByNode(ctx context.Context, nodeID int64) ([]*instanceentity.Instance, error)
	ListActiveContainerIDs(ctx context.Context) ([]string, error)
}

type instanceMaintenanceEngine interface {
	ListManagedContainers(ctx context.Context) ([]instanceports.ManagedContainer, error)
	InspectManagedContainer(ctx context.Context, containerID string) (*instanceports.ManagedContainerState, error)
	StartContainer(ctx context.Context, containerID string) error
}

type instanceMaintenanceCleaner interface {
	instanceports.RuntimeCleaner
	RemoveContainer(ctx context.Context, containerID string) error
}

type instanceMaintenanceLockStore interface {
	WithStoppingCleanupLock(ctx context.Context, fn func(context.Context)) (bool, error)
}

// InstanceMaintenanceService 收口实例 owner 视角的后台维护能力。
type InstanceMaintenanceService struct {
	repo      instanceMaintenanceRepository
	engine    instanceMaintenanceEngine
	cleaner   instanceMaintenanceCleaner
	lockStore instanceMaintenanceLockStore
	wakeup    chan struct{}
	config    *config.ContainerConfig
	logger    *zap.Logger
}

const (
	defaultStoppingCleanupPollInterval  = time.Second
	defaultStoppingCleanupMaxConcurrent = 8
)

func NewInstanceMaintenanceService(repo instanceMaintenanceRepository, engine instanceMaintenanceEngine, cleaner instanceMaintenanceCleaner, cfg *config.ContainerConfig, logger *zap.Logger, lockStores ...instanceMaintenanceLockStore) *InstanceMaintenanceService {
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
	var lockStore instanceMaintenanceLockStore
	if len(lockStores) > 0 && !isNilCommandDependency(lockStores[0]) {
		lockStore = lockStores[0]
	}
	return &InstanceMaintenanceService{
		repo:      repo,
		engine:    engine,
		cleaner:   cleaner,
		lockStore: lockStore,
		wakeup:    make(chan struct{}, 1),
		config:    cfg,
		logger:    logger,
	}
}

func (s *InstanceMaintenanceService) RegisterStoppingCleanupWakeup(bus platformevents.Bus) {
	if s == nil || bus == nil {
		return
	}
	bus.Subscribe(instancecontracts.EventInstanceStoppingCleanupWakeup, func(context.Context, platformevents.Event) error {
		s.signalStoppingCleanupWakeup()
		return nil
	})
}

func (s *InstanceMaintenanceService) signalStoppingCleanupWakeup() {
	if s == nil || s.wakeup == nil {
		return
	}
	select {
	case s.wakeup <- struct{}{}:
	default:
	}
}

func (s *InstanceMaintenanceService) CleanExpiredInstances(ctx context.Context) error {
	ctx = normalizeContext(ctx)
	instances, err := s.repo.FindExpired(ctx)
	if err != nil {
		return err
	}

	for _, instance := range instances {
		logctx.Info(ctx, s.logger, "清理过期实例", zap.Int64("instance_id", instance.ID))

		if s.cleaner != nil {
			if err := s.cleaner.CleanupRuntime(normalizeContext(ctx), instance); err != nil {
				logctx.Warn(ctx, s.logger, "清理过期实例运行时失败", zap.Int64("instance_id", instance.ID), zap.Error(err))
				continue
			}
		}
		if err := s.repo.UpdateStatusAndReleasePort(ctx, instance.ID, instancecontracts.InstanceStatusExpired); err != nil {
			logctx.Warn(ctx, s.logger, "更新过期实例状态并释放端口失败", zap.Int64("instance_id", instance.ID), zap.Int("host_port", instance.HostPort), zap.Error(err))
		}
	}

	return nil
}

func (s *InstanceMaintenanceService) ReconcileLostActiveRuntimes(ctx context.Context) error {
	ctx = normalizeContext(ctx)
	if s.engine == nil {
		logctx.Debug(ctx, s.logger, "跳过运行时丢失恢复，Docker 引擎未启用")
		return nil
	}

	instances, err := s.repo.ListRecoverableActiveInstances(ctx)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, instance := range instances {
		if instance == nil {
			continue
		}
		lost, reason, stoppedContainerIDs, err := s.isInstanceRuntimeLost(ctx, instance, now)
		if err != nil {
			logctx.Warn(ctx, s.logger, "检查实例运行时状态失败，跳过本实例",
				zap.Int64("instance_id", instance.ID),
				zap.String("status", instance.Status),
				zap.String("container_id", instance.ContainerID),
				zap.Error(err))
			continue
		}
		if !lost {
			continue
		}
		if reason == "container_not_running" && len(stoppedContainerIDs) > 0 {
			if err := s.restartStoppedContainers(ctx, instance, stoppedContainerIDs); err == nil {
				continue
			}
		}

		requeued, err := s.repo.RequeueLostRuntime(ctx, instance.ID)
		if err != nil {
			return err
		}
		if requeued {
			s.recordSystemAWDOperation(ctx, instance, instanceports.AWDServiceOperationTypeRecreate, instanceports.AWDServiceOperationStatusProvisioning, reason, "")
			logctx.Warn(ctx, s.logger, "实例运行时丢失，已重新入队",
				zap.Int64("instance_id", instance.ID),
				zap.String("status", instance.Status),
				zap.String("reason", reason),
				zap.String("container_id", instance.ContainerID))
		}
	}
	return nil
}

func (s *InstanceMaintenanceService) HandleRuntimeNodeOffline(ctx context.Context, nodeID int64) error {
	ctx = normalizeContext(ctx)
	if s == nil || s.repo == nil || nodeID <= 0 {
		return nil
	}
	instances, err := s.repo.RequeueLostRuntimesByNode(ctx, nodeID)
	if err != nil {
		return err
	}
	for _, instance := range instances {
		if instance == nil {
			continue
		}
		s.recordSystemAWDOperation(ctx, instance, instanceports.AWDServiceOperationTypeRecreate, instanceports.AWDServiceOperationStatusProvisioning, "runtime_node_offline", "")
		logctx.Warn(ctx, s.logger, "runtime node offline; instance requeued",
			zap.Int64("runtime_node_id", nodeID),
			zap.Int64("instance_id", instance.ID),
			zap.String("status", instance.Status))
	}
	return nil
}

func (s *InstanceMaintenanceService) RunStoppingCleanupLoop(ctx context.Context) {
	if s == nil || s.repo == nil || s.cleaner == nil {
		return
	}
	if ctx == nil {
		logctx.Warn(ctx, s.logger, "stopping 实例清理循环缺少上下文")
		return
	}

	ticker := time.NewTicker(s.stoppingCleanupPollInterval())
	defer ticker.Stop()

	var (
		mu       sync.Mutex
		inFlight = make(map[int64]struct{})
		wg       sync.WaitGroup
	)

	for {
		s.dispatchStoppingCleanup(ctx, &wg, &mu, inFlight)

		select {
		case <-ctx.Done():
			wg.Wait()
			return
		case <-s.wakeup:
		case <-ticker.C:
		}
	}
}

func (s *InstanceMaintenanceService) dispatchStoppingCleanup(ctx context.Context, wg *sync.WaitGroup, mu *sync.Mutex, inFlight map[int64]struct{}) {
	if s == nil || s.repo == nil || s.cleaner == nil {
		return
	}
	if s.lockStore == nil {
		s.dispatchStoppingCleanupLocked(ctx, wg, mu, inFlight)
		return
	}
	acquired, err := s.lockStore.WithStoppingCleanupLock(ctx, func(lockCtx context.Context) {
		s.dispatchStoppingCleanupLocked(lockCtx, wg, mu, inFlight)
		wg.Wait()
	})
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			logctx.Warn(ctx, s.logger, "获取 stopping 实例清理锁失败", zap.Error(err))
		}
		return
	}
	if !acquired {
		logctx.Debug(ctx, s.logger, "stopping 实例清理已由其他节点执行")
	}
}

func (s *InstanceMaintenanceService) dispatchStoppingCleanupLocked(ctx context.Context, wg *sync.WaitGroup, mu *sync.Mutex, inFlight map[int64]struct{}) {
	if s == nil || s.repo == nil || s.cleaner == nil {
		return
	}

	instances, err := s.repo.ListStoppingInstances(ctx, time.Time{}, s.stoppingCleanupMaxConcurrent())
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			logctx.Warn(ctx, s.logger, "查询 stopping 实例失败", zap.Error(err))
		}
		return
	}

	for _, instance := range instances {
		if instance == nil || instance.ID <= 0 {
			continue
		}
		if !s.tryClaimStoppingInstance(mu, inFlight, instance.ID) {
			continue
		}
		current := *instance
		wg.Add(1)
		go func(item *instanceentity.Instance) {
			defer wg.Done()
			defer s.releaseStoppingInstance(mu, inFlight, item.ID)

			if err := s.cleanupStoppingInstance(ctx, item); err != nil && !errors.Is(err, context.Canceled) {
				logctx.Warn(ctx, s.logger, "清理 stopping 实例运行时失败",
					zap.Int64("instance_id", item.ID),
					zap.Error(err))
			}
		}(&current)
	}
}

func (s *InstanceMaintenanceService) CleanupOrphans(ctx context.Context) error {
	ctx = normalizeContext(ctx)
	if s.engine == nil {
		logctx.Debug(ctx, s.logger, "跳过孤儿容器清理，Docker 引擎未启用")
		return nil
	}
	if s.cleaner == nil {
		logctx.Debug(ctx, s.logger, "跳过孤儿容器清理，运行时清理服务未启用")
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
			logctx.Warn(ctx, s.logger, "删除孤儿容器失败",
				zap.String("container_id", orphan.ID),
				zap.String("container_name", orphan.Name),
				zap.Error(err))
			continue
		}
		logctx.Warn(ctx, s.logger, "已清理孤儿容器",
			zap.String("container_id", orphan.ID),
			zap.String("container_name", orphan.Name),
			zap.Time("created_at", orphan.CreatedAt))
	}

	return nil
}

func (s *InstanceMaintenanceService) isInstanceRuntimeLost(ctx context.Context, instance *instanceentity.Instance, now time.Time) (bool, string, []string, error) {
	if instance.Status == instancecontracts.InstanceStatusStopping {
		return false, "", nil, nil
	}
	if instance.Status == instancecontracts.InstanceStatusCreating && now.Sub(instance.UpdatedAt) < s.runtimeCreateTimeout() {
		return false, "", nil, nil
	}

	containerIDs, err := s.collectRecoverableContainerIDs(ctx, instance)
	if err != nil {
		return false, "", nil, err
	}
	if len(containerIDs) == 0 {
		return true, "missing_runtime_identity", nil, nil
	}

	stoppedContainerIDs := make([]string, 0, len(containerIDs))
	for _, containerID := range containerIDs {
		state, err := s.engine.InspectManagedContainer(ctx, containerID)
		if err != nil {
			return false, "", nil, err
		}
		if state == nil || !state.Exists {
			return true, "container_missing", nil, nil
		}
		if !state.Running {
			stoppedContainerIDs = append(stoppedContainerIDs, containerID)
		}
	}
	if len(stoppedContainerIDs) > 0 {
		return true, "container_not_running", stoppedContainerIDs, nil
	}
	return false, "", nil, nil
}

func (s *InstanceMaintenanceService) collectRecoverableContainerIDs(ctx context.Context, instance *instanceentity.Instance) ([]string, error) {
	containerIDs := collectInstanceContainerIDs(instance)
	if s == nil || s.repo == nil || instance == nil || instance.ID <= 0 {
		return containerIDs, nil
	}

	workspace, err := s.repo.FindRunningAWDDefenseWorkspaceByInstanceID(ctx, instance.ID)
	if err != nil {
		return nil, err
	}
	if workspace == nil || workspace.ContainerID == "" {
		return containerIDs, nil
	}
	return appendUniqueContainerID(containerIDs, workspace.ContainerID), nil
}

func (s *InstanceMaintenanceService) restartStoppedContainers(ctx context.Context, instance *instanceentity.Instance, containerIDs []string) error {
	operationID := s.recordSystemAWDOperation(ctx, instance, instanceports.AWDServiceOperationTypeRecover, instanceports.AWDServiceOperationStatusRecovering, "container_not_running", "")
	for _, containerID := range containerIDs {
		if err := s.engine.StartContainer(ctx, containerID); err != nil {
			s.finishAWDOperation(ctx, operationID, instanceports.AWDServiceOperationStatusFailed, err.Error())
			logctx.Warn(ctx, s.logger, "恢复停止的实例容器失败，准备重新入队",
				zap.Int64("instance_id", instance.ID),
				zap.String("container_id", containerID),
				zap.Error(err))
			return err
		}
		logctx.Warn(ctx, s.logger, "实例容器已自动恢复运行",
			zap.Int64("instance_id", instance.ID),
			zap.String("container_id", containerID))
	}
	s.finishAWDOperation(ctx, operationID, instanceports.AWDServiceOperationStatusRecovered, "")
	return nil
}

func (s *InstanceMaintenanceService) recordSystemAWDOperation(ctx context.Context, instance *instanceentity.Instance, operationType, status, reason, errorMessage string) int64 {
	if s == nil || s.repo == nil || instance == nil || instance.ContestID == nil || instance.TeamID == nil || instance.ServiceID == nil {
		return 0
	}
	now := time.Now().UTC()
	operation := &instanceports.AWDServiceOperation{
		ContestID:     *instance.ContestID,
		TeamID:        *instance.TeamID,
		ServiceID:     *instance.ServiceID,
		InstanceID:    instance.ID,
		OperationType: operationType,
		RequestedBy:   instanceports.AWDServiceOperationRequestedBySystem,
		Reason:        reason,
		SLABillable:   false,
		Status:        status,
		ErrorMessage:  errorMessage,
		StartedAt:     now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.CreateAWDServiceOperation(ctx, operation); err != nil {
		logctx.Warn(ctx, s.logger, "记录 AWD 系统服务操作失败",
			zap.Int64("instance_id", instance.ID),
			zap.String("operation_type", operationType),
			zap.Error(err))
		return 0
	}
	return operation.ID
}

func (s *InstanceMaintenanceService) finishAWDOperation(ctx context.Context, operationID int64, status, errorMessage string) {
	if operationID <= 0 || s == nil || s.repo == nil {
		return
	}
	if err := s.repo.FinishAWDServiceOperation(ctx, operationID, status, errorMessage, time.Now().UTC()); err != nil {
		logctx.Warn(ctx, s.logger, "更新 AWD 系统服务操作失败",
			zap.Int64("operation_id", operationID),
			zap.String("status", status),
			zap.Error(err))
	}
}

func (s *InstanceMaintenanceService) runtimeCreateTimeout() time.Duration {
	if s == nil || s.config == nil || s.config.CreateTimeout <= 0 {
		return 30 * time.Second
	}
	return s.config.CreateTimeout
}

func (s *InstanceMaintenanceService) stoppingCleanupPollInterval() time.Duration {
	if s == nil || s.config == nil || s.config.DeletePollInterval <= 0 {
		return defaultStoppingCleanupPollInterval
	}
	return s.config.DeletePollInterval
}

func (s *InstanceMaintenanceService) stoppingCleanupMaxConcurrent() int {
	if s == nil || s.config == nil || s.config.DeleteMaxConcurrent <= 0 {
		return defaultStoppingCleanupMaxConcurrent
	}
	return s.config.DeleteMaxConcurrent
}

func (s *InstanceMaintenanceService) tryClaimStoppingInstance(mu *sync.Mutex, inFlight map[int64]struct{}, instanceID int64) bool {
	if mu == nil || inFlight == nil || instanceID <= 0 {
		return false
	}

	mu.Lock()
	defer mu.Unlock()

	if len(inFlight) >= s.stoppingCleanupMaxConcurrent() {
		return false
	}
	if _, exists := inFlight[instanceID]; exists {
		return false
	}
	inFlight[instanceID] = struct{}{}
	return true
}

func (s *InstanceMaintenanceService) releaseStoppingInstance(mu *sync.Mutex, inFlight map[int64]struct{}, instanceID int64) {
	if mu == nil || inFlight == nil || instanceID <= 0 {
		return
	}

	mu.Lock()
	delete(inFlight, instanceID)
	mu.Unlock()
}

func (s *InstanceMaintenanceService) cleanupStoppingInstance(ctx context.Context, instance *instanceentity.Instance) error {
	if instance == nil {
		return nil
	}
	if err := s.cleaner.CleanupRuntime(ctx, instance); err != nil {
		return err
	}
	if err := s.repo.FinalizeStoppedRuntime(ctx, instance.ID); err != nil {
		return err
	}
	logctx.Info(ctx, s.logger, "已收尾 stopping 实例",
		zap.Int64("instance_id", instance.ID))
	return nil
}

func selectOrphanContainers(managedContainers []instanceports.ManagedContainer, activeContainerIDs map[string]struct{}, gracePeriod time.Duration) []instanceports.ManagedContainer {
	now := time.Now()
	orphanContainers := make([]instanceports.ManagedContainer, 0, len(managedContainers))
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

func collectInstanceContainerIDs(instance *instanceentity.Instance) []string {
	if instance == nil {
		return nil
	}
	ids := make([]string, 0, 1)
	ids = appendUniqueContainerID(ids, instance.ContainerID)
	containerIDs, err := instancecontracts.ExtractInstanceRuntimeContainerIDs(instance.RuntimeDetails)
	if err != nil {
		return ids
	}
	for _, containerID := range containerIDs {
		ids = appendUniqueContainerID(ids, containerID)
	}
	return ids
}

func appendUniqueContainerID(ids []string, containerID string) []string {
	if containerID == "" {
		return ids
	}
	for _, existing := range ids {
		if existing == containerID {
			return ids
		}
	}
	return append(ids, containerID)
}
