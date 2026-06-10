package ports

import (
	"context"
	"time"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
)

type ContestPendingRealtimeRelay struct {
	ID        int64
	DedupeKey string
	Relay     contestcontracts.RealtimeRelayEvent
}

type ContestRealtimeOutboxRepository interface {
	EnqueueRealtimeRelay(ctx context.Context, relay contestcontracts.RealtimeRelayEvent, dedupeKey string) error
	ListPendingRealtimeRelays(ctx context.Context, before time.Time, limit int) ([]ContestPendingRealtimeRelay, error)
	MarkRealtimeRelaySent(ctx context.Context, outboxID int64, streamMessageID string, sentAt time.Time) error
	MarkRealtimeRelayFailed(ctx context.Context, outboxID int64, reason error, nextAttemptAt time.Time) error
}
