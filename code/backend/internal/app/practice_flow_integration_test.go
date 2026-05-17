package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/middleware"
	"ctf-platform/internal/model"
	authhttp "ctf-platform/internal/module/auth/api/http"
	authcmd "ctf-platform/internal/module/auth/application/commands"
	authqry "ctf-platform/internal/module/auth/application/queries"
	authinfra "ctf-platform/internal/module/auth/infrastructure"
	challengehttp "ctf-platform/internal/module/challenge/api/http"
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	challengeqry "ctf-platform/internal/module/challenge/application/queries"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	challengeruntime "ctf-platform/internal/module/challenge/runtime"
	identitycmd "ctf-platform/internal/module/identity/application/commands"
	identityqry "ctf-platform/internal/module/identity/application/queries"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	identityinfra "ctf-platform/internal/module/identity/infrastructure"
	instancecmd "ctf-platform/internal/module/instance/application/commands"
	instanceqry "ctf-platform/internal/module/instance/application/queries"
	opshttp "ctf-platform/internal/module/ops/api/http"
	opscmd "ctf-platform/internal/module/ops/application/commands"
	opsqry "ctf-platform/internal/module/ops/application/queries"
	opsinfra "ctf-platform/internal/module/ops/infrastructure"
	practicehttp "ctf-platform/internal/module/practice/api/http"
	practicecmd "ctf-platform/internal/module/practice/application/commands"
	practiceqry "ctf-platform/internal/module/practice/application/queries"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	runtimehttp "ctf-platform/internal/module/runtime/api/http"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimeinfrarepo "ctf-platform/internal/module/runtime/infrastructure"
	teachingqueryhttp "ctf-platform/internal/module/teaching_query/api/http"
	teachingqueryqueries "ctf-platform/internal/module/teaching_query/application/queries"
	teachingqueryinfra "ctf-platform/internal/module/teaching_query/infrastructure"
	runtimeadapters "ctf-platform/internal/testutil/runtimeadapters"
	"ctf-platform/internal/validation"
	"ctf-platform/pkg/errcode"
)

type flowTestEnv struct {
	router  *gin.Engine
	db      *gorm.DB
	cache   *redislib.Client
	admin   *model.User
	student *model.User
	image   *model.Image
}

type flowEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type flowLoginResponse struct {
	User struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Role     string `json:"role"`
	} `json:"user"`
}

type teachingQueryIdentityLookupAdapter struct {
	users identitycontracts.UserLookupRepository
}

func (a teachingQueryIdentityLookupAdapter) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	user, err := a.users.FindByID(ctx, userID)
	if errors.Is(err, identitycontracts.ErrUserNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

type flowChallengeResponse struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Difficulty  string `json:"difficulty"`
	Points      int    `json:"points"`
	ImageID     int64  `json:"image_id"`
	Status      string `json:"status"`
}

type flowChallengeListItem struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	Category      string `json:"category"`
	Difficulty    string `json:"difficulty"`
	Points        int    `json:"points"`
	SolvedCount   int64  `json:"solved_count"`
	TotalAttempts int64  `json:"total_attempts"`
	IsSolved      bool   `json:"is_solved"`
}

type flowChallengeDetail struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	Category      string `json:"category"`
	Difficulty    string `json:"difficulty"`
	Points        int    `json:"points"`
	NeedTarget    bool   `json:"need_target"`
	AttachmentURL string `json:"attachment_url"`
	Hints         []struct {
		Level   int    `json:"level"`
		Content string `json:"content"`
	} `json:"hints"`
	SolvedCount   int64 `json:"solved_count"`
	TotalAttempts int64 `json:"total_attempts"`
	IsSolved      bool  `json:"is_solved"`
}

type flowSubmissionResponse struct {
	IsCorrect bool   `json:"is_correct"`
	Message   string `json:"message"`
	Points    int    `json:"points"`
}

