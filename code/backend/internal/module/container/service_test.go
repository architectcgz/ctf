package container

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

func TestRepositoryListActiveContainerIDs(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	seedInstance(t, repo.db, &model.Instance{
		UserID:      1,
		ChallengeID: 101,
		ContainerID: "running-container",
		Status:      model.InstanceStatusRunning,
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	seedInstance(t, repo.db, &model.Instance{
		UserID:      1,
		ChallengeID: 102,
		ContainerID: "creating-container",
		Status:      model.InstanceStatusCreating,
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	seedInstance(t, repo.db, &model.Instance{
		UserID:      1,
		ChallengeID: 103,
		ContainerID: "stopped-container",
		Status:      model.InstanceStatusStopped,
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	seedInstance(t, repo.db, &model.Instance{
		UserID:      1,
		ChallengeID: 104,
		ContainerID: "",
		Status:      model.InstanceStatusRunning,
		ExpiresAt:   time.Now().Add(time.Hour),
	})

	containerIDs, err := repo.ListActiveContainerIDs()
	if err != nil {
		t.Fatalf("ListActiveContainerIDs() error = %v", err)
	}
	if len(containerIDs) != 2 {
		t.Fatalf("expected 2 active container ids, got %d (%v)", len(containerIDs), containerIDs)
	}

	got := make(map[string]struct{}, len(containerIDs))
	for _, containerID := range containerIDs {
		got[containerID] = struct{}{}
	}
	if _, exists := got["running-container"]; !exists {
		t.Fatalf("running container not returned: %v", containerIDs)
	}
	if _, exists := got["creating-container"]; !exists {
		t.Fatalf("creating container not returned: %v", containerIDs)
	}
}

func TestSelectOrphanContainersSkipsActiveAndGracePeriod(t *testing.T) {
	t.Parallel()

	now := time.Now()
	managedContainers := []ManagedContainer{
		{ID: "active", Name: "ctf-instance-active", CreatedAt: now.Add(-10 * time.Minute)},
		{ID: "fresh", Name: "ctf-instance-fresh", CreatedAt: now.Add(-2 * time.Minute)},
		{ID: "orphan", Name: "ctf-instance-orphan", CreatedAt: now.Add(-12 * time.Minute)},
	}
	activeContainerIDs := map[string]struct{}{
		"active": {},
	}

	orphanContainers := selectOrphanContainers(managedContainers, activeContainerIDs, 5*time.Minute, now)
	if len(orphanContainers) != 1 {
		t.Fatalf("expected 1 orphan container, got %d (%v)", len(orphanContainers), orphanContainers)
	}
	if orphanContainers[0].ID != "orphan" {
		t.Fatalf("unexpected orphan container: %+v", orphanContainers[0])
	}
}

func TestManagedContainerLabels(t *testing.T) {
	t.Parallel()

	labels := managedContainerLabels()
	if labels[managedByLabelKey] != managedByLabelValue {
		t.Fatalf("expected managed-by label, got %v", labels)
	}
	if labels[challengeInstanceLabelKey] != challengeInstanceLabelValue {
		t.Fatalf("expected component label, got %v", labels)
	}
}

func TestManagedNetworkLabels(t *testing.T) {
	t.Parallel()

	labels := managedNetworkLabels()
	if labels[managedByLabelKey] != managedByLabelValue {
		t.Fatalf("expected managed-by label, got %v", labels)
	}
	if labels[challengeInstanceLabelKey] != challengeInstanceLabelValue {
		t.Fatalf("expected component label, got %v", labels)
	}
}

func TestServiceCreateContainerCreatesIsolatedNetwork(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:   "net-123",
		containerID: "ctr-123",
	}
	service := NewService(repo, engine, &config.ContainerConfig{
		PortRangeStart:     30000,
		PortRangeEnd:       30010,
		DefaultExposedPort: 8080,
	}, nil)

	containerID, networkID, port, err := service.CreateContainer(context.Background(), "ctf/web:v1", map[string]string{"FLAG": "flag{1}"})
	if err != nil {
		t.Fatalf("CreateContainer() error = %v", err)
	}
	if containerID != "ctr-123" {
		t.Fatalf("unexpected container id: %s", containerID)
	}
	if networkID != "net-123" {
		t.Fatalf("unexpected network id: %s", networkID)
	}
	if port != 30000 {
		t.Fatalf("unexpected port: %d", port)
	}
	if engine.createdNetworkName == "" {
		t.Fatalf("expected isolated network to be created")
	}
	if engine.createdContainerCfg == nil || engine.createdContainerCfg.Network != engine.createdNetworkName {
		t.Fatalf("expected container to join created network, cfg=%+v network=%s", engine.createdContainerCfg, engine.createdNetworkName)
	}
}

func TestServiceCreateContainerRemovesNetworkWhenStartFails(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:   "net-456",
		containerID: "ctr-456",
		startErr:    errors.New("start failed"),
	}
	service := NewService(repo, engine, &config.ContainerConfig{
		PortRangeStart:     30000,
		PortRangeEnd:       30010,
		DefaultExposedPort: 8080,
	}, nil)

	_, _, _, err := service.CreateContainer(context.Background(), "ctf/web:v1", nil)
	if err == nil {
		t.Fatal("expected CreateContainer() to fail")
	}
	if engine.removedContainerID != "ctr-456" {
		t.Fatalf("expected container cleanup, got %s", engine.removedContainerID)
	}
	if engine.removedNetworkID != "net-456" {
		t.Fatalf("expected network cleanup, got %s", engine.removedNetworkID)
	}
}

func TestServiceListTeacherInstancesScopesTeacherAndAppliesFilters(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestService(repo)
	now := time.Now()

	seedUser(t, repo.db, &model.User{ID: 1, Username: "teacher-a", Role: model.RoleTeacher, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedUser(t, repo.db, &model.User{ID: 2, Username: "alice", StudentNo: "S-1001", Role: model.RoleStudent, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedUser(t, repo.db, &model.User{ID: 3, Username: "bob", StudentNo: "S-1002", Role: model.RoleStudent, ClassName: "Class B", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedChallenge(t, repo.db, &model.Challenge{ID: 11, Title: "web-101", Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &model.Instance{ID: 101, UserID: 2, ChallengeID: 11, ContainerID: "inst-a", Status: model.InstanceStatusRunning, ExpiresAt: now.Add(30 * time.Minute), CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &model.Instance{ID: 102, UserID: 3, ChallengeID: 11, ContainerID: "inst-b", Status: model.InstanceStatusRunning, ExpiresAt: now.Add(30 * time.Minute), CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &model.Instance{ID: 103, UserID: 2, ChallengeID: 11, ContainerID: "inst-stopped", Status: model.InstanceStatusStopped, ExpiresAt: now.Add(30 * time.Minute), CreatedAt: now, UpdatedAt: now})

	items, err := service.ListTeacherInstances(context.Background(), 1, model.RoleTeacher, nil)
	if err != nil {
		t.Fatalf("ListTeacherInstances() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 visible instance, got %d (%+v)", len(items), items)
	}
	if items[0].StudentUsername != "alice" || items[0].ClassName != "Class A" {
		t.Fatalf("unexpected item: %+v", items[0])
	}

	filtered, err := service.ListTeacherInstances(context.Background(), 1, model.RoleTeacher, &dto.TeacherInstanceQuery{
		Keyword:   "ali",
		StudentNo: "S-1001",
	})
	if err != nil {
		t.Fatalf("ListTeacherInstances() with filters error = %v", err)
	}
	if len(filtered) != 1 || filtered[0].ID != 101 {
		t.Fatalf("unexpected filtered result: %+v", filtered)
	}
}

func TestServiceListTeacherInstancesRejectsTeacherCrossClassFilter(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestService(repo)
	now := time.Now()

	seedUser(t, repo.db, &model.User{ID: 1, Username: "teacher-a", Role: model.RoleTeacher, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})

	_, err := service.ListTeacherInstances(context.Background(), 1, model.RoleTeacher, &dto.TeacherInstanceQuery{ClassName: "Class B"})
	if err == nil || err.Error() != errcode.ErrForbidden.Error() {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestServiceDestroyTeacherInstanceHonorsClassScope(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestService(repo)
	now := time.Now()

	seedUser(t, repo.db, &model.User{ID: 1, Username: "teacher-a", Role: model.RoleTeacher, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedUser(t, repo.db, &model.User{ID: 2, Username: "alice", Role: model.RoleStudent, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedUser(t, repo.db, &model.User{ID: 3, Username: "bob", Role: model.RoleStudent, ClassName: "Class B", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedChallenge(t, repo.db, &model.Challenge{ID: 11, Title: "web-101", Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &model.Instance{ID: 201, UserID: 2, ChallengeID: 11, ContainerID: "inst-a", Status: model.InstanceStatusRunning, ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &model.Instance{ID: 202, UserID: 3, ChallengeID: 11, ContainerID: "inst-b", Status: model.InstanceStatusRunning, ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})

	if err := service.DestroyTeacherInstance(context.Background(), 202, 1, model.RoleTeacher); err == nil || err.Error() != errcode.ErrForbidden.Error() {
		t.Fatalf("expected forbidden destroy, got %v", err)
	}

	if err := service.DestroyTeacherInstance(context.Background(), 201, 1, model.RoleTeacher); err != nil {
		t.Fatalf("DestroyTeacherInstance() error = %v", err)
	}

	instance, err := repo.FindByID(201)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if instance.Status != model.InstanceStatusStopped {
		t.Fatalf("expected stopped status, got %s", instance.Status)
	}
}

func newTestRepository(t *testing.T) *Repository {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Challenge{}, &model.Instance{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	return NewRepository(db)
}

func newTestService(repo *Repository) *Service {
	return NewService(repo, nil, &config.ContainerConfig{
		MaxExtends:        2,
		ExtendDuration:    30 * time.Minute,
		OrphanGracePeriod: 5 * time.Minute,
	}, nil)
}

type fakeRuntimeEngine struct {
	networkID           string
	containerID         string
	startErr            error
	createdNetworkName  string
	createdNetworkLabel map[string]string
	createdContainerCfg *model.ContainerConfig
	removedContainerID  string
	removedNetworkID    string
}

func (f *fakeRuntimeEngine) CreateNetwork(_ context.Context, name string, labels map[string]string) (string, error) {
	f.createdNetworkName = name
	f.createdNetworkLabel = labels
	return f.networkID, nil
}

func (f *fakeRuntimeEngine) CreateContainer(_ context.Context, cfg *model.ContainerConfig) (string, error) {
	f.createdContainerCfg = cfg
	return f.containerID, nil
}

func (f *fakeRuntimeEngine) StartContainer(_ context.Context, _ string) error {
	return f.startErr
}

func (f *fakeRuntimeEngine) StopContainer(_ context.Context, _ string, _ time.Duration) error {
	return nil
}

func (f *fakeRuntimeEngine) RemoveContainer(_ context.Context, containerID string, _ bool) error {
	f.removedContainerID = containerID
	return nil
}

func (f *fakeRuntimeEngine) RemoveNetwork(_ context.Context, networkID string) error {
	f.removedNetworkID = networkID
	return nil
}

func (f *fakeRuntimeEngine) ListManagedContainers(_ context.Context, _ string) ([]ManagedContainer, error) {
	return nil, nil
}

func seedInstance(t *testing.T, db *gorm.DB, instance *model.Instance) {
	t.Helper()

	if err := db.Create(instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
}

func seedUser(t *testing.T, db *gorm.DB, user *model.User) {
	t.Helper()

	if err := db.Create(user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

func seedChallenge(t *testing.T, db *gorm.DB, challenge *model.Challenge) {
	t.Helper()

	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}
}
