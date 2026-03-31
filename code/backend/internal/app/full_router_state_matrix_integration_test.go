package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	practicereadmodelhttp "ctf-platform/internal/module/practice_readmodel/api/http"
	teachinghttp "ctf-platform/internal/module/teaching_readmodel/api/http"
	rediskeys "ctf-platform/internal/pkg/redis"
	flagcrypto "ctf-platform/pkg/crypto"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	xws "golang.org/x/net/websocket"
)

type fullRouterEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func TestTeacherRoutesAreServedByTeachingReadModel(t *testing.T) {
	cfg, db, cache := newAppTestDependencies(t)

	originalBuildTeachingReadmodelModule := buildTeachingReadmodelModule
	t.Cleanup(func() {
		buildTeachingReadmodelModule = originalBuildTeachingReadmodelModule
	})

	called := false
	buildTeachingReadmodelModule = func(root *composition.Root, assessment *composition.AssessmentModule) *composition.TeachingReadmodelModule {
		module := originalBuildTeachingReadmodelModule(root, assessment)
		called = true
		if module == nil || module.Handler == nil {
			t.Fatal("expected teaching readmodel module handler")
		}
		if got, want := reflect.TypeOf(module.Handler), reflect.TypeOf(&teachinghttp.Handler{}); got != want {
			t.Fatalf("teaching readmodel handler type = %v, want %v", got, want)
		}
		return module
	}

	router, err := NewRouter(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}
	if router == nil {
		t.Fatal("expected router")
	}
	if !called {
		t.Fatal("expected teaching readmodel module builder to be called")
	}
}

func TestStudentPracticeReadRoutesAreServedByPracticeReadmodel(t *testing.T) {
	cfg, db, cache := newAppTestDependencies(t)

	originalBuildPracticeReadmodelModule := buildPracticeReadmodelModule
	t.Cleanup(func() {
		buildPracticeReadmodelModule = originalBuildPracticeReadmodelModule
	})

	called := false
	buildPracticeReadmodelModule = func(root *composition.Root) *composition.PracticeReadmodelModule {
		module := originalBuildPracticeReadmodelModule(root)
		called = true
		if module == nil || module.Handler == nil {
			t.Fatal("expected practice readmodel module handler")
		}
		if got, want := reflect.TypeOf(module.Handler), reflect.TypeOf(&practicereadmodelhttp.Handler{}); got != want {
			t.Fatalf("practice readmodel handler type = %v, want %v", got, want)
		}
		return module
	}

	router, err := NewRouter(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}
	if router == nil {
		t.Fatal("expected router")
	}
	if !called {
		t.Fatal("expected practice readmodel module builder to be called")
	}
}

