package container

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

type Service struct {
	repo   *Repository
	logger *zap.Logger
}

func NewService(repo *Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) CreateInstance(userID, challengeID int64) (*dto.InstanceResp, error) {
	// 检查用户并发实例数
	instances, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if len(instances) >= 3 {
		return nil, errcode.ErrInstanceLimitExceeded
	}

	// 创建实例记录
	instance := &model.Instance{
		UserID:      userID,
		ChallengeID: challengeID,
		ContainerID: fmt.Sprintf("container-%d-%d", userID, time.Now().Unix()),
		Status:      model.InstanceStatusCreating,
		ExpiresAt:   time.Now().Add(2 * time.Hour),
		MaxExtends:  2,
	}

	if err := s.repo.Create(instance); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	// TODO: 实际创建容器（B9-B11 实现后集成）
	instance.Status = model.InstanceStatusRunning
	instance.AccessURL = fmt.Sprintf("http://localhost:3%04d", 1000+instance.ID)
	s.repo.UpdateStatus(instance.ID, model.InstanceStatusRunning)

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

	// TODO: 停止并删除容器
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
	if instance.ExtendCount >= instance.MaxExtends {
		return errcode.ErrExtendLimitExceeded
	}

	newExpiresAt := instance.ExpiresAt.Add(1 * time.Hour)
	return s.repo.UpdateExtend(instanceID, newExpiresAt, instance.ExtendCount+1)
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
		// TODO: 停止容器、删除容器、删除网络
		s.repo.UpdateStatus(inst.ID, model.InstanceStatusExpired)
	}
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
