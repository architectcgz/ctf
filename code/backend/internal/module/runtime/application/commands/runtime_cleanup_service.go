package commands

import (
	"context"
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

// RuntimeCleanupService 收口实例运行时资源清理能力。
type RuntimeCleanupService struct {
	engine runtimeCleanupEngine
	logger *zap.Logger
}

// NewRuntimeCleanupService 创建运行时资源清理服务。
func NewRuntimeCleanupService(engine runtimeCleanupEngine, logger *zap.Logger) *RuntimeCleanupService {
	if logger == nil {
		logger = zap.NewNop()
	}
	if isNilCommandDependency(engine) {
		engine = nil
	}
	return &RuntimeCleanupService{
		engine: engine,
		logger: logger,
	}
}

// CleanupRuntime 在后台上下文中清理实例对应的容器、网络和 ACL 规则。
func (s *RuntimeCleanupService) CleanupRuntime(instance *model.Instance) error {
	return s.CleanupRuntimeWithContext(context.Background(), instance)
}

// RemoveContainer 在后台上下文中删除单个容器。
func (s *RuntimeCleanupService) RemoveContainer(containerID string) error {
	return s.RemoveContainerWithContext(context.Background(), containerID)
}

// RemoveContainerWithContext 删除单个容器。
func (s *RuntimeCleanupService) RemoveContainerWithContext(ctx context.Context, containerID string) error {
	return s.removeContainerWithContext(normalizeContext(ctx), containerID)
}

// CleanupRuntimeWithContext 清理实例对应的容器、网络和 ACL 规则。
func (s *RuntimeCleanupService) CleanupRuntimeWithContext(ctx context.Context, instance *model.Instance) error {
	ctx = normalizeContext(ctx)
	if instance == nil {
		return nil
	}

	resources := runtimedomain.ExtractManagedResources(instance)
	if err := s.removeACLRulesWithContext(ctx, resources.ACLRules); err != nil {
		s.logger.Warn("删除实例 ACL 规则失败", zap.Int64("instance_id", instance.ID), zap.Error(err))
	}
	for _, containerID := range resources.ContainerIDs {
		if err := s.RemoveContainerWithContext(ctx, containerID); err != nil {
			return err
		}
	}
	for _, networkID := range resources.NetworkIDs {
		if err := s.removeNetworkWithContext(ctx, networkID); err != nil {
			return err
		}
	}
	return nil
}

func (s *RuntimeCleanupService) removeACLRulesWithContext(ctx context.Context, rules []model.InstanceRuntimeACLRule) error {
	if len(rules) == 0 || s == nil || s.engine == nil {
		return nil
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return s.engine.RemoveACLRules(timeoutCtx, rules)
}

func (s *RuntimeCleanupService) removeContainerWithContext(ctx context.Context, containerID string) error {
	if containerID == "" {
		return nil
	}
	if s == nil || s.engine == nil {
		if s != nil && s.logger != nil {
			s.logger.Info("删除容器（降级模拟）", zap.String("container_id", containerID))
		}
		return nil
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_ = s.engine.StopContainer(timeoutCtx, containerID, 5*time.Second)
	if err := s.engine.RemoveContainer(timeoutCtx, containerID, true); err != nil {
		return err
	}
	s.logger.Info("删除容器", zap.String("container_id", containerID))
	return nil
}

func (s *RuntimeCleanupService) removeNetworkWithContext(ctx context.Context, networkID string) error {
	if networkID == "" {
		return nil
	}
	if s == nil || s.engine == nil {
		if s != nil && s.logger != nil {
			s.logger.Info("删除网络（降级跳过）", zap.String("network_id", networkID))
		}
		return nil
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := s.engine.RemoveNetwork(timeoutCtx, networkID); err != nil {
		return err
	}
	s.logger.Info("删除网络", zap.String("network_id", networkID))
	return nil
}
