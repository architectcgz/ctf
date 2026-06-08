package jobs_test

import (
	"context"
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

func TestAWDRoundUpdaterPrefersContestAWDServiceDefinitionsForRuntimeChecks(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)

	createAWDContestFixture(t, db, 144, now)
	createAWDRoundFixture(t, db, 14401, 144, 1, 50, 40, now)
	createAWDChallengeFixture(t, db, 144001, now)
	createAWDContestChallengeFixture(t, db, 144, 144001, now)
	createAWDTeamFixture(t, db, 144011, 144, "ServiceFirst", now)
	createAWDTeamMemberFixture(t, db, 144, 144011, 6441, now)

	if err := db.Model(&contestJobChallengeRow{}).Where("id = ?", 144001).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}
	serviceID := defaultAWDContestServiceID(144, 144001)
	if err := db.Model(&contestentity.ContestAWDService{}).
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

	if err := db.Create(&instanceentity.Instance{
		ID:          9441,
		UserID:      6441,
		ChallengeID: 144001,
		ServiceID:   awdServiceIDPtr(144, 144001),
		ContainerID: "ctr-service-first",
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
	}, "service-first-secret", nil, zap.NewNop())
	setAWDHTTPRuntimeForTest(updater, server.Client(), time.Second)

	if err := updater.SyncRoundServiceChecks(context.Background(), &contestentity.Contest{ID: 144}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record contestentity.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 14401, 144011, 144001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceID != serviceID {
		t.Fatalf("expected persisted service_id=%d, got %+v", serviceID, record)
	}
	if record.ServiceStatus != contestentity.AWDServiceStatusUp || record.SLAScore != 1 || record.DefenseScore != 2 || record.CheckerType != contestentity.AWDCheckerTypeHTTPStandard {
		t.Fatalf("expected runtime check to prefer contest_awd_services, got %+v", record)
	}
}
