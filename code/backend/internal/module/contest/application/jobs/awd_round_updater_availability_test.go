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
	runtimeentity "ctf-platform/internal/module/contest/entity"
	instanceentity "ctf-platform/internal/module/instance/entity"
)

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
	if err := db.Create(&instanceentity.Instance{
		ID:          9701,
		UserID:      5701,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 107001,
		ServiceID:   awdServiceIDPtr(107, 107001),
		ContainerID: "ctr-partial-ok",
		Status:      instanceentity.InstanceStatusRunning,
		AccessURL:   healthyServer.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create healthy awd instance: %v", err)
	}
	if err := db.Create(&instanceentity.Instance{
		ID:          9702,
		UserID:      5702,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 107001,
		ServiceID:   awdServiceIDPtr(107, 107001),
		ContainerID: "ctr-partial-fail",
		Status:      instanceentity.InstanceStatusRunning,
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
	setAWDHTTPRuntimeForTest(updater, healthyServer.Client(), time.Second)

	if err := updater.SyncRoundServiceChecks(context.Background(), &contestentity.Contest{ID: 107}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record contestentity.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 10701, 107011, 107001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceStatus != contestentity.AWDServiceStatusUp {
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

	if err := updater.SyncRoundServiceChecks(context.Background(), &contestentity.Contest{ID: 104}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record contestentity.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 10401, 104011, 104001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceStatus != contestentity.AWDServiceStatusDown {
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

func TestAWDRoundUpdaterExemptsSLAWhenSystemRecoveryIsActive(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Now().UTC()

	createAWDContestFixture(t, db, 174, now)
	createAWDRoundFixture(t, db, 17401, 174, 1, 50, 40, now)
	createAWDChallengeFixture(t, db, 174001, now)
	createAWDContestChallengeFixture(t, db, 174, 174001, now)
	createAWDTeamFixture(t, db, 174011, 174, "Alpha", now)

	serviceID := defaultAWDContestServiceID(174, 174001)
	if err := db.Model(&contestentity.ContestAWDService{}).
		Where("id = ?", serviceID).
		Update("score_config", `{"points":100,"awd_sla_score":1,"awd_defense_score":2}`).Error; err != nil {
		t.Fatalf("update service score config: %v", err)
	}
	finishedAt := now.Add(time.Hour)
	if err := db.Create(&runtimeentity.AWDServiceOperation{
		ContestID:     174,
		TeamID:        174011,
		ServiceID:     serviceID,
		InstanceID:    174900,
		OperationType: runtimeentity.AWDServiceOperationTypeRecover,
		RequestedBy:   runtimeentity.AWDServiceOperationRequestedBySystem,
		Reason:        "container_not_running",
		SLABillable:   false,
		Status:        runtimeentity.AWDServiceOperationStatusRecovering,
		StartedAt:     now.Add(-time.Minute),
		FinishedAt:    &finishedAt,
		CreatedAt:     now.Add(-time.Minute),
		UpdatedAt:     now.Add(-time.Minute),
	}).Error; err != nil {
		t.Fatalf("create system recovery operation: %v", err)
	}

	updater := newAWDRoundUpdaterForTest(db, nil, config.ContestAWDConfig{
		SchedulerInterval:  time.Second,
		SchedulerBatchSize: 10,
		RoundInterval:      5 * time.Minute,
		RoundLockTTL:       time.Minute,
		CheckerTimeout:     time.Second,
		CheckerHealthPath:  "/health",
	}, "test-flag-secret", nil, zap.NewNop())

	if err := updater.SyncRoundServiceChecks(context.Background(), &contestentity.Contest{ID: 174}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record contestentity.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND service_id = ?", 17401, 174011, serviceID).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceStatus != contestentity.AWDServiceStatusDown || record.SLAScore != 1 || record.DefenseScore != 0 {
		t.Fatalf("expected down service with SLA exemption only, got %+v", record)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(record.CheckResult), &result); err != nil {
		t.Fatalf("unmarshal check result: %v", err)
	}
	if result["sla_exempt"] != true || result["sla_exempt_reason"] != "system_recovery" {
		t.Fatalf("expected SLA exemption marker, got %#v", result)
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

	if err := db.Create(&instanceentity.Instance{
		ID:          9601,
		UserID:      5601,
		ChallengeID: 106001,
		ServiceID:   awdServiceIDPtr(106, 106001),
		ContainerID: "ctr-fallback",
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

	if err := updater.SyncRoundServiceChecks(context.Background(), &contestentity.Contest{ID: 106}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record contestentity.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 10601, 106011, 106001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceStatus != contestentity.AWDServiceStatusDown {
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
