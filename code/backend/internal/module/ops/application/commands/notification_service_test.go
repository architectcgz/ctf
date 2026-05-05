package commands

import (
	"context"
	"errors"
	"testing"
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

type stubNotificationRepository struct {
	created                   []*model.Notification
	createdBatch              *model.NotificationBatch
	createdBatchNotifications []*model.Notification
	findByIDFn                func(ctx context.Context, notificationID, userID int64) (*model.Notification, error)
	listAllUserIDsFn          func(ctx context.Context) ([]int64, error)
	listUserIDsByRolesFn      func(ctx context.Context, roles []string) ([]int64, error)
	listUserIDsByClassesFn    func(ctx context.Context, classNames []string) ([]int64, error)
	listExistingUserIDsFn     func(ctx context.Context, userIDs []int64) ([]int64, error)
	markAsReadCalls           int
}

func (r *stubNotificationRepository) Create(_ context.Context, notification *model.Notification) error {
	copied := *notification
	r.created = append(r.created, &copied)
	return nil
}

func (r *stubNotificationRepository) List(_ context.Context, _ opsports.NotificationListFilter) ([]model.Notification, int64, error) {
	return nil, 0, nil
}

func (r *stubNotificationRepository) FindByID(ctx context.Context, notificationID, userID int64) (*model.Notification, error) {
	if r.findByIDFn == nil {
		return nil, errors.New("unexpected FindByID call")
	}
	return r.findByIDFn(ctx, notificationID, userID)
}

func (r *stubNotificationRepository) MarkAsRead(_ context.Context, _ int64, _ int64, _ any) error {
	r.markAsReadCalls++
	return nil
}

func (r *stubNotificationRepository) CreateBatch(_ context.Context, batch *model.NotificationBatch, notifications []*model.Notification) error {
	copiedBatch := *batch
	if copiedBatch.ID == 0 {
		copiedBatch.ID = 88
		batch.ID = 88
	}
	r.createdBatch = &copiedBatch
	r.createdBatchNotifications = make([]*model.Notification, 0, len(notifications))
	for _, item := range notifications {
		item.BatchID = &copiedBatch.ID
		copied := *item
		r.createdBatchNotifications = append(r.createdBatchNotifications, &copied)
	}
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

func TestNotificationServiceRegisterPracticeEventConsumers(t *testing.T) {
	repo := &stubNotificationRepository{}
	service := NewNotificationService(repo, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, nil, zap.NewNop())
	bus := &recordingBus{}

	service.RegisterPracticeEventConsumers(bus)

	if got := len(bus.subscribers[practicecontracts.EventFlagAccepted]); got != 1 {
		t.Fatalf("flag_accepted subscribers = %d, want 1", got)
	}
	if got := len(bus.subscribers); got != 1 {
		t.Fatalf("practice subscriber count = %d, want 1", got)
	}

	err := bus.Publish(context.Background(), platformevents.Event{
		Name: practicecontracts.EventFlagAccepted,
		Payload: practicecontracts.FlagAcceptedEvent{
			UserID:      7,
			ChallengeID: 12,
			Points:      30,
		},
	})
	if err != nil {
		t.Fatalf("Publish(flag_accepted) error = %v", err)
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
}

func TestNotificationServiceMarkAsReadReturnsNotificationNotFound(t *testing.T) {
	service := NewNotificationService(&stubNotificationRepository{
		findByIDFn: func(_ context.Context, _, _ int64) (*model.Notification, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, nil, zap.NewNop())

	err := service.MarkAsRead(context.Background(), 7, 11)
	if !errors.Is(err, errcode.ErrNotificationNotFound) {
		t.Fatalf("MarkAsRead() error = %v, want ErrNotificationNotFound", err)
	}
}

func TestNotificationServiceMarkAsReadIsIdempotentForReadNotification(t *testing.T) {
	repo := &stubNotificationRepository{
		findByIDFn: func(_ context.Context, _, _ int64) (*model.Notification, error) {
			readAt := time.Now().UTC()
			return &model.Notification{
				ID:      11,
				UserID:  7,
				Title:   "already read",
				IsRead:  true,
				ReadAt:  &readAt,
				Type:    model.NotificationTypeSystem,
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

func TestNotificationServiceSendNotificationPublishesWebsocketEvent(t *testing.T) {
	repo := &stubNotificationRepository{}
	broadcaster := &stubNotificationBroadcaster{}
	service := NewNotificationService(repo, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, broadcaster, zap.NewNop())

	if err := service.SendNotification(context.Background(), 7, SendNotificationInput{
		Type:    model.NotificationTypeSystem,
		Title:   "title",
		Content: "content",
	}); err != nil {
		t.Fatalf("SendNotification() error = %v", err)
	}
	if len(repo.created) != 1 {
		t.Fatalf("created notifications len = %d, want 1", len(repo.created))
	}
	if len(broadcaster.envelopes) != 1 || broadcaster.envelopes[0].Type != "notification.created" {
		t.Fatalf("unexpected websocket events: %+v", broadcaster.envelopes)
	}
}

func TestNotificationServicePublishAdminNotificationCreatesBatchAndFanOut(t *testing.T) {
	repo := &stubNotificationRepository{
		listAllUserIDsFn: func(_ context.Context) ([]int64, error) {
			return []int64{1, 2}, nil
		},
		listUserIDsByRolesFn: func(_ context.Context, roles []string) ([]int64, error) {
			if len(roles) != 1 || roles[0] != model.RoleTeacher {
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
		Type:    model.NotificationTypeSystem,
		Title:   "系统通知",
		Content: "统一发布测试",
		Link:    &link,
		AudienceRules: NotificationAudienceRulesInput{
			Mode: "union",
			Rules: []NotificationAudienceRuleInput{
				{Type: dto.NotificationAudienceTypeAll},
				{Type: dto.NotificationAudienceTypeRole, Values: []string{model.RoleTeacher}},
				{Type: dto.NotificationAudienceTypeClass, Values: []string{"ClassA"}},
				{Type: dto.NotificationAudienceTypeUser, Values: []string{"4", "5", "999"}},
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
	if len(broadcaster.sentUsers) != 5 {
		t.Fatalf("websocket send count = %d, want 5", len(broadcaster.sentUsers))
	}
	for _, envelope := range broadcaster.envelopes {
		if envelope.Type != "notification.created" {
			t.Fatalf("unexpected websocket envelope type: %s", envelope.Type)
		}
	}
}

func TestNotificationServicePublishAdminNotificationRejectsInvalidAudienceRule(t *testing.T) {
	service := NewNotificationService(&stubNotificationRepository{}, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, nil, zap.NewNop())

	_, err := service.PublishAdminNotification(context.Background(), 99, PublishAdminNotificationInput{
		Type:    model.NotificationTypeSystem,
		Title:   "系统通知",
		Content: "invalid payload",
		AudienceRules: NotificationAudienceRulesInput{
			Mode: "union",
			Rules: []NotificationAudienceRuleInput{
				{Type: dto.NotificationAudienceTypeRole},
			},
		},
	})
	if !errors.Is(err, errcode.ErrInvalidParams) {
		t.Fatalf("PublishAdminNotification() error = %v, want ErrInvalidParams", err)
	}
}

func TestNotificationServicePublishAdminNotificationRejectsUnknownRoleValue(t *testing.T) {
	service := NewNotificationService(&stubNotificationRepository{}, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, nil, zap.NewNop())

	_, err := service.PublishAdminNotification(context.Background(), 99, PublishAdminNotificationInput{
		Type:    model.NotificationTypeSystem,
		Title:   "系统通知",
		Content: "invalid role",
		AudienceRules: NotificationAudienceRulesInput{
			Mode: "union",
			Rules: []NotificationAudienceRuleInput{
				{Type: dto.NotificationAudienceTypeRole, Values: []string{"superadmin"}},
			},
		},
	})
	if !errors.Is(err, errcode.ErrInvalidParams) {
		t.Fatalf("PublishAdminNotification() error = %v, want ErrInvalidParams", err)
	}
}
