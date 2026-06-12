package infrastructure

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	opsentity "ctf-platform/internal/module/ops/entity"
	opsports "ctf-platform/internal/module/ops/ports"
	platformevents "ctf-platform/internal/platform/events"
)

func TestNotificationRepositoryRollsBackNotificationWhenOutboxEnqueueFails(t *testing.T) {
	t.Parallel()

	db := newNotificationRepositoryTestDB(t)
	if err := db.Callback().Create().Before("gorm:create").Register("fail_notification_outbox_insert", func(tx *gorm.DB) {
		if tx.Statement != nil && tx.Statement.Table == "platform_event_outbox" {
			tx.AddError(gorm.ErrInvalidDB)
		}
	}); err != nil {
		t.Fatalf("register outbox failure callback: %v", err)
	}

	repo := NewNotificationRepository(db)
	err := repo.WithinNotificationOutboxTx(context.Background(), func(txRepo opsports.NotificationOutboxTxRepository) error {
		if err := txRepo.Create(context.Background(), &opsentity.Notification{
			UserID:  7,
			Type:    opsentity.NotificationTypeSystem,
			Title:   "outbox failure",
			Content: "rollback expected",
		}); err != nil {
			return err
		}
		return txRepo.EnqueueOutboxEvent(context.Background(), platformevents.OutboxEvent{
			Name:           "notification.created",
			PayloadVersion: 1,
			Payload:        []byte(`{"user_id":7}`),
			Route:          platformevents.OutboxRouteStream,
			DedupeKey:      "notification:created:rollback",
			OccurredAt:     time.Now().UTC(),
		})
	})
	if err == nil {
		t.Fatal("expected transaction to fail when outbox enqueue fails")
	}

	var notificationCount int64
	if err := db.Model(&opsentity.Notification{}).Count(&notificationCount).Error; err != nil {
		t.Fatalf("count notifications: %v", err)
	}
	if notificationCount != 0 {
		t.Fatalf("expected notification insert to roll back, got %d notifications", notificationCount)
	}
}

func TestNotificationRepositoryCreateIfSourceEventAbsentIsIdempotent(t *testing.T) {
	t.Parallel()

	db := newNotificationRepositoryTestDB(t)
	repo := NewNotificationRepository(db)
	ctx := context.Background()

	created, err := repo.CreateIfSourceEventAbsent(ctx, &opsentity.Notification{
		UserID:         7,
		Type:           opsentity.NotificationTypeChallenge,
		Title:          "flag accepted",
		Content:        "first",
		SourceEventKey: "outbox:practice:flag_accepted:7:11:notification",
	})
	if err != nil {
		t.Fatalf("CreateIfSourceEventAbsent(first) error = %v", err)
	}
	if !created {
		t.Fatal("CreateIfSourceEventAbsent(first) created = false, want true")
	}

	created, err = repo.CreateIfSourceEventAbsent(ctx, &opsentity.Notification{
		UserID:         7,
		Type:           opsentity.NotificationTypeChallenge,
		Title:          "flag accepted duplicate",
		Content:        "second",
		SourceEventKey: "outbox:practice:flag_accepted:7:11:notification",
	})
	if err != nil {
		t.Fatalf("CreateIfSourceEventAbsent(second) error = %v", err)
	}
	if created {
		t.Fatal("CreateIfSourceEventAbsent(second) created = true, want false")
	}

	var notifications []opsentity.Notification
	if err := db.Order("id ASC").Find(&notifications).Error; err != nil {
		t.Fatalf("load notifications: %v", err)
	}
	if len(notifications) != 1 {
		t.Fatalf("notifications len = %d, want 1", len(notifications))
	}
	if notifications[0].Content != "first" {
		t.Fatalf("notification content = %q, want first", notifications[0].Content)
	}
}

func newNotificationRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&opsentity.Notification{}, &platformevents.OutboxRecord{}); err != nil {
		t.Fatalf("auto migrate schema: %v", err)
	}
	if err := db.Exec("CREATE UNIQUE INDEX idx_notifications_source_event_key ON notifications (source_event_key) WHERE source_event_key <> ''").Error; err != nil {
		t.Fatalf("create source event key index: %v", err)
	}
	return db
}
