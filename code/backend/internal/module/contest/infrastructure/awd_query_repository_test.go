package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type awdQueryRepositorySourceStub struct {
	contestports.AWDRoundStore
	contestports.AWDTeamLookup
	contestports.AWDServiceDefinitionQuery
	contestports.AWDReadinessQuery
	contestports.AWDServiceInstanceQuery
	contestports.AWDDefenseWorkspaceSummaryQuery
	contestports.AWDServiceOperationQuery
	contestports.AWDTeamServiceStore
	contestports.AWDAttackLogStore
	contestports.AWDTrafficEventQuery

	findRoundByContestAndIDFn func(context.Context, int64, int64) (*model.AWDRound, error)
	findRunningRoundFn        func(context.Context, int64) (*model.AWDRound, error)
	findContestTeamByMemberFn func(context.Context, int64, int64) (*model.Team, error)
}

func (s awdQueryRepositorySourceStub) FindRoundByContestAndID(ctx context.Context, contestID, roundID int64) (*model.AWDRound, error) {
	return s.findRoundByContestAndIDFn(ctx, contestID, roundID)
}

func (s awdQueryRepositorySourceStub) FindRunningRound(ctx context.Context, contestID int64) (*model.AWDRound, error) {
	return s.findRunningRoundFn(ctx, contestID)
}

func (s awdQueryRepositorySourceStub) FindContestTeamByMember(ctx context.Context, contestID, userID int64) (*model.Team, error) {
	return s.findContestTeamByMemberFn(ctx, contestID, userID)
}

func TestAWDQueryRepositoryMapsFindRoundByContestAndIDNotFound(t *testing.T) {
	t.Parallel()

	repo := NewAWDQueryRepository(awdQueryRepositorySourceStub{
		findRoundByContestAndIDFn: func(context.Context, int64, int64) (*model.AWDRound, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindRoundByContestAndID(context.Background(), 1, 2); !errors.Is(err, contestports.ErrContestAWDRoundNotFound) {
		t.Fatalf("error = %v, want %v", err, contestports.ErrContestAWDRoundNotFound)
	}
}

func TestAWDQueryRepositoryMapsFindRunningRoundNotFound(t *testing.T) {
	t.Parallel()

	repo := NewAWDQueryRepository(awdQueryRepositorySourceStub{
		findRunningRoundFn: func(context.Context, int64) (*model.AWDRound, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindRunningRound(context.Background(), 1); !errors.Is(err, contestports.ErrContestAWDRoundNotFound) {
		t.Fatalf("error = %v, want %v", err, contestports.ErrContestAWDRoundNotFound)
	}
}

func TestAWDQueryRepositoryMapsFindContestTeamByMemberNotFound(t *testing.T) {
	t.Parallel()

	repo := NewAWDQueryRepository(awdQueryRepositorySourceStub{
		findContestTeamByMemberFn: func(context.Context, int64, int64) (*model.Team, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindContestTeamByMember(context.Background(), 1, 2); !errors.Is(err, contestports.ErrContestUserTeamNotFound) {
		t.Fatalf("error = %v, want %v", err, contestports.ErrContestUserTeamNotFound)
	}
}

func TestAWDQueryRepositoryPassesThroughNonNotFoundErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("boom")
	repo := NewAWDQueryRepository(awdQueryRepositorySourceStub{
		findRoundByContestAndIDFn: func(context.Context, int64, int64) (*model.AWDRound, error) {
			return nil, expectedErr
		},
		findRunningRoundFn: func(context.Context, int64) (*model.AWDRound, error) {
			return nil, expectedErr
		},
		findContestTeamByMemberFn: func(context.Context, int64, int64) (*model.Team, error) {
			return nil, expectedErr
		},
	})

	if _, err := repo.FindRoundByContestAndID(context.Background(), 1, 2); !errors.Is(err, expectedErr) {
		t.Fatalf("round error = %v, want %v", err, expectedErr)
	}
	if _, err := repo.FindRunningRound(context.Background(), 1); !errors.Is(err, expectedErr) {
		t.Fatalf("running round error = %v, want %v", err, expectedErr)
	}
	if _, err := repo.FindContestTeamByMember(context.Background(), 1, 2); !errors.Is(err, expectedErr) {
		t.Fatalf("team error = %v, want %v", err, expectedErr)
	}
}
