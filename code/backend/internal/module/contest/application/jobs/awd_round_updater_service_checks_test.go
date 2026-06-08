package jobs_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	contestentity "ctf-platform/internal/module/contest/entity"
	instanceentity "ctf-platform/internal/module/instance/entity"
)

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

	if err := db.Create(&instanceentity.Instance{
		ID:          9301,
		UserID:      5301,
		ChallengeID: 103001,
		ServiceID:   awdServiceIDPtr(103, 103001),
		ContainerID: "ctr-up",
		Status:      instanceentity.InstanceStatusRunning,
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
	setAWDHTTPRuntimeForTest(updater, server.Client(), time.Second)

	if err := updater.SyncRoundServiceChecks(context.Background(), &contestentity.Contest{ID: 103}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record contestentity.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 10301, 103011, 103001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceStatus != contestentity.AWDServiceStatusUp {
		t.Fatalf("unexpected service status: %s", record.ServiceStatus)
	}
	if record.DefenseScore != 40 {
		t.Fatalf("unexpected defense score: %d", record.DefenseScore)
	}
	if record.SLAScore != 0 || record.CheckerType != contestentity.AWDCheckerTypeLegacyProbe {
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

	syncAWDContestServiceFixture(t, db, 104, 104001, "awd-service", contestentity.AWDCheckerTypeHTTPStandard, `{"get_flag":{"path":"/internal/flag"}}`, 100, 1, 2, now)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/internal/flag":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	if err := db.Create(&instanceentity.Instance{
		ID:          9401,
		UserID:      5401,
		ChallengeID: 104001,
		ServiceID:   awdServiceIDPtr(104, 104001),
		ContainerID: "ctr-config",
		Status:      instanceentity.InstanceStatusRunning,
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
	setAWDHTTPRuntimeForTest(updater, server.Client(), time.Second)

	if err := updater.SyncRoundServiceChecks(context.Background(), &contestentity.Contest{ID: 104}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record contestentity.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 10401, 104011, 104001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceStatus != contestentity.AWDServiceStatusUp || record.DefenseScore != 2 || record.SLAScore != 1 || record.CheckerType != contestentity.AWDCheckerTypeHTTPStandard {
		t.Fatalf("unexpected configured service record: %+v", record)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(record.CheckResult), &result); err != nil {
		t.Fatalf("unmarshal check result: %v", err)
	}
	if result["health_path"] != "/internal/flag" {
		t.Fatalf("unexpected configured health path: %#v", result["health_path"])
	}
	if result["checker_type"] != string(contestentity.AWDCheckerTypeHTTPStandard) {
		t.Fatalf("unexpected checker type: %#v", result["checker_type"])
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
	if err := db.Create(&instanceentity.Instance{
		ID:          9501,
		UserID:      5501,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 105001,
		ServiceID:   awdServiceIDPtr(105, 105001),
		ContainerID: "ctr-team-scoped",
		Status:      instanceentity.InstanceStatusRunning,
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
	setAWDHTTPRuntimeForTest(updater, server.Client(), time.Second)

	if err := updater.SyncRoundServiceChecks(context.Background(), &contestentity.Contest{ID: 105}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record contestentity.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 10501, 105011, 105001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceStatus != contestentity.AWDServiceStatusUp {
		t.Fatalf("unexpected service status: %s", record.ServiceStatus)
	}
}
