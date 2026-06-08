package jobs_test

import (
	"context"
	"ctf-platform/internal/config"
	contestentity "ctf-platform/internal/module/contest/entity"
	rediskeys "ctf-platform/internal/module/contest/infrastructure/cachekeys"
	contestports "ctf-platform/internal/module/contest/ports"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"strings"
	"testing"
	"time"
)

type captureAWDFlagInjector struct {
	assignments [][]contestports.AWDFlagAssignment
}

func (i *captureAWDFlagInjector) InjectRoundFlags(_ context.Context, _ *contestentity.Contest, _ *contestentity.AWDRound, assignments []contestports.AWDFlagAssignment) error {
	cloned := append([]contestports.AWDFlagAssignment(nil), assignments...)
	i.assignments = append(i.assignments, cloned)
	return nil
}

func TestAWDRoundUpdaterCreatesAndAdvancesRounds(t *testing.T) {
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
	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)
	createAWDContestFixture(t, db, 101, now.Add(-11*time.Minute))
	createAWDChallengeFixture(t, db, 1001, now)
	createAWDContestChallengeFixture(t, db, 101, 1001, now)
	createAWDTeamFixture(t, db, 10011, 101, "Alpha", now)
	if err := db.Model(&contestentity.Contest{}).Where("id = ?", 101).Updates(map[string]any{
		"start_time": now.Add(-11 * time.Minute),
		"end_time":   now.Add(14 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("update contest time window: %v", err)
	}
	if err := db.Model(&contestJobChallengeRow{}).Where("id = ?", 1001).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("update challenge flag prefix: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, redisClient, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      roundInterval,
		RoundLockTTL:       time.Minute,
	}, "test-flag-secret", nil, zap.NewNop())

	updater.UpdateRoundsAt(context.Background(), now)

	var rounds []contestentity.AWDRound
	if err := db.Order("round_number ASC").Find(&rounds, "contest_id = ?", 101).Error; err != nil {
		t.Fatalf("list rounds: %v", err)
	}
	if len(rounds) != 3 {
		t.Fatalf("expected 3 rounds, got %d", len(rounds))
	}
	if rounds[0].Status != contestentity.AWDRoundStatusFinished || rounds[0].StartedAt == nil || rounds[0].EndedAt == nil {
		t.Fatalf("unexpected round 1: %+v", rounds[0])
	}
	if rounds[1].Status != contestentity.AWDRoundStatusFinished || rounds[1].StartedAt == nil || rounds[1].EndedAt == nil {
		t.Fatalf("unexpected round 2: %+v", rounds[1])
	}
	if rounds[2].Status != contestentity.AWDRoundStatusRunning || rounds[2].StartedAt == nil || rounds[2].EndedAt != nil {
		t.Fatalf("unexpected round 3: %+v", rounds[2])
	}

	currentRound, err := redisClient.Get(context.Background(), rediskeys.AWDCurrentRoundKey(101)).Result()
	if err != nil {
		t.Fatalf("load current round: %v", err)
	}
	if currentRound != "3" {
		t.Fatalf("unexpected current round: %s", currentRound)
	}

	flags, err := redisClient.HGetAll(context.Background(), rediskeys.AWDRoundFlagsKey(101, rounds[2].ID)).Result()
	if err != nil {
		t.Fatalf("load round flags: %v", err)
	}
	serviceID := defaultAWDContestServiceID(101, 1001)
	serviceField := rediskeys.AWDRoundFlagServiceField(10011, serviceID)
	if !strings.HasPrefix(flags[serviceField], "awd{") {
		t.Fatalf("unexpected service round flag field: %+v", flags)
	}
	if _, ok := flags["10011:1001"]; ok {
		t.Fatalf("expected legacy round flag field removed, got %+v", flags)
	}
}

func TestAWDRoundUpdaterPreservesRotatedCurrentRoundFlags(t *testing.T) {
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
	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)
	contestID := int64(301)
	challengeID := int64(3001)
	teamID := int64(30011)

	createAWDContestFixture(t, db, contestID, now.Add(-11*time.Minute))
	createAWDChallengeFixture(t, db, challengeID, now)
	createAWDContestChallengeFixture(t, db, contestID, challengeID, now)
	createAWDTeamFixture(t, db, teamID, contestID, "Alpha", now)
	if err := db.Model(&contestentity.Contest{}).Where("id = ?", contestID).Updates(map[string]any{
		"start_time": now.Add(-11 * time.Minute),
		"end_time":   now.Add(14 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("update contest time window: %v", err)
	}
	if err := db.Model(&contestJobChallengeRow{}).Where("id = ?", challengeID).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("update challenge flag prefix: %v", err)
	}

	injector := &captureAWDFlagInjector{}
	updater := newAWDRoundUpdaterForTest(db, redisClient, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      roundInterval,
		RoundLockTTL:       time.Second,
	}, "test-flag-secret", injector, zap.NewNop())

	updater.UpdateRoundsAt(context.Background(), now)

	var round contestentity.AWDRound
	if err := db.Where("contest_id = ? AND round_number = ?", contestID, 3).First(&round).Error; err != nil {
		t.Fatalf("load running round: %v", err)
	}

	serviceID := defaultAWDContestServiceID(contestID, challengeID)
	serviceField := rediskeys.AWDRoundFlagServiceField(teamID, serviceID)
	rotatedFlag := "awd{rotated-current-round-flag}"
	if err := redisClient.HSet(context.Background(), rediskeys.AWDRoundFlagsKey(contestID, round.ID), serviceField, rotatedFlag).Err(); err != nil {
		t.Fatalf("seed rotated round flag: %v", err)
	}
	injector.assignments = nil
	mini.FastForward(2 * time.Second)

	updater.UpdateRoundsAt(context.Background(), now.Add(30*time.Second))

	storedFlag, err := redisClient.HGet(context.Background(), rediskeys.AWDRoundFlagsKey(contestID, round.ID), serviceField).Result()
	if err != nil {
		t.Fatalf("load preserved round flag: %v", err)
	}
	if storedFlag != rotatedFlag {
		t.Fatalf("expected rotated round flag to be preserved, got %q want %q", storedFlag, rotatedFlag)
	}
	if len(injector.assignments) == 0 {
		t.Fatal("expected flag injector to receive assignments")
	}
	found := false
	for _, item := range injector.assignments[len(injector.assignments)-1] {
		if item.TeamID == teamID && item.ServiceID == serviceID {
			found = true
			if item.Flag != rotatedFlag {
				t.Fatalf("expected injector to receive rotated flag, got %q want %q", item.Flag, rotatedFlag)
			}
		}
	}
	if !found {
		t.Fatalf("expected injector assignment for team %d service %d", teamID, serviceID)
	}
}
