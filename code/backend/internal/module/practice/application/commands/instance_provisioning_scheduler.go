package commands

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/model"
)

func (s *Service) RunProvisioningLoop(ctx context.Context) {
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
		if nextAttemptAt := time.Now().UTC(); s.shouldRunDesiredAWDReconcile(lastDesiredReconcileAt, nextAttemptAt) {
			lastDesiredReconcileAt = nextAttemptAt
			if err := s.ReconcileDesiredAWDInstances(ctx); err != nil && !errors.Is(err, context.Canceled) {
				s.logger.Warn("对账 AWD 期望运行态失败", zap.Error(err))
			}
		}

		if err := s.dispatchPendingInstances(ctx); err != nil && !errors.Is(err, context.Canceled) {
			s.logger.Warn("调度待启动实例失败", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *Service) dispatchPendingInstances(ctx context.Context) error {
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
		claimed, err := s.instanceRepo.TryTransitionStatus(ctx, instance.ID, model.InstanceStatusPending, model.InstanceStatusCreating)
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

func (s *Service) availableProvisioningSlots(ctx context.Context) (int, error) {
	slots := s.schedulerMaxConcurrentStarts()
	if slots <= 0 {
		return 0, nil
	}

	creatingCount, err := s.instanceRepo.CountInstancesByStatus(ctx, []string{model.InstanceStatusCreating})
	if err != nil {
		return 0, err
	}
	slots -= int(creatingCount)
	if slots <= 0 {
		return 0, nil
	}

	maxActive := s.schedulerMaxActiveInstances()
	if maxActive > 0 {
		activeCount, err := s.instanceRepo.CountInstancesByStatus(ctx, []string{model.InstanceStatusCreating, model.InstanceStatusRunning})
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

func (s *Service) processPendingInstance(ctx context.Context, instanceID int64) {
	instance, err := s.instanceRepo.FindByID(ctx, instanceID)
	if err != nil {
		s.logger.Error("读取待启动实例失败", zap.Int64("instance_id", instanceID), zap.Error(err))
		return
	}
	if instance == nil || instance.Status != model.InstanceStatusCreating {
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

	if err := s.provisionInstance(ctx, instance, chal, topology, flag); err != nil {
		s.logger.Warn("实例异步启动失败", zap.Int64("instance_id", instanceID), zap.Error(err), wrappedErrorCauseField(err))
	}
}

func (s *Service) schedulerEnabled() bool {
	return s != nil && s.config != nil && s.config.Container.Scheduler.Enabled
}

func (s *Service) schedulerPollInterval() time.Duration {
	if s == nil || s.config == nil || s.config.Container.Scheduler.PollInterval <= 0 {
		return time.Second
	}
	return s.config.Container.Scheduler.PollInterval
}

func (s *Service) desiredAWDReconcileInterval() time.Duration {
	if s == nil || s.config == nil || s.config.Container.Scheduler.DesiredReconcileInterval <= 0 {
		return 15 * time.Second
	}
	return s.config.Container.Scheduler.DesiredReconcileInterval
}

func (s *Service) shouldRunDesiredAWDReconcile(lastAttemptAt, now time.Time) bool {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if lastAttemptAt.IsZero() {
		return true
	}
	return !now.Before(lastAttemptAt.Add(s.desiredAWDReconcileInterval()))
}

func (s *Service) schedulerBatchSize() int {
	if s == nil || s.config == nil || s.config.Container.Scheduler.BatchSize <= 0 {
		return 1
	}
	return s.config.Container.Scheduler.BatchSize
}

func (s *Service) schedulerMaxConcurrentStarts() int {
	if s == nil || s.config == nil || s.config.Container.Scheduler.MaxConcurrentStarts <= 0 {
		return 1
	}
	return s.config.Container.Scheduler.MaxConcurrentStarts
}

func (s *Service) schedulerMaxActiveInstances() int {
	if s == nil || s.config == nil {
		return 0
	}
	return s.config.Container.Scheduler.MaxActiveInstances
}
