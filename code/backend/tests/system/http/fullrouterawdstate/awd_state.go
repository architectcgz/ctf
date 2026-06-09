package fullrouterawdstate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"ctf-platform/internal/apperror"
	challengehttp "ctf-platform/internal/module/challenge/api/http"
	contesthttp "ctf-platform/internal/module/contest/api/http"
	instancehttp "ctf-platform/internal/module/instance/api/http"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	"ctf-platform/internal/shared/taxonomy"
)

type RequestFunc func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder

type InstanceHintAndProxyStateMatrixDriver struct {
	Request              RequestFunc
	StudentHeaders       map[string]string
	PeerHeaders          map[string]string
	TeacherHeaders       map[string]string
	ChallengeID          int64
	InstanceID           int64
	CreateDraftChallenge func(t *testing.T) int64
	ResetInstance        func(t *testing.T)
	SetInstanceAccess    func(t *testing.T, accessURL string, status string)
}

type AWDTrafficAdminStateMatrixDriver struct {
	Request           RequestFunc
	AdminHeaders      map[string]string
	StudentHeaders    map[string]string
	ContestID         int64
	RoundID           int64
	InstanceID        int64
	SetInstanceAccess func(t *testing.T, accessURL string)
}

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func VerifyInstanceHintAndProxyStateMatrix(t *testing.T, driver InstanceHintAndProxyStateMatrixDriver) {
	t.Helper()

	resp := driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d", driver.ChallengeID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var detail challengehttp.ChallengeDetailResp
	decodeEnvelopeData(t, resp, &detail)
	if len(detail.Hints) == 0 || detail.Hints[0].Content == "" {
		t.Fatalf("expected hint content in challenge detail, got %+v", detail.Hints)
	}

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/challenges/%d/hints/99/unlock", driver.ChallengeID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusNotFound)

	draftChallengeID := driver.CreateDraftChallenge(t)
	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/challenges/%d/hints/1/unlock", draftChallengeID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusNotFound)

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/instances/%d/extend", driver.InstanceID), nil, driver.PeerHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/instances/%d/extend", driver.InstanceID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)
	var extended instancecontracts.InstanceResp
	decodeEnvelopeData(t, resp, &extended)
	if extended.ID != driver.InstanceID {
		t.Fatalf("unexpected extended instance id: %+v", extended)
	}
	if extended.RemainingExtends != 1 {
		t.Fatalf("expected remaining extends 1 after first extend, got %+v", extended)
	}
	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/instances/%d/extend", driver.InstanceID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)
	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/instances/%d/extend", driver.InstanceID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	driver.ResetInstance(t)

	resp = driver.Request(http.MethodGet, "/api/v1/teacher/instances?class_name=ClassB", nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	resp = driver.Request(http.MethodGet, "/api/v1/teacher/instances", nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	var teacherInstances instancehttp.TeacherInstancePageResp
	decodeEnvelopeData(t, resp, &teacherInstances)
	if len(teacherInstances.List) == 0 {
		t.Fatalf("expected teacher instances for own class")
	}
	if !strings.Contains(resp.Body.String(), `"student_username"`) || !strings.Contains(resp.Body.String(), `"access_url"`) || !strings.Contains(resp.Body.String(), `"remaining_time"`) {
		t.Fatalf("expected teacher instance response to preserve key json fields, got %s", resp.Body.String())
	}
	if teacherInstances.List[0].StudentUsername == "" || teacherInstances.List[0].RemainingTime <= 0 {
		t.Fatalf("expected decoded teacher instance fields, got %+v", teacherInstances.List[0])
	}
	if teacherInstances.Summary.TotalCount == 0 {
		t.Fatalf("expected teacher instance summary fields, got %+v", teacherInstances.Summary)
	}

	proxyTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("proxied:" + r.URL.Path))
	}))
	defer proxyTarget.Close()

	driver.SetInstanceAccess(t, proxyTarget.URL, instancecontracts.InstanceStatusRunning)

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/instances/%d/access", driver.InstanceID), nil, driver.PeerHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/instances/%d/proxy/ping", driver.InstanceID), nil, nil)
	assertStatus(t, resp, http.StatusUnauthorized)

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/instances/%d/access", driver.InstanceID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var access instancehttp.InstanceAccessResp
	decodeEnvelopeData(t, resp, &access)
	parsedAccessURL, err := url.Parse(access.AccessURL)
	if err != nil {
		t.Fatalf("parse access url: %v", err)
	}
	ticket := parsedAccessURL.Query().Get("ticket")
	if ticket == "" {
		t.Fatalf("expected proxy ticket in access url: %s", access.AccessURL)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/instances/%d/proxy/ping?ticket=%s", driver.InstanceID, url.QueryEscape(ticket)), nil, nil)
	if resp.Code != http.StatusFound {
		t.Fatalf("expected proxy redirect, got %d body=%s", resp.Code, resp.Body.String())
	}
	if location := resp.Header().Get("Location"); location != fmt.Sprintf("/api/v1/instances/%d/proxy/ping", driver.InstanceID) {
		t.Fatalf("unexpected proxy redirect location: %s", location)
	}
	cookies := resp.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected proxy access cookie to be set")
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/instances/%d/proxy/ping", driver.InstanceID), nil, map[string]string{
		"Cookie": cookies[0].String(),
	})
	assertStatus(t, resp, http.StatusOK)
	if body := resp.Body.String(); body != "proxied:/ping" {
		t.Fatalf("unexpected proxy body: %s", body)
	}

	driver.SetInstanceAccess(t, proxyTarget.URL, instancecontracts.InstanceStatusStopped)
	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/instances/%d/access", driver.InstanceID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusGone)
}