type flowSubmissionRecord struct {
	ID          int64  `json:"id"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	Answer      string `json:"answer"`
	SubmittedAt string `json:"submitted_at"`
}

type flowInstanceResponse struct {
	ID        int64  `json:"id"`
	AccessURL string `json:"access_url"`
	Status    string `json:"status"`
}

type flowPageResponse[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

type flowProgressResponse struct {
	TotalScore  int `json:"total_score"`
	TotalSolved int `json:"total_solved"`
	Rank        int `json:"rank"`
}

type flowTimelineResponse struct {
	Events []struct {
		Type        string `json:"type"`
		ChallengeID int64  `json:"challenge_id"`
		Title       string `json:"title"`
		IsCorrect   *bool  `json:"is_correct"`
		Points      *int   `json:"points"`
		Detail      string `json:"detail"`
	} `json:"events"`
}

type flowTeacherEvidenceReviewResponse struct {
	Summary struct {
		TotalEvents       int   `json:"total_events"`
		ProxyRequestCount int   `json:"proxy_request_count"`
		SubmitCount       int   `json:"submit_count"`
		SuccessCount      int   `json:"success_count"`
		ChallengeID       int64 `json:"challenge_id"`
	} `json:"summary"`
	Events []struct {
		Type        string                 `json:"type"`
		ChallengeID int64                  `json:"challenge_id"`
		Title       string                 `json:"title"`
		Detail      string                 `json:"detail"`
		Meta        map[string]interface{} `json:"meta"`
	} `json:"events"`
}

type flowTeacherAttackSessionResponse struct {
	Summary struct {
		TotalSessions   int `json:"total_sessions"`
		SuccessCount    int `json:"success_count"`
		FailedCount     int `json:"failed_count"`
		InProgressCount int `json:"in_progress_count"`
		UnknownCount    int `json:"unknown_count"`
		EventCount      int `json:"event_count"`
	} `json:"summary"`
	Sessions []struct {
		ID          string `json:"id"`
		Mode        string `json:"mode"`
		ChallengeID *int64 `json:"challenge_id"`
		Result      string `json:"result"`
		EventCount  int    `json:"event_count"`
		Events      []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"events"`
	} `json:"sessions"`
}

type flowAuditItem struct {
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   *int64                 `json:"resource_id"`
	ActorUserID  *int64                 `json:"actor_user_id"`
	Detail       map[string]interface{} `json:"detail"`
}

func TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge(t *testing.T) {
	env := newPracticeFlowTestEnv(t)

	adminSession := loginForSession(t, env.router, "admin_user", "Password123")
	studentSession := loginForSession(t, env.router, "student_user", "Password123")

	createResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/authoring/challenges",
		map[string]any{
			"title":          "Web SQLi 101",
			"description":    "basic sql injection challenge",
			"category":       model.DimensionWeb,
			"difficulty":     model.ChallengeDifficultyEasy,
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

	if err := env.db.Model(&model.Challenge{}).
		Where("id = ?", challenge.ID).
		Update("status", model.ChallengeStatusPublished).Error; err != nil {
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
	listBefore := decodeFlowJSON[dto.PageResult[json.RawMessage]](t, listBeforeBody.Data)
	listBeforeItems := decodeFlowJSON[[]flowChallengeListItem](t, mustMarshalJSON(t, listBefore.List))
	if len(listBeforeItems) != 1 {
		t.Fatalf("expected 1 published challenge, got %+v", listBeforeItems)
	}
	if listBeforeItems[0].IsSolved {
		t.Fatalf("expected challenge to be unsolved before submission")
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
	if detailResp.Code != http.StatusOK {
		t.Fatalf("unexpected challenge detail status: %d body=%s", detailResp.Code, detailResp.Body.String())
	}
	detailBody := decodeFlowEnvelope(t, detailResp)
	if bytes.Contains(detailBody.Data, []byte(`"is_unlocked"`)) {
		t.Fatalf("expected challenge detail payload to omit is_unlocked, got %s", string(detailBody.Data))
	}
	if bytes.Contains(detailBody.Data, []byte(`"cost_points"`)) {
		t.Fatalf("expected challenge detail payload to omit cost_points, got %s", string(detailBody.Data))
	}
	detail := decodeFlowJSON[flowChallengeDetail](t, detailBody.Data)
	if detail.IsSolved {
		t.Fatalf("expected challenge detail to be unsolved before submission")
	}
	if detail.AttachmentURL != "https://example.com/files/web-sqli-101.zip" {
		t.Fatalf("unexpected attachment_url: %s", detail.AttachmentURL)
	}
	if !detail.NeedTarget {
		t.Fatalf("expected need_target=true, got false")
	}
	if len(detail.Hints) != 1 || detail.Hints[0].Content == "" {
		t.Fatalf("expected hint content available in challenge detail, got %+v", detail.Hints)
	}

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
	if instance.ID <= 0 || instance.AccessURL == "" {
		t.Fatalf("expected instance to expose access url, got %+v", instance)
	}

	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/submit" {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("submitted"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("target ok"))
	}))
	defer targetServer.Close()
	if err := env.db.Model(&model.Instance{}).
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
	if !strings.Contains(proxyAccess.AccessURL, "/api/v1/instances/"+strconv.FormatInt(instance.ID, 10)+"/proxy/") {
		t.Fatalf("expected proxied instance access url, got %+v", proxyAccess)
	}

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
	location := proxyBootstrapResp.Header().Get("Location")
	if location == "" || strings.Contains(location, "ticket=") {
		t.Fatalf("expected sanitized proxy redirect location, got %q", location)
	}
	proxyCookies := proxyBootstrapResp.Result().Cookies()
	if len(proxyCookies) == 0 {
		t.Fatal("expected proxy bootstrap to issue cookie")
	}

	proxyPageResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		location,
		nil,
		nil,
		proxyCookies,
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
		proxyCookies,
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
	wrongSubmitBody := decodeFlowEnvelope(t, wrongSubmitResp)
	wrongSubmission := decodeFlowJSON[flowSubmissionResponse](t, wrongSubmitBody.Data)
	if wrongSubmission.IsCorrect {
		t.Fatalf("expected wrong flag submission to be incorrect")
	}
	if wrongSubmission.Message != "" {
		t.Fatalf("expected wrong submission message to be omitted, got %+v", wrongSubmission)
	}

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
	correctSubmitBody := decodeFlowEnvelope(t, correctSubmitResp)
	correctSubmission := decodeFlowJSON[flowSubmissionResponse](t, correctSubmitBody.Data)
	if !correctSubmission.IsCorrect {
		t.Fatalf("expected correct flag submission to succeed")
	}
	if correctSubmission.Points != 100 {
		t.Fatalf("expected 100 points, got %d", correctSubmission.Points)
	}
	if correctSubmission.Message != "" {
		t.Fatalf("expected correct submission message to be omitted, got %+v", correctSubmission)
	}

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
	submissionHistoryBody := decodeFlowEnvelope(t, submissionHistoryResp)
	submissionHistory := decodeFlowJSON[[]flowSubmissionRecord](t, submissionHistoryBody.Data)
	if len(submissionHistory) != 2 {
		t.Fatalf("expected 2 submission history records, got %d", len(submissionHistory))
	}
	if submissionHistory[0].Status != dto.SubmissionStatusCorrect {
		t.Fatalf("unexpected latest submission record: %+v", submissionHistory[0])
	}
	if submissionHistory[0].Message != "" {
		t.Fatalf("expected latest submission record message to be omitted, got %+v", submissionHistory[0])
	}
	if submissionHistory[1].Status != dto.SubmissionStatusIncorrect {
		t.Fatalf("unexpected previous submission record: %+v", submissionHistory[1])
	}
	if submissionHistory[1].Message != "" {
		t.Fatalf("expected previous submission record message to be omitted, got %+v", submissionHistory[1])
	}

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
	repeatSubmitBody := decodeFlowEnvelope(t, repeatSubmitResp)
	repeatSubmission := decodeFlowJSON[flowSubmissionResponse](t, repeatSubmitBody.Data)
	if !repeatSubmission.IsCorrect {
		t.Fatalf("expected repeated correct submission to stay correct, got %+v", repeatSubmission)
	}
	if repeatSubmission.Points != 0 {
		t.Fatalf("expected repeated correct submission not to award points, got %+v", repeatSubmission)
	}

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
	listAfter := decodeFlowJSON[dto.PageResult[json.RawMessage]](t, listAfterBody.Data)
	listAfterItems := decodeFlowJSON[[]flowChallengeListItem](t, mustMarshalJSON(t, listAfter.List))
	if len(listAfterItems) != 1 {
		t.Fatalf("expected 1 challenge after submit, got %+v", listAfterItems)
	}
	if !listAfterItems[0].IsSolved {
		t.Fatalf("expected challenge to be solved after correct submission")
	}
	if listAfterItems[0].SolvedCount != 1 {
		t.Fatalf("expected solved_count 1, got %d", listAfterItems[0].SolvedCount)
	}
	if listAfterItems[0].TotalAttempts != 2 {
		t.Fatalf("expected total_attempts 2, got %d", listAfterItems[0].TotalAttempts)
	}

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
	progressBody := decodeFlowEnvelope(t, progressResp)
	progress := decodeFlowJSON[flowProgressResponse](t, progressBody.Data)
	if progress.TotalSolved != 1 {
		t.Fatalf("expected total_solved 1, got %d", progress.TotalSolved)
	}
	if progress.TotalScore != 100 {
		t.Fatalf("expected total_score 100, got %d", progress.TotalScore)
	}
	if progress.Rank != 1 {
		t.Fatalf("expected rank 1, got %d", progress.Rank)
	}

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
	timelineBody := decodeFlowEnvelope(t, timelineResp)
	timeline := decodeFlowJSON[flowTimelineResponse](t, timelineBody.Data)
	if len(timeline.Events) < 2 {
		t.Fatalf("expected at least two timeline events, got %+v", timeline.Events)
	}
	assertTimelineHasChallengeDetailView(t, timeline.Events, challenge.ID)
	assertTimelineHasInstanceAccess(t, timeline.Events, challenge.ID)
	assertTimelineHasProxyTrace(t, timeline.Events, challenge.ID)
	assertTimelineHasSubmit(t, timeline.Events, challenge.ID, false, 0)
	assertTimelineHasSubmit(t, timeline.Events, challenge.ID, true, 100)

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
	evidenceBody := decodeFlowEnvelope(t, evidenceResp)
	evidence := decodeFlowJSON[flowTeacherEvidenceReviewResponse](t, evidenceBody.Data)
	if evidence.Summary.TotalEvents < 4 {
		t.Fatalf("expected evidence summary to contain >= 4 events, got %+v", evidence.Summary)
	}
	if evidence.Summary.ProxyRequestCount < 1 {
		t.Fatalf("expected evidence summary to count proxy request, got %+v", evidence.Summary)
	}
	if evidence.Summary.SubmitCount != 2 {
		t.Fatalf("expected evidence summary to count 2 submissions, got %+v", evidence.Summary)
	}
	assertTeacherEvidenceHasEvent(t, evidence.Events, "instance_access", challenge.ID, "event_stage", "access")
	assertTeacherEvidenceHasEvent(t, evidence.Events, "instance_proxy_request", challenge.ID, "event_stage", "exploit")
	assertTeacherEvidenceHasEvent(t, evidence.Events, "challenge_submission", challenge.ID, "event_stage", "submit")

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
	attackSessionsBody := decodeFlowEnvelope(t, attackSessionsResp)
	attackSessions := decodeFlowJSON[flowTeacherAttackSessionResponse](t, attackSessionsBody.Data)
	if attackSessions.Summary.TotalSessions != 1 {
		t.Fatalf("expected 1 attack session, got %+v", attackSessions.Summary)
	}
	if attackSessions.Summary.SuccessCount != 1 {
		t.Fatalf("expected 1 successful attack session, got %+v", attackSessions.Summary)
	}
	if attackSessions.Summary.EventCount < 4 {
		t.Fatalf("expected aggregated attack session events >= 4, got %+v", attackSessions.Summary)
	}
	if len(attackSessions.Sessions) != 1 {
		t.Fatalf("expected 1 session payload, got %+v", attackSessions.Sessions)
	}
	if attackSessions.Sessions[0].Mode != "practice" {
		t.Fatalf("expected practice mode session, got %+v", attackSessions.Sessions[0])
	}
	if attackSessions.Sessions[0].Result != "success" {
		t.Fatalf("expected successful session, got %+v", attackSessions.Sessions[0])
	}
	if attackSessions.Sessions[0].ChallengeID == nil || *attackSessions.Sessions[0].ChallengeID != challenge.ID {
		t.Fatalf("expected challenge id %d, got %+v", challenge.ID, attackSessions.Sessions[0].ChallengeID)
	}
	if attackSessions.Sessions[0].Events != nil {
		t.Fatalf("expected events to be omitted when with_events=false, got %+v", attackSessions.Sessions[0].Events)
	}

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
	auditBody := decodeFlowEnvelope(t, auditResp)
	auditPage := decodeFlowJSON[flowPageResponse[flowAuditItem]](t, auditBody.Data)
	if len(auditPage.List) != 2 {
		t.Fatalf("expected 2 submit audit logs, got %+v", auditPage.List)
	}

	var submissions []model.Submission
	if err := env.db.Order("submitted_at ASC").Find(&submissions).Error; err != nil {
		t.Fatalf("query submissions: %v", err)
	}
	if len(submissions) != 2 {
		t.Fatalf("expected 2 submission records, got %d", len(submissions))
	}
	if submissions[0].IsCorrect {
		t.Fatalf("expected first submission to be incorrect")
	}
	if !submissions[1].IsCorrect {
		t.Fatalf("expected second submission to be correct")
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
			"category":    model.DimensionCrypto,
			"difficulty":  model.ChallengeDifficultyMedium,
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
	listPage := decodeFlowJSON[dto.PageResult[json.RawMessage]](t, listBody.Data)
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
	if submitBody.Code != errcode.ErrChallengeNotPublish.Code {
		t.Fatalf("expected challenge not published code %d, got %d", errcode.ErrChallengeNotPublish.Code, submitBody.Code)
	}
}

func newPracticeFlowTestEnv(t *testing.T) *flowTestEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)
	if err := validation.Register(); err != nil {
		t.Fatalf("register validator: %v", err)
	}

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	cache := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	if err := cache.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("ping test redis: %v", err)
	}

	dbPath := filepath.Join(t.TempDir(), "practice-flow.sqlite")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Role{},
		&model.User{},
		&model.UserRole{},
		&model.AuditLog{},
		&model.Contest{},
		&model.ContestRegistration{},
		&model.Team{},
		&model.TeamMember{},
		&model.Image{},
		&model.Challenge{},
		&model.ChallengePublishCheckJob{},
		&model.ChallengeHint{},
		&model.ChallengeWriteup{},
		&model.ChallengeTopology{},
		&model.ChallengePackageRevision{},
		&model.EnvironmentTemplate{},
		&model.ContestAWDService{},
		&model.Submission{},
		&model.Instance{},
		&model.PortAllocation{},
		&model.AWDServiceOperation{},
		&model.AWDScopeControl{},
		&model.SkillProfile{},
		&model.UserScore{},
	); err != nil {
		t.Fatalf("auto migrate test schema: %v", err)
	}

	cfg := newPracticeFlowTestConfig(t)
	logger := zap.NewNop()

	tokenService := authinfra.NewTokenService(cfg.Auth, cfg.WebSocket, cache)
	authRepo := identityinfra.NewRepository(db)
	authService := authcmd.NewService(authRepo, tokenService, cfg.RateLimit.Login, logger)
	casCommandService := authcmd.NewCASService(cfg.Auth.CAS, authRepo, tokenService, logger.Named("cas_command_service"), nil)
	casQueryService := authqry.NewCASService(cfg.Auth.CAS)
	profileCommandService := identitycmd.NewProfileService(authRepo, logger.Named("identity_profile_command_service"))
	profileQueryService := identityqry.NewProfileService(authRepo)
	auditRepo := opsinfra.NewAuditRepository(db)
	auditCommandService := opscmd.NewAuditService(auditRepo, logger)
	auditQueryService := opsqry.NewAuditService(auditRepo, cfg.Pagination, logger)
	authHandler := authhttp.NewHandler(authService, profileCommandService, profileQueryService, tokenService, casCommandService, casQueryService, authhttp.CookieConfig{
		Name:     cfg.Auth.SessionCookieName,
		Path:     cfg.Auth.SessionCookiePath,
		Secure:   cfg.Auth.SessionCookieSecure,
		HTTPOnly: cfg.Auth.SessionCookieHTTPOnly,
		SameSite: cfg.Auth.CookieSameSite(),
		MaxAge:   cfg.Auth.SessionTTL,
	}, logger, auditCommandService)
	auditHandler := opshttp.NewAuditHandler(auditQueryService)

	challengeRepo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	challengeCommandService := challengecmd.NewChallengeService(
		challengeinfra.NewChallengeCommandRepository(challengeRepo),
		challengeinfra.NewImageQueryRepository(imageRepo),
		challengeinfra.NewTopologyServiceRepository(challengeRepo),
		challengeinfra.NewTopologyPackageRevisionRepository(challengeRepo),
		nil,
		challengecmd.SelfCheckConfig{
			RuntimeCreateTimeout: cfg.Container.CreateTimeout,
			FlagGlobalSecret:     cfg.Container.FlagGlobalSecret,
		},
		logger,
	)
	challengeCommandService.SetChallengeImportTxRunner(challengeruntime.NewChallengeImportTxRunner(challengeRepo, nil))
	challengeCommandService.SetChallengePackageExportTxRunner(challengeruntime.NewChallengePackageExportTxRunner(challengeRepo))
	challengeQueryService := challengeqry.NewChallengeService(challengeRepo, challengeinfra.NewSolvedCountCache(cache), &challengeqry.Config{
		SolvedCountCacheTTL: cfg.Challenge.SolvedCountCacheTTL,
	}, logger)
	challengeHandler := challengehttp.NewHandler(challengeCommandService, challengeQueryService)

	flagQueryService, err := challengeqry.NewFlagService(challengeRepo, cfg.Container.FlagGlobalSecret)
	if err != nil {
		t.Fatalf("create flag query service: %v", err)
	}
	flagCommandService, err := challengecmd.NewFlagService(challengeRepo, cfg.Container.FlagGlobalSecret)
	if err != nil {
		t.Fatalf("create flag command service: %v", err)
	}
	flagHandler := challengehttp.NewFlagHandler(flagCommandService, flagQueryService)

	practiceRepo := practiceinfra.NewRepository(db)
	instanceRepo := runtimeinfrarepo.NewRepository(db)
	root, err := composition.BuildRoot(cfg, logger, db, cache)
	if err != nil {
		t.Fatalf("build composition root: %v", err)
	}
	runtimeModule := composition.BuildRuntimeModule(root)
	instanceModule := composition.BuildInstanceModule(root, runtimeModule)
	runtimeCleanupService := runtimecmd.NewRuntimeCleanupService(nil, nil, logger)
	runtimeInstanceCommands := instancecmd.NewInstanceService(instanceRepo, runtimeCleanupService, &cfg.Container, logger)
	runtimeInstanceQueries := instanceqry.NewInstanceService(instanceRepo, &cfg.Container)
	runtimeProxyTicketService := instanceqry.NewProxyTicketService(runtimeinfrarepo.NewProxyTicketStore(cache), instanceRepo, cfg.Container.ProxyTicketTTL)
	runtimeService := runtimeadapters.NewHTTPService(
		runtimeInstanceCommands,
		runtimeInstanceQueries,
		runtimeProxyTicketService,
		cfg.Container.ProxyBodyPreviewSize,
	)
	scoreStateStore := practiceinfra.NewScoreStateStore(cache)
	flagSubmitRateLimitStore := practiceinfra.NewFlagSubmitRateLimitStore(cache, cfg.RateLimit.RedisKeyPrefix)
	practiceScoreCommandService := practicecmd.NewScoreService(practiceRepo, scoreStateStore, logger, &cfg.Score)
	practiceService := practicecmd.NewService(
		practiceRepo,
		challengeRepo,
		imageRepo,
		instanceModule.PracticeInstanceRepository,
		instanceModule.PracticeRuntimeService,
		practiceScoreCommandService,
		flagSubmitRateLimitStore,
		cfg,
		logger).
		SetSolvedSubmissionRepository(practiceinfra.NewSolvedSubmissionRepository(practiceRepo)).
		SetManualReviewRepository(practiceinfra.NewManualReviewRepository(practiceRepo)).
		SetContestScopeRepository(practiceinfra.NewContestScopeRepository(practiceRepo)).
		SetRuntimeSubjectRepository(practiceinfra.NewRuntimeSubjectRepository(challengeRepo)).
		SetInstanceReadinessProbe(practiceinfra.NewInstanceReadinessProbe())

	practiceScoreQueryService := practiceqry.NewScoreService(practiceinfra.NewScoreQueryRepository(practiceRepo), scoreStateStore, logger, &cfg.Score)
	practiceProgressTimelineService := practiceqry.NewProgressTimelineService(
		practiceRepo,
		practiceinfra.NewProgressCache(cache),
		cfg.Cache.ProgressTTL,
		logger,
	)
	practiceHandler := practicehttp.NewHandler(practiceService, practiceScoreQueryService, practiceProgressTimelineService)
	teachingQueryRepo := teachingqueryinfra.NewRepository(db)
	teachingQueryUsers := teachingQueryIdentityLookupAdapter{users: authRepo}
	teachingQueryService := teachingqueryqueries.NewQueryService(teachingQueryUsers, teachingQueryRepo, cfg.Pagination)
	teachingQueryOverviewService := teachingqueryqueries.NewOverviewService(teachingQueryUsers, teachingQueryRepo)
	teachingQueryClassInsightService := teachingqueryqueries.NewClassInsightService(teachingQueryUsers, teachingQueryRepo, nil, logger)
	teachingQueryStudentReviewService := teachingqueryqueries.NewStudentReviewService(teachingQueryUsers, teachingQueryRepo, nil)
	teachingQueryHandler := teachingqueryhttp.NewHandler(
		teachingQueryService,
		teachingQueryOverviewService,
		teachingQueryClassInsightService,
		teachingQueryStudentReviewService,
	)
	runtimeHandler := runtimehttp.NewHandler(runtimeService, cfg.Container.PublicHost, cfg.Container.AccessHost, auditCommandService, runtimehttp.CookieConfig{}, nil)

	admin := createFlowUser(t, db, "admin_user", "Password123", model.RoleAdmin)
	student := createFlowUser(t, db, "student_user", "Password123", model.RoleStudent)
	image := createFlowImage(t, db)

	router := gin.New()
	router.Use(middleware.RequestID())

	apiV1 := router.Group("/api/v1")
	authGroup := apiV1.Group("/auth")
	authGroup.POST("/login", authHandler.Login)

	protected := apiV1.Group("")
	protected.Use(middleware.Auth(tokenService, cfg.Auth.SessionCookieName))

	authoringOnly := protected.Group("/authoring")
	authoringOnly.Use(middleware.RequireRole(model.RoleTeacher))
	authoringOnly.POST("/challenges", challengeHandler.CreateChallenge)
	authoringOnly.PUT("/challenges/:id/flag", flagHandler.ConfigureFlag)

	adminOnly := protected.Group("/admin")
	adminOnly.Use(middleware.RequireRole(model.RoleAdmin))
	adminOnly.GET("/audit-logs", auditHandler.ListAuditLogs)

	protected.GET("/challenges", challengeHandler.ListPublishedChallenges)
	protected.GET("/challenges/:id",
		middleware.Audit(auditCommandService, middleware.AuditOptions{
			Action:          model.AuditActionRead,
			ResourceType:    "challenge_detail",
			ResourceIDParam: "id",
		}, logger),
		challengeHandler.GetPublishedChallenge,
	)
	protected.POST("/challenges/:id/submit",
		middleware.Audit(auditCommandService, middleware.AuditOptions{
			Action:          model.AuditActionSubmit,
			ResourceType:    "challenge_submission",
			ResourceIDParam: "id",
		}, logger),
		practiceHandler.SubmitFlag,
	)
	protected.GET("/challenges/:id/submissions/mine", practiceHandler.ListMyChallengeSubmissions)
	protected.POST("/challenges/:id/instances", practiceHandler.StartChallenge)
	protected.POST("/instances/:id/access", runtimeHandler.AccessInstance)
	apiV1.GET("/instances/:id/proxy", runtimeHandler.ProxyInstance)
	apiV1.Any("/instances/:id/proxy/*proxyPath", runtimeHandler.ProxyInstance)
	usersGroup := protected.Group("/users")
	usersGroup.GET("/me/progress", practiceHandler.GetProgress)
	usersGroup.GET("/me/timeline", practiceHandler.GetTimeline)
	teacherGroup := protected.Group("/teacher")
	teacherGroup.Use(middleware.RequireRole(model.RoleTeacher, model.RoleAdmin))
	teacherGroup.GET("/students/:id/evidence", teachingQueryHandler.GetStudentEvidence)
	teacherGroup.GET("/students/:id/attack-sessions", teachingQueryHandler.GetStudentAttackSessions)

	t.Cleanup(func() {
		if sqlDB, sqlErr := db.DB(); sqlErr == nil {
			_ = sqlDB.Close()
		}
		_ = cache.Close()
		mini.Close()
	})

	return &flowTestEnv{
		router:  router,
		db:      db,
		cache:   cache,
		admin:   admin,
		student: student,
		image:   image,
	}
}

