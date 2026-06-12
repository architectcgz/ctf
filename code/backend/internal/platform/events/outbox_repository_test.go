package events

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func TestOutboxRepositoryEnqueueClaimAndMarkDispatched(t *testing.T) {
	t.Parallel()

	db := newOutboxRepositoryTestDB(t)
	repo := NewOutboxRepository(db)
	now := time.Date(2026, 6, 12, 10, 0, 0, 0, time.FixedZone("CST", 8*60*60))

	if err := repo.Enqueue(context.Background(), OutboxEvent{
		Name:           "notification.created",
		PayloadVersion: 1,
		Payload:        []byte(`{"id":101,"user_id":7}`),
		Route:          OutboxRouteStream,
		DedupeKey:      "notification:101:created",
		OccurredAt:     now,
	}); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}

	claimed, err := repo.ClaimDue(context.Background(), "worker-a", now.UTC().Add(time.Second), time.Minute, 10)
	if err != nil {
		t.Fatalf("ClaimDue() error = %v", err)
	}
	if len(claimed) != 1 {
		t.Fatalf("claimed len = %d, want 1", len(claimed))
	}
	if claimed[0].Event.OccurredAt.Location() != time.UTC {
		t.Fatalf("claimed occurred_at location = %v, want UTC", claimed[0].Event.OccurredAt.Location())
	}

	claimedAgain, err := repo.ClaimDue(context.Background(), "worker-b", now.UTC().Add(time.Second), time.Minute, 10)
	if err != nil {
		t.Fatalf("ClaimDue(second) error = %v", err)
	}
	if len(claimedAgain) != 0 {
		t.Fatalf("second claim should not return leased row, got %+v", claimedAgain)
	}

	if err := repo.MarkDispatched(context.Background(), claimed[0].ID, "1710000000000-0", now.UTC().Add(2*time.Second)); err != nil {
		t.Fatalf("MarkDispatched() error = %v", err)
	}

	var row outboxRepositoryTestRow
	if err := db.Table("platform_event_outbox").First(&row, claimed[0].ID).Error; err != nil {
		t.Fatalf("load outbox row: %v", err)
	}
	if row.Status != OutboxStatusDispatched {
		t.Fatalf("status = %q, want %q", row.Status, OutboxStatusDispatched)
	}
	if row.StreamMessageID != "1710000000000-0" {
		t.Fatalf("stream_message_id = %q", row.StreamMessageID)
	}
}

func TestOutboxRepositoryMarkFailedSchedulesRetry(t *testing.T) {
	t.Parallel()

	db := newOutboxRepositoryTestDB(t)
	repo := NewOutboxRepository(db)
	now := time.Now().UTC()
	if err := repo.Enqueue(context.Background(), OutboxEvent{
		Name:           "practice.flag_accepted",
		PayloadVersion: 1,
		Payload:        []byte(`{"user_id":7}`),
		Route:          OutboxRouteHandler,
		DedupeKey:      "practice:7:11:accepted",
		OccurredAt:     now,
	}); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}
	claimed, err := repo.ClaimDue(context.Background(), "worker-a", now.Add(time.Second), time.Minute, 10)
	if err != nil {
		t.Fatalf("ClaimDue() error = %v", err)
	}
	if len(claimed) != 1 {
		t.Fatalf("claimed len = %d, want 1", len(claimed))
	}

	nextAttemptAt := now.Add(5 * time.Minute)
	if err := repo.MarkFailed(context.Background(), claimed[0].ID, errors.New("redis unavailable"), nextAttemptAt); err != nil {
		t.Fatalf("MarkFailed() error = %v", err)
	}

	var row outboxRepositoryTestRow
	if err := db.Table("platform_event_outbox").First(&row, claimed[0].ID).Error; err != nil {
		t.Fatalf("load outbox row: %v", err)
	}
	if row.Status != OutboxStatusPending {
		t.Fatalf("status = %q, want %q", row.Status, OutboxStatusPending)
	}
	if row.AttemptCount != 1 {
		t.Fatalf("attempt_count = %d, want 1", row.AttemptCount)
	}
	if row.LastError == "" {
		t.Fatalf("last_error should be recorded, got %+v", row)
	}

	notDue, err := repo.ClaimDue(context.Background(), "worker-b", nextAttemptAt.Add(-time.Second), time.Minute, 10)
	if err != nil {
		t.Fatalf("ClaimDue(before retry) error = %v", err)
	}
	if len(notDue) != 0 {
		t.Fatalf("claim before retry should be empty, got %+v", notDue)
	}
}

