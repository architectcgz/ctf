package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/model"
	runtimedomain "ctf-platform/internal/module/runtime/domain"
)

type runtimeCleanupEngine interface {
	StopContainer(ctx context.Context, containerID string, timeout time.Duration) error
	RemoveContainer(ctx context.Context, containerID string, force bool) error
	RemoveNetwork(ctx context.Context, networkID string) error
	RemoveACLRules(ctx context.Context, rules []model.InstanceRuntimeACLRule) error
}

type runtimeCleanupRepository interface {
	ReleasePort(ctx context.Context, port int) error
}

// RuntimeCleanupService 收口实例运行时资源清理能力。
type RuntimeCleanupService struct {
	engine runtimeCleanupEngine
	repo   runtimeCleanupRepository
	logger *zap.Logger
}

// NewRuntimeCleanupService 创建运行时资源清理服务。
func NewRuntimeCleanupService(engine runtimeCleanupEngine, repo runtimeCleanupRepository, logger *zap.Logger) *RuntimeCleanupService {
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

// CleanupRuntime 清理实例对应的容器、网络和 ACL 规则。
func (s *RuntimeCleanupService) CleanupRuntime(ctx context.Context, instance *model.Instance) error {
	ctx = normalizeContext(ctx)
	if instance == nil {
		return nil
	}

	resources := runtimedomain.ExtractManagedResources(instance)
	if err := s.removeACLRules(ctx, resources.ACLRules); err != nil {
		s.logger.Warn("删除实例 ACL 规则失败", zap.Int64("instance_id", instance.ID), zap.Error(err))
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
	for _, hostPort := range resources.HostPorts {
		if err := s.releasePort(ctx, hostPort); err != nil {
			return err
		}
	}
	return nil
}

func (s *RuntimeCleanupService) removeACLRules(ctx context.Context, rules []model.InstanceRuntimeACLRule) error {
	if len(rules) == 0 {
		return nil
	}
	if s == nil || s.engine == nil {
		return errRuntimeEngineUnavailable()
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return s.engine.RemoveACLRules(timeoutCtx, rules)
}

func (s *RuntimeCleanupService) removeContainer(ctx context.Context, containerID string) error {
	if containerID == "" {
		return nil
	}
	if s == nil || s.engine == nil {
		return errRuntimeEngineUnavailable()
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_ = s.engine.StopContainer(timeoutCtx, containerID, 5*time.Second)
	if err := s.engine.RemoveContainer(timeoutCtx, containerID, true); err != nil {
		if isMissingContainerError(err) {
			s.logger.Info("删除容器跳过，容器不存在", zap.String("container_id", containerID))
			return nil
		}
		return err
	}
	s.logger.Info("删除容器", zap.String("container_id", containerID))
	return nil
}

func (s *RuntimeCleanupService) removeNetwork(ctx context.Context, networkID string) error {
	if networkID == "" {
		return nil
	}
	if s == nil || s.engine == nil {
		return errRuntimeEngineUnavailable()
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := s.engine.RemoveNetwork(timeoutCtx, networkID); err != nil {
		if isMissingNetworkError(err) {
			s.logger.Info("删除网络跳过，网络不存在", zap.String("network_id", networkID))
			return nil
		}
		return err
	}
	s.logger.Info("删除网络", zap.String("network_id", networkID))
	return nil
}

func (s *RuntimeCleanupService) releasePort(ctx context.Context, port int) error {
	if port <= 0 || s == nil || s.repo == nil {
		return nil
	}
	if err := s.repo.ReleasePort(ctx, port); err != nil {
		return err
	}
	s.logger.Info("释放实例端口占用", zap.Int("host_port", port))
	return nil
}

func isMissingContainerError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "no such container")
}

func isMissingNetworkError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "network") && strings.Contains(strings.ToLower(err.Error()), "not found")
}

func errRuntimeEngineUnavailable() error {
	return fmt.Errorf("runtime engine is not configured")
}
