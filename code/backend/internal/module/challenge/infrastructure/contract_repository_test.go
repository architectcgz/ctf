package infrastructure

import (
	"context"
	"errors"
	"testing"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	"ctf-platform/internal/module/challenge/testsupport"
)

func TestContractRepositoryMapsFindByIDNotFoundToChallengeContractError(t *testing.T) {
	t.Parallel()

	db := testsupport.SetupTestDB(t)
	repo := NewContractRepository(NewRepository(db))

	if _, err := repo.FindByID(context.Background(), 404); !errors.Is(err, challengecontracts.ErrChallengeNotFound) {
		t.Fatalf("error = %v, want %v", err, challengecontracts.ErrChallengeNotFound)
	}
}

func TestContractRepositoryPassesThroughFindByIDInfrastructureErrors(t *testing.T) {
	t.Parallel()

	db := testsupport.SetupTestDB(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB(): %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}
	repo := NewContractRepository(NewRepository(db))

	_, err = repo.FindByID(context.Background(), 1)
	if err == nil {
		t.Fatal("expected closed database error")
	}
	if errors.Is(err, challengecontracts.ErrChallengeNotFound) {
		t.Fatalf("error = %v, must not map non-not-found infrastructure error to %v", err, challengecontracts.ErrChallengeNotFound)
	}
}
