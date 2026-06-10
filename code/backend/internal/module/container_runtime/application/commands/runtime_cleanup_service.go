package commands

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.uber.org/zap"

	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	runtimedomain "ctf-platform/internal/module/container_runtime/domain"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
)

type RuntimeCleanupRepository interface {
	ReleaseReservedPort(ctx context.Context, port int) error
	ReleasePortForInstance(ctx context.Context, port int, instanceID int64) error
	ReleaseReservedSubnet(ctx context.Context, subnet string) error
	ReleaseSubnetForInstance(ctx context.Context, subnet string, instanceID int64) error
}

// RuntimeCleanupService 收口实例运行时资源清理能力。
type RuntimeCleanupService struct {
	engine runtimeports.ContainerCleanupRuntime
	repo   RuntimeCleanupRepository
	logger *zap.Logger
}

const runtimeCleanupContainerOpTimeout = 10 * time.Second
const runtimeCleanupContainerRemovalPollInterval = 500 * time.Millisecond
const runtimeCleanupNetworkOpTimeout = 10 * time.Second
const runtimeCleanupNetworkRemovalPollInterval = 500 * time.Millisecond

// NewRuntimeCleanupService 创建运行时资源清理服务。
func NewRuntimeCleanupService(engine runtimeports.ContainerCleanupRuntime, repo RuntimeCleanupRepository, logger *zap.Logger) *RuntimeCleanupService {
	if logger == nil {
		logger = zap.NewNop()
	}
	if isNilCommandDependency(engine) {
		engine = nil
	}
	if isNilCommandDependency(repo) {
		repo = nil
	}
	return &RuntimeCleanupService{
		engine: engine,
		repo:   repo,
		logger: logger,
	}
}

// RemoveContainer 删除单个容器。
func (s *RuntimeCleanupService) RemoveContainer(ctx context.Context, containerID string) error {
	return s.removeContainer(normalizeContext(ctx), containerID)
}

// CleanupRuntime 清理目标对应的容器、网络和 ACL 规则。
func (s *RuntimeCleanupService) CleanupRuntime(ctx context.Context, target runtimecontracts.RuntimeCleanupTarget) error {
	ctx = normalizeContext(ctx)
	if target == (runtimecontracts.RuntimeCleanupTarget{}) {
		return nil
	}

	resources := runtimedomain.ExtractManagedResources(target)
	if err := s.removeACL(ctx, resources); err != nil {
		s.logger.Warn("删除实例 ACL 规则失败", zap.Int64("instance_id", target.InstanceID), zap.Error(err))
	}
	for _, containerID := range resources.ContainerIDs {
		if err := s.RemoveContainer(ctx, containerID); err != nil {
			return err
		}
	}
	for _, networkID := range resources.NetworkIDs {
		if err := s.removeNetwork(ctx, networkID); err != nil {
			return err
		}
	}
	for _, subnet := range resources.Subnets {
		if err := s.releaseSubnet(ctx, target.InstanceID, subnet); err != nil {
			return err
		}
	}
	for _, hostPort := range resources.HostPorts {
		if err := s.releasePort(ctx, target.InstanceID, hostPort); err != nil {
			return err
		}
	}
	return nil
}

func (s *RuntimeCleanupService) removeACL(ctx context.Context, resources runtimedomain.ManagedResources) error {
	if s == nil || s.engine == nil {
		return errRuntimeEngineUnavailable()
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 新实例：优先按 ACL handle 删除。
	if resources.ACLHandle != nil {
		return s.engine.RemoveACL(timeoutCtx, resources.ACLHandle)
	}

	return nil
}

func (s *RuntimeCleanupService) removeContainer(ctx context.Context, containerID string) error {
	if containerID == "" {
		return nil
	}
	if s == nil || s.engine == nil {
		return errRuntimeEngineUnavailable()
	}

	stopCtx, stopCancel := context.WithTimeout(ctx, runtimeCleanupContainerOpTimeout)
	_ = s.engine.StopContainer(stopCtx, containerID, 5*time.Second)
	stopCancel()

	if err := s.removeContainerOnce(ctx, containerID); err != nil {
		if errors.Is(err, context.DeadlineExceeded) && ctx.Err() == nil {
			if retryErr := s.removeContainerOnce(ctx, containerID); retryErr == nil {
				s.logger.Info("删除容器在重试后完成", zap.String("container_id", containerID))
				return nil
			} else if isContainerRemovalInProgressError(retryErr) {
				if waitErr := s.waitForContainerRemoval(ctx, containerID); waitErr == nil {
					s.logger.Info("删除容器在后台移除后完成", zap.String("container_id", containerID))
					return nil
				} else {
					return waitErr
				}
			} else {
				return retryErr
			}
		}
		if isContainerRemovalInProgressError(err) && ctx.Err() == nil {
			if waitErr := s.waitForContainerRemoval(ctx, containerID); waitErr == nil {
				s.logger.Info("删除容器在后台移除后完成", zap.String("container_id", containerID))
				return nil
			} else {
				return waitErr
			}
		}
		return err
	}
	s.logger.Info("删除容器", zap.String("container_id", containerID))
	return nil
}

func (s *RuntimeCleanupService) removeContainerOnce(ctx context.Context, containerID string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, runtimeCleanupContainerOpTimeout)
	defer cancel()

	if err := s.engine.RemoveContainer(timeoutCtx, containerID, true); err != nil {
		if errors.Is(err, runtimeports.ErrRuntimeContainerNotFound) {
			s.logger.Info("删除容器跳过，容器不存在", zap.String("container_id", containerID))
			return nil
		}
		return err
	}
	return nil
}

