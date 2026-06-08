package jobs_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	contestentity "ctf-platform/internal/module/contest/entity"
	instanceentity "ctf-platform/internal/module/instance/entity"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
)

func TestAWDRoundUpdaterHTTPStandardKeepsAWDAliasHostWhenDialingRuntimeIP(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Date(2026, 3, 10, 12, 11, 0, 0, time.UTC)

	createAWDContestFixture(t, db, 142, now)
	createAWDRoundFixture(t, db, 14201, 142, 1, 50, 40, now)
	createAWDChallengeFixture(t, db, 142001, now)
	createAWDContestChallengeFixture(t, db, 142, 142001, now)
	createAWDTeamFixture(t, db, 142011, 142, "Alias", now)
	createAWDTeamMemberFixture(t, db, 142, 142011, 6421, now)

	if err := db.Model(&contestJobChallengeRow{}).Where("id = ?", 142001).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}
	syncAWDContestServiceFixture(t, db, 142, 142001, "alias-service", contestentity.AWDCheckerTypeHTTPStandard, `{
				"put_flag":{"method":"PUT","path":"/api/flag","body_template":"{{FLAG}}","expected_status":200},
				"get_flag":{"method":"GET","path":"/api/flag","expected_status":200,"expected_substring":"{{FLAG}}"}
			}`, 100, 1, 2, now)

	storedFlag := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Host, "awd-c142-t142011-s") {
			http.Error(w, "unexpected host", http.StatusMisdirectedRequest)
			return
		}
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

	parsedURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}
	aliasAccessURL := "http://awd-c142-t142011-s142001:" + parsedURL.Port()
	runtimeDetails, err := runtimecontracts.EncodeInstanceRuntimeDetails(runtimecontracts.InstanceRuntimeDetails{
		Networks: []runtimecontracts.InstanceRuntimeNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-142", Shared: true},
		},
		Containers: []runtimecontracts.InstanceRuntimeContainer{
			{
				ContainerID:    "ctr-http-alias",
				IsEntryPoint:   true,
				NetworkKeys:    []string{runtimecontracts.TopologyDefaultNetworkKey},
				NetworkAliases: []string{"awd-c142-t142011-s142001"},
				NetworkIPs:     map[string]string{"ctf-awd-contest-142": "127.0.0.1"},
			},
		},
	})
	if err != nil {
		t.Fatalf("encode runtime details: %v", err)
	}

	if err := db.Create(&instanceentity.Instance{
		ID:             9421,
		UserID:         6421,
		ChallengeID:    142001,
		ServiceID:      awdServiceIDPtr(142, 142001),
		ContainerID:    "ctr-http-alias",
		Status:         instanceentity.InstanceStatusRunning,
		AccessURL:      aliasAccessURL,
		RuntimeDetails: runtimeDetails,
		ExpiresAt:      now.Add(time.Hour),
		CreatedAt:      now,
		UpdatedAt:      now,
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
	}, "http-alias-secret", nil, zap.NewNop())

	if err := updater.SyncRoundServiceChecks(context.Background(), &contestentity.Contest{ID: 142}, 1); err != nil {
		t.Fatalf("syncRoundServiceChecks() error = %v", err)
	}

	var record contestentity.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 14201, 142011, 142001).First(&record).Error; err != nil {
		t.Fatalf("load service check: %v", err)
	}
	if record.ServiceStatus != contestentity.AWDServiceStatusUp {
		t.Fatalf("expected alias-backed service up, got status=%s result=%s", record.ServiceStatus, record.CheckResult)
	}
	if !strings.Contains(record.CheckResult, aliasAccessURL) {
		t.Fatalf("expected check result to keep alias access url, got %s", record.CheckResult)
	}
}
