package queries_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
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
	if err := db.AutoMigrate(&model.User{}, &model.Challenge{}, &model.SkillProfile{}, &model.Submission{}, &model.AWDAttackLog{}); err != nil {
		t.Fatalf("migrate recommendation tables: %v", err)
	}
	return db
}

func newRecommendationTestService(db *gorm.DB, challengeRepo assessmentports.RecommendationChallengeRepository, redisClient *redis.Client) *assessmentqry.RecommendationService {
	var cacheStore assessmentports.AssessmentRecommendationCacheStore
	if redisClient != nil {
		cacheStore = assessmentinfra.NewRecommendationCacheStore(redisClient)
	}
	return assessmentqry.NewRecommendationService(
		assessmentinfra.NewRepository(db),
		challengeRepo,
		cacheStore,
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
		{ID: 1, Title: "cached-web", Category: model.DimensionWeb, Difficulty: model.ChallengeDifficultyEasy, Points: 100, Summary: "cached"},
	}
	payload, err := json.Marshal(cached)
	if err != nil {
		t.Fatalf("marshal cached recommendations: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.RecommendationKey(1), payload, time.Hour).Err(); err != nil {
		t.Fatalf("seed recommendation cache: %v", err)
	}

	service := newRecommendationTestService(db, stubRepo, redisClient)
	items, err := service.RecommendChallenges(context.Background(), 1, 0)
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

	if err := db.Create(&model.User{
		ID:       7,
		Username: "student-7",
		Role:     model.RoleStudent,
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	challenges := []model.Challenge{
		{ID: 101, Title: "web-intro", Category: model.DimensionWeb, Difficulty: model.ChallengeDifficultyEasy, Points: 100, Status: model.ChallengeStatusPublished},
		{ID: 202, Title: "pwn-intro", Category: model.DimensionPwn, Difficulty: model.ChallengeDifficultyEasy, Points: 150, Status: model.ChallengeStatusPublished},
	}
	for _, challenge := range challenges {
		if err := db.Create(&challenge).Error; err != nil {
			t.Fatalf("seed challenge: %v", err)
		}
	}

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
		{UserID: 7, ChallengeID: 202, IsCorrect: false, SubmittedAt: now.Add(1 * time.Minute)},
		{UserID: 7, ChallengeID: 202, IsCorrect: false, SubmittedAt: now.Add(2 * time.Minute)},
		{UserID: 7, ChallengeID: 202, IsCorrect: false, SubmittedAt: now.Add(3 * time.Minute)},
	}
	for _, submission := range submissions {
		if err := db.Create(&submission).Error; err != nil {
			t.Fatalf("seed submission: %v", err)
		}
	}

	stubRepo := &stubChallengeRecommendationRepo{
		challenges: []*model.Challenge{
			{ID: 301, Title: "web-fix", Category: model.DimensionWeb, Difficulty: model.ChallengeDifficultyEasy, Points: 120},
			{ID: 302, Title: "pwn-fix", Category: model.DimensionPwn, Difficulty: model.ChallengeDifficultyMedium, Points: 180},
		},
	}
	service := newRecommendationTestService(db, stubRepo, nil)

	items, err := service.RecommendChallenges(context.Background(), 7, 99)
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
	if len(stubRepo.lastDims) != 1 || stubRepo.lastDims[0] != model.DimensionPwn {
		t.Fatalf("unexpected weak dimensions: %+v", stubRepo.lastDims)
	}
	if len(stubRepo.lastSolved) != 1 || stubRepo.lastSolved[0] != 101 {
		t.Fatalf("unexpected solved challenge ids: %+v", stubRepo.lastSolved)
	}
	if items[0].Summary == "" || items[1].Summary == "" {
		t.Fatalf("expected recommendation summary generated, got %+v", items)
	}
	if items[0].Dimension == "" || items[1].Dimension == "" {
		t.Fatalf("expected recommendation dimensions generated, got %+v", items)
	}
	if len(items[0].ReasonCodes) == 0 || len(items[1].ReasonCodes) == 0 {
		t.Fatalf("expected recommendation reason codes generated, got %+v", items)
	}
}

func TestRecommendationServiceRecommendChallengesUsesMatchedRecommendationDimension(t *testing.T) {
	db := setupRecommendationTestDB(t)
	now := time.Now()

	if err := db.Create(&model.User{
		ID:       8,
		Username: "student-8",
		Role:     model.RoleStudent,
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Create(&model.SkillProfile{
		UserID:    8,
		Dimension: model.DimensionPwn,
		Score:     0.18,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed profile: %v", err)
	}
	if err := db.Create(&model.Challenge{
		ID:         801,
		Title:      "pwn-primer",
		Category:   model.DimensionPwn,
		Difficulty: model.ChallengeDifficultyBeginner,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
	}).Error; err != nil {
		t.Fatalf("seed practice challenge: %v", err)
	}
	for index := 0; index < 3; index++ {
		if err := db.Create(&model.Submission{
			UserID:      8,
			ChallengeID: 801,
			IsCorrect:   false,
			SubmittedAt: now.Add(time.Duration(index) * time.Minute),
		}).Error; err != nil {
			t.Fatalf("seed submission %d: %v", index, err)
		}
	}

	stubRepo := &stubChallengeRecommendationRepo{
		challenges: []*model.Challenge{
			{
				ID:                      401,
				Title:                   "tagged-web-for-pwn",
				Category:                model.DimensionWeb,
				RecommendationDimension: model.DimensionPwn,
				Difficulty:              model.ChallengeDifficultyEasy,
				Points:                  120,
			},
		},
	}
	service := newRecommendationTestService(db, stubRepo, nil)

	items, err := service.RecommendChallenges(context.Background(), 8, 3)
	if err != nil {
		t.Fatalf("RecommendChallenges() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 recommendation, got %+v", items)
	}
	if items[0].Dimension != model.DimensionPwn {
		t.Fatalf("expected recommendation dimension pwn, got %+v", items[0])
	}
	if items[0].Category != model.DimensionWeb {
		t.Fatalf("expected original challenge category preserved, got %+v", items[0])
	}
	if items[0].Summary == "" || !strings.Contains(items[0].Summary, "Pwn") {
		t.Fatalf("expected summary to follow matched recommendation dimension, got %+v", items[0])
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

	resp, err := service.Recommend(context.Background(), 9, 0)
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

func TestRecommendationServiceRecommendReturnsEmptyWhenOnlyHealthyEvidenceExists(t *testing.T) {
	db := setupRecommendationTestDB(t)
	now := time.Now()

	if err := db.Create(&model.User{
		ID:       10,
		Username: "student-10",
		Role:     model.RoleStudent,
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Create(&model.SkillProfile{
		UserID:    10,
		Dimension: model.DimensionWeb,
		Score:     0.82,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed profile: %v", err)
	}
	if err := db.Create(&model.Challenge{
		ID:         901,
		Title:      "healthy-web-sample",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
	}).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}
	for index := 0; index < 3; index++ {
		if err := db.Create(&model.Submission{
			UserID:      10,
			ChallengeID: 901,
			IsCorrect:   index < 2,
			SubmittedAt: now.Add(time.Duration(index) * time.Minute),
		}).Error; err != nil {
			t.Fatalf("seed submission %d: %v", index, err)
		}
	}

	stubRepo := &stubChallengeRecommendationRepo{}
	service := newRecommendationTestService(db, stubRepo, nil)

	resp, err := service.Recommend(context.Background(), 10, 3)
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}
	if len(resp.WeakDimensions) != 0 || len(resp.Challenges) != 0 {
		t.Fatalf("expected empty recommendation response for healthy evidence-backed student, got %+v", resp)
	}
	if stubRepo.calls != 0 {
		t.Fatalf("expected no challenge query when only healthy evidence exists, got %d", stubRepo.calls)
	}
}

func TestRecommendationServiceRecommendChallengesHonorsCancellation(t *testing.T) {
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

	_, err := service.RecommendChallenges(ctx, 11, 0)
	if !errors.Is(err, context.Canceled) {
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
			UserID:         17,
			ContestID:      99,
			AWDChallengeID: 501,
			Dimension:      model.DimensionWeb,
			OccurredAt:     time.Now(),
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
