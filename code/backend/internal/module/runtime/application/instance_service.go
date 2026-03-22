package application

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

// InstanceService 收口实例 HTTP 用例，负责权限校验、查询组装和销毁编排。
type InstanceService struct {
	repo    InstanceRepository
	cleaner RuntimeCleaner
	config  *config.ContainerConfig
	logger  *zap.Logger
}

// NewInstanceService 创建实例 HTTP 用例服务。
func NewInstanceService(repo InstanceRepository, cleaner RuntimeCleaner, cfg *config.ContainerConfig, logger *zap.Logger) *InstanceService {
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

// DestroyInstanceWithContext 销毁当前用户可访问的实例。
func (s *InstanceService) DestroyInstanceWithContext(ctx context.Context, instanceID, userID int64) error {
	ctx = normalizeContext(ctx)

	instance, err := s.repo.FindAccessibleByIDForUser(ctx, instanceID, userID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if instance == nil {
		return errcode.ErrForbidden
	}

	s.logger.Info("销毁实例",
		zap.Int64("instance_id", instanceID),
		zap.Int64("user_id", userID))

	return s.destroyManagedInstanceWithContext(ctx, instance)
}

// ExtendInstanceWithContext 延长当前用户可访问实例的有效期。
func (s *InstanceService) ExtendInstanceWithContext(ctx context.Context, instanceID, userID int64) (*dto.InstanceResp, error) {
	ctx = normalizeContext(ctx)

	instance, err := s.repo.FindAccessibleByIDForUser(ctx, instanceID, userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if instance == nil {
		return nil, errcode.ErrForbidden
	}
	if instance.Status != model.InstanceStatusRunning {
		return nil, errcode.ErrInstanceExpired
	}

	if err := s.repo.AtomicExtendByIDWithContext(ctx, instanceID, s.config.MaxExtends, s.config.ExtendDuration); err != nil {
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

	return toInstanceResp(updatedInstance), nil
}

// GetAccessURLWithContext 返回当前用户可访问实例的入口地址。
func (s *InstanceService) GetAccessURLWithContext(ctx context.Context, instanceID, userID int64) (string, error) {
	ctx = normalizeContext(ctx)

	instance, err := s.repo.FindAccessibleByIDForUser(ctx, instanceID, userID)
	if err != nil {
		return "", errcode.ErrInternal.WithCause(err)
	}
	if instance == nil {
		return "", errcode.ErrForbidden
	}
	if instance.Status != model.InstanceStatusRunning || strings.TrimSpace(instance.AccessURL) == "" {
		return "", errcode.ErrInstanceExpired
	}

	return instance.AccessURL, nil
}

// GetUserInstancesWithContext 返回当前用户可见的实例列表。
func (s *InstanceService) GetUserInstancesWithContext(ctx context.Context, userID int64) ([]*dto.InstanceInfo, error) {
	ctx = normalizeContext(ctx)

	instances, err := s.repo.ListVisibleByUser(ctx, userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	now := time.Now()
	result := make([]*dto.InstanceInfo, len(instances))
	for idx, inst := range instances {
		result[idx] = toInstanceInfo(inst, now)
	}
	return result, nil
}

// ListTeacherInstances 返回教师端可见的实例列表。
func (s *InstanceService) ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error) {
	ctx = normalizeContext(ctx)

	filter := TeacherInstanceFilter{}
	if query != nil {
		filter.ClassName = strings.TrimSpace(query.ClassName)
		filter.Keyword = strings.TrimSpace(query.Keyword)
		filter.StudentNo = strings.TrimSpace(query.StudentNo)
	}

	if requesterRole != model.RoleAdmin {
		requester, err := s.repo.FindUserByID(ctx, requesterID)
		if err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		if requester == nil {
			return nil, errcode.ErrUnauthorized
		}

		className := strings.TrimSpace(requester.ClassName)
		if className == "" {
			return []dto.TeacherInstanceItem{}, nil
		}
		if filter.ClassName != "" && filter.ClassName != className {
			return nil, errcode.ErrForbidden
		}
		filter.ClassName = className
	}

	items, err := s.repo.ListTeacherInstances(ctx, filter)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	now := time.Now()
	result := make([]dto.TeacherInstanceItem, len(items))
	for idx, item := range items {
		result[idx] = dto.TeacherInstanceItem{
			ID:              item.ID,
			StudentID:       item.StudentID,
			StudentName:     item.StudentName,
			StudentUsername: item.StudentUsername,
			StudentNo:       item.StudentNo,
			ClassName:       item.ClassName,
			ChallengeID:     item.ChallengeID,
			ChallengeTitle:  item.ChallengeTitle,
			Status:          item.Status,
			AccessURL:       item.AccessURL,
			ExpiresAt:       item.ExpiresAt,
			RemainingTime:   calculateRemainingTime(item.ExpiresAt, now),
			ExtendCount:     item.ExtendCount,
			MaxExtends:      item.MaxExtends,
			CreatedAt:       item.CreatedAt,
		}
	}

	return result, nil
}

// DestroyTeacherInstance 按教师班级范围销毁实例。
func (s *InstanceService) DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error {
	ctx = normalizeContext(ctx)

	instance, err := s.repo.FindByID(instanceID)
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

	return s.destroyManagedInstanceWithContext(ctx, instance)
}

func (s *InstanceService) destroyManagedInstanceWithContext(ctx context.Context, instance *model.Instance) error {
	if s.cleaner != nil {
		if err := s.cleaner.CleanupRuntimeWithContext(ctx, instance); err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
	}
	if err := s.repo.UpdateStatusAndReleasePort(instance.ID, model.InstanceStatusStopped); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}

func toInstanceResp(inst *model.Instance) *dto.InstanceResp {
	return &dto.InstanceResp{
		ID:               inst.ID,
		ChallengeID:      inst.ChallengeID,
		Status:           inst.Status,
		AccessURL:        inst.AccessURL,
		ExpiresAt:        inst.ExpiresAt,
		ExtendCount:      inst.ExtendCount,
		MaxExtends:       inst.MaxExtends,
		RemainingExtends: remainingExtends(inst.MaxExtends, inst.ExtendCount),
		CreatedAt:        inst.CreatedAt,
	}
}

func toInstanceInfo(inst UserVisibleInstanceRow, now time.Time) *dto.InstanceInfo {
	return &dto.InstanceInfo{
		ID:               inst.ID,
		ChallengeID:      inst.ChallengeID,
		ChallengeTitle:   inst.ChallengeTitle,
		Category:         inst.Category,
		Difficulty:       inst.Difficulty,
		FlagType:         inst.FlagType,
		Status:           inst.Status,
		AccessURL:        inst.AccessURL,
		ExpiresAt:        inst.ExpiresAt,
		RemainingTime:    calculateRemainingTime(inst.ExpiresAt, now),
		ExtendCount:      inst.ExtendCount,
		MaxExtends:       inst.MaxExtends,
		RemainingExtends: remainingExtends(inst.MaxExtends, inst.ExtendCount),
		CreatedAt:        inst.CreatedAt,
	}
}

func remainingExtends(maxExtends int, extendCount int) int {
	remaining := maxExtends - extendCount
	if remaining < 0 {
		return 0
	}
	return remaining
}

func calculateRemainingTime(expiresAt, now time.Time) int64 {
	remaining := int64(expiresAt.Sub(now).Seconds())
	if remaining < 0 {
		return 0
	}
	return remaining
}

func normalizeContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
