package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

type RealtimeOutboxRepository struct {
	db *gorm.DB
}

func NewRealtimeOutboxRepository(db *gorm.DB) *RealtimeOutboxRepository {
	return &RealtimeOutboxRepository{db: db}
}

func (r *RealtimeOutboxRepository) WithDB(db *gorm.DB) *RealtimeOutboxRepository {
	return &RealtimeOutboxRepository{db: db}
}

func (r *RealtimeOutboxRepository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *RealtimeOutboxRepository) EnqueueRealtimeRelay(ctx context.Context, relay contestcontracts.RealtimeRelayEvent, dedupeKey string) error {
	payload, err := json.Marshal(relay.Payload)
	if err != nil {
		return fmt.Errorf("marshal realtime relay payload: %w", err)
	}
	record := contestentity.ContestRealtimeOutbox{
		EventName:       relay.EventName,
		Delivery:        relay.Delivery,
		Channel:         relay.Channel,
		RecipientUserID: relay.RecipientUserID,
		MessageType:     relay.MessageType,
		Payload:         string(payload),
		DedupeKey:       dedupeKey,
		Status:          contestdomain.ContestRealtimeOutboxStatusPending,
		EventOccurredAt: relayTimestamp(relay.Timestamp),
		NextAttemptAt:   relayTimestamp(relay.Timestamp),
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	return r.dbWithContext(ctx).Create(&record).Error
}

func (r *RealtimeOutboxRepository) ListPendingRealtimeRelays(ctx context.Context, before time.Time, limit int) ([]contestports.ContestPendingRealtimeRelay, error) {
	if limit <= 0 {
		limit = 100
	}
	var rows []contestentity.ContestRealtimeOutbox
	if err := r.dbWithContext(ctx).
		Where("status = ? AND next_attempt_at <= ?", contestdomain.ContestRealtimeOutboxStatusPending, relayTimestamp(before)).
		Order("id ASC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]contestports.ContestPendingRealtimeRelay, 0, len(rows))
	for _, row := range rows {
		relay, err := decodeRealtimeRelay(row)
		if err != nil {
			return nil, err
		}
		items = append(items, contestports.ContestPendingRealtimeRelay{
			ID:        row.ID,
			DedupeKey: row.DedupeKey,
			Relay:     relay,
		})
	}
	return items, nil
}

func (r *RealtimeOutboxRepository) MarkRealtimeRelaySent(ctx context.Context, outboxID int64, streamMessageID string, sentAt time.Time) error {
	now := sentAt.UTC()
	return r.dbWithContext(ctx).
		Model(&contestentity.ContestRealtimeOutbox{}).
		Where("id = ? AND status = ?", outboxID, contestdomain.ContestRealtimeOutboxStatusPending).
		Updates(map[string]any{
			"status":            contestdomain.ContestRealtimeOutboxStatusSent,
			"stream_message_id": streamMessageID,
			"sent_at":           now,
			"updated_at":        now,
			"last_error":        "",
		}).Error
}

func (r *RealtimeOutboxRepository) MarkRealtimeRelayFailed(ctx context.Context, outboxID int64, reason error, nextAttemptAt time.Time) error {
	return r.dbWithContext(ctx).
		Model(&contestentity.ContestRealtimeOutbox{}).
		Where("id = ? AND status = ?", outboxID, contestdomain.ContestRealtimeOutboxStatusPending).
		Updates(map[string]any{
			"status":          contestdomain.ContestRealtimeOutboxStatusPending,
			"attempt_count":   gorm.Expr("attempt_count + 1"),
			"next_attempt_at": relayTimestamp(nextAttemptAt),
			"last_error":      errorString(reason),
			"updated_at":      time.Now().UTC(),
		}).Error
}

func decodeRealtimeRelay(row contestentity.ContestRealtimeOutbox) (contestcontracts.RealtimeRelayEvent, error) {
	var payload any
	if row.Payload != "" {
		decoded, err := contestcontracts.DecodeRealtimeRelayPayload(row.EventName, []byte(row.Payload))
		if err != nil {
			return contestcontracts.RealtimeRelayEvent{}, fmt.Errorf("decode realtime relay payload: %w", err)
		}
		payload = decoded
	}
	return contestcontracts.RealtimeRelayEvent{
		EventName:       row.EventName,
		Delivery:        row.Delivery,
		Channel:         row.Channel,
		RecipientUserID: row.RecipientUserID,
		MessageType:     row.MessageType,
		Payload:         payload,
		Timestamp:       relayTimestamp(row.EventOccurredAt),
	}, nil
}

func relayTimestamp(ts time.Time) time.Time {
	if ts.IsZero() {
		return time.Now().UTC()
	}
	return ts.UTC()
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
