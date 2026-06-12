package commands

import (
	"context"
	opscontracts "ctf-platform/internal/module/ops/contracts"
	"errors"
	"testing"
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

type stubNotificationRepository struct {
	created                   []*opsentity.Notification
	createdBatch              *opsentity.NotificationBatch
	createdBatchNotifications []*opsentity.Notification
	enqueued                  []platformevents.OutboxEvent
	sourceEventKeys           map[string]struct{}
	nextNotificationID        int64
	findByIDFn                func(ctx context.Context, notificationID, userID int64) (*opsentity.Notification, error)
	listAllUserIDsFn          func(ctx context.Context) ([]int64, error)
	listUserIDsByRolesFn      func(ctx context.Context, roles []string) ([]int64, error)
	listUserIDsByClassesFn    func(ctx context.Context, classNames []string) ([]int64, error)
	listExistingUserIDsFn     func(ctx context.Context, userIDs []int64) ([]int64, error)
	markAsReadCalls           int
}

func (r *stubNotificationRepository) Create(_ context.Context, notification *opsentity.Notification) error {
	if notification.ID == 0 {
		if r.nextNotificationID == 0 {
			r.nextNotificationID = 101
		}
		notification.ID = r.nextNotificationID
		r.nextNotificationID++
	}
	copied := *notification
	r.created = append(r.created, &copied)
	return nil
}

func (r *stubNotificationRepository) CreateIfSourceEventAbsent(ctx context.Context, notification *opsentity.Notification) (bool, error) {
	if notification != nil && notification.SourceEventKey != "" {
		if r.sourceEventKeys == nil {
			r.sourceEventKeys = make(map[string]struct{})
		}
		if _, exists := r.sourceEventKeys[notification.SourceEventKey]; exists {
			return false, nil
		}
		r.sourceEventKeys[notification.SourceEventKey] = struct{}{}
	}
	return true, r.Create(ctx, notification)
}

func (r *stubNotificationRepository) List(_ context.Context, _ opsports.NotificationListFilter) ([]opsentity.Notification, int64, error) {
	return nil, 0, nil
}

func (r *stubNotificationRepository) FindByID(ctx context.Context, notificationID, userID int64) (*opsentity.Notification, error) {
	if r.findByIDFn == nil {
		return nil, errors.New("unexpected FindByID call")
	}
	return r.findByIDFn(ctx, notificationID, userID)
}

func (r *stubNotificationRepository) MarkAsRead(_ context.Context, _ int64, _ int64, _ any) error {
	r.markAsReadCalls++
	return nil
}

func (r *stubNotificationRepository) CreateBatch(_ context.Context, batch *opsentity.NotificationBatch, notifications []*opsentity.Notification) error {
	copiedBatch := *batch
	if copiedBatch.ID == 0 {
		copiedBatch.ID = 88
		batch.ID = 88
	}
	r.createdBatch = &copiedBatch
	r.createdBatchNotifications = make([]*opsentity.Notification, 0, len(notifications))
	for _, item := range notifications {
		if item.ID == 0 {
			if r.nextNotificationID == 0 {
				r.nextNotificationID = 201
			}
			item.ID = r.nextNotificationID
			r.nextNotificationID++
		}
		item.BatchID = &copiedBatch.ID
		copied := *item
		r.createdBatchNotifications = append(r.createdBatchNotifications, &copied)
	}
	return nil
}

func (r *stubNotificationRepository) WithinNotificationOutboxTx(ctx context.Context, fn func(txRepo opsports.NotificationOutboxTxRepository) error) error {
	return fn(r)
}

func (r *stubNotificationRepository) EnqueueOutboxEvent(_ context.Context, event platformevents.OutboxEvent) error {
	r.enqueued = append(r.enqueued, event)
	return nil
}

func (r *stubNotificationRepository) ListAllUserIDs(ctx context.Context) ([]int64, error) {
	if r.listAllUserIDsFn == nil {
		return nil, errors.New("unexpected ListAllUserIDs call")
	}
	return r.listAllUserIDsFn(ctx)
}

func (r *stubNotificationRepository) ListUserIDsByRoles(ctx context.Context, roles []string) ([]int64, error) {
	if r.listUserIDsByRolesFn == nil {
		return nil, errors.New("unexpected ListUserIDsByRoles call")
	}
	return r.listUserIDsByRolesFn(ctx, roles)
}

func (r *stubNotificationRepository) ListUserIDsByClasses(ctx context.Context, classNames []string) ([]int64, error) {
	if r.listUserIDsByClassesFn == nil {
		return nil, errors.New("unexpected ListUserIDsByClasses call")
	}
	return r.listUserIDsByClassesFn(ctx, classNames)
}

func (r *stubNotificationRepository) ListExistingUserIDs(ctx context.Context, userIDs []int64) ([]int64, error) {
	if r.listExistingUserIDsFn == nil {
		return nil, errors.New("unexpected ListExistingUserIDs call")
	}
	return r.listExistingUserIDsFn(ctx, userIDs)
}

type stubNotificationBroadcaster struct {
	sentUsers []int64
	envelopes []ctfws.Envelope
}

func (b *stubNotificationBroadcaster) SendToUser(userID int64, message ctfws.Envelope) int {
	b.sentUsers = append(b.sentUsers, userID)
	b.envelopes = append(b.envelopes, message)
	return 1
}

func (b *stubNotificationBroadcaster) Broadcast(message ctfws.Envelope) int {
	b.envelopes = append(b.envelopes, message)
	return 0
}

type recordingBus struct {
	subscribers map[string][]platformevents.Handler
}

func (b *recordingBus) Subscribe(name string, fn platformevents.Handler) {
	if b.subscribers == nil {
		b.subscribers = make(map[string][]platformevents.Handler)
	}
	b.subscribers[name] = append(b.subscribers[name], fn)
}

func (b *recordingBus) Publish(ctx context.Context, evt platformevents.Event) error {
	for _, handler := range b.subscribers[evt.Name] {
		if err := handler(ctx, evt); err != nil {
			return err
		}
	}
	return nil
}

type notificationFanoutTestPayload struct {
	UserID       int64            `json:"user_id"`
	Notification NotificationInfo `json:"notification"`
	OccurredAt   time.Time        `json:"occurred_at"`
}

func TestNotificationServiceHandlePracticeFlagAcceptedOutboxEventCreatesNotification(t *testing.T) {
	repo := &stubNotificationRepository{}
	service := NewNotificationService(repo, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, nil, zap.NewNop())
	codec := platformevents.NewOutboxCodec()
	event, err := codec.Encode(practicecontracts.EventFlagAccepted, practicecontracts.EventFlagAcceptedPayloadVersion, practicecontracts.FlagAcceptedEvent{
		UserID:      7,
		ChallengeID: 12,
		Points:      30,
		OccurredAt:  time.Date(2026, 6, 12, 10, 0, 0, 0, time.UTC),
	}, time.Date(2026, 6, 12, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("encode flag accepted event: %v", err)
	}

	if err := service.HandlePracticeFlagAcceptedOutboxEvent(context.Background(), event); err != nil {
		t.Fatalf("HandlePracticeFlagAcceptedOutboxEvent() error = %v", err)
	}

	if len(repo.created) != 1 {
		t.Fatalf("created notifications len = %d, want 1", len(repo.created))
	}
	if repo.created[0].Title != "题目解出" {
		t.Fatalf("unexpected created notifications = %+v", repo.created)
	}
	if repo.created[0].Link == nil || *repo.created[0].Link != "/challenges/12" {
		t.Fatalf("unexpected created notification link = %+v", repo.created[0].Link)
	}
	if len(repo.enqueued) != 1 || repo.enqueued[0].Name != "notification.created" {
		t.Fatalf("unexpected notification fanout outbox events: %+v", repo.enqueued)
	}
}

func TestNotificationServiceHandlePracticeFlagAcceptedOutboxEventIsIdempotent(t *testing.T) {
	repo := &stubNotificationRepository{}
	service := NewNotificationService(repo, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, nil, zap.NewNop())
	codec := platformevents.NewOutboxCodec()
	event, err := codec.Encode(practicecontracts.EventFlagAccepted, practicecontracts.EventFlagAcceptedPayloadVersion, practicecontracts.FlagAcceptedEvent{
		UserID:      7,
		ChallengeID: 12,
		Points:      30,
		OccurredAt:  time.Date(2026, 6, 12, 10, 0, 0, 0, time.UTC),
	}, time.Date(2026, 6, 12, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("encode flag accepted event: %v", err)
	}
	event.DedupeKey = "practice:flag_accepted:7:12"

	if err := service.HandlePracticeFlagAcceptedOutboxEvent(context.Background(), event); err != nil {
		t.Fatalf("HandlePracticeFlagAcceptedOutboxEvent(first) error = %v", err)
	}
	if err := service.HandlePracticeFlagAcceptedOutboxEvent(context.Background(), event); err != nil {
		t.Fatalf("HandlePracticeFlagAcceptedOutboxEvent(second) error = %v", err)
	}
	if len(repo.created) != 1 {
		t.Fatalf("created notifications len = %d, want 1", len(repo.created))
	}
	if len(repo.enqueued) != 1 {
		t.Fatalf("enqueued fanout events len = %d, want 1", len(repo.enqueued))
	}
	if repo.created[0].SourceEventKey == "" {
		t.Fatal("expected notification created from outbox handler to carry source event key")
	}
}

func TestNotificationServiceHandleChallengePublishCheckFinishedOutboxEventCreatesNotification(t *testing.T) {
	repo := &stubNotificationRepository{}
	service := NewNotificationService(repo, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, nil, zap.NewNop())
	codec := platformevents.NewOutboxCodec()
	event, err := codec.Encode(challengecontracts.EventPublishCheckFinished, challengecontracts.EventPublishCheckFinishedPayloadVersion, challengecontracts.PublishCheckFinishedEvent{
		UserID:         9,
		ChallengeID:    21,
		ChallengeTitle: "Web 101",
		Passed:         false,
		FailureSummary: "镜像缺失",
		OccurredAt:     time.Date(2026, 6, 12, 10, 0, 0, 0, time.UTC),
	}, time.Date(2026, 6, 12, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("encode publish check event: %v", err)
	}

	if err := service.HandleChallengePublishCheckFinishedOutboxEvent(context.Background(), event); err != nil {
		t.Fatalf("HandleChallengePublishCheckFinishedOutboxEvent() error = %v", err)
	}

	if len(repo.created) != 1 {
		t.Fatalf("created notifications len = %d, want 1", len(repo.created))
	}
	if repo.created[0].Title != "题目发布失败" {
		t.Fatalf("unexpected created notifications = %+v", repo.created)
	}
	if repo.created[0].Content != "《Web 101》未通过平台自检：镜像缺失" {
		t.Fatalf("unexpected created notification content = %q", repo.created[0].Content)
	}
	if repo.created[0].Link == nil || *repo.created[0].Link != "/admin/challenges/21" {
		t.Fatalf("unexpected created notification link = %+v", repo.created[0].Link)
	}
	if len(repo.enqueued) != 1 || repo.enqueued[0].Name != "notification.created" {
		t.Fatalf("unexpected notification fanout outbox events: %+v", repo.enqueued)
	}
}

func TestNotificationServiceMarkAsReadReturnsNotificationNotFound(t *testing.T) {
	service := NewNotificationService(&stubNotificationRepository{
		findByIDFn: func(_ context.Context, _, _ int64) (*opsentity.Notification, error) {
			return nil, opsports.ErrNotificationNotFound
		},
	}, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, nil, zap.NewNop())

	err := service.MarkAsRead(context.Background(), 7, 11)
	if !errors.Is(err, opscontracts.ErrNotificationNotFound) {
		t.Fatalf("MarkAsRead() error = %v, want ErrNotificationNotFound", err)
	}
}

func TestNotificationServiceMarkAsReadIsIdempotentForReadNotification(t *testing.T) {
	repo := &stubNotificationRepository{
		findByIDFn: func(_ context.Context, _, _ int64) (*opsentity.Notification, error) {
			readAt := time.Now().UTC()
			return &opsentity.Notification{
				ID:      11,
				UserID:  7,
				Title:   "already read",
				IsRead:  true,
				ReadAt:  &readAt,
				Type:    opsentity.NotificationTypeSystem,
				Content: "done",
			}, nil
		},
	}
	broadcaster := &stubNotificationBroadcaster{}
	service := NewNotificationService(repo, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, broadcaster, zap.NewNop())

	err := service.MarkAsRead(context.Background(), 7, 11)
	if err != nil {
		t.Fatalf("MarkAsRead() error = %v", err)
	}
	if repo.markAsReadCalls != 0 {
		t.Fatalf("MarkAsRead() repo calls = %d, want 0", repo.markAsReadCalls)
	}
	if len(broadcaster.envelopes) != 0 {
		t.Fatalf("MarkAsRead() should not publish websocket event for already-read notification, got %+v", broadcaster.envelopes)
	}
}

func TestNotificationServiceMarkAsReadEnqueuesReadOutboxEventWithoutDirectWebsocket(t *testing.T) {
	readAt := time.Now().UTC().Add(-time.Minute)
	repo := &stubNotificationRepository{
		findByIDFn: func(_ context.Context, notificationID, userID int64) (*opsentity.Notification, error) {
			return &opsentity.Notification{
				ID:        notificationID,
				UserID:    userID,
				Title:     "unread",
				Type:      opsentity.NotificationTypeSystem,
				Content:   "content",
				IsRead:    false,
				CreatedAt: readAt,
			}, nil
		},
	}
	broadcaster := &stubNotificationBroadcaster{}
	service := NewNotificationService(repo, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, broadcaster, zap.NewNop())

	err := service.MarkAsRead(context.Background(), 7, 11)
	if err != nil {
		t.Fatalf("MarkAsRead() error = %v", err)
	}
	if repo.markAsReadCalls != 1 {
		t.Fatalf("MarkAsRead() repo calls = %d, want 1", repo.markAsReadCalls)
	}
	if len(broadcaster.envelopes) != 0 {
		t.Fatalf("MarkAsRead should not publish websocket event directly, got %+v", broadcaster.envelopes)
	}
	if len(repo.enqueued) != 1 {
		t.Fatalf("enqueued outbox events len = %d, want 1", len(repo.enqueued))
	}
	event := repo.enqueued[0]
	if event.Name != "notification.read" {
		t.Fatalf("unexpected outbox event name: %q", event.Name)
	}
	if event.PayloadVersion != 1 || event.Route != platformevents.OutboxRouteStream {
		t.Fatalf("unexpected outbox envelope: %+v", event)
	}
	if event.DedupeKey != "notification:read:11" {
		t.Fatalf("unexpected outbox dedupe key: %s", event.DedupeKey)
	}
}

func TestNotificationServiceSendNotificationEnqueuesCreatedOutboxEventWithoutDirectWebsocket(t *testing.T) {
	repo := &stubNotificationRepository{}
	broadcaster := &stubNotificationBroadcaster{}
	service := NewNotificationService(repo, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, broadcaster, zap.NewNop())

	if err := service.SendNotification(context.Background(), 7, SendNotificationInput{
		Type:    opsentity.NotificationTypeSystem,
		Title:   "title",
		Content: "content",
	}); err != nil {
		t.Fatalf("SendNotification() error = %v", err)
	}
	if len(repo.created) != 1 {
		t.Fatalf("created notifications len = %d, want 1", len(repo.created))
	}
	if len(broadcaster.envelopes) != 0 {
		t.Fatalf("SendNotification should not publish websocket event directly, got %+v", broadcaster.envelopes)
	}
	if len(repo.enqueued) != 1 {
		t.Fatalf("enqueued outbox events len = %d, want 1", len(repo.enqueued))
	}
	event := repo.enqueued[0]
	if event.Name != "notification.created" {
		t.Fatalf("unexpected outbox event name: %q", event.Name)
	}
	if event.PayloadVersion != 1 || event.Route != platformevents.OutboxRouteStream {
		t.Fatalf("unexpected outbox envelope: %+v", event)
	}
	if event.DedupeKey != "notification:created:101" {
		t.Fatalf("unexpected outbox dedupe key: %s", event.DedupeKey)
	}

	codec := platformevents.NewOutboxCodec()
	codec.Register("notification.created", 1, func() any { return &notificationFanoutTestPayload{} })
	decoded, err := codec.Decode(event)
	if err != nil {
		t.Fatalf("decode outbox payload: %v", err)
	}
	payload, ok := decoded.Payload.(*notificationFanoutTestPayload)
	if !ok {
		t.Fatalf("unexpected decoded payload type: %T", decoded.Payload)
	}
	if payload.UserID != 7 || payload.Notification.ID != 101 || payload.Notification.Title != "title" || !payload.Notification.Unread {
		t.Fatalf("unexpected notification created payload: %+v", payload)
	}
}

func TestNotificationServicePublishAdminNotificationCreatesBatchAndFanOut(t *testing.T) {
	repo := &stubNotificationRepository{
		listAllUserIDsFn: func(_ context.Context) ([]int64, error) {
			return []int64{1, 2}, nil
		},
		listUserIDsByRolesFn: func(_ context.Context, roles []string) ([]int64, error) {
			if len(roles) != 1 || roles[0] != identitycontracts.RoleTeacher {
				t.Fatalf("unexpected roles filter: %+v", roles)
			}
			return []int64{2, 3}, nil
		},
		listUserIDsByClassesFn: func(_ context.Context, classNames []string) ([]int64, error) {
			if len(classNames) != 1 || classNames[0] != "ClassA" {
				t.Fatalf("unexpected class filter: %+v", classNames)
			}
			return []int64{3, 4}, nil
		},
		listExistingUserIDsFn: func(_ context.Context, userIDs []int64) ([]int64, error) {
			if len(userIDs) != 3 {
				t.Fatalf("unexpected user id filter: %+v", userIDs)
			}
			return []int64{4, 5}, nil
		},
	}
	broadcaster := &stubNotificationBroadcaster{}
	service := NewNotificationService(repo, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, broadcaster, zap.NewNop())
	link := "/notifications/admin"

	result, err := service.PublishAdminNotification(context.Background(), 99, PublishAdminNotificationInput{
		Type:    opsentity.NotificationTypeSystem,
		Title:   "系统通知",
		Content: "统一发布测试",
		Link:    &link,
		AudienceRules: NotificationAudienceRulesInput{
			Mode: "union",
			Rules: []NotificationAudienceRuleInput{
				{Type: NotificationAudienceTypeAll},
				{Type: NotificationAudienceTypeRole, Values: []string{identitycontracts.RoleTeacher}},
				{Type: NotificationAudienceTypeClass, Values: []string{"ClassA"}},
				{Type: NotificationAudienceTypeUser, Values: []string{"4", "5", "999"}},
			},
		},
	})
	if err != nil {
		t.Fatalf("PublishAdminNotification() error = %v", err)
	}
	if result.BatchID != 88 {
		t.Fatalf("BatchID = %d, want 88", result.BatchID)
	}
	if result.RecipientCount != 5 {
		t.Fatalf("RecipientCount = %d, want 5", result.RecipientCount)
	}
	if repo.createdBatch == nil {
		t.Fatal("expected notification batch to be created")
	}
	if repo.createdBatch.CreatedBy != 99 {
		t.Fatalf("created batch actor = %d, want 99", repo.createdBatch.CreatedBy)
	}
	if repo.createdBatch.AudienceMode != "union" {
		t.Fatalf("created batch audience mode = %q, want union", repo.createdBatch.AudienceMode)
	}
	if repo.createdBatch.RecipientCount != 5 {
		t.Fatalf("created batch recipient count = %d, want 5", repo.createdBatch.RecipientCount)
	}
	if len(repo.createdBatchNotifications) != 5 {
		t.Fatalf("batch notifications len = %d, want 5", len(repo.createdBatchNotifications))
	}

	expectedUsers := map[int64]struct{}{1: {}, 2: {}, 3: {}, 4: {}, 5: {}}
	for _, item := range repo.createdBatchNotifications {
		if _, ok := expectedUsers[item.UserID]; !ok {
			t.Fatalf("unexpected recipient user_id=%d", item.UserID)
		}
		delete(expectedUsers, item.UserID)
		if item.BatchID == nil || *item.BatchID != 88 {
			t.Fatalf("unexpected notification batch_id=%+v", item.BatchID)
		}
	}
	if len(expectedUsers) != 0 {
		t.Fatalf("missing recipients after dedupe: %+v", expectedUsers)
	}
	if len(broadcaster.sentUsers) != 0 {
		t.Fatalf("PublishAdminNotification should not publish websocket events directly, got %+v", broadcaster.sentUsers)
	}
	if len(repo.enqueued) != 5 {
		t.Fatalf("enqueued outbox events len = %d, want 5", len(repo.enqueued))
	}
	for _, event := range repo.enqueued {
		if event.Name != "notification.created" || event.PayloadVersion != 1 || event.Route != platformevents.OutboxRouteStream {
			t.Fatalf("unexpected outbox event: %+v", event)
		}
	}
}

func TestNotificationServiceHandleNotificationOutboxEventFansOutLocally(t *testing.T) {
	broadcaster := &stubNotificationBroadcaster{}
	service := NewNotificationService(&stubNotificationRepository{}, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, broadcaster, zap.NewNop())
	occurredAt := time.Date(2026, 6, 12, 10, 0, 0, 0, time.UTC)
	codec := platformevents.NewOutboxCodec()
	event, err := codec.Encode("notification.created", 1, notificationFanoutTestPayload{
		UserID: 7,
		Notification: NotificationInfo{
			ID:      101,
			Type:    opsentity.NotificationTypeSystem,
			Title:   "title",
			Content: ptrString("content"),
			Unread:  true,
		},
		OccurredAt: occurredAt,
	}, occurredAt)
	if err != nil {
		t.Fatalf("encode outbox event: %v", err)
	}

	if err := service.HandleNotificationFanoutEvent(context.Background(), event); err != nil {
		t.Fatalf("HandleNotificationFanoutEvent() error = %v", err)
	}

	if len(broadcaster.sentUsers) != 1 || broadcaster.sentUsers[0] != 7 {
		t.Fatalf("unexpected websocket recipients: %+v", broadcaster.sentUsers)
	}
	if len(broadcaster.envelopes) != 1 {
		t.Fatalf("websocket envelopes len = %d, want 1", len(broadcaster.envelopes))
	}
	if broadcaster.envelopes[0].Type != "notification.created" {
		t.Fatalf("unexpected websocket envelope type: %s", broadcaster.envelopes[0].Type)
	}
	if !broadcaster.envelopes[0].Timestamp.Equal(occurredAt) {
		t.Fatalf("unexpected websocket timestamp: %v", broadcaster.envelopes[0].Timestamp)
	}
}

func ptrString(value string) *string {
	return &value
}

func TestNotificationServicePublishAdminNotificationRejectsInvalidAudienceRule(t *testing.T) {
	service := NewNotificationService(&stubNotificationRepository{}, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, nil, zap.NewNop())

	_, err := service.PublishAdminNotification(context.Background(), 99, PublishAdminNotificationInput{
		Type:    opsentity.NotificationTypeSystem,
		Title:   "系统通知",
		Content: "invalid payload",
		AudienceRules: NotificationAudienceRulesInput{
			Mode: "union",
			Rules: []NotificationAudienceRuleInput{
				{Type: NotificationAudienceTypeRole},
			},
		},
	})
	if !errors.Is(err, apperror.ErrInvalidParams) {
		t.Fatalf("PublishAdminNotification() error = %v, want ErrInvalidParams", err)
	}
}

func TestNotificationServicePublishAdminNotificationRejectsUnknownRoleValue(t *testing.T) {
	service := NewNotificationService(&stubNotificationRepository{}, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, nil, zap.NewNop())

	_, err := service.PublishAdminNotification(context.Background(), 99, PublishAdminNotificationInput{
		Type:    opsentity.NotificationTypeSystem,
		Title:   "系统通知",
		Content: "invalid role",
		AudienceRules: NotificationAudienceRulesInput{
			Mode: "union",
			Rules: []NotificationAudienceRuleInput{
				{Type: NotificationAudienceTypeRole, Values: []string{"superadmin"}},
			},
		},
	})
	if !errors.Is(err, apperror.ErrInvalidParams) {
		t.Fatalf("PublishAdminNotification() error = %v, want ErrInvalidParams", err)
	}
}
