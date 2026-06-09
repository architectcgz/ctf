package infrastructure

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instanceports "ctf-platform/internal/module/instance/ports"
)

type countRunningContextKey string

type instanceDestroyStatusRepository interface {
	MarkStopping(ctx context.Context, id int64) (bool, error)
}

var (
	_ instanceports.InstanceLookupRepository       = (*Repository)(nil)
	_ instanceports.InstanceUserLookupRepository   = (*Repository)(nil)
	_ instanceports.InstanceAccessRepository       = (*Repository)(nil)
	_ instanceports.UserVisibleInstanceRepository  = (*Repository)(nil)
	_ instanceports.TeacherInstanceQueryRepository = (*Repository)(nil)
	_ instanceports.InstanceExtendRepository       = (*Repository)(nil)
	_ instanceDestroyStatusRepository              = (*Repository)(nil)
)

func newInstanceRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&instancecontracts.Instance{}); err != nil {
		t.Fatalf("migrate instance: %v", err)
	}
	return db
}

func TestCountRunningInstancesCountsOnlyRunningInstances(t *testing.T) {
	t.Parallel()

	db := newInstanceRepositoryTestDB(t)
	repo := NewRepository(db)

	instances := []instancecontracts.Instance{
		{ID: 101, UserID: 9, ChallengeID: 21, Status: instancecontracts.InstanceStatusRunning, ExpiresAt: time.Now().Add(time.Hour)},
		{ID: 102, UserID: 9, ChallengeID: 22, Status: instancecontracts.InstanceStatusRunning, ExpiresAt: time.Now().Add(time.Hour)},
		{ID: 103, UserID: 9, ChallengeID: 23, Status: instancecontracts.InstanceStatusStopped, ExpiresAt: time.Now().Add(time.Hour)},
	}
	if err := db.Create(&instances).Error; err != nil {
		t.Fatalf("seed instances: %v", err)
	}

	count, err := repo.CountRunningInstances(context.Background())
	if err != nil {
		t.Fatalf("CountRunningInstances() error = %v", err)
	}
	if count != 2 {
		t.Fatalf("CountRunningInstances() count = %d, want 2", count)
	}
}

func TestCountRunningInstancesPropagatesContextToGORM(t *testing.T) {
	t.Parallel()

	db := newInstanceRepositoryTestDB(t)
	repo := NewRepository(db)

	ctxKey := countRunningContextKey("count-running")
	expectedCtxValue := "ctx-count-running"
	callbackName := "count_running_instances_context_check"
	callbackCalled := false
	if err := db.Callback().Query().Before("gorm:query").Register(callbackName, func(tx *gorm.DB) {
		if got := tx.Statement.Context.Value(ctxKey); got == expectedCtxValue {
			callbackCalled = true
		}
	}); err != nil {
		t.Fatalf("register query callback: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Callback().Query().Remove(callbackName)
	})

	if err := db.Create(&instancecontracts.Instance{
		ID:          101,
		UserID:      9,
		ChallengeID: 21,
		Status:      instancecontracts.InstanceStatusRunning,
		ExpiresAt:   time.Now().Add(time.Hour),
	}).Error; err != nil {
		t.Fatalf("seed running instance: %v", err)
	}

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	count, err := repo.CountRunningInstances(ctx)
	if err != nil {
		t.Fatalf("CountRunningInstances() error = %v", err)
	}
	if !callbackCalled {
		t.Fatal("expected gorm query callback to observe context-aware running instance count query")
	}
	if count != 1 {
		t.Fatalf("CountRunningInstances() count = %d, want 1", count)
	}
}

func TestMarkStoppingTransitionsActiveInstance(t *testing.T) {
	t.Parallel()

	db := newInstanceRepositoryTestDB(t)
	repo := NewRepository(db)

	instance := instancecontracts.Instance{
		ID:          11,
		UserID:      7,
		ChallengeID: 99,
		Status:      instancecontracts.InstanceStatusRunning,
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}

	changed, err := repo.MarkStopping(context.Background(), instance.ID)
	if err != nil {
		t.Fatalf("MarkStopping() error = %v", err)
	}
	if !changed {
		t.Fatal("expected active instance to transition to stopping")
	}

	var row struct {
		Status string `gorm:"column:status"`
	}
	if err := db.Table("instances").Select("status").Where("id = ?", instance.ID).Take(&row).Error; err != nil {
		t.Fatalf("load instance: %v", err)
	}
	if row.Status != instancecontracts.InstanceStatusStopping {
		t.Fatalf("instance status = %q, want %q", row.Status, instancecontracts.InstanceStatusStopping)
	}
}

