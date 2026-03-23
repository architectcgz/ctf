package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	practiceapp "ctf-platform/internal/module/practice/application"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	"ctf-platform/internal/module/practice/testsupport"
)

func newTestScoreService(db *gorm.DB, redisClient *redis.Client) *practiceapp.ScoreService {
	return practiceapp.NewScoreService(practiceinfra.NewRepository(db), redisClient, zap.NewNop(), &config.ScoreConfig{
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

	service := newTestScoreService(db, redisClient)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.GetUserScoreWithContext(ctx, 1)
	if err != context.Canceled {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestScoreServiceUpdateUserScoreWithContextHonorsCancellation(t *testing.T) {
	db := testsupport.SetupScoreServiceTestDB(t)
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
