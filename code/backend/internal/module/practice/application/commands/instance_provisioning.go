package commands

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/model"
	"ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
)

func (s *Service) markInstanceFailed(ctx context.Context, instance *model.Instance) {
	if instance == nil {
		return
	}
	if err := s.runtimeService.CleanupRuntime(ctx, instance); err != nil {
		s.logger.Warn("清理失败实例运行时资源失败", zap.Int64("instance_id", instance.ID), zap.Error(err))
	}
	if err := s.instanceRepo.UpdateStatusAndReleasePort(ctx, instance.ID, model.InstanceStatusFailed); err != nil {
		s.logger.Warn("更新失败实例状态并释放端口失败", zap.Int64("instance_id", instance.ID), zap.Int("host_port", instance.HostPort), zap.Error(err))
	}
	if err := s.instanceRepo.FinishActiveAWDServiceOperationForInstance(ctx, instance.ID, model.AWDServiceOperationStatusFailed, "provision_failed", time.Now().UTC()); err != nil {
		s.logger.Warn("更新失败实例 AWD 操作状态失败", zap.Int64("instance_id", instance.ID), zap.Error(err))
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

	attempts := s.startProbeAttempts()
	client := &http.Client{Timeout: s.startProbeTimeout()}
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		lastErr = s.probeInstanceAccessURL(ctx, client, accessURL)
		if lastErr == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if attempt == attempts-1 {
			break
		}

		timer := time.NewTimer(s.startProbeInterval())
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return lastErr
}

func (s *Service) probeInstanceAccessURL(ctx context.Context, client *http.Client, accessURL string) error {
	parsed, err := url.Parse(accessURL)
	if err != nil {
		return err
	}
	if strings.EqualFold(parsed.Scheme, model.ChallengeTargetProtocolTCP) {
		return probeTCPAccessURL(ctx, parsed, s.startProbeTimeout())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, accessURL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 512))
	return nil
}

func probeTCPAccessURL(ctx context.Context, parsed *url.URL, timeout time.Duration) error {
	host := parsed.Host
	if strings.TrimSpace(host) == "" {
		return fmt.Errorf("tcp access url missing host")
	}
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", host)
	if err != nil {
		return err
	}
	return conn.Close()
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