func TestFindByUserAndChallengeReturnsActivePracticeInstance(t *testing.T) {
	t.Parallel()

	db := newInstanceRepositoryTestDB(t)
	repo := NewRepository(db)
	now := time.Now().UTC()

	instances := []instancecontracts.Instance{
		{ID: 1, UserID: 7, ChallengeID: 21, Status: instancecontracts.InstanceStatusStopped, ExpiresAt: now.Add(time.Hour)},
		{ID: 2, UserID: 7, ChallengeID: 21, Status: instancecontracts.InstanceStatusRunning, ExpiresAt: now.Add(time.Hour)},
		{ID: 3, UserID: 7, ChallengeID: 21, ContestID: int64Ptr(99), Status: instancecontracts.InstanceStatusRunning, ExpiresAt: now.Add(time.Hour)},
		{ID: 4, UserID: 8, ChallengeID: 21, Status: instancecontracts.InstanceStatusRunning, ExpiresAt: now.Add(time.Hour)},
	}
	if err := db.Create(&instances).Error; err != nil {
		t.Fatalf("seed instances: %v", err)
	}

	instance, err := repo.FindByUserAndChallenge(context.Background(), 7, 21)
	if err != nil {
		t.Fatalf("FindByUserAndChallenge() error = %v", err)
	}
	if instance == nil {
		t.Fatal("expected active practice instance")
	}
	if instance.ID != 2 {
		t.Fatalf("FindByUserAndChallenge() id = %d, want 2", instance.ID)
	}
}

func TestPersistProvisionedRuntimeUpdatesOnlyCreatingInstance(t *testing.T) {
	t.Parallel()

	db := newInstanceRepositoryTestDB(t)
	repo := NewRepository(db)
	now := time.Now().UTC()

	instances := []instancecontracts.Instance{
		{ID: 11, UserID: 7, ChallengeID: 21, Status: instancecontracts.InstanceStatusCreating, ExpiresAt: now.Add(time.Hour)},
		{ID: 12, UserID: 7, ChallengeID: 22, Status: instancecontracts.InstanceStatusRunning, ExpiresAt: now.Add(time.Hour)},
		{ID: 13, UserID: 7, ChallengeID: 23, Status: instancecontracts.InstanceStatusCreating, ExpiresAt: now.Add(time.Hour)},
	}
	if err := db.Create(&instances).Error; err != nil {
		t.Fatalf("seed instances: %v", err)
	}

	changed, err := repo.PersistProvisionedRuntime(context.Background(), &instancecontracts.Instance{
		ID:             11,
		HostPort:       30001,
		ContainerID:    "ctr-1",
		NetworkID:      "net-1",
		RuntimeDetails: `{"network_name":"runtime"}`,
		AccessURL:      "http://runtime.test",
		Status:         instancecontracts.InstanceStatusRunning,
	})
	if err != nil {
		t.Fatalf("PersistProvisionedRuntime() error = %v", err)
	}
	if !changed {
		t.Fatal("expected creating instance to be updated")
	}

	changed, err = repo.PersistProvisionedRuntime(context.Background(), &instancecontracts.Instance{
		ID:          12,
		HostPort:    30002,
		ContainerID: "ctr-2",
		Status:      instancecontracts.InstanceStatusRunning,
	})
	if err != nil {
		t.Fatalf("PersistProvisionedRuntime() on running instance error = %v", err)
	}
	if changed {
		t.Fatal("expected non-creating instance to stay unchanged")
	}

	if err := repo.UpdateRuntime(context.Background(), &instancecontracts.Instance{
		ID:             13,
		HostPort:       30003,
		ContainerID:    "ctr-3",
		NetworkID:      "net-3",
		RuntimeDetails: `{"network_name":"runtime-3"}`,
		AccessURL:      "http://runtime-3.test",
		Status:         instancecontracts.InstanceStatusRunning,
	}); err != nil {
		t.Fatalf("UpdateRuntime() error = %v", err)
	}

	var updated instancecontracts.Instance
	if err := db.Where("id = ?", 11).First(&updated).Error; err != nil {
		t.Fatalf("load updated instance: %v", err)
	}
	if updated.Status != instancecontracts.InstanceStatusRunning || updated.HostPort != 30001 || updated.ContainerID != "ctr-1" {
		t.Fatalf("updated instance mismatch: %+v", updated)
	}

	var wrapperUpdated instancecontracts.Instance
	if err := db.Where("id = ?", 13).First(&wrapperUpdated).Error; err != nil {
		t.Fatalf("load wrapper-updated instance: %v", err)
	}
	if wrapperUpdated.ContainerID != "ctr-3" || wrapperUpdated.Status != instancecontracts.InstanceStatusRunning {
		t.Fatalf("UpdateRuntime() did not persist expected fields: %+v", wrapperUpdated)
	}
}

