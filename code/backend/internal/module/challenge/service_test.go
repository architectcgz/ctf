package challenge

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"testing"
)

func TestServiceCreateChallengeSuccess(t *testing.T) {
	db := setupTestDB(t)

	// 创建测试镜像
	db.Create(&model.Image{ID: 1, Name: "test-image"})

	repo := NewRepository(db)
	imageRepo := NewImageRepository(db)
	service := NewService(repo, imageRepo, nil)

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
	service := NewService(repo, imageRepo, nil)

	_, err := service.CreateChallenge(&dto.CreateChallengeReq{
		ImageID: 999,
	})
	if err == nil || err.Error() != "镜像不存在" {
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
	service := NewService(repo, nil, nil)

	err := service.DeleteChallenge(challenge.ID)
	if err == nil || err.Error() != "存在运行中的实例，无法删除" {
		t.Fatalf("expected running instances error, got %v", err)
	}
}

func TestServicePublishChallengeNoImage(t *testing.T) {
	db := setupTestDB(t)

	challenge := &model.Challenge{Title: "Test", ImageID: 0}
	db.Create(challenge)

	repo := NewRepository(db)
	service := NewService(repo, nil, nil)

	err := service.PublishChallenge(challenge.ID)
	if err == nil || err.Error() != "靶场未关联镜像，无法发布" {
		t.Fatalf("expected no image error, got %v", err)
	}
}

func TestServiceGetPublishedChallengeNotPublished(t *testing.T) {
	db := setupTestDB(t)

	challenge := &model.Challenge{Title: "Test", Status: model.ChallengeStatusDraft}
	db.Create(challenge)

	repo := NewRepository(db)
	service := NewService(repo, nil, nil)

	_, err := service.GetPublishedChallenge(1, challenge.ID)
	if err == nil || err.Error() != "challenge not published" {
		t.Fatalf("expected not published error, got %v", err)
	}
}
