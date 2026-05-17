package commands

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	instancedomain "ctf-platform/internal/module/instance/domain"
	instanceports "ctf-platform/internal/module/instance/ports"
	"ctf-platform/pkg/errcode"
)

type InstanceService struct {
	repo    instanceCommandRepository
	cleaner instanceports.RuntimeCleaner
	config  *config.ContainerConfig
	logger  *zap.Logger
}

type instanceCommandRepository interface {
	instanceports.InstanceLookupRepository
	instanceports.InstanceUserLookupRepository
	instanceports.InstanceAccessRepository
	instanceports.InstanceExtendRepository
	instanceports.InstanceStatusRepository
}

func NewInstanceService(repo instanceCommandRepository, cleaner instanceports.RuntimeCleaner, cfg *config.ContainerConfig, logger *zap.Logger) *InstanceService {
	if logger == nil {
		logger = zap.NewNop()
	}
	if cfg == nil {
		cfg = &config.ContainerConfig{}
	}
	return &InstanceService{
		repo:    repo,
		cleaner: cleaner,
		config:  cfg,
		logger:  logger,
	}
}

func (s *InstanceService) DestroyInstance(ctx context.Context, instanceID, userID int64) error {
	ctx = normalizeContext(ctx)

	instance, err := s.repo.FindAccessibleByIDForUser(ctx, instanceID, userID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if instance == nil {
		return errcode.ErrForbidden
	}
	if instance.ShareScope == model.InstanceSharingShared {
		return errcode.ErrForbidden
	}
	if isAWDTeamServiceInstance(instance) {
		return errcode.ErrForbidden
	}

	s.logger.Info("销毁实例", zap.Int64("instance_id", instanceID), zap.Int64("user_id", userID))

	return s.destroyManagedInstance(ctx, instance)
}

func (s *InstanceService) ExtendInstance(ctx context.Context, instanceID, userID int64) (*dto.InstanceResp, error) {
	ctx = normalizeContext(ctx)

	instance, err := s.repo.FindAccessibleByIDForUser(ctx, instanceID, userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if instance == nil {
		return nil, errcode.ErrForbidden
	}
	if instance.ShareScope == model.InstanceSharingShared {
		return nil, errcode.ErrForbidden
	}
	if isAWDTeamServiceInstance(instance) {
		return nil, errcode.ErrForbidden
	}
	if instance.Status != model.InstanceStatusRunning || !instance.ExpiresAt.After(time.Now()) {
		return nil, errcode.ErrInstanceExpired
	}

	if err := s.repo.AtomicExtendByID(ctx, instanceID, s.config.MaxExtends, s.config.ExtendDuration); err != nil {
		return nil, err
	}

	updatedInstance, err := s.repo.FindAccessibleByIDForUser(ctx, instanceID, userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if updatedInstance == nil {
		return nil, errcode.ErrForbidden
	}

	s.logger.Info("延时实例",
		zap.Int64("instance_id", instanceID),
		zap.Int("extend_count", instance.ExtendCount+1),
		zap.Time("new_expires_at", instance.ExpiresAt.Add(s.config.ExtendDuration)))

	return s.toInstanceResp(updatedInstance), nil
}

func (s *InstanceService) DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error {
	ctx = normalizeContext(ctx)

	instance, err := s.repo.FindByID(ctx, instanceID)
	if err != nil {
		return errcode.ErrInstanceNotFound
	}

	if requesterRole != model.RoleAdmin {
		requester, err := s.repo.FindUserByID(ctx, requesterID)
		if err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
		if requester == nil {
			return errcode.ErrUnauthorized
		}

		owner, err := s.repo.FindUserByID(ctx, instance.UserID)
		if err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
		if owner == nil || strings.TrimSpace(owner.ClassName) == "" || owner.ClassName != requester.ClassName {
			return errcode.ErrForbidden
		}
	}

	s.logger.Info("教师销毁实例",
		zap.Int64("instance_id", instanceID),
		zap.Int64("requester_id", requesterID),
		zap.String("requester_role", requesterRole))

	return s.destroyManagedInstance(ctx, instance)
}

func (s *InstanceService) destroyManagedInstance(ctx context.Context, instance *model.Instance) error {
	if s.cleaner != nil {
		if err := s.cleaner.CleanupRuntime(ctx, instance); err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
	}
	if err := s.repo.UpdateStatusAndReleasePort(ctx, instance.ID, model.InstanceStatusStopped); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}

func isAWDTeamServiceInstance(instance *model.Instance) bool {
	return instance != nil && instance.ContestID != nil && instance.TeamID != nil && instance.ServiceID != nil
}

func (s *InstanceService) toInstanceResp(inst *model.Instance) *dto.InstanceResp {
	if inst == nil {
		return nil
	}
	accessURL := model.ResolveRuntimePublicAccessURL(inst.AccessURL, s.config.PublicHost, s.config.AccessHost)
	return &dto.InstanceResp{
		ID:               inst.ID,
		ChallengeID:      inst.ChallengeID,
		Status:           inst.Status,
		ShareScope:       inst.ShareScope,
		AccessURL:        accessURL,
		Access:           dto.BuildInstanceAccessInfo(accessURL),
		ExpiresAt:        inst.ExpiresAt,
		ExtendCount:      inst.ExtendCount,
		MaxExtends:       inst.MaxExtends,
		RemainingExtends: instancedomain.RemainingExtends(inst.MaxExtends, inst.ExtendCount),
		CreatedAt:        inst.CreatedAt,
	}
}

func normalizeContext(ctx context.Context) context.Context {
	return ctx
}
