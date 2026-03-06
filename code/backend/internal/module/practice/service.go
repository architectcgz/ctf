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
	challengeRepo    *challenge.Repository
	instanceRepo     *container.Repository
	containerService *container.Service
	config           *config.Config
	logger           *zap.Logger
}

func NewService(
	challengeRepo *challenge.Repository,
	instanceRepo *container.Repository,
	containerService *container.Service,
	config *config.Config,
	logger *zap.Logger,
) *Service {
	return &Service{
		challengeRepo:    challengeRepo,
		instanceRepo:     instanceRepo,
		containerService: containerService,
		config:           config,
		logger:           logger,
	}
}

func (s *Service) StartChallenge(userID, challengeID int64) (*dto.InstanceResp, error) {
	// 1. 检查是否已有该靶场的运行中实例
	existingInstance, err := s.instanceRepo.FindByUserAndChallenge(userID, challengeID)
	if err == nil && existingInstance != nil {
		return toInstanceResp(existingInstance), nil
	}

	// 2. 校验并发限制
	instances, err := s.instanceRepo.FindByUserID(userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if len(instances) >= s.config.Container.MaxConcurrentPerUser {
		s.logger.Warn("用户实例数量超限",
			zap.Int64("user_id", userID),
			zap.Int("current", len(instances)),
			zap.Int("limit", s.config.Container.MaxConcurrentPerUser))
		return nil, errcode.ErrInstanceLimitExceeded
	}

	// 3. 查询靶场信息
	chal, err := s.challengeRepo.FindByID(challengeID)
	if err != nil {
		if err.Error() == "record not found" {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if chal.Status != model.ChallengeStatusPublished {
		return nil, errcode.ErrNotFound
	}

	// 4. 生成 Flag
	var flag string
	var nonce string
	if chal.FlagType == model.FlagTypeDynamic {
		nonce, err = crypto.GenerateNonce()
		if err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		globalSecret := s.config.Container.FlagGlobalSecret
		if globalSecret == "" {
			return nil, errcode.ErrInternal.WithCause(fmt.Errorf("Flag 全局密钥未配置"))
		}
		flag = crypto.GenerateDynamicFlag(userID, challengeID, globalSecret, nonce)
		s.logger.Debug("生成动态 Flag",
			zap.Int64("user_id", userID),
			zap.Int64("challenge_id", challengeID),
			zap.String("nonce", nonce))
	} else if chal.FlagType == model.FlagTypeStatic {
		flag = chal.FlagHash
	}

	// 5. 创建实例记录
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

	// 6. 创建容器（带超时控制）
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Container.CreateTimeout)
	defer cancel()

	if err := s.createContainer(ctx, instance, chal, flag); err != nil {
		s.logger.Error("容器创建失败", zap.Error(err), zap.Int64("instance_id", instance.ID))

		// 清理资源
		if instance.NetworkID != "" {
			s.containerService.RemoveNetwork(instance.NetworkID)
		}
		if instance.ContainerID != "" {
			s.containerService.RemoveContainer(instance.ContainerID)
		}

		s.instanceRepo.UpdateStatus(instance.ID, model.InstanceStatusFailed)
		return nil, err
	}

	// 7. 更新实例状态
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
	// 构建环境变量
	env := map[string]string{
		"FLAG": flag,
	}

	// 调用 container.Service 创建容器
	containerID, networkID, port, err := s.containerService.CreateContainer(ctx, fmt.Sprintf("image-%d", chal.ImageID), env)
	if err != nil {
		return errcode.ErrContainerCreateFailed.WithCause(err)
	}

	// 更新实例信息
	instance.ContainerID = containerID
	instance.NetworkID = networkID
	instance.AccessURL = fmt.Sprintf("http://%s:%d", s.config.Container.PublicHost, port)

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
