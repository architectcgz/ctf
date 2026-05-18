package queries

import (
	"context"
	"testing"

	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	"ctf-platform/pkg/errcode"
)

type teamContestLookupStub struct{}

func (s *teamContestLookupStub) FindByID(context.Context, int64) (*contestentity.Contest, error) {
	return nil, contestdomain.ErrContestNotFound
}

type teamRepoStub struct {
	findByIDFn              func(context.Context, int64) (*contestentity.Team, error)
	findUserTeamInContestFn func(context.Context, int64, int64) (*contestentity.Team, error)
	getMembersFn            func(context.Context, int64) ([]*contestentity.TeamMember, error)
	findUsersByIDsFn        func(context.Context, []int64) ([]*identitycontracts.User, error)
}

func (s *teamRepoStub) CreateWithMember(context.Context, *contestentity.Team, int64) error {
	return nil
}
func (s *teamRepoStub) FindByID(ctx context.Context, id int64) (*contestentity.Team, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, nil
}
func (s *teamRepoStub) DeleteWithMembers(context.Context, int64) error { return nil }
func (s *teamRepoStub) AddMemberWithLock(context.Context, int64, int64, int64) error {
	return nil
}
func (s *teamRepoStub) RemoveMember(context.Context, int64, int64) error { return nil }
func (s *teamRepoStub) FindContestRegistration(context.Context, int64, int64) (*contestentity.ContestRegistration, error) {
	return nil, nil
}
func (s *teamRepoStub) GetMembers(ctx context.Context, teamID int64) ([]*contestentity.TeamMember, error) {
	if s.getMembersFn != nil {
		return s.getMembersFn(ctx, teamID)
	}
	return []*contestentity.TeamMember{}, nil
}
func (s *teamRepoStub) GetMemberCount(context.Context, int64) (int64, error) { return 0, nil }
func (s *teamRepoStub) ListByContest(context.Context, int64) ([]*contestentity.Team, error) {
	return []*contestentity.Team{}, nil
}
func (s *teamRepoStub) GetMemberCountBatch(context.Context, []int64) (map[int64]int, error) {
	return map[int64]int{}, nil
}
func (s *teamRepoStub) FindUsersByIDs(ctx context.Context, ids []int64) ([]*identitycontracts.User, error) {
	if s.findUsersByIDsFn != nil {
		return s.findUsersByIDsFn(ctx, ids)
	}
	return []*identitycontracts.User{}, nil
}
func (s *teamRepoStub) IsUniqueViolation(error, string) bool { return false }
func (s *teamRepoStub) FindUserTeamInContest(ctx context.Context, userID, contestID int64) (*contestentity.Team, error) {
	if s.findUserTeamInContestFn != nil {
		return s.findUserTeamInContestFn(ctx, userID, contestID)
	}
	return nil, nil
}

func TestTeamServiceListTeamsReturnsContestNotFound(t *testing.T) {
	t.Parallel()

	service := NewTeamService(&teamRepoStub{}, &teamContestLookupStub{})

	_, err := service.ListTeams(context.Background(), 42)
	if err != errcode.ErrContestNotFound {
		t.Fatalf("expected ErrContestNotFound, got %v", err)
	}
}

func TestTeamServiceGetTeamInfoTreatsContestTeamNotFoundAsTeamNotFound(t *testing.T) {
	t.Parallel()

	service := NewTeamService(&teamRepoStub{
		findByIDFn: func(context.Context, int64) (*contestentity.Team, error) {
			return nil, contestports.ErrContestTeamNotFound
		},
	}, &teamContestLookupStub{})

	_, _, err := service.GetTeamInfo(context.Background(), 404)
	if err != errcode.ErrTeamNotFound {
		t.Fatalf("expected ErrTeamNotFound, got %v", err)
	}
}

func TestTeamServiceGetMyTeamTreatsContestUserTeamNotFoundAsNil(t *testing.T) {
	t.Parallel()

	service := NewTeamService(&teamRepoStub{
		findUserTeamInContestFn: func(context.Context, int64, int64) (*contestentity.Team, error) {
			return nil, contestports.ErrContestUserTeamNotFound
		},
	}, &teamContestLookupStub{})

	item, err := service.GetMyTeam(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("GetMyTeam() error = %v", err)
	}
	if item != nil {
		t.Fatalf("expected nil team, got %+v", item)
	}
}
