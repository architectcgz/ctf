package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	"ctf-platform/internal/module/contest/testsupport"
	rediskeys "ctf-platform/internal/pkg/redis"
)

func TestAWDRoundUpdaterRefreshesSchedulerLockWhileRunning(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	updater := &AWDRoundUpdater{
		redis: redisClient,
		cfg: config.ContestAWDConfig{
			SchedulerLockTTL: 60 * time.Millisecond,
		},
		log: zap.NewNop(),
	}

	lockKey := rediskeys.AWDSchedulerLockKey()
	updater.withSchedulerLock(context.Background(), func(ctx context.Context) {
		time.Sleep(140 * time.Millisecond)
		if ctx.Err() != nil {
			t.Fatalf("expected scheduler lock context to remain active, got %v", ctx.Err())
		}
		if !mini.Exists(lockKey) {
			t.Fatalf("expected scheduler lock %q to be refreshed during run", lockKey)
		}
		if ttl := mini.TTL(lockKey); ttl <= 0 {
			t.Fatalf("expected positive scheduler lock ttl during run, got %s", ttl)
		}
	})

	if mini.Exists(lockKey) {
		t.Fatalf("expected scheduler lock %q to be released after run", lockKey)
	}
}

func TestAWDRoundUpdaterSyncContestRoundsSkipsCanceledContextErrorLog(t *testing.T) {
	t.Parallel()

	db := testsupport.SetupAWDTestDB(t)

	now := time.Date(2026, 5, 9, 21, 30, 0, 0, time.UTC)
	contestID := int64(901)
	testsupport.CreateAWDContestFixture(t, db, contestID, now.Add(-11*time.Minute))
	if err := db.Model(&model.Contest{}).Where("id = ?", contestID).Updates(map[string]any{
		"start_time": now.Add(-11 * time.Minute),
		"end_time":   now.Add(14 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("update contest time window: %v", err)
	}

	var contest model.Contest
	if err := db.First(&contest, contestID).Error; err != nil {
		t.Fatalf("load contest: %v", err)
	}

	core, recorded := observer.New(zap.DebugLevel)
	updater := NewAWDRoundUpdater(
		contestinfra.NewAWDRepository(db),
		nil,
		config.ContestAWDConfig{
			RoundInterval: 5 * time.Minute,
			RoundLockTTL:  time.Minute,
		},
		"",
		nil,
		zap.New(core),
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	updater.syncContestRounds(ctx, &contest, now)

	if entries := recorded.FilterMessage("sync_awd_rounds_failed").All(); len(entries) != 0 {
		t.Fatalf("expected canceled context to skip sync_awd_rounds_failed log, got %d entries", len(entries))
	}
}
