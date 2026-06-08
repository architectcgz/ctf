package commands

import (
	"context"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
	"ctf-platform/internal/shared/flagcrypto"
	"ctf-platform/internal/shared/taxonomy"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestSubmitFlagPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("submit")
	expectedCtxValue := "ctx-submit-flag"
	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()
	flagSalt := "context-submit-salt"

	findCorrectCalled := false
	createSubmissionCalled := false
	challengeLookupCalled := false
	repo := &stubPracticeRepository{
		findCorrectSubmissionFn: func(ctx context.Context, userID, challengeID int64) (*practiceports.SubmissionRecord, error) {
			findCorrectCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-correct ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil, gorm.ErrRecordNotFound
		},
		createSubmissionFn: func(ctx context.Context, submission *practiceports.SubmissionRecord) error {
			createSubmissionCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected create-submission ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	challengeRepo := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			challengeLookupCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected challenge lookup ctx value %v, got %v", expectedCtxValue, got)
			}
			return &challengecontracts.PracticeRuntimeChallenge{
				ID:       id,
				Category: taxonomy.DimensionWeb,
				Points:   100,
				Status:   challengecontracts.ChallengeStatusPublished,
				FlagType: challengecontracts.FlagTypeStatic,
				FlagSalt: flagSalt,
				FlagHash: flagcrypto.HashStaticFlag("flag{ctx-submit}", flagSalt),
			}, nil
		},
	}
	service := wirePracticeSubmissionAdapters(
		NewService(
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

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if _, err := service.SubmitFlag(ctx, 7, 11, "flag{ctx-submit}"); err != nil {
		t.Fatalf("SubmitFlag() error = %v", err)
	}
	if !challengeLookupCalled {
		t.Fatal("expected challenge lookup to be called")
	}
	if !findCorrectCalled {
		t.Fatal("expected find correct submission repository to be called")
	}
	if !createSubmissionCalled {
		t.Fatal("expected create submission repository to be called")
	}
}

func TestSubmitFlagPropagatesContextToDynamicFlagInstanceLookup(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("dynamic-flag")
	expectedCtxValue := "ctx-dynamic-flag"
	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()
	instanceLookupCalled := false
	instanceStore := &stubPracticeInstanceStore{
		findByUserAndChallengeWithContextFn: func(ctx context.Context, userID, challengeID int64) (*instanceentity.Instance, error) {
			instanceLookupCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected dynamic flag instance lookup ctx value %v, got %v", expectedCtxValue, got)
			}
			return &instanceentity.Instance{ID: 301, UserID: userID, ChallengeID: challengeID, Nonce: "nonce-301"}, nil
		},
	}
	repo := &stubPracticeRepository{
		findCorrectSubmissionFn: func(context.Context, int64, int64) (*practiceports.SubmissionRecord, error) {
			return nil, gorm.ErrRecordNotFound
		},
		createSubmissionFn: func(context.Context, *practiceports.SubmissionRecord) error {
			return nil
		},
	}
	challengeRepo := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			return &challengecontracts.PracticeRuntimeChallenge{
				ID:         id,
				Category:   taxonomy.DimensionWeb,
				Points:     100,
				Status:     challengecontracts.ChallengeStatusPublished,
				FlagType:   challengecontracts.FlagTypeDynamic,
				FlagPrefix: "flag",
			}, nil
		},
	}
	service := wirePracticeSubmissionAdapters(
		NewService(
			repo,

			nil,
			instanceStore,
			nil,
			nil,
			newPracticeFlagSubmitRateLimitStoreForTest(redisClient),
			&config.Config{
				RateLimit: config.RateLimitConfig{
					RedisKeyPrefix: "practice:test",
					FlagSubmit:     config.RateLimitPolicyConfig{Limit: 5, Window: time.Minute},
				},
				Container: config.ContainerConfig{FlagGlobalSecret: "12345678901234567890123456789012"},
			},
			nil),

		repo,
		challengeRepo,
	)

	flag := flagcrypto.GenerateDynamicFlag(7, 11, "12345678901234567890123456789012", "nonce-301", "flag")
	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if _, err := service.SubmitFlag(ctx, 7, 11, flag); err != nil {
		t.Fatalf("SubmitFlag() error = %v", err)
	}
	if !instanceLookupCalled {
		t.Fatal("expected dynamic flag instance lookup to be called")
	}
}

func TestSubmitFlagPropagatesContextToSolveGraceInstanceUpdates(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("solve-grace")
	expectedCtxValue := "ctx-solve-grace"
	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()
	lookupCalled := false
	refreshCalled := false
	instanceStore := &stubPracticeInstanceStore{
		findByUserAndChallengeWithContextFn: func(ctx context.Context, userID, challengeID int64) (*instanceentity.Instance, error) {
			lookupCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected solve grace lookup ctx value %v, got %v", expectedCtxValue, got)
			}
			return &instanceentity.Instance{ID: 401, UserID: userID, ChallengeID: challengeID, ShareScope: instanceentity.ShareScopePerUser, ExpiresAt: time.Now().Add(2 * time.Hour)}, nil
		},
		refreshInstanceExpiryWithContextFn: func(ctx context.Context, instanceID int64, expiresAt time.Time) error {
			refreshCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected solve grace refresh ctx value %v, got %v", expectedCtxValue, got)
			}
			if instanceID != 401 {
				t.Fatalf("unexpected instance id: %d", instanceID)
			}
			return nil
		},
	}
	flagSalt := "solve-grace-ctx"
	repo := &stubPracticeRepository{
		findCorrectSubmissionFn: func(context.Context, int64, int64) (*practiceports.SubmissionRecord, error) {
			return nil, gorm.ErrRecordNotFound
		},
		createSubmissionFn: func(context.Context, *practiceports.SubmissionRecord) error {
			return nil
		},
	}
	challengeRepo := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			return &challengecontracts.PracticeRuntimeChallenge{
				ID:              id,
				Category:        taxonomy.DimensionWeb,
				Points:          100,
				Status:          challengecontracts.ChallengeStatusPublished,
				FlagType:        challengecontracts.FlagTypeStatic,
				FlagSalt:        flagSalt,
				FlagHash:        flagcrypto.HashStaticFlag("flag{solve-grace-ctx}", flagSalt),
				InstanceSharing: challengecontracts.InstanceSharingPerUser,
			}, nil
		},
	}
	service := wirePracticeSubmissionAdapters(
		NewService(
			repo,

			nil,
			instanceStore,
			nil,
			nil,
			newPracticeFlagSubmitRateLimitStoreForTest(redisClient),
			&config.Config{
				RateLimit: config.RateLimitConfig{
					RedisKeyPrefix: "practice:test",
					FlagSubmit:     config.RateLimitPolicyConfig{Limit: 5, Window: time.Minute},
				},
				Container: config.ContainerConfig{SolveGracePeriod: 10 * time.Minute},
			},
			nil),

		repo,
		challengeRepo,
	)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if _, err := service.SubmitFlag(ctx, 7, 11, "flag{solve-grace-ctx}"); err != nil {
		t.Fatalf("SubmitFlag() error = %v", err)
	}
	if !lookupCalled {
		t.Fatal("expected solve grace instance lookup to be called")
	}
	if !refreshCalled {
		t.Fatal("expected solve grace refresh to be called")
	}
}
