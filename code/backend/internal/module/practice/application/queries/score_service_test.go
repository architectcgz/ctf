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
	practiceqry "ctf-platform/internal/module/practice/application/queries"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	"ctf-platform/internal/module/practice/testsupport"
)

func newTestScoreQueryService(db *gorm.DB, redisClient *redis.Client) *practiceqry.ScoreService {
	_ = redisClient
	return practiceqry.NewScoreService(practiceinfra.NewRepository(db), zap.NewNop(), &config.ScoreConfig{
		CacheTTL:        time.Minute,
		LockTimeout:     5 * time.Second,
		MaxRankingLimit: 100,
	})
}

func TestScoreServiceGetUserScoreWithContextHonorsCancellation(t *testing.T) {
	db := testsupport.SetupScoreServiceTestDB(t)
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	service := newTestScoreQueryService(db, redisClient)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.GetUserScoreWithContext(ctx, 1)
	if err != context.Canceled {
		t.Fatalf("expected context canceled, got %v", err)
	}
}
