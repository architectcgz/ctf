package fullrouteradmin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	challengehttp "ctf-platform/internal/module/challenge/api/http"
	practicecommands "ctf-platform/internal/module/practice/application/commands"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
)

type RequestFunc func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder

type AWDControlTarget struct {
	ContestID int64
	TeamID    int64
	ServiceID int64
}

type PublishRequestLifecycleTarget struct {
	ChallengeID int64
}

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func VerifyAdminCanToggleAWDControlsAndSeeOrchestrationState(
	t *testing.T,
	request RequestFunc,
	adminHeaders map[string]string,
	target AWDControlTarget,
) {
	t.Helper()

	for _, tc := range []struct {
		name    string
		path    string
		payload map[string]any
		assert  func(*testing.T, *practicecommands.AdminAWDScopeControlResp)
	}{
		{
			name: "retire team",
			path: fmt.Sprintf("/api/v1/admin/contests/%d/awd/teams/%d/retirement", target.ContestID, target.TeamID),
			payload: map[string]any{
				"retired": true,
				"reason":  "retired-by-admin",
			},
			assert: func(t *testing.T, resp *practicecommands.AdminAWDScopeControlResp) {
				t.Helper()
				if !resp.Enabled || resp.ControlType != runtimecontracts.AWDScopeControlTypeRetired || resp.TeamID != target.TeamID {
					t.Fatalf("unexpected retirement response: %+v", resp)
				}
			},
		},
		{
			name: "disable service",
			path: fmt.Sprintf("/api/v1/admin/contests/%d/awd/teams/%d/services/%d/disabled", target.ContestID, target.TeamID, target.ServiceID),
			payload: map[string]any{
				"disabled": true,
				"reason":   "disabled-by-admin",
			},
			assert: func(t *testing.T, resp *practicecommands.AdminAWDScopeControlResp) {
				t.Helper()
				if !resp.Enabled || resp.ControlType != runtimecontracts.AWDScopeControlTypeServiceDisabled || resp.ServiceID == nil || *resp.ServiceID != target.ServiceID {
					t.Fatalf("unexpected disable response: %+v", resp)
				}
			},
		},
		{
			name: "suppress desired reconcile",
			path: fmt.Sprintf("/api/v1/admin/contests/%d/awd/teams/%d/services/%d/suppression", target.ContestID, target.TeamID, target.ServiceID),
			payload: map[string]any{
				"suppressed": true,
				"reason":     "manual-suppress",
			},
			assert: func(t *testing.T, resp *practicecommands.AdminAWDScopeControlResp) {
				t.Helper()
				if !resp.Enabled || resp.ControlType != runtimecontracts.AWDScopeControlTypeDesiredReconcileSuppressed || resp.ServiceID == nil || *resp.ServiceID != target.ServiceID {
					t.Fatalf("unexpected suppress response: %+v", resp)
				}
			},
		},
	} {
		resp := request(http.MethodPut, tc.path, tc.payload, adminHeaders)
		assertStatus(t, resp, http.StatusOK)

		var result practicecommands.AdminAWDScopeControlResp
		decodeEnvelopeData(t, resp, &result)
		tc.assert(t, &result)
	}

	resp := request(http.MethodGet, fmt.Sprintf("/api/v1/admin/contests/%d/awd/instances", target.ContestID), nil, adminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var orchestration practicecommands.AdminAWDInstanceOrchestrationResp
	decodeEnvelopeData(t, resp, &orchestration)
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

func VerifyAdminChallengePublishRequestLifecycle(
	t *testing.T,
	request RequestFunc,
	teacherHeaders map[string]string,
	target PublishRequestLifecycleTarget,
) {
	t.Helper()

	createResp := request(
		http.MethodPost,
		fmt.Sprintf("/api/v1/authoring/challenges/%d/publish-requests", target.ChallengeID),
		nil,
		teacherHeaders,
	)
	assertStatus(t, createResp, http.StatusAccepted)

	var created challengehttp.ChallengePublishCheckJobResp
	decodeEnvelopeData(t, createResp, &created)
	if created.ChallengeID != target.ChallengeID {
		t.Fatalf("unexpected created publish request payload: %+v", created)
	}
	if created.Status != "queued" {
		t.Fatalf("expected queued publish request, got %+v", created)
	}
	if !created.Active {
		t.Fatalf("expected active publish request, got %+v", created)
	}

	latestResp := request(
		http.MethodGet,
		fmt.Sprintf("/api/v1/authoring/challenges/%d/publish-requests/latest", target.ChallengeID),
		nil,
		teacherHeaders,
	)
	assertStatus(t, latestResp, http.StatusOK)

	var latest challengehttp.ChallengePublishCheckJobResp
	decodeEnvelopeData(t, latestResp, &latest)
	if latest.ChallengeID != target.ChallengeID {
		t.Fatalf("expected latest publish request challenge id %d, got %+v", target.ChallengeID, latest)
	}
	if latest.ID != created.ID {
		t.Fatalf("expected latest publish request id %d, got %+v", created.ID, latest)
	}
	if latest.Status != "queued" {
		t.Fatalf("expected latest queued publish request, got %+v", latest)
	}
}

func assertStatus(t *testing.T, resp *httptest.ResponseRecorder, want int) {
	t.Helper()

	if resp.Code != want {
		t.Fatalf("expected status %d, got %d body=%s", want, resp.Code, resp.Body.String())
	}
}

func decodeEnvelopeData(t *testing.T, resp *httptest.ResponseRecorder, target any) {
	t.Helper()

	var body envelope
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response envelope: %v body=%s", err, resp.Body.String())
	}
	if len(body.Data) == 0 || string(body.Data) == "null" {
		t.Fatalf("expected response data, got empty body=%s", resp.Body.String())
	}
	if err := json.Unmarshal(body.Data, target); err != nil {
		t.Fatalf("decode response data: %v body=%s", err, resp.Body.String())
	}
}
