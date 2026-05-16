package infrastructure

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
)

func TestAddPausedDurationToActiveAWDContests(t *testing.T) {
	t.Parallel()

	db := contesttestsupport.SetupAWDTestDB(t)
	repo := NewRepository(db)
	now := time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC)
	activeAt := now
	updatedAt := now.Add(10 * time.Minute)
	recoveryKey := "boot-a|2026-05-16T10:00:00Z"

	rows := []model.Contest{
		{ID: 1, Title: "running-awd", Mode: model.ContestModeAWD, Status: model.ContestStatusRunning, StartTime: now.Add(-time.Hour), EndTime: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now},
		{ID: 2, Title: "frozen-awd", Mode: model.ContestModeAWD, Status: model.ContestStatusFrozen, StartTime: now.Add(-time.Hour), EndTime: now.Add(30 * time.Minute), CreatedAt: now, UpdatedAt: now},
		{ID: 3, Title: "expired-awd", Mode: model.ContestModeAWD, Status: model.ContestStatusRunning, StartTime: now.Add(-2 * time.Hour), EndTime: now.Add(-time.Minute), CreatedAt: now, UpdatedAt: now},
		{ID: 4, Title: "running-jeopardy", Mode: model.ContestModeJeopardy, Status: model.ContestStatusRunning, StartTime: now.Add(-time.Hour), EndTime: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now},
		{ID: 5, Title: "already-paused-awd", Mode: model.ContestModeAWD, Status: model.ContestStatusRunning, StartTime: now.Add(-2 * time.Hour), EndTime: now.Add(-time.Minute), PausedSeconds: 120, CreatedAt: now, UpdatedAt: now},
	}
	for _, row := range rows {
		contest := row
		if err := db.Create(&contest).Error; err != nil {
			t.Fatalf("create contest %d: %v", contest.ID, err)
		}
	}

	updated, err := repo.AddPausedDurationToActiveAWDContests(context.Background(), activeAt, recoveryKey, 120, updatedAt)
	if err != nil {
		t.Fatalf("AddPausedDurationToActiveAWDContests() error = %v", err)
	}
	if len(updated) != 3 {
		t.Fatalf("expected 3 updated contests, got %d", len(updated))
	}

	assertPausedSeconds(t, db, 1, 120)
	assertPausedSeconds(t, db, 2, 120)
	assertPausedSeconds(t, db, 3, 0)
	assertPausedSeconds(t, db, 4, 0)
	assertPausedSeconds(t, db, 5, 240)

	retried, err := repo.AddPausedDurationToActiveAWDContests(context.Background(), activeAt, recoveryKey, 120, updatedAt.Add(time.Minute))
	if err != nil {
		t.Fatalf("retry AddPausedDurationToActiveAWDContests() error = %v", err)
	}
	if len(retried) != 3 {
		t.Fatalf("expected retry with same target to return active contests for expiry refresh, got %d", len(retried))
	}
	assertPausedSeconds(t, db, 1, 120)
	assertPausedSeconds(t, db, 2, 120)
	assertPausedSeconds(t, db, 5, 240)

	resumed, err := repo.AddPausedDurationToActiveAWDContests(context.Background(), activeAt, recoveryKey, 150, updatedAt.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("resume AddPausedDurationToActiveAWDContests() error = %v", err)
	}
	if len(resumed) != 3 {
		t.Fatalf("expected resumed recovery to update 3 contests, got %d", len(resumed))
	}
	assertPausedSeconds(t, db, 1, 150)
	assertPausedSeconds(t, db, 2, 150)
	assertPausedSeconds(t, db, 5, 270)
}

func assertPausedSeconds(t *testing.T, db *gorm.DB, contestID int64, want int64) {
	t.Helper()

	var contest model.Contest
	if err := db.Where("id = ?", contestID).First(&contest).Error; err != nil {
		t.Fatalf("load contest %d: %v", contestID, err)
	}
	if contest.PausedSeconds != want {
		t.Fatalf("contest %d paused seconds = %d, want %d", contestID, contest.PausedSeconds, want)
	}
}
