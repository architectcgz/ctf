package commands

import (
	"context"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	"ctf-platform/internal/module/practice/testsupport/contestentity"
	"ctf-platform/internal/platform/events"
	"ctf-platform/internal/shared/flagcrypto"
	"ctf-platform/internal/shared/taxonomy"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestSubmitFlagEnqueuesFlagAcceptedOutboxEvent(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })
	flagSalt := "static-salt"

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
				FlagHash: flagcrypto.HashStaticFlag("flag{correct}", flagSalt),
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
				Cache: config.CacheConfig{
					ProgressTTL: time.Minute,
				},
			},
			nil),

		repo,
		challengeRepo,
	)

	resp, err := service.SubmitFlag(context.Background(), 7, 11, "flag{correct}")
	if err != nil {
		t.Fatalf("SubmitFlag() error = %v", err)
	}
	if !resp.IsCorrect {
		t.Fatalf("expected correct submission response, got %+v", resp)
	}

	var records []events.OutboxRecord
	if err := db.Order("id ASC").Find(&records).Error; err != nil {
		t.Fatalf("query outbox records: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected one outbox record, got %d", len(records))
	}
	record := records[0]
	if record.EventName != practicecontracts.EventFlagAccepted {
		t.Fatalf("unexpected outbox event name: %s", record.EventName)
	}
	if record.PayloadVersion != 1 {
		t.Fatalf("unexpected outbox payload version: %d", record.PayloadVersion)
	}
	if record.Route != events.OutboxRouteHandler {
		t.Fatalf("unexpected outbox route: %s", record.Route)
	}
	if record.DedupeKey != "practice:flag_accepted:7:11" {
		t.Fatalf("unexpected outbox dedupe key: %s", record.DedupeKey)
	}

	codec := events.NewOutboxCodec()
	codec.Register(practicecontracts.EventFlagAccepted, 1, func() any { return &practicecontracts.FlagAcceptedEvent{} })
	decoded, err := codec.Decode(events.OutboxEvent{
		Name:           record.EventName,
		PayloadVersion: record.PayloadVersion,
		Payload:        record.Payload,
		Route:          record.Route,
		DedupeKey:      record.DedupeKey,
		OccurredAt:     record.OccurredAt,
	})
	if err != nil {
		t.Fatalf("decode outbox payload: %v", err)
	}
	payload, ok := decoded.Payload.(*practicecontracts.FlagAcceptedEvent)
	if !ok {
		t.Fatalf("unexpected decoded payload type: %T", decoded.Payload)
	}
	if payload.UserID != 7 || payload.ChallengeID != 11 || payload.Dimension != taxonomy.DimensionWeb || payload.Points != 100 {
		t.Fatalf("unexpected outbox payload: %+v", payload)
	}
	if !payload.OccurredAt.Equal(resp.SubmittedAt) {
		t.Fatalf("expected outbox occurred_at to match submission time: payload=%v resp=%v", payload.OccurredAt, resp.SubmittedAt)
	}
}

func TestSubmitFlagRollsBackSubmissionWhenFlagAcceptedOutboxEnqueueFails(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })
	flagSalt := "static-salt"

	repo := newPracticeRepositoryWithRuntimePortOwner(db)
	if err := db.Callback().Create().Before("gorm:create").Register("fail_platform_event_outbox_insert", func(tx *gorm.DB) {
		if tx.Statement != nil && tx.Statement.Table == "platform_event_outbox" {
			tx.AddError(gorm.ErrInvalidDB)
		}
	}); err != nil {
		t.Fatalf("register outbox failure callback: %v", err)
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

	_, err := service.SubmitFlag(context.Background(), 7, 11, "flag{correct}")
	if err == nil {
		t.Fatal("expected SubmitFlag to fail when outbox enqueue fails")
	}

	var submissions int64
	if err := db.Model(&contestentity.Submission{}).Count(&submissions).Error; err != nil {
		t.Fatalf("count submissions: %v", err)
	}
	if submissions != 0 {
		t.Fatalf("expected submission insert to roll back, got %d submissions", submissions)
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
		newServiceCore(
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
