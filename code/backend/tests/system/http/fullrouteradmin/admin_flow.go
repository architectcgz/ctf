package fullrouteradmin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	challengehttp "ctf-platform/internal/module/challenge/api/http"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identityhttp "ctf-platform/internal/module/identity/api/http"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	opshttp "ctf-platform/internal/module/ops/api/http"
	opsqry "ctf-platform/internal/module/ops/application/queries"
	opsentity "ctf-platform/internal/module/ops/entity"
	practicecommands "ctf-platform/internal/module/practice/application/commands"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	"ctf-platform/internal/shared/taxonomy"
	xws "golang.org/x/net/websocket"
)

type RequestFunc func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder
type MultipartRequestFunc func(method, target, fieldName, fileName, content string, headers map[string]string) *httptest.ResponseRecorder

type AWDControlTarget struct {
	ContestID int64
	TeamID    int64
	ServiceID int64
}

type PublishRequestLifecycleTarget struct {
	ChallengeID int64
}

type ChallengeManagementDriver struct {
	Request                             RequestFunc
	AdminHeaders                        map[string]string
	TeacherHeaders                      map[string]string
	OtherTeacherHeaders                 map[string]string
	StudentHeaders                      map[string]string
	ImageID                             int64
	PracticeStudentID                   int64
	PracticeStudentUsername             string
	PublishChallenge                    func(t *testing.T, challengeID int64)
	CreatePracticeSubmission            func(t *testing.T, challengeID int64)
	SetPracticeStudentNo                func(t *testing.T, studentNo string)
	CreateDeleteBlockedChallenge        func(t *testing.T, title string) int64
	CreateRunningInstanceForDeleteBlock func(t *testing.T, challengeID int64)
	StopInstancesForChallenge           func(t *testing.T, challengeID int64)
}

