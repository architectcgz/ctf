package infrastructure

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	contestentity "ctf-platform/internal/module/contest/entity"
)

func newContestRuntimePlacementRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&contestentity.ContestRuntimePlacement{}); err != nil {
		t.Fatalf("auto migrate contest runtime placement: %v", err)
	}
	if err := db.Exec(`
		CREATE UNIQUE INDEX idx_contest_runtime_placements_active_contest
		ON contest_runtime_placements(contest_id)
		WHERE status = 'active'
	`).Error; err != nil {
		t.Fatalf("create active placement unique index: %v", err)
	}
	return db
}

func TestContestRuntimePlacementRepository(t *testing.T) {
	t.Parallel()

	db := newContestRuntimePlacementRepositoryTestDB(t)
	repo := NewContestRuntimePlacementRepository(db)
	ctx := context.Background()

	if placement, exists, err := repo.FindActiveContestRuntimePlacement(ctx, 1001); err != nil || exists || placement != nil {
		t.Fatalf("FindActiveContestRuntimePlacement() = (%+v, %v, %v), want nil false nil", placement, exists, err)
	}

	created, err := repo.EnsureActiveContestRuntimePlacement(ctx, 1001, 7001)
	if err != nil {
		t.Fatalf("EnsureActiveContestRuntimePlacement() create error = %v", err)
	}
	if created == nil || created.ContestID != 1001 || created.RuntimeNodeID != 7001 || created.Status != contestentity.ContestRuntimePlacementStatusActive {
		t.Fatalf("created placement = %+v, want contest 1001 runtime node 7001 active", created)
	}

	existing, err := repo.EnsureActiveContestRuntimePlacement(ctx, 1001, 7002)
	if err != nil {
		t.Fatalf("EnsureActiveContestRuntimePlacement() existing error = %v", err)
	}
	if existing == nil || existing.ID != created.ID || existing.RuntimeNodeID != 7001 {
		t.Fatalf("existing placement = %+v, want original %+v", existing, created)
	}
	assertContestRuntimePlacementCount(t, db, 1001, contestentity.ContestRuntimePlacementStatusActive, 1)

	releasedAt := time.Now().UTC()
	if err := db.Model(&contestentity.ContestRuntimePlacement{}).
		Where("id = ?", created.ID).
		Updates(map[string]any{
			"status":      contestentity.ContestRuntimePlacementStatusReleased,
			"released_at": releasedAt,
		}).Error; err != nil {
		t.Fatalf("release placement: %v", err)
	}

	recreated, err := repo.EnsureActiveContestRuntimePlacement(ctx, 1001, 7003)
	if err != nil {
		t.Fatalf("EnsureActiveContestRuntimePlacement() after release error = %v", err)
	}
	if recreated == nil || recreated.ID == created.ID || recreated.RuntimeNodeID != 7003 || recreated.Status != contestentity.ContestRuntimePlacementStatusActive {
		t.Fatalf("recreated placement = %+v, want new active node 7003", recreated)
	}
	assertContestRuntimePlacementCount(t, db, 1001, contestentity.ContestRuntimePlacementStatusActive, 1)
	assertContestRuntimePlacementCount(t, db, 1001, contestentity.ContestRuntimePlacementStatusReleased, 1)
}

func assertContestRuntimePlacementCount(t *testing.T, db *gorm.DB, contestID int64, status string, want int64) {
	t.Helper()

	var count int64
	if err := db.Model(&contestentity.ContestRuntimePlacement{}).
		Where("contest_id = ? AND status = ?", contestID, status).
		Count(&count).Error; err != nil {
		t.Fatalf("count placements: %v", err)
	}
	if count != want {
		t.Fatalf("placement count for contest %d status %q = %d, want %d", contestID, status, count, want)
	}
}
