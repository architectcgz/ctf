package system

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
	ctfws "ctf-platform/pkg/websocket"
)

type NotificationBroadcaster interface {
	SendToUser(userID int64, message ctfws.Envelope) int
	Broadcast(message ctfws.Envelope) int
}

type NotificationService struct {
	repo       *NotificationRepository
	pagination config.PaginationConfig
	manager    NotificationBroadcaster
	logger     *zap.Logger
}

func NewNotificationService(repo *NotificationRepository, pagination config.PaginationConfig, manager NotificationBroadcaster, logger *zap.Logger) *NotificationService {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &NotificationService{
		repo:       repo,
		pagination: pagination,
		manager:    manager,
		logger:     logger,
	}
}

func (s *NotificationService) SendNotification(ctx context.Context, userID int64, req *dto.NotificationReq) error {
	notification := &model.Notification{
		UserID:  userID,
		Type:    req.Type,
		Title:   req.Title,
		Content: req.Content,
		Link:    req.Link,
	}
	if err := s.repo.Create(ctx, notification); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}

	if s.manager != nil {
		s.manager.SendToUser(userID, ctfws.Envelope{
			Type:      "notification.created",
			Payload:   toNotificationInfo(notification),
			Timestamp: time.Now().UTC(),
		})
	}
	return nil
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

	items, total, err := s.repo.List(ctx, NotificationListFilter{
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

func (s *NotificationService) MarkAsRead(ctx context.Context, userID, notificationID int64) error {
	notification, err := s.repo.FindByID(ctx, notificationID, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.ErrNotificationNotFound
	}
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if notification.IsRead {
		return nil
	}

	readAt := time.Now().UTC()
	if err := s.repo.MarkAsRead(ctx, notificationID, userID, readAt); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	notification.IsRead = true
	notification.ReadAt = &readAt

	if s.manager != nil {
		s.manager.SendToUser(userID, ctfws.Envelope{
			Type:      "notification.read",
			Payload:   toNotificationInfo(notification),
			Timestamp: time.Now().UTC(),
		})
	}
	return nil
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
