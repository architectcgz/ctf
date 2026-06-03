package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ctf-platform/internal/auditlog"
	challengehttp "ctf-platform/internal/module/challenge/api/http"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identityhttp "ctf-platform/internal/module/identity/api/http"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	opshttp "ctf-platform/internal/module/ops/api/http"
	opsqry "ctf-platform/internal/module/ops/application/queries"
	opsentity "ctf-platform/internal/module/ops/entity"
	practicecmd "ctf-platform/internal/module/practice/application/commands"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	"ctf-platform/internal/shared/taxonomy"
	xws "golang.org/x/net/websocket"
)

func TestFullRouter_AdminChallengeManagementStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	adminHeaders := bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd))
	teacherHeaders := bearerHeaders(loginForToken(t, env.router, env.teacher.Username, env.teacherPwd))
	otherTeacherHeaders := bearerHeaders(loginForToken(t, env.router, env.otherTeacher.Username, "Password123"))
	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.peerStudent.Username, "Password123"))

	resp := performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/authoring/challenges", map[string]any{
		"title":       "Lifecycle Challenge",
		"description": "challenge lifecycle matrix",
		"category":    taxonomy.DimensionWeb,
		"difficulty":  taxonomy.DifficultyEasy,
		"points":      120,
		"image_id":    env.image.ID,
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
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var createdChallenge challengehttp.ChallengeResp
	decodeFullRouterData(t, resp, &createdChallenge)
	if createdChallenge.Status != challengecontracts.ChallengeStatusDraft || len(createdChallenge.Hints) != 2 {
		t.Fatalf("unexpected created challenge: %+v", createdChallenge)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/authoring/challenges", map[string]any{
		"title":       "Invalid Hint Challenge",
		"description": "invalid hints",
		"category":    taxonomy.DimensionWeb,
		"difficulty":  taxonomy.DifficultyEasy,
		"points":      80,
		"image_id":    env.image.ID,
		"hints": []map[string]any{
			{"level": 1, "content": "a"},
			{"level": 1, "content": "b"},
		},
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusBadRequest)

	emptyAttachment := ""
	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d", createdChallenge.ID), map[string]any{
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
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/authoring/challenges/%d", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var updatedChallenge challengehttp.ChallengeResp
	decodeFullRouterData(t, resp, &updatedChallenge)
	if updatedChallenge.Title != "Lifecycle Challenge Updated" || updatedChallenge.Points != 150 || len(updatedChallenge.Hints) != 1 {
		t.Fatalf("unexpected updated challenge: %+v", updatedChallenge)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/flag", createdChallenge.ID), map[string]any{
		"flag_type":   challengecontracts.FlagTypeStatic,
		"flag":        "invalid-flag",
		"flag_prefix": "flag",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusBadRequest)

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/flag", createdChallenge.ID), map[string]any{
		"flag_type":   challengecontracts.FlagTypeStatic,
		"flag":        "flag{lifecycle-static}",
		"flag_prefix": "flag",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/authoring/challenges/%d/flag", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var staticFlag challengehttp.FlagResp
	decodeFullRouterData(t, resp, &staticFlag)
	if !staticFlag.Configured || staticFlag.FlagType != challengecontracts.FlagTypeStatic {
		t.Fatalf("unexpected static flag config: %+v", staticFlag)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/flag", createdChallenge.ID), map[string]any{
		"flag_type":   challengecontracts.FlagTypeDynamic,
		"flag_prefix": "ctf",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/authoring/challenges/%d/flag", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var dynamicFlag challengehttp.FlagResp
	decodeFullRouterData(t, resp, &dynamicFlag)
	if !dynamicFlag.Configured || dynamicFlag.FlagType != challengecontracts.FlagTypeDynamic || dynamicFlag.FlagPrefix != "ctf" {
		t.Fatalf("unexpected dynamic flag config: %+v", dynamicFlag)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup", createdChallenge.ID), map[string]any{
		"title":      "Scheduled Writeup",
		"content":    "scheduled content",
		"visibility": "scheduled",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusBadRequest)

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup", createdChallenge.ID), map[string]any{
		"title":      "Scheduled Writeup",
		"content":    "scheduled content",
		"visibility": "scheduled",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusBadRequest)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusNotFound)

	if err := env.db.Model(&appChallengeRow{}).
		Where("id = ?", createdChallenge.ID).
		Update("status", challengecontracts.ChallengeStatusPublished).Error; err != nil {
		t.Fatalf("set created challenge published: %v", err)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d", createdChallenge.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var publishedDetail challengehttp.ChallengeDetailResp
	decodeFullRouterData(t, resp, &publishedDetail)
	if publishedDetail.ID != createdChallenge.ID {
		t.Fatalf("unexpected published detail: %+v", publishedDetail)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d/writeup", createdChallenge.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusNotFound)

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup", createdChallenge.ID), map[string]any{
		"title":      "Public Writeup",
		"content":    "public content",
		"visibility": challengeentity.WriteupVisibilityPublic,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d/writeup", createdChallenge.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var publicWriteup challengecontracts.ChallengeWriteupResp
	decodeFullRouterData(t, resp, &publicWriteup)
	if publicWriteup.Visibility != challengeentity.WriteupVisibilityPublic {
		t.Fatalf("unexpected public writeup visibility: %+v", publicWriteup)
	}
	if !publicWriteup.RequiresSpoilerWarning {
		t.Fatalf("expected spoiler warning before solving, got %+v", publicWriteup)
	}

	createPracticeSubmission(t, env, env.peerStudent.ID, createdChallenge.ID, 150)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d/writeup", createdChallenge.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var solvedWriteup challengecontracts.ChallengeWriteupResp
	decodeFullRouterData(t, resp, &solvedWriteup)
	if solvedWriteup.RequiresSpoilerWarning {
		t.Fatalf("expected spoiler warning to clear after solve, got %+v", solvedWriteup)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d/writeup-submissions/me", createdChallenge.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
	var emptyWriteupEnvelope fullRouterEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &emptyWriteupEnvelope); err != nil {
		t.Fatalf("decode empty writeup envelope: %v body=%s", err, resp.Body.String())
	}
	if string(emptyWriteupEnvelope.Data) != "null" {
		t.Fatalf("expected empty writeup submission before upsert")
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/challenges/%d/writeup-submissions", createdChallenge.ID), map[string]any{
		"title":             "首版草稿",
		"content":           "先记录思路，再整理利用链。",
		"submission_status": challengeentity.SubmissionWriteupStatusDraft,
	}, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var draftSubmission challengecontracts.SubmissionWriteupResp
	decodeFullRouterData(t, resp, &draftSubmission)
	if draftSubmission.SubmissionStatus != challengeentity.SubmissionWriteupStatusDraft || draftSubmission.PublishedAt != nil {
		t.Fatalf("unexpected draft submission response: %+v", draftSubmission)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d/writeup-submissions/me", createdChallenge.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var mySubmission challengecontracts.SubmissionWriteupResp
	decodeFullRouterData(t, resp, &mySubmission)
	if mySubmission.Title != "首版草稿" {
		t.Fatalf("unexpected my submission payload: %+v", mySubmission)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/challenges/%d/writeup-submissions", createdChallenge.ID), map[string]any{
		"title":             "正式复盘",
		"content":           "1. 判断输入点\n2. 构造 payload\n3. 读取 flag",
		"submission_status": challengeentity.SubmissionWriteupStatusSubmitted,
	}, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var submittedWriteup challengecontracts.SubmissionWriteupResp
	decodeFullRouterData(t, resp, &submittedWriteup)
	if submittedWriteup.SubmissionStatus != challengeentity.SubmissionWriteupStatusPublished || submittedWriteup.PublishedAt == nil {
		t.Fatalf("unexpected submitted writeup response: %+v", submittedWriteup)
	}
	peerStudentNo := "20240001"
	if err := env.db.Model(&identitycontracts.User{}).Where("id = ?", env.peerStudent.ID).Update("student_no", peerStudentNo).Error; err != nil {
		t.Fatalf("set peer student number: %v", err)
	}
	env.peerStudent.StudentNo = peerStudentNo

	resp = performFullRouterRequest(
		t,
		env.router,
		http.MethodGet,
		fmt.Sprintf("/api/v1/teacher/writeup-submissions?student_id=%d&challenge_id=%d", env.peerStudent.ID, createdChallenge.ID),
		nil,
		teacherHeaders,
	)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var teacherSubmissionList struct {
		List     []challengecontracts.TeacherSubmissionWriteupItemResp `json:"list"`
		Total    int64                                                 `json:"total"`
		Page     int                                                   `json:"page"`
		PageSize int                                                   `json:"page_size"`
	}
	decodeFullRouterData(t, resp, &teacherSubmissionList)
	if teacherSubmissionList.Total != 1 || len(teacherSubmissionList.List) != 1 {
		t.Fatalf("unexpected teacher submission list: %+v", teacherSubmissionList)
	}
	if teacherSubmissionList.List[0].StudentUsername != env.peerStudent.Username ||
		teacherSubmissionList.List[0].StudentNo != peerStudentNo ||
		teacherSubmissionList.List[0].ChallengeID != createdChallenge.ID {
		t.Fatalf("unexpected teacher submission list item: %+v", teacherSubmissionList.List[0])
	}

	resp = performFullRouterRequest(
		t,
		env.router,
		http.MethodGet,
		fmt.Sprintf("/api/v1/teacher/writeup-submissions?student_id=%d", env.peerStudent.ID),
		nil,
		otherTeacherHeaders,
	)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var inaccessibleList struct {
		List  []challengecontracts.TeacherSubmissionWriteupItemResp `json:"list"`
		Total int64                                                 `json:"total"`
	}
	decodeFullRouterData(t, resp, &inaccessibleList)
	if inaccessibleList.Total != 0 || len(inaccessibleList.List) != 0 {
		t.Fatalf("expected other teacher to see empty submission list, got %+v", inaccessibleList)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/writeup-submissions/%d", submittedWriteup.ID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var teacherSubmissionDetail challengecontracts.TeacherSubmissionWriteupDetailResp
	decodeFullRouterData(t, resp, &teacherSubmissionDetail)
	if teacherSubmissionDetail.StudentUsername != env.peerStudent.Username ||
		teacherSubmissionDetail.StudentNo != peerStudentNo ||
		teacherSubmissionDetail.Content == "" {
		t.Fatalf("unexpected teacher submission detail: %+v", teacherSubmissionDetail)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/authoring/challenges", map[string]any{
		"title":       "Manual Review Challenge",
		"description": "submit an answer for teacher review",
		"category":    taxonomy.DimensionMisc,
		"difficulty":  taxonomy.DifficultyMedium,
		"points":      120,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var manualChallenge challengehttp.ChallengeResp
	decodeFullRouterData(t, resp, &manualChallenge)

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/flag", manualChallenge.ID), map[string]any{
		"flag_type": challengecontracts.FlagTypeManualReview,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	if err := env.db.Model(&appChallengeRow{}).
		Where("id = ?", manualChallenge.ID).
		Update("status", challengecontracts.ChallengeStatusPublished).Error; err != nil {
		t.Fatalf("set manual challenge published: %v", err)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/challenges/%d/submit", manualChallenge.ID), map[string]any{
		"flag": "exploit trace and reasoning",
	}, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var pendingManualSubmission practicecmd.SubmissionResp
	decodeFullRouterData(t, resp, &pendingManualSubmission)
	if pendingManualSubmission.Status != practicecmd.SubmissionStatusPendingReview || pendingManualSubmission.IsCorrect {
		t.Fatalf("unexpected pending manual review response: %+v", pendingManualSubmission)
	}

	resp = performFullRouterRequest(
		t,
		env.router,
		http.MethodGet,
		fmt.Sprintf("/api/v1/teacher/manual-review-submissions?student_id=%d&challenge_id=%d", env.peerStudent.ID, manualChallenge.ID),
		nil,
		teacherHeaders,
	)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var manualReviewList struct {
		List     []practicecontracts.TeacherManualReviewSubmissionItemResp `json:"list"`
		Total    int64                                                     `json:"total"`
		Page     int                                                       `json:"page"`
		PageSize int                                                       `json:"page_size"`
	}
	decodeFullRouterData(t, resp, &manualReviewList)
	if manualReviewList.Total != 1 || len(manualReviewList.List) != 1 {
		t.Fatalf("unexpected manual review list: %+v", manualReviewList)
	}
	if manualReviewList.List[0].ChallengeID != manualChallenge.ID || manualReviewList.List[0].StudentUsername != env.peerStudent.Username {
		t.Fatalf("unexpected manual review list item: %+v", manualReviewList.List[0])
	}

	resp = performFullRouterRequest(
		t,
		env.router,
		http.MethodGet,
		fmt.Sprintf("/api/v1/teacher/manual-review-submissions?student_id=%d", env.peerStudent.ID),
		nil,
		otherTeacherHeaders,
	)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var otherTeacherManualList struct {
		List  []practicecontracts.TeacherManualReviewSubmissionItemResp `json:"list"`
		Total int64                                                     `json:"total"`
	}
	decodeFullRouterData(t, resp, &otherTeacherManualList)
	if otherTeacherManualList.Total != 0 || len(otherTeacherManualList.List) != 0 {
		t.Fatalf("expected other teacher to see empty manual review list, got %+v", otherTeacherManualList)
	}

	manualSubmissionID := manualReviewList.List[0].ID

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/manual-review-submissions/%d", manualSubmissionID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var manualReviewDetail practicecontracts.TeacherManualReviewSubmissionDetailResp
	decodeFullRouterData(t, resp, &manualReviewDetail)
	if manualReviewDetail.Answer == "" || manualReviewDetail.ChallengeID != manualChallenge.ID {
		t.Fatalf("unexpected manual review detail: %+v", manualReviewDetail)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/teacher/manual-review-submissions/%d/review", manualSubmissionID), map[string]any{
		"review_status":  contestcontracts.SubmissionReviewStatusApproved,
		"review_comment": "证据完整，通过。",
	}, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var reviewedManualSubmission practicecontracts.TeacherManualReviewSubmissionDetailResp
	decodeFullRouterData(t, resp, &reviewedManualSubmission)
	if reviewedManualSubmission.ReviewStatus != contestcontracts.SubmissionReviewStatusApproved || !reviewedManualSubmission.IsCorrect || reviewedManualSubmission.Score != 120 {
		t.Fatalf("unexpected reviewed manual submission: %+v", reviewedManualSubmission)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/authoring/environment-templates", map[string]any{
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
				"image_id":     env.image.ID,
				"service_port": 8080,
				"tier":         challengecontracts.TopologyTierPublic,
				"network_keys": []string{"default"},
			},
		},
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var template challengehttp.EnvironmentTemplateResp
	decodeFullRouterData(t, resp, &template)
	if template.EntryNodeKey != "web" {
		t.Fatalf("unexpected template: %+v", template)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/authoring/environment-templates?keyword=Lifecycle", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var templates []challengehttp.EnvironmentTemplateResp
	decodeFullRouterData(t, resp, &templates)
	if len(templates) == 0 {
		t.Fatalf("expected template list to include created template")
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/topology", createdChallenge.ID), map[string]any{
		"template_id": template.ID,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var topology challengehttp.ChallengeTopologyResp
	decodeFullRouterData(t, resp, &topology)
	if topology.TemplateID == nil || *topology.TemplateID != template.ID {
		t.Fatalf("unexpected topology template binding: %+v", topology)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/authoring/challenges/%d/topology", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/authoring/environment-templates/%d", template.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var loadedTemplate challengehttp.EnvironmentTemplateResp
	decodeFullRouterData(t, resp, &loadedTemplate)
	if loadedTemplate.UsageCount < 1 {
		t.Fatalf("expected template usage count increment, got %+v", loadedTemplate)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/topology", createdChallenge.ID), map[string]any{
		"entry_node_key": "ghost",
		"nodes": []map[string]any{
			{
				"key":          "web",
				"name":         "Web",
				"image_id":     env.image.ID,
				"service_port": 8080,
			},
		},
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusBadRequest)

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/authoring/environment-templates/%d", template.ID), map[string]any{
		"name":           "Lifecycle Template Updated",
		"description":    "updated template",
		"entry_node_key": "web",
		"nodes": []map[string]any{
			{
				"key":          "web",
				"name":         "Web Updated",
				"image_id":     env.image.ID,
				"service_port": 9090,
			},
		},
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/authoring/challenges/%d/topology", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/authoring/challenges/%d/topology", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusNotFound)

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusNotFound)

	instanceChallenge := createDraftChallengeRecord(t, env, "DeleteBlocked Challenge")
	createRunningInstanceForChallenge(t, env, instanceChallenge.ID, env.student.ID)

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/authoring/challenges/%d", instanceChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusConflict)

	stopInstancesForChallenge(t, env, instanceChallenge.ID)
	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/authoring/challenges/%d", instanceChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/authoring/challenges/%d", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/authoring/challenges/%d", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusNotFound)

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/authoring/environment-templates/%d", template.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
}

func TestFullRouter_AdminOpsAndNotificationStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	adminHeaders := bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd))
	teacherHeaders := bearerHeaders(loginForToken(t, env.router, env.teacher.Username, env.teacherPwd))
	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd))
	peerHeaders := bearerHeaders(loginForToken(t, env.router, env.peerStudent.Username, "Password123"))

	resp := performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/authoring/images", map[string]any{
		"name":        "matrix/webapp",
		"tag":         "v2",
		"description": "integration image",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var freeImage challengehttp.ImageResp
	decodeFullRouterData(t, resp, &freeImage)
	if freeImage.Name != "matrix/webapp" {
		t.Fatalf("unexpected created image: %+v", freeImage)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/authoring/images", map[string]any{
		"name":        "matrix/webapp",
		"tag":         "v2",
		"description": "duplicate image",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusConflict)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/authoring/images?name=matrix/status=available", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/authoring/images/%d", freeImage.ID), map[string]any{
		"description": "updated image",
		"status":      "failed",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/authoring/images/%d", freeImage.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var loadedImage challengehttp.ImageResp
	decodeFullRouterData(t, resp, &loadedImage)
	if loadedImage.Status != "failed" || loadedImage.Description != "updated image" {
		t.Fatalf("unexpected loaded image: %+v", loadedImage)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/authoring/images/%d", env.image.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusConflict)

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/authoring/images/%d", freeImage.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/admin/users?role=student&class_name=ClassA", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var userPage map[string]any
	decodeFullRouterData(t, resp, &userPage)
	if int(userPage["total"].(float64)) < 2 {
		t.Fatalf("expected student page results, got %+v", userPage)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/admin/users", map[string]any{
		"username":   "admin_created_student",
		"name":       "Created Student",
		"password":   "Password123",
		"email":      "created_student@example.com",
		"student_no": "20260001",
		"class_name": "ClassA",
		"role":       identitycontracts.RoleStudent,
		"status":     identitycontracts.UserStatusActive,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var createdUserWrap map[string]json.RawMessage
	decodeFullRouterData(t, resp, &createdUserWrap)
	createdUser := decodeFullRouterJSON[identityhttp.AdminUserResp](t, createdUserWrap["user"])
	if createdUser.Username != "admin_created_student" {
		t.Fatalf("unexpected created user: %+v", createdUser)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/admin/users", map[string]any{
		"username": "admin_created_student",
		"password": "Password123",
		"role":     identitycontracts.RoleStudent,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusConflict)

	updatedTeacherNo := "T-9001"
	updatedRole := identitycontracts.RoleTeacher
	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%d", createdUser.ID), map[string]any{
		"role":       updatedRole,
		"teacher_no": updatedTeacherNo,
		"class_name": "ClassTeach",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var updatedUserWrap map[string]json.RawMessage
	decodeFullRouterData(t, resp, &updatedUserWrap)
	updatedUser := decodeFullRouterJSON[identityhttp.AdminUserResp](t, updatedUserWrap["user"])
	if updatedUser.TeacherNo == nil || *updatedUser.TeacherNo != updatedTeacherNo || updatedUser.StudentNo != nil {
		t.Fatalf("unexpected updated user: %+v", updatedUser)
	}

	csvContent := strings.Join([]string{
		"username,password,email,class_name,role,status,student_no,teacher_no,name",
		"import_new,Password123,import_new@example.com,ClassA,student,active,20260002,,Import New",
		"admin_created_student,,updated_import@example.com,ClassTeach,teacher,inactive,,T-9002,Imported Update",
		",Password123,bad@example.com,ClassA,student,active,20260003,,Bad Row",
	}, "\n")
	resp = performFullRouterMultipartRequest(t, env.router, http.MethodPost, "/api/v1/admin/users/import", "file", "users.csv", csvContent, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusCreated)

	var importResult identityhttp.ImportUsersResp
	decodeFullRouterData(t, resp, &importResult)
	if importResult.Created != 1 || importResult.Updated != 1 || importResult.Failed != 1 {
		t.Fatalf("unexpected import result: %+v", importResult)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/admin/users/%d", createdUser.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	if err := env.cache.Set(context.Background(), "ctf:auth:session:manual-online", "online", time.Hour).Err(); err != nil {
		t.Fatalf("seed session key: %v", err)
	}
	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/admin/dashboard", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var dashboard opshttp.DashboardStats
	decodeFullRouterData(t, resp, &dashboard)
	if dashboard.OnlineUsers < 1 || dashboard.ActiveContainers < 1 {
		t.Fatalf("unexpected dashboard stats: %+v", dashboard)
	}

	submitDetail, _ := json.Marshal(map[string]any{"username": env.student.Username, "source": "matrix"})
	for i := 0; i < 5; i++ {
		if err := env.db.Create(&opsentity.AuditLog{
			UserID:       &env.student.ID,
			Action:       auditlog.ActionSubmit,
			ResourceType: "challenge_submission",
			Detail:       string(submitDetail),
			IPAddress:    "10.0.0.1",
			CreatedAt:    time.Now().Add(-time.Duration(i) * time.Minute),
		}).Error; err != nil {
			t.Fatalf("seed submit audit log: %v", err)
		}
	}
	for _, user := range []*identitycontracts.User{env.student, env.peerStudent} {
		if err := env.db.Create(&opsentity.AuditLog{
			UserID:       &user.ID,
			Action:       auditlog.ActionLogin,
			ResourceType: "auth_login",
			Detail:       `{"username":"` + user.Username + `"}`,
			IPAddress:    "10.0.0.99",
			CreatedAt:    time.Now().Add(-10 * time.Minute),
		}).Error; err != nil {
			t.Fatalf("seed login audit log: %v", err)
		}
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/admin/audit-logs?action=submit&page=1&page_size=10", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var auditPage map[string]any
	decodeFullRouterData(t, resp, &auditPage)
	if int(auditPage["total"].(float64)) < 5 {
		t.Fatalf("unexpected audit page: %+v", auditPage)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/admin/audit-logs?start_time=bad-time", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusBadRequest)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/admin/cheat-detection", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var cheat opsqry.CheatDetectionResp
	decodeFullRouterData(t, resp, &cheat)
	if cheat.Summary.SubmitBurstUsers < 1 || cheat.Summary.SharedIPGroups < 1 {
		t.Fatalf("unexpected cheat detection response: %+v", cheat)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/admin/notifications", map[string]any{
		"type":    opsentity.NotificationTypeSystem,
		"title":   "全员通知",
		"content": "full-router matrix admin publish",
		"audience_rules": map[string]any{
			"mode": "union",
			"rules": []map[string]any{
				{"type": "all"},
			},
		},
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var publishResult opshttp.AdminNotificationPublishResp
	decodeFullRouterData(t, resp, &publishResult)
	if publishResult.BatchID <= 0 || publishResult.RecipientCount < 4 {
		t.Fatalf("unexpected publish result: %+v", publishResult)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/admin/notifications", map[string]any{
		"type":    opsentity.NotificationTypeSystem,
		"title":   "teacher forbidden",
		"content": "teacher should not publish",
		"audience_rules": map[string]any{
			"mode": "union",
			"rules": []map[string]any{
				{"type": "all"},
			},
		},
	}, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/admin/notifications", map[string]any{
		"type":    opsentity.NotificationTypeSystem,
		"title":   "invalid audience",
		"content": "missing roles",
		"audience_rules": map[string]any{
			"mode": "union",
			"rules": []map[string]any{
				{"type": "role"},
			},
		},
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusBadRequest)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/notifications?page=1&page_size=10", nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var notificationPage map[string]any
	decodeFullRouterData(t, resp, &notificationPage)
	if int(notificationPage["total"].(float64)) < 2 {
		t.Fatalf("unexpected notifications page: %+v", notificationPage)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/notifications/%d/read", env.notification.ID), nil, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusNotFound)

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/notifications/%d/read", env.notification.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	server := httptest.NewServer(env.router)
	defer server.Close()

	ticketResp := performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/auth/ws-ticket", nil, studentHeaders)
	assertFullRouterStatus(t, ticketResp, http.StatusOK)

	var wsTicket map[string]any
	decodeFullRouterData(t, ticketResp, &wsTicket)
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

	message := receiveFullRouterWSMessageByType(t, conn, "system.connected")
	if message.Type != "system.connected" {
		t.Fatalf("unexpected websocket message: %+v", message)
	}

	reusedConfig, _ := xws.NewConfig(wsURL, server.URL)
	if _, err := xws.DialConfig(reusedConfig); err == nil {
		t.Fatal("expected consumed websocket ticket to be rejected")
	}
}

func TestFullRouter_AdminImagesCapsOversizedPageSize(t *testing.T) {
	env := newFullRouterTestEnv(t)

	adminHeaders := bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd))

	resp := performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/authoring/images?page=1&page_size=200", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var payload struct {
		List []challengehttp.ImageResp `json:"list"`
		Page int                       `json:"page"`
		Size int                       `json:"page_size"`
	}
	decodeFullRouterData(t, resp, &payload)

	if payload.Page != 1 {
		t.Fatalf("expected page=1, got %d", payload.Page)
	}
	if payload.Size != 100 {
		t.Fatalf("expected capped page_size=100, got %d", payload.Size)
	}
	if len(payload.List) == 0 {
		t.Fatal("expected image list to contain seeded records")
	}
}
