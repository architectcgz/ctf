package events

import (
	"context"
	"time"
)

const (
	OutboxRouteHandler = "handler"
	OutboxRouteStream  = "stream"

	OutboxStatusPending    = "pending"
	OutboxStatusDispatched = "dispatched"
)

type OutboxEvent struct {
	Name           string
	PayloadVersion int
	Payload        []byte
	Route          string
	DedupeKey      string
	OccurredAt     time.Time
}

type PendingOutboxEvent struct {
	ID    int64
	Event OutboxEvent
}

type OutboxRecord struct {
	ID              int64      `gorm:"column:id;primaryKey"`
	EventName       string     `gorm:"column:event_name"`
	Payload         []byte     `gorm:"column:payload"`
	PayloadVersion  int        `gorm:"column:payload_version"`
	Route           string     `gorm:"column:route"`
	DedupeKey       string     `gorm:"column:dedupe_key;uniqueIndex"`
	Status          string     `gorm:"column:status"`
	AttemptCount    int        `gorm:"column:attempt_count"`
	NextAttemptAt   time.Time  `gorm:"column:next_attempt_at"`
	LockedBy        string     `gorm:"column:locked_by"`
	LockedUntil     *time.Time `gorm:"column:locked_until"`
	StreamMessageID string     `gorm:"column:stream_message_id"`
	DispatchedAt    *time.Time `gorm:"column:dispatched_at"`
	LastError       string     `gorm:"column:last_error"`
	OccurredAt      time.Time  `gorm:"column:occurred_at"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
}

func (OutboxRecord) TableName() string {
	return "platform_event_outbox"
}

type OutboxStore interface {
	ClaimDue(ctx context.Context, workerID string, now time.Time, lease time.Duration, limit int) ([]PendingOutboxEvent, error)
	MarkDispatched(ctx context.Context, id int64, streamMessageID string, dispatchedAt time.Time) error
	MarkFailed(ctx context.Context, id int64, reason error, nextAttemptAt time.Time) error
}

type OutboxEventEnqueuer interface {
	EnqueueOutboxEvent(ctx context.Context, event OutboxEvent) error
}

func normalizeOutboxTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value.UTC()
}
