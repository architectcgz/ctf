package app

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	practicecommands "ctf-platform/internal/module/practice/application/commands"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	"ctf-platform/internal/shared/taxonomy"
)

func TestFullRouter_AdminCanToggleAWDControlsAndSeeOrchestrationState(t *testing.T) {
	env := newFullRouterTestEnv(t)
	now := time.Now().UTC()

	awdTeam := &contestcontracts.Team{
		ContestID:  env.awdContest.ID,
		Name:       "AWD Control Team",
		CaptainID:  env.student.ID,
		InviteCode: "AWDCTRL1",
		MaxMembers: 4,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := env.db.Create(awdTeam).Error; err != nil {
		t.Fatalf("create awd team: %v", err)
	}
	if err := env.db.Create(&contestcontracts.TeamMember{
		ContestID: env.awdContest.ID,
		TeamID:    awdTeam.ID,
		UserID:    env.student.ID,
		JoinedAt:  now,
		CreatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create awd team member: %v", err)
	}

	serviceSnapshot, err := contestcontracts.EncodeContestAWDServiceSnapshot(contestcontracts.ContestAWDServiceSnapshot{
		Name:       "AWD Web",
		Category:   taxonomy.DimensionWeb,
		Difficulty: taxonomy.DifficultyEasy,
		RuntimeConfig: map[string]any{
			"image_id":         env.image.ID,
			"instance_sharing": challengecontracts.InstanceSharingPerTeam,
		},
		FlagConfig: map[string]any{
			"flag_type":   challengecontracts.FlagTypeStatic,
			"flag_prefix": "flag",
		},
	})
	if err != nil {
		t.Fatalf("encode awd service snapshot: %v", err)
	}
	awdService := &contestcontracts.ContestAWDService{
		ContestID:       env.awdContest.ID,
		AWDChallengeID:  env.challenge.ID,
		DisplayName:     "AWD Web",
		IsVisible:       true,
		ServiceSnapshot: serviceSnapshot,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := env.db.Create(awdService).Error; err != nil {
		t.Fatalf("create awd service: %v", err)
	}

	adminHeaders := sessionHeaders(loginForSession(t, env.router, env.admin.Username, env.adminPwd))

	for _, tc := range []struct {
		name    string
		path    string
		payload map[string]any
		assert  func(*testing.T, *practicecommands.AdminAWDScopeControlResp)
	}{
		{
			name: "retire team",
			path: fmt.Sprintf("/api/v1/admin/contests/%d/awd/teams/%d/retirement", env.awdContest.ID, awdTeam.ID),
			payload: map[string]any{
				"retired": true,
				"reason":  "retired-by-admin",
			},
			assert: func(t *testing.T, resp *practicecommands.AdminAWDScopeControlResp) {
				t.Helper()
				if !resp.Enabled || resp.ControlType != runtimecontracts.AWDScopeControlTypeRetired || resp.TeamID != awdTeam.ID {
					t.Fatalf("unexpected retirement response: %+v", resp)
				}
			},
		},
		{
			name: "disable service",
			path: fmt.Sprintf("/api/v1/admin/contests/%d/awd/teams/%d/services/%d/disabled", env.awdContest.ID, awdTeam.ID, awdService.ID),
			payload: map[string]any{
				"disabled": true,
				"reason":   "disabled-by-admin",
			},
			assert: func(t *testing.T, resp *practicecommands.AdminAWDScopeControlResp) {
				t.Helper()
				if !resp.Enabled || resp.ControlType != runtimecontracts.AWDScopeControlTypeServiceDisabled || resp.ServiceID == nil || *resp.ServiceID != awdService.ID {
					t.Fatalf("unexpected disable response: %+v", resp)
				}
			},
		},
		{
			name: "suppress desired reconcile",
			path: fmt.Sprintf("/api/v1/admin/contests/%d/awd/teams/%d/services/%d/suppression", env.awdContest.ID, awdTeam.ID, awdService.ID),
			payload: map[string]any{
				"suppressed": true,
				"reason":     "manual-suppress",
			},
			assert: func(t *testing.T, resp *practicecommands.AdminAWDScopeControlResp) {
				t.Helper()
				if !resp.Enabled || resp.ControlType != runtimecontracts.AWDScopeControlTypeDesiredReconcileSuppressed || resp.ServiceID == nil || *resp.ServiceID != awdService.ID {
					t.Fatalf("unexpected suppress response: %+v", resp)
				}
			},
		},
	} {
		resp := performFullRouterRequest(t, env.router, http.MethodPut, tc.path, tc.payload, adminHeaders)
		assertFullRouterStatus(t, resp, http.StatusOK)

		var result practicecommands.AdminAWDScopeControlResp
		decodeFullRouterData(t, resp, &result)
		tc.assert(t, &result)
	}

	resp := performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/admin/contests/%d/awd/instances", env.awdContest.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var orchestration practicecommands.AdminAWDInstanceOrchestrationResp
	decodeFullRouterData(t, resp, &orchestration)
	if len(orchestration.Controls) < 3 {
		t.Fatalf("expected 3 awd controls in orchestration view, got %+v", orchestration.Controls)
	}

	seen := make(map[string]bool, len(orchestration.Controls))
	for _, control := range orchestration.Controls {
		if control == nil {
			continue
		}
		seen[control.ControlType] = true
	}
	for _, controlType := range []string{
		runtimecontracts.AWDScopeControlTypeRetired,
		runtimecontracts.AWDScopeControlTypeServiceDisabled,
		runtimecontracts.AWDScopeControlTypeDesiredReconcileSuppressed,
	} {
		if !seen[controlType] {
			t.Fatalf("expected orchestration to include control %q, got %+v", controlType, orchestration.Controls)
		}
	}
}

func TestFullRouter_AdminChallengePublishRequestLifecycle(t *testing.T) {
	env := newFullRouterTestEnv(t)
	if err := env.db.Model(&appChallengeRow{}).
		Where("id = ?", env.challenge.ID).
		Update("status", challengecontracts.ChallengeStatusDraft).Error; err != nil {
		t.Fatalf("set challenge draft: %v", err)
	}
	env.challenge.Status = challengecontracts.ChallengeStatusDraft

	teacherHeaders := sessionHeaders(loginForSession(t, env.router, env.teacher.Username, env.teacherPwd))
	createResp := performFullRouterRequest(
		t,
		env.router,
		http.MethodPost,
		fmt.Sprintf("/api/v1/authoring/challenges/%d/publish-requests", env.challenge.ID),
		nil,
		teacherHeaders,
	)
	assertFullRouterStatus(t, createResp, http.StatusAccepted)

	var created map[string]any
	decodeFullRouterData(t, createResp, &created)
	if created["challenge_id"] != float64(env.challenge.ID) {
		t.Fatalf("unexpected created publish request payload: %+v", created)
	}
	if created["status"] != "queued" {
		t.Fatalf("expected queued publish request, got %+v", created)
	}
	if created["active"] != true {
		t.Fatalf("expected active publish request, got %+v", created)
	}

	latestResp := performFullRouterRequest(
		t,
		env.router,
		http.MethodGet,
		fmt.Sprintf("/api/v1/authoring/challenges/%d/publish-requests/latest", env.challenge.ID),
		nil,
		teacherHeaders,
	)
	assertFullRouterStatus(t, latestResp, http.StatusOK)

	var latest map[string]any
	decodeFullRouterData(t, latestResp, &latest)
	if latest["id"] != created["id"] {
		t.Fatalf("expected latest publish request id %v, got %+v", created["id"], latest)
	}
	if latest["status"] != "queued" {
		t.Fatalf("expected latest queued publish request, got %+v", latest)
	}
}
