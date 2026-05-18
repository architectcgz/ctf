package queries

import (
	"context"
	"testing"

	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

type awdQueryContestLookupStub struct {
	contest *contestentity.Contest
	err     error
}

func (s awdQueryContestLookupStub) FindByID(context.Context, int64) (*contestentity.Contest, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.contest, nil
}

type awdQueryRepositoryStub struct {
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

	findRoundByContestAndIDFn func(context.Context, int64, int64) (*contestentity.AWDRound, error)
	findRunningRoundFn        func(context.Context, int64) (*contestentity.AWDRound, error)
	findContestTeamByMemberFn func(context.Context, int64, int64) (*contestentity.Team, error)
}

func (s awdQueryRepositoryStub) FindRoundByContestAndID(ctx context.Context, contestID, roundID int64) (*contestentity.AWDRound, error) {
	return s.findRoundByContestAndIDFn(ctx, contestID, roundID)
}

func (s awdQueryRepositoryStub) FindRunningRound(ctx context.Context, contestID int64) (*contestentity.AWDRound, error) {
	return s.findRunningRoundFn(ctx, contestID)
}

func (s awdQueryRepositoryStub) FindContestTeamByMember(ctx context.Context, contestID, userID int64) (*contestentity.Team, error) {
	return s.findContestTeamByMemberFn(ctx, contestID, userID)
}

func TestAWDServiceGetRoundSummaryMapsRoundNotFoundToErrNotFound(t *testing.T) {
	t.Parallel()

	service := NewAWDService(
		awdQueryRepositoryStub{
			findRoundByContestAndIDFn: func(context.Context, int64, int64) (*contestentity.AWDRound, error) {
				return nil, contestports.ErrContestAWDRoundNotFound
			},
		},
		awdQueryContestLookupStub{
			contest: &contestentity.Contest{ID: 1, Mode: contestentity.ContestModeAWD},
		},
	)

	_, err := service.GetRoundSummary(context.Background(), 1, 2)
	if err != errcode.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAWDServiceGetUserWorkspaceTreatsCurrentRoundAndTeamNotFoundAsEmptyState(t *testing.T) {
	t.Parallel()

	service := NewAWDService(
		awdQueryRepositoryStub{
			findRunningRoundFn: func(context.Context, int64) (*contestentity.AWDRound, error) {
				return nil, contestports.ErrContestAWDRoundNotFound
			},
			findContestTeamByMemberFn: func(context.Context, int64, int64) (*contestentity.Team, error) {
				return nil, contestports.ErrContestUserTeamNotFound
			},
		},
		awdQueryContestLookupStub{
			contest: &contestentity.Contest{ID: 1, Mode: contestentity.ContestModeAWD},
		},
	)

	resp, err := service.GetUserWorkspace(context.Background(), 101, 1)
	if err != nil {
		t.Fatalf("GetUserWorkspace() error = %v", err)
	}
	if resp.CurrentRound != nil {
		t.Fatalf("expected nil current round, got %+v", resp.CurrentRound)
	}
	if resp.MyTeam != nil {
		t.Fatalf("expected nil my team, got %+v", resp.MyTeam)
	}
	if len(resp.Services) != 0 || len(resp.Targets) != 0 || len(resp.RecentEvents) != 0 {
		t.Fatalf("expected empty workspace state, got %+v", resp)
	}
}
