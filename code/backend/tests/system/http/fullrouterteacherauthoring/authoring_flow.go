package fullrouterteacherauthoring

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	challengehttp "ctf-platform/internal/module/challenge/api/http"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
)

type RequestFunc func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder

type ChallengeOwnershipDriver struct {
	Request                       RequestFunc
	AdminHeaders                  map[string]string
	TeacherHeaders                map[string]string
	CreatePayload                 func(title string) map[string]any
	TemplateID                    int64
	ImageID                       int64
	PrepareArchivedAdminChallenge func(t *testing.T, adminChallengeID int64) int64
}

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func VerifyTeacherCanBrowseArchivedAndDraftChallengesButOnlyManageOwnChallenges(t *testing.T, driver ChallengeOwnershipDriver) {
	t.Helper()

	adminChallenge := createChallenge(t, driver.Request, driver.AdminHeaders, driver.CreatePayload("admin-owned"))
	teacherChallenge := createChallenge(t, driver.Request, driver.TeacherHeaders, driver.CreatePayload("teacher-owned"))

	listResp := driver.Request(http.MethodGet, "/api/v1/authoring/challenges?page=1&page_size=50", nil, driver.TeacherHeaders)
	assertStatus(t, listResp, http.StatusOK)

	var listResult struct {
		List []challengehttp.ChallengeResp `json:"list"`
	}
	decodeEnvelopeData(t, listResp, &listResult)

	foundTeacherOwned := false
	foundAdminOwned := false
	for _, item := range listResult.List {
		if item.ID == teacherChallenge.ID {
			foundTeacherOwned = true
		}
		if item.ID == adminChallenge.ID {
			foundAdminOwned = true
		}
	}
	if !foundTeacherOwned {
		t.Fatalf("teacher should see own challenge %d in list, got %+v", teacherChallenge.ID, listResult.List)
	}
	if !foundAdminOwned {
		t.Fatalf("teacher should see admin challenge %d in list, got %+v", adminChallenge.ID, listResult.List)
	}

	resp := driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/topology", adminChallenge.ID), map[string]any{
		"template_id": driver.TemplateID,
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup", adminChallenge.ID), map[string]any{
		"title":      "admin writeup",
		"content":    "admin writeup content",
		"visibility": challengeentity.WriteupVisibilityPublic,
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/flag", adminChallenge.ID), map[string]any{
		"flag_type":   challengecontracts.FlagTypeStatic,
		"flag":        "flag{ownership-check}",
		"flag_prefix": "flag",
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	packageRevisionID := driver.PrepareArchivedAdminChallenge(t, adminChallenge.ID)

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/authoring/challenges/%d/publish-requests", adminChallenge.ID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusAccepted)

	var publishJob challengehttp.ChallengePublishCheckJobResp
	decodeEnvelopeData(t, resp, &publishJob)
	if publishJob.ChallengeID != adminChallenge.ID {
		t.Fatalf("unexpected publish request payload: %+v", publishJob)
	}

	readChecks := []struct {
		name    string
		method  string
		path    string
		payload any
		assert  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "get archived detail",
			method: http.MethodGet,
			path:   fmt.Sprintf("/api/v1/authoring/challenges/%d", adminChallenge.ID),
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assertStatus(t, resp, http.StatusOK)
				var detail challengehttp.ChallengeResp
				decodeEnvelopeData(t, resp, &detail)
				if detail.ID != adminChallenge.ID || detail.Status != challengecontracts.ChallengeStatusArchived {
					t.Fatalf("unexpected archived challenge detail: %+v", detail)
				}
			},
		},
		{
			name:   "get admin writeup",
			method: http.MethodGet,
			path:   fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup", adminChallenge.ID),
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assertStatus(t, resp, http.StatusOK)
				var writeup challengecontracts.AdminChallengeWriteupResp
				decodeEnvelopeData(t, resp, &writeup)
				if writeup.Title != "admin writeup" || writeup.Visibility != challengeentity.WriteupVisibilityPublic {
					t.Fatalf("unexpected admin writeup: %+v", writeup)
				}
			},
		},
		{
			name:   "get admin flag",
			method: http.MethodGet,
			path:   fmt.Sprintf("/api/v1/authoring/challenges/%d/flag", adminChallenge.ID),
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assertStatus(t, resp, http.StatusOK)
				var flagResp challengehttp.FlagResp
				decodeEnvelopeData(t, resp, &flagResp)
				if !flagResp.Configured || flagResp.FlagType != challengecontracts.FlagTypeStatic {
					t.Fatalf("unexpected flag config: %+v", flagResp)
				}
			},
		},
		{
			name:   "get admin topology",
			method: http.MethodGet,
			path:   fmt.Sprintf("/api/v1/authoring/challenges/%d/topology", adminChallenge.ID),
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assertStatus(t, resp, http.StatusOK)
				var topology challengehttp.ChallengeTopologyResp
				decodeEnvelopeData(t, resp, &topology)
				if topology.TemplateID == nil || *topology.TemplateID != driver.TemplateID {
					t.Fatalf("unexpected topology template binding: %+v", topology)
				}
			},
		},
		{
			name:   "get latest publish request",
			method: http.MethodGet,
			path:   fmt.Sprintf("/api/v1/authoring/challenges/%d/publish-requests/latest", adminChallenge.ID),
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assertStatus(t, resp, http.StatusOK)
				var latest challengehttp.ChallengePublishCheckJobResp
				decodeEnvelopeData(t, resp, &latest)
				if latest.ChallengeID != adminChallenge.ID || latest.Status != "queued" {
					t.Fatalf("unexpected latest publish request: %+v", latest)
				}
			},
		},
		{
			name:   "download package export",
			method: http.MethodGet,
			path:   fmt.Sprintf("/api/v1/authoring/challenges/%d/package-export/download?revision_id=%d", adminChallenge.ID, packageRevisionID),
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assertStatus(t, resp, http.StatusOK)
			},
		},
	}
	for _, tc := range readChecks {
		resp := driver.Request(tc.method, tc.path, tc.payload, driver.TeacherHeaders)
		tc.assert(t, resp)
	}

	for _, tc := range []struct {
		name    string
		method  string
		path    string
		payload any
	}{
		{name: "update challenge", method: http.MethodPut, path: fmt.Sprintf("/api/v1/authoring/challenges/%d", adminChallenge.ID), payload: map[string]any{"title": "forbidden-update"}},
		{name: "configure flag", method: http.MethodPut, path: fmt.Sprintf("/api/v1/authoring/challenges/%d/flag", adminChallenge.ID), payload: map[string]any{
			"flag_type":   challengecontracts.FlagTypeStatic,
			"flag":        "flag{ownership-check}",
			"flag_prefix": "flag",
		}},
		{name: "upsert writeup", method: http.MethodPut, path: fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup", adminChallenge.ID), payload: map[string]any{
			"title":      "forbidden writeup",
			"content":    "forbidden content",
			"visibility": challengeentity.WriteupVisibilityPublic,
		}},
		{name: "save topology", method: http.MethodPut, path: fmt.Sprintf("/api/v1/authoring/challenges/%d/topology", adminChallenge.ID), payload: map[string]any{
			"entry_node_key": "web",
			"nodes": []map[string]any{
				{
					"key":          "web",
					"name":         "web",
					"image_id":     driver.ImageID,
					"service_port": 80,
					"inject_flag":  true,
					"tier":         challengecontracts.TopologyTierPublic,
				},
			},
		}},
		{name: "self check", method: http.MethodPost, path: fmt.Sprintf("/api/v1/authoring/challenges/%d/self-check", adminChallenge.ID)},
	} {
		resp := driver.Request(tc.method, tc.path, tc.payload, driver.TeacherHeaders)
		if resp.Code != http.StatusForbidden {
			t.Fatalf("%s should be forbidden, got status=%d body=%s", tc.name, resp.Code, resp.Body.String())
		}
	}

	ownDetailResp := driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/authoring/challenges/%d", teacherChallenge.ID), nil, driver.TeacherHeaders)
	assertStatus(t, ownDetailResp, http.StatusOK)

	var ownDetail challengehttp.ChallengeResp
	decodeEnvelopeData(t, ownDetailResp, &ownDetail)
	if ownDetail.Status != challengecontracts.ChallengeStatusDraft {
		t.Fatalf("expected own draft challenge to stay readable, got %+v", ownDetail)
	}
}

