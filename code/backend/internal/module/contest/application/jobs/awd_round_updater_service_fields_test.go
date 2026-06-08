package jobs_test

import (
	"context"
	"ctf-platform/internal/config"
	contestentity "ctf-platform/internal/module/contest/entity"
	rediskeys "ctf-platform/internal/module/contest/infrastructure/cachekeys"
	instanceentity "ctf-platform/internal/module/instance/entity"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

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
	if err := db.Model(&contestentity.Contest{}).Where("id = ?", 153).Updates(map[string]any{
		"start_time": now.Add(-11 * time.Minute),
		"end_time":   now.Add(14 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("update contest time window: %v", err)
	}
	if err := db.Model(&contestJobChallengeRow{}).Where("id = ?", 153001).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("update challenge flag prefix: %v", err)
	}
	serviceID := defaultAWDContestServiceID(153, 153001)
	if err := db.Model(&contestentity.ContestAWDService{}).
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

	var rounds []contestentity.AWDRound
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
	if err := db.Create(&contestentity.ContestChallenge{
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

	if err := db.Create(&instanceentity.Instance{
		ID:          154901,
		UserID:      154101,
		ChallengeID: 154001,
		ContainerID: "ctr-legacy-only",
		Status:      instanceentity.InstanceStatusRunning,
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
	setAWDHTTPRuntimeForTest(updater, server.Client(), time.Second)

	if err := updater.RunRoundServiceChecks(context.Background(), &contestentity.Contest{ID: 154}, &contestentity.AWDRound{ID: 15401, ContestID: 154, RoundNumber: 1}, awdCheckSourceManualCurrent); err != nil {
		t.Fatalf("RunRoundServiceChecks() error = %v", err)
	}

	var count int64
	if err := db.Model(&contestentity.AWDTeamService{}).Where("round_id = ?", 15401).Count(&count).Error; err != nil {
		t.Fatalf("count service checks: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no service checks without contest_awd_services definition, got %d", count)
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
	if err := db.Model(&contestentity.Contest{}).Where("id = ?", 111).Updates(map[string]any{
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

	var round contestentity.AWDRound
	if err := db.Where("contest_id = ? AND round_number = ?", 111, 2).First(&round).Error; err != nil {
		t.Fatalf("load inherited round: %v", err)
	}
	if round.AttackScore != 80 || round.DefenseScore != 25 {
		t.Fatalf("unexpected inherited round scores: %+v", round)
	}
}