func newPracticeFlowTestConfig(t *testing.T) *config.Config {
	t.Helper()

	return &config.Config{
		App: config.AppConfig{
			Name: "ctf-platform-test",
			Env:  "test",
		},
		Auth: config.AuthConfig{
			SessionTTL:            24 * time.Hour,
			SessionCookieName:     "ctf_session",
			SessionCookiePath:     "/",
			SessionCookieHTTPOnly: true,
			SessionCookieSameSite: "lax",
			SessionKeyPrefix:      "test:session",
		},
		RateLimit: config.RateLimitConfig{
			RedisKeyPrefix: "test:rate_limit",
			FlagSubmit: config.RateLimitPolicyConfig{
				Enabled: true,
				Limit:   10,
				Window:  time.Minute,
			},
		},
		Challenge: config.ChallengeConfig{
			SolvedCountCacheTTL: time.Minute,
		},
		Cache: config.CacheConfig{
			ProgressTTL: time.Minute,
		},
		Container: config.ContainerConfig{
			FlagGlobalSecret:     "12345678901234567890123456789012",
			MaxConcurrentPerUser: 3,
			MaxExtends:           2,
			DefaultTTL:           2 * time.Hour,
			ExtendDuration:       30 * time.Minute,
			CreateTimeout:        5 * time.Second,
			PublicHost:           "127.0.0.1",
			DefaultExposedPort:   80,
			PortRangeStart:       30000,
			PortRangeEnd:         30100,
			ProxyTicketTTL:       15 * time.Minute,
			ProxyBodyPreviewSize: 1024,
		},
		Assessment: config.AssessmentConfig{
			IncrementalUpdateDelay:   10 * time.Millisecond,
			IncrementalUpdateTimeout: time.Second,
		},
		Pagination: config.PaginationConfig{
			DefaultPageSize: 20,
			MaxPageSize:     100,
		},
		WebSocket: config.WebSocketConfig{
			TicketTTL:         30 * time.Second,
			TicketKeyPrefix:   "test:ws:ticket",
			HeartbeatInterval: 100 * time.Millisecond,
			ReadTimeout:       time.Second,
			RetryInitialDelay: time.Second,
			RetryMaxDelay:     5 * time.Second,
		},
	}
}

