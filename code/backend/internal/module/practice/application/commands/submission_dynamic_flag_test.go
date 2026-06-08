package commands

import (
	"context"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
	practiceentity "ctf-platform/internal/module/practice/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
	flagcrypto "ctf-platform/internal/shared/flagcrypto"
	"ctf-platform/internal/shared/taxonomy"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"testing"
	"time"
)

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

func TestSubmitFlagValidatesDynamicFlagWithInstanceNonce(t *testing.T) {
	t.Parallel()

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()
	instanceStore := &stubPracticeInstanceStore{
		findByUserAndChallengeWithContextFn: func(ctx context.Context, userID, challengeID int64) (*instanceentity.Instance, error) {
			return &instanceentity.Instance{
				ID:          302,
				UserID:      userID,
				ChallengeID: challengeID,
				Nonce:       "nonce-302",
			}, nil
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
					RedisKeyPrefix: "practice:dynamic-flag",
					FlagSubmit:     config.RateLimitPolicyConfig{Limit: 5, Window: time.Minute},
				},
				Container: config.ContainerConfig{
					FlagGlobalSecret: "dynamic-secret-12345678901234567890",
				},
			},
			nil),
		repo,
		challengeRepo,
	)

	flag := flagcrypto.GenerateDynamicFlag(7, 11, "dynamic-secret-12345678901234567890", "nonce-302", "flag")
	resp, err := service.SubmitFlag(context.Background(), 7, 11, flag)
	if err != nil {
		t.Fatalf("SubmitFlag() error = %v", err)
	}
	if resp == nil || !resp.IsCorrect {
		t.Fatalf("expected dynamic flag validated with instance nonce to be correct, got %+v", resp)
	}
}

func TestBuildInstanceFlagUsesGlobalSecret(t *testing.T) {
	t.Parallel()

	service := NewService(nil, nil, nil, nil, nil, nil, &config.Config{
		Container: config.ContainerConfig{
			FlagGlobalSecret: "active-secret-12345678901234567890",
		},
	}, nil)

	flag, nonce, err := service.buildInstanceFlag(7, 11, &practiceentity.Challenge{
		ID:         11,
		FlagType:   practiceentity.FlagTypeDynamic,
		FlagPrefix: "flag",
	})
	if err != nil {
		t.Fatalf("buildInstanceFlag() error = %v", err)
	}
	if nonce == "" {
		t.Fatal("expected nonce to be generated")
	}
	expected := flagcrypto.GenerateDynamicFlag(7, 11, "active-secret-12345678901234567890", nonce, "flag")
	if flag != expected {
		t.Fatalf("flag = %q, want %q", flag, expected)
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
