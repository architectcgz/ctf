package commands_test

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
	practicecmd "ctf-platform/internal/module/practice/application/commands"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	"ctf-platform/internal/module/practice/testsupport"
	"ctf-platform/internal/pkg/cache"
)

func newTestScoreCommandService(db *gorm.DB, redisClient *redis.Client) *practicecmd.ScoreService {
	return practicecmd.NewScoreService(practiceinfra.NewRepository(db), practiceinfra.NewScoreStateStore(redisClient), zap.NewNop(), &config.ScoreConfig{
		CacheTTL:        time.Minute,
		LockTimeout:     5 * time.Second,
		MaxRankingLimit: 100,
	})
}

func TestScoreServiceUpdateUserScoreHonorsCancellation(t *testing.T) {
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

	service := newTestScoreCommandService(db, redisClient)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := service.UpdateUserScore(ctx, 7)
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestScoreServiceCalculateScoreUsesChallengePointsDirectly(t *testing.T) {
	db := testsupport.SetupScoreServiceTestDB(t)
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	now := time.Now()
	if err := db.Create(&model.Challenge{
		ID:         11,
		Title:      "web-2",
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}

	service := newTestScoreCommandService(db, redisClient)

	score := service.CalculateScore(context.Background(), 11)
	if score != 100 {
		t.Fatalf("expected direct challenge points 100, got %d", score)
	}
}

func TestScoreServiceUpdateUserScoreUsesSolvedChallengePointsSum(t *testing.T) {
	db := testsupport.SetupScoreServiceTestDB(t)
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	now := time.Now()
	if err := db.Create([]*model.User{
		{ID: 9, Username: "student09", CreatedAt: now, UpdatedAt: now},
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Create([]*model.Challenge{
		{
			ID:         21,
			Title:      "easy-web",
			Difficulty: model.ChallengeDifficultyEasy,
			Points:     100,
			Status:     model.ChallengeStatusPublished,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         22,
			Title:      "hard-pwn",
			Difficulty: model.ChallengeDifficultyHard,
			Points:     300,
			Status:     model.ChallengeStatusPublished,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}).Error; err != nil {
		t.Fatalf("seed challenges: %v", err)
	}
	if err := db.Create([]*model.Submission{
		{UserID: 9, ChallengeID: 21, IsCorrect: true, SubmittedAt: now},
		{UserID: 9, ChallengeID: 22, IsCorrect: true, SubmittedAt: now.Add(time.Minute)},
	}).Error; err != nil {
		t.Fatalf("seed submissions: %v", err)
	}

	service := newTestScoreCommandService(db, redisClient)

	if err := service.UpdateUserScore(context.Background(), 9); err != nil {
		t.Fatalf("UpdateUserScore() error = %v", err)
	}

	var userScore model.UserScore
	if err := db.First(&userScore, "user_id = ?", 9).Error; err != nil {
		t.Fatalf("load user score: %v", err)
	}
	if userScore.TotalScore != 400 {
		t.Fatalf("expected total_score 400, got %+v", userScore)
	}
	if userScore.SolvedCount != 2 {
		t.Fatalf("expected solved_count 2, got %+v", userScore)
	}

	rankingScore, err := redisClient.ZScore(context.Background(), cache.RankingKey(), "9").Result()
	if err != nil {
		t.Fatalf("load ranking score: %v", err)
	}
	if rankingScore != 400 {
		t.Fatalf("expected ranking score 400, got %v", rankingScore)
	}
}
