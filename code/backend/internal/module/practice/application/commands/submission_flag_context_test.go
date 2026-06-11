package commands

import (
	"context"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
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
			nil,
		),
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
