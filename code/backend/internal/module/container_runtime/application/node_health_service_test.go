package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	runtimeentity "ctf-platform/internal/module/container_runtime/entity"
)

func TestNodeHealthServiceEvaluateOnceRecordsHeartbeatAndCapacitySnapshot(t *testing.T) {
	t.Parallel()

	seenAt := time.Date(2026, 6, 12, 10, 0, 0, 0, time.UTC)
	repo := &nodeHealthTestRepository{
		nodes: []runtimeentity.RuntimeNode{
			nodeHealthTestNode(11, "node-a", runtimeentity.RuntimeNodeHealthUnknown, true, nil),
		},
	}
	probe := &nodeHealthTestProbe{
		statsByNodeID: map[int64][]ManagedContainerStat{
			11: {
				{ContainerID: "container-a", CPUPercent: 12.5, MemoryUsage: 1024, MemoryLimit: 4096},
				{ContainerID: "container-b", CPUPercent: 7.5, MemoryUsage: 2048, MemoryLimit: 4096},
			},
		},
	}
	service := NewNodeHealthService(repo, probe, NodeHealthOptions{
		ProbeTimeout:     time.Second,
		StaleAfter:       time.Minute,
		FailureThreshold: 1,
	}, zap.NewNop()).SetNow(func() time.Time { return seenAt })

	if err := service.EvaluateOnce(context.Background()); err != nil {
		t.Fatalf("EvaluateOnce() error = %v", err)
	}
	if len(repo.heartbeats) != 1 {
		t.Fatalf("expected one heartbeat, got %+v", repo.heartbeats)
	}
	heartbeat := repo.heartbeats[0]
	if heartbeat.nodeID != 11 || heartbeat.health != runtimeentity.RuntimeNodeHealthReady || !heartbeat.seenAt.Equal(seenAt) {
		t.Fatalf("unexpected heartbeat: %+v", heartbeat)
	}
	for _, snippet := range []string{
		`"containers":2`,
		`"memory_usage":3072`,
		`"memory_limit":8192`,
		`"max_cpu_percent":12.5`,
	} {
		if !strings.Contains(heartbeat.snapshot, snippet) {
			t.Fatalf("capacity snapshot should contain %s, got %s", snippet, heartbeat.snapshot)
		}
	}
}

func TestNodeHealthServiceEvaluatesUnschedulableNodes(t *testing.T) {
	t.Parallel()

	seenAt := time.Date(2026, 6, 12, 10, 10, 0, 0, time.UTC)
	repo := &nodeHealthTestRepository{
		nodes: []runtimeentity.RuntimeNode{
			nodeHealthTestNode(12, "cordoned-node", runtimeentity.RuntimeNodeHealthReady, false, nil),
		},
	}
	service := NewNodeHealthService(repo, &nodeHealthTestProbe{}, NodeHealthOptions{
		ProbeTimeout:     time.Second,
		StaleAfter:       time.Minute,
		FailureThreshold: 1,
	}, zap.NewNop()).SetNow(func() time.Time { return seenAt })

	if err := service.EvaluateOnce(context.Background()); err != nil {
		t.Fatalf("EvaluateOnce() error = %v", err)
	}
	if len(repo.heartbeats) != 1 || repo.heartbeats[0].nodeID != 12 {
		t.Fatalf("expected heartbeat for unschedulable node 12, got %+v", repo.heartbeats)
	}
}

func TestNodeHealthServiceMarksNodeOfflineAfterProbeFailureThreshold(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 12, 10, 30, 0, 0, time.UTC)
	lastSeenAt := now.Add(-10 * time.Second)
	repo := &nodeHealthTestRepository{
		nodes: []runtimeentity.RuntimeNode{
			nodeHealthTestNode(21, "node-b", runtimeentity.RuntimeNodeHealthReady, true, &lastSeenAt),
		},
	}
	service := NewNodeHealthService(repo, &nodeHealthTestProbe{errByNodeID: map[int64]error{21: errors.New("agent unavailable")}}, NodeHealthOptions{
		ProbeTimeout:     time.Second,
		StaleAfter:       time.Minute,
		FailureThreshold: 2,
	}, zap.NewNop()).SetNow(func() time.Time { return now })

	if err := service.EvaluateOnce(context.Background()); err != nil {
		t.Fatalf("first EvaluateOnce() error = %v", err)
	}
	if len(repo.offlineIDs) != 0 {
		t.Fatalf("expected first probe failure below threshold to stay online, got offline %v", repo.offlineIDs)
	}
	if err := service.EvaluateOnce(context.Background()); err != nil {
		t.Fatalf("second EvaluateOnce() error = %v", err)
	}
	if len(repo.offlineIDs) != 1 || repo.offlineIDs[0] != 21 {
		t.Fatalf("expected node 21 offline after threshold, got %v", repo.offlineIDs)
	}
}

