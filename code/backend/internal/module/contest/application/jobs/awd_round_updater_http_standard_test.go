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

func TestAWDRoundUpdaterSyncsHTTPStandardChecksAsUp(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)

	createAWDContestFixture(t, db, 141, now)
	createAWDRoundFixture(t, db, 14101, 141, 1, 50, 40, now)
	createAWDChallengeFixture(t, db, 141001, now)
	createAWDContestChallengeFixture(t, db, 141, 141001, now)
	createAWDTeamFixture(t, db, 141011, 141, "HTTP", now)
	createAWDTeamMemberFixture(t, db, 141, 141011, 6411, now)

	if err := db.Model(&contestJobChallengeRow{}).Where("id = ?", 141001).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}
	syncAWDContestServiceFixture(t, db, 141, 141001, "awd-service", contestentity.AWDCheckerTypeHTTPStandard, `{
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

	if err := db.Create(&instanceentity.Instance{
		ID:          9411,
		UserID:      6411,
		ChallengeID: 141001,
		ServiceID:   awdServiceIDPtr(141, 141001),
		ContainerID: "ctr-http-up",
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

	if err := updater.SyncRoundServiceChecks(context.Background(), &contestentity.Contest{ID: 141}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record contestentity.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 14101, 141011, 141001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceStatus != contestentity.AWDServiceStatusUp || record.SLAScore != 1 || record.DefenseScore != 2 {
		t.Fatalf("unexpected http_standard record: %+v", record)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(record.CheckResult), &result); err != nil {
		t.Fatalf("unmarshal check result: %v", err)
	}
	if result["checker_type"] != string(contestentity.AWDCheckerTypeHTTPStandard) {
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
