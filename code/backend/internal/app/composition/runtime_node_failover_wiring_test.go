package composition

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	runtimeapp "ctf-platform/internal/module/container_runtime/application"
	runtimeentity "ctf-platform/internal/module/container_runtime/entity"
)

func TestWireRuntimeNodeFailoverRequeuesThenReconcilesDesiredAWD(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 12, 12, 0, 0, 0, time.UTC)
	lastSeenAt := now.Add(-10 * time.Second)
	repo := &runtimeNodeFailoverHealthRepo{
		node: runtimeentity.RuntimeNode{
			ID:               4401,
			Name:             "node-offline",
			Endpoint:         "local://docker",
			Schedulable:      true,
			HealthStatus:     runtimeentity.RuntimeNodeHealthReady,
			CapacitySnapshot: "{}",
			LastSeenAt:       &lastSeenAt,
			CreatedAt:        now.Add(-time.Hour),
			UpdatedAt:        now.Add(-time.Hour),
		},
	}
	health := runtimeapp.NewNodeHealthService(repo, runtimeNodeFailoverProbe{err: errors.New("agent offline")}, runtimeapp.NodeHealthOptions{
		ProbeTimeout:     time.Second,
		StaleAfter:       time.Minute,
		FailureThreshold: 1,
	}, zap.NewNop()).SetNow(func() time.Time { return now })

	runtime := &ContainerRuntimeModule{runtimeNodeHealth: health}
	calls := make([]string, 0, 2)
	instance := &InstanceModule{
		runtimeNodeOfflineHandler: runtimeNodeOfflineHandlerFunc(func(ctx context.Context, nodeID int64) error {
			if nodeID != 4401 {
				t.Fatalf("node id = %d, want 4401", nodeID)
			}
			calls = append(calls, "instance")
			return nil
		}),
	}
	practice := &PracticeModule{
		AWDDesiredRuntimeReconciler: desiredRuntimeReconcilerFunc(func(ctx context.Context) error {
			calls = append(calls, "desired")
			return nil
		}),
	}

	WireRuntimeNodeFailover(runtime, instance, practice)
	if err := health.EvaluateOnce(context.Background()); err != nil {
		t.Fatalf("EvaluateOnce() error = %v", err)
	}
	if got := strings.Join(calls, ","); got != "instance,desired" {
		t.Fatalf("call order = %s, want instance,desired", got)
	}
}

type runtimeNodeFailoverHealthRepo struct {
	node runtimeentity.RuntimeNode
}

func (r *runtimeNodeFailoverHealthRepo) ListHealthCheckNodes(context.Context) ([]runtimeentity.RuntimeNode, error) {
	return []runtimeentity.RuntimeNode{r.node}, nil
}

func (r *runtimeNodeFailoverHealthRepo) MarkNodeHeartbeat(_ context.Context, nodeID int64, healthStatus, capacitySnapshot string, seenAt time.Time) (*runtimeentity.RuntimeNode, error) {
	r.node.HealthStatus = healthStatus
	r.node.CapacitySnapshot = capacitySnapshot
	r.node.LastSeenAt = &seenAt
	r.node.UpdatedAt = seenAt
	return &r.node, nil
}

func (r *runtimeNodeFailoverHealthRepo) MarkNodeOffline(_ context.Context, nodeID int64, updatedAt time.Time) (*runtimeentity.RuntimeNode, error) {
	r.node.HealthStatus = runtimeentity.RuntimeNodeHealthOffline
	r.node.UpdatedAt = updatedAt
	return &r.node, nil
}

type runtimeNodeFailoverProbe struct {
	err error
}

func (p runtimeNodeFailoverProbe) ListManagedContainerStats(context.Context, runtimeentity.RuntimeNode) ([]runtimeapp.ManagedContainerStat, error) {
	return nil, p.err
}

type runtimeNodeOfflineHandlerFunc func(context.Context, int64) error

func (f runtimeNodeOfflineHandlerFunc) HandleRuntimeNodeOffline(ctx context.Context, nodeID int64) error {
	return f(ctx, nodeID)
}

type desiredRuntimeReconcilerFunc func(context.Context) error

func (f desiredRuntimeReconcilerFunc) ReconcileDesiredAWDInstances(ctx context.Context) error {
	return f(ctx)
}
