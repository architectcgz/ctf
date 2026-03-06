package practice

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge"
	"ctf-platform/internal/module/container"
	"ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
)

type Service struct {
	challengeRepo *challenge.Repository
	instanceRepo  *container.Repository
	config        *config.Config
	logger        *zap.Logger
}

func NewService(
	challengeRepo *challenge.Repository,
	instanceRepo *container.Repository,
	config *config.Config,
	logger *zap.Logger,
) *Service {
	return &Service{
		challengeRepo: challengeRepo,
		instanceRepo:  instanceRepo,
		config:        config,
		logger:        logger,
	}
}

func (s *Service) StartChallenge(userID, challengeID int64) (*dto.InstanceResp, error) {
	// 1. 校验并发限制
	instances, err := s.instanceRepo.FindByUserID(userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if len(instances) >= s.config.Container.MaxConcurrentPerUser {
		return nil, errcode.ErrInstanceLimitExceeded
	}

	// 2. 查询靶场信息
	chal, err := s.challengeRepo.FindByID(challengeID)
	if err != nil {
		return nil, errcode.ErrNotFound
	}
	if chal.Status != model.ChallengeStatusPublished {
		return nil, errcode.ErrNotFound
	}

	// 3. 生成动态 Flag（如果需要）
	var flag string
	var nonce string
	if chal.FlagType == model.FlagTypeDynamic {
		nonce, err = crypto.GenerateNonce()
		if err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		flag = crypto.GenerateDynamicFlag(userID, challengeID, chal.FlagSalt, nonce)
	}

	// 4. 创建实例记录
	instance := &model.Instance{
		UserID:      userID,
		ChallengeID: challengeID,
		Status:      model.InstanceStatusCreating,
		Nonce:       nonce,
		ExpiresAt:   time.Now().Add(s.config.Container.DefaultTTL),
		MaxExtends:  s.config.Container.MaxExtends,
	}

	if err := s.instanceRepo.Create(instance); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	// 5. 创建容器（带超时控制）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.createContainer(ctx, instance, chal, flag); err != nil {
		s.logger.Error("容器创建失败", zap.Error(err), zap.Int64("instance_id", instance.ID))
		s.instanceRepo.UpdateStatus(instance.ID, model.InstanceStatusFailed)
		return nil, err
	}

	// 6. 更新实例状态
	instance.Status = model.InstanceStatusRunning
	if err := s.instanceRepo.UpdateStatus(instance.ID, model.InstanceStatusRunning); err != nil {
		s.logger.Error("更新实例状态失败", zap.Error(err))
		return nil, errcode.ErrInternal.WithCause(err)
	}

	s.logger.Info("实例启动成功",
		zap.Int64("user_id", userID),
		zap.Int64("challenge_id", challengeID),
		zap.Int64("instance_id", instance.ID))

	return toInstanceResp(instance), nil
}

func (s *Service) createContainer(ctx context.Context, instance *model.Instance, chal *model.Challenge, flag string) error {
	// TODO: 实际的 Docker 容器创建逻辑
	// 这里使用模拟实现，实际应调用 Docker SDK

	// 模拟容器创建延迟
	select {
	case <-ctx.Done():
		return errcode.ErrContainerCreateFailed.WithCause(ctx.Err())
	case <-time.After(100 * time.Millisecond):
	}

	// 生成容器 ID 和访问地址
	instance.ContainerID = fmt.Sprintf("ctf-%d-%d-%d", instance.UserID, instance.ChallengeID, time.Now().Unix())
	instance.NetworkID = fmt.Sprintf("net-%d", instance.ID)

	// 分配端口
	port := s.config.Container.PortRangeStart + int(instance.ID%int64(s.config.Container.PortRangeEnd-s.config.Container.PortRangeStart))
	instance.AccessURL = fmt.Sprintf("http://localhost:%d", port)

	return nil
}

func (s *Service) GetInstance(instanceID, userID int64) (*dto.InstanceInfo, error) {
	instance, err := s.instanceRepo.FindByID(instanceID)
	if err != nil {
		return nil, errcode.ErrInstanceNotFound
	}
	if instance.UserID != userID {
		return nil, errcode.ErrForbidden
	}

	return toInstanceInfo(instance), nil
}

func (s *Service) ListUserInstances(userID int64) ([]*dto.InstanceInfo, error) {
	instances, err := s.instanceRepo.FindByUserID(userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := make([]*dto.InstanceInfo, len(instances))
	for i, inst := range instances {
		result[i] = toInstanceInfo(inst)
	}
	return result, nil
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
