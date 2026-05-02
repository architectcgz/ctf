package queries

import (
	"context"
	"strings"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	runtimedomain "ctf-platform/internal/module/runtime/domain"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	"ctf-platform/pkg/errcode"
)

type InstanceService struct {
	repo runtimeports.InstanceRepository
}

func NewInstanceService(repo runtimeports.InstanceRepository) *InstanceService {
	return &InstanceService{repo: repo}
}

func (s *InstanceService) GetAccessURL(ctx context.Context, instanceID, userID int64) (string, error) {
	ctx = normalizeContext(ctx)

	instance, err := s.repo.FindAccessibleByIDForUser(ctx, instanceID, userID)
	if err != nil {
		return "", errcode.ErrInternal.WithCause(err)
	}
	if instance == nil {
		return "", errcode.ErrForbidden
	}
	if visibleInstanceStatus(instance.Status, instance.ExpiresAt, time.Now()) != model.InstanceStatusRunning || strings.TrimSpace(instance.AccessURL) == "" {
		return "", errcode.ErrInstanceExpired
	}

	return instance.AccessURL, nil
}

func (s *InstanceService) GetUserInstances(ctx context.Context, userID int64) ([]*dto.InstanceInfo, error) {
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

func (s *InstanceService) ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error) {
	ctx = normalizeContext(ctx)

	filter := runtimeports.TeacherInstanceFilter{}
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
		resp := runtimeResponseMapper.ToTeacherInstanceItemPtr(&item)
		resp.Status = visibleInstanceStatus(item.Status, item.ExpiresAt, now)
		resp.Access = dto.BuildInstanceAccessInfo(item.AccessURL)
		resp.RemainingTime = runtimedomain.RemainingTime(item.ExpiresAt, now)
		result[idx] = *resp
	}

	return result, nil
}

func toInstanceInfo(inst runtimeports.UserVisibleInstanceRow, now time.Time) *dto.InstanceInfo {
	accessURL := inst.AccessURL
	if inst.ContestMode == model.ContestModeAWD {
		accessURL = ""
	}
	resp := runtimeResponseMapper.ToInstanceInfoPtr(&inst)
	resp.Status = visibleInstanceStatus(inst.Status, inst.ExpiresAt, now)
	resp.AccessURL = accessURL
	resp.Access = dto.BuildInstanceAccessInfo(accessURL)
	resp.RemainingTime = runtimedomain.RemainingTime(inst.ExpiresAt, now)
	resp.RemainingExtends = runtimedomain.RemainingExtends(inst.MaxExtends, inst.ExtendCount)
	return resp
}

func visibleInstanceStatus(status string, expiresAt, now time.Time) string {
	if status == model.InstanceStatusRunning && !expiresAt.After(now) {
		return model.InstanceStatusExpired
	}
	return status
}

func normalizeContext(ctx context.Context) context.Context {
	return ctx
}
