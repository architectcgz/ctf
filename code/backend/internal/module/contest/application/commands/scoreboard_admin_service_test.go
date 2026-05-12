package commands

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
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

func TestScoreboardAdminServiceFreezeScoreboardCreatesTransitionAndSnapshot(t *testing.T) {
	t.Parallel()

	service, db, redisClient, mini := newScoreboardAdminServiceForTest(t)
	contestID := int64(88)
	now := time.Now().UTC()
	createScoreboardContest(t, db, &model.Contest{
		ID:            contestID,
		Title:         "freeze-check",
		Mode:          model.ContestModeJeopardy,
		Status:        model.ContestStatusRunning,
		StatusVersion: 2,
		StartTime:     now.Add(-time.Hour),
		EndTime:       now.Add(30 * time.Minute),
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err := redisClient.ZAdd(context.Background(), rediskeys.RankContestTeamKey(contestID), redislib.Z{Score: 300, Member: "11"}).Err(); err != nil {
		t.Fatalf("seed live scoreboard: %v", err)
	}

	if err := service.FreezeScoreboard(context.Background(), contestID, 40); err != nil {
		t.Fatalf("FreezeScoreboard() error = %v", err)
	}
	if !mini.Exists(rediskeys.RankContestFrozenKey(contestID)) {
		t.Fatal("expected frozen snapshot to be created")
	}

	var contest model.Contest
	if err := db.First(&contest, contestID).Error; err != nil {
		t.Fatalf("load contest: %v", err)
	}
	if contest.Status != model.ContestStatusFrozen || contest.StatusVersion != 3 {
		t.Fatalf("unexpected frozen contest: %+v", contest)
	}

	var transition model.ContestStatusTransition
	if err := db.Where("contest_id = ? AND status_version = ?", contestID, 3).First(&transition).Error; err != nil {
		t.Fatalf("load transition: %v", err)
	}
	if transition.SideEffectStatus != contestdomain.ContestStatusTransitionSideEffectSucceeded {
		t.Fatalf("unexpected transition record: %+v", transition)
	}
}

func TestScoreboardAdminServiceUnfreezeScoreboardClearsSnapshot(t *testing.T) {
	t.Parallel()

	service, db, redisClient, mini := newScoreboardAdminServiceForTest(t)
	contestID := int64(89)
	now := time.Now().UTC()
	createScoreboardContest(t, db, &model.Contest{
		ID:            contestID,
		Title:         "unfreeze-check",
		Mode:          model.ContestModeJeopardy,
		Status:        model.ContestStatusFrozen,
		StatusVersion: 4,
		StartTime:     now.Add(-time.Hour),
		EndTime:       now.Add(30 * time.Minute),
		FreezeTime:    timePtr(now.Add(-10 * time.Minute)),
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err := redisClient.ZAdd(context.Background(), rediskeys.RankContestFrozenKey(contestID), redislib.Z{Score: 300, Member: "11"}).Err(); err != nil {
		t.Fatalf("seed frozen scoreboard: %v", err)
	}

	if err := service.UnfreezeScoreboard(context.Background(), contestID); err != nil {
		t.Fatalf("UnfreezeScoreboard() error = %v", err)
	}
	if mini.Exists(rediskeys.RankContestFrozenKey(contestID)) {
		t.Fatal("expected frozen snapshot to be cleared")
	}

	var contest model.Contest
	if err := db.First(&contest, contestID).Error; err != nil {
		t.Fatalf("load contest: %v", err)
	}
	if contest.Status != model.ContestStatusRunning || contest.StatusVersion != 5 || contest.FreezeTime != nil {
		t.Fatalf("unexpected unfrozen contest: %+v", contest)
	}
}

func newScoreboardAdminServiceForTest(t *testing.T) (*ScoreboardAdminService, *gorm.DB, *redislib.Client, *miniredis.Miniredis) {
	t.Helper()

	db := contesttestsupport.SetupContestTestDB(t)
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})
	service := NewScoreboardAdminService(contestinfra.NewRepository(db), redisClient, nil)
	service.SetStatusSideEffectStore(contestinfra.NewContestStatusSideEffectStore(redisClient))
	return service, db, redisClient, mini
}

func createScoreboardContest(t *testing.T, db *gorm.DB, contest *model.Contest) {
	t.Helper()

	if err := db.Create(contest).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
}

func timePtr(v time.Time) *time.Time {
	return &v
}
