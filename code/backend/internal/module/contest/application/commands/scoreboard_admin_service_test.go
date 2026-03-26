package commands

import (
	"context"
	"testing"

	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

type scoreboardAdminRepoStub struct{}

func (s *scoreboardAdminRepoStub) FindByID(context.Context, int64) (*model.Contest, error) {
	return nil, contestdomain.ErrContestNotFound
}

func (s *scoreboardAdminRepoStub) Update(context.Context, *model.Contest) error {
	return nil
}

func (s *scoreboardAdminRepoStub) FindTeamsByContest(context.Context, int64) ([]*model.Team, error) {
	return []*model.Team{}, nil
}

func TestScoreboardAdminServiceFreezeScoreboardReturnsContestNotFound(t *testing.T) {
	t.Parallel()

	redisClient := redislib.NewClient(&redislib.Options{Addr: "127.0.0.1:0"})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := NewScoreboardAdminService(&scoreboardAdminRepoStub{}, redisClient, nil)

	err := service.FreezeScoreboard(context.Background(), 42, 30)
	if err != errcode.ErrContestNotFound {
		t.Fatalf("expected ErrContestNotFound, got %v", err)
	}
}
