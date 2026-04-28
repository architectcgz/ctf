package events

import (
	"context"
	"testing"
)

func TestBusPublishDeliversToSubscriber(t *testing.T) {
	t.Parallel()

	bus := NewBus()
	got := make(chan Event, 1)

	bus.Subscribe("contest.created", func(ctx context.Context, evt Event) error {
		if ctx == nil {
			t.Fatal("expected publish context")
		}
		got <- evt
		return nil
	})

	want := Event{
		Name: "contest.created",
		Payload: map[string]any{
			"id": int64(42),
		},
	}
	if err := bus.Publish(context.Background(), want); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	select {
	case evt := <-got:
		if evt.Name != want.Name {
			t.Fatalf("event name = %q, want %q", evt.Name, want.Name)
		}
		payload, ok := evt.Payload.(map[string]any)
		if !ok {
			t.Fatalf("payload type = %T, want map[string]any", evt.Payload)
		}
		if payload["id"] != want.Payload.(map[string]any)["id"] {
			t.Fatalf("payload id = %v, want %v", payload["id"], want.Payload.(map[string]any)["id"])
		}
	default:
		t.Fatal("expected subscriber to receive published event")
	}
}

func TestBusPublishDoesNotCreateBackgroundContext(t *testing.T) {
	t.Parallel()

	bus := NewBus()
	called := false

	bus.Subscribe("practice.created", func(ctx context.Context, evt Event) error {
		called = true
		if ctx != nil {
			t.Fatalf("expected publish ctx to stay nil, got %v", ctx)
		}
		return nil
	})

	if err := bus.Publish(nil, Event{Name: "practice.created"}); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if !called {
		t.Fatal("expected subscriber to be called")
	}
}
