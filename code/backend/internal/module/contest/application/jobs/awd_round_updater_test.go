package jobs_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	rediskeys "ctf-platform/internal/pkg/redis"
)

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
	if err := db.Model(&model.Contest{}).Where("id = ?", 101).Updates(map[string]any{
		"start_time": now.Add(-11 * time.Minute),
		"end_time":   now.Add(14 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("update contest time window: %v", err)
	}
	if err := db.Model(&model.Challenge{}).Where("id = ?", 1001).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("update challenge flag prefix: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, redisClient, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      roundInterval,
		RoundLockTTL:       time.Minute,
	}, "test-flag-secret", nil, zap.NewNop())

	updater.UpdateRoundsAt(context.Background(), now)

	var rounds []model.AWDRound
	if err := db.Order("round_number ASC").Find(&rounds, "contest_id = ?", 101).Error; err != nil {
		t.Fatalf("list rounds: %v", err)
	}
	if len(rounds) != 3 {
		t.Fatalf("expected 3 rounds, got %d", len(rounds))
	}
	if rounds[0].Status != model.AWDRoundStatusFinished || rounds[0].StartedAt == nil || rounds[0].EndedAt == nil {
		t.Fatalf("unexpected round 1: %+v", rounds[0])
	}
	if rounds[1].Status != model.AWDRoundStatusFinished || rounds[1].StartedAt == nil || rounds[1].EndedAt == nil {
		t.Fatalf("unexpected round 2: %+v", rounds[1])
	}
	if rounds[2].Status != model.AWDRoundStatusRunning || rounds[2].StartedAt == nil || rounds[2].EndedAt != nil {
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

func TestAWDRoundUpdaterSkipsRegistrationContest(t *testing.T) {
	db := newAWDTestDB(t)

	roundInterval := 5 * time.Minute
	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)
	contestID := int64(158)
	createAWDContestFixture(t, db, contestID, now.Add(-11*time.Minute))
	if err := db.Model(&model.Contest{}).Where("id = ?", contestID).Updates(map[string]any{
		"status":     model.ContestStatusRegistration,
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
	if err := db.Model(&model.AWDRound{}).Where("contest_id = ?", contestID).Count(&count).Error; err != nil {
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
	if err := db.Model(&model.Contest{}).Where("id = ?", contestID).Updates(map[string]any{
		"status":     model.ContestStatusEnded,
		"start_time": startedAt,
		"end_time":   startedAt.Add(2 * roundInterval),
	}).Error; err != nil {
		t.Fatalf("update contest time window: %v", err)
	}
	createAWDRoundFixtureWithWindow(t, db, 15501, contestID, 1, 70, 35, startedAt, time.Time{})
	if err := db.Model(&model.AWDRound{}).Where("id = ?", 15501).Update("ended_at", nil).Error; err != nil {
		t.Fatalf("clear stale round end: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(contestID), "1", 0).Err(); err != nil {
		t.Fatalf("seed current round: %v", err)
	}
	if err := redisClient.HSet(context.Background(), rediskeys.AWDServiceStatusKey(contestID), "1:s:1", model.AWDServiceStatusUp).Err(); err != nil {
		t.Fatalf("seed service status: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, redisClient, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      roundInterval,
		RoundLockTTL:       time.Minute,
	}, "test-flag-secret", nil, zap.NewNop())

	updater.UpdateRoundsAt(context.Background(), now)

	var rounds []model.AWDRound
	if err := db.Order("round_number ASC").Find(&rounds, "contest_id = ?", contestID).Error; err != nil {
		t.Fatalf("list rounds: %v", err)
	}
	if len(rounds) != 2 {
		t.Fatalf("expected finalized contest to have 2 rounds, got %d: %+v", len(rounds), rounds)
	}
	for i, round := range rounds {
		if round.Status != model.AWDRoundStatusFinished || round.EndedAt == nil {
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
	if err := db.Model(&model.Contest{}).Where("id = ?", contestID).Updates(map[string]any{
		"status":     model.ContestStatusEnded,
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

	var rounds []model.AWDRound
	if err := db.Order("round_number ASC").Find(&rounds, "contest_id = ?", contestID).Error; err != nil {
		t.Fatalf("list rounds: %v", err)
	}
	if len(rounds) != 2 {
		t.Fatalf("expected finalized contest to have 2 rounds, got %d: %+v", len(rounds), rounds)
	}
	for i, round := range rounds {
		if round.Status != model.AWDRoundStatusFinished || round.EndedAt == nil {
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
	if err := db.Model(&model.Contest{}).Where("id = ?", contestID).Updates(map[string]any{
		"status":     model.ContestStatusEnded,
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

	var rounds []model.AWDRound
	if err := db.Order("round_number ASC").Find(&rounds, "contest_id = ?", contestID).Error; err != nil {
		t.Fatalf("list rounds: %v", err)
	}
	if len(rounds) != 2 {
		t.Fatalf("expected finalized contest to have 2 rounds, got %d: %+v", len(rounds), rounds)
	}
	if rounds[1].Status != model.AWDRoundStatusFinished || rounds[1].EndedAt == nil || !rounds[1].EndedAt.Equal(contestEnd) {
		t.Fatalf("expected recovered final round to end at contest end, got %+v", rounds[1])
	}
}

func TestAWDRoundUpdaterCreatesAndAdvancesRoundsWritesOnlyServiceFlagFields(t *testing.T) {
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
	createAWDContestFixture(t, db, 153, now.Add(-11*time.Minute))
	createAWDChallengeFixture(t, db, 153001, now)
	createAWDContestChallengeFixture(t, db, 153, 153001, now)
	createAWDTeamFixture(t, db, 153011, 153, "Alpha", now)
	if err := db.Model(&model.Contest{}).Where("id = ?", 153).Updates(map[string]any{
		"start_time": now.Add(-11 * time.Minute),
		"end_time":   now.Add(14 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("update contest time window: %v", err)
	}
	if err := db.Model(&model.Challenge{}).Where("id = ?", 153001).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("update challenge flag prefix: %v", err)
	}
	serviceID := defaultAWDContestServiceID(153, 153001)
	if err := db.Model(&model.ContestAWDService{}).
		Where("contest_id = ? AND awd_challenge_id = ?", 153, 153001).
		Updates(map[string]any{
			"display_name":   "Bridge Service",
			"order":          0,
			"is_visible":     true,
			"score_config":   `{"points":100,"awd_sla_score":1,"awd_defense_score":2}`,
			"runtime_config": `{"awd_challenge_id":153001,"checker_type":"legacy_probe","checker_config":{}}`,
			"updated_at":     now,
		}).Error; err != nil {
		t.Fatalf("update contest awd service: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, redisClient, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      roundInterval,
		RoundLockTTL:       time.Minute,
	}, "test-flag-secret", nil, zap.NewNop())

	updater.UpdateRoundsAt(context.Background(), now)

	var rounds []model.AWDRound
	if err := db.Order("round_number ASC").Find(&rounds, "contest_id = ?", 153).Error; err != nil {
		t.Fatalf("list rounds: %v", err)
	}
	if len(rounds) != 3 {
		t.Fatalf("expected 3 rounds, got %d", len(rounds))
	}

	flags, err := redisClient.HGetAll(context.Background(), rediskeys.AWDRoundFlagsKey(153, rounds[2].ID)).Result()
	if err != nil {
		t.Fatalf("load round flags: %v", err)
	}
	serviceField := rediskeys.AWDRoundFlagServiceField(153011, serviceID)
	if !strings.HasPrefix(flags[serviceField], "awd{") {
		t.Fatalf("expected service round flag field, got %+v", flags)
	}
	if _, ok := flags["153011:153001"]; ok {
		t.Fatalf("expected legacy round flag field removed, got %+v", flags)
	}
}

func TestAWDRoundUpdaterIgnoresLegacyContestChallengeBridgeWithoutServiceDefinition(t *testing.T) {
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

	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)
	createAWDContestFixture(t, db, 154, now)
	createAWDRoundFixture(t, db, 15401, 154, 1, 50, 40, now)
	createAWDChallengeFixture(t, db, 154001, now)
	createAWDTeamFixture(t, db, 154011, 154, "LegacyOnly", now)
	createAWDTeamMemberFixture(t, db, 154, 154011, 154101, now)
	if err := db.Create(&model.ContestChallenge{
		ContestID:   154,
		ChallengeID: 154001,
		Points:      100,
		IsVisible:   true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create legacy-only contest challenge: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	if err := db.Create(&model.Instance{
		ID:          154901,
		UserID:      154101,
		ChallengeID: 154001,
		ContainerID: "ctr-legacy-only",
		Status:      model.InstanceStatusRunning,
		AccessURL:   server.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd instance: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, redisClient, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "test-flag-secret", nil, zap.NewNop())
	updater.SetHTTPClient(server.Client())

	if err := updater.RunRoundServiceChecks(context.Background(), &model.Contest{ID: 154}, &model.AWDRound{ID: 15401, ContestID: 154, RoundNumber: 1}, awdCheckSourceManualCurrent); err != nil {
		t.Fatalf("RunRoundServiceChecks() error = %v", err)
	}

	var count int64
	if err := db.Model(&model.AWDTeamService{}).Where("round_id = ?", 15401).Count(&count).Error; err != nil {
		t.Fatalf("count service checks: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no service checks without contest_awd_services definition, got %d", count)
	}
}

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
	if err := db.Model(&model.Contest{}).Where("id = ?", 102).Updates(map[string]any{
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
	if err := db.Model(&model.AWDRound{}).Where("contest_id = ?", 102).Count(&count).Error; err != nil {
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
	if err := db.Model(&model.Contest{}).Where("id = ?", 152).Updates(map[string]any{
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
	if err := db.Model(&model.AWDRound{}).Where("contest_id = ?", 152).Count(&count).Error; err != nil {
		t.Fatalf("count rounds: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no rounds when scheduler lock held, got %d", count)
	}
}

func TestAWDRoundUpdaterReconcileRoundsInheritsPreviousRoundScores(t *testing.T) {
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
	createAWDContestFixture(t, db, 111, now.Add(-6*time.Minute))
	createAWDRoundFixtureWithWindow(t, db, 11101, 111, 1, 80, 25, now.Add(-6*time.Minute), now.Add(-time.Minute))
	if err := db.Model(&model.Contest{}).Where("id = ?", 111).Updates(map[string]any{
		"start_time": now.Add(-6 * time.Minute),
		"end_time":   now.Add(14 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("update contest time window: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, redisClient, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
	}, "test-flag-secret", nil, zap.NewNop())

	updater.UpdateRoundsAt(context.Background(), now)

	var round model.AWDRound
	if err := db.Where("contest_id = ? AND round_number = ?", 111, 2).First(&round).Error; err != nil {
		t.Fatalf("load inherited round: %v", err)
	}
	if round.AttackScore != 80 || round.DefenseScore != 25 {
		t.Fatalf("unexpected inherited round scores: %+v", round)
	}
}

func TestAWDRoundUpdaterSyncsServiceChecksAsUp(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)

	createAWDContestFixture(t, db, 103, now)
	createAWDRoundFixture(t, db, 10301, 103, 1, 50, 40, now)
	createAWDChallengeFixture(t, db, 103001, now)
	createAWDContestChallengeFixture(t, db, 103, 103001, now)
	createAWDTeamFixture(t, db, 103011, 103, "Alpha", now)
	createAWDTeamMemberFixture(t, db, 103, 103011, 5301, now)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	if err := db.Create(&model.Instance{
		ID:          9301,
		UserID:      5301,
		ChallengeID: 103001,
		ServiceID:   awdServiceIDPtr(103, 103001),
		ContainerID: "ctr-up",
		Status:      model.InstanceStatusRunning,
		AccessURL:   server.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd instance: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, nil, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "test-flag-secret", nil, zap.NewNop())
	updater.SetHTTPClient(server.Client())

	if err := updater.SyncRoundServiceChecks(context.Background(), &model.Contest{ID: 103}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record model.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 10301, 103011, 103001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceStatus != model.AWDServiceStatusUp {
		t.Fatalf("unexpected service status: %s", record.ServiceStatus)
	}
	if record.DefenseScore != 40 {
		t.Fatalf("unexpected defense score: %d", record.DefenseScore)
	}
	if record.SLAScore != 0 || record.CheckerType != model.AWDCheckerTypeLegacyProbe {
		t.Fatalf("unexpected sla/checker fields: %+v", record)
	}
	if record.CreatedAt.Location() != time.UTC || record.UpdatedAt.Location() != time.UTC {
		t.Fatalf("expected UTC service check timestamps, got created=%v updated=%v", record.CreatedAt.Location(), record.UpdatedAt.Location())
	}
	if !strings.Contains(record.CheckResult, "\"healthy_instance_count\":1") {
		t.Fatalf("unexpected check result: %s", record.CheckResult)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(record.CheckResult), &result); err != nil {
		t.Fatalf("unmarshal check result: %v", err)
	}
	if result["check_source"] != awdCheckSourceScheduler {
		t.Fatalf("unexpected check_source: %#v", result["check_source"])
	}
	if result["status_reason"] != "healthy" {
		t.Fatalf("unexpected status_reason: %#v", result["status_reason"])
	}
	targets, ok := result["targets"].([]any)
	if !ok || len(targets) != 1 {
		t.Fatalf("unexpected targets: %#v", result["targets"])
	}
	target, ok := targets[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected target payload: %#v", targets[0])
	}
	if target["healthy"] != true || target["probe"] != "http" {
		t.Fatalf("unexpected target result: %#v", target)
	}
	attempts, ok := target["attempts"].([]any)
	if !ok || len(attempts) != 1 {
		t.Fatalf("unexpected attempts: %#v", target["attempts"])
	}
}

func TestAWDRoundUpdaterUsesContestServiceCheckerConfig(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)

	createAWDContestFixture(t, db, 104, now)
	createAWDRoundFixture(t, db, 10401, 104, 1, 50, 40, now)
	createAWDChallengeFixture(t, db, 104001, now)
	createAWDContestChallengeFixture(t, db, 104, 104001, now)
	createAWDTeamFixture(t, db, 104011, 104, "Config", now)
	createAWDTeamMemberFixture(t, db, 104, 104011, 5401, now)

	syncAWDContestServiceFixture(t, db, 104, 104001, "awd-service", model.AWDCheckerTypeHTTPStandard, `{"get_flag":{"path":"/internal/flag"}}`, 100, 1, 2, now)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/internal/flag":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	if err := db.Create(&model.Instance{
		ID:          9401,
		UserID:      5401,
		ChallengeID: 104001,
		ServiceID:   awdServiceIDPtr(104, 104001),
		ContainerID: "ctr-config",
		Status:      model.InstanceStatusRunning,
		AccessURL:   server.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd instance: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, nil, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "test-flag-secret", nil, zap.NewNop())
	updater.SetHTTPClient(server.Client())

	if err := updater.SyncRoundServiceChecks(context.Background(), &model.Contest{ID: 104}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record model.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 10401, 104011, 104001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceStatus != model.AWDServiceStatusUp || record.DefenseScore != 2 || record.SLAScore != 1 || record.CheckerType != model.AWDCheckerTypeHTTPStandard {
		t.Fatalf("unexpected configured service record: %+v", record)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(record.CheckResult), &result); err != nil {
		t.Fatalf("unmarshal check result: %v", err)
	}
	if result["health_path"] != "/internal/flag" {
		t.Fatalf("unexpected configured health path: %#v", result["health_path"])
	}
	if result["checker_type"] != string(model.AWDCheckerTypeHTTPStandard) {
		t.Fatalf("unexpected checker type: %#v", result["checker_type"])
	}
}

func TestAWDRoundUpdaterSyncsHTTPStandardChecksAsUp(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)

	createAWDContestFixture(t, db, 141, now)
	createAWDRoundFixture(t, db, 14101, 141, 1, 50, 40, now)
	createAWDChallengeFixture(t, db, 141001, now)
	createAWDContestChallengeFixture(t, db, 141, 141001, now)
	createAWDTeamFixture(t, db, 141011, 141, "HTTP", now)
	createAWDTeamMemberFixture(t, db, 141, 141011, 6411, now)

	if err := db.Model(&model.Challenge{}).Where("id = ?", 141001).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}
	syncAWDContestServiceFixture(t, db, 141, 141001, "awd-service", model.AWDCheckerTypeHTTPStandard, `{
				"put_flag":{"method":"PUT","path":"/api/flag","body_template":"{{FLAG}}","expected_status":200},
				"get_flag":{"method":"GET","path":"/api/flag","expected_status":200,"expected_substring":"{{FLAG}}"}
			}`, 100, 1, 2, now)

	storedFlag := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/api/flag":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read put body: %v", err)
			}
			storedFlag = string(body)
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/flag":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(storedFlag))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	if err := db.Create(&model.Instance{
		ID:          9411,
		UserID:      6411,
		ChallengeID: 141001,
		ServiceID:   awdServiceIDPtr(141, 141001),
		ContainerID: "ctr-http-up",
		Status:      model.InstanceStatusRunning,
		AccessURL:   server.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd instance: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, nil, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "http-secret", nil, zap.NewNop())
	updater.SetHTTPClient(server.Client())

	if err := updater.SyncRoundServiceChecks(context.Background(), &model.Contest{ID: 141}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record model.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 14101, 141011, 141001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceStatus != model.AWDServiceStatusUp || record.SLAScore != 1 || record.DefenseScore != 2 {
		t.Fatalf("unexpected http_standard record: %+v", record)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(record.CheckResult), &result); err != nil {
		t.Fatalf("unmarshal check result: %v", err)
	}
	if result["checker_type"] != string(model.AWDCheckerTypeHTTPStandard) {
		t.Fatalf("unexpected checker_type: %#v", result["checker_type"])
	}
	putFlag, ok := result["put_flag"].(map[string]any)
	if !ok || putFlag["healthy"] != true {
		t.Fatalf("unexpected put_flag result: %#v", result["put_flag"])
	}
	getFlag, ok := result["get_flag"].(map[string]any)
	if !ok || getFlag["healthy"] != true {
		t.Fatalf("unexpected get_flag result: %#v", result["get_flag"])
	}
}

func TestAWDRoundUpdaterPrefersContestAWDServiceDefinitionsForRuntimeChecks(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)

	createAWDContestFixture(t, db, 144, now)
	createAWDRoundFixture(t, db, 14401, 144, 1, 50, 40, now)
	createAWDChallengeFixture(t, db, 144001, now)
	createAWDContestChallengeFixture(t, db, 144, 144001, now)
	createAWDTeamFixture(t, db, 144011, 144, "ServiceFirst", now)
	createAWDTeamMemberFixture(t, db, 144, 144011, 6441, now)

	if err := db.Model(&model.Challenge{}).Where("id = ?", 144001).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}
	serviceID := defaultAWDContestServiceID(144, 144001)
	if err := db.Model(&model.ContestAWDService{}).
		Where("contest_id = ? AND awd_challenge_id = ?", 144, 144001).
		Updates(map[string]any{
			"display_name": "Service First",
			"order":        0,
			"is_visible":   true,
			"score_config": `{"points":100,"awd_sla_score":1,"awd_defense_score":2}`,
			"runtime_config": `{
				"awd_challenge_id":144001,
				"checker_type":"http_standard",
				"checker_config":{
					"put_flag":{"method":"PUT","path":"/api/service-flag","body_template":"{{FLAG}}","expected_status":200},
					"get_flag":{"method":"GET","path":"/api/service-flag","expected_status":200,"expected_substring":"{{FLAG}}"}
				}
			}`,
			"updated_at": now,
		}).Error; err != nil {
		t.Fatalf("update contest awd service: %v", err)
	}

	storedFlag := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/api/service-flag":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read put body: %v", err)
			}
			storedFlag = string(body)
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/service-flag":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(storedFlag))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	if err := db.Create(&model.Instance{
		ID:          9441,
		UserID:      6441,
		ChallengeID: 144001,
		ServiceID:   awdServiceIDPtr(144, 144001),
		ContainerID: "ctr-service-first",
		Status:      model.InstanceStatusRunning,
		AccessURL:   server.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd instance: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, nil, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "service-first-secret", nil, zap.NewNop())
	updater.SetHTTPClient(server.Client())

	if err := updater.SyncRoundServiceChecks(context.Background(), &model.Contest{ID: 144}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record model.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 14401, 144011, 144001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceID != serviceID {
		t.Fatalf("expected persisted service_id=%d, got %+v", serviceID, record)
	}
	if record.ServiceStatus != model.AWDServiceStatusUp || record.SLAScore != 1 || record.DefenseScore != 2 || record.CheckerType != model.AWDCheckerTypeHTTPStandard {
		t.Fatalf("expected runtime check to prefer contest_awd_services, got %+v", record)
	}
}

func TestAWDRoundUpdaterMarksHTTPStandardChecksCompromisedOnFlagMismatch(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)

	createAWDContestFixture(t, db, 142, now)
	createAWDRoundFixture(t, db, 14201, 142, 1, 50, 40, now)
	createAWDChallengeFixture(t, db, 142001, now)
	createAWDContestChallengeFixture(t, db, 142, 142001, now)
	createAWDTeamFixture(t, db, 142011, 142, "Mismatch", now)
	createAWDTeamMemberFixture(t, db, 142, 142011, 6421, now)

	if err := db.Model(&model.Challenge{}).Where("id = ?", 142001).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}
	syncAWDContestServiceFixture(t, db, 142, 142001, "awd-service", model.AWDCheckerTypeHTTPStandard, `{
				"put_flag":{"method":"PUT","path":"/api/flag","body_template":"{{FLAG}}","expected_status":200},
				"get_flag":{"method":"GET","path":"/api/flag","expected_status":200,"expected_substring":"{{FLAG}}"}
			}`, 100, 1, 2, now)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/api/flag":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/flag":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("awd{broken-flag}"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	if err := db.Create(&model.Instance{
		ID:          9421,
		UserID:      6421,
		ChallengeID: 142001,
		ServiceID:   awdServiceIDPtr(142, 142001),
		ContainerID: "ctr-http-mismatch",
		Status:      model.InstanceStatusRunning,
		AccessURL:   server.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd instance: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, nil, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "http-secret", nil, zap.NewNop())
	updater.SetHTTPClient(server.Client())

	if err := updater.SyncRoundServiceChecks(context.Background(), &model.Contest{ID: 142}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record model.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 14201, 142011, 142001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceStatus != model.AWDServiceStatusCompromised || record.SLAScore != 0 || record.DefenseScore != 0 {
		t.Fatalf("unexpected compromised record: %+v", record)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(record.CheckResult), &result); err != nil {
		t.Fatalf("unmarshal check result: %v", err)
	}
	if result["status_reason"] != "flag_mismatch" {
		t.Fatalf("unexpected status_reason: %#v", result["status_reason"])
	}
	getFlag, ok := result["get_flag"].(map[string]any)
	if !ok || getFlag["error_code"] != "flag_mismatch" {
		t.Fatalf("unexpected get_flag result: %#v", result["get_flag"])
	}
}

func TestAWDRoundUpdaterMarksHTTPStandardChecksDownWhenHavocFails(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)

	createAWDContestFixture(t, db, 143, now)
	createAWDRoundFixture(t, db, 14301, 143, 1, 50, 40, now)
	createAWDChallengeFixture(t, db, 143001, now)
	createAWDContestChallengeFixture(t, db, 143, 143001, now)
	createAWDTeamFixture(t, db, 143011, 143, "Havoc", now)
	createAWDTeamMemberFixture(t, db, 143, 143011, 6431, now)

	if err := db.Model(&model.Challenge{}).Where("id = ?", 143001).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}
	syncAWDContestServiceFixture(t, db, 143, 143001, "awd-service", model.AWDCheckerTypeHTTPStandard, `{
				"put_flag":{"method":"PUT","path":"/api/flag","body_template":"{{FLAG}}","expected_status":200},
				"get_flag":{"method":"GET","path":"/api/flag","expected_status":200,"expected_substring":"{{FLAG}}"},
				"havoc":{"method":"GET","path":"/api/ping","expected_status":200}
			}`, 100, 1, 2, now)

	storedFlag := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/api/flag":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read put body: %v", err)
			}
			storedFlag = string(body)
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/flag":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(storedFlag))
		case r.Method == http.MethodGet && r.URL.Path == "/api/ping":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	if err := db.Create(&model.Instance{
		ID:          9431,
		UserID:      6431,
		ChallengeID: 143001,
		ServiceID:   awdServiceIDPtr(143, 143001),
		ContainerID: "ctr-http-havoc",
		Status:      model.InstanceStatusRunning,
		AccessURL:   server.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd instance: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, nil, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "http-secret", nil, zap.NewNop())
	updater.SetHTTPClient(server.Client())

	if err := updater.SyncRoundServiceChecks(context.Background(), &model.Contest{ID: 143}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record model.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 14301, 143011, 143001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceStatus != model.AWDServiceStatusDown || record.SLAScore != 0 || record.DefenseScore != 0 {
		t.Fatalf("unexpected havoc failure record: %+v", record)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(record.CheckResult), &result); err != nil {
		t.Fatalf("unmarshal check result: %v", err)
	}
	havoc, ok := result["havoc"].(map[string]any)
	if !ok || havoc["error_code"] != "unexpected_http_status" {
		t.Fatalf("unexpected havoc result: %#v", result["havoc"])
	}
}

func TestAWDRoundUpdaterSyncsServiceChecksForContestScopedTeamInstance(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)

	createAWDContestFixture(t, db, 105, now)
	createAWDRoundFixture(t, db, 10501, 105, 1, 50, 40, now)
	createAWDChallengeFixture(t, db, 105001, now)
	createAWDContestChallengeFixture(t, db, 105, 105001, now)
	createAWDTeamFixture(t, db, 105011, 105, "Scoped", now)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	contestID := int64(105)
	teamID := int64(105011)
	if err := db.Create(&model.Instance{
		ID:          9501,
		UserID:      5501,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 105001,
		ServiceID:   awdServiceIDPtr(105, 105001),
		ContainerID: "ctr-team-scoped",
		Status:      model.InstanceStatusRunning,
		AccessURL:   server.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create scoped awd instance: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, nil, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "test-flag-secret", nil, zap.NewNop())
	updater.SetHTTPClient(server.Client())

	if err := updater.SyncRoundServiceChecks(context.Background(), &model.Contest{ID: 105}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record model.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 10501, 105011, 105001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceStatus != model.AWDServiceStatusUp {
		t.Fatalf("unexpected service status: %s", record.ServiceStatus)
	}
}

func TestAWDRoundUpdaterHistoricalRoundChecksDoNotOverwriteLiveStatusCache(t *testing.T) {
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

	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)
	createAWDContestFixture(t, db, 108, now)
	createAWDRoundFixtureWithWindow(t, db, 10801, 108, 1, 50, 40, now.Add(-10*time.Minute), now.Add(-5*time.Minute))
	createAWDRoundFixture(t, db, 10802, 108, 2, 50, 40, now)
	createAWDChallengeFixture(t, db, 108001, now)
	createAWDContestChallengeFixture(t, db, 108, 108001, now)
	createAWDTeamFixture(t, db, 108011, 108, "History", now)
	createAWDTeamMemberFixture(t, db, 108, 108011, 5801, now)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	if err := db.Create(&model.Instance{
		ID:          9801,
		UserID:      5801,
		ChallengeID: 108001,
		ServiceID:   awdServiceIDPtr(108, 108001),
		ContainerID: "ctr-history",
		Status:      model.InstanceStatusRunning,
		AccessURL:   server.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd instance: %v", err)
	}

	serviceID := defaultAWDContestServiceID(108, 108001)
	field := rediskeys.AWDRoundFlagServiceField(108011, serviceID)
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(108), "2", 0).Err(); err != nil {
		t.Fatalf("seed current round: %v", err)
	}
	if err := redisClient.HSet(context.Background(), rediskeys.AWDServiceStatusKey(108), field, model.AWDServiceStatusCompromised).Err(); err != nil {
		t.Fatalf("seed live status cache: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, redisClient, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "test-flag-secret", nil, zap.NewNop())
	updater.SetHTTPClient(server.Client())

	if err := updater.RunRoundServiceChecks(context.Background(), &model.Contest{ID: 108}, &model.AWDRound{ID: 10801, ContestID: 108, RoundNumber: 1}, awdCheckSourceManualSelected); err != nil {
		t.Fatalf("RunRoundServiceChecks() error = %v", err)
	}

	assertAWDServiceStatusCache(t, redisClient, 108, 108011, serviceID, model.AWDServiceStatusCompromised)

	var record model.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 10801, 108011, 108001).First(&record).Error; err != nil {
		t.Fatalf("load historical service check: %v", err)
	}
	if record.ServiceStatus != model.AWDServiceStatusUp {
		t.Fatalf("unexpected historical service status: %s", record.ServiceStatus)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(record.CheckResult), &result); err != nil {
		t.Fatalf("unmarshal historical check result: %v", err)
	}
	if result["check_source"] != awdCheckSourceManualSelected {
		t.Fatalf("unexpected check_source: %#v", result["check_source"])
	}
}

func TestAWDRoundUpdaterCurrentRoundChecksRefreshLiveStatusCache(t *testing.T) {
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

	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)
	createAWDContestFixture(t, db, 109, now)
	createAWDRoundFixture(t, db, 10902, 109, 2, 50, 40, now)
	createAWDChallengeFixture(t, db, 109001, now)
	createAWDContestChallengeFixture(t, db, 109, 109001, now)
	createAWDTeamFixture(t, db, 109011, 109, "Current", now)
	createAWDTeamMemberFixture(t, db, 109, 109011, 5901, now)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	if err := db.Create(&model.Instance{
		ID:          9901,
		UserID:      5901,
		ChallengeID: 109001,
		ServiceID:   awdServiceIDPtr(109, 109001),
		ContainerID: "ctr-current",
		Status:      model.InstanceStatusRunning,
		AccessURL:   server.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd instance: %v", err)
	}

	serviceID := defaultAWDContestServiceID(109, 109001)
	field := rediskeys.AWDRoundFlagServiceField(109011, serviceID)
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(109), "2", 0).Err(); err != nil {
		t.Fatalf("seed current round: %v", err)
	}
	if err := redisClient.HSet(context.Background(), rediskeys.AWDServiceStatusKey(109), field, model.AWDServiceStatusDown).Err(); err != nil {
		t.Fatalf("seed live status cache: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, redisClient, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "test-flag-secret", nil, zap.NewNop())
	updater.SetHTTPClient(server.Client())

	if err := updater.RunRoundServiceChecks(context.Background(), &model.Contest{ID: 109}, &model.AWDRound{ID: 10902, ContestID: 109, RoundNumber: 2}, awdCheckSourceManualCurrent); err != nil {
		t.Fatalf("RunRoundServiceChecks() error = %v", err)
	}

	assertAWDServiceStatusCache(t, redisClient, 109, 109011, serviceID, model.AWDServiceStatusUp)
}

func TestAWDRoundUpdaterHistoricalRoundChecksIgnoreStaleCurrentRoundPointer(t *testing.T) {
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

	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)
	createAWDContestFixture(t, db, 110, now)
	createAWDRoundFixtureWithWindow(t, db, 11001, 110, 1, 50, 40, now.Add(-10*time.Minute), now.Add(-5*time.Minute))
	createAWDRoundFixture(t, db, 11002, 110, 2, 50, 40, now)
	createAWDChallengeFixture(t, db, 110001, now)
	createAWDContestChallengeFixture(t, db, 110, 110001, now)
	createAWDTeamFixture(t, db, 110011, 110, "StalePointer", now)
	createAWDTeamMemberFixture(t, db, 110, 110011, 6001, now)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	if err := db.Create(&model.Instance{
		ID:          10001,
		UserID:      6001,
		ChallengeID: 110001,
		ServiceID:   awdServiceIDPtr(110, 110001),
		ContainerID: "ctr-stale-pointer",
		Status:      model.InstanceStatusRunning,
		AccessURL:   server.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd instance: %v", err)
	}

	serviceID := defaultAWDContestServiceID(110, 110001)
	field := rediskeys.AWDRoundFlagServiceField(110011, serviceID)
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(110), "1", 0).Err(); err != nil {
		t.Fatalf("seed stale current round: %v", err)
	}
	if err := redisClient.HSet(context.Background(), rediskeys.AWDServiceStatusKey(110), field, model.AWDServiceStatusCompromised).Err(); err != nil {
		t.Fatalf("seed live status cache: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, redisClient, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "test-flag-secret", nil, zap.NewNop())
	updater.SetHTTPClient(server.Client())

	if err := updater.RunRoundServiceChecks(context.Background(), &model.Contest{ID: 110}, &model.AWDRound{ID: 11001, ContestID: 110, RoundNumber: 1}, awdCheckSourceManualSelected); err != nil {
		t.Fatalf("RunRoundServiceChecks() error = %v", err)
	}

	assertAWDServiceStatusCache(t, redisClient, 110, 110011, serviceID, model.AWDServiceStatusCompromised)
}

func TestAWDRoundUpdaterSyncsServiceChecksWithPartialAvailability(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)

	createAWDContestFixture(t, db, 107, now)
	createAWDRoundFixture(t, db, 10701, 107, 1, 50, 40, now)
	createAWDChallengeFixture(t, db, 107001, now)
	createAWDContestChallengeFixture(t, db, 107, 107001, now)
	createAWDTeamFixture(t, db, 107011, 107, "Partial", now)

	healthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(healthyServer.Close)

	failedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	t.Cleanup(failedServer.Close)

	contestID := int64(107)
	teamID := int64(107011)
	if err := db.Create(&model.Instance{
		ID:          9701,
		UserID:      5701,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 107001,
		ServiceID:   awdServiceIDPtr(107, 107001),
		ContainerID: "ctr-partial-ok",
		Status:      model.InstanceStatusRunning,
		AccessURL:   healthyServer.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create healthy awd instance: %v", err)
	}
	if err := db.Create(&model.Instance{
		ID:          9702,
		UserID:      5702,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 107001,
		ServiceID:   awdServiceIDPtr(107, 107001),
		ContainerID: "ctr-partial-fail",
		Status:      model.InstanceStatusRunning,
		AccessURL:   failedServer.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create failed awd instance: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, nil, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "test-flag-secret", nil, zap.NewNop())
	updater.SetHTTPClient(healthyServer.Client())

	if err := updater.SyncRoundServiceChecks(context.Background(), &model.Contest{ID: 107}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record model.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 10701, 107011, 107001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceStatus != model.AWDServiceStatusUp {
		t.Fatalf("unexpected service status: %s", record.ServiceStatus)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(record.CheckResult), &result); err != nil {
		t.Fatalf("unmarshal check result: %v", err)
	}
	if result["check_source"] != awdCheckSourceScheduler {
		t.Fatalf("unexpected check_source: %#v", result["check_source"])
	}
	if result["status_reason"] != "partial_available" {
		t.Fatalf("unexpected status_reason: %#v", result["status_reason"])
	}
	if result["healthy_instance_count"] != float64(1) || result["failed_instance_count"] != float64(1) {
		t.Fatalf("unexpected instance counts: %#v", result)
	}
	targets, ok := result["targets"].([]any)
	if !ok || len(targets) != 2 {
		t.Fatalf("unexpected targets: %#v", result["targets"])
	}
}

func TestAWDRoundUpdaterSyncsServiceChecksAsDownWithoutHealthyInstance(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)

	createAWDContestFixture(t, db, 104, now)
	createAWDRoundFixture(t, db, 10401, 104, 1, 50, 40, now)
	createAWDChallengeFixture(t, db, 104001, now)
	createAWDContestChallengeFixture(t, db, 104, 104001, now)
	createAWDTeamFixture(t, db, 104011, 104, "Alpha", now)

	updater := newAWDRoundUpdaterForTest(db, nil, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "test-flag-secret", nil, zap.NewNop())

	if err := updater.SyncRoundServiceChecks(context.Background(), &model.Contest{ID: 104}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record model.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 10401, 104011, 104001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceStatus != model.AWDServiceStatusDown {
		t.Fatalf("unexpected service status: %s", record.ServiceStatus)
	}
	if record.DefenseScore != 0 {
		t.Fatalf("unexpected defense score: %d", record.DefenseScore)
	}
	if !strings.Contains(record.CheckResult, "\"error\":\"no_running_instances\"") {
		t.Fatalf("unexpected check result: %s", record.CheckResult)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(record.CheckResult), &result); err != nil {
		t.Fatalf("unmarshal check result: %v", err)
	}
	if result["check_source"] != awdCheckSourceScheduler {
		t.Fatalf("unexpected check_source: %#v", result["check_source"])
	}
	if result["status_reason"] != "no_running_instances" {
		t.Fatalf("unexpected status_reason: %#v", result["status_reason"])
	}
	if result["failed_instance_count"] != float64(0) {
		t.Fatalf("unexpected failed_instance_count: %#v", result["failed_instance_count"])
	}
}

func TestAWDRoundUpdaterMarksServiceDownAfterHTTPFailure(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)

	createAWDContestFixture(t, db, 106, now)
	createAWDRoundFixture(t, db, 10601, 106, 1, 50, 40, now)
	createAWDChallengeFixture(t, db, 106001, now)
	createAWDContestChallengeFixture(t, db, 106, 106001, now)
	createAWDTeamFixture(t, db, 106011, 106, "Fallback", now)
	createAWDTeamMemberFixture(t, db, 106, 106011, 5601, now)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)

	if err := db.Create(&model.Instance{
		ID:          9601,
		UserID:      5601,
		ChallengeID: 106001,
		ServiceID:   awdServiceIDPtr(106, 106001),
		ContainerID: "ctr-fallback",
		Status:      model.InstanceStatusRunning,
		AccessURL:   server.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd instance: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, nil, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "test-flag-secret", nil, zap.NewNop())
	updater.SetHTTPClient(server.Client())

	if err := updater.SyncRoundServiceChecks(context.Background(), &model.Contest{ID: 106}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record model.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 10601, 106011, 106001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceStatus != model.AWDServiceStatusDown {
		t.Fatalf("unexpected service status: %s", record.ServiceStatus)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(record.CheckResult), &result); err != nil {
		t.Fatalf("unmarshal check result: %v", err)
	}
	if result["check_source"] != awdCheckSourceScheduler {
		t.Fatalf("unexpected check_source: %#v", result["check_source"])
	}
	if result["status_reason"] != "unexpected_http_status" {
		t.Fatalf("unexpected status_reason: %#v", result["status_reason"])
	}
	targets, ok := result["targets"].([]any)
	if !ok || len(targets) != 1 {
		t.Fatalf("unexpected targets: %#v", result["targets"])
	}
	target, ok := targets[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected target payload: %#v", targets[0])
	}
	if target["probe"] != "http" || target["healthy"] != false || target["error_code"] != "unexpected_http_status" {
		t.Fatalf("unexpected target result: %#v", target)
	}
	attempts, ok := target["attempts"].([]any)
	if !ok || len(attempts) != 1 {
		t.Fatalf("unexpected attempts: %#v", target["attempts"])
	}
	firstAttempt, ok := attempts[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected first attempt payload: %#v", attempts[0])
	}
	if firstAttempt["probe"] != "http" || firstAttempt["error_code"] != "unexpected_http_status" {
		t.Fatalf("unexpected first attempt: %#v", firstAttempt)
	}
}
