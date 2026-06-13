package requestctx

import (
	"context"
	"testing"
)

func TestWithRequestIDStoresAndLoadsRequestID(t *testing.T) {
	t.Parallel()

	ctx := WithRequestID(context.Background(), "req-platform-1")
	if got := RequestIDFromContext(ctx); got != "req-platform-1" {
		t.Fatalf("request id = %q, want req-platform-1", got)
	}
}

func TestWithRequestIDSkipsBlankValue(t *testing.T) {
	t.Parallel()

	ctx := WithRequestID(context.Background(), "   ")
	if got := RequestIDFromContext(ctx); got != "" {
		t.Fatalf("request id = %q, want empty", got)
	}
}
