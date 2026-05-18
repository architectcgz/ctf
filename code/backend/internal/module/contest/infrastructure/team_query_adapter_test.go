package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

type teamQueryAdapterSourceStub struct {
	findByIDFn              func(context.Context, int64) (*contestentity.Team, error)
	findUserTeamInContestFn func(context.Context, int64, int64) (*contestentity.Team, error)
}

func (s teamQueryAdapterSourceStub) FindByID(ctx context.Context, id int64) (*contestentity.Team, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return &contestentity.Team{}, nil
}

func (s teamQueryAdapterSourceStub) FindUserTeamInContest(ctx context.Context, userID, contestID int64) (*contestentity.Team, error) {
	if s.findUserTeamInContestFn != nil {
		return s.findUserTeamInContestFn(ctx, userID, contestID)
	}
	return &contestentity.Team{}, nil
}

func (s teamQueryAdapterSourceStub) GetMembers(context.Context, int64) ([]*contestentity.TeamMember, error) {
	return nil, nil
}

func (s teamQueryAdapterSourceStub) AddMemberWithLock(context.Context, int64, int64, int64) error {
	return nil
}

func (s teamQueryAdapterSourceStub) RemoveMember(context.Context, int64, int64) error {
	return nil
}

func (s teamQueryAdapterSourceStub) GetMemberCount(context.Context, int64) (int64, error) {
	return 0, nil
}

func (s teamQueryAdapterSourceStub) GetMemberCountBatch(context.Context, []int64) (map[int64]int, error) {
	return map[int64]int{}, nil
}

func (s teamQueryAdapterSourceStub) ListByContest(context.Context, int64) ([]*contestentity.Team, error) {
	return nil, nil
}

func (s teamQueryAdapterSourceStub) FindUsersByIDs(context.Context, []int64) ([]*model.User, error) {
	return nil, nil
}

func TestTeamQueryAdapterMapsFindByIDNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewTeamQueryAdapter(teamQueryAdapterSourceStub{
		findByIDFn: func(context.Context, int64) (*contestentity.Team, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindByID(context.Background(), 1); !errors.Is(err, contestports.ErrContestTeamNotFound) {
		t.Fatalf("error = %v, want %v", err, contestports.ErrContestTeamNotFound)
	}
}

func TestTeamQueryAdapterMapsFindUserTeamInContestNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewTeamQueryAdapter(teamQueryAdapterSourceStub{
		findUserTeamInContestFn: func(context.Context, int64, int64) (*contestentity.Team, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindUserTeamInContest(context.Background(), 1, 2); !errors.Is(err, contestports.ErrContestUserTeamNotFound) {
		t.Fatalf("error = %v, want %v", err, contestports.ErrContestUserTeamNotFound)
	}
}