func (s *RuntimeCleanupService) waitForContainerRemoval(ctx context.Context, containerID string) error {
	waitCtx, cancel := context.WithTimeout(ctx, runtimeCleanupContainerOpTimeout)
	defer cancel()

	ticker := time.NewTicker(runtimeCleanupContainerRemovalPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-waitCtx.Done():
			return waitCtx.Err()
		default:
		}

		err := s.removeContainerOnce(waitCtx, containerID)
		if err == nil {
			return nil
		}
		if !isContainerRemovalInProgressError(err) {
			return err
		}

		select {
		case <-waitCtx.Done():
			return waitCtx.Err()
		case <-ticker.C:
		}
	}
}

func (s *RuntimeCleanupService) removeNetwork(ctx context.Context, networkID string) error {
	if networkID == "" {
		return nil
	}
	if s == nil || s.engine == nil {
		return errRuntimeEngineUnavailable()
	}

	if err := s.removeNetworkOnce(ctx, networkID); err != nil {
		if isNetworkRemovalRetryableError(err) && ctx.Err() == nil {
			if waitErr := s.waitForNetworkRemoval(ctx, networkID); waitErr == nil {
				s.logger.Info("删除网络在重试后完成", zap.String("network_id", networkID))
				return nil
			} else {
				return waitErr
			}
		}
		return err
	}
	s.logger.Info("删除网络", zap.String("network_id", networkID))
	return nil
}

func (s *RuntimeCleanupService) removeNetworkOnce(ctx context.Context, networkID string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, runtimeCleanupNetworkOpTimeout)
	defer cancel()

	if err := s.engine.RemoveNetwork(timeoutCtx, networkID); err != nil {
		if errors.Is(err, runtimeports.ErrRuntimeNetworkNotFound) {
			s.logger.Info("删除网络跳过，网络不存在", zap.String("network_id", networkID))
			return nil
		}
		return err
	}
	return nil
}

func (s *RuntimeCleanupService) waitForNetworkRemoval(ctx context.Context, networkID string) error {
	waitCtx, cancel := context.WithTimeout(ctx, runtimeCleanupNetworkOpTimeout)
	defer cancel()

	ticker := time.NewTicker(runtimeCleanupNetworkRemovalPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-waitCtx.Done():
			return waitCtx.Err()
		default:
		}

		err := s.removeNetworkOnce(waitCtx, networkID)
		if err == nil {
			return nil
		}
		if !isNetworkRemovalRetryableError(err) {
			return err
		}

		select {
		case <-waitCtx.Done():
			return waitCtx.Err()
		case <-ticker.C:
		}
	}
}

func (s *RuntimeCleanupService) releasePort(ctx context.Context, instanceID int64, port int) error {
	if port <= 0 || s == nil || s.repo == nil {
		return nil
	}
	var err error
	if instanceID > 0 {
		err = s.repo.ReleasePortForInstance(ctx, port, instanceID)
	} else {
		err = s.repo.ReleaseReservedPort(ctx, port)
	}
	if err != nil {
		return err
	}
	s.logger.Info("释放运行时端口占用",
		zap.Int64("instance_id", instanceID),
		zap.Int("host_port", port))
	return nil
}

func (s *RuntimeCleanupService) releaseSubnet(ctx context.Context, instanceID int64, subnet string) error {
	subnet = strings.TrimSpace(subnet)
	if subnet == "" || s == nil || s.repo == nil {
		return nil
	}
	var err error
	if instanceID > 0 {
		err = s.repo.ReleaseSubnetForInstance(ctx, subnet, instanceID)
	} else {
		err = s.repo.ReleaseReservedSubnet(ctx, subnet)
	}
	if err != nil {
		return err
	}
	s.logger.Info("释放运行时子网占用",
		zap.Int64("instance_id", instanceID),
		zap.String("subnet", subnet))
	return nil
}

func errRuntimeEngineUnavailable() error {
	return runtimeports.ErrRuntimeEngineUnavailable
}

func isContainerRemovalInProgressError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(message, "removal of container") && strings.Contains(message, "already in progress")
}

func isNetworkRemovalRetryableError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(message, "has active endpoints")
}