func TestNodeHealthServiceMarksStaleNodeOfflineOnProbeFailure(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 12, 11, 0, 0, 0, time.UTC)
	lastSeenAt := now.Add(-3 * time.Minute)
	repo := &nodeHealthTestRepository{
		nodes: []runtimeentity.RuntimeNode{
			nodeHealthTestNode(31, "node-c", runtimeentity.RuntimeNodeHealthReady, true, &lastSeenAt),
		},
	}
	service := NewNodeHealthService(repo, &nodeHealthTestProbe{errByNodeID: map[int64]error{31: errors.New("agent stale")}}, NodeHealthOptions{
		ProbeTimeout:     time.Second,
		StaleAfter:       time.Minute,
		FailureThreshold: 10,
	}, zap.NewNop()).SetNow(func() time.Time { return now })

	if err := service.EvaluateOnce(context.Background()); err != nil {
		t.Fatalf("EvaluateOnce() error = %v", err)
	}
	if len(repo.offlineIDs) != 1 || repo.offlineIDs[0] != 31 {
		t.Fatalf("expected stale failed node 31 offline, got %v", repo.offlineIDs)
	}
}

func TestNodeHealthServiceCallsOfflineHandlerAfterMarkingNodeOffline(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 12, 11, 30, 0, 0, time.UTC)
	lastSeenAt := now.Add(-10 * time.Second)
	repo := &nodeHealthTestRepository{
		nodes: []runtimeentity.RuntimeNode{
			nodeHealthTestNode(36, "node-offline-callback", runtimeentity.RuntimeNodeHealthReady, true, &lastSeenAt),
		},
	}
	offlineIDs := make([]int64, 0, 1)
	service := NewNodeHealthService(repo, &nodeHealthTestProbe{errByNodeID: map[int64]error{36: errors.New("agent unavailable")}}, NodeHealthOptions{
		ProbeTimeout:     time.Second,
		StaleAfter:       time.Minute,
		FailureThreshold: 1,
	}, zap.NewNop()).
		SetNow(func() time.Time { return now }).
		SetOfflineHandler(func(ctx context.Context, node runtimeentity.RuntimeNode) error {
			if ctx == nil {
				t.Fatal("expected offline handler to receive context")
			}
			offlineIDs = append(offlineIDs, node.ID)
			return nil
		})

	if err := service.EvaluateOnce(context.Background()); err != nil {
		t.Fatalf("EvaluateOnce() error = %v", err)
	}
	if len(repo.offlineIDs) != 1 || repo.offlineIDs[0] != 36 {
		t.Fatalf("expected node 36 to be marked offline, got %v", repo.offlineIDs)
	}
	if len(offlineIDs) != 1 || offlineIDs[0] != 36 {
		t.Fatalf("expected offline handler for node 36, got %v", offlineIDs)
	}
}

