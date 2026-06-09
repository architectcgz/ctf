package queries

import (
	"context"
	"strings"
	"time"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instancedomain "ctf-platform/internal/module/instance/domain"
	instanceports "ctf-platform/internal/module/instance/ports"
)

type InstanceService struct {
	repo       instanceQueryRepository
	cfg        *config.ContainerConfig
	pagination config.PaginationConfig
}

type instanceQueryRepository interface {
	instanceports.InstanceUserLookupRepository
	instanceports.InstanceAccessRepository
	instanceports.UserVisibleInstanceRepository
	instanceports.TeacherInstanceQueryRepository
}

func NewInstanceService(
	repo instanceQueryRepository,
	cfg *config.ContainerConfig,
	pagination ...config.PaginationConfig,
) *InstanceService {
	if cfg == nil {
		cfg = &config.ContainerConfig{}
	}
	var resolvedPagination config.PaginationConfig
	if len(pagination) > 0 {
		resolvedPagination = pagination[0]
	}
	return &InstanceService{repo: repo, cfg: cfg, pagination: resolvedPagination}
}

func (s *InstanceService) GetAccessURL(ctx context.Context, instanceID, userID int64) (string, error) {
	ctx = normalizeContext(ctx)

	instance, err := s.repo.FindAccessibleByIDForUser(ctx, instanceID, userID)
	if err != nil {
		return "", apperror.ErrInternal.WithCause(err)
	}
	if instance == nil {
		return "", apperror.ErrForbidden
	}
	if visibleInstanceStatus(instance.Status, instance.ExpiresAt, time.Now().UTC()) != instancecontracts.InstanceStatusRunning || strings.TrimSpace(instance.AccessURL) == "" {
		return "", instancecontracts.ErrInstanceExpired
	}

	return instance.AccessURL, nil
}

func (s *InstanceService) GetUserInstances(ctx context.Context, userID int64) ([]*instancecontracts.InstanceInfo, error) {
	ctx = normalizeContext(ctx)

	instances, err := s.repo.ListVisibleByUser(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	now := time.Now().UTC()
	result := make([]*instancecontracts.InstanceInfo, len(instances))
	for idx, inst := range instances {
		result[idx] = toInstanceInfo(inst, now, s.cfg.PublicHost, s.cfg.AccessHost)
	}
	return result, nil
}

func (s *InstanceService) ListTeacherInstances(
	ctx context.Context,
	requesterID int64,
	requesterRole string,
	query instancecontracts.TeacherInstanceListQuery,
) (*instancecontracts.TeacherInstancePageResult, error) {
	ctx = normalizeContext(ctx)

	page, pageSize := s.normalizeTeacherInstancePagination(query.Page, query.PageSize)
	filter := instanceports.TeacherInstanceFilter{
		ClassName: strings.TrimSpace(query.ClassName),
		Keyword:   strings.TrimSpace(query.Keyword),
		StudentNo: strings.TrimSpace(query.StudentNo),
		Status:    strings.TrimSpace(query.Status),
		Page:      page,
		PageSize:  pageSize,
	}

	if requesterRole != identitycontracts.RoleAdmin {
		requester, err := s.repo.FindUserByID(ctx, requesterID)
		if err != nil {
			return nil, apperror.ErrInternal.WithCause(err)
		}
		if requester == nil {
			return nil, apperror.ErrUnauthorized
		}

		className := strings.TrimSpace(requester.ClassName)
		if className == "" {
			return &instancecontracts.TeacherInstancePageResult{
				List:     []instancecontracts.TeacherInstanceItem{},
				Total:    0,
				Page:     page,
				PageSize: pageSize,
				Summary:  instancecontracts.TeacherInstanceListSummary{},
			}, nil
		}
		if filter.ClassName != "" && filter.ClassName != className {
			return nil, apperror.ErrForbidden
		}
		filter.ClassName = className
	}

	items, err := s.repo.ListTeacherInstances(ctx, filter)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if items == nil {
		items = &instanceports.TeacherInstancePage{}
	}

	now := time.Now().UTC()
	result := make([]instancecontracts.TeacherInstanceItem, len(items.List))
	for idx, item := range items.List {
		result[idx] = toTeacherInstanceItem(item, now, s.cfg.PublicHost, s.cfg.AccessHost)
	}

	return &instancecontracts.TeacherInstancePageResult{
		List:     result,
		Total:    items.Total,
		Page:     page,
		PageSize: pageSize,
		Summary: instancecontracts.TeacherInstanceListSummary{
			TotalCount:        items.Summary.TotalCount,
			RunningCount:      items.Summary.RunningCount,
			ExpiringSoonCount: items.Summary.ExpiringSoonCount,
			WarningCount:      items.Summary.WarningCount,
		},
	}, nil
}

func toInstanceInfo(inst instanceports.UserVisibleInstanceRow, now time.Time, publicHost, accessHost string) *instancecontracts.InstanceInfo {
	status := visibleInstanceStatus(inst.Status, inst.ExpiresAt, now)
	accessURL := ""
	if status == instancecontracts.InstanceStatusRunning {
		accessURL = instancecontracts.ResolveInstancePublicAccessURL(inst.AccessURL, publicHost, accessHost)
	}
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
		Status:           status,
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
	status := visibleInstanceStatus(item.Status, item.ExpiresAt, now)
	accessURL := ""
	if status == instancecontracts.InstanceStatusRunning {
		accessURL = instancecontracts.ResolveInstancePublicAccessURL(item.AccessURL, publicHost, accessHost)
	}
	return instancecontracts.TeacherInstanceItem{
		ID:              item.ID,
		StudentID:       item.StudentID,
		StudentName:     item.StudentName,
		StudentUsername: item.StudentUsername,
		StudentNo:       item.StudentNo,
		ClassName:       item.ClassName,
		ChallengeID:     item.ChallengeID,
		ChallengeTitle:  item.ChallengeTitle,
		Status:          status,
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
	if status == instancecontracts.InstanceStatusStopping {
		return "destroying"
	}
	return status
}

func normalizeContext(ctx context.Context) context.Context {
	return ctx
}

func (s *InstanceService) normalizeTeacherInstancePagination(page, pageSize int) (int, int) {
	normalizedPage := page
	if normalizedPage < 1 {
		normalizedPage = 1
	}

	defaultPageSize := s.pagination.DefaultPageSize
	if defaultPageSize < 1 {
		defaultPageSize = 20
	}
	maxPageSize := s.pagination.MaxPageSize
	if maxPageSize < 1 {
		maxPageSize = 100
	}

	normalizedPageSize := pageSize
	if normalizedPageSize < 1 {
		normalizedPageSize = defaultPageSize
	}
	if normalizedPageSize > maxPageSize {
		normalizedPageSize = maxPageSize
	}

	return normalizedPage, normalizedPageSize
}
