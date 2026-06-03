package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"ctf-platform/internal/apperror"
	challengehttp "ctf-platform/internal/module/challenge/api/http"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contesthttp "ctf-platform/internal/module/contest/api/http"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	runtimehttp "ctf-platform/internal/module/runtime/api/http"
	"ctf-platform/internal/shared/taxonomy"
)

func TestFullRouter_InstanceHintAndProxyStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd))
	peerHeaders := bearerHeaders(loginForToken(t, env.router, env.peerStudent.Username, "Password123"))
	teacherHeaders := bearerHeaders(loginForToken(t, env.router, env.teacher.Username, env.teacherPwd))

	resp := performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d", env.challenge.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var detail challengehttp.ChallengeDetailResp
	decodeFullRouterData(t, resp, &detail)
	if len(detail.Hints) == 0 || detail.Hints[0].Content == "" {
		t.Fatalf("expected hint content in challenge detail, got %+v", detail.Hints)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/challenges/%d/hints/99/unlock", env.challenge.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusNotFound)

	draftChallenge := createDraftChallengeRecord(t, env, "Draft Hint Challenge")
	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/challenges/%d/hints/1/unlock", draftChallenge.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusNotFound)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/instances/%d/extend", env.instance.ID), nil, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/instances/%d/extend", env.instance.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
	var extended instancecontracts.InstanceResp
	decodeFullRouterData(t, resp, &extended)
	if extended.ID != env.instance.ID {
		t.Fatalf("unexpected extended instance id: %+v", extended)
	}
	if extended.RemainingExtends != 1 {
		t.Fatalf("expected remaining extends 1 after first extend, got %+v", extended)
	}
	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/instances/%d/extend", env.instance.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/instances/%d/extend", env.instance.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resetInstanceForAccessMatrix(t, env, env.instance.ID)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/teacher/instances?class_name=ClassB", nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/teacher/instances", nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var teacherInstances runtimehttp.TeacherInstancePageResp
	decodeFullRouterData(t, resp, &teacherInstances)
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

	if err := env.db.Model(&instancecontracts.Instance{}).Where("id = ?", env.instance.ID).Updates(map[string]any{
		"access_url": proxyTarget.URL,
		"status":     instancecontracts.InstanceStatusRunning,
	}).Error; err != nil {
		t.Fatalf("update instance access url: %v", err)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/instances/%d/access", env.instance.ID), nil, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/instances/%d/proxy/ping", env.instance.ID), nil, nil)
	assertFullRouterStatus(t, resp, http.StatusUnauthorized)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/instances/%d/access", env.instance.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var access runtimehttp.InstanceAccessResp
	decodeFullRouterData(t, resp, &access)
	parsedAccessURL, err := url.Parse(access.AccessURL)
	if err != nil {
		t.Fatalf("parse access url: %v", err)
	}
	ticket := parsedAccessURL.Query().Get("ticket")
	if ticket == "" {
		t.Fatalf("expected proxy ticket in access url: %s", access.AccessURL)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/instances/%d/proxy/ping?ticket=%s", env.instance.ID, url.QueryEscape(ticket)), nil, nil)
	if resp.Code != http.StatusFound {
		t.Fatalf("expected proxy redirect, got %d body=%s", resp.Code, resp.Body.String())
	}
	if location := resp.Header().Get("Location"); location != fmt.Sprintf("/api/v1/instances/%d/proxy/ping", env.instance.ID) {
		t.Fatalf("unexpected proxy redirect location: %s", location)
	}
	cookies := resp.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected proxy access cookie to be set")
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/instances/%d/proxy/ping", env.instance.ID), nil, map[string]string{
		"Cookie": cookies[0].String(),
	})
	assertFullRouterStatus(t, resp, http.StatusOK)
	if body := resp.Body.String(); body != "proxied:/ping" {
		t.Fatalf("unexpected proxy body: %s", body)
	}

	if err := env.db.Model(&instancecontracts.Instance{}).Where("id = ?", env.instance.ID).Updates(map[string]any{
		"status": instancecontracts.InstanceStatusStopped,
	}).Error; err != nil {
		t.Fatalf("stop instance: %v", err)
	}
	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/instances/%d/access", env.instance.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusGone)
}

