package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	systemapp "ctf-platform/internal/testutil/systemapp"
)

type flowTestEnv struct {
	router *gin.Engine
	db     *gorm.DB
	cache  *redislib.Client
	image  *appImageRow
}

type flowEnvelope = systemapp.FlowEnvelope
type flowLoginResponse = systemapp.FlowLoginResponse
type flowChallengeResponse = systemapp.FlowChallengeResponse
type flowChallengeListItem = systemapp.FlowChallengeListItem
type flowChallengeDetail = systemapp.FlowChallengeDetail
type flowSubmissionResponse = systemapp.FlowSubmissionResponse
type flowSubmissionRecord = systemapp.FlowSubmissionRecord
type flowInstanceResponse = systemapp.FlowInstanceResponse
type flowPageResponse[T any] = systemapp.FlowPageResponse[T]
type flowProgressResponse = systemapp.FlowProgressResponse
type flowTimelineResponse = systemapp.FlowTimelineResponse
type flowTeacherEvidenceReviewResponse = systemapp.FlowTeacherEvidenceReviewResponse
type flowTeacherAttackSessionResponse = systemapp.FlowTeacherAttackSessionResponse
type flowAuditItem = systemapp.FlowAuditItem

func newPracticeFlowTestEnv(t *testing.T) *flowTestEnv {
	t.Helper()

	env := systemapp.NewPracticeFlowTestEnv(t)
	return &flowTestEnv{
		router: env.Router,
		db:     env.DB,
		cache:  env.Cache,
		image:  &appImageRow{ID: env.ImageID},
	}
}

func newPracticeFlowTestConfig(t *testing.T) *config.Config {
	t.Helper()
	return systemapp.NewPracticeFlowTestConfig(t)
}

func createFlowImage(t *testing.T, db *gorm.DB) *appImageRow {
	t.Helper()

	image := &appImageRow{
		Name:   "ctf/web-basic",
		Tag:    "v1",
		Status: challengecontracts.ImageStatusAvailable,
	}
	if err := db.Create(image).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	return image
}

func loginForSession(t *testing.T, router http.Handler, username, password string) *http.Cookie {
	t.Helper()
	return systemapp.LoginForSession(t, router, username, password)
}

func sessionHeaders(cookie *http.Cookie) map[string]string {
	return systemapp.SessionHeaders(cookie)
}

func loginForToken(t *testing.T, router http.Handler, username, password string) string {
	t.Helper()
	return systemapp.LoginForToken(t, router, username, password)
}

func bearerHeaders(token string) map[string]string {
	return systemapp.BearerHeaders(token)
}

func performFlowJSONRequest(
	t *testing.T,
	router http.Handler,
	method string,
	target string,
	payload any,
	headers map[string]string,
	cookies []*http.Cookie,
) *httptest.ResponseRecorder {
	t.Helper()
	return systemapp.PerformFlowJSONRequest(t, router, method, target, payload, headers, cookies)
}

func decodeFlowEnvelope(t *testing.T, recorder *httptest.ResponseRecorder) flowEnvelope {
	t.Helper()
	return systemapp.DecodeFlowEnvelope(t, recorder)
}

func decodeFlowJSON[T any](t *testing.T, data []byte) T {
	t.Helper()
	return systemapp.DecodeFlowJSON[T](t, data)
}

func mustMarshalJSON(t *testing.T, value any) []byte {
	t.Helper()
	return systemapp.MustMarshalJSON(t, value)
}

func assertTimelineHasSubmit(t *testing.T, events []systemapp.FlowTimelineEvent, challengeID int64, isCorrect bool, points int) {
	t.Helper()
	systemapp.AssertTimelineHasSubmit(t, events, challengeID, isCorrect, points)
}

func assertTeacherEvidenceHasEvent(t *testing.T, events []systemapp.FlowTeacherEvidenceReviewEvent, wantType string, challengeID int64, metaKey, metaValue string) {
	t.Helper()
	systemapp.AssertTeacherEvidenceHasEvent(t, events, wantType, challengeID, metaKey, metaValue)
}

func assertTimelineHasChallengeDetailView(t *testing.T, events []systemapp.FlowTimelineEvent, challengeID int64) {
	t.Helper()
	systemapp.AssertTimelineHasChallengeDetailView(t, events, challengeID)
}

func assertTimelineHasInstanceAccess(t *testing.T, events []systemapp.FlowTimelineEvent, challengeID int64) {
	t.Helper()
	systemapp.AssertTimelineHasInstanceAccess(t, events, challengeID)
}

func assertTimelineHasProxyTrace(t *testing.T, events []systemapp.FlowTimelineEvent, challengeID int64) {
	t.Helper()
	systemapp.AssertTimelineHasProxyTrace(t, events, challengeID)
}
