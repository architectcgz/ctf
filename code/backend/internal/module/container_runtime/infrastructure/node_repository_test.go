package infrastructure

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	runtimeentity "ctf-platform/internal/module/container_runtime/entity"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
)

func newRuntimeNodeRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&runtimeentity.RuntimeNode{}); err != nil {
		t.Fatalf("migrate runtime node: %v", err)
	}
	return db
}

func TestRuntimeNodeRepositoryHeartbeatUpdatesHealthAndLastSeen(t *testing.T) {
	t.Parallel()

	db := newRuntimeNodeRepositoryTestDB(t)
	repo := NewRuntimeNodeRepository(db)
	node, err := repo.EnsureDefaultNode(context.Background(), runtimecontracts.RuntimeNodeBootstrapSpec{
		Name:        "agent-a",
		Endpoint:    "grpc://agent-a",
		Schedulable: true,
	})
	if err != nil {
		t.Fatalf("EnsureDefaultNode() error = %v", err)
	}

	seenAt := time.Date(2026, 6, 12, 8, 15, 0, 0, time.FixedZone("CST", 8*60*60))
	updated, err := repo.MarkNodeHeartbeat(context.Background(), node.ID, runtimeentity.RuntimeNodeHealthDegraded, `{"containers":2}`, seenAt)
	if err != nil {
		t.Fatalf("MarkNodeHeartbeat() error = %v", err)
	}
	if updated == nil || updated.LastSeenAt == nil {
		t.Fatalf("expected heartbeat to return last_seen_at, got %+v", updated)
	}
	if updated.HealthStatus != runtimeentity.RuntimeNodeHealthDegraded {
		t.Fatalf("health_status = %q, want degraded", updated.HealthStatus)
	}
	if updated.CapacitySnapshot != `{"containers":2}` {
		t.Fatalf("capacity_snapshot = %q", updated.CapacitySnapshot)
	}
	if !updated.LastSeenAt.Equal(seenAt.UTC()) {
		t.Fatalf("last_seen_at = %s, want %s", updated.LastSeenAt, seenAt.UTC())
	}

	var stored runtimeentity.RuntimeNode
	if err := db.First(&stored, node.ID).Error; err != nil {
		t.Fatalf("load stored node: %v", err)
	}
	if stored.LastSeenAt == nil || !stored.LastSeenAt.Equal(seenAt.UTC()) {
		t.Fatalf("stored last_seen_at = %+v, want %s", stored.LastSeenAt, seenAt.UTC())
	}
}

func TestRuntimeNodeRepositoryEnsureDefaultNodePersistsAccessHosts(t *testing.T) {
	t.Parallel()

	db := newRuntimeNodeRepositoryTestDB(t)
	repo := NewRuntimeNodeRepository(db)
	node, err := repo.EnsureDefaultNode(context.Background(), runtimecontracts.RuntimeNodeBootstrapSpec{
		Name:        "agent-a",
		Endpoint:    "grpc://agent-a",
		PublicHost:  "public-a.ctf.local",
		AccessHost:  "access-a.internal",
		Schedulable: true,
	})
	if err != nil {
		t.Fatalf("EnsureDefaultNode() create error = %v", err)
	}
	if node.PublicHost != "public-a.ctf.local" || node.AccessHost != "access-a.internal" {
		t.Fatalf("created hosts = public %q access %q", node.PublicHost, node.AccessHost)
	}

	updated, err := repo.EnsureDefaultNode(context.Background(), runtimecontracts.RuntimeNodeBootstrapSpec{
		Name:        "agent-a",
		Endpoint:    "grpc://agent-a",
		PublicHost:  "public-b.ctf.local",
		AccessHost:  "access-b.internal",
		Schedulable: true,
	})
	if err != nil {
		t.Fatalf("EnsureDefaultNode() update error = %v", err)
	}
	if updated.ID != node.ID {
		t.Fatalf("updated node id = %d, want %d", updated.ID, node.ID)
	}
	if updated.PublicHost != "public-b.ctf.local" || updated.AccessHost != "access-b.internal" {
		t.Fatalf("updated hosts = public %q access %q", updated.PublicHost, updated.AccessHost)
	}

	var stored runtimeentity.RuntimeNode
	if err := db.First(&stored, node.ID).Error; err != nil {
		t.Fatalf("load stored node: %v", err)
	}
	if stored.PublicHost != "public-b.ctf.local" || stored.AccessHost != "access-b.internal" {
		t.Fatalf("stored hosts = public %q access %q", stored.PublicHost, stored.AccessHost)
	}
}

