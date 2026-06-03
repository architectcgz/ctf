package practiceflow

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	practicecmd "ctf-platform/internal/module/practice/application/commands"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	"ctf-platform/internal/shared/taxonomy"
	systemapp "ctf-platform/internal/testutil/systemapp"
)

func VerifyPublishedChallengeLifecycleAndAccess(t *testing.T) {
	t.Helper()

	result := systemapp.RunPublishedPracticeFlowScenario(t)

	if len(result.ListBeforeItems) != 1 {
		t.Fatalf("expected 1 published challenge, got %+v", result.ListBeforeItems)
	}
	if result.ListBeforeItems[0].IsSolved {
		t.Fatalf("expected challenge to be unsolved before submission")
	}

	if bytes.Contains(result.DetailBody, []byte(`"is_unlocked"`)) {
		t.Fatalf("expected challenge detail payload to omit is_unlocked, got %s", string(result.DetailBody))
	}
	if bytes.Contains(result.DetailBody, []byte(`"cost_points"`)) {
		t.Fatalf("expected challenge detail payload to omit cost_points, got %s", string(result.DetailBody))
	}
	if result.Detail.IsSolved {
		t.Fatalf("expected challenge detail to be unsolved before submission")
	}
	if result.Detail.AttachmentURL != "https://example.com/files/web-sqli-101.zip" {
		t.Fatalf("unexpected attachment_url: %s", result.Detail.AttachmentURL)
	}
	if !result.Detail.NeedTarget {
		t.Fatalf("expected need_target=true, got false")
	}
	if len(result.Detail.Hints) != 1 || result.Detail.Hints[0].Content == "" {
		t.Fatalf("expected hint content available in challenge detail, got %+v", result.Detail.Hints)
	}

	if result.Instance.ID <= 0 || result.Instance.AccessURL == "" {
		t.Fatalf("expected instance to expose access url, got %+v", result.Instance)
	}
	if !strings.Contains(result.ProxyAccess.AccessURL, "/api/v1/instances/") || !strings.Contains(result.ProxyAccess.AccessURL, "/proxy/") {
		t.Fatalf("expected proxied instance access url, got %+v", result.ProxyAccess)
	}
	if result.ProxyLocation == "" || strings.Contains(result.ProxyLocation, "ticket=") {
		t.Fatalf("expected sanitized proxy redirect location, got %q", result.ProxyLocation)
	}
	if len(result.ProxyCookies) == 0 {
		t.Fatal("expected proxy bootstrap to issue cookie")
	}
}

func VerifyPublishedChallengeSubmissionsAndProgress(t *testing.T) {
	t.Helper()

	result := systemapp.RunPublishedPracticeFlowScenario(t)

	if result.WrongSubmission.IsCorrect {
		t.Fatalf("expected wrong flag submission to be incorrect")
	}
	if result.WrongSubmission.Message != "" {
		t.Fatalf("expected wrong submission message to be omitted, got %+v", result.WrongSubmission)
	}

	if !result.CorrectSubmission.IsCorrect {
		t.Fatalf("expected correct flag submission to succeed")
	}
	if result.CorrectSubmission.Points != 100 {
		t.Fatalf("expected 100 points, got %d", result.CorrectSubmission.Points)
	}
	if result.CorrectSubmission.Message != "" {
		t.Fatalf("expected correct submission message to be omitted, got %+v", result.CorrectSubmission)
	}

	if len(result.SubmissionHistory) != 2 {
		t.Fatalf("expected 2 submission history records, got %d", len(result.SubmissionHistory))
	}
	if result.SubmissionHistory[0].Status != practicecmd.SubmissionStatusCorrect {
		t.Fatalf("unexpected latest submission record: %+v", result.SubmissionHistory[0])
	}
	if result.SubmissionHistory[0].Message != "" {
		t.Fatalf("expected latest submission record message to be omitted, got %+v", result.SubmissionHistory[0])
	}
	if result.SubmissionHistory[1].Status != practicecmd.SubmissionStatusIncorrect {
		t.Fatalf("unexpected previous submission record: %+v", result.SubmissionHistory[1])
	}
	if result.SubmissionHistory[1].Message != "" {
		t.Fatalf("expected previous submission record message to be omitted, got %+v", result.SubmissionHistory[1])
	}

	if !result.RepeatSubmission.IsCorrect {
		t.Fatalf("expected repeated correct submission to stay correct, got %+v", result.RepeatSubmission)
	}
	if result.RepeatSubmission.Points != 0 {
		t.Fatalf("expected repeated correct submission not to award points, got %+v", result.RepeatSubmission)
	}

	if len(result.ListAfterItems) != 1 {
		t.Fatalf("expected 1 challenge after submit, got %+v", result.ListAfterItems)
	}
	if !result.ListAfterItems[0].IsSolved {
		t.Fatalf("expected challenge to be solved after correct submission")
	}
	if result.ListAfterItems[0].SolvedCount != 1 {
		t.Fatalf("expected solved_count 1, got %d", result.ListAfterItems[0].SolvedCount)
	}
	if result.ListAfterItems[0].TotalAttempts != 2 {
		t.Fatalf("expected total_attempts 2, got %d", result.ListAfterItems[0].TotalAttempts)
	}

	if result.Progress.TotalSolved != 1 {
		t.Fatalf("expected total_solved 1, got %d", result.Progress.TotalSolved)
	}
	if result.Progress.TotalScore != 100 {
		t.Fatalf("expected total_score 100, got %d", result.Progress.TotalScore)
	}
	if result.Progress.Rank != 1 {
		t.Fatalf("expected rank 1, got %d", result.Progress.Rank)
	}

	if len(result.Submissions) != 2 {
		t.Fatalf("expected 2 submission records, got %d", len(result.Submissions))
	}
	if result.Submissions[0].IsCorrect {
		t.Fatalf("expected first submission to be incorrect")
	}
	if !result.Submissions[1].IsCorrect {
		t.Fatalf("expected second submission to be correct")
	}
}

