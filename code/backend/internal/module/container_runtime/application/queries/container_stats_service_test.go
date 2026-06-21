package queries

import (
	"context"
	"testing"

	runtimeapp "ctf-platform/internal/module/container_runtime/application"
)

type stubManagedContainerStatsReader struct {
	calls int
	stats []runtimeapp.ManagedContainerStat
}

func (s *stubManagedContainerStatsReader) ListManagedContainerStats(_ context.Context) ([]runtimeapp.ManagedContainerStat, error) {
	s.calls++
	return append([]runtimeapp.ManagedContainerStat(nil), s.stats...), nil
}

func TestContainerStatsServiceListManagedContainerStatsSkipsNilReader(t *testing.T) {
	t.Parallel()

	service := NewContainerStatsService(nil)
	stats, err := service.ListManagedContainerStats(context.Background())
	if err != nil {
		t.Fatalf("ListManagedContainerStats() error = %v", err)
	}
	if len(stats) != 0 {
		t.Fatalf("ListManagedContainerStats() len = %d, want 0", len(stats))
	}
}

func TestContainerStatsServiceListManagedContainerStatsDelegatesToReader(t *testing.T) {
	t.Parallel()

	reader := &stubManagedContainerStatsReader{
		stats: []runtimeapp.ManagedContainerStat{
			{ContainerID: "abc", ContainerName: "ctf-instance", CPUPercent: 1.5},
		},
	}
	service := NewContainerStatsService(reader)

	stats, err := service.ListManagedContainerStats(context.Background())
	if err != nil {
		t.Fatalf("ListManagedContainerStats() error = %v", err)
	}
	if reader.calls != 1 {
		t.Fatalf("ListManagedContainerStats() calls = %d, want 1", reader.calls)
	}
	if len(stats) != 1 || stats[0].ContainerID != "abc" || stats[0].ContainerName != "ctf-instance" {
		t.Fatalf("ListManagedContainerStats() stats = %+v", stats)
	}
}
