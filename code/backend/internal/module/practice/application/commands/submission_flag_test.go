package commands

import (
	"context"
	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	practiceports "ctf-platform/internal/module/practice/ports"
	"ctf-platform/internal/shared/flagcrypto"
	"ctf-platform/internal/shared/taxonomy"
	"errors"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestSubmitFlagWithRegexChallengeMatchesPattern(t *testing.T) {
	t.Parallel()

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()

	repo := &stubPracticeRepository{}
	challengeRepo := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			return &challengecontracts.PracticeRuntimeChallenge{
				ID:        id,
				Category:  taxonomy.DimensionWeb,
				Points:    80,
				Status:    challengecontracts.ChallengeStatusPublished,
				FlagType:  challengecontracts.FlagTypeRegex,
				FlagRegex: `^flag\{regex-[0-9]{2}\}$`,
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
			nil),

		repo,
		challengeRepo,
	)

	resp, err := service.SubmitFlag(context.Background(), 9, 19, "flag{regex-42}")
	if err != nil {
		t.Fatalf("SubmitFlag() error = %v", err)
	}
	if !resp.IsCorrect || resp.Status != SubmissionStatusCorrect {
		t.Fatalf("expected regex submission success, got %+v", resp)
	}
}

func TestSubmitFlagWithSharedStaticChallengeUsesRegularFlagValidation(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	flagSalt := "shared-static-salt"

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()

	repo := newPracticeRepositoryWithRuntimePortOwner(db)
	challengeRepo := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			return &challengecontracts.PracticeRuntimeChallenge{
				ID:              id,
				Category:        taxonomy.DimensionWeb,
				Points:          100,
				Status:          challengecontracts.ChallengeStatusPublished,
				FlagType:        challengecontracts.FlagTypeStatic,
				FlagSalt:        flagSalt,
				FlagHash:        flagcrypto.HashStaticFlag("flag{shared-static}", flagSalt),
				InstanceSharing: challengecontracts.InstanceSharingShared,
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
			nil),

		repo,
		challengeRepo,
	)

	resp, err := service.SubmitFlag(context.Background(), 7, 11, "flag{shared-static}")
	if err != nil {
		t.Fatalf("SubmitFlag() error = %v", err)
	}
	if !resp.IsCorrect || resp.Status != SubmissionStatusCorrect {
		t.Fatalf("expected shared static submission success, got %+v", resp)
	}
}

func TestSubmitFlagRejectsUnknownFlagType(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()

	repo := newPracticeRepositoryWithRuntimePortOwner(db)
	challengeRepo := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			return &challengecontracts.PracticeRuntimeChallenge{
				ID:       id,
				Category: taxonomy.DimensionWeb,
				Points:   100,
				Status:   challengecontracts.ChallengeStatusPublished,
				FlagType: "shared_proof",
			}, nil
		},
	}
	service := wirePracticeSubmissionAdapters(
		newServiceCore(
			repo,

			nil,
			&stubPracticeInstanceStore{},
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

	_, err := service.SubmitFlag(context.Background(), 7, 11, "flag{legacy}")
	if err == nil || err.Error() != apperror.ErrInvalidParams.Error() {
		t.Fatalf("expected invalid params for unknown flag type, got %v", err)
	}
}

func TestSubmitFlagTreatsPracticeChallengeNotFoundAsChallengeNotFound(t *testing.T) {
	t.Parallel()

	runtimeSubjectSource := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(context.Context, int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}

	service := wirePracticeSubmissionAdapters(
		newServiceCore(&stubPracticeRepository{}, nil, nil, nil, nil, nil, &config.Config{}, nil),
		nil,
		runtimeSubjectSource,
	)

	_, err := service.SubmitFlag(context.Background(), 7, 11, "flag{missing}")
	if err == nil {
		t.Fatal("expected challenge not found")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != challengecontracts.ErrChallengeNotFound.Code {
		t.Fatalf("expected challenge not found error, got %v", err)
	}
}

func TestSubmitFlagTreatsPracticeSolvedSubmissionNotFoundAsUnsolved(t *testing.T) {
	t.Parallel()

	flagSalt := "submission-sentinel-salt"
	runtimeSubjectSource := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			return &challengecontracts.PracticeRuntimeChallenge{
				ID:       id,
				Category: taxonomy.DimensionWeb,
				Points:   100,
				Status:   challengecontracts.ChallengeStatusPublished,
				FlagType: challengecontracts.FlagTypeStatic,
				FlagSalt: flagSalt,
				FlagHash: flagcrypto.HashStaticFlag("flag{sentinel-correct}", flagSalt),
			}, nil
		},
	}

	createCalled := false
	service := wirePracticeSubmissionAdapters(
		newServiceCore(
			&stubPracticeRepository{
				findCorrectSubmissionFn: func(context.Context, int64, int64) (*practiceports.SubmissionRecord, error) {
					return nil, errors.New("raw solved submission repo should not be called")
				},
				createSubmissionFn: func(context.Context, *practiceports.SubmissionRecord) error {
					createCalled = true
					return nil
				},
			},

			nil,
			nil,
			nil,
			nil,
			nil,
			&config.Config{},
			nil),

		&stubPracticeRepository{
			findCorrectSubmissionFn: func(context.Context, int64, int64) (*practiceports.SubmissionRecord, error) {
				return nil, gorm.ErrRecordNotFound
			},
		},
		runtimeSubjectSource,
	)

	resp, err := service.SubmitFlag(context.Background(), 7, 11, "flag{sentinel-correct}")
	if err != nil {
		t.Fatalf("SubmitFlag() error = %v", err)
	}
	if !resp.IsCorrect {
		t.Fatalf("expected correct submission, got %+v", resp)
	}
	if !createCalled {
		t.Fatal("expected submission to be created")
	}
}
