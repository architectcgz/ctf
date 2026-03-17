package practice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
)

func setupScoreServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Challenge{}, &model.Submission{}, &model.User{}, &model.UserScore{}); err != nil {
		t.Fatalf("migrate score tables: %v", err)
	}
	return db
}

func newTestScoreService(db *gorm.DB, redisClient *redis.Client) *ScoreService {
	return NewScoreService(NewRepository(db), redisClient, zap.NewNop(), &config.ScoreConfig{
		CacheTTL:        time.Minute,
		LockTimeout:     5 * time.Second,
		MaxRankingLimit: 100,
	})
}

func TestScoreServiceGetUserScoreWithContextHonorsCancellation(t *testing.T) {
	db := setupScoreServiceTestDB(t)
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	service := newTestScoreService(db, redisClient)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.GetUserScoreWithContext(ctx, 1)
	if err != context.Canceled {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestScoreServiceUpdateUserScoreWithContextHonorsCancellation(t *testing.T) {
	db := setupScoreServiceTestDB(t)
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	now := time.Now()
	if err := db.Create(&model.Challenge{
		ID:         1,
		Title:      "web-1",
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}
	if err := db.Create(&model.Submission{
		UserID:      7,
		ChallengeID: 1,
		IsCorrect:   true,
		SubmittedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed submission: %v", err)
	}

	service := newTestScoreService(db, redisClient)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := service.UpdateUserScoreWithContext(ctx, 7)
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}
