package queries

import (
	"context"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	opsports "ctf-platform/internal/module/ops/ports"
	commonmapper "ctf-platform/internal/shared/mapperhelper"
	"ctf-platform/pkg/errcode"
)

type NotificationService struct {
	repo       opsports.NotificationRepository
	pagination config.PaginationConfig
	logger     *zap.Logger
}

func NewNotificationService(repo opsports.NotificationRepository, pagination config.PaginationConfig, logger *zap.Logger) *NotificationService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &NotificationService{
		repo:       repo,
		pagination: pagination,
		logger:     logger,
	}
}

func (s *NotificationService) GetNotifications(ctx context.Context, userID int64, query *dto.NotificationQuery) ([]dto.NotificationInfo, int64, int, int, error) {
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
		return nil, 0, 0, 0, errcode.ErrInternal.WithCause(err)
	}

	result := make([]dto.NotificationInfo, 0, len(items))
	for _, item := range items {
		result = append(result, toNotificationInfo(&item))
	}

	return result, total, page, pageSize, nil
}

func toNotificationInfo(notification *model.Notification) dto.NotificationInfo {
	resp := notificationMapper.ToNotificationInfoPtr(notification)
	resp.Content = commonmapper.NormalizeOptionalString(notification.Content)
	resp.Unread = !notification.IsRead
	return *resp
}
