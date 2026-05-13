package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type teamFinderSourceStub struct {
	findUserTeamInContestFn func(context.Context, int64, int64) (*model.Team, error)
}

func (s teamFinderSourceStub) FindUserTeamInContest(ctx context.Context, userID, contestID int64) (*model.Team, error) {
	if s.findUserTeamInContestFn != nil {
		return s.findUserTeamInContestFn(ctx, userID, contestID)
	}
	return &model.Team{}, nil
}

func TestTeamFinderRepositoryMapsNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewTeamFinderRepository(teamFinderSourceStub{
		findUserTeamInContestFn: func(context.Context, int64, int64) (*model.Team, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindUserTeamInContest(context.Background(), 1, 2); !errors.Is(err, contestports.ErrContestUserTeamNotFound) {
		t.Fatalf("error = %v, want %v", err, contestports.ErrContestUserTeamNotFound)
	}
}
