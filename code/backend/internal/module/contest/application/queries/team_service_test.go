package queries

import (
	"context"
	"testing"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

type teamContestLookupStub struct{}

func (s *teamContestLookupStub) FindByID(context.Context, int64) (*model.Contest, error) {
	return nil, contestdomain.ErrContestNotFound
}

type teamRepoStub struct{}

func (s *teamRepoStub) CreateWithMember(context.Context, *model.Team, int64) error { return nil }
func (s *teamRepoStub) FindByID(context.Context, int64) (*model.Team, error)       { return nil, nil }
func (s *teamRepoStub) DeleteWithMembers(context.Context, int64) error             { return nil }
func (s *teamRepoStub) AddMemberWithLock(context.Context, int64, int64, int64) error {
	return nil
}
func (s *teamRepoStub) RemoveMember(context.Context, int64, int64) error { return nil }
func (s *teamRepoStub) FindContestRegistration(context.Context, int64, int64) (*model.ContestRegistration, error) {
	return nil, nil
}
func (s *teamRepoStub) GetMembers(context.Context, int64) ([]*model.TeamMember, error) {
	return []*model.TeamMember{}, nil
}
func (s *teamRepoStub) GetMemberCount(context.Context, int64) (int64, error) { return 0, nil }
func (s *teamRepoStub) ListByContest(context.Context, int64) ([]*model.Team, error) {
	return []*model.Team{}, nil
}
func (s *teamRepoStub) GetMemberCountBatch(context.Context, []int64) (map[int64]int, error) {
	return map[int64]int{}, nil
}
func (s *teamRepoStub) FindUsersByIDs(context.Context, []int64) ([]*model.User, error) {
	return []*model.User{}, nil
}
func (s *teamRepoStub) IsUniqueViolation(error, string) bool { return false }
func (s *teamRepoStub) FindUserTeamInContest(context.Context, int64, int64) (*model.Team, error) {
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
