package queries_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	practiceqry "ctf-platform/internal/module/practice/application/queries"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	practiceports "ctf-platform/internal/module/practice/ports"
	"ctf-platform/internal/module/practice/testsupport"
)

type scoreQueryRepoStub struct {
	findUserScoreFn     func(context.Context, int64) (*model.UserScore, error)
	listTopUserScoresFn func(context.Context, int) ([]model.UserScore, error)
	findUsersByIDsFn    func(context.Context, []int64) ([]model.User, error)
}

func (s scoreQueryRepoStub) FindUserScore(ctx context.Context, userID int64) (*model.UserScore, error) {
	if s.findUserScoreFn == nil {
		return nil, nil
	}
	return s.findUserScoreFn(ctx, userID)
}

func (s scoreQueryRepoStub) ListTopUserScores(ctx context.Context, limit int) ([]model.UserScore, error) {
	if s.listTopUserScoresFn == nil {
		return []model.UserScore{}, nil
	}
	return s.listTopUserScoresFn(ctx, limit)
}

func (s scoreQueryRepoStub) FindUsersByIDs(ctx context.Context, userIDs []int64) ([]model.User, error) {
	if s.findUsersByIDsFn == nil {
		return []model.User{}, nil
	}
	return s.findUsersByIDsFn(ctx, userIDs)
}

func newTestScoreQueryService(db *gorm.DB, redisClient *redis.Client) *practiceqry.ScoreService {
	return practiceqry.NewScoreService(practiceinfra.NewScoreQueryRepository(practiceinfra.NewRepository(db)), practiceinfra.NewScoreStateStore(redisClient), zap.NewNop(), &config.ScoreConfig{
		CacheTTL:        time.Minute,
		LockTimeout:     5 * time.Second,
		MaxRankingLimit: 100,
	})
}

func TestScoreServiceGetUserScoreHonorsCancellation(t *testing.T) {
	db := testsupport.SetupScoreServiceTestDB(t)
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	service := newTestScoreQueryService(db, redisClient)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.GetUserScore(ctx, 1)
	if err != context.Canceled {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestScoreServiceGetUserScoreTreatsPracticeScoreNotFoundAsZeroScore(t *testing.T) {
	t.Parallel()

	service := practiceqry.NewScoreService(scoreQueryRepoStub{
		findUserScoreFn: func(context.Context, int64) (*model.UserScore, error) {
			return nil, practiceports.ErrPracticeUserScoreNotFound
		},
	}, nil, zap.NewNop(), &config.ScoreConfig{MaxRankingLimit: 100})

	info, err := service.GetUserScore(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetUserScore error = %v", err)
	}
	if info == nil {
		t.Fatal("GetUserScore returned nil info")
	}
	if info.UserID != 42 || info.TotalScore != 0 || info.SolvedCount != 0 || info.Rank != 0 {
		t.Fatalf("GetUserScore info = %+v, want zero score fallback", info)
	}
}

func TestScoreServiceGetRankingHonorsCancellation(t *testing.T) {
	db := testsupport.SetupScoreServiceTestDB(t)
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	service := newTestScoreQueryService(db, redisClient)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.GetRanking(ctx, 10)
	if err != context.Canceled {
		t.Fatalf("expected context canceled, got %v", err)
	}
}