func createFlowUser(t *testing.T, db *gorm.DB, username, password, role string) *model.User {
	t.Helper()

	user := &model.User{
		Username: username,
		Email:    fmt.Sprintf("%s@example.com", username),
		Role:     role,
		Status:   model.UserStatusActive,
	}
	setTestPassword(t, user, password)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func createFlowImage(t *testing.T, db *gorm.DB) *model.Image {
	t.Helper()

	image := &model.Image{
		Name:   "ctf/web-basic",
		Tag:    "v1",
		Status: model.ImageStatusAvailable,
	}
	if err := db.Create(image).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	return image
}

func loginForSession(t *testing.T, router http.Handler, username, password string) *http.Cookie {
	t.Helper()

	resp := performFlowJSONRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/auth/login",
		map[string]any{
			"username": username,
			"password": password,
		},
		nil,
		nil,
	)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected login status for %s: %d body=%s", username, resp.Code, resp.Body.String())
	}
	body := decodeFlowEnvelope(t, resp)
	_ = decodeFlowJSON[flowLoginResponse](t, body.Data)
	sessionCookie := cloneCookie(resp.Result().Cookies(), "ctf_session")
	if sessionCookie == nil {
		t.Fatalf("expected session cookie for %s", username)
	}
	return sessionCookie
}

