package commands

import (
	"context"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	practiceports "ctf-platform/internal/module/practice/ports"
	"ctf-platform/internal/module/practice/testsupport/contestentity"
	"ctf-platform/internal/platform/events"
	"ctf-platform/internal/shared/flagcrypto"
	"ctf-platform/internal/shared/taxonomy"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"testing"
	"time"
)

func TestPracticePublishesFlagAcceptedEvent(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&contestentity.Submission{}); err != nil {
		t.Fatalf("migrate submissions: %v", err)
	}

	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })
	flagSalt := "static-salt"

	bus := events.NewBus()
	repo := &stubPracticeRepository{
		findCorrectSubmissionFn: func(ctx context.Context, userID, challengeID int64) (*practiceports.SubmissionRecord, error) {
			return nil, gorm.ErrRecordNotFound
		},
		createSubmissionFn: func(ctx context.Context, submission *practiceports.SubmissionRecord) error {
			entity := contestSubmissionEntityFromPracticeRecord(submission)
			if err := db.Create(entity).Error; err != nil {
				return err
			}
			submission.ID = entity.ID
			return nil
		},
	}
	challengeRepo := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			return &challengecontracts.PracticeRuntimeChallenge{
				ID:       id,
				Category: taxonomy.DimensionWeb,
				Points:   100,
				Status:   challengecontracts.ChallengeStatusPublished,
				FlagType: challengecontracts.FlagTypeStatic,
				FlagSalt: flagSalt,
				FlagHash: flagcrypto.HashStaticFlag("flag{correct}", flagSalt),
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
				Cache: config.CacheConfig{
					ProgressTTL: time.Minute,
				},
			},
			nil),

		repo,
		challengeRepo,
	)

	service.SetEventBus(bus)

	received := make(chan practicecontracts.FlagAcceptedEvent, 1)
	bus.Subscribe(practicecontracts.EventFlagAccepted, func(_ context.Context, evt events.Event) error {
		payload, ok := evt.Payload.(practicecontracts.FlagAcceptedEvent)
		if !ok {
			t.Fatalf("unexpected payload type: %T", evt.Payload)
		}
		received <- payload
		return nil
	})

	resp, err := service.SubmitFlag(context.Background(), 7, 11, "flag{correct}")
	if err != nil {
		t.Fatalf("SubmitFlag() error = %v", err)
	}
	if !resp.IsCorrect {
		t.Fatalf("expected correct submission response, got %+v", resp)
	}

	select {
	case evt := <-received:
		if evt.UserID != 7 || evt.ChallengeID != 11 || evt.Dimension != taxonomy.DimensionWeb {
			t.Fatalf("unexpected event payload: %+v", evt)
		}
	case <-time.After(time.Second):
		t.Fatal("expected practice.flag_accepted event to be published")
	}
}

func TestSubmitFlagAllowsRepeatCorrectSubmissionWithoutExtraPoints(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	flagSalt := "repeat-submit-salt"

	if err := db.Create(&identitycontracts.User{
		ID:        71,
		Username:  "student-repeat",
		Role:      identitycontracts.RoleStudent,
		Status:    identitycontracts.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

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
				FlagType: challengecontracts.FlagTypeStatic,
				FlagSalt: flagSalt,
				FlagHash: flagcrypto.HashStaticFlag("flag{repeatable}", flagSalt),
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

	first, err := service.SubmitFlag(context.Background(), 71, 11, "flag{repeatable}")
	if err != nil {
		t.Fatalf("SubmitFlag() first error = %v", err)
	}
	if !first.IsCorrect || first.Points != 100 {
		t.Fatalf("expected first correct submission to score once, got %+v", first)
	}

	repeat, err := service.SubmitFlag(context.Background(), 71, 11, "flag{repeatable}")
	if err != nil {
		t.Fatalf("SubmitFlag() repeat error = %v", err)
	}
	if !repeat.IsCorrect || repeat.Status != SubmissionStatusCorrect {
		t.Fatalf("expected repeated correct submission to stay correct, got %+v", repeat)
	}
	if repeat.Points != 0 {
		t.Fatalf("expected repeated correct submission not to award points, got %+v", repeat)
	}

	var count int64
	if err := db.Model(&contestentity.Submission{}).
		Where("user_id = ? AND challenge_id = ?", 71, 11).
		Count(&count).Error; err != nil {
		t.Fatalf("count submissions: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected repeated correct submission not to create extra record, got %d", count)
	}
}

func TestSubmitFlagShrinksOwnedInstanceExpiryAfterSolve(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	flagSalt := "solve-grace-salt"

	if err := db.Create(&identitycontracts.User{
		ID:        7,
		Username:  "student7",
		Role:      identitycontracts.RoleStudent,
		Status:    identitycontracts.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	originalExpiry := now.Add(2 * time.Hour)
	if err := db.Create(&instanceentity.Instance{
		ID:          1001,
		UserID:      7,
		ChallengeID: 11,
		Status:      instanceentity.InstanceStatusRunning,
		ShareScope:  instanceentity.ShareScopePerUser,
		ExpiresAt:   originalExpiry,
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create instance: %v", err)
	}

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
				FlagHash:        flagcrypto.HashStaticFlag("flag{correct}", flagSalt),
				InstanceSharing: challengecontracts.InstanceSharingPerUser,
			}, nil
		},
	}
	service := wirePracticeSubmissionAdapters(
		NewService(
			repo,

			nil,
			newPracticeTestInstanceRepository(db),
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
				Container: config.ContainerConfig{
					SolveGracePeriod: 10 * time.Minute,
				},
			},
			nil),

		repo,
		challengeRepo,
	)

	beforeSubmit := time.Now()
	resp, err := service.SubmitFlag(context.Background(), 7, 11, "flag{correct}")
	if err != nil {
		t.Fatalf("SubmitFlag() error = %v", err)
	}
	if !resp.IsCorrect {
		t.Fatalf("expected correct submission response, got %+v", resp)
	}
	if resp.InstanceShutdownAt == nil {
		t.Fatalf("expected shutdown hint, got %+v", resp)
	}
	if resp.Message != "" {
		t.Fatalf("expected practice submit message to be omitted, got %q", resp.Message)
	}

	expectedMax := beforeSubmit.Add(10*time.Minute + 5*time.Second)
	expectedMin := beforeSubmit.Add(9*time.Minute + 50*time.Second)
	if resp.InstanceShutdownAt.Before(expectedMin) || resp.InstanceShutdownAt.After(expectedMax) {
		t.Fatalf("unexpected shutdown time: got %v, want around %v", resp.InstanceShutdownAt, beforeSubmit.Add(10*time.Minute))
	}

	var stored instanceentity.Instance
	if err := db.First(&stored, 1001).Error; err != nil {
		t.Fatalf("load instance: %v", err)
	}
	if !stored.ExpiresAt.Equal(*resp.InstanceShutdownAt) {
		t.Fatalf("expected instance expiry to match response: stored=%v response=%v", stored.ExpiresAt, *resp.InstanceShutdownAt)
	}
	if !stored.ExpiresAt.Before(originalExpiry) {
		t.Fatalf("expected instance expiry to shrink: before=%v after=%v", originalExpiry, stored.ExpiresAt)
	}
}
