package commands

import (
	"context"
	"time"

	"go.uber.org/zap"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestports "ctf-platform/internal/module/contest/ports"
)

type contestRealtimeOutbox interface {
	ListPendingRealtimeRelays(ctx context.Context, before time.Time, limit int) ([]contestports.ContestPendingRealtimeRelay, error)
	MarkRealtimeRelaySent(ctx context.Context, outboxID int64, streamMessageID string, sentAt time.Time) error
	MarkRealtimeRelayFailed(ctx context.Context, outboxID int64, reason error, nextAttemptAt time.Time) error
}

type contestRealtimeRelayStreamPublisher interface {
	Publish(ctx context.Context, relay contestcontracts.RealtimeRelayEvent, dedupeKey string) (string, error)
}

type ContestRealtimeOutboxDispatcher struct {
	outbox       contestRealtimeOutbox
	publisher    contestRealtimeRelayStreamPublisher
	logger       *zap.Logger
	batchSize    int
	retryBackoff time.Duration
}

func NewContestRealtimeOutboxDispatcher(outbox contestRealtimeOutbox, publisher contestRealtimeRelayStreamPublisher, logger *zap.Logger) *ContestRealtimeOutboxDispatcher {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ContestRealtimeOutboxDispatcher{
		outbox:       outbox,
		publisher:    publisher,
		logger:       logger,
		batchSize:    32,
		retryBackoff: 5 * time.Second,
	}
}

func (d *ContestRealtimeOutboxDispatcher) Run(ctx context.Context) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		if err := d.DispatchOnce(ctx); err != nil && d.logger != nil {
			d.logger.Warn("dispatch contest realtime outbox failed", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (d *ContestRealtimeOutboxDispatcher) DispatchOnce(ctx context.Context) error {
	if d == nil || d.outbox == nil || d.publisher == nil {
		return nil
	}
	now := time.Now().UTC()
	items, err := d.outbox.ListPendingRealtimeRelays(ctx, now, d.batchSize)
	if err != nil {
		return err
	}
	for _, item := range items {
		streamMessageID, err := d.publisher.Publish(ctx, item.Relay, item.DedupeKey)
		if err != nil {
			if markErr := d.outbox.MarkRealtimeRelayFailed(ctx, item.ID, err, now.Add(d.retryBackoff)); markErr != nil {
				return markErr
			}
			continue
		}
		if err := d.outbox.MarkRealtimeRelaySent(ctx, item.ID, streamMessageID, now); err != nil {
			return err
		}
	}
	return nil
}
