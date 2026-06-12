package commands

import (
	"context"
	"crypto/sha256"
	opscontracts "ctf-platform/internal/module/ops/contracts"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	opsentity "ctf-platform/internal/module/ops/entity"
	opsports "ctf-platform/internal/module/ops/ports"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	platformevents "ctf-platform/internal/platform/events"
	ctfws "ctf-platform/internal/websocket"
)

type NotificationService struct {
	repo       notificationCommandRepository
	pagination config.PaginationConfig
	manager    opsports.NotificationBroadcaster
	logger     *zap.Logger
}

type notificationCommandRepository interface {
	opsports.NotificationCommandRepository
	opsports.NotificationOutboxTxManager
}

func NewNotificationService(repo notificationCommandRepository, pagination config.PaginationConfig, manager opsports.NotificationBroadcaster, logger *zap.Logger) *NotificationService {
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

func (s *NotificationService) HandlePracticeFlagAcceptedOutboxEvent(ctx context.Context, event platformevents.OutboxEvent) error {
	decoded, err := notificationOutboxCodec().Decode(event)
	if err != nil {
		return err
	}
	payload, ok := decoded.Payload.(*practicecontracts.FlagAcceptedEvent)
	if !ok {
		return fmt.Errorf("unexpected practice flag event payload: %T", decoded.Payload)
	}
	link := fmt.Sprintf("/challenges/%d", payload.ChallengeID)
	return s.sendNotificationFromOutboxEvent(ctx, event, "practice_flag_accepted_notification", payload.UserID, SendNotificationInput{
		Type:    "challenge",
		Title:   "题目解出",
		Content: fmt.Sprintf("你已成功提交题目 #%d 的 Flag，获得 %d 分。", payload.ChallengeID, payload.Points),
		Link:    &link,
	})
}

func (s *NotificationService) HandleChallengePublishCheckFinishedOutboxEvent(ctx context.Context, event platformevents.OutboxEvent) error {
	decoded, err := notificationOutboxCodec().Decode(event)
	if err != nil {
		return err
	}
	payload, ok := decoded.Payload.(*challengecontracts.PublishCheckFinishedEvent)
	if !ok {
		return fmt.Errorf("unexpected challenge publish check event payload: %T", decoded.Payload)
	}

	title := "题目发布失败"
	content := fmt.Sprintf("《%s》未通过平台自检。", payload.ChallengeTitle)
	if payload.Passed {
		title = "题目发布成功"
		content = fmt.Sprintf("《%s》已通过平台自检并自动发布。", payload.ChallengeTitle)
	} else if payload.FailureSummary != "" {
		content = fmt.Sprintf("《%s》未通过平台自检：%s", payload.ChallengeTitle, payload.FailureSummary)
	}

	link := fmt.Sprintf("/admin/challenges/%d", payload.ChallengeID)
	return s.sendNotificationFromOutboxEvent(ctx, event, "challenge_publish_check_notification", payload.UserID, SendNotificationInput{
		Type:    opsentity.NotificationTypeChallenge,
		Title:   title,
		Content: content,
		Link:    &link,
	})
}

func (s *NotificationService) HandleNotificationFanoutEvent(ctx context.Context, event platformevents.OutboxEvent) error {
	decoded, err := notificationOutboxCodec().Decode(event)
	if err != nil {
		return err
	}
	payload, ok := decoded.Payload.(*opscontracts.NotificationFanoutEvent)
	if !ok {
		return fmt.Errorf("unexpected notification fanout payload: %T", decoded.Payload)
	}
	if s == nil || s.manager == nil || payload.UserID <= 0 {
		return nil
	}
	timestamp := payload.OccurredAt
	if timestamp.IsZero() {
		timestamp = decoded.Event.OccurredAt
	}
	s.manager.SendToUser(payload.UserID, ctfws.Envelope{
		Type:      event.Name,
		Payload:   payload.Notification,
		Timestamp: timestamp.UTC(),
	})
	return nil
}

func (s *NotificationService) SendNotification(ctx context.Context, userID int64, req SendNotificationInput) error {
	notification := &opsentity.Notification{
		UserID:  userID,
		Type:    req.Type,
		Title:   req.Title,
		Content: req.Content,
		Link:    req.Link,
	}
	if err := s.repo.WithinNotificationOutboxTx(ctx, func(txRepo opsports.NotificationOutboxTxRepository) error {
		if err := txRepo.Create(ctx, notification); err != nil {
			return err
		}
		return enqueueNotificationFanoutOutboxEvent(ctx, txRepo, opscontracts.EventNotificationCreated, notification, time.Now().UTC())
	}); err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	return nil
}

func (s *NotificationService) sendNotificationFromOutboxEvent(ctx context.Context, source platformevents.OutboxEvent, handlerName string, userID int64, req SendNotificationInput) error {
	notification := &opsentity.Notification{
		UserID:         userID,
		Type:           req.Type,
		Title:          req.Title,
		Content:        req.Content,
		Link:           req.Link,
		SourceEventKey: notificationSourceEventKey(source, handlerName),
	}
	if err := s.repo.WithinNotificationOutboxTx(ctx, func(txRepo opsports.NotificationOutboxTxRepository) error {
		created, err := txRepo.CreateIfSourceEventAbsent(ctx, notification)
		if err != nil || !created {
			return err
		}
		return enqueueNotificationFanoutOutboxEvent(ctx, txRepo, opscontracts.EventNotificationCreated, notification, time.Now().UTC())
	}); err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	return nil
}

func (s *NotificationService) PublishAdminNotification(ctx context.Context, actorUserID int64, req PublishAdminNotificationInput) (*AdminNotificationPublishResp, error) {
	if req.AudienceRules.Mode != "union" || len(req.AudienceRules.Rules) == 0 {
		return nil, apperror.ErrInvalidParams
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
		return nil, apperror.ErrInternal.WithCause(err)
	}
	batch := &opsentity.NotificationBatch{
		Type:           req.Type,
		Title:          req.Title,
		Content:        req.Content,
		Link:           req.Link,
		AudienceMode:   req.AudienceRules.Mode,
		AudienceRules:  string(audienceRules),
		RecipientCount: len(recipientIDs),
		CreatedBy:      actorUserID,
	}

	notifications := make([]*opsentity.Notification, 0, len(recipientIDs))
	for _, userID := range recipientIDs {
		notifications = append(notifications, &opsentity.Notification{
			UserID:  userID,
			Type:    req.Type,
			Title:   req.Title,
			Content: req.Content,
			Link:    req.Link,
		})
	}
	if err := s.repo.WithinNotificationOutboxTx(ctx, func(txRepo opsports.NotificationOutboxTxRepository) error {
		if err := txRepo.CreateBatch(ctx, batch, notifications); err != nil {
			return err
		}
		now := time.Now().UTC()
		for _, item := range notifications {
			if err := enqueueNotificationFanoutOutboxEvent(ctx, txRepo, opscontracts.EventNotificationCreated, item, now); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	return notificationMapper.ToAdminNotificationPublishRespPtr(adminNotificationPublishRespSource{
		BatchID:        batch.ID,
		RecipientCount: len(notifications),
	}), nil
}

func (s *NotificationService) MarkAsRead(ctx context.Context, userID, notificationID int64) error {
	notification, err := s.repo.FindByID(ctx, notificationID, userID)
	if errors.Is(err, opsports.ErrNotificationNotFound) {
		return opscontracts.ErrNotificationNotFound
	}
	if err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	if notification.IsRead {
		return nil
	}

	readAt := time.Now().UTC()
	if err := s.repo.WithinNotificationOutboxTx(ctx, func(txRepo opsports.NotificationOutboxTxRepository) error {
		if err := txRepo.MarkAsRead(ctx, notificationID, userID, readAt); err != nil {
			return err
		}
		notification.IsRead = true
		notification.ReadAt = &readAt
		return enqueueNotificationFanoutOutboxEvent(ctx, txRepo, opscontracts.EventNotificationRead, notification, readAt)
	}); err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	return nil
}

func toNotificationInfo(notification *opsentity.Notification) NotificationInfo {
	resp := notificationMapper.ToNotificationInfoPtr(notification)
	resp.Content = normalizeOptionalString(notification.Content)
	resp.Unread = !notification.IsRead
	return *resp
}

func enqueueNotificationFanoutOutboxEvent(ctx context.Context, repo opsports.NotificationOutboxTxRepository, eventName string, notification *opsentity.Notification, occurredAt time.Time) error {
	if repo == nil {
		return fmt.Errorf("notification outbox repository is not configured")
	}
	if notification == nil {
		return fmt.Errorf("notification is required")
	}
	occurredAt = occurredAt.UTC()
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	payload := opscontracts.NotificationFanoutEvent{
		UserID:       notification.UserID,
		Notification: toNotificationInfo(notification),
		OccurredAt:   occurredAt,
	}
	event, err := notificationOutboxCodec().Encode(eventName, opscontracts.EventNotificationPayloadVersion, payload, occurredAt)
	if err != nil {
		return err
	}
	event.Route = platformevents.OutboxRouteStream
	event.DedupeKey = notificationFanoutDedupeKey(eventName, notification.ID)
	return repo.EnqueueOutboxEvent(ctx, event)
}

func notificationFanoutDedupeKey(eventName string, notificationID int64) string {
	switch eventName {
	case opscontracts.EventNotificationCreated:
		return fmt.Sprintf("notification:created:%d", notificationID)
	case opscontracts.EventNotificationRead:
		return fmt.Sprintf("notification:read:%d", notificationID)
	default:
		return fmt.Sprintf("notification:%s:%d", eventName, notificationID)
	}
}

func notificationSourceEventKey(event platformevents.OutboxEvent, handlerName string) string {
	handlerName = strings.TrimSpace(handlerName)
	if handlerName == "" {
		handlerName = "notification"
	}
	if event.DedupeKey != "" {
		return fmt.Sprintf("outbox:%s:%s", event.DedupeKey, handlerName)
	}
	sum := sha256.Sum256(event.Payload)
	return fmt.Sprintf("outbox:%s:v%d:%s:%s", event.Name, event.PayloadVersion, hex.EncodeToString(sum[:]), handlerName)
}

func notificationOutboxCodec() *platformevents.OutboxCodec {
	codec := platformevents.NewOutboxCodec()
	codec.Register(practicecontracts.EventFlagAccepted, practicecontracts.EventFlagAcceptedPayloadVersion, func() any {
		return &practicecontracts.FlagAcceptedEvent{}
	})
	codec.Register(challengecontracts.EventPublishCheckFinished, challengecontracts.EventPublishCheckFinishedPayloadVersion, func() any {
		return &challengecontracts.PublishCheckFinishedEvent{}
	})
	codec.Register(opscontracts.EventNotificationCreated, opscontracts.EventNotificationPayloadVersion, func() any {
		return &opscontracts.NotificationFanoutEvent{}
	})
	codec.Register(opscontracts.EventNotificationRead, opscontracts.EventNotificationPayloadVersion, func() any {
		return &opscontracts.NotificationFanoutEvent{}
	})
	return codec
}

func (s *NotificationService) resolveAudienceRule(ctx context.Context, rule NotificationAudienceRuleInput) ([]int64, error) {
	switch rule.Type {
	case NotificationAudienceTypeAll:
		userIDs, err := s.repo.ListAllUserIDs(ctx)
		if err != nil {
			return nil, apperror.ErrInternal.WithCause(err)
		}
		return userIDs, nil
	case NotificationAudienceTypeRole:
		roles, err := normalizeRoleSlice(rule.Values)
		if err != nil {
			return nil, apperror.ErrInvalidParams
		}
		if len(roles) == 0 {
			return nil, apperror.ErrInvalidParams
		}
		userIDs, err := s.repo.ListUserIDsByRoles(ctx, roles)
		if err != nil {
			return nil, apperror.ErrInternal.WithCause(err)
		}
		return userIDs, nil
	case NotificationAudienceTypeClass:
		classNames := normalizeStringSlice(rule.Values)
		if len(classNames) == 0 {
			return nil, apperror.ErrInvalidParams
		}
		userIDs, err := s.repo.ListUserIDsByClasses(ctx, classNames)
		if err != nil {
			return nil, apperror.ErrInternal.WithCause(err)
		}
		return userIDs, nil
	case NotificationAudienceTypeUser:
		userIDs, err := normalizeUserIDSlice(rule.Values)
		if err != nil {
			return nil, apperror.ErrInvalidParams
		}
		if len(userIDs) == 0 {
			return nil, apperror.ErrInvalidParams
		}
		resolvedIDs, err := s.repo.ListExistingUserIDs(ctx, userIDs)
		if err != nil {
			return nil, apperror.ErrInternal.WithCause(err)
		}
		return resolvedIDs, nil
	default:
		return nil, apperror.ErrInvalidParams
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
			return nil, apperror.ErrInvalidParams
		}
		parsed = append(parsed, userID)
	}
	return normalizeInt64Slice(parsed), nil
}

func normalizeRoleSlice(values []string) ([]string, error) {
	roles := normalizeStringSlice(values)
	allowed := map[string]struct{}{
		identitycontracts.RoleStudent: {},
		identitycontracts.RoleTeacher: {},
		identitycontracts.RoleAdmin:   {},
	}
	for _, role := range roles {
		if _, ok := allowed[role]; !ok {
			return nil, apperror.ErrInvalidParams
		}
	}
	return roles, nil
}
