package app

import (
	"archive/zip"
	"bytes"
	"compress/zlib"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"ctf-platform/internal/app/composition"
	assessmentcmd "ctf-platform/internal/module/assessment/application/commands"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	rediskeys "ctf-platform/internal/module/contest/infrastructure/cachekeys"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	practicehttp "ctf-platform/internal/module/practice/api/http"
	teachinghttp "ctf-platform/internal/module/teaching_query/api/http"
	"ctf-platform/internal/platform/randomstring"
	flagcrypto "ctf-platform/internal/shared/flagcrypto"
	"ctf-platform/internal/shared/taxonomy"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	xws "golang.org/x/net/websocket"
)

type fullRouterEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func TestTeacherRoutesAreServedByTeachingQuery(t *testing.T) {
	cfg, db, cache := newAppTestDependencies(t)

	originalBuildTeachingQueryModule := buildTeachingQueryModule
	t.Cleanup(func() {
		buildTeachingQueryModule = originalBuildTeachingQueryModule
	})

	called := false
	buildTeachingQueryModule = func(root *composition.Root, assessment *composition.AssessmentModule, identity *composition.IdentityModule) *composition.TeachingQueryModule {
		module := originalBuildTeachingQueryModule(root, assessment, identity)
		called = true
		if module == nil || module.Handler == nil {
			t.Fatal("expected teaching query module handler")
		}
		if got, want := reflect.TypeOf(module.Handler), reflect.TypeOf(&teachinghttp.Handler{}); got != want {
			t.Fatalf("teaching query handler type = %v, want %v", got, want)
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
		t.Fatal("expected teaching query module builder to be called")
	}
}

func TestStudentPracticeReadRoutesAreServedByPracticeModule(t *testing.T) {
	cfg, db, cache := newAppTestDependencies(t)

	originalBuildPracticeModule := buildPracticeModule
	t.Cleanup(func() {
		buildPracticeModule = originalBuildPracticeModule
	})

	called := false
	buildPracticeModule = func(root *composition.Root, challenge *composition.ChallengeModule, instance *composition.InstanceModule) *composition.PracticeModule {
		module := originalBuildPracticeModule(root, challenge, instance)
		called = true
		if module == nil || module.Handler == nil {
			t.Fatal("expected practice module handler")
		}
		if got, want := reflect.TypeOf(module.Handler), reflect.TypeOf(&practicehttp.Handler{}); got != want {
			t.Fatalf("practice handler type = %v, want %v", got, want)
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
		t.Fatal("expected practice module builder to be called")
	}
}

func TestFullRouter_ReportPreviewAndDownloadStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	adminHeaders := bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd))
	teacherHeaders := bearerHeaders(loginForToken(t, env.router, env.teacher.Username, env.teacherPwd))
	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd))

	resp := performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/reports/personal", map[string]any{
		"format": assessmententity.ReportFormatExcel,
	}, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var personalReport assessmentcmd.ReportExportData
	decodeFullRouterData(t, resp, &personalReport)
	if personalReport.Status != assessmententity.ReportStatusReady || personalReport.DownloadURL == nil {
		t.Fatalf("expected ready personal report with download url, got %+v", personalReport)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", personalReport.ReportID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var personalStatus assessmentcmd.ReportExportData
	decodeFullRouterData(t, resp, &personalStatus)
	if personalStatus.Status != assessmententity.ReportStatusReady || personalStatus.DownloadURL == nil {
		t.Fatalf("expected ready personal report status, got %+v", personalStatus)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/download", personalReport.ReportID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
	if contentType := resp.Header().Get("Content-Type"); contentType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Fatalf("expected xlsx content-type, got %q", contentType)
	}

	processingReport := createReportRecord(t, env, assessmententity.Report{
		Type:   assessmententity.ReportTypePersonal,
		Format: assessmententity.ReportFormatPDF,
		UserID: &env.student.ID,
		Status: assessmententity.ReportStatusProcessing,
	})

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", processingReport.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var processingStatus assessmentcmd.ReportExportData
	decodeFullRouterData(t, resp, &processingStatus)
	if processingStatus.Status != assessmententity.ReportStatusProcessing {
		t.Fatalf("expected processing status, got %+v", processingStatus)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/download", processingReport.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusConflict)

	failedMessage := "generation failed in matrix"
	failedReport := createReportRecord(t, env, assessmententity.Report{
		Type:     assessmententity.ReportTypePersonal,
		Format:   assessmententity.ReportFormatPDF,
		UserID:   &env.student.ID,
		Status:   assessmententity.ReportStatusFailed,
		ErrorMsg: &failedMessage,
	})

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", failedReport.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var failedStatus assessmentcmd.ReportExportData
	decodeFullRouterData(t, resp, &failedStatus)
	if failedStatus.Status != assessmententity.ReportStatusFailed || failedStatus.ErrorMessage == nil || *failedStatus.ErrorMessage != failedMessage {
		t.Fatalf("expected failed status with message, got %+v", failedStatus)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/download", failedReport.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusConflict)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/reports/class", map[string]any{
		"class_name": env.otherStudent.ClassName,
		"format":     assessmententity.ReportFormatPDF,
	}, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/reports/class", map[string]any{
		"class_name": env.className,
		"format":     assessmententity.ReportFormatPDF,
	}, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var classReport assessmentcmd.ReportExportData
	decodeFullRouterData(t, resp, &classReport)
	if classReport.Status != assessmententity.ReportStatusProcessing {
		t.Fatalf("expected class report to start in processing state, got %+v", classReport)
	}

	classReady := waitForReportStatus(t, env, classReport.ReportID, teacherHeaders, assessmententity.ReportStatusReady, 5*time.Second)
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

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/export", env.contest.ID), map[string]any{
		"format": assessmententity.ReportFormatJSON,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var contestExport assessmentcmd.ReportExportData
	decodeFullRouterData(t, resp, &contestExport)
	if contestExport.Status != assessmententity.ReportStatusProcessing {
		t.Fatalf("expected contest export processing status, got %+v", contestExport)
	}

	contestExportReady := waitForReportStatus(t, env, contestExport.ReportID, adminHeaders, assessmententity.ReportStatusReady, 5*time.Second)
	if contestExportReady.DownloadURL == nil {
		t.Fatalf("expected contest export download url, got %+v", contestExportReady)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/download", contestExport.ReportID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
	if contentType := resp.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected json content-type, got %q", contentType)
	}
	if !strings.Contains(resp.Body.String(), "\"contest\"") {
		t.Fatalf("expected contest export json payload, got %s", resp.Body.String())
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/teacher/students/%d/review-archive/export", env.otherStudent.ID), map[string]any{
		"format": assessmententity.ReportFormatJSON,
	}, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/teacher/students/%d/review-archive/export", env.student.ID), map[string]any{
		"format": assessmententity.ReportFormatJSON,
	}, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var reviewArchive assessmentcmd.ReportExportData
	decodeFullRouterData(t, resp, &reviewArchive)
	if reviewArchive.Status != assessmententity.ReportStatusProcessing {
		t.Fatalf("expected review archive processing status, got %+v", reviewArchive)
	}

	reviewArchiveReady := waitForReportStatus(t, env, reviewArchive.ReportID, teacherHeaders, assessmententity.ReportStatusReady, 5*time.Second)
	if reviewArchiveReady.DownloadURL == nil {
		t.Fatalf("expected review archive download url, got %+v", reviewArchiveReady)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/download", reviewArchive.ReportID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
	if contentType := resp.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected json content-type, got %q", contentType)
	}
	if !strings.Contains(resp.Body.String(), "\"manual_reviews\"") {
		t.Fatalf("expected review archive json payload, got %s", resp.Body.String())
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", reviewArchive.ReportID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)
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

func readFullRouterZIPEntry(t *testing.T, archive *zip.Reader, name string) []byte {
	t.Helper()
	for _, file := range archive.File {
		if file.Name != name {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatalf("open zip entry %s: %v", name, err)
		}
		defer reader.Close()

		content, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("read zip entry %s: %v", name, err)
		}
		return content
	}
	t.Fatalf("zip entry %s not found", name)
	return nil
}

func fullRouterPDFContainsText(content []byte, token string) bool {
	needle := []byte(token)
	if bytes.Contains(content, needle) {
		return true
	}

	for pos := 0; pos < len(content); {
		idx := bytes.Index(content[pos:], []byte("stream"))
		if idx < 0 {
			return false
		}
		start := pos + idx + len("stream")
		for start < len(content) && (content[start] == '\n' || content[start] == '\r' || content[start] == ' ') {
			start++
		}

		endOffset := bytes.Index(content[start:], []byte("endstream"))
		if endOffset < 0 {
			return false
		}
		streamData := bytes.TrimRight(content[start:start+endOffset], "\r\n")
		reader, err := zlib.NewReader(bytes.NewReader(streamData))
		if err == nil {
			decoded, readErr := io.ReadAll(reader)
			reader.Close()
			if readErr == nil && bytes.Contains(decoded, needle) {
				return true
			}
		}
		pos = start + endOffset + len("endstream")
	}

	return false
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

func waitForReportStatus(t *testing.T, env *fullRouterTestEnv, reportID int64, headers map[string]string, wantStatus string, timeout time.Duration) *assessmentcmd.ReportExportData {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp := performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", reportID), nil, headers)
		if resp.Code != http.StatusOK {
			t.Fatalf("unexpected report status response: %d body=%s", resp.Code, resp.Body.String())
		}

		var report assessmentcmd.ReportExportData
		decodeFullRouterData(t, resp, &report)
		if report.Status == wantStatus {
			return &report
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for report %d status %s", reportID, wantStatus)
	return nil
}

func createFullRouterContest(t *testing.T, env *fullRouterTestEnv, title, status string) *contestcontracts.Contest {
	t.Helper()

	now := time.Now()
	contest := &contestcontracts.Contest{
		Title:       title,
		Description: "state matrix contest",
		Mode:        contestcontracts.ContestModeJeopardy,
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

func createContestRegistration(t *testing.T, env *fullRouterTestEnv, contestID, userID int64, status string, teamID *int64) *contestcontracts.ContestRegistration {
	t.Helper()

	registration := &contestcontracts.ContestRegistration{
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

func findContestRegistration(t *testing.T, env *fullRouterTestEnv, contestID, userID int64) *contestcontracts.ContestRegistration {
	t.Helper()

	var registration contestcontracts.ContestRegistration
	if err := env.db.Where("contest_id = ? AND user_id = ?", contestID, userID).First(&registration).Error; err != nil {
		t.Fatalf("find contest registration: %v", err)
	}
	return &registration
}

func createContestTeam(t *testing.T, env *fullRouterTestEnv, contestID, captainID int64, name string, maxMembers int) *contestcontracts.Team {
	t.Helper()

	team := &contestcontracts.Team{
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
	if err := env.db.Model(&contestcontracts.ContestRegistration{}).
		Where("contest_id = ? AND user_id = ?", contestID, captainID).
		Updates(map[string]any{"team_id": team.ID, "updated_at": time.Now()}).Error; err != nil {
		t.Fatalf("bind captain registration to team: %v", err)
	}
	return team
}

func createContestTeamMember(t *testing.T, env *fullRouterTestEnv, contestID, teamID, userID int64) {
	t.Helper()

	if err := env.db.Create(&contestcontracts.TeamMember{
		ContestID: contestID,
		TeamID:    teamID,
		UserID:    userID,
		JoinedAt:  time.Now(),
	}).Error; err != nil {
		t.Fatalf("create contest team member: %v", err)
	}
	if err := env.db.Model(&contestcontracts.ContestRegistration{}).
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		Updates(map[string]any{"team_id": teamID, "updated_at": time.Now()}).Error; err != nil {
		t.Fatalf("bind member registration to team: %v", err)
	}
}

func createContestSubmission(t *testing.T, env *fullRouterTestEnv, contestID, teamID, userID, challengeID int64, score int) {
	t.Helper()

	if err := env.db.Create(&contestcontracts.Submission{
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

	if err := env.db.Create(&contestcontracts.Submission{
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

	if err := env.db.Model(&instancecontracts.Instance{}).Where("id = ?", instanceID).Updates(map[string]any{
		"status":       instancecontracts.InstanceStatusRunning,
		"extend_count": 0,
		"expires_at":   time.Now().Add(time.Hour),
	}).Error; err != nil {
		t.Fatalf("reset instance for access matrix: %v", err)
	}
}

func createReportRecord(t *testing.T, env *fullRouterTestEnv, report assessmententity.Report) *assessmententity.Report {
	t.Helper()

	if report.CreatedAt.IsZero() {
		report.CreatedAt = time.Now()
	}
	if err := env.db.Create(&report).Error; err != nil {
		t.Fatalf("create report record: %v", err)
	}
	return &report
}

func createDraftChallengeRecord(t *testing.T, env *fullRouterTestEnv, title string) *appChallengeRow {
	t.Helper()

	salt, err := randomstring.Generate()
	if err != nil {
		t.Fatalf("generate flag salt: %v", err)
	}

	challenge := &appChallengeRow{
		Title:       title,
		Description: "draft challenge for delete matrix",
		Category:    taxonomy.DimensionWeb,
		Difficulty:  taxonomy.DifficultyEasy,
		Points:      90,
		ImageID:     env.image.ID,
		Status:      challengecontracts.ChallengeStatusDraft,
		FlagType:    challengecontracts.FlagTypeStatic,
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

	instance := &instancecontracts.Instance{
		UserID:      userID,
		ChallengeID: challengeID,
		Status:      instancecontracts.InstanceStatusRunning,
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

	if err := env.db.Model(&instancecontracts.Instance{}).
		Where("challenge_id = ?", challengeID).
		Updates(map[string]any{
			"status":     instancecontracts.InstanceStatusStopped,
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
	if err := env.db.Model(&contestcontracts.Contest{}).Where("id = ?", contestID).Updates(updates).Error; err != nil {
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

func createRecommendationChallenge(t *testing.T, env *fullRouterTestEnv, title, category string) *appChallengeRow {
	t.Helper()

	salt, err := randomstring.Generate()
	if err != nil {
		t.Fatalf("generate flag salt: %v", err)
	}

	challenge := &appChallengeRow{
		Title:       title,
		Description: "recommendation challenge",
		Category:    category,
		Difficulty:  taxonomy.DifficultyEasy,
		Points:      150,
		ImageID:     env.image.ID,
		Status:      challengecontracts.ChallengeStatusPublished,
		FlagType:    challengecontracts.FlagTypeStatic,
		FlagSalt:    salt,
		FlagHash:    flagcrypto.HashStaticFlag("flag{recommend}", salt),
		FlagPrefix:  "flag",
	}
	if err := env.db.Create(challenge).Error; err != nil {
		t.Fatalf("create recommendation challenge: %v", err)
	}
	return challenge
}