func TestFullRouter_AWDTrafficAdminStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	adminHeaders := bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd))
	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd))

	awdTeam := createContestTeam(t, env, env.awdContest.ID, env.student.ID, "AWD-Traffic-Blue", 4)
	trafficServiceID := contesttestsupport.DefaultAWDContestServiceID(env.awdContest.ID, env.challenge.ID)
	contesttestsupport.SyncAWDContestServiceFixture(
		t,
		env.db,
		env.awdContest.ID,
		env.challenge.ID,
		"traffic-service",
		contestcontracts.AWDCheckerTypeHTTPStandard,
		`{"method":"GET","path":"/ping"}`,
		100,
		60,
		40,
		time.Now(),
	)

	proxyTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("awd-proxied:" + r.URL.Path))
	}))
	defer proxyTarget.Close()

	if err := env.db.Model(&instancecontracts.Instance{}).Where("id = ?", env.instance.ID).Updates(map[string]any{
		"user_id":      env.student.ID,
		"contest_id":   env.awdContest.ID,
		"team_id":      awdTeam.ID,
		"service_id":   trafficServiceID,
		"challenge_id": env.challenge.ID,
		"access_url":   proxyTarget.URL,
		"status":       instancecontracts.InstanceStatusRunning,
	}).Error; err != nil {
		t.Fatalf("update awd instance scope: %v", err)
	}

	resp := performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/instances/%d/access", env.instance.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var access runtimehttp.InstanceAccessResp
	decodeFullRouterData(t, resp, &access)
	parsedAccessURL, err := url.Parse(access.AccessURL)
	if err != nil {
		t.Fatalf("parse access url: %v", err)
	}
	ticket := parsedAccessURL.Query().Get("ticket")
	if ticket == "" {
		t.Fatalf("expected proxy ticket in access url: %s", access.AccessURL)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/instances/%d/proxy/ping?ticket=%s", env.instance.ID, url.QueryEscape(ticket)), nil, nil)
	if resp.Code != http.StatusFound {
		t.Fatalf("expected proxy redirect, got %d body=%s", resp.Code, resp.Body.String())
	}
	cookies := resp.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected proxy access cookie to be set")
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/instances/%d/proxy/ping", env.instance.ID), nil, map[string]string{
		"Cookie": cookies[0].String(),
	})
	assertFullRouterStatus(t, resp, http.StatusOK)
	if body := resp.Body.String(); body != "awd-proxied:/ping" {
		t.Fatalf("unexpected proxy body: %s", body)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/admin/contests/%d/awd/rounds/%d/traffic/summary", env.awdContest.ID, env.awdRound.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var summary contesthttp.AWDTrafficSummaryResp
	decodeFullRouterData(t, resp, &summary)
	if summary.TotalRequests < 1 {
		t.Fatalf("expected traffic requests in summary, got %+v", summary)
	}
	if len(summary.TopPaths) == 0 || summary.TopPaths[0].Path != "/ping" {
		t.Fatalf("unexpected top paths summary: %+v", summary.TopPaths)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/admin/contests/%d/awd/rounds/%d/traffic/events?page=1&page_size=20", env.awdContest.ID, env.awdRound.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var page contesthttp.AWDTrafficEventPageResp
	decodeFullRouterData(t, resp, &page)
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

func TestFullRouter_VisibleAWDContestChallengesIncludeAWDServiceID(t *testing.T) {
	env := newFullRouterTestEnv(t)

	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd))

	awdChallenge := &challengecontracts.AWDChallenge{
		Name:           "Visible AWD Challenge",
		Category:       taxonomy.DimensionWeb,
		Difficulty:     taxonomy.DifficultyMedium,
		ServiceType:    challengecontracts.AWDServiceTypeWebHTTP,
		DeploymentMode: challengecontracts.AWDDeploymentModeSingleContainer,
		Status:         challengecontracts.AWDChallengeStatusPublished,
	}
	if err := env.db.Create(awdChallenge).Error; err != nil {
		t.Fatalf("create awd challenge: %v", err)
	}

	contest := createFullRouterContest(t, env, "Visible AWD Contest", contestcontracts.ContestStatusRunning)
	contest.Mode = contestcontracts.ContestModeAWD
	if err := env.db.Save(contest).Error; err != nil {
		t.Fatalf("update contest mode: %v", err)
	}

	now := time.Now()
	awdService := &contestcontracts.ContestAWDService{
		ContestID:      contest.ID,
		AWDChallengeID: awdChallenge.ID,
		DisplayName:    "Visible Bank Portal",
		Order:          0,
		IsVisible:      true,
		ScoreConfig:    `{"points":260}`,
		RuntimeConfig:  fmt.Sprintf(`{"awd_challenge_id":%d}`, awdChallenge.ID),
		ServiceSnapshot: `{
			"name":"Visible Bank Portal",
			"category":"web",
			"difficulty":"medium"
		}`,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := env.db.Create(awdService).Error; err != nil {
		t.Fatalf("create contest awd service: %v", err)
	}

	createContestRegistration(t, env, contest.ID, env.student.ID, contestcontracts.ContestRegistrationStatusApproved, nil)

	resp := performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/challenges", contest.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var visibleChallenges []contesthttp.ContestChallengeInfo
	decodeFullRouterData(t, resp, &visibleChallenges)
	if len(visibleChallenges) != 1 || visibleChallenges[0].AWDServiceID == nil || *visibleChallenges[0].AWDServiceID != awdService.ID {
		t.Fatalf("unexpected visible awd service metadata: %+v", visibleChallenges)
	}
}

func TestFullRouter_AWDContestLegacyChallengeInstanceRouteRejected(t *testing.T) {
	env := newFullRouterTestEnv(t)

	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd))
	awdTeam := createContestTeam(t, env, env.awdContest.ID, env.student.ID, "AWD Legacy Route Team", 4)
	contesttestsupport.SyncAWDContestServiceFixture(
		t,
		env.db,
		env.awdContest.ID,
		env.challenge.ID,
		"Matrix AWD Service",
		contestcontracts.AWDCheckerTypeHTTPStandard,
		`{"get_flag":{"path":"/health"}}`,
		100,
		18,
		28,
		time.Now(),
	)

	resp := performFullRouterRequest(
		t,
		env.router,
		http.MethodPost,
		fmt.Sprintf("/api/v1/contests/%d/challenges/%d/instances", env.awdContest.ID, env.challenge.ID),
		nil,
		studentHeaders,
	)
	assertFullRouterStatus(t, resp, http.StatusBadRequest)

	var envelope fullRouterEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode challenge-based awd instance route response: %v body=%s", err, resp.Body.String())
	}
	if envelope.Code != apperror.ErrInvalidParams.Code || envelope.Message != apperror.ErrInvalidParams.Message {
		t.Fatalf("expected invalid params envelope, got %+v", envelope)
	}

	var awdInstanceCount int64
	if err := env.db.Model(&instancecontracts.Instance{}).
		Where("contest_id = ? AND team_id = ? AND user_id = ?", env.awdContest.ID, awdTeam.ID, env.student.ID).
		Count(&awdInstanceCount).Error; err != nil {
		t.Fatalf("count awd instances after challenge-based route request: %v", err)
	}
	if awdInstanceCount != 0 {
		t.Fatalf("expected no awd instance created through challenge-based route, got %d", awdInstanceCount)
	}
}

