package commands

import (
	"context"
	"errors"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	rediskeys "ctf-platform/internal/pkg/redis"
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

type scoreboardAdminMutableRepoStub struct {
	contest  *model.Contest
	updateFn func(context.Context, *model.Contest) error
}

func (s *scoreboardAdminMutableRepoStub) FindByID(context.Context, int64) (*model.Contest, error) {
	return s.contest, nil
}

func (s *scoreboardAdminMutableRepoStub) Update(ctx context.Context, contest *model.Contest) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, contest)
	}
	return nil
}

func (s *scoreboardAdminMutableRepoStub) FindTeamsByContest(context.Context, int64) ([]*model.Team, error) {
	return []*model.Team{}, nil
}

func TestScoreboardAdminServiceFreezeScoreboardRollsBackSnapshotWhenUpdateFails(t *testing.T) {
	t.Parallel()

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mini.Close()

	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	defer func() {
		_ = redisClient.Close()
	}()

	contestID := int64(88)
	now := time.Now()
	liveKey := rediskeys.RankContestTeamKey(contestID)
	frozenKey := rediskeys.RankContestFrozenKey(contestID)
	if err := redisClient.ZAdd(context.Background(), liveKey, redislib.Z{Score: 300, Member: "11"}).Err(); err != nil {
		t.Fatalf("seed live scoreboard: %v", err)
	}

	updateErr := errors.New("write contest freeze state failed")
	repo := &scoreboardAdminMutableRepoStub{
		contest: &model.Contest{
			ID:        contestID,
			Title:     "freeze-check",
			StartTime: now.Add(-time.Hour),
			EndTime:   now.Add(30 * time.Minute),
			Status:    model.ContestStatusRunning,
		},
		updateFn: func(context.Context, *model.Contest) error {
			return updateErr
		},
	}

	service := NewScoreboardAdminService(repo, redisClient, nil)
	err = service.FreezeScoreboard(context.Background(), contestID, 40)
	if !errors.Is(err, updateErr) {
		t.Fatalf("expected update error, got %v", err)
	}

	exists, err := redisClient.Exists(context.Background(), frozenKey).Result()
	if err != nil {
		t.Fatalf("check frozen snapshot: %v", err)
	}
	if exists != 0 {
		t.Fatalf("expected frozen snapshot to be rolled back on update failure")
	}
}
