package events

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func TestStreamFanoutPublishConsumeAndAdvanceCursor(t *testing.T) {
	t.Parallel()

	cache := newStreamFanoutRedis(t)
	fanout := NewStreamFanout(cache, StreamFanoutOptions{
		StreamKey:        "platform:events:test",
		CursorKeyPrefix:  "platform:events:test:cursor:",
		PublishKeyPrefix: "platform:events:test:publish:",
	}, zap.NewNop())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := cache.Set(ctx, "platform:events:test:cursor:replica-a", "0-0", 0).Err(); err != nil {
		t.Fatalf("seed cursor: %v", err)
	}
	if _, err := fanout.Publish(ctx, OutboxEvent{
		Name:           "notification.created",
		PayloadVersion: 1,
		Payload:        []byte(`{"id":101,"user_id":7}`),
		Route:          OutboxRouteStream,
		DedupeKey:      "notification:101:created",
		OccurredAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	var handled []OutboxEvent
	if err := fanout.ConsumeOnce(ctx, "replica-a", func(_ context.Context, event OutboxEvent) error {
		handled = append(handled, event)
		return nil
	}); err != nil {
		t.Fatalf("ConsumeOnce() error = %v", err)
	}
	if len(handled) != 1 || handled[0].Name != "notification.created" {
		t.Fatalf("handled events = %+v", handled)
	}
	cursor, err := cache.Get(ctx, "platform:events:test:cursor:replica-a").Result()
	if err != nil || cursor == "0-0" || cursor == "" {
		t.Fatalf("cursor = %q err=%v, want advanced cursor", cursor, err)
	}
}

func TestStreamFanoutDoesNotAdvanceCursorWhenHandlerFails(t *testing.T) {
	t.Parallel()

	cache := newStreamFanoutRedis(t)
	fanout := NewStreamFanout(cache, StreamFanoutOptions{
		StreamKey:        "platform:events:test",
		CursorKeyPrefix:  "platform:events:test:cursor:",
		PublishKeyPrefix: "platform:events:test:publish:",
	}, zap.NewNop())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := cache.Set(ctx, "platform:events:test:cursor:replica-a", "0-0", 0).Err(); err != nil {
		t.Fatalf("seed cursor: %v", err)
	}
	if _, err := fanout.Publish(ctx, OutboxEvent{
		Name:           "notification.read",
		PayloadVersion: 1,
		Payload:        []byte(`{"id":101}`),
		Route:          OutboxRouteStream,
		DedupeKey:      "notification:101:read",
		OccurredAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if err := fanout.ConsumeOnce(ctx, "replica-a", func(context.Context, OutboxEvent) error {
		return errors.New("websocket fanout failed")
	}); err == nil {
		t.Fatal("ConsumeOnce() error = nil, want handler error")
	}

	cursor, err := cache.Get(ctx, "platform:events:test:cursor:replica-a").Result()
	if err != nil {
		t.Fatalf("load cursor: %v", err)
	}
	if cursor != "0-0" {
		t.Fatalf("cursor = %q, want original cursor after handler failure", cursor)
	}
}

func TestStreamFanoutPublishIsIdempotentByDedupeKey(t *testing.T) {
	t.Parallel()

	cache := newStreamFanoutRedis(t)
	fanout := NewStreamFanout(cache, StreamFanoutOptions{
		StreamKey:        "platform:events:test",
		CursorKeyPrefix:  "platform:events:test:cursor:",
		PublishKeyPrefix: "platform:events:test:publish:",
	}, zap.NewNop())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	event := OutboxEvent{
		Name:           "notification.created",
		PayloadVersion: 1,
		Payload:        []byte(`{"id":101}`),
		Route:          OutboxRouteStream,
		DedupeKey:      "notification:101:created",
		OccurredAt:     time.Now().UTC(),
	}
	first, err := fanout.Publish(ctx, event)
	if err != nil {
		t.Fatalf("Publish(first) error = %v", err)
	}
	second, err := fanout.Publish(ctx, event)
	if err != nil {
		t.Fatalf("Publish(second) error = %v", err)
	}
	if first == "" || first != second {
		t.Fatalf("message IDs = %q/%q, want identical non-empty IDs", first, second)
	}
	messages, err := cache.XRange(ctx, "platform:events:test", "-", "+").Result()
	if err != nil {
		t.Fatalf("XRange() error = %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("stream messages len = %d, want 1", len(messages))
	}
}

func newStreamFanoutRedis(t *testing.T) *redislib.Client {
	t.Helper()
	mr := miniredis.RunT(t)
	cache := redislib.NewClient(&redislib.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = cache.Close() })
	return cache
}