func TestRefreshAndLifecycleQueriesStayOnInstanceOwner(t *testing.T) {
	t.Parallel()

	db := newInstanceRepositoryTestDB(t)
	repo := NewRepository(db)
	now := time.Now().UTC()

	instances := []instancecontracts.Instance{
		{ID: 21, UserID: 7, ChallengeID: 31, Status: instancecontracts.InstanceStatusRunning, ExpiresAt: now.Add(-time.Minute)},
		{ID: 22, UserID: 7, ChallengeID: 32, Status: instancecontracts.InstanceStatusCreating, ExpiresAt: now.Add(time.Hour), UpdatedAt: now.Add(-4 * time.Minute)},
		{ID: 23, UserID: 7, ChallengeID: 33, Status: instancecontracts.InstanceStatusRunning, ExpiresAt: now.Add(2 * time.Hour), UpdatedAt: now.Add(-3 * time.Minute)},
		{ID: 24, UserID: 7, ChallengeID: 34, Status: instancecontracts.InstanceStatusStopping, ExpiresAt: now.Add(time.Hour), UpdatedAt: now.Add(-5 * time.Minute)},
		{ID: 25, UserID: 7, ChallengeID: 35, Status: instancecontracts.InstanceStatusStopping, ExpiresAt: now.Add(time.Hour), UpdatedAt: now.Add(-time.Minute)},
		{ID: 26, UserID: 7, ChallengeID: 36, ContestID: int64Ptr(88), ServiceID: int64Ptr(1001), Status: instancecontracts.InstanceStatusRunning, ExpiresAt: now.Add(90 * time.Minute)},
		{ID: 27, UserID: 7, ChallengeID: 37, ContestID: int64Ptr(88), ServiceID: int64Ptr(1002), Status: instancecontracts.InstanceStatusStopped, ExpiresAt: now.Add(90 * time.Minute)},
	}
	if err := db.Create(&instances).Error; err != nil {
		t.Fatalf("seed instances: %v", err)
	}

	newExpiry := now.Add(4 * time.Hour)
	if err := repo.RefreshInstanceExpiry(context.Background(), 21, newExpiry); err != nil {
		t.Fatalf("RefreshInstanceExpiry() error = %v", err)
	}
	if err := repo.RefreshActiveAWDInstanceExpiryByContest(context.Background(), 88, now, now.Add(3*time.Hour)); err != nil {
		t.Fatalf("RefreshActiveAWDInstanceExpiryByContest() error = %v", err)
	}

	expired, err := repo.FindExpired(context.Background())
	if err != nil {
		t.Fatalf("FindExpired() error = %v", err)
	}
	if len(expired) != 0 {
		t.Fatalf("FindExpired() count = %d, want 0 after refresh", len(expired))
	}

	recoverable, err := repo.ListRecoverableActiveInstances(context.Background())
	if err != nil {
		t.Fatalf("ListRecoverableActiveInstances() error = %v", err)
	}
	if len(recoverable) != 4 {
		t.Fatalf("ListRecoverableActiveInstances() count = %d, want 4", len(recoverable))
	}
	if recoverable[0].ID != 22 || recoverable[1].ID != 23 || recoverable[2].ID != 21 || recoverable[3].ID != 26 {
		t.Fatalf("ListRecoverableActiveInstances() order = %+v, want [22 23 21 26]", extractInstanceIDs(recoverable))
	}

	stopping, err := repo.ListStoppingInstances(context.Background(), now.Add(-2*time.Minute), 10)
	if err != nil {
		t.Fatalf("ListStoppingInstances() error = %v", err)
	}
	if len(stopping) != 1 || stopping[0].ID != 24 {
		t.Fatalf("ListStoppingInstances() result = %+v, want only instance 24", stopping)
	}

	var refreshed instancecontracts.Instance
	if err := db.Where("id = ?", 26).First(&refreshed).Error; err != nil {
		t.Fatalf("load awd instance: %v", err)
	}
	if !refreshed.ExpiresAt.Equal(now.Add(3 * time.Hour)) {
		t.Fatalf("awd instance expiry = %s, want %s", refreshed.ExpiresAt, now.Add(3*time.Hour))
	}
}

