package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/apperror"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	practiceentity "ctf-platform/internal/module/practice/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
	crypto "ctf-platform/internal/shared/flagcrypto"
)

func bestEffortFailureContext(ctx context.Context) context.Context {
	if ctx == nil || ctx.Err() == nil {
		return ctx
	}
	return context.WithoutCancel(ctx)
}

func (s *serviceCore) markInstanceFailed(ctx context.Context, instance *instancecontracts.Instance) {
	if instance == nil {
		return
	}
	ctx = bestEffortFailureContext(ctx)
	failedAt := time.Now().UTC()
	if err := s.runtimeService.CleanupRuntime(ctx, instance); err != nil {
		s.logger.Warn("清理失败实例运行时资源失败", zap.Int64("instance_id", instance.ID), zap.Error(err))
	}
	failed, err := s.instanceRepo.FailProvisioning(ctx, instance.ID)
	if err != nil {
		s.logger.Warn("更新失败实例状态并释放端口失败", zap.Int64("instance_id", instance.ID), zap.Int("host_port", instance.HostPort), zap.Error(err))
	}
	if !failed {
		s.logger.Info("实例已不再处于 creating，跳过 failed 状态写回",
			zap.Int64("instance_id", instance.ID))
		return
	}
	if err := s.instanceRepo.FinishActiveAWDServiceOperationForInstance(ctx, instance.ID, contestcontracts.AWDServiceOperationStatusFailed, "provision_failed", failedAt); err != nil {
		s.logger.Warn("更新失败实例 AWD 操作状态失败", zap.Int64("instance_id", instance.ID), zap.Error(err))
	}
	if instance.ContestID != nil && instance.TeamID != nil && instance.ServiceID != nil {
		s.recordDesiredAWDReconcileFailure(ctx, *instance.ContestID, *instance.TeamID, *instance.ServiceID, fmt.Errorf("provision_failed"), failedAt)
	}
}

func (s *serviceCore) markInstanceRescheduling(ctx context.Context, instance *instancecontracts.Instance, cause error) {
	if instance == nil {
		return
	}
	ctx = bestEffortFailureContext(ctx)
	if err := s.runtimeService.CleanupRuntime(ctx, instance); err != nil {
		s.logger.Warn("清理待重新调度实例运行时资源失败", zap.Int64("instance_id", instance.ID), zap.Error(err))
	}
	attempt := instance.ProvisioningAttempt + 1
	if attempt <= 0 {
		attempt = 1
	}
	if err := s.recordProvisioningProgress(ctx, instancecontracts.ProvisioningProgress{
		InstanceID:            instance.ID,
		Attempt:               attempt,
		Stage:                 instancecontracts.ProvisioningStageRescheduling,
		Severity:              instancecontracts.ProvisioningEventSeverityWarning,
		RuntimeNodeID:         instance.RuntimeNodeID,
		LastProvisioningError: safeProvisioningError(cause),
	}); err != nil {
		s.logger.Warn("记录实例重新调度进度失败", zap.Int64("instance_id", instance.ID), zap.Error(err))
	}
}

func (s *serviceCore) provisionInstance(ctx context.Context, instance *instancecontracts.Instance, chal *practiceentity.Challenge, topology *practiceports.RuntimeChallengeTopology, flag string) error {
	createCtx, cancel := context.WithTimeout(ctx, s.config.Container.CreateTimeout)
	defer cancel()

	if err := s.createContainer(createCtx, instance, chal, topology, flag); err != nil {
		s.logger.Error("容器创建失败", zap.Error(err), wrappedErrorCauseField(err), zap.Int64("instance_id", instance.ID))
		if shouldRescheduleProvisioningFailure(instance, err) {
			s.markInstanceRescheduling(ctx, instance, err)
			return err
		}
		s.markInstanceFailed(ctx, instance)
		return err
	}
	if !usesAWDStableNetworkAlias(instance) {
		if err := s.waitForInstanceReadiness(createCtx, instance.AccessURL); err != nil {
			s.logger.Error("实例访问地址未就绪", zap.Error(err), zap.Int64("instance_id", instance.ID), zap.String("access_url", instance.AccessURL))
			s.markInstanceFailed(ctx, instance)
			return instancecontracts.ErrContainerStartFailed.WithCause(err)
		}
	} else {
		s.logger.Info("跳过宿主机探活，AWD 实例使用赛内稳定网络访问",
			zap.Int64("instance_id", instance.ID),
			zap.String("access_url", instance.AccessURL))
	}

	instance.Status = instancecontracts.InstanceStatusRunning
	persisted, err := s.instanceRepo.PersistProvisionedRuntime(ctx, instance)
	if err != nil {
		s.logger.Error("更新实例状态失败", zap.Error(err), zap.Int64("instance_id", instance.ID))
		s.markInstanceFailed(ctx, instance)
		return apperror.ErrInternal.WithCause(err)
	}
	if !persisted {
		cleanupCtx := bestEffortFailureContext(ctx)
		if err := s.runtimeService.CleanupRuntime(cleanupCtx, instance); err != nil {
			s.logger.Warn("实例已退出 creating，清理已创建运行时失败",
				zap.Int64("instance_id", instance.ID),
				zap.Error(err))
		}
		s.logger.Info("实例已不再处于 creating，跳过运行时持久化",
			zap.Int64("instance_id", instance.ID))
		return nil
	}
	if instance.ContestID != nil && instance.TeamID != nil && instance.ServiceID != nil {
		s.clearDesiredAWDReconcileFailure(ctx, *instance.ContestID, *instance.TeamID, *instance.ServiceID)
	}
	if err := s.instanceRepo.FinishActiveAWDServiceOperationForInstance(ctx, instance.ID, contestcontracts.AWDServiceOperationStatusSucceeded, "", time.Now().UTC()); err != nil {
		s.logger.Warn("更新实例 AWD 操作完成状态失败", zap.Int64("instance_id", instance.ID), zap.Error(err))
	}

	s.logger.Info("实例启动成功",
		zap.Int64("user_id", instance.UserID),
		zap.Int64("challenge_id", instance.ChallengeID),
		zap.Int64("instance_id", instance.ID))
	return nil
}