func TestNodeHealthServiceDoesNotRepeatOfflineHandlerAfterSuccessfulOfflineHandling(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 12, 11, 45, 0, 0, time.UTC)
	lastSeenAt := now.Add(-10 * time.Second)
	repo := &nodeHealthTestRepository{
		nodes: []runtimeentity.RuntimeNode{
			nodeHealthTestNode(37, "node-repeated-offline", runtimeentity.RuntimeNodeHealthReady, true, &lastSeenAt),
		},
	}
	offlineCalls := 0
	service := NewNodeHealthService(repo, &nodeHealthTestProbe{errByNodeID: map[int64]error{37: errors.New("agent still unavailable")}}, NodeHealthOptions{
		ProbeTimeout:     time.Second,
		StaleAfter:       time.Minute,
		FailureThreshold: 1,
	}, zap.NewNop()).
		SetNow(func() time.Time { return now }).
		SetOfflineHandler(func(ctx context.Context, node runtimeentity.RuntimeNode) error {
			offlineCalls++
			return nil
		})

	if err := service.EvaluateOnce(context.Background()); err != nil {
		t.Fatalf("first EvaluateOnce() error = %v", err)
	}
	if err := service.EvaluateOnce(context.Background()); err != nil {
		t.Fatalf("second EvaluateOnce() error = %v", err)
	}
	if offlineCalls != 1 {
		t.Fatalf("offline handler calls = %d, want 1 after successful offline handling", offlineCalls)
	}
	if len(repo.offlineIDs) != 2 || repo.offlineIDs[0] != 37 || repo.offlineIDs[1] != 37 {
		t.Fatalf("expected offline state to stay refreshed, got %v", repo.offlineIDs)
	}
}

func TestNodeHealthServiceRetriesOfflineHandlerWhenFirstAttemptFails(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 12, 11, 48, 0, 0, time.UTC)
	lastSeenAt := now.Add(-10 * time.Second)
	repo := &nodeHealthTestRepository{
		nodes: []runtimeentity.RuntimeNode{
			nodeHealthTestNode(39, "node-handler-retry", runtimeentity.RuntimeNodeHealthReady, true, &lastSeenAt),
		},
	}
	offlineCalls := 0
	service := NewNodeHealthService(repo, &nodeHealthTestProbe{errByNodeID: map[int64]error{39: errors.New("agent unavailable")}}, NodeHealthOptions{
		ProbeTimeout:     time.Second,
		StaleAfter:       time.Minute,
		FailureThreshold: 1,
	}, zap.NewNop()).
		SetNow(func() time.Time { return now }).
		SetOfflineHandler(func(ctx context.Context, node runtimeentity.RuntimeNode) error {
			offlineCalls++
			if offlineCalls == 1 {
				return errors.New("temporary requeue failure")
			}
			return nil
		})

	if err := service.EvaluateOnce(context.Background()); err != nil {
		t.Fatalf("first EvaluateOnce() error = %v", err)
	}
	if err := service.EvaluateOnce(context.Background()); err != nil {
		t.Fatalf("second EvaluateOnce() error = %v", err)
	}
	if offlineCalls != 2 {
		t.Fatalf("offline handler calls = %d, want retry after first handler failure", offlineCalls)
	}
}

func TestNodeHealthServiceDoesNotMarkNodeOfflineWhenParentContextCanceled(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 12, 11, 50, 0, 0, time.UTC)
	lastSeenAt := now.Add(-10 * time.Second)
	repo := &nodeHealthTestRepository{
		nodes: []runtimeentity.RuntimeNode{
			nodeHealthTestNode(38, "node-cancelled-probe", runtimeentity.RuntimeNodeHealthReady, true, &lastSeenAt),
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	service := NewNodeHealthService(repo, nodeHealthTestProbeFunc(func(probeCtx context.Context, node runtimeentity.RuntimeNode) ([]ManagedContainerStat, error) {
		cancel()
		return nil, probeCtx.Err()
	}), NodeHealthOptions{
		ProbeTimeout:     time.Second,
		StaleAfter:       time.Minute,
		FailureThreshold: 1,
	}, zap.NewNop()).
		SetNow(func() time.Time { return now }).
		SetOfflineHandler(func(ctx context.Context, node runtimeentity.RuntimeNode) error {
			t.Fatalf("offline handler should not run for parent context cancellation")
			return nil
		})

	if err := service.EvaluateOnce(ctx); err != nil {
		t.Fatalf("EvaluateOnce() error = %v", err)
	}
	if len(repo.offlineIDs) != 0 {
		t.Fatalf("expected no offline mark for parent context cancellation, got %v", repo.offlineIDs)
	}
}

func TestNodeHealthServiceRunStopsWhenContextIsCanceled(t *testing.T) {
	t.Parallel()

	repo := &nodeHealthTestRepository{
		nodes: []runtimeentity.RuntimeNode{
			nodeHealthTestNode(41, "node-d", runtimeentity.RuntimeNodeHealthReady, true, nil),
		},
	}
	service := NewNodeHealthService(repo, &nodeHealthTestProbe{}, NodeHealthOptions{
		PollInterval:     time.Millisecond,
		ProbeTimeout:     time.Second,
		StaleAfter:       time.Minute,
		FailureThreshold: 1,
	}, zap.NewNop())

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		service.Run(ctx)
	}()

	deadline := time.Now().Add(time.Second)
	for len(repo.heartbeats) == 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("expected Run() to stop after context cancellation")
	}
	if len(repo.heartbeats) == 0 {
		t.Fatal("expected Run() to evaluate at least once before cancellation")
	}
}

