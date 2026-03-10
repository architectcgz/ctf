package challenge

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
	"testing"
	"time"
)

func newTestService(repo *Repository, imageRepo *ImageRepository) *Service {
	return NewService(repo, imageRepo, nil, &Config{SolvedCountCacheTTL: time.Minute}, nil)
}

func TestServiceCreateChallengeSuccess(t *testing.T) {
	db := setupTestDB(t)

	// 创建测试镜像
	db.Create(&model.Image{ID: 1, Name: "test-image"})

	repo := NewRepository(db)
	imageRepo := NewImageRepository(db)
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
	db := setupTestDB(t)

	repo := NewRepository(db)
	imageRepo := NewImageRepository(db)
	service := newTestService(repo, imageRepo)

	_, err := service.CreateChallenge(&dto.CreateChallengeReq{
		ImageID: 999,
	})
	if err == nil || err.Error() != errcode.ErrNotFound.Error() {
		t.Fatalf("expected image not found error, got %v", err)
	}
}

func TestServiceDeleteChallengeWithRunningInstances(t *testing.T) {
	db := setupTestDB(t)

	// 创建靶场和运行中的实例
	challenge := &model.Challenge{Title: "Test", Status: "draft"}
	db.Create(challenge)
	db.Create(&model.Instance{ChallengeID: challenge.ID, Status: "running"})

	repo := NewRepository(db)
	service := newTestService(repo, nil)

	err := service.DeleteChallenge(challenge.ID)
	if err == nil || err.Error() != errcode.ErrConflict.Error() {
		t.Fatalf("expected running instances error, got %v", err)
	}
}

func TestServicePublishChallengeNoImage(t *testing.T) {
	db := setupTestDB(t)

	challenge := &model.Challenge{Title: "Test", ImageID: 0}
	db.Create(challenge)

	repo := NewRepository(db)
	service := newTestService(repo, nil)

	err := service.PublishChallenge(challenge.ID)
	if err == nil || err.Error() != errcode.ErrInvalidParams.Error() {
		t.Fatalf("expected no image error, got %v", err)
	}
}

func TestServiceGetPublishedChallengeNotPublished(t *testing.T) {
	db := setupTestDB(t)

	challenge := &model.Challenge{Title: "Test", Status: model.ChallengeStatusDraft}
	db.Create(challenge)

	repo := NewRepository(db)
	service := newTestService(repo, nil)

	_, err := service.GetPublishedChallenge(1, challenge.ID)
	if err == nil || err.Error() != errcode.ErrForbidden.Error() {
		t.Fatalf("expected not published error, got %v", err)
	}
}

func TestServiceGetChallengeIncludesHintsAndAttachment(t *testing.T) {
	db := setupTestDB(t)

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

	repo := NewRepository(db)
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