func TestRequeuePendingTransitionAndCountStatus(t *testing.T) {
	t.Parallel()

	db := newInstanceRepositoryTestDB(t)
	repo := NewRepository(db)
	now := time.Now().UTC()

	instances := []instancecontracts.Instance{
		{ID: 31, UserID: 7, ChallengeID: 41, Status: instancecontracts.InstanceStatusCreating, HostPort: 30001, ContainerID: "ctr-1", NetworkID: "net-1", RuntimeDetails: "detail", AccessURL: "http://runtime-1", ExpiresAt: now.Add(time.Hour), CreatedAt: now.Add(-5 * time.Minute)},
		{ID: 32, UserID: 7, ChallengeID: 42, Status: instancecontracts.InstanceStatusPending, ExpiresAt: now.Add(time.Hour), CreatedAt: now.Add(-4 * time.Minute)},
		{ID: 33, UserID: 7, ChallengeID: 43, Status: instancecontracts.InstanceStatusPending, ExpiresAt: now.Add(time.Hour), CreatedAt: now.Add(-3 * time.Minute)},
		{ID: 34, UserID: 7, ChallengeID: 44, Status: instancecontracts.InstanceStatusRunning, ExpiresAt: now.Add(time.Hour)},
	}
	if err := db.Create(&instances).Error; err != nil {
		t.Fatalf("seed instances: %v", err)
	}

	requeued, err := repo.RequeueLostRuntime(context.Background(), 31)
	if err != nil {
		t.Fatalf("RequeueLostRuntime() error = %v", err)
	}
	if !requeued {
		t.Fatal("expected creating instance to be requeued")
	}

	pending, err := repo.ListPendingInstances(context.Background(), 2)
	if err != nil {
		t.Fatalf("ListPendingInstances() error = %v", err)
	}
	if len(pending) != 2 || pending[0].ID != 31 || pending[1].ID != 32 {
		t.Fatalf("ListPendingInstances() ids = %+v, want [31 32]", extractInstanceIDs(pending))
	}

	transitioned, err := repo.TryTransitionStatus(context.Background(), 32, instancecontracts.InstanceStatusPending, instancecontracts.InstanceStatusCreating)
	if err != nil {
		t.Fatalf("TryTransitionStatus() error = %v", err)
	}
	if !transitioned {
		t.Fatal("expected pending instance to transition")
	}

	transitioned, err = repo.TryTransitionStatus(context.Background(), 34, instancecontracts.InstanceStatusPending, instancecontracts.InstanceStatusCreating)
	if err != nil {
		t.Fatalf("TryTransitionStatus() on mismatched status error = %v", err)
	}
	if transitioned {
		t.Fatal("expected mismatched status transition to fail")
	}

	count, err := repo.CountInstancesByStatus(context.Background(), []string{
		instancecontracts.InstanceStatusPending,
		instancecontracts.InstanceStatusCreating,
	})
	if err != nil {
		t.Fatalf("CountInstancesByStatus() error = %v", err)
	}
	if count != 3 {
		t.Fatalf("CountInstancesByStatus() count = %d, want 3", count)
	}

	var requeuedRow instancecontracts.Instance
	if err := db.Where("id = ?", 31).First(&requeuedRow).Error; err != nil {
		t.Fatalf("load requeued instance: %v", err)
	}
	if requeuedRow.Status != instancecontracts.InstanceStatusPending || requeuedRow.ContainerID != "" || requeuedRow.NetworkID != "" || requeuedRow.RuntimeDetails != "" || requeuedRow.AccessURL != "" {
		t.Fatalf("requeued instance mismatch: %+v", requeuedRow)
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}

func extractInstanceIDs(instances []*instancecontracts.Instance) []int64 {
	ids := make([]int64, 0, len(instances))
	for _, instance := range instances {
		if instance == nil {
			continue
		}
		ids = append(ids, instance.ID)
	}
	return ids
}
