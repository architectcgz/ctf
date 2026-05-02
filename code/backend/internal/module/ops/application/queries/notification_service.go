package queries

import (
	"context"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	opsports "ctf-platform/internal/module/ops/ports"
	"ctf-platform/pkg/errcode"
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
	var content *string
	if notification.Content != "" {
		content = &notification.Content
	}

	return dto.NotificationInfo{
		ID:        notification.ID,
		Type:      notification.Type,
		Title:     notification.Title,
		Content:   content,
		Unread:    !notification.IsRead,
		Link:      notification.Link,
		CreatedAt: notification.CreatedAt,
		ReadAt:    notification.ReadAt,
	}
}
