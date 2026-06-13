package commands

import (
	"context"
	"errors"
	"testing"
	"time"

	"ctf-platform/internal/platform/requestctx"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLogTopologyStageIncludesRequestID(t *testing.T) {
	t.Parallel()

	core, observed := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	service := NewProvisioningService(nil, nil, nil, logger)
	ctx := requestctx.WithRequestID(context.Background(), "req-runtime-log-1")

	service.logTopologyStage(ctx, 10*time.Millisecond, errors.New("boom"), topologyStageContext{
		stage:      "create_container",
		instanceID: 42,
	})

	if observed.Len() != 1 {
		t.Fatalf("expected 1 log entry, got %d", observed.Len())
	}
	fields := observed.All()[0].ContextMap()
	if got := fields["request_id"]; got != "req-runtime-log-1" {
		t.Fatalf("request_id = %v, want req-runtime-log-1", got)
	}
}
