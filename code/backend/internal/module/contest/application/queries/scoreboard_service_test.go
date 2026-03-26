package queries

import (
	"context"
	"testing"

	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

type scoreboardRepoStub struct{}

func (s *scoreboardRepoStub) FindByID(context.Context, int64) (*model.Contest, error) {
	return nil, contestdomain.ErrContestNotFound
}

func (s *scoreboardRepoStub) FindTeamsByIDs(context.Context, []int64) ([]*model.Team, error) {
	return nil, nil
}

func (s *scoreboardRepoStub) FindScoreboardTeamStats(context.Context, int64, string, []int64) (map[int64]contestports.ScoreboardTeamStats, error) {
	return nil, nil
}

func TestScoreboardServiceGetScoreboardReturnsContestNotFound(t *testing.T) {
	t.Parallel()

	redisClient := redislib.NewClient(&redislib.Options{Addr: "127.0.0.1:0"})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := NewScoreboardService(&scoreboardRepoStub{}, redisClient, &config.ContestConfig{}, nil)

	_, err := service.GetScoreboard(context.Background(), 42, 1, 20)
	if err != errcode.ErrContestNotFound {
		t.Fatalf("expected ErrContestNotFound, got %v", err)
	}
}
