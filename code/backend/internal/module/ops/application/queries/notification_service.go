package queries

import (
	"context"

	"go.uber.org/zap"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	opsentity "ctf-platform/internal/module/ops/entity"
	opsports "ctf-platform/internal/module/ops/ports"
	commonmapper "ctf-platform/internal/shared/mapperhelper"
)

type NotificationService struct {
	repo       opsports.NotificationQueryRepository
	pagination config.PaginationConfig
	logger     *zap.Logger
}

func NewNotificationService(repo opsports.NotificationQueryRepository, pagination config.PaginationConfig, logger *zap.Logger) *NotificationService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &NotificationService{
		repo:       repo,
		pagination: pagination,
		logger:     logger,
	}
}

func (s *NotificationService) GetNotifications(ctx context.Context, userID int64, query *NotificationQuery) ([]NotificationInfo, int64, int, int, error) {
	page := query.Page
	if page < 1 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize < 1 {
		pageSize = s.pagination.DefaultPageSize
	}
	if pageSize > s.pagination.MaxPageSize {
		pageSize = s.pagination.MaxPageSize
	}

	items, total, err := s.repo.List(ctx, opsports.NotificationListFilter{
		UserID: userID,
		Type:   query.Type,
		Offset: (page - 1) * pageSize,
		Limit:  pageSize,
	})
	if err != nil {
		return nil, 0, 0, 0, apperror.ErrInternal.WithCause(err)
	}

	result := make([]NotificationInfo, 0, len(items))
	for _, item := range items {
		result = append(result, toNotificationInfo(&item))
	}

	return result, total, page, pageSize, nil
}

func toNotificationInfo(notification *opsentity.Notification) NotificationInfo {
	resp := notificationMapper.ToNotificationInfoPtr(notification)
	resp.Content = commonmapper.NormalizeOptionalString(notification.Content)
	resp.Unread = !notification.IsRead
	return *resp
}