func shouldRescheduleProvisioningFailure(instance *instancecontracts.Instance, err error) bool {
	if instance == nil || isAWDInstance(instance) {
		return false
	}
	return errors.Is(err, runtimeports.ErrRuntimeNodeUnavailable) ||
		errors.Is(err, runtimeports.ErrRuntimeNetworkSubnetConflict) ||
		errors.Is(err, runtimeports.ErrPublishedHostPortConflict)
}

func safeProvisioningError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func (s *serviceCore) waitForInstanceReadiness(ctx context.Context, accessURL string) error {
	if strings.TrimSpace(accessURL) == "" {
		return fmt.Errorf("instance access url is empty")
	}
	if s.readinessProbe == nil {
		return fmt.Errorf("instance readiness probe is not configured")
	}

	timeout := s.startProbeTimeout()
	attempts := s.effectiveStartProbeAttempts(ctx, timeout, s.startProbeInterval())
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		attemptStartedAt := time.Now()
		lastErr = s.readinessProbe.ProbeAccessURL(ctx, accessURL, timeout)
		if lastErr == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if attempt == attempts-1 {
			break
		}

		waitDuration := s.startProbeInterval()
		if remainingProbeBudget := timeout - time.Since(attemptStartedAt); remainingProbeBudget > 0 {
			waitDuration += remainingProbeBudget
		}
		timer := time.NewTimer(waitDuration)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return lastErr
}

func (s *serviceCore) buildProvisioningFlag(instance *instancecontracts.Instance, chal *practiceentity.Challenge) (string, error) {
	if instance == nil || chal == nil {
		return "", apperror.ErrInternal.WithCause(fmt.Errorf("instance or challenge is nil"))
	}

	switch chal.FlagType {
	case practiceentity.FlagTypeDynamic:
		if strings.TrimSpace(instance.Nonce) == "" {
			return "", apperror.ErrInternal.WithCause(fmt.Errorf("instance nonce is empty"))
		}
		secret, ok := s.flagSecretForKeyID(instance.FlagKeyID)
		if !ok {
			return "", apperror.ErrInternal.WithCause(fmt.Errorf("flag global secret is empty"))
		}
		subjectID := instance.UserID
		if instance.TeamID != nil && *instance.TeamID > 0 {
			subjectID = *instance.TeamID
		}
		return crypto.GenerateDynamicFlag(subjectID, chal.ID, secret, instance.Nonce, chal.FlagPrefix), nil
	case practiceentity.FlagTypeStatic:
		return chal.FlagHash, nil
	default:
		return "", nil
	}
}

func (s *serviceCore) startProbeTimeout() time.Duration {
	if s == nil || s.config == nil || s.config.Container.StartProbeTimeout <= 0 {
		return 800 * time.Millisecond
	}
	return s.config.Container.StartProbeTimeout
}

func (s *serviceCore) startProbeInterval() time.Duration {
	if s == nil || s.config == nil || s.config.Container.StartProbeInterval <= 0 {
		return 300 * time.Millisecond
	}
	return s.config.Container.StartProbeInterval
}

func (s *serviceCore) startProbeAttempts() int {
	if s == nil || s.config == nil || s.config.Container.StartProbeAttempts <= 0 {
		return 5
	}
	return s.config.Container.StartProbeAttempts
}

func (s *serviceCore) effectiveStartProbeAttempts(ctx context.Context, timeout, interval time.Duration) int {
	attempts := s.startProbeAttempts()
	if attempts < 1 {
		attempts = 1
	}

	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		return attempts
	}

	remaining := time.Until(deadline)
	if remaining <= 0 {
		return attempts
	}

	cycleDuration := timeout + interval
	if cycleDuration <= 0 {
		cycleDuration = time.Second
	}

	derivedAttempts := int(remaining / cycleDuration)
	if remaining%cycleDuration != 0 {
		derivedAttempts++
	}
	if derivedAttempts < 1 {
		derivedAttempts = 1
	}
	if derivedAttempts > attempts {
		return derivedAttempts
	}
	return attempts
}
