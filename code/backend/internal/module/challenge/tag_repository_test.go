package challenge

import (
	"ctf-platform/internal/model"
	"testing"
)

func TestTagRepositoryCreate(t *testing.T) {
	db := setupTagTestDB(t)
	repo := NewTagRepository(db)

	tag := &model.Tag{Name: "SQL注入", Dimension: "技术"}
	err := repo.Create(tag)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if tag.ID == 0 {
		t.Fatal("ID should be set")
	}
}

func TestTagRepositoryList(t *testing.T) {
	db := setupTagTestDB(t)
	repo := NewTagRepository(db)

	db.Create(&model.Tag{Name: "SQL注入", Dimension: "技术"})
	db.Create(&model.Tag{Name: "XSS", Dimension: "技术"})

	tags, err := repo.List("技术")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("unexpected count: %d", len(tags))
	}
}

func TestTagRepositoryAttachToChallenge(t *testing.T) {
	db := setupTagTestDB(t)
	repo := NewTagRepository(db)

	err := repo.AttachToChallenge(1, 1)
	if err != nil {
		t.Fatalf("AttachToChallenge() error = %v", err)
	}
}
