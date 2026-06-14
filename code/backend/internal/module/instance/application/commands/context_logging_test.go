package commands_test

import (
	"context"
	"testing"

	"ctf-platform/internal/config"
	instancecmd "ctf-platform/internal/module/instance/application/commands"
	instanceentity "ctf-platform/internal/module/instance/entity"
	"ctf-platform/internal/platform/requestctx"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestRuntimeMaintenanceServiceOfflineLogIncludesRequestID(t *testing.T) {
	t.Parallel()

	core, observed := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	service := instancecmd.NewInstanceMaintenanceService(&maintenanceTestRepository{
		findExpiredFn: func(ctx context.Context) ([]*instanceentity.Instance, error) {
			return []*instanceentity.Instance{{
				ID:     7,
				Status: "running",
			}}, nil
		},
	}, nil, nil, &config.ContainerConfig{}, logger)

	ctx := requestctx.WithRequestID(context.Background(), "req-instance-log-1")
	if err := service.CleanExpiredInstances(ctx); err != nil {
		t.Fatalf("CleanExpiredInstances() error = %v", err)
	}

	entries := observed.All()
	if len(entries) == 0 {
		t.Fatal("expected maintenance warning log entry")
	}
	fields := entries[len(entries)-1].ContextMap()
	if got := fields["request_id"]; got != "req-instance-log-1" {
		t.Fatalf("request_id = %v, want req-instance-log-1", got)
	}
}
