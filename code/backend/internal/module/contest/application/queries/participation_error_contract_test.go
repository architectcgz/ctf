package queries

import (
	"context"
	"testing"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type participationQueryRepoStub struct {
	findRegistrationFn func(context.Context, int64, int64) (*model.ContestRegistration, error)
}

func (s participationQueryRepoStub) FindRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
	if s.findRegistrationFn != nil {
		return s.findRegistrationFn(ctx, contestID, userID)
	}
	return nil, contestports.ErrContestParticipationRegistrationNotFound
}

func (s participationQueryRepoStub) FindRegistrationByID(context.Context, int64, int64) (*model.ContestRegistration, error) {
	return nil, contestports.ErrContestParticipationRegistrationNotFound
}

func (s participationQueryRepoStub) ListRegistrations(context.Context, int64, *string, int, int) ([]*contestports.ContestParticipationRegistrationRow, int64, error) {
	return nil, 0, nil
}

func (s participationQueryRepoStub) ListAnnouncements(context.Context, int64) ([]*model.ContestAnnouncement, error) {
	return nil, nil
}

func (s participationQueryRepoStub) ListSolvedProgress(context.Context, int64, int64) ([]*contestports.ContestParticipationSolvedProgressRow, error) {
	return nil, nil
}

type participationQueryContestLookupStub struct{}

func (participationQueryContestLookupStub) FindByID(context.Context, int64) (*model.Contest, error) {
	return &model.Contest{ID: 10, Status: model.ContestStatusRunning}, nil
}

func (participationQueryContestLookupStub) List(context.Context, *string, int, int) ([]*model.Contest, int64, error) {
	return nil, 0, nil
}

type participationQueryTeamFinderStub struct {
	findUserTeamInContestFn func(context.Context, int64, int64) (*model.Team, error)
}

func (s participationQueryTeamFinderStub) FindUserTeamInContest(ctx context.Context, userID, contestID int64) (*model.Team, error) {
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
			findRegistrationFn: func(context.Context, int64, int64) (*model.ContestRegistration, error) {
				return nil, contestports.ErrContestParticipationRegistrationNotFound
			},
		},
		participationQueryTeamFinderStub{
			findUserTeamInContestFn: func(context.Context, int64, int64) (*model.Team, error) {
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