func VerifyPublishedChallengeGeneratesTeacherEvidenceAndAuditTrail(t *testing.T) {
	t.Helper()

	result := systemapp.RunPublishedPracticeFlowScenario(t)

	if len(result.Timeline.Events) < 2 {
		t.Fatalf("expected at least two timeline events, got %+v", result.Timeline.Events)
	}
	systemapp.AssertTimelineHasChallengeDetailView(t, result.Timeline.Events, result.Challenge.ID)
	systemapp.AssertTimelineHasInstanceAccess(t, result.Timeline.Events, result.Challenge.ID)
	systemapp.AssertTimelineHasProxyTrace(t, result.Timeline.Events, result.Challenge.ID)
	systemapp.AssertTimelineHasSubmit(t, result.Timeline.Events, result.Challenge.ID, false, 0)
	systemapp.AssertTimelineHasSubmit(t, result.Timeline.Events, result.Challenge.ID, true, 100)

	if result.Evidence.Summary.TotalEvents < 4 {
		t.Fatalf("expected evidence summary to contain >= 4 events, got %+v", result.Evidence.Summary)
	}
	if result.Evidence.Summary.ProxyRequestCount < 1 {
		t.Fatalf("expected evidence summary to count proxy request, got %+v", result.Evidence.Summary)
	}
	if result.Evidence.Summary.SubmitCount != 2 {
		t.Fatalf("expected evidence summary to count 2 submissions, got %+v", result.Evidence.Summary)
	}
	systemapp.AssertTeacherEvidenceHasEvent(t, result.Evidence.Events, "instance_access", result.Challenge.ID, "event_stage", "access")
	systemapp.AssertTeacherEvidenceHasEvent(t, result.Evidence.Events, "instance_proxy_request", result.Challenge.ID, "event_stage", "exploit")
	systemapp.AssertTeacherEvidenceHasEvent(t, result.Evidence.Events, "challenge_submission", result.Challenge.ID, "event_stage", "submit")

	if result.AttackSessions.Summary.TotalSessions != 1 {
		t.Fatalf("expected 1 attack session, got %+v", result.AttackSessions.Summary)
	}
	if result.AttackSessions.Summary.SuccessCount != 1 {
		t.Fatalf("expected 1 successful attack session, got %+v", result.AttackSessions.Summary)
	}
	if result.AttackSessions.Summary.EventCount < 4 {
		t.Fatalf("expected aggregated attack session events >= 4, got %+v", result.AttackSessions.Summary)
	}
	if len(result.AttackSessions.Sessions) != 1 {
		t.Fatalf("expected 1 session payload, got %+v", result.AttackSessions.Sessions)
	}
	if result.AttackSessions.Sessions[0].Mode != "practice" {
		t.Fatalf("expected practice mode session, got %+v", result.AttackSessions.Sessions[0])
	}
	if result.AttackSessions.Sessions[0].Result != "success" {
		t.Fatalf("expected successful session, got %+v", result.AttackSessions.Sessions[0])
	}
	if result.AttackSessions.Sessions[0].ChallengeID == nil || *result.AttackSessions.Sessions[0].ChallengeID != result.Challenge.ID {
		t.Fatalf("expected challenge id %d, got %+v", result.Challenge.ID, result.AttackSessions.Sessions[0].ChallengeID)
	}
	if result.AttackSessions.Sessions[0].Events != nil {
		t.Fatalf("expected events to be omitted when with_events=false, got %+v", result.AttackSessions.Sessions[0].Events)
	}

	challengeSubmitAuditCount := 0
	proxySubmitAuditCount := 0
	for _, item := range result.AuditPage.List {
		switch item.ResourceType {
		case "challenge_submission":
			challengeSubmitAuditCount++
		case "instance_proxy_request":
			proxySubmitAuditCount++
		}
	}
	if challengeSubmitAuditCount != 2 {
		t.Fatalf("expected 2 challenge submission audit logs, got %+v", result.AuditPage.List)
	}
	if proxySubmitAuditCount < 1 {
		t.Fatalf("expected submit audit page to include proxy request logs, got %+v", result.AuditPage.List)
	}
}

