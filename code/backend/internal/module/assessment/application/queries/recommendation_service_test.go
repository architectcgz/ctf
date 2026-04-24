package queries_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	assessmentinfra "ctf-platform/internal/module/assessment/infrastructure"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	rediskeys "ctf-platform/internal/pkg/redis"
	platformevents "ctf-platform/internal/platform/events"
)

type stubChallengeRecommendationRepo struct {
	challenges []*model.Challenge
	calls      int
	lastLimit  int
	lastDims   []string
	lastSolved []int64
}

func (s *stubChallengeRecommendationRepo) FindPublishedForRecommendation(_ context.Context, limit int, dimensions []string, excludeSolved []int64) ([]*model.Challenge, error) {
	s.calls++
	s.lastLimit = limit
	s.lastDims = append([]string(nil), dimensions...)
	s.lastSolved = append([]int64(nil), excludeSolved...)
	return s.challenges, nil
}

func setupRecommendationTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.SkillProfile{}, &model.Submission{}, &model.AWDAttackLog{}); err != nil {
		t.Fatalf("migrate recommendation tables: %v", err)
	}
	return db
}

func newRecommendationTestService(db *gorm.DB, challengeRepo assessmentports.ChallengeRepository, redisClient *redis.Client) *assessmentqry.RecommendationService {
	return assessmentqry.NewRecommendationService(
		assessmentinfra.NewRepository(db),
		challengeRepo,
		redisClient,
		config.RecommendationConfig{
			WeakThreshold: 0.4,
			CacheTTL:      time.Hour,
			DefaultLimit:  3,
			MaxLimit:      5,
		},
		zap.NewNop(),
	)
}

