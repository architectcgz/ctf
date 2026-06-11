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
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	assessmentinfra "ctf-platform/internal/module/assessment/infrastructure"
	assessmentcachekeys "ctf-platform/internal/module/assessment/infrastructure/cachekeys"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	platformevents "ctf-platform/internal/platform/events"
	"ctf-platform/internal/shared/taxonomy"
)

type assessmentRecommendationChallengeRow struct {
	ID                      int64          `gorm:"column:id;primaryKey"`
	Title                   string         `gorm:"column:title"`
	Category                string         `gorm:"column:category"`
	Difficulty              string         `gorm:"column:difficulty"`
	Points                  int            `gorm:"column:points"`
	Status                  string         `gorm:"column:status"`
	RecommendationDimension string         `gorm:"column:recommendation_dimension"`
	CreatedAt               time.Time      `gorm:"column:created_at"`
	UpdatedAt               time.Time      `gorm:"column:updated_at"`
	DeletedAt               gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (assessmentRecommendationChallengeRow) TableName() string {
	return "challenges"
}

type assessmentRecommendationTagRow struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	Name        string    `gorm:"column:name"`
	Type        string    `gorm:"column:type"`
	Description string    `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (assessmentRecommendationTagRow) TableName() string {
	return "tags"
}

type assessmentRecommendationChallengeTagRow struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	ChallengeID int64     `gorm:"column:challenge_id"`
	TagID       int64     `gorm:"column:tag_id"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (assessmentRecommendationChallengeTagRow) TableName() string {
	return "challenge_tags"
}

type stubChallengeRecommendationRepo struct {
	challenges     []*challengecontracts.RecommendationChallenge
	calls          int
	lastLimit      int
	lastDims       []string
	lastDifficulty string
	lastSolved     []int64
}

func (s *stubChallengeRecommendationRepo) FindPublishedForRecommendation(_ context.Context, limit int, dimensions []string, preferredDifficulty string, excludeSolved []int64) ([]*challengecontracts.RecommendationChallenge, error) {
	s.calls++
	s.lastLimit = limit
	s.lastDims = append([]string(nil), dimensions...)
	s.lastDifficulty = preferredDifficulty
	s.lastSolved = append([]int64(nil), excludeSolved...)
	return s.challenges, nil
}

func setupRecommendationTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&identitycontracts.User{},
		&assessmentRecommendationChallengeRow{},
		&challengecontracts.AWDChallenge{},
		&assessmententity.SkillProfile{},
		&contestcontracts.Submission{},
		&contestcontracts.AWDAttackLog{},
		&assessmentRecommendationTagRow{},
		&assessmentRecommendationChallengeTagRow{},
	); err != nil {
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
	now := time.Now().UTC()

	if err := db.Create(&identitycontracts.User{
		ID:       1,
		Username: "student-1",
		Role:     identitycontracts.RoleStudent,
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Create(&assessmententity.SkillProfile{
		UserID:    1,
		Dimension: taxonomy.DimensionWeb,
		Score:     0.2,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed profile: %v", err)
	}
	if err := db.Create(&assessmentRecommendationChallengeRow{
		ID:         11,
		Title:      "web-training-gap",
		Category:   taxonomy.DimensionWeb,
		Difficulty: taxonomy.DifficultyEasy,
		Points:     100,
		Status:     challengecontracts.ChallengeStatusPublished,
	}).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}
	for index := 0; index < 3; index++ {
		if err := db.Create(&contestcontracts.Submission{
			UserID:      1,
			ChallengeID: 11,
			IsCorrect:   false,
			SubmittedAt: now.Add(time.Duration(index) * time.Minute),
		}).Error; err != nil {
			t.Fatalf("seed submission %d: %v", index, err)
		}
	}

	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	cached := []*assessmentcontracts.ChallengeRecommendation{
		{ID: 1, Title: "cached-web", Category: taxonomy.DimensionWeb, Difficulty: taxonomy.DifficultyEasy, Points: 100, Summary: "cached"},
	}
	payload, err := json.Marshal(cached)
	if err != nil {
		t.Fatalf("marshal cached recommendations: %v", err)
	}
	if err := redisClient.Set(context.Background(), assessmentcachekeys.RecommendationKey(1), payload, time.Hour).Err(); err != nil {
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

	if err := db.Create(&identitycontracts.User{
		ID:       7,
		Username: "student-7",
		Role:     identitycontracts.RoleStudent,
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	challenges := []assessmentRecommendationChallengeRow{
		{ID: 101, Title: "web-intro", Category: taxonomy.DimensionWeb, Difficulty: taxonomy.DifficultyEasy, Points: 100, Status: challengecontracts.ChallengeStatusPublished},
		{ID: 202, Title: "pwn-intro", Category: taxonomy.DimensionPwn, Difficulty: taxonomy.DifficultyEasy, Points: 150, Status: challengecontracts.ChallengeStatusPublished},
	}
	for _, challenge := range challenges {
		if err := db.Create(&challenge).Error; err != nil {
			t.Fatalf("seed challenge: %v", err)
		}
	}

	profiles := []assessmententity.SkillProfile{
		{UserID: 7, Dimension: taxonomy.DimensionWeb, Score: 0.2, UpdatedAt: now},
		{UserID: 7, Dimension: taxonomy.DimensionCrypto, Score: 0.8, UpdatedAt: now},
		{UserID: 7, Dimension: taxonomy.DimensionPwn, Score: 0.1, UpdatedAt: now},
	}
	for _, profile := range profiles {
		if err := db.Create(&profile).Error; err != nil {
			t.Fatalf("seed profile: %v", err)
		}
	}

	submissions := []contestcontracts.Submission{
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
		challenges: []*challengecontracts.RecommendationChallenge{
			{ID: 301, Title: "web-fix", Category: taxonomy.DimensionWeb, Difficulty: taxonomy.DifficultyEasy, Points: 120},
			{ID: 302, Title: "pwn-fix", Category: taxonomy.DimensionPwn, Difficulty: taxonomy.DifficultyMedium, Points: 180},
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
	if len(stubRepo.lastDims) != 1 || stubRepo.lastDims[0] != taxonomy.DimensionPwn {
		t.Fatalf("unexpected weak dimensions: %+v", stubRepo.lastDims)
	}
	if stubRepo.lastDifficulty != taxonomy.DifficultyEasy {
		t.Fatalf("expected preferred difficulty easy for weakest evidence-backed weak dimension, got %s", stubRepo.lastDifficulty)
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

	if err := db.Create(&identitycontracts.User{
		ID:       8,
		Username: "student-8",
		Role:     identitycontracts.RoleStudent,
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Create(&assessmententity.SkillProfile{
		UserID:    8,
		Dimension: taxonomy.DimensionPwn,
		Score:     0.18,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed profile: %v", err)
	}
	if err := db.Create(&assessmentRecommendationChallengeRow{
		ID:         801,
		Title:      "pwn-primer",
		Category:   taxonomy.DimensionPwn,
		Difficulty: taxonomy.DifficultyBeginner,
		Points:     100,
		Status:     challengecontracts.ChallengeStatusPublished,
	}).Error; err != nil {
		t.Fatalf("seed practice challenge: %v", err)
	}
	for index := 0; index < 3; index++ {
		if err := db.Create(&contestcontracts.Submission{
			UserID:      8,
			ChallengeID: 801,
			IsCorrect:   false,
			SubmittedAt: now.Add(time.Duration(index) * time.Minute),
		}).Error; err != nil {
			t.Fatalf("seed submission %d: %v", index, err)
		}
	}

	stubRepo := &stubChallengeRecommendationRepo{
		challenges: []*challengecontracts.RecommendationChallenge{
			{
				ID:                      401,
				Title:                   "tagged-web-for-pwn",
				Category:                taxonomy.DimensionWeb,
				RecommendationDimension: taxonomy.DimensionPwn,
				Difficulty:              taxonomy.DifficultyEasy,
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
	if items[0].Dimension != taxonomy.DimensionPwn {
		t.Fatalf("expected recommendation dimension pwn, got %+v", items[0])
	}
	if items[0].Category != taxonomy.DimensionWeb {
		t.Fatalf("expected original challenge category preserved, got %+v", items[0])
	}
	if items[0].Summary == "" || !strings.Contains(items[0].Summary, "Pwn") {
		t.Fatalf("expected summary to follow matched recommendation dimension, got %+v", items[0])
	}
}

func TestRecommendationServiceRecommendChallengesPrefersPreferredDifficultyCandidates(t *testing.T) {
	db := setupRecommendationTestDB(t)
	now := time.Now()

	if err := db.Create(&identitycontracts.User{
		ID:       18,
		Username: "student-18",
		Role:     identitycontracts.RoleStudent,
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Create(&assessmententity.SkillProfile{
		UserID:    18,
		Dimension: taxonomy.DimensionPwn,
		Score:     0.35,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed profile: %v", err)
	}
	if err := db.Create(&assessmentRecommendationChallengeRow{
		ID:         1801,
		Title:      "training-pwn-sample",
		Category:   taxonomy.DimensionPwn,
		Difficulty: taxonomy.DifficultyEasy,
		Points:     80,
		Status:     challengecontracts.ChallengeStatusPublished,
	}).Error; err != nil {
		t.Fatalf("seed practice sample challenge: %v", err)
	}
	for index := 0; index < 2; index++ {
		if err := db.Create(&contestcontracts.Submission{
			UserID:      18,
			ChallengeID: 1801,
			IsCorrect:   false,
			SubmittedAt: now.Add(time.Duration(index) * time.Minute),
		}).Error; err != nil {
			t.Fatalf("seed submission %d: %v", index, err)
		}
	}

	candidates := []assessmentRecommendationChallengeRow{
		{
			ID:         1802,
			Title:      "pwn-beginner",
			Category:   taxonomy.DimensionPwn,
			Difficulty: taxonomy.DifficultyBeginner,
			Points:     90,
			Status:     challengecontracts.ChallengeStatusPublished,
		},
		{
			ID:         1803,
			Title:      "pwn-easy",
			Category:   taxonomy.DimensionPwn,
			Difficulty: taxonomy.DifficultyEasy,
			Points:     120,
			Status:     challengecontracts.ChallengeStatusPublished,
		},
	}
	for _, challenge := range candidates {
		if err := db.Create(&challenge).Error; err != nil {
			t.Fatalf("seed candidate %s: %v", challenge.Title, err)
		}
	}

	service := newRecommendationTestService(db, challengeinfra.NewContractRepository(challengeinfra.NewRepository(db)), nil)

	items, err := service.RecommendChallenges(context.Background(), 18, 2)
	if err != nil {
		t.Fatalf("RecommendChallenges() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 recommendations, got %+v", items)
	}
	if items[0].DifficultyBand != taxonomy.DifficultyEasy {
		t.Fatalf("expected preferred difficulty band easy, got %+v", items[0])
	}
	if items[0].Difficulty != taxonomy.DifficultyEasy {
		t.Fatalf("expected easy candidate to rank first when preferred difficulty is easy, got %+v", items)
	}
}

func TestRecommendationServiceRecommendReturnsEmptyWhenNoWeakDimension(t *testing.T) {
	db := setupRecommendationTestDB(t)
	now := time.Now()
	if err := db.Create(&assessmententity.SkillProfile{
		UserID:    9,
		Dimension: taxonomy.DimensionWeb,
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

	if err := db.Create(&identitycontracts.User{
		ID:       10,
		Username: "student-10",
		Role:     identitycontracts.RoleStudent,
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Create(&assessmententity.SkillProfile{
		UserID:    10,
		Dimension: taxonomy.DimensionWeb,
		Score:     0.82,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed profile: %v", err)
	}
	if err := db.Create(&assessmentRecommendationChallengeRow{
		ID:         901,
		Title:      "healthy-web-sample",
		Category:   taxonomy.DimensionWeb,
		Difficulty: taxonomy.DifficultyEasy,
		Points:     100,
		Status:     challengecontracts.ChallengeStatusPublished,
	}).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}
	for index := 0; index < 3; index++ {
		if err := db.Create(&contestcontracts.Submission{
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

func TestRecommendationServiceRecommendReturnsWeakDimensionsFromEvidenceBackedLiveSnapshot(t *testing.T) {
	db := setupRecommendationTestDB(t)
	now := time.Now().UTC()

	if err := db.Create(&identitycontracts.User{
		ID:       30,
		Username: "student-30",
		Role:     identitycontracts.RoleStudent,
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	profiles := []assessmententity.SkillProfile{
		{UserID: 30, Dimension: taxonomy.DimensionCrypto, Score: 0.12, UpdatedAt: now},
		{UserID: 30, Dimension: taxonomy.DimensionPwn, Score: 0.28, UpdatedAt: now},
		{UserID: 30, Dimension: taxonomy.DimensionWeb, Score: 0.65, UpdatedAt: now},
	}
	for _, profile := range profiles {
		if err := db.Create(&profile).Error; err != nil {
			t.Fatalf("seed profile: %v", err)
		}
	}
	challenges := []assessmentRecommendationChallengeRow{
		{ID: 3001, Title: "crypto-training", Category: taxonomy.DimensionCrypto, Difficulty: taxonomy.DifficultyEasy, Points: 80, Status: challengecontracts.ChallengeStatusPublished},
		{ID: 3002, Title: "pwn-training", Category: taxonomy.DimensionPwn, Difficulty: taxonomy.DifficultyEasy, Points: 90, Status: challengecontracts.ChallengeStatusPublished},
	}
	for _, challenge := range challenges {
		if err := db.Create(&challenge).Error; err != nil {
			t.Fatalf("seed challenge %s: %v", challenge.Title, err)
		}
	}
	for index := 0; index < 4; index++ {
		if err := db.Create(&contestcontracts.Submission{
			UserID:      30,
			ChallengeID: 3001,
			IsCorrect:   false,
			SubmittedAt: now.Add(time.Duration(index) * time.Minute),
		}).Error; err != nil {
			t.Fatalf("seed crypto submission %d: %v", index, err)
		}
	}
	for index := 0; index < 3; index++ {
		if err := db.Create(&contestcontracts.Submission{
			UserID:      30,
			ChallengeID: 3002,
			IsCorrect:   false,
			SubmittedAt: now.Add(10*time.Minute + time.Duration(index)*time.Minute),
		}).Error; err != nil {
			t.Fatalf("seed pwn submission %d: %v", index, err)
		}
	}

	stubRepo := &stubChallengeRecommendationRepo{
		challenges: []*challengecontracts.RecommendationChallenge{
			{ID: 3005, Title: "crypto-primer", Category: taxonomy.DimensionCrypto, Difficulty: taxonomy.DifficultyBeginner, Points: 160},
			{ID: 3006, Title: "pwn-retry", Category: taxonomy.DimensionPwn, Difficulty: taxonomy.DifficultyEasy, Points: 220},
		},
	}
	service := newRecommendationTestService(db, stubRepo, nil)

	resp, err := service.Recommend(context.Background(), 30, 3)
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}
	if len(resp.WeakDimensions) != 2 {
		t.Fatalf("expected 2 weak dimensions, got %+v", resp.WeakDimensions)
	}
	if resp.WeakDimensions[0].Dimension != taxonomy.DimensionCrypto || resp.WeakDimensions[1].Dimension != taxonomy.DimensionPwn {
		t.Fatalf("expected weak dimensions sorted by profile score, got %+v", resp.WeakDimensions)
	}
	if len(resp.Challenges) == 0 {
		t.Fatalf("expected challenge recommendations for weakest dimension, got %+v", resp)
	}
	if stubRepo.calls != 1 {
		t.Fatalf("expected challenge repo called once, got %d", stubRepo.calls)
	}
	if resp.Challenges[0].Dimension == "" || resp.Challenges[0].Summary == "" {
		t.Fatalf("expected structured recommendation contract, got %+v", resp.Challenges[0])
	}
}

func TestRecommendationServiceRecommendTurnsAWDSuccessIntoProgressionTarget(t *testing.T) {
	db := setupRecommendationTestDB(t)
	now := time.Now().UTC()

	if err := db.Create(&identitycontracts.User{
		ID:       31,
		Username: "student-31",
		Role:     identitycontracts.RoleStudent,
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Create(&assessmententity.SkillProfile{
		UserID:    31,
		Dimension: taxonomy.DimensionPwn,
		Score:     0.12,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed low practice profile: %v", err)
	}

	awdChallenges := []challengecontracts.AWDChallenge{
		{ID: 3101, Name: "pwn-awd-easy-a", Category: taxonomy.DimensionPwn, Difficulty: taxonomy.DifficultyEasy, Status: challengecontracts.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 3102, Name: "pwn-awd-easy-b", Category: taxonomy.DimensionPwn, Difficulty: taxonomy.DifficultyEasy, Status: challengecontracts.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 3103, Name: "pwn-awd-medium-a", Category: taxonomy.DimensionPwn, Difficulty: taxonomy.DifficultyMedium, Status: challengecontracts.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 3104, Name: "pwn-awd-medium-b", Category: taxonomy.DimensionPwn, Difficulty: taxonomy.DifficultyMedium, Status: challengecontracts.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
	}
	for _, challenge := range awdChallenges {
		if err := db.Create(&challenge).Error; err != nil {
			t.Fatalf("seed awd challenge %s: %v", challenge.Name, err)
		}
	}

	attackLogs := []contestcontracts.AWDAttackLog{
		{ID: 1, RoundID: 4101, AttackerTeamID: 5101, VictimTeamID: 6101, ServiceID: 7101, AWDChallengeID: 3101, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 80, SubmittedByUserID: ptrRecommendationInt64(31), CreatedAt: now},
		{ID: 2, RoundID: 4101, AttackerTeamID: 5101, VictimTeamID: 6102, ServiceID: 7102, AWDChallengeID: 3102, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 80, SubmittedByUserID: ptrRecommendationInt64(31), CreatedAt: now.Add(1 * time.Minute)},
		{ID: 3, RoundID: 4102, AttackerTeamID: 5101, VictimTeamID: 6103, ServiceID: 7103, AWDChallengeID: 3103, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 90, SubmittedByUserID: ptrRecommendationInt64(31), CreatedAt: now.Add(2 * time.Minute)},
		{ID: 4, RoundID: 4102, AttackerTeamID: 5101, VictimTeamID: 6104, ServiceID: 7104, AWDChallengeID: 3104, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 90, SubmittedByUserID: ptrRecommendationInt64(31), CreatedAt: now.Add(3 * time.Minute)},
	}
	for _, log := range attackLogs {
		if err := db.Create(&log).Error; err != nil {
			t.Fatalf("seed awd attack log %d: %v", log.ID, err)
		}
	}

	candidateChallenges := []assessmentRecommendationChallengeRow{
		{ID: 3111, Title: "pwn-hard-next", Category: taxonomy.DimensionPwn, Difficulty: taxonomy.DifficultyHard, Points: 240, Status: challengecontracts.ChallengeStatusPublished},
		{ID: 3112, Title: "pwn-easy-next", Category: taxonomy.DimensionPwn, Difficulty: taxonomy.DifficultyEasy, Points: 180, Status: challengecontracts.ChallengeStatusPublished},
	}
	for _, challenge := range candidateChallenges {
		if err := db.Create(&challenge).Error; err != nil {
			t.Fatalf("seed candidate challenge %s: %v", challenge.Title, err)
		}
	}

	service := newRecommendationTestService(db, challengeinfra.NewContractRepository(challengeinfra.NewRepository(db)), nil)

	resp, err := service.Recommend(context.Background(), 31, 2)
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}
	if len(resp.WeakDimensions) != 0 {
		t.Fatalf("expected awd-backed stable direction to clear explicit weak dimensions, got %+v", resp.WeakDimensions)
	}
	if len(resp.Challenges) == 0 {
		t.Fatalf("expected progression recommendation after awd-backed stable direction, got %+v", resp)
	}
	if resp.Challenges[0].DifficultyBand != taxonomy.DifficultyHard {
		t.Fatalf("expected hard progression difficulty band, got %+v", resp.Challenges[0])
	}
	if resp.Challenges[0].Difficulty != taxonomy.DifficultyHard {
		t.Fatalf("expected hard candidate to rank first for progression recommendation, got %+v", resp.Challenges)
	}
	if !strings.Contains(resp.Challenges[0].Summary, "进阶") {
		t.Fatalf("expected progression summary, got %+v", resp.Challenges[0])
	}
}

func TestRecommendationServiceRecommendExcludesContestSolvedChallengeFromFallbackSample(t *testing.T) {
	db := setupRecommendationTestDB(t)
	now := time.Now().UTC()
	contestID := int64(401)

	if err := db.Create(&identitycontracts.User{
		ID:       32,
		Username: "student-32",
		Role:     identitycontracts.RoleStudent,
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Create(&assessmententity.SkillProfile{
		UserID:    32,
		Dimension: taxonomy.DimensionWeb,
		Score:     0.15,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed web profile: %v", err)
	}

	challenges := []assessmentRecommendationChallengeRow{
		{ID: 3201, Title: "contest-web-solved", Category: taxonomy.DimensionWeb, Difficulty: taxonomy.DifficultyEasy, Points: 80, Status: challengecontracts.ChallengeStatusPublished},
		{ID: 3202, Title: "contest-web-next", Category: taxonomy.DimensionWeb, Difficulty: taxonomy.DifficultyEasy, Points: 100, Status: challengecontracts.ChallengeStatusPublished},
	}
	for _, challenge := range challenges {
		if err := db.Create(&challenge).Error; err != nil {
			t.Fatalf("seed challenge %s: %v", challenge.Title, err)
		}
	}

	submissions := []contestcontracts.Submission{
		{
			UserID:       32,
			ChallengeID:  3201,
			ContestID:    &contestID,
			IsCorrect:    true,
			ReviewStatus: contestcontracts.SubmissionReviewStatusApproved,
			SubmittedAt:  now,
			UpdatedAt:    now,
		},
		{
			UserID:      32,
			ChallengeID: 3201,
			ContestID:   &contestID,
			IsCorrect:   true,
			SubmittedAt: now.Add(1 * time.Minute),
			UpdatedAt:   now.Add(1 * time.Minute),
		},
	}
	for index, submission := range submissions {
		if err := db.Create(&submission).Error; err != nil {
			t.Fatalf("seed contest submission %d: %v", index, err)
		}
	}

	service := newRecommendationTestService(db, challengeinfra.NewContractRepository(challengeinfra.NewRepository(db)), nil)

	resp, err := service.Recommend(context.Background(), 32, 2)
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}
	if len(resp.WeakDimensions) != 0 {
		t.Fatalf("expected contest-backed stable sample to avoid explicit weak dimension, got %+v", resp.WeakDimensions)
	}
	if len(resp.Challenges) == 0 {
		t.Fatalf("expected fallback sample recommendation, got %+v", resp)
	}
	if resp.Challenges[0].ID != 3202 {
		t.Fatalf("expected solved contest challenge to be excluded from recommendation, got %+v", resp.Challenges)
	}
	if resp.Challenges[0].Dimension != taxonomy.DimensionWeb {
		t.Fatalf("expected recommendation to stay on web dimension, got %+v", resp.Challenges[0])
	}
}

func TestRecommendationServiceRecommendChallengesHonorsCancellation(t *testing.T) {
	db := setupRecommendationTestDB(t)
	now := time.Now()
	if err := db.Create(&assessmententity.SkillProfile{
		UserID:    11,
		Dimension: taxonomy.DimensionWeb,
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

	cacheKey := assessmentcachekeys.RecommendationKey(17)
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
			Dimension:      taxonomy.DimensionWeb,
			OccurredAt:     time.Now(),
		},
	}); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if redisClient.Exists(context.Background(), cacheKey).Val() != 0 {
		t.Fatalf("expected recommendation cache to be cleared after awd event")
	}
}

func TestRecommendationServiceRegistersContestFlagAcceptedConsumer(t *testing.T) {
	db := setupRecommendationTestDB(t)

	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	cacheKey := assessmentcachekeys.RecommendationKey(18)
	if err := redisClient.Set(context.Background(), cacheKey, `[{"id":"cached"}]`, time.Hour).Err(); err != nil {
		t.Fatalf("seed recommendation cache: %v", err)
	}

	service := newRecommendationTestService(db, &stubChallengeRecommendationRepo{}, redisClient)
	bus := platformevents.NewBus()
	service.RegisterContestEventConsumers(bus)

	if err := bus.Publish(context.Background(), platformevents.Event{
		Name: contestcontracts.EventFlagAccepted,
		Payload: contestcontracts.FlagAcceptedEvent{
			UserID:      18,
			ContestID:   101,
			ChallengeID: 601,
			Dimension:   taxonomy.DimensionWeb,
			OccurredAt:  time.Now(),
		},
	}); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if redisClient.Exists(context.Background(), cacheKey).Val() != 0 {
		t.Fatalf("expected recommendation cache to be cleared after contest flag event")
	}
}

func ptrRecommendationInt64(value int64) *int64 {
	return &value
}
