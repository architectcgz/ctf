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
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	practicecmd "ctf-platform/internal/module/practice/application/commands"
	practiceentity "ctf-platform/internal/module/practice/entity"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	practicecachekeys "ctf-platform/internal/module/practice/infrastructure/cachekeys"
	"ctf-platform/internal/module/practice/testsupport"
	contestentity "ctf-platform/internal/module/practice/testsupport/contestentity"
	"ctf-platform/internal/shared/taxonomy"
)

func newTestScoreCommandService(db *gorm.DB, redisClient *redis.Client) *practicecmd.ScoreService {
	return practicecmd.NewScoreService(newPracticeRepositoryWithRuntimePortOwner(db), practiceinfra.NewScoreStateStore(redisClient), zap.NewNop(), &config.ScoreConfig{
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
	if err := db.Create(&practiceCommandChallengeRow{
		ID:         1,
		Title:      "web-1",
		Difficulty: taxonomy.DifficultyEasy,
		Points:     100,
		Status:     challengecontracts.ChallengeStatusPublished,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}
	if err := db.Create(&contestentity.Submission{
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
	if err := db.Create(&practiceCommandChallengeRow{
		ID:         11,
		Title:      "web-2",
		Difficulty: taxonomy.DifficultyEasy,
		Points:     100,
		Status:     challengecontracts.ChallengeStatusPublished,
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
	if err := db.Create([]*identitycontracts.User{
		{ID: 9, Username: "student09", CreatedAt: now, UpdatedAt: now},
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Create([]*practiceCommandChallengeRow{
		{
			ID:         21,
			Title:      "easy-web",
			Difficulty: taxonomy.DifficultyEasy,
			Points:     100,
			Status:     challengecontracts.ChallengeStatusPublished,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         22,
			Title:      "hard-pwn",
			Difficulty: taxonomy.DifficultyHard,
			Points:     300,
			Status:     challengecontracts.ChallengeStatusPublished,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}).Error; err != nil {
		t.Fatalf("seed challenges: %v", err)
	}
	if err := db.Create([]*contestentity.Submission{
		{UserID: 9, ChallengeID: 21, IsCorrect: true, SubmittedAt: now},
		{UserID: 9, ChallengeID: 22, IsCorrect: true, SubmittedAt: now.Add(time.Minute)},
	}).Error; err != nil {
		t.Fatalf("seed submissions: %v", err)
	}

	service := newTestScoreCommandService(db, redisClient)

	if err := service.UpdateUserScore(context.Background(), 9); err != nil {
		t.Fatalf("UpdateUserScore() error = %v", err)
	}

	var userScore practiceentity.UserScore
	if err := db.First(&userScore, "user_id = ?", 9).Error; err != nil {
		t.Fatalf("load user score: %v", err)
	}
	if userScore.TotalScore != 400 {
		t.Fatalf("expected total_score 400, got %+v", userScore)
	}
	if userScore.SolvedCount != 2 {
		t.Fatalf("expected solved_count 2, got %+v", userScore)
	}

	rankingScore, err := redisClient.ZScore(context.Background(), practicecachekeys.RankingKey(), "9").Result()
	if err != nil {
		t.Fatalf("load ranking score: %v", err)
	}
	if rankingScore != 400 {
		t.Fatalf("expected ranking score 400, got %v", rankingScore)
	}
}