func TestFullRouter_ContestParticipationStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	adminHeaders := bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd))
	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd))
	peerHeaders := bearerHeaders(loginForToken(t, env.router, env.peerStudent.Username, "Password123"))
	otherHeaders := bearerHeaders(loginForToken(t, env.router, env.otherStudent.Username, "Password123"))

	registrationContest := createFullRouterContest(t, env, "Registration Matrix", model.ContestStatusRegistration)
	retryStudent := createFullRouterUser(t, env.db, "student_retry", "Password123", model.RoleStudent, env.className)
	retryHeaders := bearerHeaders(loginForToken(t, env.router, retryStudent.Username, "Password123"))
	fillerStudent := createFullRouterUser(t, env.db, "student_filler", "Password123", model.RoleStudent, env.className)

	resp := performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/register", registrationContest.ID), nil, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	peerRegistration := findContestRegistration(t, env, registrationContest.ID, env.peerStudent.ID)
	if peerRegistration.Status != model.ContestRegistrationStatusPending {
		t.Fatalf("expected pending registration, got %s", peerRegistration.Status)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/teams", registrationContest.ID), map[string]any{
		"name":        "PendingTeam",
		"max_members": 3,
	}, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/admin/contests/%d/registrations?status=pending", registrationContest.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
	var registrationPage map[string]any
	decodeFullRouterData(t, resp, &registrationPage)
	if total := int(registrationPage["total"].(float64)); total != 1 {
		t.Fatalf("expected 1 pending registration, got %d", total)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/contests/%d/registrations/%d", registrationContest.ID, peerRegistration.ID), map[string]any{
		"status": model.ContestRegistrationStatusApproved,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/register", env.contest.ID), nil, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	createContestRegistration(t, env, registrationContest.ID, env.student.ID, model.ContestRegistrationStatusApproved, nil)
	createContestRegistration(t, env, registrationContest.ID, env.otherStudent.ID, model.ContestRegistrationStatusApproved, nil)
	createContestRegistration(t, env, registrationContest.ID, fillerStudent.ID, model.ContestRegistrationStatusApproved, nil)
	createContestRegistration(t, env, registrationContest.ID, retryStudent.ID, model.ContestRegistrationStatusRejected, nil)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/register", registrationContest.ID), nil, retryHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
	retryRegistration := findContestRegistration(t, env, registrationContest.ID, retryStudent.ID)
	if retryRegistration.Status != model.ContestRegistrationStatusPending {
		t.Fatalf("expected rejected registration to requeue as pending, got %s", retryRegistration.Status)
	}

	fullTeam := createContestTeam(t, env, registrationContest.ID, env.otherStudent.ID, "FullTeam", 2)
	createContestTeamMember(t, env, registrationContest.ID, fullTeam.ID, fillerStudent.ID)
	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/teams/%d/join", registrationContest.ID, fullTeam.ID), nil, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/teams", registrationContest.ID), map[string]any{
		"name":        "AlphaTeam",
		"max_members": 4,
	}, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var createdTeam dto.TeamResp
	decodeFullRouterData(t, resp, &createdTeam)
	if createdTeam.CaptainID != env.student.ID {
		t.Fatalf("expected student captain id %d, got %d", env.student.ID, createdTeam.CaptainID)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/teams", registrationContest.ID), map[string]any{
		"name":        "DuplicateTeam",
		"max_members": 4,
	}, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusConflict)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/teams/%d/join", registrationContest.ID, createdTeam.ID), nil, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/contests/%d/teams/%d", registrationContest.ID, createdTeam.ID), nil, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/contests/%d/teams/%d/leave", registrationContest.ID, createdTeam.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/contests/%d/teams/%d/leave", registrationContest.ID, createdTeam.ID), nil, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/contests/%d/teams/%d", registrationContest.ID, createdTeam.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	createContestSubmission(t, env, env.contest.ID, env.team.ID, env.student.ID, env.challenge.ID, 100)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/my-progress", env.contest.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var progress dto.ContestMyProgressResp
	decodeFullRouterData(t, resp, &progress)
	if progress.ContestID != env.contest.ID || len(progress.Solved) == 0 {
		t.Fatalf("expected existing contest progress, got %+v", progress)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/my-progress", env.contest.ID), nil, otherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var emptyProgress dto.ContestMyProgressResp
	decodeFullRouterData(t, resp, &emptyProgress)
	if emptyProgress.TeamID != nil || len(emptyProgress.Solved) != 0 {
		t.Fatalf("expected empty progress for unregistered student, got %+v", emptyProgress)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/announcements", env.contest.ID), nil, nil)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var announcements []dto.ContestAnnouncementResp
	decodeFullRouterData(t, resp, &announcements)
	if len(announcements) == 0 {
		t.Fatalf("expected seeded announcement")
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/announcements", env.contest.ID), map[string]any{
		"title":   "新的公告",
		"content": "integration notice",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var createdAnnouncement dto.ContestAnnouncementResp
	decodeFullRouterData(t, resp, &createdAnnouncement)
	if createdAnnouncement.Title != "新的公告" {
		t.Fatalf("unexpected announcement title: %+v", createdAnnouncement)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/admin/contests/%d/announcements/%d", env.contest.ID, createdAnnouncement.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
}

func TestFullRouter_ReportPreviewAndDownloadStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	adminHeaders := bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd))
	teacherHeaders := bearerHeaders(loginForToken(t, env.router, env.teacher.Username, env.teacherPwd))
	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd))

	resp := performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/reports/personal", map[string]any{
		"format": model.ReportFormatExcel,
	}, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var personalReport dto.ReportExportData
	decodeFullRouterData(t, resp, &personalReport)
	if personalReport.Status != model.ReportStatusReady || personalReport.DownloadURL == nil {
		t.Fatalf("expected ready personal report with download url, got %+v", personalReport)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", personalReport.ReportID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var personalStatus dto.ReportExportData
	decodeFullRouterData(t, resp, &personalStatus)
	if personalStatus.Status != model.ReportStatusReady || personalStatus.DownloadURL == nil {
		t.Fatalf("expected ready personal report status, got %+v", personalStatus)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/download", personalReport.ReportID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
	if contentType := resp.Header().Get("Content-Type"); contentType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Fatalf("expected xlsx content-type, got %q", contentType)
	}

	processingReport := createReportRecord(t, env, model.Report{
		Type:   model.ReportTypePersonal,
		Format: model.ReportFormatPDF,
		UserID: &env.student.ID,
		Status: model.ReportStatusProcessing,
	})

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", processingReport.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var processingStatus dto.ReportExportData
	decodeFullRouterData(t, resp, &processingStatus)
	if processingStatus.Status != model.ReportStatusProcessing {
		t.Fatalf("expected processing status, got %+v", processingStatus)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/download", processingReport.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusConflict)

	failedMessage := "generation failed in matrix"
	failedReport := createReportRecord(t, env, model.Report{
		Type:     model.ReportTypePersonal,
		Format:   model.ReportFormatPDF,
		UserID:   &env.student.ID,
		Status:   model.ReportStatusFailed,
		ErrorMsg: &failedMessage,
	})

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", failedReport.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var failedStatus dto.ReportExportData
	decodeFullRouterData(t, resp, &failedStatus)
	if failedStatus.Status != model.ReportStatusFailed || failedStatus.ErrorMessage == nil || *failedStatus.ErrorMessage != failedMessage {
		t.Fatalf("expected failed status with message, got %+v", failedStatus)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/download", failedReport.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusConflict)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/reports/class", map[string]any{
		"class_name": env.otherStudent.ClassName,
		"format":     model.ReportFormatPDF,
	}, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/reports/class", map[string]any{
		"class_name": env.className,
		"format":     model.ReportFormatPDF,
	}, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var classReport dto.ReportExportData
	decodeFullRouterData(t, resp, &classReport)
	if classReport.Status != model.ReportStatusProcessing {
		t.Fatalf("expected class report to start in processing state, got %+v", classReport)
	}

	classReady := waitForReportStatus(t, env, classReport.ReportID, teacherHeaders, model.ReportStatusReady, 5*time.Second)
	if classReady.DownloadURL == nil {
		t.Fatalf("expected class report download url after ready, got %+v", classReady)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/download", classReport.ReportID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
	if contentType := resp.Header().Get("Content-Type"); contentType != "application/pdf" {
		t.Fatalf("expected pdf content-type, got %q", contentType)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", classReport.ReportID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", classReport.ReportID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
}

func TestFullRouter_TeacherAccessAndRecommendationStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)
	createRecommendationChallenge(t, env, "Matrix Weak Web 2", model.DimensionWeb)

	adminHeaders := bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd))
	teacherHeaders := bearerHeaders(loginForToken(t, env.router, env.teacher.Username, env.teacherPwd))
	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd))

	resp := performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/teacher/classes", nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var teacherClasses []dto.TeacherClassItem
	decodeFullRouterData(t, resp, &teacherClasses)
	if len(teacherClasses) != 1 || teacherClasses[0].Name != env.className {
		t.Fatalf("expected only teacher class, got %+v", teacherClasses)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/teacher/classes", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var adminClasses []dto.TeacherClassItem
	decodeFullRouterData(t, resp, &adminClasses)
	if len(adminClasses) < 2 {
		t.Fatalf("expected admin to see all classes, got %+v", adminClasses)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/classes/%s/summary", env.className), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/classes/%s/summary", env.otherStudent.ClassName), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/classes/%s/trend", env.className), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/classes/%s/review", env.className), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/students/%d/progress", env.student.ID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var progress dto.TeacherProgressResp
	decodeFullRouterData(t, resp, &progress)
	if progress.SolvedChallenges == 0 {
		t.Fatalf("expected solved challenges in teacher progress, got %+v", progress)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/students/%d/timeline", env.student.ID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var timeline dto.TimelineResp
	decodeFullRouterData(t, resp, &timeline)
	if len(timeline.Events) == 0 {
		t.Fatalf("expected timeline events, got %+v", timeline)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/students/%d/recommendations", env.student.ID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var teacherRecommendations []dto.TeacherRecommendationItem
	decodeFullRouterData(t, resp, &teacherRecommendations)
	if len(teacherRecommendations) == 0 {
		t.Fatalf("expected teacher recommendations, got %+v", teacherRecommendations)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/users/%d/skill-profile", env.student.ID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var skillProfile dto.SkillProfileResp
	decodeFullRouterData(t, resp, &skillProfile)
	if skillProfile.UserID != env.student.ID {
		t.Fatalf("expected skill profile for student %d, got %+v", env.student.ID, skillProfile)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/students/%d/progress", env.otherStudent.ID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/users/%d/skill-profile", env.otherStudent.ID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/users/me/recommendations", nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var selfRecommendations dto.RecommendationResp
	decodeFullRouterData(t, resp, &selfRecommendations)
	if len(selfRecommendations.Challenges) == 0 {
		t.Fatalf("expected self recommendations, got %+v", selfRecommendations)
	}
}

func TestFullRouter_AdminChallengeManagementStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	adminHeaders := bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd))
	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.peerStudent.Username, "Password123"))

	resp := performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/admin/challenges", map[string]any{
		"title":       "Lifecycle Challenge",
		"description": "challenge lifecycle matrix",
		"category":    model.DimensionWeb,
		"difficulty":  model.ChallengeDifficultyEasy,
		"points":      120,
		"image_id":    env.image.ID,
		"hints": []map[string]any{
			{
				"level":       1,
				"title":       "入口",
				"content":     "look at login",
				"cost_points": 5,
			},
			{
				"level":       2,
				"title":       "深入",
				"content":     "check cookies",
				"cost_points": 10,
			},
		},
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var createdChallenge dto.ChallengeResp
	decodeFullRouterData(t, resp, &createdChallenge)
	if createdChallenge.Status != model.ChallengeStatusDraft || len(createdChallenge.Hints) != 2 {
		t.Fatalf("unexpected created challenge: %+v", createdChallenge)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/admin/challenges", map[string]any{
		"title":       "Invalid Hint Challenge",
		"description": "invalid hints",
		"category":    model.DimensionWeb,
		"difficulty":  model.ChallengeDifficultyEasy,
		"points":      80,
		"image_id":    env.image.ID,
		"hints": []map[string]any{
			{"level": 1, "content": "a"},
			{"level": 1, "content": "b"},
		},
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusBadRequest)

	emptyAttachment := ""
	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/challenges/%d", createdChallenge.ID), map[string]any{
		"title":          "Lifecycle Challenge Updated",
		"points":         150,
		"attachment_url": emptyAttachment,
		"hints": []map[string]any{
			{
				"level":       1,
				"title":       "更新提示",
				"content":     "updated content",
				"cost_points": 8,
			},
		},
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/admin/challenges/%d", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var updatedChallenge dto.ChallengeResp
	decodeFullRouterData(t, resp, &updatedChallenge)
	if updatedChallenge.Title != "Lifecycle Challenge Updated" || updatedChallenge.Points != 150 || len(updatedChallenge.Hints) != 1 {
		t.Fatalf("unexpected updated challenge: %+v", updatedChallenge)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/challenges/%d/flag", createdChallenge.ID), map[string]any{
		"flag_type":   model.FlagTypeStatic,
		"flag":        "invalid-flag",
		"flag_prefix": "flag",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusBadRequest)

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/challenges/%d/flag", createdChallenge.ID), map[string]any{
		"flag_type":   model.FlagTypeStatic,
		"flag":        "flag{lifecycle-static}",
		"flag_prefix": "flag",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/admin/challenges/%d/flag", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var staticFlag dto.FlagResp
	decodeFullRouterData(t, resp, &staticFlag)
	if !staticFlag.Configured || staticFlag.FlagType != model.FlagTypeStatic {
		t.Fatalf("unexpected static flag config: %+v", staticFlag)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/challenges/%d/flag", createdChallenge.ID), map[string]any{
		"flag_type":   model.FlagTypeDynamic,
		"flag_prefix": "ctf",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/admin/challenges/%d/flag", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var dynamicFlag dto.FlagResp
	decodeFullRouterData(t, resp, &dynamicFlag)
	if !dynamicFlag.Configured || dynamicFlag.FlagType != model.FlagTypeDynamic || dynamicFlag.FlagPrefix != "ctf" {
		t.Fatalf("unexpected dynamic flag config: %+v", dynamicFlag)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/challenges/%d/writeup", createdChallenge.ID), map[string]any{
		"title":      "Scheduled Writeup",
		"content":    "scheduled content",
		"visibility": model.WriteupVisibilityScheduled,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusBadRequest)

	releaseAt := time.Now().Add(time.Hour).Format(time.RFC3339)
	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/challenges/%d/writeup", createdChallenge.ID), map[string]any{
		"title":      "Scheduled Writeup",
		"content":    "scheduled content",
		"visibility": model.WriteupVisibilityScheduled,
		"release_at": releaseAt,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/admin/challenges/%d/writeup", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var adminWriteup dto.AdminChallengeWriteupResp
	decodeFullRouterData(t, resp, &adminWriteup)
	if adminWriteup.Visibility != model.WriteupVisibilityScheduled || adminWriteup.ReleaseAt == nil {
		t.Fatalf("unexpected admin writeup: %+v", adminWriteup)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/challenges/%d/publish", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d", createdChallenge.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var publishedDetail dto.ChallengeDetailResp
	decodeFullRouterData(t, resp, &publishedDetail)
	if publishedDetail.ID != createdChallenge.ID {
		t.Fatalf("unexpected published detail: %+v", publishedDetail)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d/writeup", createdChallenge.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusNotFound)

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/challenges/%d/writeup", createdChallenge.ID), map[string]any{
		"title":      "Public Writeup",
		"content":    "public content",
		"visibility": model.WriteupVisibilityPublic,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d/writeup", createdChallenge.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var publicWriteup dto.ChallengeWriteupResp
	decodeFullRouterData(t, resp, &publicWriteup)
	if !publicWriteup.RequiresSpoilerWarning {
		t.Fatalf("expected spoiler warning before solving, got %+v", publicWriteup)
	}

	createPracticeSubmission(t, env, env.peerStudent.ID, createdChallenge.ID, 150)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d/writeup", createdChallenge.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var solvedWriteup dto.ChallengeWriteupResp
	decodeFullRouterData(t, resp, &solvedWriteup)
	if solvedWriteup.RequiresSpoilerWarning {
		t.Fatalf("expected spoiler warning to clear after solve, got %+v", solvedWriteup)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/admin/environment-templates", map[string]any{
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
				"tier":         model.TopologyTierPublic,
				"network_keys": []string{"default"},
			},
		},
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var template dto.EnvironmentTemplateResp
	decodeFullRouterData(t, resp, &template)
	if template.EntryNodeKey != "web" {
		t.Fatalf("unexpected template: %+v", template)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/admin/environment-templates?keyword=Lifecycle", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var templates []dto.EnvironmentTemplateResp
	decodeFullRouterData(t, resp, &templates)
	if len(templates) == 0 {
		t.Fatalf("expected template list to include created template")
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/challenges/%d/topology", createdChallenge.ID), map[string]any{
		"template_id": template.ID,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var topology dto.ChallengeTopologyResp
	decodeFullRouterData(t, resp, &topology)
	if topology.TemplateID == nil || *topology.TemplateID != template.ID {
		t.Fatalf("unexpected topology template binding: %+v", topology)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/admin/challenges/%d/topology", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/admin/environment-templates/%d", template.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var loadedTemplate dto.EnvironmentTemplateResp
	decodeFullRouterData(t, resp, &loadedTemplate)
	if loadedTemplate.UsageCount < 1 {
		t.Fatalf("expected template usage count increment, got %+v", loadedTemplate)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/challenges/%d/topology", createdChallenge.ID), map[string]any{
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

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/environment-templates/%d", template.ID), map[string]any{
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

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/admin/challenges/%d/topology", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/admin/challenges/%d/topology", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusNotFound)

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/admin/challenges/%d/writeup", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/admin/challenges/%d/writeup", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusNotFound)

	instanceChallenge := createDraftChallengeRecord(t, env, "DeleteBlocked Challenge")
	createRunningInstanceForChallenge(t, env, instanceChallenge.ID, env.student.ID)

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/admin/challenges/%d", instanceChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusConflict)

	stopInstancesForChallenge(t, env, instanceChallenge.ID)
	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/admin/challenges/%d", instanceChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/admin/challenges/%d", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/admin/challenges/%d", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusNotFound)

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/admin/environment-templates/%d", template.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
}

func TestFullRouter_InstanceHintAndProxyStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd))
	peerHeaders := bearerHeaders(loginForToken(t, env.router, env.peerStudent.Username, "Password123"))
	teacherHeaders := bearerHeaders(loginForToken(t, env.router, env.teacher.Username, env.teacherPwd))

	resp := performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/challenges/%d/hints/1/unlock", env.challenge.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var unlocked dto.UnlockHintResp
	decodeFullRouterData(t, resp, &unlocked)
	if unlocked.Hint == nil || !unlocked.Hint.IsUnlocked || unlocked.Hint.Level != 1 {
		t.Fatalf("unexpected unlock response: %+v", unlocked)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/challenges/%d/hints/99/unlock", env.challenge.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusNotFound)

	draftChallenge := createDraftChallengeRecord(t, env, "Draft Hint Challenge")
	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/challenges/%d/hints/1/unlock", draftChallenge.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/instances/%d/extend", env.instance.ID), nil, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/instances/%d/extend", env.instance.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
	var extended dto.InstanceResp
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

	var teacherInstances []dto.TeacherInstanceItem
	decodeFullRouterData(t, resp, &teacherInstances)
	if len(teacherInstances) == 0 {
		t.Fatalf("expected teacher instances for own class")
	}

	proxyTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("proxied:" + r.URL.Path))
	}))
	defer proxyTarget.Close()

	if err := env.db.Model(&model.Instance{}).Where("id = ?", env.instance.ID).Updates(map[string]any{
		"access_url": proxyTarget.URL,
		"status":     model.InstanceStatusRunning,
	}).Error; err != nil {
		t.Fatalf("update instance access url: %v", err)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/instances/%d/access", env.instance.ID), nil, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/instances/%d/proxy/ping", env.instance.ID), nil, nil)
	assertFullRouterStatus(t, resp, http.StatusUnauthorized)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/instances/%d/access", env.instance.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var access dto.InstanceAccessResp
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

	if err := env.db.Model(&model.Instance{}).Where("id = ?", env.instance.ID).Updates(map[string]any{
		"status": model.InstanceStatusStopped,
	}).Error; err != nil {
		t.Fatalf("stop instance: %v", err)
	}
	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/instances/%d/access", env.instance.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusGone)
}

