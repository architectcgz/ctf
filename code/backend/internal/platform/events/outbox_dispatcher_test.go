package events

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestOutboxDispatcherPublishesStreamRouteAndMarksDispatched(t *testing.T) {
	t.Parallel()

	store := &stubOutboxStore{
		claimed: []PendingOutboxEvent{{
			ID: 42,
			Event: OutboxEvent{
				Name:           "notification.created",
				PayloadVersion: 1,
				Payload:        []byte(`{"id":101}`),
				Route:          OutboxRouteStream,
				DedupeKey:      "notification:101:created",
				OccurredAt:     time.Now().UTC(),
			},
		}},
	}
	publisher := &stubStreamPublisher{messageID: "1710000000000-0"}
	dispatcher := NewOutboxDispatcher(store, publisher, nil, zap.NewNop())

	if err := dispatcher.DispatchOnce(context.Background(), "worker-a"); err != nil {
		t.Fatalf("DispatchOnce() error = %v", err)
	}
	if len(publisher.published) != 1 || publisher.published[0].Name != "notification.created" {
		t.Fatalf("published = %+v", publisher.published)
	}
	if len(store.dispatchedIDs) != 1 || store.dispatchedIDs[0] != 42 {
		t.Fatalf("dispatched ids = %+v", store.dispatchedIDs)
	}
	if len(store.failedIDs) != 0 {
		t.Fatalf("failed ids = %+v", store.failedIDs)
	}
}

func TestOutboxDispatcherRetriesFailedStreamPublish(t *testing.T) {
	t.Parallel()

	store := &stubOutboxStore{
		claimed: []PendingOutboxEvent{{
			ID: 9,
			Event: OutboxEvent{
				Name:           "notification.read",
				PayloadVersion: 1,
				Payload:        []byte(`{"id":101}`),
				Route:          OutboxRouteStream,
				DedupeKey:      "notification:101:read",
				OccurredAt:     time.Now().UTC(),
			},
		}},
	}
	dispatcher := NewOutboxDispatcher(store, &stubStreamPublisher{err: errors.New("redis unavailable")}, nil, zap.NewNop())

	if err := dispatcher.DispatchOnce(context.Background(), "worker-a"); err != nil {
		t.Fatalf("DispatchOnce() error = %v", err)
	}
	if len(store.failedIDs) != 1 || store.failedIDs[0] != 9 {
		t.Fatalf("failed ids = %+v", store.failedIDs)
	}
	if len(store.dispatchedIDs) != 0 {
		t.Fatalf("dispatched ids = %+v", store.dispatchedIDs)
	}
}

func TestOutboxDispatcherInvokesRegisteredHandlerRoute(t *testing.T) {
	t.Parallel()

	store := &stubOutboxStore{
		claimed: []PendingOutboxEvent{{
			ID: 77,
			Event: OutboxEvent{
				Name:           "practice.flag_accepted",
				PayloadVersion: 1,
				Payload:        []byte(`{"user_id":7}`),
				Route:          OutboxRouteHandler,
				DedupeKey:      "practice:7:11:accepted",
				OccurredAt:     time.Now().UTC(),
			},
		}},
	}
	registry := NewOutboxHandlerRegistry()
	called := false
	registry.Register("practice.flag_accepted", func(ctx context.Context, event OutboxEvent) error {
		called = true
		if event.Name != "practice.flag_accepted" {
			t.Fatalf("handler event = %+v", event)
		}
		return nil
	})
	dispatcher := NewOutboxDispatcher(store, nil, registry, zap.NewNop())

	if err := dispatcher.DispatchOnce(context.Background(), "worker-a"); err != nil {
		t.Fatalf("DispatchOnce() error = %v", err)
	}
	if !called {
		t.Fatal("expected registered handler to be called")
	}
	if len(store.dispatchedIDs) != 1 || store.dispatchedIDs[0] != 77 {
		t.Fatalf("dispatched ids = %+v", store.dispatchedIDs)
	}
}

func TestOutboxHandlerRegistryInvokesMultipleHandlersForSameEvent(t *testing.T) {
	t.Parallel()

	registry := NewOutboxHandlerRegistry()
	calls := make([]string, 0, 2)
	registry.Register("practice.flag_accepted", func(ctx context.Context, event OutboxEvent) error {
		calls = append(calls, "notification")
		return nil
	})
	registry.Register("practice.flag_accepted", func(ctx context.Context, event OutboxEvent) error {
		calls = append(calls, "progress")
		return nil
	})

	if err := registry.Handle(context.Background(), OutboxEvent{Name: "practice.flag_accepted"}); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if len(calls) != 2 || calls[0] != "notification" || calls[1] != "progress" {
		t.Fatalf("handler calls = %+v, want notification then progress", calls)
	}
}

type stubOutboxStore struct {
	claimed       []PendingOutboxEvent
	dispatchedIDs []int64
	failedIDs     []int64
}

func (s *stubOutboxStore) ClaimDue(context.Context, string, time.Time, time.Duration, int) ([]PendingOutboxEvent, error) {
	return s.claimed, nil
}

func (s *stubOutboxStore) MarkDispatched(_ context.Context, id int64, _ string, _ time.Time) error {
	s.dispatchedIDs = append(s.dispatchedIDs, id)
	return nil
}

func (s *stubOutboxStore) MarkFailed(_ context.Context, id int64, _ error, _ time.Time) error {
	s.failedIDs = append(s.failedIDs, id)
	return nil
}

type stubStreamPublisher struct {
	published []OutboxEvent
	messageID string
	err       error
}

func (s *stubStreamPublisher) Publish(_ context.Context, event OutboxEvent) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	s.published = append(s.published, event)
	if s.messageID == "" {
		s.messageID = "1710000000000-0"
	}
	return s.messageID, nil
}
