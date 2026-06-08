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

func TestAWDRoundUpdaterSkipsRegistrationContest(t *testing.T) {
	db := newAWDTestDB(t)

	roundInterval := 5 * time.Minute
	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)
	contestID := int64(158)
	createAWDContestFixture(t, db, contestID, now.Add(-11*time.Minute))
	if err := db.Model(&contestentity.Contest{}).Where("id = ?", contestID).Updates(map[string]any{
		"status":     contestentity.ContestStatusRegistration,
		"start_time": now.Add(-11 * time.Minute),
		"end_time":   now.Add(14 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("update contest time window: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, nil, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      roundInterval,
		RoundLockTTL:       time.Minute,
	}, "test-flag-secret", nil, zap.NewNop())

	updater.UpdateRoundsAt(context.Background(), now)

	var count int64
	if err := db.Model(&contestentity.AWDRound{}).Where("contest_id = ?", contestID).Count(&count).Error; err != nil {
		t.Fatalf("count rounds: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected registration contest to have no scheduled rounds, got %d", count)
	}
}

func TestAWDRoundUpdaterFinalizesStaleEndedContestAfterLongDowntime(t *testing.T) {
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
	now := time.Date(2026, 3, 10, 12, 30, 0, 0, time.UTC)
	contestID := int64(155)
	startedAt := now.Add(-30 * time.Minute)
	firstRoundEnd := startedAt.Add(roundInterval)
	createAWDContestFixture(t, db, contestID, startedAt)
	if err := db.Model(&contestentity.Contest{}).Where("id = ?", contestID).Updates(map[string]any{
		"status":     contestentity.ContestStatusEnded,
		"start_time": startedAt,
		"end_time":   startedAt.Add(2 * roundInterval),
	}).Error; err != nil {
		t.Fatalf("update contest time window: %v", err)
	}
	createAWDRoundFixtureWithWindow(t, db, 15501, contestID, 1, 70, 35, startedAt, time.Time{})
	if err := db.Model(&contestentity.AWDRound{}).Where("id = ?", 15501).Update("ended_at", nil).Error; err != nil {
		t.Fatalf("clear stale round end: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(contestID), "1", 0).Err(); err != nil {
		t.Fatalf("seed current round: %v", err)
	}
	if err := redisClient.HSet(context.Background(), rediskeys.AWDServiceStatusKey(contestID), "1:s:1", contestentity.AWDServiceStatusUp).Err(); err != nil {
		t.Fatalf("seed service status: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, redisClient, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      roundInterval,
		RoundLockTTL:       time.Minute,
	}, "test-flag-secret", nil, zap.NewNop())

	updater.UpdateRoundsAt(context.Background(), now)

	var rounds []contestentity.AWDRound
	if err := db.Order("round_number ASC").Find(&rounds, "contest_id = ?", contestID).Error; err != nil {
		t.Fatalf("list rounds: %v", err)
	}
	if len(rounds) != 2 {
		t.Fatalf("expected finalized contest to have 2 rounds, got %d: %+v", len(rounds), rounds)
	}
	for i, round := range rounds {
		if round.Status != contestentity.AWDRoundStatusFinished || round.EndedAt == nil {
			t.Fatalf("expected round %d to be finished with ended_at, got %+v", i+1, round)
		}
	}
	if !rounds[0].EndedAt.Equal(firstRoundEnd) {
		t.Fatalf("unexpected first round end: %s", rounds[0].EndedAt)
	}
	if mini.Exists(rediskeys.AWDCurrentRoundKey(contestID)) {
		t.Fatal("expected stale current round key to be cleared")
	}
	if mini.Exists(rediskeys.AWDServiceStatusKey(contestID)) {
		t.Fatal("expected stale service status key to be cleared")
	}
}

func TestAWDRoundUpdaterFinalizesEndedContestWithNoMaterializedRoundsAfterLongDowntime(t *testing.T) {
	db := newAWDTestDB(t)

	roundInterval := 5 * time.Minute
	now := time.Date(2026, 3, 10, 12, 30, 0, 0, time.UTC)
	contestID := int64(156)
	startedAt := now.Add(-30 * time.Minute)
	createAWDContestFixture(t, db, contestID, startedAt)
	if err := db.Model(&contestentity.Contest{}).Where("id = ?", contestID).Updates(map[string]any{
		"status":     contestentity.ContestStatusEnded,
		"start_time": startedAt,
		"end_time":   startedAt.Add(2 * roundInterval),
	}).Error; err != nil {
		t.Fatalf("update contest time window: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, nil, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      roundInterval,
		RoundLockTTL:       time.Minute,
	}, "test-flag-secret", nil, zap.NewNop())

	updater.UpdateRoundsAt(context.Background(), now)

	var rounds []contestentity.AWDRound
	if err := db.Order("round_number ASC").Find(&rounds, "contest_id = ?", contestID).Error; err != nil {
		t.Fatalf("list rounds: %v", err)
	}
	if len(rounds) != 2 {
		t.Fatalf("expected finalized contest to have 2 rounds, got %d: %+v", len(rounds), rounds)
	}
	for i, round := range rounds {
		if round.Status != contestentity.AWDRoundStatusFinished || round.EndedAt == nil {
			t.Fatalf("expected round %d to be finished with ended_at, got %+v", i+1, round)
		}
	}
}

func TestAWDRoundUpdaterCompletesEndedContestWhenFinishedRoundsDoNotReachEnd(t *testing.T) {
	db := newAWDTestDB(t)

	roundInterval := 5 * time.Minute
	now := time.Date(2026, 3, 10, 12, 30, 0, 0, time.UTC)
	contestID := int64(157)
	startedAt := now.Add(-30 * time.Minute)
	firstRoundEnd := startedAt.Add(roundInterval)
	contestEnd := startedAt.Add(2 * roundInterval)
	createAWDContestFixture(t, db, contestID, startedAt)
	if err := db.Model(&contestentity.Contest{}).Where("id = ?", contestID).Updates(map[string]any{
		"status":     contestentity.ContestStatusEnded,
		"start_time": startedAt,
		"end_time":   contestEnd,
	}).Error; err != nil {
		t.Fatalf("update contest time window: %v", err)
	}
	createAWDRoundFixtureWithWindow(t, db, 15701, contestID, 1, 70, 35, startedAt, firstRoundEnd)

	updater := newAWDRoundUpdaterForTest(db, nil, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      roundInterval,
		RoundLockTTL:       time.Minute,
	}, "test-flag-secret", nil, zap.NewNop())

	updater.UpdateRoundsAt(context.Background(), now)

	var rounds []contestentity.AWDRound
	if err := db.Order("round_number ASC").Find(&rounds, "contest_id = ?", contestID).Error; err != nil {
		t.Fatalf("list rounds: %v", err)
	}
	if len(rounds) != 2 {
		t.Fatalf("expected finalized contest to have 2 rounds, got %d: %+v", len(rounds), rounds)
	}
	if rounds[1].Status != contestentity.AWDRoundStatusFinished || rounds[1].EndedAt == nil || !rounds[1].EndedAt.Equal(contestEnd) {
		t.Fatalf("expected recovered final round to end at contest end, got %+v", rounds[1])
	}
}
