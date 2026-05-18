package queries

import (
	"context"
	"strings"
	"time"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instancedomain "ctf-platform/internal/module/instance/domain"
	instanceports "ctf-platform/internal/module/instance/ports"
	"ctf-platform/pkg/errcode"
)

type InstanceService struct {
	repo instanceQueryRepository
	cfg  *config.ContainerConfig
}

type instanceQueryRepository interface {
	instanceports.InstanceUserLookupRepository
	instanceports.InstanceAccessRepository
	instanceports.UserVisibleInstanceRepository
	instanceports.TeacherInstanceQueryRepository
}

func NewInstanceService(repo instanceQueryRepository, cfg *config.ContainerConfig) *InstanceService {
	if cfg == nil {
		cfg = &config.ContainerConfig{}
	}
	return &InstanceService{repo: repo, cfg: cfg}
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
	if visibleInstanceStatus(instance.Status, instance.ExpiresAt, time.Now()) != instancecontracts.InstanceStatusRunning || strings.TrimSpace(instance.AccessURL) == "" {
		return "", errcode.ErrInstanceExpired
	}

	return instance.AccessURL, nil
}

func (s *InstanceService) GetUserInstances(ctx context.Context, userID int64) ([]*instancecontracts.InstanceInfo, error) {
	ctx = normalizeContext(ctx)

	instances, err := s.repo.ListVisibleByUser(ctx, userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	now := time.Now()
	result := make([]*instancecontracts.InstanceInfo, len(instances))
	for idx, inst := range instances {
		result[idx] = toInstanceInfo(inst, now, s.cfg.PublicHost, s.cfg.AccessHost)
	}
	return result, nil
}

func (s *InstanceService) ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query instancecontracts.TeacherInstanceListQuery) ([]instancecontracts.TeacherInstanceItem, error) {
	ctx = normalizeContext(ctx)

	filter := instanceports.TeacherInstanceFilter{}
	filter.ClassName = strings.TrimSpace(query.ClassName)
	filter.Keyword = strings.TrimSpace(query.Keyword)
	filter.StudentNo = strings.TrimSpace(query.StudentNo)

	if requesterRole != identitycontracts.RoleAdmin {
		requester, err := s.repo.FindUserByID(ctx, requesterID)
		if err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		if requester == nil {
			return nil, errcode.ErrUnauthorized
		}

		className := strings.TrimSpace(requester.ClassName)
		if className == "" {
			return []instancecontracts.TeacherInstanceItem{}, nil
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
	result := make([]instancecontracts.TeacherInstanceItem, len(items))
	for idx, item := range items {
		result[idx] = toTeacherInstanceItem(item, now, s.cfg.PublicHost, s.cfg.AccessHost)
	}

	return result, nil
}

func toInstanceInfo(inst instanceports.UserVisibleInstanceRow, now time.Time, publicHost, accessHost string) *instancecontracts.InstanceInfo {
	accessURL := model.ResolveRuntimePublicAccessURL(inst.AccessURL, publicHost, accessHost)
	if inst.ContestMode == contestcontracts.ContestModeAWD {
		accessURL = ""
	}
	return &instancecontracts.InstanceInfo{
		ID:               inst.ID,
		ContestMode:      inst.ContestMode,
		ChallengeID:      inst.ChallengeID,
		ChallengeTitle:   inst.ChallengeTitle,
		Category:         inst.Category,
		Difficulty:       inst.Difficulty,
		FlagType:         inst.FlagType,
		Status:           visibleInstanceStatus(inst.Status, inst.ExpiresAt, now),
		ShareScope:       inst.ShareScope,
		AccessURL:        accessURL,
		Access:           instancecontracts.BuildInstanceAccessInfo(accessURL),
		ExpiresAt:        inst.ExpiresAt,
		RemainingTime:    instancedomain.RemainingTime(inst.ExpiresAt, now),
		ExtendCount:      inst.ExtendCount,
		MaxExtends:       inst.MaxExtends,
		RemainingExtends: instancedomain.RemainingExtends(inst.MaxExtends, inst.ExtendCount),
		CreatedAt:        inst.CreatedAt,
	}
}

func toTeacherInstanceItem(item instanceports.TeacherInstanceRow, now time.Time, publicHost, accessHost string) instancecontracts.TeacherInstanceItem {
	accessURL := model.ResolveRuntimePublicAccessURL(item.AccessURL, publicHost, accessHost)
	return instancecontracts.TeacherInstanceItem{
		ID:              item.ID,
		StudentID:       item.StudentID,
		StudentName:     item.StudentName,
		StudentUsername: item.StudentUsername,
		StudentNo:       item.StudentNo,
		ClassName:       item.ClassName,
		ChallengeID:     item.ChallengeID,
		ChallengeTitle:  item.ChallengeTitle,
		Status:          visibleInstanceStatus(item.Status, item.ExpiresAt, now),
		AccessURL:       accessURL,
		Access:          instancecontracts.BuildInstanceAccessInfo(accessURL),
		ExpiresAt:       item.ExpiresAt,
		RemainingTime:   instancedomain.RemainingTime(item.ExpiresAt, now),
		ExtendCount:     item.ExtendCount,
		MaxExtends:      item.MaxExtends,
		CreatedAt:       item.CreatedAt,
	}
}

func visibleInstanceStatus(status string, expiresAt, now time.Time) string {
	if status == instancecontracts.InstanceStatusRunning && !expiresAt.After(now) {
		return instancecontracts.InstanceStatusExpired
	}
	return status
}

func normalizeContext(ctx context.Context) context.Context {
	return ctx
}
