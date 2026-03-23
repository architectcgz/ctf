package infrastructure

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/testsupport"
	"testing"
)

func TestRepositoryCreate(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := NewRepository(db)

	challenge := &model.Challenge{Title: "Test", Status: "draft"}
	err := repo.Create(challenge)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if challenge.ID == 0 {
		t.Fatal("ID should be set")
	}
}

func TestRepositoryFindByID(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := NewRepository(db)

	challenge := &model.Challenge{Title: "Test"}
	db.Create(challenge)

	found, err := repo.FindByID(challenge.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.Title != "Test" {
		t.Fatalf("unexpected title: %s", found.Title)
	}
}

func TestRepositoryList(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := NewRepository(db)

	db.Create(&model.Challenge{Title: "C1", Category: "web"})
	db.Create(&model.Challenge{Title: "C2", Category: "pwn"})

	challenges, total, err := repo.List(&dto.ChallengeQuery{Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 2 {
		t.Fatalf("unexpected total: %d", total)
	}
	if len(challenges) != 2 {
		t.Fatalf("unexpected count: %d", len(challenges))
	}
}

func TestRepositoryHasRunningInstances(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := NewRepository(db)

	challenge := &model.Challenge{Title: "Test"}
	db.Create(challenge)
	db.Create(&model.Instance{ChallengeID: challenge.ID, Status: "running"})

	has, err := repo.HasRunningInstances(challenge.ID)
	if err != nil {
		t.Fatalf("HasRunningInstances() error = %v", err)
	}
	if !has {
		t.Fatal("should have running instances")
	}
}
