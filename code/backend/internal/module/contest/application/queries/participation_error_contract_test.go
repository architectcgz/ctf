package queries

import (
	"context"
	"testing"

	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

type participationQueryRepoStub struct {
	findRegistrationFn func(context.Context, int64, int64) (*contestentity.ContestRegistration, error)
}

func (s participationQueryRepoStub) FindRegistration(ctx context.Context, contestID, userID int64) (*contestentity.ContestRegistration, error) {
	if s.findRegistrationFn != nil {
		return s.findRegistrationFn(ctx, contestID, userID)
	}
	return nil, contestports.ErrContestParticipationRegistrationNotFound
}

func (s participationQueryRepoStub) FindRegistrationByID(context.Context, int64, int64) (*contestentity.ContestRegistration, error) {
	return nil, contestports.ErrContestParticipationRegistrationNotFound
}

func (s participationQueryRepoStub) ListRegistrations(context.Context, int64, *string, int, int) ([]*contestports.ContestParticipationRegistrationRow, int64, error) {
	return nil, 0, nil
}

func (s participationQueryRepoStub) ListAnnouncements(context.Context, int64) ([]*contestentity.ContestAnnouncement, error) {
	return nil, nil
}

func (s participationQueryRepoStub) ListSolvedProgress(context.Context, int64, int64) ([]*contestports.ContestParticipationSolvedProgressRow, error) {
	return nil, nil
}

type participationQueryContestLookupStub struct{}

func (participationQueryContestLookupStub) FindByID(context.Context, int64) (*contestentity.Contest, error) {
	return &contestentity.Contest{ID: 10, Status: contestentity.ContestStatusRunning}, nil
}

func (participationQueryContestLookupStub) List(context.Context, *string, int, int) ([]*contestentity.Contest, int64, error) {
	return nil, 0, nil
}

type participationQueryTeamFinderStub struct {
	findUserTeamInContestFn func(context.Context, int64, int64) (*contestentity.Team, error)
}

func (s participationQueryTeamFinderStub) FindUserTeamInContest(ctx context.Context, userID, contestID int64) (*contestentity.Team, error) {
	if s.findUserTeamInContestFn != nil {
		return s.findUserTeamInContestFn(ctx, userID, contestID)
	}
	return nil, contestports.ErrContestUserTeamNotFound
}

func TestParticipationServiceResolveUserTeamIDTreatsMissingRegistrationAndTeamAsNoTeam(t *testing.T) {
	t.Parallel()

	service := NewParticipationService(
		participationQueryContestLookupStub{},
		participationQueryRepoStub{
			findRegistrationFn: func(context.Context, int64, int64) (*contestentity.ContestRegistration, error) {
				return nil, contestports.ErrContestParticipationRegistrationNotFound
			},
		},
		participationQueryTeamFinderStub{
			findUserTeamInContestFn: func(context.Context, int64, int64) (*contestentity.Team, error) {
				return nil, contestports.ErrContestUserTeamNotFound
			},
		},
	)

	teamID, err := service.resolveUserTeamID(context.Background(), 10, 1001)
	if err != nil {
		t.Fatalf("resolveUserTeamID() error = %v", err)
	}
	if teamID != nil {
		t.Fatalf("expected nil team id, got %v", *teamID)
	}
}
