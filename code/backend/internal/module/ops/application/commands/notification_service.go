package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	opsports "ctf-platform/internal/module/ops/ports"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	platformevents "ctf-platform/internal/platform/events"
	commonmapper "ctf-platform/internal/shared/mapperhelper"
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
}

func (s *NotificationService) handlePracticeFlagAccepted(ctx context.Context, evt platformevents.Event) error {
	payload, ok := evt.Payload.(practicecontracts.FlagAcceptedEvent)
	if !ok {
		return fmt.Errorf("unexpected practice flag event payload: %T", evt.Payload)
	}
	link := fmt.Sprintf("/challenges/%d", payload.ChallengeID)
	return s.SendNotification(ctx, payload.UserID, SendNotificationInput{
		Type:    "challenge",
		Title:   "题目解出",
		Content: fmt.Sprintf("你已成功提交题目 #%d 的 Flag，获得 %d 分。", payload.ChallengeID, payload.Points),
		Link:    &link,
	})
}

func (s *NotificationService) SendNotification(ctx context.Context, userID int64, req SendNotificationInput) error {
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

func (s *NotificationService) PublishAdminNotification(ctx context.Context, actorUserID int64, req PublishAdminNotificationInput) (*dto.AdminNotificationPublishResp, error) {
	if req.AudienceRules.Mode != "union" || len(req.AudienceRules.Rules) == 0 {
		return nil, errcode.ErrInvalidParams
	}

	recipientSet := make(map[int64]struct{})
	for _, rule := range req.AudienceRules.Rules {
		userIDs, err := s.resolveAudienceRule(ctx, rule)
		if err != nil {
			return nil, err
		}
		for _, userID := range userIDs {
			recipientSet[userID] = struct{}{}
		}
	}

	recipientIDs := make([]int64, 0, len(recipientSet))
	for userID := range recipientSet {
		recipientIDs = append(recipientIDs, userID)
	}
	sort.Slice(recipientIDs, func(i, j int) bool { return recipientIDs[i] < recipientIDs[j] })

	audienceRules, err := json.Marshal(req.AudienceRules)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	batch := &model.NotificationBatch{
		Type:           req.Type,
		Title:          req.Title,
		Content:        req.Content,
		Link:           req.Link,
		AudienceMode:   req.AudienceRules.Mode,
		AudienceRules:  string(audienceRules),
		RecipientCount: len(recipientIDs),
		CreatedBy:      actorUserID,
	}

	notifications := make([]*model.Notification, 0, len(recipientIDs))
	for _, userID := range recipientIDs {
		notifications = append(notifications, &model.Notification{
			UserID:  userID,
			Type:    req.Type,
			Title:   req.Title,
			Content: req.Content,
			Link:    req.Link,
		})
	}
	if err := s.repo.CreateBatch(ctx, batch, notifications); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	if s.manager != nil {
		for _, item := range notifications {
			s.manager.SendToUser(item.UserID, ctfws.Envelope{
				Type:      "notification.created",
				Payload:   toNotificationInfo(item),
				Timestamp: time.Now().UTC(),
			})
		}
	}

	return &dto.AdminNotificationPublishResp{
		BatchID:        batch.ID,
		RecipientCount: len(notifications),
	}, nil
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
	mapped := notificationMapper.ToNotificationInfo(*notification)
	mapped.Content = commonmapper.NormalizeOptionalString(notification.Content)
	mapped.Unread = !notification.IsRead
	return mapped
}

func (s *NotificationService) resolveAudienceRule(ctx context.Context, rule NotificationAudienceRuleInput) ([]int64, error) {
	switch rule.Type {
	case dto.NotificationAudienceTypeAll:
		userIDs, err := s.repo.ListAllUserIDs(ctx)
		if err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		return userIDs, nil
	case dto.NotificationAudienceTypeRole:
		roles, err := normalizeRoleSlice(rule.Values)
		if err != nil {
			return nil, errcode.ErrInvalidParams
		}
		if len(roles) == 0 {
			return nil, errcode.ErrInvalidParams
		}
		userIDs, err := s.repo.ListUserIDsByRoles(ctx, roles)
		if err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		return userIDs, nil
	case dto.NotificationAudienceTypeClass:
		classNames := normalizeStringSlice(rule.Values)
		if len(classNames) == 0 {
			return nil, errcode.ErrInvalidParams
		}
		userIDs, err := s.repo.ListUserIDsByClasses(ctx, classNames)
		if err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		return userIDs, nil
	case dto.NotificationAudienceTypeUser:
		userIDs, err := normalizeUserIDSlice(rule.Values)
		if err != nil {
			return nil, errcode.ErrInvalidParams
		}
		if len(userIDs) == 0 {
			return nil, errcode.ErrInvalidParams
		}
		resolvedIDs, err := s.repo.ListExistingUserIDs(ctx, userIDs)
		if err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		return resolvedIDs, nil
	default:
		return nil, errcode.ErrInvalidParams
	}
}

func normalizeStringSlice(values []string) []string {
	set := make(map[string]struct{}, len(values))
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		set[value] = struct{}{}
	}
	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func normalizeInt64Slice(values []int64) []int64 {
	set := make(map[int64]struct{}, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		set[value] = struct{}{}
	}
	result := make([]int64, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func normalizeUserIDSlice(values []string) ([]int64, error) {
	parsed := make([]int64, 0, len(values))
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		userID, err := strconv.ParseInt(value, 10, 64)
		if err != nil || userID <= 0 {
			return nil, errcode.ErrInvalidParams
		}
		parsed = append(parsed, userID)
	}
	return normalizeInt64Slice(parsed), nil
}

func normalizeRoleSlice(values []string) ([]string, error) {
	roles := normalizeStringSlice(values)
	allowed := map[string]struct{}{
		model.RoleStudent: {},
		model.RoleTeacher: {},
		model.RoleAdmin:   {},
	}
	for _, role := range roles {
		if _, ok := allowed[role]; !ok {
			return nil, errcode.ErrInvalidParams
		}
	}
	return roles, nil
}
