package contest

import (
	"context"
	"encoding/json"
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

	updater := NewAWDRoundUpdater(db, redisClient, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      roundInterval,
		RoundLockTTL:       time.Minute,
	}, "test-flag-secret", nil, zap.NewNop())

	updater.updateRoundsAt(context.Background(), now)

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
	field := rediskeys.AWDRoundFlagField(10011, 1001)
	if len(flags) != 1 || !strings.HasPrefix(flags[field], "awd{") {
		t.Fatalf("unexpected round flags: %+v", flags)
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

	updater := NewAWDRoundUpdater(db, redisClient, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      roundInterval,
		RoundLockTTL:       time.Minute,
	}, "test-flag-secret", nil, zap.NewNop())

	updater.updateRoundsAt(context.Background(), now)

	var count int64
	if err := db.Model(&model.AWDRound{}).Where("contest_id = ?", 102).Count(&count).Error; err != nil {
		t.Fatalf("count rounds: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no rounds when lock held, got %d", count)
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

	updater := NewAWDRoundUpdater(db, redisClient, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
	}, "test-flag-secret", nil, zap.NewNop())

	updater.updateRoundsAt(context.Background(), now)

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
		ContainerID: "ctr-up",
		Status:      model.InstanceStatusRunning,
		AccessURL:   server.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd instance: %v", err)
	}

	updater := NewAWDRoundUpdater(db, nil, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "test-flag-secret", nil, zap.NewNop())
	updater.httpClient = server.Client()

	if err := updater.syncRoundServiceChecks(context.Background(), &model.Contest{ID: 103}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record model.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND challenge_id = ?", 10301, 103011, 103001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceStatus != model.AWDServiceStatusUp {
		t.Fatalf("unexpected service status: %s", record.ServiceStatus)
	}
	if record.DefenseScore != 40 {
		t.Fatalf("unexpected defense score: %d", record.DefenseScore)
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
		ContainerID: "ctr-team-scoped",
		Status:      model.InstanceStatusRunning,
		AccessURL:   server.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create scoped awd instance: %v", err)
	}

	updater := NewAWDRoundUpdater(db, nil, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "test-flag-secret", nil, zap.NewNop())
	updater.httpClient = server.Client()

	if err := updater.syncRoundServiceChecks(context.Background(), &model.Contest{ID: 105}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record model.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND challenge_id = ?", 10501, 105011, 105001).First(&record).Error; err != nil {
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
		ContainerID: "ctr-history",
		Status:      model.InstanceStatusRunning,
		AccessURL:   server.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd instance: %v", err)
	}

	field := rediskeys.AWDRoundFlagField(108011, 108001)
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(108), "2", 0).Err(); err != nil {
		t.Fatalf("seed current round: %v", err)
	}
	if err := redisClient.HSet(context.Background(), rediskeys.AWDServiceStatusKey(108), field, model.AWDServiceStatusCompromised).Err(); err != nil {
		t.Fatalf("seed live status cache: %v", err)
	}

	updater := NewAWDRoundUpdater(db, redisClient, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "test-flag-secret", nil, zap.NewNop())
	updater.httpClient = server.Client()

	if err := updater.RunRoundServiceChecks(context.Background(), &model.Contest{ID: 108}, &model.AWDRound{ID: 10801, ContestID: 108, RoundNumber: 1}, awdCheckSourceManualSelected); err != nil {
		t.Fatalf("RunRoundServiceChecks() error = %v", err)
	}

	assertAWDServiceStatusCache(t, redisClient, 108, 108011, 108001, model.AWDServiceStatusCompromised)

	var record model.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND challenge_id = ?", 10801, 108011, 108001).First(&record).Error; err != nil {
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
		ContainerID: "ctr-current",
		Status:      model.InstanceStatusRunning,
		AccessURL:   server.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd instance: %v", err)
	}

	field := rediskeys.AWDRoundFlagField(109011, 109001)
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(109), "2", 0).Err(); err != nil {
		t.Fatalf("seed current round: %v", err)
	}
	if err := redisClient.HSet(context.Background(), rediskeys.AWDServiceStatusKey(109), field, model.AWDServiceStatusDown).Err(); err != nil {
		t.Fatalf("seed live status cache: %v", err)
	}

	updater := NewAWDRoundUpdater(db, redisClient, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "test-flag-secret", nil, zap.NewNop())
	updater.httpClient = server.Client()

	if err := updater.RunRoundServiceChecks(context.Background(), &model.Contest{ID: 109}, &model.AWDRound{ID: 10902, ContestID: 109, RoundNumber: 2}, awdCheckSourceManualCurrent); err != nil {
		t.Fatalf("RunRoundServiceChecks() error = %v", err)
	}

	assertAWDServiceStatusCache(t, redisClient, 109, 109011, 109001, model.AWDServiceStatusUp)
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
		ContainerID: "ctr-stale-pointer",
		Status:      model.InstanceStatusRunning,
		AccessURL:   server.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd instance: %v", err)
	}

	field := rediskeys.AWDRoundFlagField(110011, 110001)
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(110), "1", 0).Err(); err != nil {
		t.Fatalf("seed stale current round: %v", err)
	}
	if err := redisClient.HSet(context.Background(), rediskeys.AWDServiceStatusKey(110), field, model.AWDServiceStatusCompromised).Err(); err != nil {
		t.Fatalf("seed live status cache: %v", err)
	}

	updater := NewAWDRoundUpdater(db, redisClient, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "test-flag-secret", nil, zap.NewNop())
	updater.httpClient = server.Client()

	if err := updater.RunRoundServiceChecks(context.Background(), &model.Contest{ID: 110}, &model.AWDRound{ID: 11001, ContestID: 110, RoundNumber: 1}, awdCheckSourceManualSelected); err != nil {
		t.Fatalf("RunRoundServiceChecks() error = %v", err)
	}

	assertAWDServiceStatusCache(t, redisClient, 110, 110011, 110001, model.AWDServiceStatusCompromised)
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
		ContainerID: "ctr-partial-fail",
		Status:      model.InstanceStatusRunning,
		AccessURL:   failedServer.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create failed awd instance: %v", err)
	}

	updater := NewAWDRoundUpdater(db, nil, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "test-flag-secret", nil, zap.NewNop())
	updater.httpClient = healthyServer.Client()

	if err := updater.syncRoundServiceChecks(context.Background(), &model.Contest{ID: 107}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record model.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND challenge_id = ?", 10701, 107011, 107001).First(&record).Error; err != nil {
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

	updater := NewAWDRoundUpdater(db, nil, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "test-flag-secret", nil, zap.NewNop())

	if err := updater.syncRoundServiceChecks(context.Background(), &model.Contest{ID: 104}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record model.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND challenge_id = ?", 10401, 104011, 104001).First(&record).Error; err != nil {
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
		ContainerID: "ctr-fallback",
		Status:      model.InstanceStatusRunning,
		AccessURL:   server.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd instance: %v", err)
	}

	updater := NewAWDRoundUpdater(db, nil, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "test-flag-secret", nil, zap.NewNop())
	updater.httpClient = server.Client()

	if err := updater.syncRoundServiceChecks(context.Background(), &model.Contest{ID: 106}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record model.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND challenge_id = ?", 10601, 106011, 106001).First(&record).Error; err != nil {
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
