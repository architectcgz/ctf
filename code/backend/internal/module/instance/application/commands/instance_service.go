package commands

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instancedomain "ctf-platform/internal/module/instance/domain"
	instanceports "ctf-platform/internal/module/instance/ports"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
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
		return apperror.ErrInternal.WithCause(err)
	}
	if instance == nil {
		return apperror.ErrForbidden
	}
	if instance.ShareScope == instancecontracts.ShareScopeShared {
		return apperror.ErrForbidden
	}
	if isAWDTeamServiceInstance(instance) {
		return apperror.ErrForbidden
	}

	s.logger.Info("销毁实例", zap.Int64("instance_id", instanceID), zap.Int64("user_id", userID))

	return s.destroyManagedInstance(ctx, instance)
}

func (s *InstanceService) ExtendInstance(ctx context.Context, instanceID, userID int64) (*instancecontracts.InstanceResp, error) {
	ctx = normalizeContext(ctx)

	instance, err := s.repo.FindAccessibleByIDForUser(ctx, instanceID, userID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if instance == nil {
		return nil, apperror.ErrForbidden
	}
	if instance.ShareScope == instancecontracts.ShareScopeShared {
		return nil, apperror.ErrForbidden
	}
	if isAWDTeamServiceInstance(instance) {
		return nil, apperror.ErrForbidden
	}
	if instance.Status != instancecontracts.InstanceStatusRunning || !instance.ExpiresAt.After(time.Now()) {
		return nil, instancecontracts.ErrInstanceExpired
	}

	if err := s.repo.AtomicExtendByID(ctx, instanceID, s.config.MaxExtends, s.config.ExtendDuration); err != nil {
		return nil, err
	}

	updatedInstance, err := s.repo.FindAccessibleByIDForUser(ctx, instanceID, userID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if updatedInstance == nil {
		return nil, apperror.ErrForbidden
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
		return instancecontracts.ErrInstanceNotFound
	}

	if requesterRole != identitycontracts.RoleAdmin {
		requester, err := s.repo.FindUserByID(ctx, requesterID)
		if err != nil {
			return apperror.ErrInternal.WithCause(err)
		}
		if requester == nil {
			return apperror.ErrUnauthorized
		}

		owner, err := s.repo.FindUserByID(ctx, instance.UserID)
		if err != nil {
			return apperror.ErrInternal.WithCause(err)
		}
		if owner == nil || strings.TrimSpace(owner.ClassName) == "" || owner.ClassName != requester.ClassName {
			return apperror.ErrForbidden
		}
	}

	s.logger.Info("教师销毁实例",
		zap.Int64("instance_id", instanceID),
		zap.Int64("requester_id", requesterID),
		zap.String("requester_role", requesterRole))

	return s.destroyManagedInstance(ctx, instance)
}

func (s *InstanceService) destroyManagedInstance(ctx context.Context, instance *instancecontracts.Instance) error {
	if instance == nil {
		return nil
	}
	if instance.Status == instancecontracts.InstanceStatusStopped ||
		instance.Status == instancecontracts.InstanceStatusExpired ||
		instance.Status == instancecontracts.InstanceStatusStopping {
		return nil
	}
	stopping, err := s.repo.MarkStopping(ctx, instance.ID)
	if err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	if !stopping {
		return nil
	}
	return nil
}

func isAWDTeamServiceInstance(instance *instancecontracts.Instance) bool {
	return instance != nil && instance.ContestID != nil && instance.TeamID != nil && instance.ServiceID != nil
}

func (s *InstanceService) toInstanceResp(inst *instancecontracts.Instance) *instancecontracts.InstanceResp {
	if inst == nil {
		return nil
	}
	accessURL := runtimecontracts.ResolveRuntimePublicAccessURL(inst.AccessURL, s.config.PublicHost, s.config.AccessHost)
	return &instancecontracts.InstanceResp{
		ID:               inst.ID,
		ChallengeID:      inst.ChallengeID,
		Status:           inst.Status,
		ShareScope:       inst.ShareScope,
		AccessURL:        accessURL,
		Access:           instancecontracts.BuildInstanceAccessInfo(accessURL),
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
