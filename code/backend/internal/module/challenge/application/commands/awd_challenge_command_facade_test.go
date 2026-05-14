package commands

import (
	"context"
	"testing"

	"ctf-platform/internal/model"
)

func TestNewAWDChallengeCommandFacadeUsesProvidedImportService(t *testing.T) {
	importService := &AWDChallengeImportService{}

	facade := NewAWDChallengeCommandFacade(awdChallengeCommandFacadeRepoStub{}, importService)

	if facade == nil {
		t.Fatal("expected facade")
	}
	if facade.core == nil {
		t.Fatal("expected core service")
	}
	if facade.imports != importService {
		t.Fatalf("expected provided import service reused, got %+v", facade.imports)
	}
}

type awdChallengeCommandFacadeRepoStub struct{}

func (awdChallengeCommandFacadeRepoStub) CreateAWDChallenge(context.Context, *model.AWDChallenge) error {
	return nil
}

func (awdChallengeCommandFacadeRepoStub) FindAWDChallengeByID(context.Context, int64) (*model.AWDChallenge, error) {
	return nil, nil
}

func (awdChallengeCommandFacadeRepoStub) UpdateAWDChallenge(context.Context, *model.AWDChallenge) error {
	return nil
}

func (awdChallengeCommandFacadeRepoStub) DeleteAWDChallenge(context.Context, int64) error {
	return nil
}
