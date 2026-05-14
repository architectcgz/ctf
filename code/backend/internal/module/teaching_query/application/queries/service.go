package queries

import (
	"context"
	"strings"
	"time"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	queryports "ctf-platform/internal/module/teaching_query/ports"
	commonmapper "ctf-platform/internal/shared/mapperhelper"
	"ctf-platform/pkg/errcode"
)

type QueryService struct {
	users      queryports.TeachingUserLookupRepository
	repo       teachingQueryRepository
	pagination config.PaginationConfig
}

type teachingQueryRepository interface {
	queryports.TeachingClassQueryRepository
	queryports.TeachingStudentDirectoryRepository
}

var _ Service = (*QueryService)(nil)

func NewQueryService(
	users queryports.TeachingUserLookupRepository,
	repo teachingQueryRepository,
	pagination config.PaginationConfig,
) *QueryService {
	return &QueryService{
		users:      users,
		repo:       repo,
		pagination: pagination,
	}
}

func (s *QueryService) ListClasses(
	ctx context.Context,
	requesterID int64,
	requesterRole string,
	query *dto.TeacherClassQuery,
) ([]dto.TeacherClassItem, int64, int, int, error) {
	page, size := s.normalizeClassPagination(query)

	requester, err := s.users.FindUserByID(ctx, requesterID)
	if err != nil {
		return nil, 0, 0, 0, errcode.ErrInternal.WithCause(err)
	}
	if requester == nil {
		return nil, 0, 0, 0, errcode.ErrUnauthorized
	}

	if requesterRole == model.RoleAdmin {
		total, err := s.repo.CountClasses(ctx)
		if err != nil {
			return nil, 0, 0, 0, errcode.ErrInternal.WithCause(err)
		}
		if total == 0 {
			return []dto.TeacherClassItem{}, 0, page, size, nil
		}

		items, err := s.repo.ListClasses(ctx, (page-1)*size, size)
		if err != nil {
			return nil, 0, 0, 0, errcode.ErrInternal.WithCause(err)
		}
		return commonmapper.NonNilSlice(teachingQueryMapper.ToClassItems(items)), total, page, size, nil
	}

	className := strings.TrimSpace(requester.ClassName)
	if className == "" {
		return []dto.TeacherClassItem{}, 0, page, size, nil
	}

	count, err := s.repo.CountStudentsByClass(ctx, className)
	if err != nil {
		return nil, 0, 0, 0, errcode.ErrInternal.WithCause(err)
	}

	if (page-1)*size >= 1 {
		return []dto.TeacherClassItem{}, 1, page, size, nil
	}

	return []dto.TeacherClassItem{{
		Name:         className,
		StudentCount: count,
	}}, 1, page, size, nil
}

func (s *QueryService) normalizeClassPagination(query *dto.TeacherClassQuery) (int, int) {
	page := 1
	size := s.pagination.DefaultPageSize

	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.Size > 0 {
			size = query.Size
		}
	}

	if size < 1 {
		size = 20
	}
	if s.pagination.MaxPageSize > 0 && size > s.pagination.MaxPageSize {
		size = s.pagination.MaxPageSize
	}

	return page, size
}

func (s *QueryService) ListStudents(
	ctx context.Context,
	requesterID int64,
	requesterRole string,
	query *dto.TeacherStudentDirectoryQuery,
) ([]dto.TeacherStudentItem, int64, int, int, error) {
	page, size := s.normalizeStudentPagination(query)

	var requester *model.User
	if requesterRole != model.RoleAdmin {
		var err error
		requester, err = s.users.FindUserByID(ctx, requesterID)
		if err != nil {
			return nil, 0, 0, 0, errcode.ErrInternal.WithCause(err)
		}
		if requester == nil {
			return nil, 0, 0, 0, errcode.ErrUnauthorized
		}
	}

	className := ""
	keyword := ""
	studentNo := ""
	sortKey := "solved_count"
	sortOrder := "desc"
	if query != nil {
		className = strings.TrimSpace(query.ClassName)
		keyword = strings.TrimSpace(query.Keyword)
		studentNo = strings.TrimSpace(query.StudentNo)
		if strings.TrimSpace(query.SortKey) != "" {
			sortKey = strings.TrimSpace(query.SortKey)
		}
		if strings.TrimSpace(query.SortOrder) != "" {
			sortOrder = strings.TrimSpace(query.SortOrder)
		}
	}

	if requesterRole != model.RoleAdmin {
		requesterClassName := strings.TrimSpace(requester.ClassName)
		if requesterClassName == "" {
			return []dto.TeacherStudentItem{}, 0, page, size, nil
		}
		if className == "" {
			className = requesterClassName
		} else if className != requesterClassName {
			return nil, 0, 0, 0, errcode.ErrForbidden
		}
	}

	since := time.Now().AddDate(0, 0, -6)
	startOfDay := time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, since.Location())
	items, total, err := s.repo.ListStudents(ctx, className, keyword, studentNo, sortKey, sortOrder, startOfDay, (page-1)*size, size)
	if err != nil {
		return nil, 0, 0, 0, errcode.ErrInternal.WithCause(err)
	}
	return commonmapper.NonNilSlice(teachingQueryMapper.ToStudentItems(items)), total, page, size, nil
}

func (s *QueryService) normalizeStudentPagination(query *dto.TeacherStudentDirectoryQuery) (int, int) {
	page := 1
	size := s.pagination.DefaultPageSize

	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.Size > 0 {
			size = query.Size
		}
	}

	if size < 1 {
		size = 20
	}
	if s.pagination.MaxPageSize > 0 && size > s.pagination.MaxPageSize {
		size = s.pagination.MaxPageSize
	}

	return page, size
}

func (s *QueryService) ListClassStudents(ctx context.Context, requesterID int64, requesterRole, className string, query *dto.TeacherStudentQuery) ([]dto.TeacherStudentItem, error) {
	normalized := strings.TrimSpace(className)
	if normalized == "" {
		return nil, errcode.New(errcode.ErrInvalidParams.Code, "class_name 不能为空", errcode.ErrInvalidParams.HTTPStatus)
	}

	if err := ensureClassAccess(ctx, s.users, requesterID, requesterRole, normalized); err != nil {
		return nil, err
	}

	studentNo := ""
	keyword := ""
	if query != nil {
		studentNo = strings.TrimSpace(query.StudentNo)
		keyword = strings.TrimSpace(query.Keyword)
	}

	since := time.Now().AddDate(0, 0, -6)
	startOfDay := time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, since.Location())
	items, err := s.repo.ListStudentsByClass(ctx, normalized, keyword, studentNo, startOfDay)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return commonmapper.NonNilSlice(teachingQueryMapper.ToStudentItems(items)), nil
}

type classAccessRepository interface {
	queryports.TeachingUserLookupRepository
}

func ensureClassAccess(ctx context.Context, repo classAccessRepository, requesterID int64, requesterRole, className string) error {
	if requesterRole == model.RoleAdmin {
		return nil
	}

	requester, err := repo.FindUserByID(ctx, requesterID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if requester == nil {
		return errcode.ErrUnauthorized
	}

	if strings.TrimSpace(requester.ClassName) == "" || requester.ClassName != className {
		return errcode.ErrForbidden
	}
	return nil
}
