package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/model"
	"ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
)

func bestEffortFailureContext(ctx context.Context) context.Context {
	if ctx == nil || ctx.Err() == nil {
		return ctx
	}
	return context.WithoutCancel(ctx)
}

func (s *Service) markInstanceFailed(ctx context.Context, instance *model.Instance) {
	if instance == nil {
		return
	}
	ctx = bestEffortFailureContext(ctx)
	failedAt := time.Now().UTC()
	if err := s.runtimeService.CleanupRuntime(ctx, instance); err != nil {
		s.logger.Warn("清理失败实例运行时资源失败", zap.Int64("instance_id", instance.ID), zap.Error(err))
	}
	if err := s.instanceRepo.UpdateStatusAndReleasePort(ctx, instance.ID, model.InstanceStatusFailed); err != nil {
		s.logger.Warn("更新失败实例状态并释放端口失败", zap.Int64("instance_id", instance.ID), zap.Int("host_port", instance.HostPort), zap.Error(err))
	}
	if err := s.instanceRepo.FinishActiveAWDServiceOperationForInstance(ctx, instance.ID, model.AWDServiceOperationStatusFailed, "provision_failed", failedAt); err != nil {
		s.logger.Warn("更新失败实例 AWD 操作状态失败", zap.Int64("instance_id", instance.ID), zap.Error(err))
	}
	if instance.ContestID != nil && instance.TeamID != nil && instance.ServiceID != nil {
		s.recordDesiredAWDReconcileFailure(ctx, *instance.ContestID, *instance.TeamID, *instance.ServiceID, fmt.Errorf("provision_failed"), failedAt)
	}
}

func (s *Service) provisionInstance(ctx context.Context, instance *model.Instance, chal *model.Challenge, topology *model.ChallengeTopology, flag string) error {
	createCtx, cancel := context.WithTimeout(ctx, s.config.Container.CreateTimeout)
	defer cancel()

	if err := s.createContainer(createCtx, instance, chal, topology, flag); err != nil {
		s.logger.Error("容器创建失败", zap.Error(err), wrappedErrorCauseField(err), zap.Int64("instance_id", instance.ID))
		s.markInstanceFailed(ctx, instance)
		return err
	}
	if !usesAWDStableNetworkAlias(instance) {
		if err := s.waitForInstanceReadiness(createCtx, instance.AccessURL); err != nil {
			s.logger.Error("实例访问地址未就绪", zap.Error(err), zap.Int64("instance_id", instance.ID), zap.String("access_url", instance.AccessURL))
			s.markInstanceFailed(ctx, instance)
			return errcode.ErrContainerStartFailed.WithCause(err)
		}
	} else {
		s.logger.Info("跳过宿主机探活，AWD 实例使用赛内稳定网络访问",
			zap.Int64("instance_id", instance.ID),
			zap.String("access_url", instance.AccessURL))
	}

	instance.Status = model.InstanceStatusRunning
	if err := s.instanceRepo.UpdateRuntime(ctx, instance); err != nil {
		s.logger.Error("更新实例状态失败", zap.Error(err), zap.Int64("instance_id", instance.ID))
		s.markInstanceFailed(ctx, instance)
		return errcode.ErrInternal.WithCause(err)
	}
	if instance.ContestID != nil && instance.TeamID != nil && instance.ServiceID != nil {
		s.clearDesiredAWDReconcileFailure(ctx, *instance.ContestID, *instance.TeamID, *instance.ServiceID)
	}
	if err := s.instanceRepo.FinishActiveAWDServiceOperationForInstance(ctx, instance.ID, model.AWDServiceOperationStatusSucceeded, "", time.Now().UTC()); err != nil {
		s.logger.Warn("更新实例 AWD 操作完成状态失败", zap.Int64("instance_id", instance.ID), zap.Error(err))
	}

	s.logger.Info("实例启动成功",
		zap.Int64("user_id", instance.UserID),
		zap.Int64("challenge_id", instance.ChallengeID),
		zap.Int64("instance_id", instance.ID))
	return nil
}

func (s *Service) waitForInstanceReadiness(ctx context.Context, accessURL string) error {
	if strings.TrimSpace(accessURL) == "" {
		return fmt.Errorf("instance access url is empty")
	}
	if s.readinessProbe == nil {
		return fmt.Errorf("instance readiness probe is not configured")
	}

	attempts := s.startProbeAttempts()
	timeout := s.startProbeTimeout()
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

func (s *Service) buildProvisioningFlag(instance *model.Instance, chal *model.Challenge) (string, error) {
	if instance == nil || chal == nil {
		return "", errcode.ErrInternal.WithCause(fmt.Errorf("instance or challenge is nil"))
	}

	switch chal.FlagType {
	case model.FlagTypeDynamic:
		if strings.TrimSpace(instance.Nonce) == "" {
			return "", errcode.ErrInternal.WithCause(fmt.Errorf("instance nonce is empty"))
		}
		if strings.TrimSpace(s.config.Container.FlagGlobalSecret) == "" {
			return "", errcode.ErrInternal.WithCause(fmt.Errorf("flag global secret is empty"))
		}
		subjectID := instance.UserID
		if instance.TeamID != nil && *instance.TeamID > 0 {
			subjectID = *instance.TeamID
		}
		return crypto.GenerateDynamicFlag(subjectID, chal.ID, s.config.Container.FlagGlobalSecret, instance.Nonce, chal.FlagPrefix), nil
	case model.FlagTypeStatic:
		return chal.FlagHash, nil
	default:
		return "", nil
	}
}

func (s *Service) startProbeTimeout() time.Duration {
	if s == nil || s.config == nil || s.config.Container.StartProbeTimeout <= 0 {
		return 800 * time.Millisecond
	}
	return s.config.Container.StartProbeTimeout
}

func (s *Service) startProbeInterval() time.Duration {
	if s == nil || s.config == nil || s.config.Container.StartProbeInterval <= 0 {
		return 300 * time.Millisecond
	}
	return s.config.Container.StartProbeInterval
}

func (s *Service) startProbeAttempts() int {
	if s == nil || s.config == nil || s.config.Container.StartProbeAttempts <= 0 {
		return 5
	}
	return s.config.Container.StartProbeAttempts
}