func TestOutboxRepositoryDedupeConflictTargetsPartialIndex(t *testing.T) {
	t.Parallel()

	db := newOutboxRepositoryTestDB(t)
	repo := NewOutboxRepository(db)
	sawOutboxCreate := false
	if err := db.Callback().Create().Before("gorm:create").Register("assert_outbox_dedupe_conflict_target", func(tx *gorm.DB) {
		if tx.Statement == nil || tx.Statement.Table != "platform_event_outbox" {
			return
		}
		sawOutboxCreate = true
		raw, ok := tx.Statement.Clauses["ON CONFLICT"]
		if !ok {
			tx.AddError(errors.New("missing ON CONFLICT clause"))
			return
		}
		onConflict, ok := raw.Expression.(clause.OnConflict)
		if !ok {
			tx.AddError(fmt.Errorf("unexpected ON CONFLICT expression %T", raw.Expression))
			return
		}
		if len(onConflict.TargetWhere.Exprs) != 1 {
			tx.AddError(fmt.Errorf("target where exprs len = %d, want 1", len(onConflict.TargetWhere.Exprs)))
			return
		}
		expr, ok := onConflict.TargetWhere.Exprs[0].(clause.Expr)
		if !ok || expr.SQL != "dedupe_key <> ''" {
			tx.AddError(fmt.Errorf("target where expr = %#v, want dedupe_key <> ''", onConflict.TargetWhere.Exprs[0]))
		}
	}); err != nil {
		t.Fatalf("register create callback: %v", err)
	}

	if err := repo.Enqueue(context.Background(), OutboxEvent{
		Name:           "notification.created",
		PayloadVersion: 1,
		Payload:        []byte(`{"id":101}`),
		Route:          OutboxRouteStream,
		DedupeKey:      "notification:101:created",
		OccurredAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}
	if !sawOutboxCreate {
		t.Fatal("expected outbox create callback to inspect ON CONFLICT clause")
	}
}

func TestOutboxRepositoryClaimDueDoesNotBypassFutureRetryAfterStaleRead(t *testing.T) {
	t.Parallel()

	db := newOutboxRepositoryTestDB(t)
	repo := NewOutboxRepository(db)
	now := time.Now().UTC()
	if err := repo.Enqueue(context.Background(), OutboxEvent{
		Name:           "practice.flag_accepted",
		PayloadVersion: 1,
		Payload:        []byte(`{"user_id":7}`),
		Route:          OutboxRouteHandler,
		DedupeKey:      "practice:7:11:accepted",
		OccurredAt:     now,
	}); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}
	var row OutboxRecord
	if err := db.Table("platform_event_outbox").First(&row).Error; err != nil {
		t.Fatalf("load outbox row: %v", err)
	}

	futureRetryAt := now.Add(5 * time.Minute)
	movedRetryForward := false
	if err := db.Callback().Update().Before("gorm:update").Register("move_outbox_retry_forward_before_claim", func(tx *gorm.DB) {
		if movedRetryForward || tx.Statement == nil || tx.Statement.Table != "platform_event_outbox" {
			return
		}
		movedRetryForward = true
		err := tx.Session(&gorm.Session{NewDB: true, SkipHooks: true}).
			Table("platform_event_outbox").
			Where("id = ?", row.ID).
			Updates(map[string]any{
				"next_attempt_at": futureRetryAt,
				"locked_by":       "",
				"locked_until":    nil,
				"updated_at":      now,
			}).Error
		if err != nil {
			tx.AddError(err)
		}
	}); err != nil {
		t.Fatalf("register update callback: %v", err)
	}

	claimed, err := repo.ClaimDue(context.Background(), "worker-b", now.Add(time.Second), time.Minute, 10)
	if err != nil {
		t.Fatalf("ClaimDue() error = %v", err)
	}
	if !movedRetryForward {
		t.Fatal("expected test callback to simulate another worker scheduling retry")
	}
	if len(claimed) != 0 {
		t.Fatalf("claimed len = %d, want 0 after retry moved to future: %+v", len(claimed), claimed)
	}
}

type outboxRepositoryTestRow struct {
	ID              int64
	Status          string
	AttemptCount    int
	StreamMessageID string
	LastError       string
}

func newOutboxRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "outbox.sqlite")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&OutboxRecord{}); err != nil {
		t.Fatalf("migrate outbox: %v", err)
	}
	return db
}
