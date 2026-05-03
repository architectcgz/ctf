package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
)

func TestRepositoryApplyStatusTransitionApplied(t *testing.T) {
	db := contesttestsupport.SetupContestTestDB(t)
	repo := contestinfra.NewRepository(db)
	now := time.Now().UTC()

	if err := db.Create(&model.Contest{
		ID:            101,
		Title:         "status-transition",
		Mode:          model.ContestModeJeopardy,
		Status:        model.ContestStatusRegistration,
		StatusVersion: 0,
		StartTime:     now.Add(-time.Minute),
		EndTime:       now.Add(time.Hour),
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}

	result, err := repo.ApplyStatusTransition(context.Background(), contestdomain.ContestStatusTransition{
		ContestID:         101,
		FromStatus:        model.ContestStatusRegistration,
		ToStatus:          model.ContestStatusRunning,
		FromStatusVersion: 0,
		OccurredAt:        now,
	})
	if err != nil {
		t.Fatalf("ApplyStatusTransition() error = %v", err)
	}
	if !result.Applied {
		t.Fatal("expected transition to apply")
	}
	if result.StatusVersion != 1 {
		t.Fatalf("expected status version 1, got %d", result.StatusVersion)
	}

	var contest model.Contest
	if err := db.First(&contest, 101).Error; err != nil {
		t.Fatalf("load contest: %v", err)
	}
	if contest.Status != model.ContestStatusRunning {
		t.Fatalf("expected status running, got %q", contest.Status)
	}
	if contest.StatusVersion != 1 {
		t.Fatalf("expected persisted status version 1, got %d", contest.StatusVersion)
	}

	var transition model.ContestStatusTransition
	if err := db.Where("contest_id = ? AND status_version = ?", 101, 1).First(&transition).Error; err != nil {
		t.Fatalf("load transition record: %v", err)
	}
	if transition.SideEffectStatus != contestdomain.ContestStatusTransitionSideEffectPending {
		t.Fatalf("unexpected transition record: %+v", transition)
	}
}

func TestRepositoryApplyStatusTransitionReturnsStaleWhenVersionChanged(t *testing.T) {
	db := contesttestsupport.SetupContestTestDB(t)
	repo := contestinfra.NewRepository(db)
	now := time.Now().UTC()

	if err := db.Create(&model.Contest{
		ID:            102,
		Title:         "stale-transition",
		Mode:          model.ContestModeJeopardy,
		Status:        model.ContestStatusRunning,
		StatusVersion: 2,
		StartTime:     now.Add(-time.Hour),
		EndTime:       now.Add(time.Hour),
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}

	result, err := repo.ApplyStatusTransition(context.Background(), contestdomain.ContestStatusTransition{
		ContestID:         102,
		FromStatus:        model.ContestStatusRunning,
		ToStatus:          model.ContestStatusFrozen,
		FromStatusVersion: 1,
		OccurredAt:        now,
	})
	if err != nil {
		t.Fatalf("ApplyStatusTransition() error = %v", err)
	}
	if result.Applied {
		t.Fatal("expected stale transition to be skipped")
	}
}

func TestRepositoryApplyStatusTransitionMissingContest(t *testing.T) {
	db := contesttestsupport.SetupContestTestDB(t)
	repo := contestinfra.NewRepository(db)

	_, err := repo.ApplyStatusTransition(context.Background(), contestdomain.ContestStatusTransition{
		ContestID:         404,
		FromStatus:        model.ContestStatusRegistration,
		ToStatus:          model.ContestStatusRunning,
		FromStatusVersion: 0,
		OccurredAt:        time.Now().UTC(),
	})
	if err != contestdomain.ErrContestNotFound {
		t.Fatalf("expected ErrContestNotFound, got %v", err)
	}
}

func TestRepositoryUpdateContestWithStatusTransitionUsesCompareAndSet(t *testing.T) {
	db := contesttestsupport.SetupContestTestDB(t)
	repo := contestinfra.NewRepository(db)
	now := time.Now().UTC()

	if err := db.Create(&model.Contest{
		ID:            103,
		Title:         "manual-cas",
		Description:   "before",
		Mode:          model.ContestModeJeopardy,
		Status:        model.ContestStatusRegistration,
		StatusVersion: 1,
		StartTime:     now.Add(-time.Minute),
		EndTime:       now.Add(time.Hour),
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}

	contest := &model.Contest{
		ID:            103,
		Title:         "manual-cas-updated",
		Description:   "after",
		Mode:          model.ContestModeJeopardy,
		Status:        model.ContestStatusRunning,
		StatusVersion: 2,
		StartTime:     now.Add(-time.Minute),
		EndTime:       now.Add(time.Hour),
		CreatedAt:     now,
		UpdatedAt:     now.Add(time.Minute),
	}

	result, err := repo.UpdateContestWithStatusTransition(context.Background(), contest, contestdomain.ContestStatusTransition{
		ContestID:         103,
		FromStatus:        model.ContestStatusRegistration,
		ToStatus:          model.ContestStatusRunning,
		FromStatusVersion: 1,
		OccurredAt:        contest.UpdatedAt,
		Reason:            contestdomain.ContestStatusTransitionReasonManualUpdate,
		AppliedBy:         "contest_service",
	})
	if err != nil {
		t.Fatalf("UpdateContestWithStatusTransition() error = %v", err)
	}
	if !result.Applied || result.StatusVersion != 2 || result.RecordID <= 0 {
		t.Fatalf("unexpected transition result: %+v", result)
	}

	staleContest := &model.Contest{
		ID:            103,
		Title:         "should-not-win",
		Description:   "stale",
		Mode:          model.ContestModeJeopardy,
		Status:        model.ContestStatusFrozen,
		StatusVersion: 2,
		StartTime:     now.Add(-time.Minute),
		EndTime:       now.Add(time.Hour),
		CreatedAt:     now,
		UpdatedAt:     now.Add(2 * time.Minute),
	}
	staleResult, err := repo.UpdateContestWithStatusTransition(context.Background(), staleContest, contestdomain.ContestStatusTransition{
		ContestID:         103,
		FromStatus:        model.ContestStatusRegistration,
		ToStatus:          model.ContestStatusFrozen,
		FromStatusVersion: 1,
		OccurredAt:        staleContest.UpdatedAt,
		Reason:            contestdomain.ContestStatusTransitionReasonManualUpdate,
		AppliedBy:         "contest_service",
	})
	if err != nil {
		t.Fatalf("stale UpdateContestWithStatusTransition() error = %v", err)
	}
	if staleResult.Applied {
		t.Fatalf("expected stale manual transition to be skipped, got %+v", staleResult)
	}

	var persisted model.Contest
	if err := db.First(&persisted, 103).Error; err != nil {
		t.Fatalf("load contest: %v", err)
	}
	if persisted.Status != model.ContestStatusRunning || persisted.Title != "manual-cas-updated" {
		t.Fatalf("unexpected persisted contest: %+v", persisted)
	}
}
