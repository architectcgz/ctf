package composition

import (
	"context"
	"errors"
	"testing"

	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type stubOpsRuntimeCountQuery struct {
	ctx   context.Context
	count int64
	err   error
}

func (s *stubOpsRuntimeCountQuery) CountRunning(ctx context.Context) (int64, error) {
	s.ctx = ctx
	return s.count, s.err
}

type stubOpsRuntimeStatsReader struct {
	ctx   context.Context
	stats []runtimeports.ManagedContainerStat
	err   error
}

func (s *stubOpsRuntimeStatsReader) ListManagedContainerStats(ctx context.Context) ([]runtimeports.ManagedContainerStat, error) {
	s.ctx = ctx
	return s.stats, s.err
}

func TestOpsRuntimeQueryAdapterHandlesNilDependency(t *testing.T) {
	t.Parallel()

	query := newOpsRuntimeQueryAdapter(nil)
	count, err := query.CountRunning(context.Background())
	if err != nil {
		t.Fatalf("CountRunning returned error: %v", err)
	}
	if count != 0 {
		t.Fatalf("CountRunning = %d, want 0", count)
	}
}

func TestOpsRuntimeQueryAdapterDelegatesCountRunning(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), testContextKey("runtime-query"), "marker")
	repo := &stubOpsRuntimeCountQuery{count: 7}

	count, err := newOpsRuntimeQueryAdapter(repo).CountRunning(ctx)
	if err != nil {
		t.Fatalf("CountRunning returned error: %v", err)
	}
	if count != 7 {
		t.Fatalf("CountRunning = %d, want 7", count)
	}
	if repo.ctx != ctx {
		t.Fatal("CountRunning did not receive caller context")
	}
}

func TestOpsRuntimeQueryAdapterPropagatesErrors(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("count failed")
	repo := &stubOpsRuntimeCountQuery{err: wantErr}

	_, err := newOpsRuntimeQueryAdapter(repo).CountRunning(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("CountRunning error = %v, want %v", err, wantErr)
	}
}

func TestOpsRuntimeStatsProviderAdapterHandlesNilDependency(t *testing.T) {
	t.Parallel()

	stats, err := newOpsRuntimeStatsProviderAdapter(nil).ListManagedContainerStats(context.Background())
	if err != nil {
		t.Fatalf("ListManagedContainerStats returned error: %v", err)
	}
	if len(stats) != 0 {
		t.Fatalf("ListManagedContainerStats returned %d items, want 0", len(stats))
	}
}

func TestOpsRuntimeStatsProviderAdapterMapsManagedContainerStats(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), testContextKey("runtime-stats"), "marker")
	reader := &stubOpsRuntimeStatsReader{
		stats: []runtimeports.ManagedContainerStat{
			{
				ContainerID:   "container-1",
				ContainerName: "challenge-runtime-1",
				CPUPercent:    12.5,
				MemoryPercent: 34.5,
				MemoryUsage:   1024,
				MemoryLimit:   4096,
			},
		},
	}

	stats, err := newOpsRuntimeStatsProviderAdapter(reader).ListManagedContainerStats(ctx)
	if err != nil {
		t.Fatalf("ListManagedContainerStats returned error: %v", err)
	}
	if reader.ctx != ctx {
		t.Fatal("ListManagedContainerStats did not receive caller context")
	}
	if len(stats) != 1 {
		t.Fatalf("ListManagedContainerStats returned %d items, want 1", len(stats))
	}

	got := stats[0]
	if got.ContainerID != "container-1" ||
		got.ContainerName != "challenge-runtime-1" ||
		got.CPUPercent != 12.5 ||
		got.MemoryPercent != 34.5 ||
		got.MemoryUsage != 1024 ||
		got.MemoryLimit != 4096 {
		t.Fatalf("mapped stat = %#v", got)
	}
}

func TestOpsRuntimeStatsProviderAdapterPropagatesErrors(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("stats failed")
	reader := &stubOpsRuntimeStatsReader{err: wantErr}

	_, err := newOpsRuntimeStatsProviderAdapter(reader).ListManagedContainerStats(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("ListManagedContainerStats error = %v, want %v", err, wantErr)
	}
}

type testContextKey string
