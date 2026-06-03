package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	"ctf-platform/internal/shared/taxonomy"
)

type practiceFlowScenarioResult struct {
	env               *flowTestEnv
	adminSession      *http.Cookie
	studentSession    *http.Cookie
	challenge         flowChallengeResponse
	listBeforeItems   []flowChallengeListItem
	detailBody        []byte
	detail            flowChallengeDetail
	instance          flowInstanceResponse
	proxyAccess       flowInstanceResponse
	proxyLocation     string
	proxyCookies      []*http.Cookie
	wrongSubmission   flowSubmissionResponse
	correctSubmission flowSubmissionResponse
	repeatSubmission  flowSubmissionResponse
	submissionHistory []flowSubmissionRecord
	listAfterItems    []flowChallengeListItem
	progress          flowProgressResponse
	timeline          flowTimelineResponse
	evidence          flowTeacherEvidenceReviewResponse
	attackSessions    flowTeacherAttackSessionResponse
	auditPage         flowPageResponse[flowAuditItem]
	submissions       []contestcontracts.Submission
}

func runPublishedPracticeFlowScenario(t *testing.T) *practiceFlowScenarioResult {
	t.Helper()

	env := newPracticeFlowTestEnv(t)
	result := &practiceFlowScenarioResult{env: env}

	adminSession := loginForSession(t, env.router, "admin_user", "Password123")
	studentSession := loginForSession(t, env.router, "student_user", "Password123")
	result.adminSession = adminSession
	result.studentSession = studentSession

	createResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/authoring/challenges",
		map[string]any{
			"title":          "Web SQLi 101",
			"description":    "basic sql injection challenge",
			"category":       taxonomy.DimensionWeb,
			"difficulty":     taxonomy.DifficultyEasy,
			"points":         100,
			"image_id":       env.image.ID,
			"attachment_url": "https://example.com/files/web-sqli-101.zip",
			"hints": []map[string]any{
				{
					"level":   1,
					"title":   "入口提示",
					"content": "先观察登录表单的参数。",
				},
			},
		},
		sessionHeaders(adminSession),
		nil,
	)
	if createResp.Code != http.StatusOK {
		t.Fatalf("unexpected create challenge status: %d body=%s", createResp.Code, createResp.Body.String())
	}
	createBody := decodeFlowEnvelope(t, createResp)
	challenge := decodeFlowJSON[flowChallengeResponse](t, createBody.Data)
	result.challenge = challenge

	configureFlagResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodPut,
		"/api/v1/authoring/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/flag",
		map[string]any{
			"flag_type": "static",
			"flag":      "flag{sqli_success}",
		},
		sessionHeaders(adminSession),
		nil,
	)
	if configureFlagResp.Code != http.StatusOK {
		t.Fatalf("unexpected configure flag status: %d body=%s", configureFlagResp.Code, configureFlagResp.Body.String())
	}

	if err := env.db.Model(&appChallengeRow{}).
		Where("id = ?", challenge.ID).
		Update("status", challengecontracts.ChallengeStatusPublished).Error; err != nil {
		t.Fatalf("set challenge published: %v", err)
	}

	listBeforeResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/challenges",
		nil,
		sessionHeaders(studentSession),
		nil,
	)
	if listBeforeResp.Code != http.StatusOK {
		t.Fatalf("unexpected list challenges status: %d body=%s", listBeforeResp.Code, listBeforeResp.Body.String())
	}
	listBeforeBody := decodeFlowEnvelope(t, listBeforeResp)
	listBefore := decodeFlowJSON[practicecontracts.PageResult[json.RawMessage]](t, listBeforeBody.Data)
	result.listBeforeItems = decodeFlowJSON[[]flowChallengeListItem](t, mustMarshalJSON(t, listBefore.List))

	detailResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10),
		nil,
		sessionHeaders(studentSession),
		nil,
	)
	if detailResp.Code != http.StatusOK {
		t.Fatalf("unexpected challenge detail status: %d body=%s", detailResp.Code, detailResp.Body.String())
	}
	detailBody := decodeFlowEnvelope(t, detailResp)
	result.detailBody = detailBody.Data
	result.detail = decodeFlowJSON[flowChallengeDetail](t, detailBody.Data)

	instanceCreateResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/instances",
		nil,
		sessionHeaders(studentSession),
		nil,
	)
	if instanceCreateResp.Code != http.StatusOK {
		t.Fatalf("unexpected create instance status: %d body=%s", instanceCreateResp.Code, instanceCreateResp.Body.String())
	}
	instanceCreateBody := decodeFlowEnvelope(t, instanceCreateResp)
	instance := decodeFlowJSON[flowInstanceResponse](t, instanceCreateBody.Data)
	result.instance = instance

	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/submit" {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("submitted"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("target ok"))
	}))
	t.Cleanup(targetServer.Close)
	if err := env.db.Model(&instancecontracts.Instance{}).
		Where("id = ?", instance.ID).
		Update("access_url", targetServer.URL).Error; err != nil {
		t.Fatalf("update instance access url: %v", err)
	}

	instanceAccessResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/instances/"+strconv.FormatInt(instance.ID, 10)+"/access",
		nil,
		sessionHeaders(studentSession),
		nil,
	)
	if instanceAccessResp.Code != http.StatusOK {
		t.Fatalf("unexpected instance access status: %d body=%s", instanceAccessResp.Code, instanceAccessResp.Body.String())
	}
	instanceAccessBody := decodeFlowEnvelope(t, instanceAccessResp)
	proxyAccess := decodeFlowJSON[flowInstanceResponse](t, instanceAccessBody.Data)
	result.proxyAccess = proxyAccess

	proxyBootstrapResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		proxyAccess.AccessURL,
		nil,
		nil,
		nil,
	)
	if proxyBootstrapResp.Code != http.StatusFound {
		t.Fatalf("expected proxy bootstrap redirect, got %d body=%s", proxyBootstrapResp.Code, proxyBootstrapResp.Body.String())
	}
	result.proxyLocation = proxyBootstrapResp.Header().Get("Location")
	result.proxyCookies = proxyBootstrapResp.Result().Cookies()

	proxyPageResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		result.proxyLocation,
		nil,
		nil,
		result.proxyCookies,
	)
	if proxyPageResp.Code != http.StatusOK || !strings.Contains(proxyPageResp.Body.String(), "target ok") {
		t.Fatalf("expected proxied page response, got %d body=%s", proxyPageResp.Code, proxyPageResp.Body.String())
	}

	proxySubmitResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/instances/"+strconv.FormatInt(instance.ID, 10)+"/proxy/submit",
		map[string]any{"payload": "' OR 1=1 --"},
		nil,
		result.proxyCookies,
	)
	if proxySubmitResp.Code != http.StatusCreated {
		t.Fatalf("expected proxied submit response, got %d body=%s", proxySubmitResp.Code, proxySubmitResp.Body.String())
	}

	wrongSubmitResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/submit",
		map[string]any{"flag": "flag{wrong_answer}"},
		sessionHeaders(studentSession),
		nil,
	)
	if wrongSubmitResp.Code != http.StatusOK {
		t.Fatalf("unexpected wrong submit status: %d body=%s", wrongSubmitResp.Code, wrongSubmitResp.Body.String())
	}
	result.wrongSubmission = decodeFlowJSON[flowSubmissionResponse](t, decodeFlowEnvelope(t, wrongSubmitResp).Data)

	correctSubmitResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/submit",
		map[string]any{"flag": "flag{sqli_success}"},
		sessionHeaders(studentSession),
		nil,
	)
	if correctSubmitResp.Code != http.StatusOK {
		t.Fatalf("unexpected correct submit status: %d body=%s", correctSubmitResp.Code, correctSubmitResp.Body.String())
	}
	result.correctSubmission = decodeFlowJSON[flowSubmissionResponse](t, decodeFlowEnvelope(t, correctSubmitResp).Data)

	submissionHistoryResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/submissions/mine",
		nil,
		sessionHeaders(studentSession),
		nil,
	)
	if submissionHistoryResp.Code != http.StatusOK {
		t.Fatalf("unexpected submission history status: %d body=%s", submissionHistoryResp.Code, submissionHistoryResp.Body.String())
	}
	result.submissionHistory = decodeFlowJSON[[]flowSubmissionRecord](t, decodeFlowEnvelope(t, submissionHistoryResp).Data)

	repeatSubmitResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/submit",
		map[string]any{"flag": "flag{sqli_success}"},
		sessionHeaders(studentSession),
		nil,
	)
	if repeatSubmitResp.Code != http.StatusOK {
		t.Fatalf("expected repeated correct submission to return 200, got %d body=%s", repeatSubmitResp.Code, repeatSubmitResp.Body.String())
	}
	result.repeatSubmission = decodeFlowJSON[flowSubmissionResponse](t, decodeFlowEnvelope(t, repeatSubmitResp).Data)

	listAfterResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/challenges",
		nil,
		sessionHeaders(studentSession),
		nil,
	)
	if listAfterResp.Code != http.StatusOK {
		t.Fatalf("unexpected post-submit list status: %d body=%s", listAfterResp.Code, listAfterResp.Body.String())
	}
	listAfterBody := decodeFlowEnvelope(t, listAfterResp)
	listAfter := decodeFlowJSON[practicecontracts.PageResult[json.RawMessage]](t, listAfterBody.Data)
	result.listAfterItems = decodeFlowJSON[[]flowChallengeListItem](t, mustMarshalJSON(t, listAfter.List))

	progressResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/users/me/progress",
		nil,
		sessionHeaders(studentSession),
		nil,
	)
	if progressResp.Code != http.StatusOK {
		t.Fatalf("unexpected progress status: %d body=%s", progressResp.Code, progressResp.Body.String())
	}
	result.progress = decodeFlowJSON[flowProgressResponse](t, decodeFlowEnvelope(t, progressResp).Data)

	timelineResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/users/me/timeline",
		nil,
		sessionHeaders(studentSession),
		nil,
	)
	if timelineResp.Code != http.StatusOK {
		t.Fatalf("unexpected timeline status: %d body=%s", timelineResp.Code, timelineResp.Body.String())
	}
	result.timeline = decodeFlowJSON[flowTimelineResponse](t, decodeFlowEnvelope(t, timelineResp).Data)

	evidenceResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/teacher/students/"+strconv.FormatInt(env.student.ID, 10)+"/evidence?challenge_id="+strconv.FormatInt(challenge.ID, 10),
		nil,
		sessionHeaders(adminSession),
		nil,
	)
	if evidenceResp.Code != http.StatusOK {
		t.Fatalf("unexpected evidence status: %d body=%s", evidenceResp.Code, evidenceResp.Body.String())
	}
	result.evidence = decodeFlowJSON[flowTeacherEvidenceReviewResponse](t, decodeFlowEnvelope(t, evidenceResp).Data)

	attackSessionsResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/teacher/students/"+strconv.FormatInt(env.student.ID, 10)+"/attack-sessions?challenge_id="+strconv.FormatInt(challenge.ID, 10)+"&mode=practice&result=success&with_events=false",
		nil,
		sessionHeaders(adminSession),
		nil,
	)
	if attackSessionsResp.Code != http.StatusOK {
		t.Fatalf("unexpected attack sessions status: %d body=%s", attackSessionsResp.Code, attackSessionsResp.Body.String())
	}
	result.attackSessions = decodeFlowJSON[flowTeacherAttackSessionResponse](t, decodeFlowEnvelope(t, attackSessionsResp).Data)

	auditResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/admin/audit-logs?action=submit&resource_type=challenge_submission&user_id="+strconv.FormatInt(env.student.ID, 10),
		nil,
		sessionHeaders(adminSession),
		nil,
	)
	if auditResp.Code != http.StatusOK {
		t.Fatalf("unexpected audit status: %d body=%s", auditResp.Code, auditResp.Body.String())
	}
	result.auditPage = decodeFlowJSON[flowPageResponse[flowAuditItem]](t, decodeFlowEnvelope(t, auditResp).Data)

	var submissions []contestcontracts.Submission
	if err := env.db.Order("submitted_at ASC").Find(&submissions).Error; err != nil {
		t.Fatalf("query submissions: %v", err)
	}
	result.submissions = submissions

	return result
}
