package app

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	"ctf-platform/internal/shared/taxonomy"
)

func TestPracticeFlow_PublishedChallengeGeneratesTeacherEvidenceAndAuditTrail(t *testing.T) {
	result := runPublishedPracticeFlowScenario(t)

	if len(result.timeline.Events) < 2 {
		t.Fatalf("expected at least two timeline events, got %+v", result.timeline.Events)
	}
	assertTimelineHasChallengeDetailView(t, result.timeline.Events, result.challenge.ID)
	assertTimelineHasInstanceAccess(t, result.timeline.Events, result.challenge.ID)
	assertTimelineHasProxyTrace(t, result.timeline.Events, result.challenge.ID)
	assertTimelineHasSubmit(t, result.timeline.Events, result.challenge.ID, false, 0)
	assertTimelineHasSubmit(t, result.timeline.Events, result.challenge.ID, true, 100)

	if result.evidence.Summary.TotalEvents < 4 {
		t.Fatalf("expected evidence summary to contain >= 4 events, got %+v", result.evidence.Summary)
	}
	if result.evidence.Summary.ProxyRequestCount < 1 {
		t.Fatalf("expected evidence summary to count proxy request, got %+v", result.evidence.Summary)
	}
	if result.evidence.Summary.SubmitCount != 2 {
		t.Fatalf("expected evidence summary to count 2 submissions, got %+v", result.evidence.Summary)
	}
	assertTeacherEvidenceHasEvent(t, result.evidence.Events, "instance_access", result.challenge.ID, "event_stage", "access")
	assertTeacherEvidenceHasEvent(t, result.evidence.Events, "instance_proxy_request", result.challenge.ID, "event_stage", "exploit")
	assertTeacherEvidenceHasEvent(t, result.evidence.Events, "challenge_submission", result.challenge.ID, "event_stage", "submit")

	if result.attackSessions.Summary.TotalSessions != 1 {
		t.Fatalf("expected 1 attack session, got %+v", result.attackSessions.Summary)
	}
	if result.attackSessions.Summary.SuccessCount != 1 {
		t.Fatalf("expected 1 successful attack session, got %+v", result.attackSessions.Summary)
	}
	if result.attackSessions.Summary.EventCount < 4 {
		t.Fatalf("expected aggregated attack session events >= 4, got %+v", result.attackSessions.Summary)
	}
	if len(result.attackSessions.Sessions) != 1 {
		t.Fatalf("expected 1 session payload, got %+v", result.attackSessions.Sessions)
	}
	if result.attackSessions.Sessions[0].Mode != "practice" {
		t.Fatalf("expected practice mode session, got %+v", result.attackSessions.Sessions[0])
	}
	if result.attackSessions.Sessions[0].Result != "success" {
		t.Fatalf("expected successful session, got %+v", result.attackSessions.Sessions[0])
	}
	if result.attackSessions.Sessions[0].ChallengeID == nil || *result.attackSessions.Sessions[0].ChallengeID != result.challenge.ID {
		t.Fatalf("expected challenge id %d, got %+v", result.challenge.ID, result.attackSessions.Sessions[0].ChallengeID)
	}
	if result.attackSessions.Sessions[0].Events != nil {
		t.Fatalf("expected events to be omitted when with_events=false, got %+v", result.attackSessions.Sessions[0].Events)
	}

	if len(result.auditPage.List) != 2 {
		t.Fatalf("expected 2 submit audit logs, got %+v", result.auditPage.List)
	}
}

func TestPracticeFlow_UnpublishedChallengeCannotBeSolved(t *testing.T) {
	env := newPracticeFlowTestEnv(t)

	adminSession := loginForSession(t, env.router, "admin_user", "Password123")
	studentSession := loginForSession(t, env.router, "student_user", "Password123")

	createResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/authoring/challenges",
		map[string]any{
			"title":       "Draft Crypto",
			"description": "not published yet",
			"category":    taxonomy.DimensionCrypto,
			"difficulty":  taxonomy.DifficultyMedium,
			"points":      150,
			"image_id":    env.image.ID,
		},
		sessionHeaders(adminSession),
		nil,
	)
	if createResp.Code != http.StatusOK {
		t.Fatalf("unexpected create challenge status: %d body=%s", createResp.Code, createResp.Body.String())
	}
	createBody := decodeFlowEnvelope(t, createResp)
	challenge := decodeFlowJSON[flowChallengeResponse](t, createBody.Data)

	configureFlagResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodPut,
		"/api/v1/authoring/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/flag",
		map[string]any{
			"flag_type": "static",
			"flag":      "flag{draft_secret}",
		},
		sessionHeaders(adminSession),
		nil,
	)
	if configureFlagResp.Code != http.StatusOK {
		t.Fatalf("unexpected configure draft flag status: %d body=%s", configureFlagResp.Code, configureFlagResp.Body.String())
	}

	listResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/challenges",
		nil,
		sessionHeaders(studentSession),
		nil,
	)
	if listResp.Code != http.StatusOK {
		t.Fatalf("unexpected list challenges status: %d body=%s", listResp.Code, listResp.Body.String())
	}
	listBody := decodeFlowEnvelope(t, listResp)
	listPage := decodeFlowJSON[practicecontracts.PageResult[json.RawMessage]](t, listBody.Data)
	listItems := decodeFlowJSON[[]flowChallengeListItem](t, mustMarshalJSON(t, listPage.List))
	if len(listItems) != 0 {
		t.Fatalf("expected unpublished challenge to stay hidden, got %+v", listItems)
	}

	detailResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10),
		nil,
		sessionHeaders(studentSession),
		nil,
	)
	if detailResp.Code != http.StatusForbidden {
		t.Fatalf("expected unpublished challenge detail to return 403, got %d body=%s", detailResp.Code, detailResp.Body.String())
	}

	submitResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/submit",
		map[string]any{"flag": "flag{draft_secret}"},
		sessionHeaders(studentSession),
		nil,
	)
	if submitResp.Code != http.StatusForbidden {
		t.Fatalf("expected unpublished challenge submit to return 403, got %d body=%s", submitResp.Code, submitResp.Body.String())
	}
	submitBody := decodeFlowEnvelope(t, submitResp)
	if submitBody.Code != challengecontracts.ErrChallengeNotPublish.Code {
		t.Fatalf("expected challenge not published code %d, got %d", challengecontracts.ErrChallengeNotPublish.Code, submitBody.Code)
	}
}
