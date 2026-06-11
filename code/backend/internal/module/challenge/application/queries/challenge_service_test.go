package queries

import (
	"context"
	"errors"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"ctf-platform/internal/apperror"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/internal/module/challenge/testsupport"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
)

type challengeQueryRepositoryStub struct {
	findByIDFn             func(context.Context, int64) (*challengeports.ChallengeReadModel, error)
	listFn                 func(context.Context, *challengecontracts.ChallengeQuery) ([]*challengeports.ChallengeReadModel, int64, error)
	listPublishedFn        func(context.Context, *challengecontracts.ChallengeQuery) ([]*challengeports.ChallengeReadModel, int64, error)
	listHintsByChallengeID func(context.Context, int64) ([]*challengeentity.ChallengeHint, error)
	getSolvedStatusFn      func(context.Context, int64, int64) (bool, error)
	getSolvedCountFn       func(context.Context, int64) (int64, error)
	getTotalAttemptsFn     func(context.Context, int64) (int64, error)
	batchSolvedStatusFn    func(context.Context, int64, []int64) (map[int64]bool, error)
	batchSolvedCountFn     func(context.Context, []int64) (map[int64]int64, error)
	batchTotalAttemptsFn   func(context.Context, []int64) (map[int64]int64, error)
}

func (s *challengeQueryRepositoryStub) FindByID(ctx context.Context, id int64) (*challengeports.ChallengeReadModel, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (s *challengeQueryRepositoryStub) List(ctx context.Context, query *challengecontracts.ChallengeQuery) ([]*challengeports.ChallengeReadModel, int64, error) {
	if s.listFn != nil {
		return s.listFn(ctx, query)
	}
	return nil, 0, nil
}

func (s *challengeQueryRepositoryStub) ListPublished(ctx context.Context, query *challengecontracts.ChallengeQuery) ([]*challengeports.ChallengeReadModel, int64, error) {
	if s.listPublishedFn != nil {
		return s.listPublishedFn(ctx, query)
	}
	return nil, 0, nil
}

func (s *challengeQueryRepositoryStub) ListHintsByChallengeID(ctx context.Context, challengeID int64) ([]*challengeentity.ChallengeHint, error) {
	if s.listHintsByChallengeID != nil {
		return s.listHintsByChallengeID(ctx, challengeID)
	}
	return nil, nil
}

func (s *challengeQueryRepositoryStub) GetSolvedStatus(ctx context.Context, userID, challengeID int64) (bool, error) {
	if s.getSolvedStatusFn != nil {
		return s.getSolvedStatusFn(ctx, userID, challengeID)
	}
	return false, nil
}

func (s *challengeQueryRepositoryStub) GetSolvedCount(ctx context.Context, challengeID int64) (int64, error) {
	if s.getSolvedCountFn != nil {
		return s.getSolvedCountFn(ctx, challengeID)
	}
	return 0, nil
}

func (s *challengeQueryRepositoryStub) GetTotalAttempts(ctx context.Context, challengeID int64) (int64, error) {
	if s.getTotalAttemptsFn != nil {
		return s.getTotalAttemptsFn(ctx, challengeID)
	}
	return 0, nil
}

func (s *challengeQueryRepositoryStub) BatchGetSolvedStatus(ctx context.Context, userID int64, challengeIDs []int64) (map[int64]bool, error) {
	if s.batchSolvedStatusFn != nil {
		return s.batchSolvedStatusFn(ctx, userID, challengeIDs)
	}
	return map[int64]bool{}, nil
}

func (s *challengeQueryRepositoryStub) BatchGetSolvedCount(ctx context.Context, challengeIDs []int64) (map[int64]int64, error) {
	if s.batchSolvedCountFn != nil {
		return s.batchSolvedCountFn(ctx, challengeIDs)
	}
	return map[int64]int64{}, nil
}

func (s *challengeQueryRepositoryStub) BatchGetTotalAttempts(ctx context.Context, challengeIDs []int64) (map[int64]int64, error) {
	if s.batchTotalAttemptsFn != nil {
		return s.batchTotalAttemptsFn(ctx, challengeIDs)
	}
	return map[int64]int64{}, nil
}

func challengeReadModelFromModel(source *challengeentity.Challenge) *challengeports.ChallengeReadModel {
	if source == nil {
		return nil
	}
	return &challengeports.ChallengeReadModel{
		ID:              source.ID,
		PackageSlug:     source.PackageSlug,
		Title:           source.Title,
		Description:     source.Description,
		Category:        source.Category,
		Difficulty:      source.Difficulty,
		Points:          source.Points,
		ImageID:         source.ImageID,
		AttachmentURL:   source.AttachmentURL,
		Status:          string(source.Status),
		FlagType:        source.FlagType,
		FlagPrefix:      source.FlagPrefix,
		InstanceSharing: string(source.InstanceSharing),
		CreatedBy:       source.CreatedBy,
		CreatedAt:       source.CreatedAt,
		UpdatedAt:       source.UpdatedAt,
	}
}

func TestServiceGetPublishedChallengeDraftChallengeReturnsDraftAccessError(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	challenge := &challengeentity.Challenge{Title: "Test", Status: challengeentity.ChallengeStatusDraft}
	db.Create(challenge)

	repo := challengeinfra.NewChallengeQueryRepository(challengeinfra.NewRepository(db))
	service := NewChallengeService(repo, nil, &Config{SolvedCountCacheTTL: time.Minute}, nil)

	_, err := service.GetPublishedChallenge(context.Background(), 1, challenge.ID)
	if err == nil {
		t.Fatal("expected draft access error")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected app error, got %v", err)
	}
	if appErr.Code != challengecontracts.ErrChallengeNotPublish.Code {
		t.Fatalf("expected challenge not publish code, got %+v", appErr)
	}
	if appErr.Message != "题目为草稿，无法访问" {
		t.Fatalf("expected draft access message, got %+v", appErr)
	}
}

func TestServiceGetPublishedChallengeArchivedChallengeReturnsArchivedAccessError(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	challenge := &challengeentity.Challenge{Title: "Test", Status: challengeentity.ChallengeStatusArchived}
	db.Create(challenge)

	repo := challengeinfra.NewChallengeQueryRepository(challengeinfra.NewRepository(db))
	service := NewChallengeService(repo, nil, &Config{SolvedCountCacheTTL: time.Minute}, nil)

	_, err := service.GetPublishedChallenge(context.Background(), 1, challenge.ID)
	if err == nil {
		t.Fatal("expected archived access error")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected app error, got %v", err)
	}
	if appErr.Code != challengecontracts.ErrChallengeNotPublish.Code {
		t.Fatalf("expected challenge not publish code, got %+v", appErr)
	}
	if appErr.Message != "题目已归档，无法访问" {
		t.Fatalf("expected archived access message, got %+v", appErr)
	}
}

func TestChallengeServiceGetChallengeTreatsChallengeQueryNotFoundAsChallengeNotFound(t *testing.T) {
	t.Parallel()

	service := NewChallengeService(&challengeQueryRepositoryStub{
		findByIDFn: func(context.Context, int64) (*challengeports.ChallengeReadModel, error) {
			return nil, challengeports.ErrChallengeQueryChallengeNotFound
		},
	}, nil, &Config{SolvedCountCacheTTL: time.Minute}, nil)

	_, err := service.GetChallenge(context.Background(), 404)
	if err == nil {
		t.Fatal("expected challenge not found")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != challengecontracts.ErrChallengeNotFound.Code {
		t.Fatalf("expected challengecontracts.ErrChallengeNotFound, got %v", err)
	}
}

func TestChallengeServiceGetPublishedChallengeTreatsChallengeQueryNotFoundAsNotFound(t *testing.T) {
	t.Parallel()

	service := NewChallengeService(&challengeQueryRepositoryStub{
		findByIDFn: func(context.Context, int64) (*challengeports.ChallengeReadModel, error) {
			return nil, challengeports.ErrChallengeQueryChallengeNotFound
		},
	}, nil, &Config{SolvedCountCacheTTL: time.Minute}, nil)

	_, err := service.GetPublishedChallenge(context.Background(), 7, 404)
	if err == nil {
		t.Fatal("expected published challenge not found")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.ErrNotFound.Code {
		t.Fatalf("expected apperror.ErrNotFound, got %v", err)
	}
}

func TestServiceGetChallengeIncludesHintsAndAttachment(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	challenge := &challengeentity.Challenge{
		Title:         "Hint Challenge",
		Description:   "desc",
		Category:      "web",
		Difficulty:    challengeentity.ChallengeDifficultyEasy,
		Points:        100,
		ImageID:       int64Ptr(1),
		AttachmentURL: "https://example.com/files/hint.zip",
		Status:        challengeentity.ChallengeStatusDraft,
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := db.Create(&challengeentity.ChallengeHint{
		ChallengeID: challenge.ID,
		Level:       1,
		Title:       "第一条提示",
		Content:     "从登录入口开始",
	}).Error; err != nil {
		t.Fatalf("create hint: %v", err)
	}

	repo := challengeinfra.NewChallengeQueryRepository(challengeinfra.NewRepository(db))
	service := NewChallengeService(repo, nil, &Config{SolvedCountCacheTTL: time.Minute}, nil)

	resp, err := service.GetChallenge(context.Background(), challenge.ID)
	if err != nil {
		t.Fatalf("GetChallenge() error = %v", err)
	}
	if resp.AttachmentURL != challenge.AttachmentURL {
		t.Fatalf("unexpected attachment url: %s", resp.AttachmentURL)
	}
	if len(resp.Hints) != 1 || resp.Hints[0].Content != "从登录入口开始" {
		t.Fatalf("unexpected hints: %+v", resp.Hints)
	}
}

func TestServiceGetSolvedCountCachedHonorsContextCancellation(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	challenge := &challengeentity.Challenge{
		Title:  "Published",
		Status: challengeentity.ChallengeStatusPublished,
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := db.Create(&contestcontracts.Submission{
		UserID:      1,
		ChallengeID: challenge.ID,
		IsCorrect:   true,
	}).Error; err != nil {
		t.Fatalf("create submission: %v", err)
	}

	mini := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := NewChallengeService(challengeinfra.NewChallengeQueryRepository(challengeinfra.NewRepository(db)), challengeinfra.NewSolvedCountCache(redisClient), &Config{SolvedCountCacheTTL: time.Minute}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.getSolvedCountCached(ctx, challenge.ID)
	if err == nil || err != context.Canceled {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestServiceGetPublishedChallengeUsesSolvedCountCache(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	challenge := &challengeentity.Challenge{
		Title:  "Published Cached",
		Status: challengeentity.ChallengeStatusPublished,
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	mini := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	solvedCountCache := challengeinfra.NewSolvedCountCache(redisClient)
	if err := solvedCountCache.StoreSolvedCount(context.Background(), challenge.ID, 7, time.Minute); err != nil {
		t.Fatalf("StoreSolvedCount() error = %v", err)
	}

	service := NewChallengeService(challengeinfra.NewChallengeQueryRepository(challengeinfra.NewRepository(db)), solvedCountCache, &Config{SolvedCountCacheTTL: time.Minute}, nil)

	resp, err := service.GetPublishedChallenge(context.Background(), 0, challenge.ID)
	if err != nil {
		t.Fatalf("GetPublishedChallenge() error = %v", err)
	}
	if resp.SolvedCount != 7 {
		t.Fatalf("unexpected solved count: %d", resp.SolvedCount)
	}
}

func TestServiceGetSolvedCountCachedWarmsCacheOnMiss(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	challenge := &challengeentity.Challenge{
		Title:  "Published Warm Cache",
		Status: challengeentity.ChallengeStatusPublished,
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := db.Create(&contestcontracts.Submission{
		UserID:      1,
		ChallengeID: challenge.ID,
		IsCorrect:   true,
	}).Error; err != nil {
		t.Fatalf("create submission: %v", err)
	}

	mini := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	solvedCountCache := challengeinfra.NewSolvedCountCache(redisClient)
	service := NewChallengeService(challengeinfra.NewChallengeQueryRepository(challengeinfra.NewRepository(db)), solvedCountCache, &Config{SolvedCountCacheTTL: time.Minute}, nil)

	count, err := service.getSolvedCountCached(context.Background(), challenge.ID)
	if err != nil {
		t.Fatalf("getSolvedCountCached() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("unexpected solved count: %d", count)
	}

	cachedCount, hit, err := solvedCountCache.GetSolvedCount(context.Background(), challenge.ID)
	if err != nil {
		t.Fatalf("GetSolvedCount() error = %v", err)
	}
	if !hit || cachedCount != 1 {
		t.Fatalf("unexpected cached solved count: hit=%v count=%d", hit, cachedCount)
	}
}

func TestServiceGetChallengeHonorsCancellation(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	challenge := &challengeentity.Challenge{Title: "ctx get", Status: challengeentity.ChallengeStatusDraft}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	service := NewChallengeService(challengeinfra.NewChallengeQueryRepository(challengeinfra.NewRepository(db)), nil, &Config{SolvedCountCacheTTL: time.Minute}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.GetChallenge(ctx, challenge.ID)
	if err == nil || err != context.Canceled {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestServiceListChallengesHonorsCancellation(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	challenge := &challengeentity.Challenge{Title: "ctx list", Status: challengeentity.ChallengeStatusDraft}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	service := NewChallengeService(challengeinfra.NewChallengeQueryRepository(challengeinfra.NewRepository(db)), nil, &Config{SolvedCountCacheTTL: time.Minute}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.ListChallenges(ctx, &challengecontracts.ChallengeQuery{})
	if err == nil || err != context.Canceled {
		t.Fatalf("expected context canceled, got %v", err)
	}
}