func VerifyCreateChallengeStoresCreatorResponse(
	t *testing.T,
	request RequestFunc,
	teacherHeaders map[string]string,
	payload map[string]any,
) int64 {
	t.Helper()

	resp := request(http.MethodPost, "/api/v1/authoring/challenges", payload, teacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	var challengeData map[string]any
	decodeEnvelopeData(t, resp, &challengeData)
	if _, ok := challengeData["created_by"]; !ok {
		t.Fatalf("expected challenge response to include created_by, got %+v", challengeData)
	}

	challengeIDFloat, ok := challengeData["id"].(float64)
	if !ok {
		t.Fatalf("expected numeric challenge id, got %+v", challengeData["id"])
	}

	return int64(challengeIDFloat)
}

func VerifyChallengeSelfCheckRunsPrecheckAndRuntime(
	t *testing.T,
	request RequestFunc,
	teacherHeaders map[string]string,
	challengeID int64,
) {
	t.Helper()

	resp := request(
		http.MethodPost,
		fmt.Sprintf("/api/v1/authoring/challenges/%d/self-check", challengeID),
		nil,
		teacherHeaders,
	)
	assertStatus(t, resp, http.StatusOK)

	var result challengehttp.ChallengeSelfCheckResp
	decodeEnvelopeData(t, resp, &result)
	if result.ChallengeID != challengeID {
		t.Fatalf("expected challenge_id=%d, got %d", challengeID, result.ChallengeID)
	}
	if !result.Precheck.Passed {
		t.Fatalf("expected precheck passed, got %+v", result.Precheck)
	}
	if !result.Runtime.Passed {
		t.Fatalf("expected runtime passed, got %+v", result.Runtime)
	}
	if result.Runtime.AccessURL == "" {
		t.Fatalf("expected runtime access url, got empty")
	}
	if len(result.Runtime.Steps) == 0 {
		t.Fatalf("expected runtime steps, got empty")
	}
}

func createChallenge(t *testing.T, request RequestFunc, headers map[string]string, payload map[string]any) challengehttp.ChallengeResp {
	t.Helper()

	resp := request(http.MethodPost, "/api/v1/authoring/challenges", payload, headers)
	assertStatus(t, resp, http.StatusOK)

	var challenge challengehttp.ChallengeResp
	decodeEnvelopeData(t, resp, &challenge)
	return challenge
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