func TestFullRouter_ContestChallengeAndScoreboardStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	adminHeaders := bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd))
	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd))
	peerHeaders := bearerHeaders(loginForToken(t, env.router, env.peerStudent.Username, "Password123"))
	otherHeaders := bearerHeaders(loginForToken(t, env.router, env.otherStudent.Username, "Password123"))

	challengeA := createRecommendationChallenge(t, env, "Contest Matrix A", model.DimensionWeb)
	challengeB := createRecommendationChallenge(t, env, "Contest Matrix B", model.DimensionWeb)
	editableContest := createFullRouterContest(t, env, "Editable Contest", model.ContestStatusRegistration)

	resp := performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/challenges", editableContest.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	hidden := false
	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/challenges", editableContest.ID), map[string]any{
		"challenge_id": challengeA.ID,
		"points":       220,
		"order":        1,
		"is_visible":   hidden,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var contestChallenge dto.ContestChallengeResp
	decodeFullRouterData(t, resp, &contestChallenge)
	if contestChallenge.Points != 220 || contestChallenge.IsVisible {
		t.Fatalf("unexpected contest challenge: %+v", contestChallenge)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/challenges", editableContest.ID), map[string]any{
		"challenge_id": challengeA.ID,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusConflict)

	updatedVisible := true
	updatedPoints := 260
	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/contests/%d/challenges/%d", editableContest.ID, challengeA.ID), map[string]any{
		"points":     updatedPoints,
		"order":      2,
		"is_visible": updatedVisible,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/admin/contests/%d/challenges", editableContest.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var adminChallenges []dto.ContestChallengeResp
	decodeFullRouterData(t, resp, &adminChallenges)
	if len(adminChallenges) != 1 || adminChallenges[0].Points != updatedPoints || adminChallenges[0].Order != 2 || !adminChallenges[0].IsVisible {
		t.Fatalf("unexpected admin contest challenges: %+v", adminChallenges)
	}

	createContestRegistration(t, env, editableContest.ID, env.student.ID, model.ContestRegistrationStatusApproved, nil)
	createContestRegistration(t, env, editableContest.ID, env.peerStudent.ID, model.ContestRegistrationStatusPending, nil)

	setContestStatus(t, env, editableContest.ID, model.ContestStatusRunning, nil)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/challenges", editableContest.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var visibleChallenges []dto.ContestChallengeInfo
	decodeFullRouterData(t, resp, &visibleChallenges)
	if len(visibleChallenges) != 1 || visibleChallenges[0].ChallengeID != challengeA.ID || visibleChallenges[0].Points != updatedPoints {
		t.Fatalf("unexpected visible contest challenges: %+v", visibleChallenges)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/challenges/%d/instances", editableContest.ID, challengeA.ID), nil, otherHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/challenges/%d/instances", editableContest.ID, challengeA.ID), nil, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/challenges/%d/instances", editableContest.ID, challengeA.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var startedContestInstance dto.InstanceResp
	decodeFullRouterData(t, resp, &startedContestInstance)
	if startedContestInstance.ChallengeID != challengeA.ID {
		t.Fatalf("unexpected started contest instance: %+v", startedContestInstance)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/challenges", editableContest.ID), map[string]any{
		"challenge_id": challengeB.ID,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	conflictContest := createFullRouterContest(t, env, "Conflict Contest", model.ContestStatusRegistration)
	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/challenges", conflictContest.ID), map[string]any{
		"challenge_id": challengeB.ID,
		"points":       180,
		"order":        1,
		"is_visible":   true,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	createContestSubmission(t, env, conflictContest.ID, env.team.ID, env.student.ID, challengeB.ID, 180)
	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/admin/contests/%d/challenges/%d", conflictContest.ID, challengeB.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusConflict)

	if err := env.db.Where("contest_id = ? AND challenge_id = ?", conflictContest.ID, challengeB.ID).Delete(&model.Submission{}).Error; err != nil {
		t.Fatalf("delete conflict contest submissions: %v", err)
	}
	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/admin/contests/%d/challenges/%d", conflictContest.ID, challengeB.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	scoreboardContest := createFullRouterContest(t, env, "Scoreboard Contest", model.ContestStatusRunning)
	createContestRegistration(t, env, scoreboardContest.ID, env.student.ID, model.ContestRegistrationStatusApproved, nil)
	createContestRegistration(t, env, scoreboardContest.ID, env.peerStudent.ID, model.ContestRegistrationStatusApproved, nil)
	teamAlpha := createContestTeam(t, env, scoreboardContest.ID, env.student.ID, "Alpha", 4)
	teamBeta := createContestTeam(t, env, scoreboardContest.ID, env.peerStudent.ID, "Beta", 4)
	seedContestScore(t, env, scoreboardContest.ID, teamAlpha.ID, 100)
	seedContestScore(t, env, scoreboardContest.ID, teamBeta.ID, 80)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/scoreboard", scoreboardContest.ID), nil, nil)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var publicScoreboard dto.ScoreboardResp
	decodeFullRouterData(t, resp, &publicScoreboard)
	if publicScoreboard.Frozen || len(publicScoreboard.Scoreboard.List) != 2 || publicScoreboard.Scoreboard.List[0].TeamID != teamAlpha.ID {
		t.Fatalf("unexpected public scoreboard: %+v", publicScoreboard)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/freeze", scoreboardContest.ID), map[string]any{
		"minutes_before_end": 180,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	seedContestScore(t, env, scoreboardContest.ID, teamBeta.ID, 200)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/scoreboard", scoreboardContest.ID), nil, nil)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var frozenScoreboard dto.ScoreboardResp
	decodeFullRouterData(t, resp, &frozenScoreboard)
	if !frozenScoreboard.Frozen || frozenScoreboard.Scoreboard.List[0].TeamID != teamAlpha.ID || frozenScoreboard.Scoreboard.List[0].Score != 100 {
		t.Fatalf("unexpected frozen scoreboard: %+v", frozenScoreboard)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/admin/contests/%d/scoreboard/live", scoreboardContest.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var liveScoreboard dto.ScoreboardResp
	decodeFullRouterData(t, resp, &liveScoreboard)
	if liveScoreboard.Scoreboard.List[0].TeamID != teamBeta.ID || liveScoreboard.Scoreboard.List[0].Score != 200 {
		t.Fatalf("unexpected live scoreboard: %+v", liveScoreboard)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/unfreeze", scoreboardContest.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/scoreboard", scoreboardContest.ID), nil, nil)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var unfrozenScoreboard dto.ScoreboardResp
	decodeFullRouterData(t, resp, &unfrozenScoreboard)
	if unfrozenScoreboard.Frozen || unfrozenScoreboard.Scoreboard.List[0].TeamID != teamBeta.ID || unfrozenScoreboard.Scoreboard.List[0].Score != 200 {
		t.Fatalf("unexpected unfrozen scoreboard: %+v", unfrozenScoreboard)
	}

	notFrozenContest := createFullRouterContest(t, env, "Not Frozen Contest", model.ContestStatusRunning)
	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/unfreeze", notFrozenContest.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusBadRequest)
}

func TestFullRouter_AdminOpsAndNotificationStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	adminHeaders := bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd))
	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd))
	peerHeaders := bearerHeaders(loginForToken(t, env.router, env.peerStudent.Username, "Password123"))

	resp := performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/admin/images", map[string]any{
		"name":        "matrix/webapp",
		"tag":         "v2",
		"description": "integration image",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var freeImage dto.ImageResp
	decodeFullRouterData(t, resp, &freeImage)
	if freeImage.Name != "matrix/webapp" {
		t.Fatalf("unexpected created image: %+v", freeImage)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/admin/images", map[string]any{
		"name":        "matrix/webapp",
		"tag":         "v2",
		"description": "duplicate image",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusConflict)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/admin/images?name=matrix/status=available", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/images/%d", freeImage.ID), map[string]any{
		"description": "updated image",
		"status":      model.ImageStatusFailed,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/admin/images/%d", freeImage.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var loadedImage dto.ImageResp
	decodeFullRouterData(t, resp, &loadedImage)
	if loadedImage.Status != model.ImageStatusFailed || loadedImage.Description != "updated image" {
		t.Fatalf("unexpected loaded image: %+v", loadedImage)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/admin/images/%d", env.image.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusConflict)

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/admin/images/%d", freeImage.ID), nil, adminHeaders)
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
		"role":       model.RoleStudent,
		"status":     model.UserStatusActive,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var createdUserWrap map[string]json.RawMessage
	decodeFullRouterData(t, resp, &createdUserWrap)
	createdUser := decodeFullRouterJSON[dto.AdminUserResp](t, createdUserWrap["user"])
	if createdUser.Username != "admin_created_student" {
		t.Fatalf("unexpected created user: %+v", createdUser)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/admin/users", map[string]any{
		"username": "admin_created_student",
		"password": "Password123",
		"role":     model.RoleStudent,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusConflict)

	updatedTeacherNo := "T-9001"
	updatedRole := model.RoleTeacher
	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%d", createdUser.ID), map[string]any{
		"role":       updatedRole,
		"teacher_no": updatedTeacherNo,
		"class_name": "ClassTeach",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var updatedUserWrap map[string]json.RawMessage
	decodeFullRouterData(t, resp, &updatedUserWrap)
	updatedUser := decodeFullRouterJSON[dto.AdminUserResp](t, updatedUserWrap["user"])
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

	var importResult dto.ImportUsersResp
	decodeFullRouterData(t, resp, &importResult)
	if importResult.Created != 1 || importResult.Updated != 1 || importResult.Failed != 1 {
		t.Fatalf("unexpected import result: %+v", importResult)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/admin/users/%d", createdUser.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	if err := env.cache.Set(context.Background(), rediskeys.TokenKey(env.student.ID), "online", time.Hour).Err(); err != nil {
		t.Fatalf("seed token key: %v", err)
	}
	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/admin/dashboard", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var dashboard dto.DashboardStats
	decodeFullRouterData(t, resp, &dashboard)
	if dashboard.OnlineUsers < 1 || dashboard.ActiveContainers < 1 {
		t.Fatalf("unexpected dashboard stats: %+v", dashboard)
	}

	submitDetail, _ := json.Marshal(map[string]any{"username": env.student.Username, "source": "matrix"})
	for i := 0; i < 5; i++ {
		if err := env.db.Create(&model.AuditLog{
			UserID:       &env.student.ID,
			Action:       model.AuditActionSubmit,
			ResourceType: "challenge_submission",
			Detail:       string(submitDetail),
			IPAddress:    "10.0.0.1",
			CreatedAt:    time.Now().Add(-time.Duration(i) * time.Minute),
		}).Error; err != nil {
			t.Fatalf("seed submit audit log: %v", err)
		}
	}
	for _, user := range []*model.User{env.student, env.peerStudent} {
		if err := env.db.Create(&model.AuditLog{
			UserID:       &user.ID,
			Action:       model.AuditActionLogin,
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

	var cheat dto.CheatDetectionResp
	decodeFullRouterData(t, resp, &cheat)
	if cheat.Summary.SubmitBurstUsers < 1 || cheat.Summary.SharedIPGroups < 1 {
		t.Fatalf("unexpected cheat detection response: %+v", cheat)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/notifications?page=1&page_size=10", nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var notificationPage map[string]any
	decodeFullRouterData(t, resp, &notificationPage)
	if int(notificationPage["total"].(float64)) < 1 {
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

	resp := performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/admin/images?page=1&page_size=200", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var payload struct {
		List []dto.ImageResp `json:"list"`
		Page int             `json:"page"`
		Size int             `json:"page_size"`
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

func assertFullRouterStatus(t *testing.T, resp *httptest.ResponseRecorder, want int) {
	t.Helper()
	if resp.Code != want {
		t.Fatalf("expected status %d, got %d body=%s", want, resp.Code, resp.Body.String())
	}
}

func decodeFullRouterData(t *testing.T, resp *httptest.ResponseRecorder, target any) {
	t.Helper()

	var envelope fullRouterEnvelope
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response envelope: %v body=%s", err, resp.Body.String())
	}
	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		t.Fatalf("expected response data, got empty body=%s", resp.Body.String())
	}
	if err := json.Unmarshal(envelope.Data, target); err != nil {
		t.Fatalf("decode response data: %v body=%s", err, resp.Body.String())
	}
}

func decodeFullRouterJSON[T any](t *testing.T, data []byte) T {
	t.Helper()
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("decode nested json: %v payload=%s", err, string(data))
	}
	return value
}

type fullRouterWSEnvelope struct {
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

func receiveFullRouterWSMessageByType(t *testing.T, conn *xws.Conn, expectedType string) fullRouterWSEnvelope {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	if err := conn.SetDeadline(deadline); err != nil {
		t.Fatalf("set websocket deadline: %v", err)
	}
	for {
		var message fullRouterWSEnvelope
		if err := xws.JSON.Receive(conn, &message); err != nil {
			t.Fatalf("receive websocket message: %v", err)
		}
		if message.Type == expectedType {
			return message
		}
	}
}

func waitForReportStatus(t *testing.T, env *fullRouterTestEnv, reportID int64, headers map[string]string, wantStatus string, timeout time.Duration) *dto.ReportExportData {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp := performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", reportID), nil, headers)
		if resp.Code != http.StatusOK {
			t.Fatalf("unexpected report status response: %d body=%s", resp.Code, resp.Body.String())
		}

		var report dto.ReportExportData
		decodeFullRouterData(t, resp, &report)
		if report.Status == wantStatus {
			return &report
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for report %d status %s", reportID, wantStatus)
	return nil
}

func createFullRouterContest(t *testing.T, env *fullRouterTestEnv, title, status string) *model.Contest {
	t.Helper()

	now := time.Now()
	contest := &model.Contest{
		Title:       title,
		Description: "state matrix contest",
		Mode:        model.ContestModeJeopardy,
		StartTime:   now.Add(-30 * time.Minute),
		EndTime:     now.Add(2 * time.Hour),
		Status:      status,
	}
	if err := env.db.Create(contest).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	return contest
}

func performFullRouterMultipartRequest(
	t *testing.T,
	router http.Handler,
	method string,
	target string,
	fieldName string,
	fileName string,
	content string,
	headers map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write multipart file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(method, target, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func createContestRegistration(t *testing.T, env *fullRouterTestEnv, contestID, userID int64, status string, teamID *int64) *model.ContestRegistration {
	t.Helper()

	registration := &model.ContestRegistration{
		ContestID: contestID,
		UserID:    userID,
		TeamID:    teamID,
		Status:    status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := env.db.Create(registration).Error; err != nil {
		t.Fatalf("create contest registration: %v", err)
	}
	return registration
}

func findContestRegistration(t *testing.T, env *fullRouterTestEnv, contestID, userID int64) *model.ContestRegistration {
	t.Helper()

	var registration model.ContestRegistration
	if err := env.db.Where("contest_id = ? AND user_id = ?", contestID, userID).First(&registration).Error; err != nil {
		t.Fatalf("find contest registration: %v", err)
	}
	return &registration
}

func createContestTeam(t *testing.T, env *fullRouterTestEnv, contestID, captainID int64, name string, maxMembers int) *model.Team {
	t.Helper()

	team := &model.Team{
		ContestID:  contestID,
		Name:       name,
		CaptainID:  captainID,
		InviteCode: fmt.Sprintf("TEAM%d", time.Now().UnixNano()),
		MaxMembers: maxMembers,
	}
	if err := env.db.Create(team).Error; err != nil {
		t.Fatalf("create contest team: %v", err)
	}
	createContestTeamMember(t, env, contestID, team.ID, captainID)
	if err := env.db.Model(&model.ContestRegistration{}).
		Where("contest_id = ? AND user_id = ?", contestID, captainID).
		Updates(map[string]any{"team_id": team.ID, "updated_at": time.Now()}).Error; err != nil {
		t.Fatalf("bind captain registration to team: %v", err)
	}
	return team
}

func createContestTeamMember(t *testing.T, env *fullRouterTestEnv, contestID, teamID, userID int64) {
	t.Helper()

	if err := env.db.Create(&model.TeamMember{
		ContestID: contestID,
		TeamID:    teamID,
		UserID:    userID,
		JoinedAt:  time.Now(),
	}).Error; err != nil {
		t.Fatalf("create contest team member: %v", err)
	}
	if err := env.db.Model(&model.ContestRegistration{}).
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		Updates(map[string]any{"team_id": teamID, "updated_at": time.Now()}).Error; err != nil {
		t.Fatalf("bind member registration to team: %v", err)
	}
}

func createContestSubmission(t *testing.T, env *fullRouterTestEnv, contestID, teamID, userID, challengeID int64, score int) {
	t.Helper()

	if err := env.db.Create(&model.Submission{
		UserID:      userID,
		ChallengeID: challengeID,
		ContestID:   &contestID,
		TeamID:      &teamID,
		IsCorrect:   true,
		Score:       score,
		SubmittedAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("create contest submission: %v", err)
	}
}

func createPracticeSubmission(t *testing.T, env *fullRouterTestEnv, userID, challengeID int64, score int) {
	t.Helper()

	if err := env.db.Create(&model.Submission{
		UserID:      userID,
		ChallengeID: challengeID,
		IsCorrect:   true,
		Score:       score,
		SubmittedAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("create practice submission: %v", err)
	}
}

func resetInstanceForAccessMatrix(t *testing.T, env *fullRouterTestEnv, instanceID int64) {
	t.Helper()

	if err := env.db.Model(&model.Instance{}).Where("id = ?", instanceID).Updates(map[string]any{
		"status":       model.InstanceStatusRunning,
		"extend_count": 0,
		"expires_at":   time.Now().Add(time.Hour),
	}).Error; err != nil {
		t.Fatalf("reset instance for access matrix: %v", err)
	}
}

func createReportRecord(t *testing.T, env *fullRouterTestEnv, report model.Report) *model.Report {
	t.Helper()

	if report.CreatedAt.IsZero() {
		report.CreatedAt = time.Now()
	}
	if err := env.db.Create(&report).Error; err != nil {
		t.Fatalf("create report record: %v", err)
	}
	return &report
}

func createDraftChallengeRecord(t *testing.T, env *fullRouterTestEnv, title string) *model.Challenge {
	t.Helper()

	salt, err := flagcrypto.GenerateSalt()
	if err != nil {
		t.Fatalf("generate flag salt: %v", err)
	}

	challenge := &model.Challenge{
		Title:       title,
		Description: "draft challenge for delete matrix",
		Category:    model.DimensionWeb,
		Difficulty:  model.ChallengeDifficultyEasy,
		Points:      90,
		ImageID:     env.image.ID,
		Status:      model.ChallengeStatusDraft,
		FlagType:    model.FlagTypeStatic,
		FlagSalt:    salt,
		FlagHash:    flagcrypto.HashStaticFlag("flag{draft}", salt),
		FlagPrefix:  "flag",
	}
	if err := env.db.Create(challenge).Error; err != nil {
		t.Fatalf("create draft challenge: %v", err)
	}
	return challenge
}

func createRunningInstanceForChallenge(t *testing.T, env *fullRouterTestEnv, challengeID, userID int64) {
	t.Helper()

	instance := &model.Instance{
		UserID:      userID,
		ChallengeID: challengeID,
		Status:      model.InstanceStatusRunning,
		ContainerID: fmt.Sprintf("instance-%d", time.Now().UnixNano()),
		NetworkID:   "matrix-running-network",
		AccessURL:   "http://127.0.0.1:30002",
		Nonce:       "matrix-running-nonce",
		ExpiresAt:   time.Now().Add(time.Hour),
		MaxExtends:  2,
	}
	if err := env.db.Create(instance).Error; err != nil {
		t.Fatalf("create running instance: %v", err)
	}
}

func stopInstancesForChallenge(t *testing.T, env *fullRouterTestEnv, challengeID int64) {
	t.Helper()

	if err := env.db.Model(&model.Instance{}).
		Where("challenge_id = ?", challengeID).
		Updates(map[string]any{
			"status":     model.InstanceStatusStopped,
			"updated_at": time.Now(),
		}).Error; err != nil {
		t.Fatalf("stop instances for challenge: %v", err)
	}
}

func setContestStatus(t *testing.T, env *fullRouterTestEnv, contestID int64, status string, freezeTime *time.Time) {
	t.Helper()

	updates := map[string]any{
		"status":     status,
		"updated_at": time.Now(),
	}
	if freezeTime != nil {
		updates["freeze_time"] = freezeTime
	}
	if err := env.db.Model(&model.Contest{}).Where("id = ?", contestID).Updates(updates).Error; err != nil {
		t.Fatalf("set contest status: %v", err)
	}
}

func seedContestScore(t *testing.T, env *fullRouterTestEnv, contestID, teamID int64, score float64) {
	t.Helper()

	if err := env.cache.ZAdd(context.Background(), rediskeys.RankContestTeamKey(contestID), redislib.Z{
		Score:  score,
		Member: fmt.Sprintf("%d", teamID),
	}).Err(); err != nil {
		t.Fatalf("seed contest score: %v", err)
	}
}

func createRecommendationChallenge(t *testing.T, env *fullRouterTestEnv, title, category string) *model.Challenge {
	t.Helper()

	salt, err := flagcrypto.GenerateSalt()
	if err != nil {
		t.Fatalf("generate flag salt: %v", err)
	}

	challenge := &model.Challenge{
		Title:       title,
		Description: "recommendation challenge",
		Category:    category,
		Difficulty:  model.ChallengeDifficultyEasy,
		Points:      150,
		ImageID:     env.image.ID,
		Status:      model.ChallengeStatusPublished,
		FlagType:    model.FlagTypeStatic,
		FlagSalt:    salt,
		FlagHash:    flagcrypto.HashStaticFlag("flag{recommend}", salt),
		FlagPrefix:  "flag",
	}
	if err := env.db.Create(challenge).Error; err != nil {
		t.Fatalf("create recommendation challenge: %v", err)
	}
	return challenge
}