func TestRecommendationServiceRecommendChallengesUsesCacheForDefaultLimit(t *testing.T) {
	db := setupRecommendationTestDB(t)
	stubRepo := &stubChallengeRecommendationRepo{}

	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	cached := []*dto.ChallengeRecommendation{
		{ID: 1, Title: "cached-web", Category: model.DimensionWeb, Difficulty: model.ChallengeDifficultyEasy, Points: 100, Reason: "cached"},
	}
	payload, err := json.Marshal(cached)
	if err != nil {
		t.Fatalf("marshal cached recommendations: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.RecommendationKey(1), payload, time.Hour).Err(); err != nil {
		t.Fatalf("seed recommendation cache: %v", err)
	}

	service := newRecommendationTestService(db, stubRepo, redisClient)
	items, err := service.RecommendChallenges(1, 0)
	if err != nil {
		t.Fatalf("RecommendChallenges() error = %v", err)
	}
	if len(items) != 1 || items[0].Title != "cached-web" {
		t.Fatalf("expected cached recommendations, got %+v", items)
	}
	if stubRepo.calls != 0 {
		t.Fatalf("expected challenge repo not called on cache hit, got %d", stubRepo.calls)
	}
}

func TestRecommendationServiceRecommendChallengesUsesWeakDimensionsAndSolvedFilter(t *testing.T) {
	db := setupRecommendationTestDB(t)
	now := time.Now()

	profiles := []model.SkillProfile{
		{UserID: 7, Dimension: model.DimensionWeb, Score: 0.2, UpdatedAt: now},
		{UserID: 7, Dimension: model.DimensionCrypto, Score: 0.8, UpdatedAt: now},
		{UserID: 7, Dimension: model.DimensionPwn, Score: 0.1, UpdatedAt: now},
	}
	for _, profile := range profiles {
		if err := db.Create(&profile).Error; err != nil {
			t.Fatalf("seed profile: %v", err)
		}
	}

	submissions := []model.Submission{
		{UserID: 7, ChallengeID: 101, IsCorrect: true, SubmittedAt: now},
		{UserID: 7, ChallengeID: 202, IsCorrect: false, SubmittedAt: now},
	}
	for _, submission := range submissions {
		if err := db.Create(&submission).Error; err != nil {
			t.Fatalf("seed submission: %v", err)
		}
	}
	if err := db.Create(&model.AWDAttackLog{
		ID:                1,
		RoundID:           701,
		AttackerTeamID:    801,
		VictimTeamID:      802,
		ChallengeID:       303,
		AttackType:        model.AWDAttackTypeFlagCapture,
		Source:            model.AWDAttackSourceSubmission,
		IsSuccess:         true,
		ScoreGained:       80,
		SubmittedByUserID: ptrRecommendationInt64(7),
		CreatedAt:         now,
	}).Error; err != nil {
		t.Fatalf("seed awd attack log: %v", err)
	}

	stubRepo := &stubChallengeRecommendationRepo{
		challenges: []*model.Challenge{
			{ID: 301, Title: "web-fix", Category: model.DimensionWeb, Difficulty: model.ChallengeDifficultyEasy, Points: 120},
			{ID: 302, Title: "pwn-fix", Category: model.DimensionPwn, Difficulty: model.ChallengeDifficultyMedium, Points: 180},
		},
	}
	service := newRecommendationTestService(db, stubRepo, nil)

	items, err := service.RecommendChallenges(7, 99)
	if err != nil {
		t.Fatalf("RecommendChallenges() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 recommendations, got %+v", items)
	}
	if stubRepo.calls != 1 {
		t.Fatalf("expected challenge repo called once, got %d", stubRepo.calls)
	}
	if stubRepo.lastLimit != 5 {
		t.Fatalf("expected limit capped to max limit 5, got %d", stubRepo.lastLimit)
	}
	if len(stubRepo.lastDims) != 2 || stubRepo.lastDims[0] != model.DimensionWeb || stubRepo.lastDims[1] != model.DimensionPwn {
		t.Fatalf("unexpected weak dimensions: %+v", stubRepo.lastDims)
	}
	if len(stubRepo.lastSolved) != 2 || stubRepo.lastSolved[0] != 101 || stubRepo.lastSolved[1] != 303 {
		t.Fatalf("unexpected solved challenge ids: %+v", stubRepo.lastSolved)
	}
	if items[0].Reason == "" || items[1].Reason == "" {
		t.Fatalf("expected recommendation reason generated, got %+v", items)
	}
}

func TestRecommendationServiceRecommendReturnsEmptyWhenNoWeakDimension(t *testing.T) {
	db := setupRecommendationTestDB(t)
	now := time.Now()
	if err := db.Create(&model.SkillProfile{
		UserID:    9,
		Dimension: model.DimensionWeb,
		Score:     0.95,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed profile: %v", err)
	}

	stubRepo := &stubChallengeRecommendationRepo{}
	service := newRecommendationTestService(db, stubRepo, nil)

	resp, err := service.Recommend(9, 0)
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}
	if len(resp.WeakDimensions) != 0 || len(resp.Challenges) != 0 {
		t.Fatalf("expected empty recommendation response, got %+v", resp)
	}
	if stubRepo.calls != 0 {
		t.Fatalf("expected no challenge query when no weak dimension, got %d", stubRepo.calls)
	}
}

func TestRecommendationServiceRecommendChallengesWithContextHonorsCancellation(t *testing.T) {
	db := setupRecommendationTestDB(t)
	now := time.Now()
	if err := db.Create(&model.SkillProfile{
		UserID:    11,
		Dimension: model.DimensionWeb,
		Score:     0.2,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed profile: %v", err)
	}

	service := newRecommendationTestService(db, &stubChallengeRecommendationRepo{}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.RecommendChallengesWithContext(ctx, 11, 0)
	if err == nil || err != context.Canceled {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestRecommendationServiceRegistersContestAttackAcceptedConsumer(t *testing.T) {
	db := setupRecommendationTestDB(t)

	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	cacheKey := rediskeys.RecommendationKey(17)
	if err := redisClient.Set(context.Background(), cacheKey, `[{"id":"cached"}]`, time.Hour).Err(); err != nil {
		t.Fatalf("seed recommendation cache: %v", err)
	}

	service := newRecommendationTestService(db, &stubChallengeRecommendationRepo{}, redisClient)
	bus := platformevents.NewBus()
	service.RegisterContestEventConsumers(bus)

	if err := bus.Publish(context.Background(), platformevents.Event{
		Name: contestcontracts.EventAWDAttackAccepted,
		Payload: contestcontracts.AWDAttackAcceptedEvent{
			UserID:      17,
			ContestID:   99,
			ChallengeID: 501,
			Dimension:   model.DimensionWeb,
			OccurredAt:  time.Now(),
		},
	}); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if redisClient.Exists(context.Background(), cacheKey).Val() != 0 {
		t.Fatalf("expected recommendation cache to be cleared after awd event")
	}
}

func ptrRecommendationInt64(value int64) *int64 {
	return &value
}