type AdminOpsAndNotificationDriver struct {
	Request           RequestFunc
	MultipartRequest  MultipartRequestFunc
	Router            http.Handler
	AdminHeaders      map[string]string
	TeacherHeaders    map[string]string
	StudentHeaders    map[string]string
	PeerHeaders       map[string]string
	SeededImageID     int64
	NotificationID    int64
	SeedOnlineSession func(t *testing.T)
	SeedAuditLogs     func(t *testing.T)
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

func VerifyAdminChallengeManagementStateMatrix(t *testing.T, driver ChallengeManagementDriver) {
	t.Helper()

	resp := driver.Request(http.MethodPost, "/api/v1/authoring/challenges", map[string]any{
		"title":       "Lifecycle Challenge",
		"description": "challenge lifecycle matrix",
		"category":    taxonomy.DimensionWeb,
		"difficulty":  taxonomy.DifficultyEasy,
		"points":      120,
		"image_id":    driver.ImageID,
		"hints": []map[string]any{
			{
				"level":   1,
				"title":   "入口",
				"content": "look at login",
			},
			{
				"level":   2,
				"title":   "深入",
				"content": "check cookies",
			},
		},
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var createdChallenge challengehttp.ChallengeResp
	decodeEnvelopeData(t, resp, &createdChallenge)
	if createdChallenge.Status != challengecontracts.ChallengeStatusDraft || len(createdChallenge.Hints) != 2 {
		t.Fatalf("unexpected created challenge: %+v", createdChallenge)
	}

	resp = driver.Request(http.MethodPost, "/api/v1/authoring/challenges", map[string]any{
		"title":       "Invalid Hint Challenge",
		"description": "invalid hints",
		"category":    taxonomy.DimensionWeb,
		"difficulty":  taxonomy.DifficultyEasy,
		"points":      80,
		"image_id":    driver.ImageID,
		"hints": []map[string]any{
			{"level": 1, "content": "a"},
			{"level": 1, "content": "b"},
		},
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusBadRequest)

	emptyAttachment := ""
	resp = driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d", createdChallenge.ID), map[string]any{
		"title":          "Lifecycle Challenge Updated",
		"points":         150,
		"attachment_url": emptyAttachment,
		"hints": []map[string]any{
			{
				"level":   1,
				"title":   "更新提示",
				"content": "updated content",
			},
		},
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/authoring/challenges/%d", createdChallenge.ID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var updatedChallenge challengehttp.ChallengeResp
	decodeEnvelopeData(t, resp, &updatedChallenge)
	if updatedChallenge.Title != "Lifecycle Challenge Updated" || updatedChallenge.Points != 150 || len(updatedChallenge.Hints) != 1 {
		t.Fatalf("unexpected updated challenge: %+v", updatedChallenge)
	}

	resp = driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/flag", createdChallenge.ID), map[string]any{
		"flag_type":   challengecontracts.FlagTypeStatic,
		"flag":        "invalid-flag",
		"flag_prefix": "flag",
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusBadRequest)

	resp = driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/flag", createdChallenge.ID), map[string]any{
		"flag_type":   challengecontracts.FlagTypeStatic,
		"flag":        "flag{lifecycle-static}",
		"flag_prefix": "flag",
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/authoring/challenges/%d/flag", createdChallenge.ID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var staticFlag challengehttp.FlagResp
	decodeEnvelopeData(t, resp, &staticFlag)
	if !staticFlag.Configured || staticFlag.FlagType != challengecontracts.FlagTypeStatic {
		t.Fatalf("unexpected static flag config: %+v", staticFlag)
	}

	resp = driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/flag", createdChallenge.ID), map[string]any{
		"flag_type":   challengecontracts.FlagTypeDynamic,
		"flag_prefix": "ctf",
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/authoring/challenges/%d/flag", createdChallenge.ID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var dynamicFlag challengehttp.FlagResp
	decodeEnvelopeData(t, resp, &dynamicFlag)
	if !dynamicFlag.Configured || dynamicFlag.FlagType != challengecontracts.FlagTypeDynamic || dynamicFlag.FlagPrefix != "ctf" {
		t.Fatalf("unexpected dynamic flag config: %+v", dynamicFlag)
	}

	resp = driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup", createdChallenge.ID), map[string]any{
		"title":      "Scheduled Writeup",
		"content":    "scheduled content",
		"visibility": "scheduled",
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusBadRequest)

	resp = driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup", createdChallenge.ID), map[string]any{
		"title":      "Scheduled Writeup",
		"content":    "scheduled content",
		"visibility": "scheduled",
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusBadRequest)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup", createdChallenge.ID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusNotFound)

	driver.PublishChallenge(t, createdChallenge.ID)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d", createdChallenge.ID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var publishedDetail challengehttp.ChallengeDetailResp
	decodeEnvelopeData(t, resp, &publishedDetail)
	if publishedDetail.ID != createdChallenge.ID {
		t.Fatalf("unexpected published detail: %+v", publishedDetail)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d/writeup", createdChallenge.ID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusNotFound)

	resp = driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup", createdChallenge.ID), map[string]any{
		"title":      "Public Writeup",
		"content":    "public content",
		"visibility": challengeentity.WriteupVisibilityPublic,
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d/writeup", createdChallenge.ID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var publicWriteup challengecontracts.ChallengeWriteupResp
	decodeEnvelopeData(t, resp, &publicWriteup)
	if publicWriteup.Visibility != challengeentity.WriteupVisibilityPublic {
		t.Fatalf("unexpected public writeup visibility: %+v", publicWriteup)
	}
	if !publicWriteup.RequiresSpoilerWarning {
		t.Fatalf("expected spoiler warning before solving, got %+v", publicWriteup)
	}

	driver.CreatePracticeSubmission(t, createdChallenge.ID)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d/writeup", createdChallenge.ID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var solvedWriteup challengecontracts.ChallengeWriteupResp
	decodeEnvelopeData(t, resp, &solvedWriteup)
	if solvedWriteup.RequiresSpoilerWarning {
		t.Fatalf("expected spoiler warning to clear after solve, got %+v", solvedWriteup)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d/writeup-submissions/me", createdChallenge.ID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var emptyWriteupEnvelope envelope
	if err := json.Unmarshal(resp.Body.Bytes(), &emptyWriteupEnvelope); err != nil {
		t.Fatalf("decode empty writeup envelope: %v body=%s", err, resp.Body.String())
	}
	if string(emptyWriteupEnvelope.Data) != "null" {
		t.Fatalf("expected empty writeup submission before upsert")
	}

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/challenges/%d/writeup-submissions", createdChallenge.ID), map[string]any{
		"title":             "首版草稿",
		"content":           "先记录思路，再整理利用链。",
		"submission_status": challengeentity.SubmissionWriteupStatusDraft,
	}, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var draftSubmission challengecontracts.SubmissionWriteupResp
	decodeEnvelopeData(t, resp, &draftSubmission)
	if draftSubmission.SubmissionStatus != challengeentity.SubmissionWriteupStatusDraft || draftSubmission.PublishedAt != nil {
		t.Fatalf("unexpected draft submission response: %+v", draftSubmission)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d/writeup-submissions/me", createdChallenge.ID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var mySubmission challengecontracts.SubmissionWriteupResp
	decodeEnvelopeData(t, resp, &mySubmission)
	if mySubmission.Title != "首版草稿" {
		t.Fatalf("unexpected my submission payload: %+v", mySubmission)
	}

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/challenges/%d/writeup-submissions", createdChallenge.ID), map[string]any{
		"title":             "正式复盘",
		"content":           "1. 判断输入点\n2. 构造 payload\n3. 读取 flag",
		"submission_status": challengeentity.SubmissionWriteupStatusSubmitted,
	}, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var submittedWriteup challengecontracts.SubmissionWriteupResp
	decodeEnvelopeData(t, resp, &submittedWriteup)
	if submittedWriteup.SubmissionStatus != challengeentity.SubmissionWriteupStatusPublished || submittedWriteup.PublishedAt == nil {
		t.Fatalf("unexpected submitted writeup response: %+v", submittedWriteup)
	}

	peerStudentNo := "20240001"
	driver.SetPracticeStudentNo(t, peerStudentNo)

	resp = driver.Request(
		http.MethodGet,
		fmt.Sprintf("/api/v1/teacher/writeup-submissions?student_id=%d&challenge_id=%d", driver.PracticeStudentID, createdChallenge.ID),
		nil,
		driver.TeacherHeaders,
	)
	assertStatus(t, resp, http.StatusOK)

	var teacherSubmissionList struct {
		List     []challengecontracts.TeacherSubmissionWriteupItemResp `json:"list"`
		Total    int64                                                 `json:"total"`
		Page     int                                                   `json:"page"`
		PageSize int                                                   `json:"page_size"`
	}
	decodeEnvelopeData(t, resp, &teacherSubmissionList)
	if teacherSubmissionList.Total != 1 || len(teacherSubmissionList.List) != 1 {
		t.Fatalf("unexpected teacher submission list: %+v", teacherSubmissionList)
	}
	if teacherSubmissionList.List[0].StudentUsername != driver.PracticeStudentUsername ||
		teacherSubmissionList.List[0].StudentNo != peerStudentNo ||
		teacherSubmissionList.List[0].ChallengeID != createdChallenge.ID {
		t.Fatalf("unexpected teacher submission list item: %+v", teacherSubmissionList.List[0])
	}

	resp = driver.Request(
		http.MethodGet,
		fmt.Sprintf("/api/v1/teacher/writeup-submissions?student_id=%d", driver.PracticeStudentID),
		nil,
		driver.OtherTeacherHeaders,
	)
	assertStatus(t, resp, http.StatusOK)

	var inaccessibleList struct {
		List  []challengecontracts.TeacherSubmissionWriteupItemResp `json:"list"`
		Total int64                                                 `json:"total"`
	}
	decodeEnvelopeData(t, resp, &inaccessibleList)
	if inaccessibleList.Total != 0 || len(inaccessibleList.List) != 0 {
		t.Fatalf("expected other teacher to see empty submission list, got %+v", inaccessibleList)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/teacher/writeup-submissions/%d", submittedWriteup.ID), nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	var teacherSubmissionDetail challengecontracts.TeacherSubmissionWriteupDetailResp
	decodeEnvelopeData(t, resp, &teacherSubmissionDetail)
	if teacherSubmissionDetail.StudentUsername != driver.PracticeStudentUsername ||
		teacherSubmissionDetail.StudentNo != peerStudentNo ||
		teacherSubmissionDetail.Content == "" {
		t.Fatalf("unexpected teacher submission detail: %+v", teacherSubmissionDetail)
	}

	resp = driver.Request(http.MethodPost, "/api/v1/authoring/challenges", map[string]any{
		"title":       "Manual Review Challenge",
		"description": "submit an answer for teacher review",
		"category":    taxonomy.DimensionMisc,
		"difficulty":  taxonomy.DifficultyMedium,
		"points":      120,
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var manualChallenge challengehttp.ChallengeResp
	decodeEnvelopeData(t, resp, &manualChallenge)

	resp = driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/flag", manualChallenge.ID), map[string]any{
		"flag_type": challengecontracts.FlagTypeManualReview,
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	driver.PublishChallenge(t, manualChallenge.ID)

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/challenges/%d/submit", manualChallenge.ID), map[string]any{
		"flag": "exploit trace and reasoning",
	}, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var pendingManualSubmission practicecommands.SubmissionResp
	decodeEnvelopeData(t, resp, &pendingManualSubmission)
	if pendingManualSubmission.Status != practicecommands.SubmissionStatusPendingReview || pendingManualSubmission.IsCorrect {
		t.Fatalf("unexpected pending manual review response: %+v", pendingManualSubmission)
	}

	resp = driver.Request(
		http.MethodGet,
		fmt.Sprintf("/api/v1/teacher/manual-review-submissions?student_id=%d&challenge_id=%d", driver.PracticeStudentID, manualChallenge.ID),
		nil,
		driver.TeacherHeaders,
	)
	assertStatus(t, resp, http.StatusOK)

	var manualReviewList struct {
		List     []practicecontracts.TeacherManualReviewSubmissionItemResp `json:"list"`
		Total    int64                                                     `json:"total"`
		Page     int                                                       `json:"page"`
		PageSize int                                                       `json:"page_size"`
	}
	decodeEnvelopeData(t, resp, &manualReviewList)
	if manualReviewList.Total != 1 || len(manualReviewList.List) != 1 {
		t.Fatalf("unexpected manual review list: %+v", manualReviewList)
	}
	if manualReviewList.List[0].ChallengeID != manualChallenge.ID || manualReviewList.List[0].StudentUsername != driver.PracticeStudentUsername {
		t.Fatalf("unexpected manual review list item: %+v", manualReviewList.List[0])
	}

	resp = driver.Request(
		http.MethodGet,
		fmt.Sprintf("/api/v1/teacher/manual-review-submissions?student_id=%d", driver.PracticeStudentID),
		nil,
		driver.OtherTeacherHeaders,
	)
	assertStatus(t, resp, http.StatusOK)

	var otherTeacherManualList struct {
		List  []practicecontracts.TeacherManualReviewSubmissionItemResp `json:"list"`
		Total int64                                                     `json:"total"`
	}
	decodeEnvelopeData(t, resp, &otherTeacherManualList)
	if otherTeacherManualList.Total != 0 || len(otherTeacherManualList.List) != 0 {
		t.Fatalf("expected other teacher to see empty manual review list, got %+v", otherTeacherManualList)
	}

	manualSubmissionID := manualReviewList.List[0].ID

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/teacher/manual-review-submissions/%d", manualSubmissionID), nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	var manualReviewDetail practicecontracts.TeacherManualReviewSubmissionDetailResp
	decodeEnvelopeData(t, resp, &manualReviewDetail)
	if manualReviewDetail.Answer == "" || manualReviewDetail.ChallengeID != manualChallenge.ID {
		t.Fatalf("unexpected manual review detail: %+v", manualReviewDetail)
	}

	resp = driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/teacher/manual-review-submissions/%d/review", manualSubmissionID), map[string]any{
		"review_status":  contestcontracts.SubmissionReviewStatusApproved,
		"review_comment": "证据完整，通过。",
	}, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	var reviewedManualSubmission practicecontracts.TeacherManualReviewSubmissionDetailResp
	decodeEnvelopeData(t, resp, &reviewedManualSubmission)
	if reviewedManualSubmission.ReviewStatus != contestcontracts.SubmissionReviewStatusApproved || !reviewedManualSubmission.IsCorrect || reviewedManualSubmission.Score != 120 {
		t.Fatalf("unexpected reviewed manual submission: %+v", reviewedManualSubmission)
	}

	resp = driver.Request(http.MethodPost, "/api/v1/authoring/environment-templates", map[string]any{
		"name":           "Lifecycle Template",
		"description":    "template for lifecycle test",
		"entry_node_key": "web",
		"networks": []map[string]any{
			{"key": "default", "name": "Default"},
		},
		"nodes": []map[string]any{
			{
				"key":          "web",
				"name":         "Web",
				"image_id":     driver.ImageID,
				"service_port": 8080,
				"tier":         challengecontracts.TopologyTierPublic,
				"network_keys": []string{"default"},
			},
		},
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var template challengehttp.EnvironmentTemplateResp
	decodeEnvelopeData(t, resp, &template)
	if template.EntryNodeKey != "web" {
		t.Fatalf("unexpected template: %+v", template)
	}

	resp = driver.Request(http.MethodGet, "/api/v1/authoring/environment-templates?keyword=Lifecycle", nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var templates []challengehttp.EnvironmentTemplateResp
	decodeEnvelopeData(t, resp, &templates)
	if len(templates) == 0 {
		t.Fatalf("expected template list to include created template")
	}

	resp = driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/topology", createdChallenge.ID), map[string]any{
		"template_id": template.ID,
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var topology challengehttp.ChallengeTopologyResp
	decodeEnvelopeData(t, resp, &topology)
	if topology.TemplateID == nil || *topology.TemplateID != template.ID {
		t.Fatalf("unexpected topology template binding: %+v", topology)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/authoring/challenges/%d/topology", createdChallenge.ID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/authoring/environment-templates/%d", template.ID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var loadedTemplate challengehttp.EnvironmentTemplateResp
	decodeEnvelopeData(t, resp, &loadedTemplate)
	if loadedTemplate.UsageCount < 1 {
		t.Fatalf("expected template usage count increment, got %+v", loadedTemplate)
	}

	resp = driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/topology", createdChallenge.ID), map[string]any{
		"entry_node_key": "ghost",
		"nodes": []map[string]any{
			{
				"key":          "web",
				"name":         "Web",
				"image_id":     driver.ImageID,
				"service_port": 8080,
			},
		},
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusBadRequest)

	resp = driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/authoring/environment-templates/%d", template.ID), map[string]any{
		"name":           "Lifecycle Template Updated",
		"description":    "updated template",
		"entry_node_key": "web",
		"nodes": []map[string]any{
			{
				"key":          "web",
				"name":         "Web Updated",
				"image_id":     driver.ImageID,
				"service_port": 9090,
			},
		},
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodDelete, fmt.Sprintf("/api/v1/authoring/challenges/%d/topology", createdChallenge.ID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/authoring/challenges/%d/topology", createdChallenge.ID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusNotFound)

	resp = driver.Request(http.MethodDelete, fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup", createdChallenge.ID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup", createdChallenge.ID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusNotFound)

	instanceChallengeID := driver.CreateDeleteBlockedChallenge(t, "DeleteBlocked Challenge")
	driver.CreateRunningInstanceForDeleteBlock(t, instanceChallengeID)

	resp = driver.Request(http.MethodDelete, fmt.Sprintf("/api/v1/authoring/challenges/%d", instanceChallengeID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusConflict)

	driver.StopInstancesForChallenge(t, instanceChallengeID)

	resp = driver.Request(http.MethodDelete, fmt.Sprintf("/api/v1/authoring/challenges/%d", instanceChallengeID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodDelete, fmt.Sprintf("/api/v1/authoring/challenges/%d", createdChallenge.ID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/authoring/challenges/%d", createdChallenge.ID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusNotFound)

	resp = driver.Request(http.MethodDelete, fmt.Sprintf("/api/v1/authoring/environment-templates/%d", template.ID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)
}

func VerifyAdminOpsAndNotificationStateMatrix(t *testing.T, driver AdminOpsAndNotificationDriver) {
	t.Helper()

	resp := driver.Request(http.MethodPost, "/api/v1/authoring/images", map[string]any{
		"name":        "matrix/webapp",
		"tag":         "v2",
		"description": "integration image",
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var freeImage challengehttp.ImageResp
	decodeEnvelopeData(t, resp, &freeImage)
	if freeImage.Name != "matrix/webapp" {
		t.Fatalf("unexpected created image: %+v", freeImage)
	}

	resp = driver.Request(http.MethodPost, "/api/v1/authoring/images", map[string]any{
		"name":        "matrix/webapp",
		"tag":         "v2",
		"description": "duplicate image",
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusConflict)

	resp = driver.Request(http.MethodGet, "/api/v1/authoring/images?name=matrix/status=available", nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/authoring/images/%d", freeImage.ID), map[string]any{
		"description": "updated image",
		"status":      "failed",
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/authoring/images/%d", freeImage.ID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var loadedImage challengehttp.ImageResp
	decodeEnvelopeData(t, resp, &loadedImage)
	if loadedImage.Status != "failed" || loadedImage.Description != "updated image" {
		t.Fatalf("unexpected loaded image: %+v", loadedImage)
	}

	resp = driver.Request(http.MethodDelete, fmt.Sprintf("/api/v1/authoring/images/%d", driver.SeededImageID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusConflict)

	resp = driver.Request(http.MethodDelete, fmt.Sprintf("/api/v1/authoring/images/%d", freeImage.ID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodGet, "/api/v1/admin/users?role=student&class_name=ClassA", nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var userPage map[string]any
	decodeEnvelopeData(t, resp, &userPage)
	if int(userPage["total"].(float64)) < 2 {
		t.Fatalf("expected student page results, got %+v", userPage)
	}

	resp = driver.Request(http.MethodPost, "/api/v1/admin/users", map[string]any{
		"username":   "admin_created_student",
		"name":       "Created Student",
		"password":   "Password123",
		"email":      "created_student@example.com",
		"student_no": "20260001",
		"class_name": "ClassA",
		"role":       identitycontracts.RoleStudent,
		"status":     identitycontracts.UserStatusActive,
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var createdUserWrap map[string]json.RawMessage
	decodeEnvelopeData(t, resp, &createdUserWrap)

	var createdUser identityhttp.AdminUserResp
	if err := json.Unmarshal(createdUserWrap["user"], &createdUser); err != nil {
		t.Fatalf("decode created user: %v", err)
	}
	if createdUser.Username != "admin_created_student" {
		t.Fatalf("unexpected created user: %+v", createdUser)
	}

	resp = driver.Request(http.MethodPost, "/api/v1/admin/users", map[string]any{
		"username": "admin_created_student",
		"password": "Password123",
		"role":     identitycontracts.RoleStudent,
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusConflict)

	updatedTeacherNo := "T-9001"
	updatedRole := identitycontracts.RoleTeacher
	resp = driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%d", createdUser.ID), map[string]any{
		"role":       updatedRole,
		"teacher_no": updatedTeacherNo,
		"class_name": "ClassTeach",
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var updatedUserWrap map[string]json.RawMessage
	decodeEnvelopeData(t, resp, &updatedUserWrap)

	var updatedUser identityhttp.AdminUserResp
	if err := json.Unmarshal(updatedUserWrap["user"], &updatedUser); err != nil {
		t.Fatalf("decode updated user: %v", err)
	}
	if updatedUser.TeacherNo == nil || *updatedUser.TeacherNo != updatedTeacherNo || updatedUser.StudentNo != nil {
		t.Fatalf("unexpected updated user: %+v", updatedUser)
	}

	csvContent := strings.Join([]string{
		"username,password,email,class_name,role,status,student_no,teacher_no,name",
		"import_new,Password123,import_new@example.com,ClassA,student,active,20260002,,Import New",
		"admin_created_student,,updated_import@example.com,ClassTeach,teacher,inactive,,T-9002,Imported Update",
		",Password123,bad@example.com,ClassA,student,active,20260003,,Bad Row",
	}, "\n")
	resp = driver.MultipartRequest(http.MethodPost, "/api/v1/admin/users/import", "file", "users.csv", csvContent, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusCreated)

	var importResult identityhttp.ImportUsersResp
	decodeEnvelopeData(t, resp, &importResult)
	if importResult.Created != 1 || importResult.Updated != 1 || importResult.Failed != 1 {
		t.Fatalf("unexpected import result: %+v", importResult)
	}

	resp = driver.Request(http.MethodDelete, fmt.Sprintf("/api/v1/admin/users/%d", createdUser.ID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	driver.SeedOnlineSession(t)

	resp = driver.Request(http.MethodGet, "/api/v1/admin/dashboard", nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var dashboard opshttp.DashboardStats
	decodeEnvelopeData(t, resp, &dashboard)
	if dashboard.OnlineUsers < 1 || dashboard.ActiveContainers < 1 {
		t.Fatalf("unexpected dashboard stats: %+v", dashboard)
	}

	driver.SeedAuditLogs(t)

	resp = driver.Request(http.MethodGet, "/api/v1/admin/audit-logs?action=submit&page=1&page_size=10", nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var auditPage map[string]any
	decodeEnvelopeData(t, resp, &auditPage)
	if int(auditPage["total"].(float64)) < 5 {
		t.Fatalf("unexpected audit page: %+v", auditPage)
	}

	resp = driver.Request(http.MethodGet, "/api/v1/admin/audit-logs?start_time=bad-time", nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusBadRequest)

	resp = driver.Request(http.MethodGet, "/api/v1/admin/cheat-detection", nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var cheat opsqry.CheatDetectionResp
	decodeEnvelopeData(t, resp, &cheat)
	if cheat.Summary.SubmitBurstUsers < 1 || cheat.Summary.SharedIPGroups < 1 {
		t.Fatalf("unexpected cheat detection response: %+v", cheat)
	}

	resp = driver.Request(http.MethodPost, "/api/v1/admin/notifications", map[string]any{
		"type":    opsentity.NotificationTypeSystem,
		"title":   "全员通知",
		"content": "full-router matrix admin publish",
		"audience_rules": map[string]any{
			"mode": "union",
			"rules": []map[string]any{
				{"type": "all"},
			},
		},
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var publishResult opshttp.AdminNotificationPublishResp
	decodeEnvelopeData(t, resp, &publishResult)
	if publishResult.BatchID <= 0 || publishResult.RecipientCount < 4 {
		t.Fatalf("unexpected publish result: %+v", publishResult)
	}

	resp = driver.Request(http.MethodPost, "/api/v1/admin/notifications", map[string]any{
		"type":    opsentity.NotificationTypeSystem,
		"title":   "teacher forbidden",
		"content": "teacher should not publish",
		"audience_rules": map[string]any{
			"mode": "union",
			"rules": []map[string]any{
				{"type": "all"},
			},
		},
	}, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	resp = driver.Request(http.MethodPost, "/api/v1/admin/notifications", map[string]any{
		"type":    opsentity.NotificationTypeSystem,
		"title":   "invalid audience",
		"content": "missing roles",
		"audience_rules": map[string]any{
			"mode": "union",
			"rules": []map[string]any{
				{"type": "role"},
			},
		},
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusBadRequest)

	resp = driver.Request(http.MethodGet, "/api/v1/notifications?page=1&page_size=10", nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var notificationPage map[string]any
	decodeEnvelopeData(t, resp, &notificationPage)
	if int(notificationPage["total"].(float64)) < 2 {
		t.Fatalf("unexpected notifications page: %+v", notificationPage)
	}

	resp = driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/notifications/%d/read", driver.NotificationID), nil, driver.PeerHeaders)
	assertStatus(t, resp, http.StatusNotFound)

	resp = driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/notifications/%d/read", driver.NotificationID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	server := httptest.NewServer(driver.Router)
	defer server.Close()

	ticketResp := driver.Request(http.MethodPost, "/api/v1/auth/ws-ticket", nil, driver.StudentHeaders)
	assertStatus(t, ticketResp, http.StatusOK)

	var wsTicket map[string]any
	decodeEnvelopeData(t, ticketResp, &wsTicket)
	ticket, _ := wsTicket["ticket"].(string)
	if ticket == "" {
		t.Fatalf("expected websocket ticket")
	}

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/notifications?ticket=" + ticket
	wsConfig, err := xws.NewConfig(wsURL, server.URL)
	if err != nil {
		t.Fatalf("new websocket config: %v", err)
	}
	conn, err := xws.DialConfig(wsConfig)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	message := receiveWSMessageByType(t, conn, "system.connected")
	if message.Type != "system.connected" {
		t.Fatalf("unexpected websocket message: %+v", message)
	}

	reusedConfig, _ := xws.NewConfig(wsURL, server.URL)
	if _, err := xws.DialConfig(reusedConfig); err == nil {
		t.Fatal("expected consumed websocket ticket to be rejected")
	}
}

type wsEnvelope struct {
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

func receiveWSMessageByType(t *testing.T, conn *xws.Conn, expectedType string) wsEnvelope {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)
	if err := conn.SetDeadline(deadline); err != nil {
		t.Fatalf("set websocket deadline: %v", err)
	}
	for {
		var message wsEnvelope
		if err := xws.JSON.Receive(conn, &message); err != nil {
			t.Fatalf("receive websocket message: %v", err)
		}
		if message.Type == expectedType {
			return message
		}
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
