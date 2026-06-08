package jobs_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	contestentity "ctf-platform/internal/module/contest/entity"
	instanceentity "ctf-platform/internal/module/instance/entity"
)

func TestAWDRoundUpdaterMarksHTTPStandardChecksCompromisedOnFlagMismatch(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)

	createAWDContestFixture(t, db, 142, now)
	createAWDRoundFixture(t, db, 14201, 142, 1, 50, 40, now)
	createAWDChallengeFixture(t, db, 142001, now)
	createAWDContestChallengeFixture(t, db, 142, 142001, now)
	createAWDTeamFixture(t, db, 142011, 142, "Mismatch", now)
	createAWDTeamMemberFixture(t, db, 142, 142011, 6421, now)

	if err := db.Model(&contestJobChallengeRow{}).Where("id = ?", 142001).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}
	syncAWDContestServiceFixture(t, db, 142, 142001, "awd-service", contestentity.AWDCheckerTypeHTTPStandard, `{
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

	if err := db.Create(&instanceentity.Instance{
		ID:          9421,
		UserID:      6421,
		ChallengeID: 142001,
		ServiceID:   awdServiceIDPtr(142, 142001),
		ContainerID: "ctr-http-mismatch",
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
	}, "http-secret", nil, zap.NewNop())
	setAWDHTTPRuntimeForTest(updater, server.Client(), time.Second)

	if err := updater.SyncRoundServiceChecks(context.Background(), &contestentity.Contest{ID: 142}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record contestentity.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 14201, 142011, 142001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceStatus != contestentity.AWDServiceStatusCompromised || record.SLAScore != 0 || record.DefenseScore != 0 {
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

	if err := db.Model(&contestJobChallengeRow{}).Where("id = ?", 143001).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}
	syncAWDContestServiceFixture(t, db, 143, 143001, "awd-service", contestentity.AWDCheckerTypeHTTPStandard, `{
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

	if err := db.Create(&instanceentity.Instance{
		ID:          9431,
		UserID:      6431,
		ChallengeID: 143001,
		ServiceID:   awdServiceIDPtr(143, 143001),
		ContainerID: "ctr-http-havoc",
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
	}, "http-secret", nil, zap.NewNop())
	setAWDHTTPRuntimeForTest(updater, server.Client(), time.Second)

	if err := updater.SyncRoundServiceChecks(context.Background(), &contestentity.Contest{ID: 143}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record contestentity.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 14301, 143011, 143001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceStatus != contestentity.AWDServiceStatusDown || record.SLAScore != 0 || record.DefenseScore != 0 {
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
