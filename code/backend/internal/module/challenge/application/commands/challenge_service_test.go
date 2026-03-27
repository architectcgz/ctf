package commands

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/internal/module/challenge/testsupport"
	"ctf-platform/pkg/errcode"
	"testing"
)

func newTestService(repo challengeports.ChallengeCommandRepository, imageRepo challengeports.ImageRepository) *ChallengeService {
	return NewChallengeService(repo, imageRepo)
}

func TestServiceCreateChallengeSuccess(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	// 创建测试镜像
	db.Create(&model.Image{ID: 1, Name: "test-image"})

	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	service := newTestService(repo, imageRepo)

	resp, err := service.CreateChallenge(1001, &dto.CreateChallengeReq{
		Title:       "Test Challenge",
		Description: "Test",
		Category:    "web",
		Difficulty:  "easy",
		Points:      100,
		ImageID:     1,
	})

	if err != nil {
		t.Fatalf("CreateChallenge() error = %v", err)
	}
	if resp.Status != model.ChallengeStatusDraft {
		t.Fatalf("unexpected status: %s", resp.Status)
	}
	if resp.CreatedBy == nil || *resp.CreatedBy != 1001 {
		t.Fatalf("unexpected created_by: %+v", resp.CreatedBy)
	}
}

func TestServiceCreateChallengeImageNotFound(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	service := newTestService(repo, imageRepo)

	_, err := service.CreateChallenge(1001, &dto.CreateChallengeReq{
		ImageID: 999,
	})
	if err == nil || err.Error() != errcode.ErrNotFound.Error() {
		t.Fatalf("expected image not found error, got %v", err)
	}
}

func TestServiceCreateChallengeWithoutImageSuccess(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	service := newTestService(repo, imageRepo)

	resp, err := service.CreateChallenge(1001, &dto.CreateChallengeReq{
		Title:       "No Target Challenge",
		Description: "No target required",
		Category:    "misc",
		Difficulty:  "easy",
		Points:      50,
		ImageID:     0,
	})

	if err != nil {
		t.Fatalf("CreateChallenge() without image error = %v", err)
	}
	if resp.ImageID != 0 {
		t.Fatalf("expected image_id=0, got %d", resp.ImageID)
	}
}

func TestServiceDeleteChallengeWithRunningInstances(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	// 创建靶场和运行中的实例
	challenge := &model.Challenge{Title: "Test", Status: "draft"}
	db.Create(challenge)
	db.Create(&model.Instance{ChallengeID: challenge.ID, Status: "running"})

	repo := challengeinfra.NewRepository(db)
	service := newTestService(repo, nil)

	err := service.DeleteChallenge(challenge.ID)
	if err == nil || err.Error() != errcode.ErrConflict.Error() {
		t.Fatalf("expected running instances error, got %v", err)
	}
}

func TestServicePublishChallengeNoImage(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	challenge := &model.Challenge{Title: "Test", ImageID: 0}
	db.Create(challenge)

	repo := challengeinfra.NewRepository(db)
	service := newTestService(repo, nil)

	err := service.PublishChallenge(challenge.ID)
	if err != nil {
		t.Fatalf("PublishChallenge() error = %v", err)
	}

	published, findErr := repo.FindByID(challenge.ID)
	if findErr != nil {
		t.Fatalf("FindByID() error = %v", findErr)
	}
	if published.Status != model.ChallengeStatusPublished {
		t.Fatalf("expected published status, got %s", published.Status)
	}
}
