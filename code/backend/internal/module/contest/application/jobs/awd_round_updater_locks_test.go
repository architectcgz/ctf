package jobs_test

import (
	"context"
	"ctf-platform/internal/config"
	contestentity "ctf-platform/internal/module/contest/entity"
	rediskeys "ctf-platform/internal/module/contest/infrastructure/cachekeys"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestAWDRoundUpdaterSkipsWhenRoundLockHeld(t *testing.T) {
	db := newAWDTestDB(t)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	roundInterval := 5 * time.Minute
	now := time.Date(2026, 3, 10, 12, 6, 0, 0, time.UTC)
	createAWDContestFixture(t, db, 102, now.Add(-6*time.Minute))
	if err := db.Model(&contestentity.Contest{}).Where("id = ?", 102).Updates(map[string]any{
		"start_time": now.Add(-6 * time.Minute),
		"end_time":   now.Add(24 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("update contest time window: %v", err)
	}

	lockKey := rediskeys.AWDRoundLockKey(102, 2)
	if err := mini.Set(lockKey, "1"); err != nil {
		t.Fatalf("seed round lock: %v", err)
	}
	mini.SetTTL(lockKey, time.Minute)

	updater := newAWDRoundUpdaterForTest(db, redisClient, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      roundInterval,
		RoundLockTTL:       time.Minute,
	}, "test-flag-secret", nil, zap.NewNop())

	updater.UpdateRoundsAt(context.Background(), now)

	var count int64
	if err := db.Model(&contestentity.AWDRound{}).Where("contest_id = ?", 102).Count(&count).Error; err != nil {
		t.Fatalf("count rounds: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no rounds when lock held, got %d", count)
	}
}

func TestAWDRoundUpdaterSkipsWhenSchedulerLockHeld(t *testing.T) {
	db := newAWDTestDB(t)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	now := time.Date(2026, 3, 10, 12, 6, 0, 0, time.UTC)
	createAWDContestFixture(t, db, 152, now.Add(-6*time.Minute))
	if err := db.Model(&contestentity.Contest{}).Where("id = ?", 152).Updates(map[string]any{
		"start_time": now.Add(-6 * time.Minute),
		"end_time":   now.Add(24 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("update contest time window: %v", err)
	}

	if err := mini.Set(rediskeys.AWDSchedulerLockKey(), "busy"); err != nil {
		t.Fatalf("seed scheduler lock: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, redisClient, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerLockTTL:   time.Minute,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
	}, "test-flag-secret", nil, zap.NewNop())

	updater.UpdateRoundsAt(context.Background(), now)

	var count int64
	if err := db.Model(&contestentity.AWDRound{}).Where("contest_id = ?", 152).Count(&count).Error; err != nil {
		t.Fatalf("count rounds: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no rounds when scheduler lock held, got %d", count)
	}
}
