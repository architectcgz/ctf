package queries

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/dto"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	"ctf-platform/internal/module/challenge/testsupport"
	"ctf-platform/pkg/errcode"
)

func TestServiceGetPublishedChallengeNotPublished(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	challenge := &model.Challenge{Title: "Test", Status: model.ChallengeStatusDraft}
	db.Create(challenge)

	repo := challengeinfra.NewRepository(db)
	service := NewChallengeService(repo, nil, &Config{SolvedCountCacheTTL: time.Minute}, nil)

	_, err := service.GetPublishedChallenge(context.Background(), 1, challenge.ID)
	if err == nil || err.Error() != errcode.ErrForbidden.Error() {
		t.Fatalf("expected not published error, got %v", err)
	}
}

func TestServiceGetChallengeIncludesHintsAndAttachment(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	challenge := &model.Challenge{
		Title:         "Hint Challenge",
		Description:   "desc",
		Category:      "web",
		Difficulty:    model.ChallengeDifficultyEasy,
		Points:        100,
		ImageID:       1,
		AttachmentURL: "https://example.com/files/hint.zip",
		Status:        model.ChallengeStatusDraft,
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := db.Create(&model.ChallengeHint{
		ChallengeID: challenge.ID,
		Level:       1,
		Title:       "第一条提示",
		Content:     "从登录入口开始",
	}).Error; err != nil {
		t.Fatalf("create hint: %v", err)
	}

	repo := challengeinfra.NewRepository(db)
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

	challenge := &model.Challenge{
		Title:  "Published",
		Status: model.ChallengeStatusPublished,
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := db.Create(&model.Submission{
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

	service := NewChallengeService(challengeinfra.NewRepository(db), challengeinfra.NewSolvedCountCache(redisClient), &Config{SolvedCountCacheTTL: time.Minute}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.getSolvedCountCached(ctx, challenge.ID)
	if err == nil || err != context.Canceled {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestServiceGetPublishedChallengeUsesSolvedCountCache(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	challenge := &model.Challenge{
		Title:  "Published Cached",
		Status: model.ChallengeStatusPublished,
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

	service := NewChallengeService(challengeinfra.NewRepository(db), solvedCountCache, &Config{SolvedCountCacheTTL: time.Minute}, nil)

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

	challenge := &model.Challenge{
		Title:  "Published Warm Cache",
		Status: model.ChallengeStatusPublished,
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := db.Create(&model.Submission{
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
	service := NewChallengeService(challengeinfra.NewRepository(db), solvedCountCache, &Config{SolvedCountCacheTTL: time.Minute}, nil)

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

	challenge := &model.Challenge{Title: "ctx get", Status: model.ChallengeStatusDraft}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	service := NewChallengeService(challengeinfra.NewRepository(db), nil, &Config{SolvedCountCacheTTL: time.Minute}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.GetChallenge(ctx, challenge.ID)
	if err == nil || err != context.Canceled {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestServiceListChallengesHonorsCancellation(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	challenge := &model.Challenge{Title: "ctx list", Status: model.ChallengeStatusDraft}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	service := NewChallengeService(challengeinfra.NewRepository(db), nil, &Config{SolvedCountCacheTTL: time.Minute}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.ListChallenges(ctx, &dto.ChallengeQuery{})
	if err == nil || err != context.Canceled {
		t.Fatalf("expected context canceled, got %v", err)
	}
}
