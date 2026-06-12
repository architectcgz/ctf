package events

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OutboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) *OutboxRepository {
	if db == nil {
		return nil
	}
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) WithDB(db *gorm.DB) *OutboxRepository {
	return NewOutboxRepository(db)
}

func (r *OutboxRepository) Enqueue(ctx context.Context, event OutboxEvent) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("outbox repository is not configured")
	}
	now := time.Now().UTC()
	occurredAt := normalizeOutboxTime(event.OccurredAt)
	route := event.Route
	if route == "" {
		route = OutboxRouteHandler
	}
	record := OutboxRecord{
		EventName:      event.Name,
		Payload:        event.Payload,
		PayloadVersion: event.PayloadVersion,
		Route:          route,
		DedupeKey:      event.DedupeKey,
		Status:         OutboxStatusPending,
		NextAttemptAt:  occurredAt,
		OccurredAt:     occurredAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	query := r.db.WithContext(ctx)
	if event.DedupeKey != "" {
		query = query.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "dedupe_key"}},
			TargetWhere: clause.Where{Exprs: []clause.Expression{
				clause.Expr{SQL: "dedupe_key <> ''"},
			}},
			DoNothing: true,
		})
	}
	return query.Create(&record).Error
}

func (r *OutboxRepository) ClaimDue(ctx context.Context, workerID string, now time.Time, lease time.Duration, limit int) ([]PendingOutboxEvent, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("outbox repository is not configured")
	}
	if limit <= 0 {
		limit = 32
	}
	if lease <= 0 {
		lease = time.Minute
	}
	now = normalizeOutboxTime(now)
	lockedUntil := now.Add(lease)

	var rows []OutboxRecord
	if err := r.db.WithContext(ctx).
		Where("status = ? AND next_attempt_at <= ? AND (locked_until IS NULL OR locked_until <= ?)", OutboxStatusPending, now, now).
		Order("id ASC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]PendingOutboxEvent, 0, len(rows))
	for _, row := range rows {
		result := r.db.WithContext(ctx).
			Model(&OutboxRecord{}).
			Where("id = ? AND status = ? AND next_attempt_at <= ? AND (locked_until IS NULL OR locked_until <= ?)", row.ID, OutboxStatusPending, now, now).
			Updates(map[string]any{
				"locked_by":    workerID,
				"locked_until": lockedUntil,
				"updated_at":   now,
			})
		if result.Error != nil {
			return nil, result.Error
		}
		if result.RowsAffected == 0 {
			continue
		}
		items = append(items, PendingOutboxEvent{ID: row.ID, Event: outboxEventFromRecord(row)})
	}
	return items, nil
}

func (r *OutboxRepository) MarkDispatched(ctx context.Context, id int64, streamMessageID string, dispatchedAt time.Time) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("outbox repository is not configured")
	}
	now := normalizeOutboxTime(dispatchedAt)
	return r.db.WithContext(ctx).
		Model(&OutboxRecord{}).
		Where("id = ? AND status = ?", id, OutboxStatusPending).
		Updates(map[string]any{
			"status":            OutboxStatusDispatched,
			"stream_message_id": streamMessageID,
			"dispatched_at":     now,
			"locked_by":         "",
			"locked_until":      nil,
			"last_error":        "",
			"updated_at":        now,
		}).Error
}

func (r *OutboxRepository) MarkFailed(ctx context.Context, id int64, reason error, nextAttemptAt time.Time) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("outbox repository is not configured")
	}
	now := time.Now().UTC()
	message := ""
	if reason != nil {
		message = reason.Error()
	}
	return r.db.WithContext(ctx).
		Model(&OutboxRecord{}).
		Where("id = ? AND status = ?", id, OutboxStatusPending).
		Updates(map[string]any{
			"status":          OutboxStatusPending,
			"attempt_count":   gorm.Expr("attempt_count + 1"),
			"next_attempt_at": normalizeOutboxTime(nextAttemptAt),
			"locked_by":       "",
			"locked_until":    nil,
			"last_error":      message,
			"updated_at":      now,
		}).Error
}

func outboxEventFromRecord(row OutboxRecord) OutboxEvent {
	return OutboxEvent{
		Name:           row.EventName,
		PayloadVersion: row.PayloadVersion,
		Payload:        row.Payload,
		Route:          row.Route,
		DedupeKey:      row.DedupeKey,
		OccurredAt:     normalizeOutboxTime(row.OccurredAt),
	}
}

func isUniqueConstraint(err error) bool {
	return errors.Is(err, gorm.ErrDuplicatedKey)
}

var _ OutboxStore = (*OutboxRepository)(nil)