func sessionHeaders(cookie *http.Cookie) map[string]string {
	if cookie == nil {
		return nil
	}
	return map[string]string{
		"Cookie": fmt.Sprintf("%s=%s", cookie.Name, cookie.Value),
	}
}

func loginForToken(t *testing.T, router http.Handler, username, password string) string {
	t.Helper()
	return loginForSession(t, router, username, password).Value
}

func bearerHeaders(token string) map[string]string {
	if token == "" {
		return nil
	}
	return map[string]string{
		"Cookie": "ctf_session=" + token,
	}
}

func cloneCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			cloned := *cookie
			return &cloned
		}
	}
	return nil
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

	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatalf("encode request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, target, &body)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func decodeFlowEnvelope(t *testing.T, recorder *httptest.ResponseRecorder) flowEnvelope {
	t.Helper()

	var envelope flowEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v body=%s", err, recorder.Body.String())
	}
	return envelope
}

func decodeFlowJSON[T any](t *testing.T, data []byte) T {
	t.Helper()

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("decode payload: %v payload=%s", err, string(data))
	}
	return value
}

func mustMarshalJSON(t *testing.T, value any) []byte {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return data
}

func assertTimelineHasSubmit(t *testing.T, events []struct {
	Type        string `json:"type"`
	ChallengeID int64  `json:"challenge_id"`
	Title       string `json:"title"`
	IsCorrect   *bool  `json:"is_correct"`
	Points      *int   `json:"points"`
	Detail      string `json:"detail"`
}, challengeID int64, isCorrect bool, points int) {
	t.Helper()

	for _, event := range events {
		if event.Type != "flag_submit" || event.ChallengeID != challengeID || event.IsCorrect == nil || *event.IsCorrect != isCorrect {
			continue
		}
		if isCorrect {
			if event.Points == nil || *event.Points != points {
				t.Fatalf("expected correct timeline event to include %d points, got %+v", points, event)
			}
		}
		if event.Detail == "" {
			t.Fatalf("expected timeline submit event to include detail, got %+v", event)
		}
		return
	}

	t.Fatalf("expected timeline to contain submit event challenge_id=%d is_correct=%t", challengeID, isCorrect)
}