func TestRuntimeNodeRepositorySelectsOnlyFreshReadyOrDegradedNodes(t *testing.T) {
	t.Parallel()

	db := newRuntimeNodeRepositoryTestDB(t)
	repo := NewRuntimeNodeRepository(db)
	now := time.Date(2026, 6, 12, 9, 0, 0, 0, time.UTC)
	fresh := now.Add(-10 * time.Second)
	stale := now.Add(-2 * time.Minute)
	nodes := []runtimeentity.RuntimeNode{
		runtimeNodeTestRow("ready-fresh", runtimeentity.RuntimeNodeHealthReady, true, &fresh),
		runtimeNodeTestRow("degraded-fresh", runtimeentity.RuntimeNodeHealthDegraded, true, &fresh),
		runtimeNodeTestRow("offline-fresh", runtimeentity.RuntimeNodeHealthOffline, true, &fresh),
		runtimeNodeTestRow("unknown-fresh", runtimeentity.RuntimeNodeHealthUnknown, true, &fresh),
		runtimeNodeTestRow("ready-stale", runtimeentity.RuntimeNodeHealthReady, true, &stale),
		runtimeNodeTestRow("ready-unschedulable", runtimeentity.RuntimeNodeHealthReady, false, &fresh),
		runtimeNodeTestRow("ready-never-seen", runtimeentity.RuntimeNodeHealthReady, true, nil),
	}
	if err := db.Select("*").Create(&nodes).Error; err != nil {
		t.Fatalf("seed runtime nodes: %v", err)
	}
	if err := db.Model(&runtimeentity.RuntimeNode{}).
		Where("name = ?", "ready-unschedulable").
		Update("schedulable", false).Error; err != nil {
		t.Fatalf("mark ready-unschedulable node: %v", err)
	}

	selected, err := repo.ListSchedulableHealthyNodes(context.Background(), 30*time.Second, now)
	if err != nil {
		t.Fatalf("ListSchedulableHealthyNodes() error = %v", err)
	}
	if got := runtimeNodeNames(selected); len(got) != 2 || got[0] != "ready-fresh" || got[1] != "degraded-fresh" {
		t.Fatalf("selected nodes = %v, want [ready-fresh degraded-fresh]", got)
	}
}

func TestDefaultRuntimeNodeSelectorFallsBackWhenConfiguredNodeIsOffline(t *testing.T) {
	t.Parallel()

	db := newRuntimeNodeRepositoryTestDB(t)
	repo := NewRuntimeNodeRepository(db)
	now := time.Now().UTC()
	nodes := []runtimeentity.RuntimeNode{
		runtimeNodeTestRow("node-a", runtimeentity.RuntimeNodeHealthReady, true, &now),
		runtimeNodeTestRow("node-b", runtimeentity.RuntimeNodeHealthOffline, true, &now),
	}
	if err := db.Create(&nodes).Error; err != nil {
		t.Fatalf("seed runtime nodes: %v", err)
	}

	selector := NewDefaultRuntimeNodeSelector(repo, "node-b", 30*time.Second)
	binding, err := selector.SelectDefaultNode(context.Background())
	if err != nil {
		t.Fatalf("SelectDefaultNode() error = %v", err)
	}
	if binding == nil || binding.NodeName != "node-a" {
		t.Fatalf("expected selector to fall back to node-a, got %+v", binding)
	}
}