func VerifyAWDTrafficAdminStateMatrix(t *testing.T, driver AWDTrafficAdminStateMatrixDriver) {
	t.Helper()

	proxyTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("awd-proxied:" + r.URL.Path))
	}))
	defer proxyTarget.Close()

	driver.SetInstanceAccess(t, proxyTarget.URL)

	resp := driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/instances/%d/access", driver.InstanceID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var access instancehttp.InstanceAccessResp
	decodeEnvelopeData(t, resp, &access)
	parsedAccessURL, err := url.Parse(access.AccessURL)
	if err != nil {
		t.Fatalf("parse access url: %v", err)
	}
	ticket := parsedAccessURL.Query().Get("ticket")
	if ticket == "" {
		t.Fatalf("expected proxy ticket in access url: %s", access.AccessURL)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/instances/%d/proxy/ping?ticket=%s", driver.InstanceID, url.QueryEscape(ticket)), nil, nil)
	if resp.Code != http.StatusFound {
		t.Fatalf("expected proxy redirect, got %d body=%s", resp.Code, resp.Body.String())
	}
	cookies := resp.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected proxy access cookie to be set")
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/instances/%d/proxy/ping", driver.InstanceID), nil, map[string]string{
		"Cookie": cookies[0].String(),
	})
	assertStatus(t, resp, http.StatusOK)
	if body := resp.Body.String(); body != "awd-proxied:/ping" {
		t.Fatalf("unexpected proxy body: %s", body)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/admin/contests/%d/awd/rounds/%d/traffic/summary", driver.ContestID, driver.RoundID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var summary contesthttp.AWDTrafficSummaryResp
	decodeEnvelopeData(t, resp, &summary)
	if summary.TotalRequests < 1 {
		t.Fatalf("expected traffic requests in summary, got %+v", summary)
	}
	if len(summary.TopPaths) == 0 || summary.TopPaths[0].Path != "/ping" {
		t.Fatalf("unexpected top paths summary: %+v", summary.TopPaths)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/admin/contests/%d/awd/rounds/%d/traffic/events?page=1&page_size=20", driver.ContestID, driver.RoundID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var page contesthttp.AWDTrafficEventPageResp
	decodeEnvelopeData(t, resp, &page)
	if page.Total < 1 || len(page.List) == 0 {
		t.Fatalf("expected traffic events page data, got %+v", page)
	}
	if page.List[0].Method != http.MethodGet || page.List[0].Path != "/ping" {
		t.Fatalf("unexpected traffic event item: %+v", page.List[0])
	}
	if page.List[0].ServiceID <= 0 {
		t.Fatalf("expected traffic event service_id, got %+v", page.List[0])
	}
}

func VerifyVisibleAWDContestChallengesIncludeAWDServiceID(
	t *testing.T,
	request RequestFunc,
	studentHeaders map[string]string,
	contestID int64,
	serviceID int64,
) {
	t.Helper()

	resp := request(http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/challenges", contestID), nil, studentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var visibleChallenges []contesthttp.ContestChallengeInfo
	decodeEnvelopeData(t, resp, &visibleChallenges)
	if len(visibleChallenges) != 1 || visibleChallenges[0].AWDServiceID == nil || *visibleChallenges[0].AWDServiceID != serviceID {
		t.Fatalf("unexpected visible awd service metadata: %+v", visibleChallenges)
	}
}

func VerifyAWDContestLegacyChallengeInstanceRouteRejected(
	t *testing.T,
	request RequestFunc,
	studentHeaders map[string]string,
	contestID int64,
	challengeID int64,
) {
	t.Helper()

	resp := request(
		http.MethodPost,
		fmt.Sprintf("/api/v1/contests/%d/challenges/%d/instances", contestID, challengeID),
		nil,
		studentHeaders,
	)
	assertStatus(t, resp, http.StatusBadRequest)

	var body envelope
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode challenge-based awd instance route response: %v body=%s", err, resp.Body.String())
	}
	if body.Code != apperror.ErrInvalidParams.Code || body.Message != apperror.ErrInvalidParams.Message {
		t.Fatalf("expected invalid params envelope, got %+v", body)
	}
}

func VerifyAWDChallengeAuthoringStateMatrix(
	t *testing.T,
	request RequestFunc,
	adminHeaders map[string]string,
	teacherHeaders map[string]string,
	studentHeaders map[string]string,
) {
	t.Helper()

	resp := request(http.MethodPost, "/api/v1/authoring/awd-challenges", map[string]any{
		"name":            "Bank Portal AWD",
		"slug":            "bank-portal-awd",
		"category":        "web",
		"difficulty":      taxonomy.DifficultyHard,
		"description":     "multi-step banking target",
		"service_type":    "web_http",
		"deployment_mode": "single_container",
	}, adminHeaders)
	assertStatus(t, resp, http.StatusOK)

	invalidCategoryResp := request(http.MethodPost, "/api/v1/authoring/awd-challenges", map[string]any{
		"name":            "Invalid Category AWD",
		"slug":            "invalid-category-awd",
		"category":        "mobile",
		"difficulty":      taxonomy.DifficultyHard,
		"description":     "invalid category",
		"service_type":    "web_http",
		"deployment_mode": "single_container",
	}, adminHeaders)
	assertStatus(t, invalidCategoryResp, http.StatusBadRequest)

	var createdChallenge challengehttp.AWDChallengeResp
	decodeEnvelopeData(t, resp, &createdChallenge)
	if createdChallenge.ID == 0 || createdChallenge.Slug != "bank-portal-awd" {
		t.Fatalf("unexpected created awd challenge: %+v", createdChallenge)
	}

	resp = request(http.MethodGet, "/api/v1/authoring/awd-challenges?page=1&page_size=10", nil, teacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	var page challengehttp.AWDChallengePageResp
	decodeEnvelopeData(t, resp, &page)
	if page.Total < 1 || len(page.Items) == 0 {
		t.Fatalf("expected awd challenge page items, got %+v", page)
	}

	resp = request(http.MethodGet, fmt.Sprintf("/api/v1/authoring/awd-challenges/%d", createdChallenge.ID), nil, adminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = request(http.MethodPut, fmt.Sprintf("/api/v1/authoring/awd-challenges/%d", createdChallenge.ID), map[string]any{
		"name":   "Bank Portal AWD v2",
		"status": "published",
	}, teacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	var updatedChallenge challengehttp.AWDChallengeResp
	decodeEnvelopeData(t, resp, &updatedChallenge)
	if updatedChallenge.Name != "Bank Portal AWD v2" || updatedChallenge.Status != "published" {
		t.Fatalf("unexpected updated awd challenge: %+v", updatedChallenge)
	}

	resp = request(http.MethodGet, "/api/v1/authoring/awd-challenges", nil, studentHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	resp = request(http.MethodDelete, fmt.Sprintf("/api/v1/authoring/awd-challenges/%d", createdChallenge.ID), nil, adminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = request(http.MethodGet, fmt.Sprintf("/api/v1/authoring/awd-challenges/%d", createdChallenge.ID), nil, adminHeaders)
	assertStatus(t, resp, http.StatusNotFound)
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
