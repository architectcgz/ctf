package commands

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	runtimeports "ctf-platform/internal/module/container_runtime/ports"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	practiceports "ctf-platform/internal/module/practice/ports"
	"ctf-platform/internal/shared/lockkeepalive"
)

func (s *serviceCore) RunProvisioningLoop(ctx context.Context) {
	if !s.schedulerEnabled() {
		return
	}
	if ctx == nil {
		s.logger.Warn("实例启动调度循环缺少上下文")
		return
	}

	ticker := time.NewTicker(s.schedulerPollInterval())
	defer ticker.Stop()
	var lastDesiredReconcileAt time.Time

	for {
		if nextLastAttemptAt, acquired, err := s.runProvisioningCycle(ctx, lastDesiredReconcileAt); err != nil {
			if !errors.Is(err, context.Canceled) {
				s.logger.Warn("调度待启动实例失败", zap.Error(err))
			}
		} else if acquired {
			lastDesiredReconcileAt = nextLastAttemptAt
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *serviceCore) runProvisioningCycle(ctx context.Context, lastDesiredReconcileAt time.Time) (time.Time, bool, error) {
	nextLastDesiredReconcileAt := lastDesiredReconcileAt
	acquired, err := s.withProvisioningSchedulerLock(ctx, func(lockCtx context.Context) error {
		if nextAttemptAt := time.Now().UTC(); s.shouldRunDesiredAWDReconcile(lastDesiredReconcileAt, nextAttemptAt) {
			nextLastDesiredReconcileAt = nextAttemptAt
			if err := s.ReconcileDesiredAWDInstances(lockCtx); err != nil && !errors.Is(err, context.Canceled) {
				s.logger.Warn("对账 AWD 期望运行态失败", zap.Error(err))
			}
		}
		return s.dispatchPendingInstances(lockCtx)
	})
	return nextLastDesiredReconcileAt, acquired, err
}

func (s *serviceCore) dispatchPendingInstances(ctx context.Context) error {
	limit, err := s.availableProvisioningSlots(ctx)
	if err != nil {
		return err
	}
	if limit <= 0 {
		return nil
	}

	instances, err := s.instanceRepo.ListPendingInstances(ctx, limit)
	if err != nil {
		return err
	}
	for _, instance := range instances {
		if instance == nil {
			continue
		}
		claimed, err := s.instanceRepo.TryTransitionStatus(ctx, instance.ID, instancecontracts.InstanceStatusPending, instancecontracts.InstanceStatusCreating)
		if err != nil {
			return err
		}
		if !claimed {
			continue
		}

		instanceID := instance.ID
		s.runAsyncTask(func(taskCtx context.Context) {
			s.processPendingInstance(taskCtx, instanceID)
		})
	}
	return nil
}

func (s *serviceCore) availableProvisioningSlots(ctx context.Context) (int, error) {
	slots := s.schedulerMaxConcurrentStarts()
	if slots <= 0 {
		return 0, nil
	}

	creatingCount, err := s.instanceRepo.CountInstancesByStatus(ctx, []string{instancecontracts.InstanceStatusCreating})
	if err != nil {
		return 0, err
	}
	slots -= int(creatingCount)
	if slots <= 0 {
		return 0, nil
	}

	maxActive := s.schedulerMaxActiveInstances()
	if maxActive > 0 {
		activeCount, err := s.instanceRepo.CountInstancesByStatus(ctx, []string{instancecontracts.InstanceStatusCreating, instancecontracts.InstanceStatusRunning})
		if err != nil {
			return 0, err
		}
		remainingCapacity := maxActive - int(activeCount)
		if remainingCapacity <= 0 {
			return 0, nil
		}
		if remainingCapacity < slots {
			slots = remainingCapacity
		}
	}

	batchSize := s.schedulerBatchSize()
	if batchSize > 0 && batchSize < slots {
		slots = batchSize
	}
	return slots, nil
}

func (s *serviceCore) processPendingInstance(ctx context.Context, instanceID int64) {
	instance, err := s.instanceRepo.FindByID(ctx, instanceID)
	if err != nil {
		s.logger.Error("读取待启动实例失败", zap.Int64("instance_id", instanceID), zap.Error(err))
		return
	}
	if instance == nil || instance.Status != instancecontracts.InstanceStatusCreating {
		return
	}

	chal, topology, err := s.loadRuntimeSubjectForInstance(ctx, instance)
	if err != nil {
		s.logger.Error("读取题目失败", zap.Int64("instance_id", instanceID), zap.Int64("challenge_id", instance.ChallengeID), zap.Error(err))
		s.markInstanceFailed(ctx, instance)
		return
	}

	flag, err := s.buildProvisioningFlag(instance, chal)
	if err != nil {
		s.logger.Error("生成实例 Flag 失败", zap.Int64("instance_id", instanceID), zap.Error(err))
		s.markInstanceFailed(ctx, instance)
		return
	}

	nodeBinding, err := s.selectRuntimeNode(ctx, instanceScopeFromInstance(instance))
	if err != nil {
		if errors.Is(err, runtimeports.ErrRuntimeNodeUnavailable) {
			s.requeuePendingInstance(ctx, instance)
			return
		}
		s.logger.Warn("待启动实例选择运行节点失败", zap.Int64("instance_id", instanceID), zap.Error(err))
		s.markInstanceFailed(ctx, instance)
		return
	}
	instance.RuntimeNodeID = runtimeNodeIDFromBinding(nodeBinding)
	bound, err := s.instanceRepo.BindRuntimeNode(ctx, instance.ID, instance.RuntimeNodeID)
	if err != nil {
		s.logger.Warn("待启动实例绑定运行节点失败", zap.Int64("instance_id", instanceID), zap.Error(err))
		s.markInstanceFailed(ctx, instance)
		return
	}
	if !bound {
		s.logger.Info("实例已不再处于 creating，跳过运行节点绑定后启动",
			zap.Int64("instance_id", instanceID))
		return
	}

	if err := s.provisionInstance(ctx, instance, chal, topology, flag); err != nil {
		s.logger.Warn("实例异步启动失败", zap.Int64("instance_id", instanceID), zap.Error(err), wrappedErrorCauseField(err))
	}
}

func (s *serviceCore) requeuePendingInstance(ctx context.Context, instance *instancecontracts.Instance) {
	if s == nil || s.instanceRepo == nil || instance == nil {
		return
	}
	requeued, err := s.instanceRepo.RequeueLostRuntime(ctx, instance.ID)
	if err != nil {
		s.logger.Warn("待启动实例回到 pending 失败", zap.Int64("instance_id", instance.ID), zap.Error(err))
		return
	}
	if requeued {
		s.logger.Info("待启动实例等待健康运行节点",
			zap.Int64("instance_id", instance.ID))
	}
}

func instanceScopeFromInstance(instance *instancecontracts.Instance) practiceports.InstanceScope {
	if instance == nil {
		return practiceports.InstanceScope{}
	}
	scope := practiceports.InstanceScope{
		ContestID:  instance.ContestID,
		TeamID:     instance.TeamID,
		ServiceID:  instance.ServiceID,
		ShareScope: instance.ShareScope,
	}
	switch {
	case instance.ServiceID != nil && *instance.ServiceID > 0:
		scope.ContestMode = practiceports.ContestModeAWD
	}
	switch instance.ShareScope {
	case instancecontracts.ShareScopePerTeam:
		if instance.TeamID != nil && *instance.TeamID > 0 {
			scope.FlagSubjectID = *instance.TeamID
		}
	default:
		scope.FlagSubjectID = instance.UserID
	}
	return scope
}

func (s *serviceCore) schedulerEnabled() bool {
	return s != nil && s.config != nil && s.config.Container.Scheduler.Enabled
}

func (s *serviceCore) schedulerPollInterval() time.Duration {
	if s == nil || s.config == nil || s.config.Container.Scheduler.PollInterval <= 0 {
		return time.Second
	}
	return s.config.Container.Scheduler.PollInterval
}

func (s *serviceCore) desiredAWDReconcileInterval() time.Duration {
	if s == nil || s.config == nil || s.config.Container.Scheduler.DesiredReconcileInterval <= 0 {
		return 15 * time.Second
	}
	return s.config.Container.Scheduler.DesiredReconcileInterval
}

func (s *serviceCore) desiredAWDReconcileFailureInitialBackoff() time.Duration {
	if s == nil || s.config == nil || s.config.Container.Scheduler.DesiredReconcileFailureInitialBackoff <= 0 {
		return 30 * time.Second
	}
	return s.config.Container.Scheduler.DesiredReconcileFailureInitialBackoff
}

func (s *serviceCore) desiredAWDReconcileFailureMaxBackoff() time.Duration {
	if s == nil || s.config == nil || s.config.Container.Scheduler.DesiredReconcileFailureMaxBackoff <= 0 {
		return 10 * time.Minute
	}
	return s.config.Container.Scheduler.DesiredReconcileFailureMaxBackoff
}

func (s *serviceCore) desiredAWDReconcileSuppressAfterFailures() int {
	if s == nil || s.config == nil || s.config.Container.Scheduler.DesiredReconcileSuppressAfterFailures < 0 {
		return 0
	}
	return s.config.Container.Scheduler.DesiredReconcileSuppressAfterFailures
}

func (s *serviceCore) desiredAWDReconcileSuppressDuration() time.Duration {
	if s == nil || s.config == nil || s.config.Container.Scheduler.DesiredReconcileSuppressDuration <= 0 {
		return 30 * time.Minute
	}
	return s.config.Container.Scheduler.DesiredReconcileSuppressDuration
}

func (s *serviceCore) shouldRunDesiredAWDReconcile(lastAttemptAt, now time.Time) bool {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if lastAttemptAt.IsZero() {
		return true
	}
	return !now.Before(lastAttemptAt.Add(s.desiredAWDReconcileInterval()))
}

func (s *serviceCore) schedulerBatchSize() int {
	if s == nil || s.config == nil || s.config.Container.Scheduler.BatchSize <= 0 {
		return 1
	}
	return s.config.Container.Scheduler.BatchSize
}

func (s *serviceCore) schedulerMaxConcurrentStarts() int {
	if s == nil || s.config == nil || s.config.Container.Scheduler.MaxConcurrentStarts <= 0 {
		return 1
	}
	return s.config.Container.Scheduler.MaxConcurrentStarts
}

func (s *serviceCore) schedulerMaxActiveInstances() int {
	if s == nil || s.config == nil {
		return 0
	}
	return s.config.Container.Scheduler.MaxActiveInstances
}

func (s *serviceCore) schedulerLockTTL() time.Duration {
	if s == nil || s.config == nil || s.config.Container.Scheduler.LockTTL <= 0 {
		return 30 * time.Second
	}
	return s.config.Container.Scheduler.LockTTL
}

func (s *serviceCore) tryAcquireProvisioningSchedulerLock(ctx context.Context) (practiceSchedulerLockLease, bool, error) {
	if s == nil || s.schedulerLockStore == nil {
		return nil, true, nil
	}
	return s.schedulerLockStore.AcquireProvisioningSchedulerLock(ctx, s.schedulerLockTTL())
}

func (s *serviceCore) withProvisioningSchedulerLock(ctx context.Context, fn func(context.Context) error) (bool, error) {
	lock, acquired, err := s.tryAcquireProvisioningSchedulerLock(ctx)
	if err != nil {
		return false, err
	}
	if !acquired {
		return false, nil
	}

	lockCtx, stopKeepalive := lockkeepalive.Start(ctx, s.logger, lock, "practice_instance_scheduler", s.schedulerLockTTL())
	defer stopKeepalive()
	defer s.releaseProvisioningSchedulerLock(ctx, lock)
	return true, fn(lockCtx)
}

func (s *serviceCore) releaseProvisioningSchedulerLock(ctx context.Context, lock practiceSchedulerLockLease) {
	if lock == nil || ctx == nil {
		return
	}

	releaseCtx := context.WithoutCancel(ctx)
	if timeout := s.schedulerLockTTL(); timeout > 0 {
		var cancel context.CancelFunc
		releaseCtx, cancel = context.WithTimeout(releaseCtx, timeout)
		defer cancel()
	}

	released, err := lock.Release(releaseCtx)
	if err != nil {
		s.logger.Error("practice_instance_scheduler_lock_release_failed",
			zap.String("lock_key", lock.Key(releaseCtx)),
			zap.Error(err))
		return
	}
	if !released && ctx.Err() == nil {
		s.logger.Warn("practice_instance_scheduler_lock_already_lost",
			zap.String("lock_key", lock.Key(releaseCtx)))
	}
}
