package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/model"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
)

func TestRepositoryUpdateDoesNotOverwriteStatusFields(t *testing.T) {
	db := contesttestsupport.SetupContestTestDB(t)
	repo := contestinfra.NewRepository(db)
	now := time.Now().UTC()

	if err := db.Create(&model.Contest{
		ID:            301,
		Title:         "before",
		Description:   "before-description",
		Mode:          model.ContestModeJeopardy,
		Status:        model.ContestStatusRunning,
		StatusVersion: 5,
		StartTime:     now.Add(-time.Hour),
		EndTime:       now.Add(time.Hour),
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}

	stale := &model.Contest{
		ID:            301,
		Title:         "after",
		Description:   "after-description",
		Mode:          model.ContestModeJeopardy,
		Status:        model.ContestStatusRegistration,
		StatusVersion: 0,
		StartTime:     now.Add(-30 * time.Minute),
		EndTime:       now.Add(2 * time.Hour),
	}
	if err := repo.Update(context.Background(), stale); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	var persisted model.Contest
	if err := db.First(&persisted, 301).Error; err != nil {
		t.Fatalf("load contest: %v", err)
	}
	if persisted.Title != "after" || persisted.Description != "after-description" {
		t.Fatalf("expected metadata to update, got %+v", persisted)
	}
	if persisted.Status != model.ContestStatusRunning || persisted.StatusVersion != 5 {
		t.Fatalf("expected status fields to stay untouched, got %+v", persisted)
	}
}
