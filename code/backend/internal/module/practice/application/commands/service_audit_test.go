package commands

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	flagcrypto "ctf-platform/internal/shared/flagcrypto"
	"ctf-platform/internal/shared/taxonomy"
)

func TestSubmitFlagRequestsAuditSkipForRepeatCorrectSubmission(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	flagSalt := "repeat-audit-salt"

	if err := db.Create(&identitycontracts.User{
		ID:        71,
		Username:  "student71",
		Role:      identitycontracts.RoleStudent,
		Status:    identitycontracts.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&practiceCommandChallengeRow{
		ID:        11,
		Category:  taxonomy.DimensionWeb,
		Points:    100,
		Status:    challengecontracts.ChallengeStatusPublished,
		FlagType:  challengecontracts.FlagTypeStatic,
		FlagSalt:  flagSalt,
		FlagHash:  flagcrypto.HashStaticFlag("flag{repeatable}", flagSalt),
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	repo := newPracticeRepositoryWithRuntimePortOwner(db)
	challengeRepo := challengeinfra.NewRepository(db)
	service := wirePracticeSubmissionAdapters(
		newServiceCore(
			repo,

			nil,
			nil,
			nil,
			nil,
			newPracticeFlagSubmitRateLimitStoreForTest(redisClient),
			&config.Config{
				RateLimit: config.RateLimitConfig{
					RedisKeyPrefix: "practice:test",
					FlagSubmit: config.RateLimitPolicyConfig{
						Limit:  5,
						Window: time.Minute,
					},
				},
			},
			nil),

		repo,
		challengeRepo,
	)

	if _, err := service.SubmitFlag(context.Background(), 71, 11, "flag{repeatable}"); err != nil {
		t.Fatalf("SubmitFlag() first error = %v", err)
	}

	control := &auditlog.Control{}
	ctx := auditlog.WithControl(context.Background(), control)

	if _, err := service.SubmitFlag(ctx, 71, 11, "flag{repeatable}"); err != nil {
		t.Fatalf("SubmitFlag() repeat error = %v", err)
	}
	if !control.Skip {
		t.Fatal("expected repeat correct submission to request audit skip")
	}
}

func TestSubmitFlagRejectsTooFrequentAttempts(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	flagSalt := "rate-limit-salt"

	if err := db.Create(&identitycontracts.User{
		ID:        81,
		Username:  "student81",
		Role:      identitycontracts.RoleStudent,
		Status:    identitycontracts.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&practiceCommandChallengeRow{
		ID:        12,
		Category:  taxonomy.DimensionWeb,
		Points:    50,
		Status:    challengecontracts.ChallengeStatusPublished,
		FlagType:  challengecontracts.FlagTypeStatic,
		FlagSalt:  flagSalt,
		FlagHash:  flagcrypto.HashStaticFlag("flag{limited}", flagSalt),
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	repo := newPracticeRepositoryWithRuntimePortOwner(db)
	challengeRepo := challengeinfra.NewRepository(db)
	service := wirePracticeSubmissionAdapters(
		newServiceCore(
			repo,

			nil,
			nil,
			nil,
			nil,
			newPracticeFlagSubmitRateLimitStoreForTest(redisClient),
			&config.Config{
				RateLimit: config.RateLimitConfig{
					RedisKeyPrefix: "practice:test",
					FlagSubmit: config.RateLimitPolicyConfig{
						Limit:  1,
						Window: time.Minute,
					},
				},
			},
			nil),

		repo,
		challengeRepo,
	)

	if _, err := service.SubmitFlag(context.Background(), 81, 12, "flag{wrong}"); err != nil {
		t.Fatalf("SubmitFlag() first error = %v", err)
	}

	_, err := service.SubmitFlag(context.Background(), 81, 12, "flag{wrong-again}")
	if err == nil || err.Error() != challengecontracts.ErrSubmitTooFrequent.Error() {
		t.Fatalf("expected submit too frequent, got %v", err)
	}
}