func assertTeacherEvidenceHasEvent(
	t *testing.T,
	events []struct {
		Type        string                 `json:"type"`
		ChallengeID int64                  `json:"challenge_id"`
		Title       string                 `json:"title"`
		Detail      string                 `json:"detail"`
		Meta        map[string]interface{} `json:"meta"`
	},
	wantType string,
	challengeID int64,
	metaKey string,
	metaValue string,
) {
	t.Helper()
	for _, event := range events {
		if event.Type != wantType || event.ChallengeID != challengeID {
			continue
		}
		value, ok := event.Meta[metaKey]
		if !ok {
			t.Fatalf("expected evidence event %s to contain meta key %s: %+v", wantType, metaKey, event)
		}
		if value != metaValue {
			t.Fatalf("expected evidence event %s meta[%s]=%s, got %+v", wantType, metaKey, metaValue, event.Meta)
		}
		return
	}
	t.Fatalf("expected evidence to contain event type=%s challenge_id=%d", wantType, challengeID)
}

func assertTimelineHasChallengeDetailView(t *testing.T, events []struct {
	Type        string `json:"type"`
	ChallengeID int64  `json:"challenge_id"`
	Title       string `json:"title"`
	IsCorrect   *bool  `json:"is_correct"`
	Points      *int   `json:"points"`
	Detail      string `json:"detail"`
}, challengeID int64) {
	t.Helper()

	for _, event := range events {
		if event.Type != "challenge_detail_view" || event.ChallengeID != challengeID {
			continue
		}
		if event.Detail == "" {
			t.Fatalf("expected challenge detail view to include detail, got %+v", event)
		}
		return
	}

	t.Fatalf("expected timeline to contain challenge detail view event challenge_id=%d", challengeID)
}