func TestFullRouter_AWDChallengeAuthoringStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	adminHeaders := bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd))
	teacherHeaders := bearerHeaders(loginForToken(t, env.router, env.teacher.Username, env.teacherPwd))
	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd))

	resp := performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/authoring/awd-challenges", map[string]any{
		"name":            "Bank Portal AWD",
		"slug":            "bank-portal-awd",
		"category":        "web",
		"difficulty":      taxonomy.DifficultyHard,
		"description":     "multi-step banking target",
		"service_type":    "web_http",
		"deployment_mode": "single_container",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	invalidCategoryResp := performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/authoring/awd-challenges", map[string]any{
		"name":            "Invalid Category AWD",
		"slug":            "invalid-category-awd",
		"category":        "mobile",
		"difficulty":      taxonomy.DifficultyHard,
		"description":     "invalid category",
		"service_type":    "web_http",
		"deployment_mode": "single_container",
	}, adminHeaders)
	assertFullRouterStatus(t, invalidCategoryResp, http.StatusBadRequest)

	var createdChallenge challengehttp.AWDChallengeResp
	decodeFullRouterData(t, resp, &createdChallenge)
	if createdChallenge.ID == 0 || createdChallenge.Slug != "bank-portal-awd" {
		t.Fatalf("unexpected created awd challenge: %+v", createdChallenge)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/authoring/awd-challenges?page=1&page_size=10", nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var page challengehttp.AWDChallengePageResp
	decodeFullRouterData(t, resp, &page)
	if page.Total < 1 || len(page.Items) == 0 {
		t.Fatalf("expected awd challenge page items, got %+v", page)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/authoring/awd-challenges/%d", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/authoring/awd-challenges/%d", createdChallenge.ID), map[string]any{
		"name":   "Bank Portal AWD v2",
		"status": "published",
	}, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var updatedChallenge challengehttp.AWDChallengeResp
	decodeFullRouterData(t, resp, &updatedChallenge)
	if updatedChallenge.Name != "Bank Portal AWD v2" || updatedChallenge.Status != "published" {
		t.Fatalf("unexpected updated awd challenge: %+v", updatedChallenge)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/authoring/awd-challenges", nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/authoring/awd-challenges/%d", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/authoring/awd-challenges/%d", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusNotFound)
}
