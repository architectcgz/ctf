package queries

import (
	"context"
	"testing"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

type teamContestLookupStub struct{}

func (s *teamContestLookupStub) FindByID(context.Context, int64) (*model.Contest, error) {
	return nil, contestdomain.ErrContestNotFound
}

type teamRepoStub struct {
	findByIDFn              func(context.Context, int64) (*model.Team, error)
	findUserTeamInContestFn func(context.Context, int64, int64) (*model.Team, error)
	getMembersFn            func(context.Context, int64) ([]*model.TeamMember, error)
	findUsersByIDsFn        func(context.Context, []int64) ([]*model.User, error)
}

func (s *teamRepoStub) CreateWithMember(context.Context, *model.Team, int64) error { return nil }
func (s *teamRepoStub) FindByID(ctx context.Context, id int64) (*model.Team, error) {
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
func (s *teamRepoStub) FindContestRegistration(context.Context, int64, int64) (*model.ContestRegistration, error) {
	return nil, nil
}
func (s *teamRepoStub) GetMembers(ctx context.Context, teamID int64) ([]*model.TeamMember, error) {
	if s.getMembersFn != nil {
		return s.getMembersFn(ctx, teamID)
	}
	return []*model.TeamMember{}, nil
}
func (s *teamRepoStub) GetMemberCount(context.Context, int64) (int64, error) { return 0, nil }
func (s *teamRepoStub) ListByContest(context.Context, int64) ([]*model.Team, error) {
	return []*model.Team{}, nil
}
func (s *teamRepoStub) GetMemberCountBatch(context.Context, []int64) (map[int64]int, error) {
	return map[int64]int{}, nil
}
func (s *teamRepoStub) FindUsersByIDs(ctx context.Context, ids []int64) ([]*model.User, error) {
	if s.findUsersByIDsFn != nil {
		return s.findUsersByIDsFn(ctx, ids)
	}
	return []*model.User{}, nil
}
func (s *teamRepoStub) IsUniqueViolation(error, string) bool { return false }
func (s *teamRepoStub) FindUserTeamInContest(ctx context.Context, userID, contestID int64) (*model.Team, error) {
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
		findByIDFn: func(context.Context, int64) (*model.Team, error) {
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
		findUserTeamInContestFn: func(context.Context, int64, int64) (*model.Team, error) {
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
