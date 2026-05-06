package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

func TestAWDDefenseWorkspaceUniqueScopeConstraint(t *testing.T) {
	t.Parallel()

	db := newAWDDefenseWorkspaceRepositoryTestDB(t)
	first := &model.AWDDefenseWorkspace{
		ContestID:         101,
		TeamID:            201,
		ServiceID:         301,
		InstanceID:        401,
		WorkspaceRevision: 1,
		Status:            model.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-ctr-1",
		SeedSignature:     "seed:v1",
	}
	if err := db.Create(first).Error; err != nil {
		t.Fatalf("create first workspace: %v", err)
	}

	second := &model.AWDDefenseWorkspace{
		ContestID:         101,
		TeamID:            201,
		ServiceID:         301,
		InstanceID:        402,
		WorkspaceRevision: 1,
		Status:            model.AWDDefenseWorkspaceStatusProvisioning,
		SeedSignature:     "seed:v2",
	}
	if err := db.Create(second).Error; err == nil {
		t.Fatal("expected duplicate workspace scope to violate unique constraint")
	}
}

func TestRepositoryFindAWDDefenseWorkspaceReturnsScopedRecord(t *testing.T) {
	t.Parallel()

	db := newAWDDefenseWorkspaceRepositoryTestDB(t)
	workspace := &model.AWDDefenseWorkspace{
		ContestID:         102,
		TeamID:            202,
		ServiceID:         302,
		InstanceID:        402,
		WorkspaceRevision: 3,
		Status:            model.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-ctr-2",
		SeedSignature:     "seed:v3",
	}
	if err := db.Create(workspace).Error; err != nil {
		t.Fatalf("seed workspace: %v", err)
	}

	found, err := NewRepository(db).FindAWDDefenseWorkspace(context.Background(), 102, 202, 302)
	if err != nil {
		t.Fatalf("FindAWDDefenseWorkspace() error = %v", err)
	}
	if found == nil {
		t.Fatal("expected workspace record")
	}
	if found.ID != workspace.ID || found.ContainerID != "workspace-ctr-2" || found.WorkspaceRevision != 3 {
		t.Fatalf("unexpected workspace record: %+v", found)
	}
}

func TestRepositoryUpsertAWDDefenseWorkspaceCreatesAndUpdatesScope(t *testing.T) {
	t.Parallel()

	db := newAWDDefenseWorkspaceRepositoryTestDB(t)
	repo := NewRepository(db)

	workspace := &model.AWDDefenseWorkspace{
		ContestID:         103,
		TeamID:            203,
		ServiceID:         303,
		InstanceID:        403,
		WorkspaceRevision: 1,
		Status:            model.AWDDefenseWorkspaceStatusProvisioning,
		SeedSignature:     "seed:v1",
	}
	if err := repo.UpsertAWDDefenseWorkspace(context.Background(), workspace); err != nil {
		t.Fatalf("UpsertAWDDefenseWorkspace() create error = %v", err)
	}

	updated := &model.AWDDefenseWorkspace{
		ContestID:         103,
		TeamID:            203,
		ServiceID:         303,
		InstanceID:        404,
		WorkspaceRevision: 1,
		Status:            model.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-ctr-3",
		SeedSignature:     "seed:v1",
	}
	if err := repo.UpsertAWDDefenseWorkspace(context.Background(), updated); err != nil {
		t.Fatalf("UpsertAWDDefenseWorkspace() update error = %v", err)
	}

	stored, err := repo.FindAWDDefenseWorkspace(context.Background(), 103, 203, 303)
	if err != nil {
		t.Fatalf("FindAWDDefenseWorkspace() error = %v", err)
	}
	if stored == nil {
		t.Fatal("expected updated workspace")
	}
	if stored.InstanceID != 404 || stored.WorkspaceRevision != 1 || stored.ContainerID != "workspace-ctr-3" || stored.SeedSignature != "seed:v1" {
		t.Fatalf("unexpected stored workspace: %+v", stored)
	}

	var count int64
	if err := db.Model(&model.AWDDefenseWorkspace{}).
		Where("contest_id = ? AND team_id = ? AND service_id = ?", 103, 203, 303).
		Count(&count).Error; err != nil {
		t.Fatalf("count workspace rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one workspace row after upsert, got %d", count)
	}
}

func TestRepositoryBumpAWDDefenseWorkspaceRevisionResetsProvisioningState(t *testing.T) {
	t.Parallel()

	db := newAWDDefenseWorkspaceRepositoryTestDB(t)
	repo := NewRepository(db)

	if err := repo.UpsertAWDDefenseWorkspace(context.Background(), &model.AWDDefenseWorkspace{
		ContestID:         104,
		TeamID:            204,
		ServiceID:         304,
		InstanceID:        405,
		WorkspaceRevision: 7,
		Status:            model.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-ctr-old",
		SeedSignature:     "seed:old",
	}); err != nil {
		t.Fatalf("seed workspace: %v", err)
	}

	if err := repo.BumpAWDDefenseWorkspaceRevision(context.Background(), 104, 204, 304, 406, "seed:new"); err != nil {
		t.Fatalf("BumpAWDDefenseWorkspaceRevision() error = %v", err)
	}

	stored, err := repo.FindAWDDefenseWorkspace(context.Background(), 104, 204, 304)
	if err != nil {
		t.Fatalf("FindAWDDefenseWorkspace() error = %v", err)
	}
	if stored == nil {
		t.Fatal("expected bumped workspace")
	}
	if stored.WorkspaceRevision != 8 {
		t.Fatalf("workspace revision = %d, want 8", stored.WorkspaceRevision)
	}
	if stored.InstanceID != 406 {
		t.Fatalf("instance id = %d, want 406", stored.InstanceID)
	}
	if stored.Status != model.AWDDefenseWorkspaceStatusProvisioning {
		t.Fatalf("status = %q, want %q", stored.Status, model.AWDDefenseWorkspaceStatusProvisioning)
	}
	if stored.ContainerID != "" {
		t.Fatalf("container id = %q, want empty after reseed bump", stored.ContainerID)
	}
	if stored.SeedSignature != "seed:new" {
		t.Fatalf("seed signature = %q, want %q", stored.SeedSignature, "seed:new")
	}
}

func newAWDDefenseWorkspaceRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", name)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.AWDDefenseWorkspace{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}
	return db
}