type nodeHealthTestRepository struct {
	nodes      []runtimeentity.RuntimeNode
	heartbeats []nodeHealthHeartbeatCall
	offlineIDs []int64
}

type nodeHealthHeartbeatCall struct {
	nodeID   int64
	health   string
	snapshot string
	seenAt   time.Time
}

func (r *nodeHealthTestRepository) ListHealthCheckNodes(context.Context) ([]runtimeentity.RuntimeNode, error) {
	return append([]runtimeentity.RuntimeNode(nil), r.nodes...), nil
}

func (r *nodeHealthTestRepository) MarkNodeHeartbeat(_ context.Context, nodeID int64, healthStatus, capacitySnapshot string, seenAt time.Time) (*runtimeentity.RuntimeNode, error) {
	r.heartbeats = append(r.heartbeats, nodeHealthHeartbeatCall{
		nodeID:   nodeID,
		health:   healthStatus,
		snapshot: capacitySnapshot,
		seenAt:   seenAt,
	})
	for idx := range r.nodes {
		if r.nodes[idx].ID != nodeID {
			continue
		}
		r.nodes[idx].HealthStatus = healthStatus
		r.nodes[idx].CapacitySnapshot = capacitySnapshot
		r.nodes[idx].LastSeenAt = &seenAt
		return &r.nodes[idx], nil
	}
	return nil, nil
}

func (r *nodeHealthTestRepository) MarkNodeOffline(_ context.Context, nodeID int64, updatedAt time.Time) (*runtimeentity.RuntimeNode, error) {
	r.offlineIDs = append(r.offlineIDs, nodeID)
	for idx := range r.nodes {
		if r.nodes[idx].ID != nodeID {
			continue
		}
		r.nodes[idx].HealthStatus = runtimeentity.RuntimeNodeHealthOffline
		r.nodes[idx].UpdatedAt = updatedAt
		return &r.nodes[idx], nil
	}
	return nil, nil
}

type nodeHealthTestProbe struct {
	statsByNodeID map[int64][]ManagedContainerStat
	errByNodeID   map[int64]error
}

type nodeHealthTestProbeFunc func(context.Context, runtimeentity.RuntimeNode) ([]ManagedContainerStat, error)

func (f nodeHealthTestProbeFunc) ListManagedContainerStats(ctx context.Context, node runtimeentity.RuntimeNode) ([]ManagedContainerStat, error) {
	return f(ctx, node)
}

func (p *nodeHealthTestProbe) ListManagedContainerStats(_ context.Context, node runtimeentity.RuntimeNode) ([]ManagedContainerStat, error) {
	if p != nil && p.errByNodeID != nil && p.errByNodeID[node.ID] != nil {
		return nil, p.errByNodeID[node.ID]
	}
	if p == nil || p.statsByNodeID == nil {
		return []ManagedContainerStat{}, nil
	}
	return append([]ManagedContainerStat(nil), p.statsByNodeID[node.ID]...), nil
}

func nodeHealthTestNode(id int64, name, health string, schedulable bool, lastSeenAt *time.Time) runtimeentity.RuntimeNode {
	return runtimeentity.RuntimeNode{
		ID:               id,
		Name:             name,
		Endpoint:         "local://docker",
		Schedulable:      schedulable,
		HealthStatus:     health,
		CapacitySnapshot: "{}",
		LastSeenAt:       lastSeenAt,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
}
