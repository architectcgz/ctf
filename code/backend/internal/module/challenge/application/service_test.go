package application

import (
	"context"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengemodule "ctf-platform/internal/module/challenge"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	"ctf-platform/internal/module/challenge/testsupport"
	"ctf-platform/pkg/errcode"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestService(repo ChallengeRepository, imageRepo ImageRepository) *Service {
	return NewService(repo, imageRepo, nil, &challengemodule.Config{SolvedCountCacheTTL: time.Minute}, nil)
}

func TestServiceCreateChallengeSuccess(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	// 创建测试镜像
	db.Create(&model.Image{ID: 1, Name: "test-image"})

	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	service := newTestService(repo, imageRepo)

	resp, err := service.CreateChallenge(&dto.CreateChallengeReq{
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
}

func TestServiceCreateChallengeImageNotFound(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	service := newTestService(repo, imageRepo)

	_, err := service.CreateChallenge(&dto.CreateChallengeReq{
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

	resp, err := service.CreateChallenge(&dto.CreateChallengeReq{
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

func TestServiceGetPublishedChallengeNotPublished(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	challenge := &model.Challenge{Title: "Test", Status: model.ChallengeStatusDraft}
	db.Create(challenge)

	repo := challengeinfra.NewRepository(db)
	service := newTestService(repo, nil)

	_, err := service.GetPublishedChallenge(1, challenge.ID)
	if err == nil || err.Error() != errcode.ErrForbidden.Error() {
		t.Fatalf("expected not published error, got %v", err)
	}
}

func TestServiceGetChallengeIncludesHintsAndAttachment(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	challenge := &model.Challenge{
		Title:         "Hint Challenge",
		Description:   "desc",
		Category:      "web",
		Difficulty:    model.ChallengeDifficultyEasy,
		Points:        100,
		ImageID:       1,
		AttachmentURL: "https://example.com/files/hint.zip",
		Status:        model.ChallengeStatusDraft,
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := db.Create(&model.ChallengeHint{
		ChallengeID: challenge.ID,
		Level:       1,
		Title:       "第一条提示",
		Content:     "从登录入口开始",
	}).Error; err != nil {
		t.Fatalf("create hint: %v", err)
	}

	repo := challengeinfra.NewRepository(db)
	service := newTestService(repo, nil)

	resp, err := service.GetChallenge(challenge.ID)
	if err != nil {
		t.Fatalf("GetChallenge() error = %v", err)
	}
	if resp.AttachmentURL != challenge.AttachmentURL {
		t.Fatalf("unexpected attachment url: %s", resp.AttachmentURL)
	}
	if len(resp.Hints) != 1 || resp.Hints[0].Content != "从登录入口开始" {
		t.Fatalf("unexpected hints: %+v", resp.Hints)
	}
}

func TestServiceGetSolvedCountCachedHonorsContextCancellation(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	challenge := &model.Challenge{
		Title:  "Published",
		Status: model.ChallengeStatusPublished,
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := db.Create(&model.Submission{
		UserID:      1,
		ChallengeID: challenge.ID,
		IsCorrect:   true,
	}).Error; err != nil {
		t.Fatalf("create submission: %v", err)
	}

	mini := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := NewService(challengeinfra.NewRepository(db), nil, redisClient, &challengemodule.Config{SolvedCountCacheTTL: time.Minute}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.getSolvedCountCached(ctx, challenge.ID)
	if err == nil || err != context.Canceled {
		t.Fatalf("expected context canceled, got %v", err)
	}
}