func VerifyUnpublishedChallengeCannotBeSolved(t *testing.T) {
	t.Helper()

	env := systemapp.NewPracticeFlowTestEnv(t)

	adminSession := systemapp.LoginForSession(t, env.Router, "admin_user", "Password123")
	studentSession := systemapp.LoginForSession(t, env.Router, "student_user", "Password123")

	createResp := systemapp.PerformFlowJSONRequest(
		t,
		env.Router,
		http.MethodPost,
		"/api/v1/authoring/challenges",
		map[string]any{
			"title":       "Draft Crypto",
			"description": "not published yet",
			"category":    taxonomy.DimensionCrypto,
			"difficulty":  taxonomy.DifficultyMedium,
			"points":      150,
			"image_id":    env.ImageID,
		},
		systemapp.SessionHeaders(adminSession),
		nil,
	)
	if createResp.Code != http.StatusOK {
		t.Fatalf("unexpected create challenge status: %d body=%s", createResp.Code, createResp.Body.String())
	}
	createBody := systemapp.DecodeFlowEnvelope(t, createResp)
	challenge := systemapp.DecodeFlowJSON[systemapp.FlowChallengeResponse](t, createBody.Data)

	configureFlagResp := systemapp.PerformFlowJSONRequest(
		t,
		env.Router,
		http.MethodPut,
		"/api/v1/authoring/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/flag",
		map[string]any{
			"flag_type": "static",
			"flag":      "flag{draft_secret}",
		},
		systemapp.SessionHeaders(adminSession),
		nil,
	)
	if configureFlagResp.Code != http.StatusOK {
		t.Fatalf("unexpected configure draft flag status: %d body=%s", configureFlagResp.Code, configureFlagResp.Body.String())
	}

	listResp := systemapp.PerformFlowJSONRequest(
		t,
		env.Router,
		http.MethodGet,
		"/api/v1/challenges",
		nil,
		systemapp.SessionHeaders(studentSession),
		nil,
	)
	if listResp.Code != http.StatusOK {
		t.Fatalf("unexpected list challenges status: %d body=%s", listResp.Code, listResp.Body.String())
	}
	listBody := systemapp.DecodeFlowEnvelope(t, listResp)
	listPage := systemapp.DecodeFlowJSON[practicecontracts.PageResult[json.RawMessage]](t, listBody.Data)
	listItems := systemapp.DecodeFlowJSON[[]systemapp.FlowChallengeListItem](t, systemapp.MustMarshalJSON(t, listPage.List))
	if len(listItems) != 0 {
		t.Fatalf("expected unpublished challenge to stay hidden, got %+v", listItems)
	}

	detailResp := systemapp.PerformFlowJSONRequest(
		t,
		env.Router,
		http.MethodGet,
		"/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10),
		nil,
		systemapp.SessionHeaders(studentSession),
		nil,
	)
	if detailResp.Code != http.StatusForbidden {
		t.Fatalf("expected unpublished challenge detail to return 403, got %d body=%s", detailResp.Code, detailResp.Body.String())
	}

	submitResp := systemapp.PerformFlowJSONRequest(
		t,
		env.Router,
		http.MethodPost,
		"/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/submit",
		map[string]any{"flag": "flag{draft_secret}"},
		systemapp.SessionHeaders(studentSession),
		nil,
	)
	if submitResp.Code != http.StatusForbidden {
		t.Fatalf("expected unpublished challenge submit to return 403, got %d body=%s", submitResp.Code, submitResp.Body.String())
	}
	submitBody := systemapp.DecodeFlowEnvelope(t, submitResp)
	if submitBody.Code != challengecontracts.ErrChallengeNotPublish.Code {
		t.Fatalf("expected challenge not published code %d, got %d", challengecontracts.ErrChallengeNotPublish.Code, submitBody.Code)
	}
}
