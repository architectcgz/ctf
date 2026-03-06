package container

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

type Service struct {
	repo   *Repository
	config *config.ContainerConfig
	logger *zap.Logger
}

func NewService(repo *Repository, cfg *config.ContainerConfig, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		config: cfg,
		logger: logger,
	}
}

// CreateContainer 创建容器（B9-B11 实现后替换为真实 Docker 调用）
func (s *Service) CreateContainer(ctx context.Context, imageName string, env map[string]string) (containerID, networkID string, port int, err error) {
	// TODO: 实际的 Docker 容器创建逻辑
	// 当前为模拟实现，等待 B9-B11 完成后集成

	select {
	case <-ctx.Done():
		return "", "", 0, ctx.Err()
	case <-time.After(100 * time.Millisecond):
	}

	containerID = fmt.Sprintf("ctf-%d", time.Now().UnixNano())
	networkID = fmt.Sprintf("net-%d", time.Now().UnixNano())
	port = s.config.PortRangeStart + int(time.Now().Unix()%int64(s.config.PortRangeEnd-s.config.PortRangeStart))

	return containerID, networkID, port, nil
}

// RemoveContainer 删除容器
func (s *Service) RemoveContainer(containerID string) error {
	// TODO: 实际的 Docker 容器删除逻辑
	s.logger.Info("删除容器（模拟）", zap.String("container_id", containerID))
	return nil
}

// RemoveNetwork 删除网络
func (s *Service) RemoveNetwork(networkID string) error {
	// TODO: 实际的 Docker 网络删除逻辑
	s.logger.Info("删除网络（模拟）", zap.String("network_id", networkID))
	return nil
}

func (s *Service) CreateInstance(userID, challengeID int64) (*dto.InstanceResp, error) {
	// 检查用户并发实例数
	instances, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if len(instances) >= s.config.MaxConcurrentPerUser {
		return nil, errcode.ErrInstanceLimitExceeded
	}

	// 创建实例记录
	instance := &model.Instance{
		UserID:      userID,
		ChallengeID: challengeID,
		ContainerID: fmt.Sprintf("container-%d-%d", userID, time.Now().Unix()),
		Status:      model.InstanceStatusCreating,
		ExpiresAt:   time.Now().Add(s.config.DefaultTTL),
		MaxExtends:  s.config.MaxExtends,
	}

	if err := s.repo.Create(instance); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	// TODO: 实际创建容器（B9-B11 实现后集成）
	instance.Status = model.InstanceStatusRunning
	instance.AccessURL = fmt.Sprintf("http://localhost:3%04d", 1000+instance.ID)
	if err := s.repo.UpdateStatus(instance.ID, model.InstanceStatusRunning); err != nil {
		s.logger.Error("更新实例状态失败", zap.Error(err))
		return nil, errcode.ErrInternal.WithCause(err)
	}

	s.logger.Info("创建实例",
		zap.Int64("user_id", userID),
		zap.Int64("challenge_id", challengeID),
		zap.Int64("instance_id", instance.ID),
		zap.Time("expires_at", instance.ExpiresAt))

	return toInstanceResp(instance), nil
}

func (s *Service) DestroyInstance(instanceID, userID int64) error {
	instance, err := s.repo.FindByID(instanceID)
	if err != nil {
		return errcode.ErrInstanceNotFound
	}
	if instance.UserID != userID {
		return errcode.ErrForbidden
	}

	// TODO: 停止并删除容器（B9-B11 实现后集成）
	// if err := s.containerEngine.RemoveContainer(ctx, instance.ContainerID); err != nil {
	//     s.logger.Error("删除容器失败", zap.Error(err))
	// }
	// if err := s.containerEngine.RemoveNetwork(ctx, instance.NetworkID); err != nil {
	//     s.logger.Error("删除网络失败", zap.Error(err))
	// }

	s.logger.Info("销毁实例",
		zap.Int64("instance_id", instanceID),
		zap.Int64("user_id", userID))

	return s.repo.UpdateStatus(instanceID, model.InstanceStatusStopped)
}

func (s *Service) ExtendInstance(instanceID, userID int64) error {
	instance, err := s.repo.FindByID(instanceID)
	if err != nil {
		return errcode.ErrInstanceNotFound
	}
	if instance.UserID != userID {
		return errcode.ErrForbidden
	}
	if instance.Status != model.InstanceStatusRunning {
		return errcode.ErrInstanceExpired
	}

	// 使用原子更新避免并发竞争
	if err := s.repo.AtomicExtend(instanceID, userID, s.config.MaxExtends, s.config.ExtendDuration); err != nil {
		return err
	}

	s.logger.Info("延时实例",
		zap.Int64("instance_id", instanceID),
		zap.Int("extend_count", instance.ExtendCount+1),
		zap.Time("new_expires_at", instance.ExpiresAt.Add(s.config.ExtendDuration)))

	return nil
}

func (s *Service) GetUserInstances(userID int64) ([]*dto.InstanceInfo, error) {
	instances, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := make([]*dto.InstanceInfo, len(instances))
	for i, inst := range instances {
		result[i] = toInstanceInfo(inst)
	}
	return result, nil
}

func (s *Service) CleanExpiredInstances(ctx context.Context) error {
	instances, err := s.repo.FindExpired()
	if err != nil {
		return err
	}

	for _, inst := range instances {
		s.logger.Info("清理过期实例", zap.Int64("instance_id", inst.ID))

		// TODO: 实际清理容器资源（B9-B11 实现后集成）
		// if err := s.containerEngine.RemoveContainer(ctx, inst.ContainerID); err != nil {
		//     s.logger.Error("删除容器失败", zap.Error(err))
		// }
		// if inst.NetworkID != "" {
		//     if err := s.containerEngine.RemoveNetwork(ctx, inst.NetworkID); err != nil {
		//         s.logger.Error("删除网络失败", zap.Error(err))
		//     }
		// }

		s.repo.UpdateStatus(inst.ID, model.InstanceStatusExpired)
	}
	return nil
}

func (s *Service) CleanupOrphans(ctx context.Context) error {
	// TODO: 实现孤儿容器清理（B9-B11 实现后集成）
	// 1. 获取所有运行中的容器列表
	// 2. 查询数据库中的实例记录
	// 3. 找出数据库中不存在但容器仍在运行的孤儿容器
	// 4. 删除孤儿容器和网络
	s.logger.Info("孤儿容器清理功能待实现")
	return nil
}

func toInstanceResp(inst *model.Instance) *dto.InstanceResp {
	return &dto.InstanceResp{
		ID:          inst.ID,
		ChallengeID: inst.ChallengeID,
		Status:      inst.Status,
		AccessURL:   inst.AccessURL,
		ExpiresAt:   inst.ExpiresAt,
		ExtendCount: inst.ExtendCount,
		MaxExtends:  inst.MaxExtends,
		CreatedAt:   inst.CreatedAt,
	}
}

func toInstanceInfo(inst *model.Instance) *dto.InstanceInfo {
	remaining := int64(time.Until(inst.ExpiresAt).Seconds())
	if remaining < 0 {
		remaining = 0
	}
	return &dto.InstanceInfo{
		ID:            inst.ID,
		ChallengeID:   inst.ChallengeID,
		Status:        inst.Status,
		AccessURL:     inst.AccessURL,
		ExpiresAt:     inst.ExpiresAt,
		RemainingTime: remaining,
		ExtendCount:   inst.ExtendCount,
		MaxExtends:    inst.MaxExtends,
		CreatedAt:     inst.CreatedAt,
	}
}
