package commands

import (
	"context"
	"testing"

	platformevents "ctf-platform/internal/platform/events"
)

type stubContestEventBus struct {
	publishFn func(ctx context.Context, evt platformevents.Event) error
}

func (s *stubContestEventBus) Subscribe(string, platformevents.Handler) {}

func (s *stubContestEventBus) Publish(ctx context.Context, evt platformevents.Event) error {
	if s.publishFn != nil {
		return s.publishFn(ctx, evt)
	}
	return nil
}

func TestPublishContestWeakEventDoesNotCreateBackgroundContext(t *testing.T) {
	t.Parallel()

	publishCalled := false
	publishContestWeakEvent(nil, &stubContestEventBus{
		publishFn: func(ctx context.Context, evt platformevents.Event) error {
			publishCalled = true
			if ctx != nil {
				t.Fatalf("expected publish ctx to stay nil, got %v", ctx)
			}
			return nil
		},
	}, platformevents.Event{Name: "contest.test"})

	if !publishCalled {
		t.Fatal("expected event to be published")
	}
}

func TestAWDServicePublishWeakEventDoesNotCreateBackgroundContext(t *testing.T) {
	t.Parallel()

	publishCalled := false
	service := (&AWDService{}).SetEventBus(&stubContestEventBus{
		publishFn: func(ctx context.Context, evt platformevents.Event) error {
			publishCalled = true
			if ctx != nil {
				t.Fatalf("expected publish ctx to stay nil, got %v", ctx)
			}
			return nil
		},
	})

	service.publishWeakEvent(nil, platformevents.Event{Name: "contest.test"})
	if !publishCalled {
		t.Fatal("expected event to be published")
	}
}

func TestWithAWDPreviewRequesterDoesNotCreateBackgroundContext(t *testing.T) {
	t.Parallel()

	if ctx := WithAWDPreviewRequester(nil, 42); ctx != nil {
		t.Fatalf("expected nil context to stay nil, got %v", ctx)
	}
}

func TestWithAWDReadinessGateTraceDoesNotCreateBackgroundContext(t *testing.T) {
	t.Parallel()

	ctx, trace := WithAWDReadinessGateTrace(nil)
	if ctx != nil {
		t.Fatalf("expected nil context to stay nil, got %v", ctx)
	}
	if trace != nil {
		t.Fatalf("expected nil trace without context, got %+v", trace)
	}
}
