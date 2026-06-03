package app

import (
	"net/http"
	"testing"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
	systemapp "ctf-platform/internal/testutil/systemapp"
)

type practiceFlowScenarioResult struct {
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

	result := systemapp.RunPublishedPracticeFlowScenario(t)
	return &practiceFlowScenarioResult{
		adminSession:      result.AdminSession,
		studentSession:    result.StudentSession,
		challenge:         result.Challenge,
		listBeforeItems:   result.ListBeforeItems,
		detailBody:        result.DetailBody,
		detail:            result.Detail,
		instance:          result.Instance,
		proxyAccess:       result.ProxyAccess,
		proxyLocation:     result.ProxyLocation,
		proxyCookies:      result.ProxyCookies,
		wrongSubmission:   result.WrongSubmission,
		correctSubmission: result.CorrectSubmission,
		repeatSubmission:  result.RepeatSubmission,
		submissionHistory: result.SubmissionHistory,
		listAfterItems:    result.ListAfterItems,
		progress:          result.Progress,
		timeline:          result.Timeline,
		evidence:          result.Evidence,
		attackSessions:    result.AttackSessions,
		auditPage:         result.AuditPage,
		submissions:       result.Submissions,
	}
}
