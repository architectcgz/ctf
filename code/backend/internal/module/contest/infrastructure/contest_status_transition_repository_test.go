package infrastructure_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
)

func TestRepositoryRecordAppliedTransition(t *testing.T) {
	db := contesttestsupport.SetupContestTestDB(t)
	repo := contestinfra.NewRepository(db)
	now := time.Now().UTC()

	if err := db.Create(&model.Contest{
		ID:            201,
		Title:         "record-transition",
		Mode:          model.ContestModeJeopardy,
		Status:        model.ContestStatusRunning,
		StatusVersion: 1,
		StartTime:     now.Add(-time.Hour),
		EndTime:       now.Add(time.Hour),
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}

	recordID, err := repo.RecordAppliedTransition(context.Background(), contestdomain.ContestStatusTransitionResult{
		Transition: contestdomain.ContestStatusTransition{
			ContestID:         201,
			FromStatus:        model.ContestStatusRegistration,
			ToStatus:          model.ContestStatusRunning,
			FromStatusVersion: 0,
			Reason:            contestdomain.ContestStatusTransitionReasonTimeWindow,
			OccurredAt:        now,
			AppliedBy:         "contest_status_updater",
		},
		Applied:       true,
		StatusVersion: 1,
	})
	if err != nil {
		t.Fatalf("RecordAppliedTransition() error = %v", err)
	}
	if recordID <= 0 {
		t.Fatalf("expected positive record id, got %d", recordID)
	}

	var record model.ContestStatusTransition
	if err := db.First(&record, recordID).Error; err != nil {
		t.Fatalf("load transition record: %v", err)
	}
	if record.StatusVersion != 1 || record.SideEffectStatus != contestdomain.ContestStatusTransitionSideEffectPending {
		t.Fatalf("unexpected transition record: %+v", record)
	}
}

func TestRepositoryMarkTransitionSideEffects(t *testing.T) {
	db := contesttestsupport.SetupContestTestDB(t)
	repo := contestinfra.NewRepository(db)
	now := time.Now().UTC()
	if err := db.Create(&model.Contest{
		ID:            201,
		Title:         "mark-transition",
		Mode:          model.ContestModeJeopardy,
		Status:        model.ContestStatusRunning,
		StatusVersion: 1,
		StartTime:     now.Add(-time.Hour),
		EndTime:       now.Add(time.Hour),
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	record := model.ContestStatusTransition{
		ID:               301,
		ContestID:        201,
		StatusVersion:    1,
		FromStatus:       model.ContestStatusRegistration,
		ToStatus:         model.ContestStatusRunning,
		Reason:           contestdomain.ContestStatusTransitionReasonTimeWindow,
		AppliedBy:        "contest_status_updater",
		SideEffectStatus: contestdomain.ContestStatusTransitionSideEffectPending,
		OccurredAt:       now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := db.Create(&record).Error; err != nil {
		t.Fatalf("create transition record: %v", err)
	}

	if err := repo.MarkTransitionSideEffectsSucceeded(context.Background(), 301); err != nil {
		t.Fatalf("MarkTransitionSideEffectsSucceeded() error = %v", err)
	}

	var succeeded model.ContestStatusTransition
	if err := db.First(&succeeded, 301).Error; err != nil {
		t.Fatalf("load succeeded record: %v", err)
	}
	if succeeded.SideEffectStatus != contestdomain.ContestStatusTransitionSideEffectSucceeded {
		t.Fatalf("unexpected success status: %+v", succeeded)
	}

	if err := repo.MarkTransitionSideEffectsFailed(context.Background(), 301, errors.New("boom")); err != nil {
		t.Fatalf("MarkTransitionSideEffectsFailed() error = %v", err)
	}

	var failed model.ContestStatusTransition
	if err := db.First(&failed, 301).Error; err != nil {
		t.Fatalf("load failed record: %v", err)
	}
	if failed.SideEffectStatus != contestdomain.ContestStatusTransitionSideEffectFailed || failed.SideEffectError != "boom" {
		t.Fatalf("unexpected failed status: %+v", failed)
	}
}

func TestRepositoryRecordAppliedTransitionReturnsExistingRecordOnDuplicate(t *testing.T) {
	db := contesttestsupport.SetupContestTestDB(t)
	repo := contestinfra.NewRepository(db)
	now := time.Now().UTC()
	if err := db.Create(&model.Contest{
		ID:            202,
		Title:         "duplicate-transition",
		Mode:          model.ContestModeJeopardy,
		Status:        model.ContestStatusRunning,
		StatusVersion: 1,
		StartTime:     now.Add(-time.Hour),
		EndTime:       now.Add(time.Hour),
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}

	firstID, err := repo.RecordAppliedTransition(context.Background(), contestdomain.ContestStatusTransitionResult{
		Transition: contestdomain.ContestStatusTransition{
			ContestID:         202,
			FromStatus:        model.ContestStatusRegistration,
			ToStatus:          model.ContestStatusRunning,
			FromStatusVersion: 0,
			Reason:            contestdomain.ContestStatusTransitionReasonManualUpdate,
			OccurredAt:        now,
			AppliedBy:         "contest_service",
		},
		Applied:       true,
		StatusVersion: 1,
	})
	if err != nil {
		t.Fatalf("first RecordAppliedTransition() error = %v", err)
	}

	secondID, err := repo.RecordAppliedTransition(context.Background(), contestdomain.ContestStatusTransitionResult{
		Transition: contestdomain.ContestStatusTransition{
			ContestID:         202,
			FromStatus:        model.ContestStatusRegistration,
			ToStatus:          model.ContestStatusRunning,
			FromStatusVersion: 0,
			Reason:            contestdomain.ContestStatusTransitionReasonManualUpdate,
			OccurredAt:        now,
			AppliedBy:         "contest_service",
		},
		Applied:       true,
		StatusVersion: 1,
	})
	if err != nil {
		t.Fatalf("second RecordAppliedTransition() error = %v", err)
	}
	if firstID != secondID {
		t.Fatalf("expected duplicate insert to return existing id %d, got %d", firstID, secondID)
	}
}

func TestRepositoryListTransitionsForSideEffectReplay(t *testing.T) {
	db := contesttestsupport.SetupContestTestDB(t)
	repo := contestinfra.NewRepository(db)
	now := time.Now().UTC()
	if err := db.Create(&model.Contest{
		ID:            203,
		Title:         "replayable-transition",
		Mode:          model.ContestModeJeopardy,
		Status:        model.ContestStatusFrozen,
		StatusVersion: 2,
		StartTime:     now.Add(-time.Hour),
		EndTime:       now.Add(time.Hour),
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	records := []model.ContestStatusTransition{
		{
			ID:               401,
			ContestID:        203,
			StatusVersion:    1,
			FromStatus:       model.ContestStatusRegistration,
			ToStatus:         model.ContestStatusRunning,
			Reason:           contestdomain.ContestStatusTransitionReasonTimeWindow,
			AppliedBy:        "contest_status_updater",
			SideEffectStatus: contestdomain.ContestStatusTransitionSideEffectPending,
			OccurredAt:       now.Add(-time.Minute),
			CreatedAt:        now.Add(-time.Minute),
			UpdatedAt:        now.Add(-time.Minute),
		},
		{
			ID:               402,
			ContestID:        203,
			StatusVersion:    2,
			FromStatus:       model.ContestStatusRunning,
			ToStatus:         model.ContestStatusFrozen,
			Reason:           contestdomain.ContestStatusTransitionReasonTimeWindow,
			AppliedBy:        "contest_status_updater",
			SideEffectStatus: contestdomain.ContestStatusTransitionSideEffectFailed,
			SideEffectError:  "redis timeout",
			OccurredAt:       now,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}
	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("create transition records: %v", err)
	}

	results, err := repo.ListTransitionsForSideEffectReplay(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListTransitionsForSideEffectReplay() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 replayable transitions, got %d", len(results))
	}
	if results[0].RecordID != 401 || results[1].RecordID != 402 {
		t.Fatalf("unexpected replay order: %+v", results)
	}
	if !results[0].Applied || !results[1].Applied {
		t.Fatalf("expected replay entries to be marked applied: %+v", results)
	}
}
