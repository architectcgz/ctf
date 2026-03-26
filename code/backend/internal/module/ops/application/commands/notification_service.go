package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	opsports "ctf-platform/internal/module/ops/ports"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	platformevents "ctf-platform/internal/platform/events"
	"ctf-platform/pkg/errcode"
	ctfws "ctf-platform/pkg/websocket"
)

type NotificationService struct {
	repo       opsports.NotificationRepository
	pagination config.PaginationConfig
	manager    opsports.NotificationBroadcaster
	logger     *zap.Logger
}

func NewNotificationService(repo opsports.NotificationRepository, pagination config.PaginationConfig, manager opsports.NotificationBroadcaster, logger *zap.Logger) *NotificationService {
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

func (s *NotificationService) RegisterPracticeEventConsumers(bus platformevents.Bus) {
	if s == nil || bus == nil {
		return
	}
	bus.Subscribe(practicecontracts.EventFlagAccepted, s.handlePracticeFlagAccepted)
	bus.Subscribe(practicecontracts.EventHintUnlocked, s.handlePracticeHintUnlocked)
}

func (s *NotificationService) handlePracticeFlagAccepted(ctx context.Context, evt platformevents.Event) error {
	payload, ok := evt.Payload.(practicecontracts.FlagAcceptedEvent)
	if !ok {
		return fmt.Errorf("unexpected practice flag event payload: %T", evt.Payload)
	}
	link := fmt.Sprintf("/challenges/%d", payload.ChallengeID)
	return s.SendNotification(ctx, payload.UserID, &dto.NotificationReq{
		Type:    "challenge",
		Title:   "题目解出",
		Content: fmt.Sprintf("你已成功提交题目 #%d 的 Flag，获得 %d 分。", payload.ChallengeID, payload.Points),
		Link:    &link,
	})
}

func (s *NotificationService) handlePracticeHintUnlocked(ctx context.Context, evt platformevents.Event) error {
	payload, ok := evt.Payload.(practicecontracts.HintUnlockedEvent)
	if !ok {
		return fmt.Errorf("unexpected practice hint event payload: %T", evt.Payload)
	}
	link := fmt.Sprintf("/challenges/%d", payload.ChallengeID)
	return s.SendNotification(ctx, payload.UserID, &dto.NotificationReq{
		Type:    "challenge",
		Title:   "提示已解锁",
		Content: fmt.Sprintf("你已解锁题目 #%d 的第 %d 级提示。", payload.ChallengeID, payload.HintLevel),
		Link:    &link,
	})
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
