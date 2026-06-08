package jobs_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/config"
	contestentity "ctf-platform/internal/module/contest/entity"
	rediskeys "ctf-platform/internal/module/contest/infrastructure/cachekeys"
	instanceentity "ctf-platform/internal/module/instance/entity"
)

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

	if err := db.Create(&instanceentity.Instance{
		ID:          9801,
		UserID:      5801,
		ChallengeID: 108001,
		ServiceID:   awdServiceIDPtr(108, 108001),
		ContainerID: "ctr-history",
		Status:      instanceentity.InstanceStatusRunning,
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
	if err := redisClient.HSet(context.Background(), rediskeys.AWDServiceStatusKey(108), field, contestentity.AWDServiceStatusCompromised).Err(); err != nil {
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
	setAWDHTTPRuntimeForTest(updater, server.Client(), time.Second)

	if err := updater.RunRoundServiceChecks(context.Background(), &contestentity.Contest{ID: 108}, &contestentity.AWDRound{ID: 10801, ContestID: 108, RoundNumber: 1}, awdCheckSourceManualSelected); err != nil {
		t.Fatalf("RunRoundServiceChecks() error = %v", err)
	}

	assertAWDServiceStatusCache(t, redisClient, 108, 108011, serviceID, contestentity.AWDServiceStatusCompromised)

	var record contestentity.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 10801, 108011, 108001).First(&record).Error; err != nil {
		t.Fatalf("load historical service check: %v", err)
	}
	if record.ServiceStatus != contestentity.AWDServiceStatusUp {
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

	if err := db.Create(&instanceentity.Instance{
		ID:          9901,
		UserID:      5901,
		ChallengeID: 109001,
		ServiceID:   awdServiceIDPtr(109, 109001),
		ContainerID: "ctr-current",
		Status:      instanceentity.InstanceStatusRunning,
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
	if err := redisClient.HSet(context.Background(), rediskeys.AWDServiceStatusKey(109), field, contestentity.AWDServiceStatusDown).Err(); err != nil {
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
	setAWDHTTPRuntimeForTest(updater, server.Client(), time.Second)

	if err := updater.RunRoundServiceChecks(context.Background(), &contestentity.Contest{ID: 109}, &contestentity.AWDRound{ID: 10902, ContestID: 109, RoundNumber: 2}, awdCheckSourceManualCurrent); err != nil {
		t.Fatalf("RunRoundServiceChecks() error = %v", err)
	}

	assertAWDServiceStatusCache(t, redisClient, 109, 109011, serviceID, contestentity.AWDServiceStatusUp)
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

	if err := db.Create(&instanceentity.Instance{
		ID:          10001,
		UserID:      6001,
		ChallengeID: 110001,
		ServiceID:   awdServiceIDPtr(110, 110001),
		ContainerID: "ctr-stale-pointer",
		Status:      instanceentity.InstanceStatusRunning,
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
	if err := redisClient.HSet(context.Background(), rediskeys.AWDServiceStatusKey(110), field, contestentity.AWDServiceStatusCompromised).Err(); err != nil {
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
	setAWDHTTPRuntimeForTest(updater, server.Client(), time.Second)

	if err := updater.RunRoundServiceChecks(context.Background(), &contestentity.Contest{ID: 110}, &contestentity.AWDRound{ID: 11001, ContestID: 110, RoundNumber: 1}, awdCheckSourceManualSelected); err != nil {
		t.Fatalf("RunRoundServiceChecks() error = %v", err)
	}

	assertAWDServiceStatusCache(t, redisClient, 110, 110011, serviceID, contestentity.AWDServiceStatusCompromised)
}