func assertTimelineHasInstanceAccess(t *testing.T, events []struct {
	Type        string `json:"type"`
	ChallengeID int64  `json:"challenge_id"`
	Title       string `json:"title"`
	IsCorrect   *bool  `json:"is_correct"`
	Points      *int   `json:"points"`
	Detail      string `json:"detail"`
}, challengeID int64) {
	t.Helper()

	for _, event := range events {
		if event.Type != "instance_access" || event.ChallengeID != challengeID {
			continue
		}
		if event.Detail == "" {
			t.Fatalf("expected instance access event to include detail, got %+v", event)
		}
		return
	}

	t.Fatalf("expected timeline to contain instance access event challenge_id=%d", challengeID)
}

func assertTimelineHasProxyTrace(t *testing.T, events []struct {
	Type        string `json:"type"`
	ChallengeID int64  `json:"challenge_id"`
	Title       string `json:"title"`
	IsCorrect   *bool  `json:"is_correct"`
	Points      *int   `json:"points"`
	Detail      string `json:"detail"`
}, challengeID int64) {
	t.Helper()

	for _, event := range events {
		if event.Type != "instance_proxy_request" || event.ChallengeID != challengeID {
			continue
		}
		if !strings.Contains(event.Detail, "经平台代理发起") {
			t.Fatalf("expected proxy trace event detail, got %+v", event)
		}
		return
	}

	t.Fatalf("expected timeline to contain proxy trace event challenge_id=%d", challengeID)
}