func TestRuntimeNodeRepositoryFindHealthyByIDRejectsOfflineNode(t *testing.T) {
	t.Parallel()

	db := newRuntimeNodeRepositoryTestDB(t)
	repo := NewRuntimeNodeRepository(db)
	now := time.Now().UTC()
	node := runtimeNodeTestRow("offline-node", runtimeentity.RuntimeNodeHealthOffline, true, &now)
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("seed runtime node: %v", err)
	}

	_, err := repo.FindHealthyByID(context.Background(), node.ID, 30*time.Second, now)
	if !errors.Is(err, runtimeports.ErrRuntimeNodeUnavailable) {
		t.Fatalf("FindHealthyByID() error = %v, want ErrRuntimeNodeUnavailable", err)
	}
}

func TestRuntimeNodeRepositoryFindHealthyByIDAllowsUnschedulableFreshNode(t *testing.T) {
	t.Parallel()

	db := newRuntimeNodeRepositoryTestDB(t)
	repo := NewRuntimeNodeRepository(db)
	now := time.Date(2026, 6, 12, 12, 0, 0, 0, time.UTC)
	node := runtimeNodeTestRow("cordoned-node", runtimeentity.RuntimeNodeHealthReady, false, &now)
	if err := db.Create(&node).Error; err != nil {
		t.Fatalf("seed runtime node: %v", err)
	}
	if err := db.Model(&runtimeentity.RuntimeNode{}).
		Where("id = ?", node.ID).
		Update("schedulable", false).Error; err != nil {
		t.Fatalf("cordon runtime node: %v", err)
	}

	_, err := repo.ListSchedulableHealthyNodes(context.Background(), 30*time.Second, now)
	if !errors.Is(err, runtimeports.ErrRuntimeNodeUnavailable) {
		t.Fatalf("ListSchedulableHealthyNodes() error = %v, want ErrRuntimeNodeUnavailable", err)
	}

	selected, err := repo.FindHealthyByID(context.Background(), node.ID, 30*time.Second, now)
	if err != nil {
		t.Fatalf("FindHealthyByID() error = %v", err)
	}
	if selected == nil || selected.ID != node.ID {
		t.Fatalf("FindHealthyByID() = %+v, want node %d", selected, node.ID)
	}
}

func TestRuntimeNodeRepositoryListHealthCheckNodesIncludesUnschedulableNodes(t *testing.T) {
	t.Parallel()

	db := newRuntimeNodeRepositoryTestDB(t)
	repo := NewRuntimeNodeRepository(db)
	now := time.Date(2026, 6, 12, 12, 5, 0, 0, time.UTC)
	nodes := []runtimeentity.RuntimeNode{
		runtimeNodeTestRow("schedulable-node", runtimeentity.RuntimeNodeHealthReady, true, &now),
		runtimeNodeTestRow("cordoned-node", runtimeentity.RuntimeNodeHealthReady, false, &now),
	}
	if err := db.Create(&nodes).Error; err != nil {
		t.Fatalf("seed runtime nodes: %v", err)
	}
	if err := db.Model(&runtimeentity.RuntimeNode{}).
		Where("name = ?", "cordoned-node").
		Update("schedulable", false).Error; err != nil {
		t.Fatalf("cordon runtime node: %v", err)
	}

	selected, err := repo.ListHealthCheckNodes(context.Background())
	if err != nil {
		t.Fatalf("ListHealthCheckNodes() error = %v", err)
	}
	if got := runtimeNodeNames(selected); len(got) != 2 || got[0] != "schedulable-node" || got[1] != "cordoned-node" {
		t.Fatalf("health check nodes = %v, want [schedulable-node cordoned-node]", got)
	}
}

func runtimeNodeTestRow(name, health string, schedulable bool, lastSeenAt *time.Time) runtimeentity.RuntimeNode {
	return runtimeentity.RuntimeNode{
		Name:             name,
		Endpoint:         "local://docker",
		Schedulable:      schedulable,
		Labels:           "{}",
		HealthStatus:     health,
		CapacitySnapshot: "{}",
		LastSeenAt:       lastSeenAt,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
}

func runtimeNodeNames(nodes []runtimeentity.RuntimeNode) []string {
	names := make([]string, 0, len(nodes))
	for _, node := range nodes {
		names = append(names, node.Name)
	}
	return names
}
