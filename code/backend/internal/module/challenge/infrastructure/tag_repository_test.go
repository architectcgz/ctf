package infrastructure

import (
	"context"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/testsupport"
	"testing"
)

func TestTagRepositoryCreate(t *testing.T) {
	db := testsupport.SetupTagTestDB(t)
	repo := NewTagRepository(db)

	tag := &model.Tag{Name: "SQL注入", Type: model.TagTypeVulnerability}
	err := repo.Create(context.Background(), tag)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if tag.ID == 0 {
		t.Fatal("ID should be set")
	}
}

func TestTagRepositoryList(t *testing.T) {
	db := testsupport.SetupTagTestDB(t)
	repo := NewTagRepository(db)

	db.Create(&model.Tag{Name: "SQL注入", Type: model.TagTypeVulnerability})
	db.Create(&model.Tag{Name: "XSS", Type: model.TagTypeVulnerability})

	tags, err := repo.List(context.Background(), model.TagTypeVulnerability)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("unexpected count: %d", len(tags))
	}
}

func TestTagRepositoryAttachToChallenge(t *testing.T) {
	db := testsupport.SetupTagTestDB(t)
	repo := NewTagRepository(db)

	db.Create(&model.Challenge{ID: 1, Title: "test", Status: model.ChallengeStatusDraft})
	db.Create(&model.Tag{ID: 1, Name: "SQL注入", Type: model.TagTypeVulnerability})

	err := repo.AttachTagsInTx(context.Background(), 1, []int64{1})
	if err != nil {
		t.Fatalf("AttachTagsInTx() error = %v", err)
	}
}
