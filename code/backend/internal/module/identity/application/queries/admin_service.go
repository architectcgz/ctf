package queries

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	"ctf-platform/pkg/errcode"
)

type AdminService struct {
	repo       identitycontracts.UserListRepository
	pagination config.PaginationConfig
	log        *zap.Logger
}

var _ identitycontracts.AdminQueryService = (*AdminService)(nil)

func NewAdminService(repo identitycontracts.UserListRepository, pagination config.PaginationConfig, log *zap.Logger) *AdminService {
	if log == nil {
		log = zap.NewNop()
	}
	return &AdminService{
		repo:       repo,
		pagination: pagination,
		log:        log,
	}
}

func (s *AdminService) ListUsers(ctx context.Context, query *dto.AdminUserQuery) ([]dto.AdminUserResp, int64, int, int, error) {
	page := query.Page
	if page < 1 {
		page = 1
	}
	size := query.Size
	if size < 1 {
		size = s.pagination.DefaultPageSize
	}
	if size > s.pagination.MaxPageSize {
		size = s.pagination.MaxPageSize
	}

	users, total, err := s.repo.List(ctx, identitycontracts.UserListFilter{
		Keyword:   strings.TrimSpace(query.Keyword),
		StudentNo: strings.TrimSpace(query.StudentNo),
		TeacherNo: strings.TrimSpace(query.TeacherNo),
		Role:      strings.TrimSpace(query.Role),
		Status:    strings.TrimSpace(query.Status),
		ClassName: strings.TrimSpace(query.ClassName),
		Offset:    (page - 1) * size,
		Limit:     size,
	})
	if err != nil {
		return nil, 0, 0, 0, errcode.ErrInternal.WithCause(err)
	}

	items := make([]dto.AdminUserResp, 0, len(users))
	for _, user := range users {
		items = append(items, toAdminUserResp(user))
	}
	return items, total, page, size, nil
}
