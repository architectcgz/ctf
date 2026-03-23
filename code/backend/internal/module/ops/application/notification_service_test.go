package application

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
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	platformevents "ctf-platform/internal/platform/events"
	"ctf-platform/pkg/errcode"
	ctfws "ctf-platform/pkg/websocket"
)

type stubNotificationRepository struct {
	created         []*model.Notification
	findByIDFn      func(ctx context.Context, notificationID, userID int64) (*model.Notification, error)
	markAsReadCalls int
}

func (r *stubNotificationRepository) Create(_ context.Context, notification *model.Notification) error {
	copied := *notification
	r.created = append(r.created, &copied)
	return nil
}

func (r *stubNotificationRepository) List(_ context.Context, _ NotificationListFilter) ([]model.Notification, int64, error) {
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
	if got := len(bus.subscribers[practicecontracts.EventHintUnlocked]); got != 1 {
		t.Fatalf("hint_unlocked subscribers = %d, want 1", got)
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

	err = bus.Publish(context.Background(), platformevents.Event{
		Name: practicecontracts.EventHintUnlocked,
		Payload: practicecontracts.HintUnlockedEvent{
			UserID:      7,
			ChallengeID: 12,
			HintLevel:   2,
		},
	})
	if err != nil {
		t.Fatalf("Publish(hint_unlocked) error = %v", err)
	}

	if len(repo.created) != 2 {
		t.Fatalf("created notifications len = %d, want 2", len(repo.created))
	}
	if repo.created[0].Title != "题目解出" || repo.created[1].Title != "提示已解锁" {
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

func TestNotificationServiceGetNotificationsNormalizesPagination(t *testing.T) {
	service := NewNotificationService(&stubNotificationRepository{}, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, nil, zap.NewNop())

	items, total, page, pageSize, err := service.GetNotifications(context.Background(), 7, &dto.NotificationQuery{
		Page:     0,
		PageSize: 999,
	})
	if err != nil {
		t.Fatalf("GetNotifications() error = %v", err)
	}
	if len(items) != 0 || total != 0 {
		t.Fatalf("GetNotifications() items=%d total=%d, want empty result", len(items), total)
	}
	if page != 1 || pageSize != 100 {
		t.Fatalf("GetNotifications() page=%d pageSize=%d, want 1 and 100", page, pageSize)
	}
}
