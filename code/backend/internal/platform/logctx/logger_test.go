package logctx

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"ctf-platform/internal/platform/requestctx"
)

func TestErrorAddsRequestIDFromContext(t *testing.T) {
	t.Parallel()

	core, observed := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	ctx := requestctx.WithRequestID(context.Background(), "req-logctx-1")

	Error(ctx, logger, "failed", zap.String("component", "auth"))

	if observed.Len() != 1 {
		t.Fatalf("expected 1 log entry, got %d", observed.Len())
	}
	fields := observed.All()[0].ContextMap()
	if got := fields["request_id"]; got != "req-logctx-1" {
		t.Fatalf("request_id = %v, want req-logctx-1", got)
	}
}

func TestWarnSkipsRequestIDWhenContextDoesNotHaveIt(t *testing.T) {
	t.Parallel()

	core, observed := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	Warn(context.Background(), logger, "warned")

	fields := observed.All()[0].ContextMap()
	if _, ok := fields["request_id"]; ok {
		t.Fatalf("unexpected request_id field: %v", fields["request_id"])
	}
}
