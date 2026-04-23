package commands

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	flagcrypto "ctf-platform/pkg/crypto"
)

func TestSubmitFlagWithContextRequestsAuditSkipForRepeatCorrectSubmission(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	flagSalt := "repeat-audit-salt"

	if err := db.Create(&model.User{
		ID:        71,
		Username:  "student71",
		Role:      model.RoleStudent,
		Status:    model.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&model.Challenge{
		ID:        11,
		Category:  model.DimensionWeb,
		Points:    100,
		Status:    model.ChallengeStatusPublished,
		FlagType:  model.FlagTypeStatic,
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

	service := NewService(
		practiceinfra.NewRepository(db),
		challengeinfra.NewRepository(db),
		nil,
		nil,
		nil,
		nil,
		nil,
		redisClient,
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
	)

	if _, err := service.SubmitFlagWithContext(context.Background(), 71, 11, "flag{repeatable}"); err != nil {
		t.Fatalf("SubmitFlagWithContext() first error = %v", err)
	}

	control := &auditlog.Control{}
	ctx := auditlog.WithControl(context.Background(), control)

	if _, err := service.SubmitFlagWithContext(ctx, 71, 11, "flag{repeatable}"); err != nil {
		t.Fatalf("SubmitFlagWithContext() repeat error = %v", err)
	}
	if !control.Skip {
		t.Fatal("expected repeat correct submission to request audit skip")
	}
}
