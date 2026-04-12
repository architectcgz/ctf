package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	"ctf-platform/pkg/errcode"
)

type testReportRepository struct {
	db              *gorm.DB
	users           map[int64]*assessmentdomain.ReportUser
	contests        map[int64]*model.Contest
	personalStats   *assessmentdomain.PersonalReportStats
	totalChallenges int64
	timeline        []assessmentdomain.ReviewArchiveTimelineEvent
	evidence        []assessmentdomain.ReviewArchiveEvidenceEvent
	writeups        []assessmentdomain.ReviewArchiveWriteupItem
	manualReviews   []assessmentdomain.ReviewArchiveManualReviewItem
}

func (r *testReportRepository) Create(ctx context.Context, report *model.Report) error {
	if r == nil || r.db == nil {
		return nil
	}
	return r.db.WithContext(ctx).Create(report).Error
}

func (r *testReportRepository) FindByID(context.Context, int64) (*model.Report, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *testReportRepository) MarkReady(context.Context, int64, string, time.Time) error {
	return nil
}

func (r *testReportRepository) MarkFailed(context.Context, int64, string) error {
	return nil
}

func (r *testReportRepository) FindUserByID(ctx context.Context, userID int64) (*assessmentdomain.ReportUser, error) {
	if r != nil && r.users != nil {
		user, ok := r.users[userID]
		if !ok {
			return nil, gorm.ErrRecordNotFound
		}
		return user, nil
	}
	if r == nil || r.db == nil {
		return nil, gorm.ErrRecordNotFound
	}

	var user assessmentdomain.ReportUser
	if err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Select("id, username, class_name, role").
		Where("id = ? AND deleted_at IS NULL", userID).
		Scan(&user).Error; err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &user, nil
}

func (r *testReportRepository) FindContestByID(ctx context.Context, contestID int64) (*model.Contest, error) {
	if r != nil && r.contests != nil {
		contest, ok := r.contests[contestID]
		if !ok {
			return nil, gorm.ErrRecordNotFound
		}
		return contest, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *testReportRepository) GetPersonalStats(context.Context, int64) (*assessmentdomain.PersonalReportStats, error) {
	if r != nil && r.personalStats != nil {
		return r.personalStats, nil
	}
	return &assessmentdomain.PersonalReportStats{}, nil
}

func (r *testReportRepository) ListPersonalDimensionStats(context.Context, int64) ([]assessmentdomain.ReportDimensionStat, error) {
	return []assessmentdomain.ReportDimensionStat{}, nil
}

func (r *testReportRepository) CountClassStudents(context.Context, string) (int, error) {
	return 0, nil
}

func (r *testReportRepository) GetClassAverageScore(context.Context, string) (float64, error) {
	return 0, nil
}

func (r *testReportRepository) ListClassDimensionAverages(context.Context, string) ([]assessmentdomain.ClassDimensionAverage, error) {
	return []assessmentdomain.ClassDimensionAverage{}, nil
}

func (r *testReportRepository) ListClassTopStudents(context.Context, string, int) ([]assessmentdomain.ClassTopStudent, error) {
	return []assessmentdomain.ClassTopStudent{}, nil
}

func (r *testReportRepository) ListContestScoreboard(context.Context, int64) ([]assessmentdomain.ContestExportScoreboardItem, error) {
	return []assessmentdomain.ContestExportScoreboardItem{}, nil
}

func (r *testReportRepository) ListContestChallenges(context.Context, int64) ([]assessmentdomain.ContestExportChallengeItem, error) {
	return []assessmentdomain.ContestExportChallengeItem{}, nil
}

func (r *testReportRepository) ListContestTeams(context.Context, int64) ([]assessmentdomain.ContestExportTeamItem, error) {
	return []assessmentdomain.ContestExportTeamItem{}, nil
}

func (r *testReportRepository) CountPublishedChallenges(context.Context) (int64, error) {
	if r != nil && r.totalChallenges > 0 {
		return r.totalChallenges, nil
	}
	return 0, nil
}

func (r *testReportRepository) GetStudentTimeline(context.Context, int64, int, int) ([]assessmentdomain.ReviewArchiveTimelineEvent, error) {
	if r != nil && r.timeline != nil {
		return r.timeline, nil
	}
	return []assessmentdomain.ReviewArchiveTimelineEvent{}, nil
}

func (r *testReportRepository) GetStudentEvidence(context.Context, int64, *int64) ([]assessmentdomain.ReviewArchiveEvidenceEvent, error) {
	if r != nil && r.evidence != nil {
		return r.evidence, nil
	}
	return []assessmentdomain.ReviewArchiveEvidenceEvent{}, nil
}

func (r *testReportRepository) ListStudentWriteups(context.Context, int64) ([]assessmentdomain.ReviewArchiveWriteupItem, error) {
	if r != nil && r.writeups != nil {
		return r.writeups, nil
	}
	return []assessmentdomain.ReviewArchiveWriteupItem{}, nil
}

func (r *testReportRepository) ListStudentManualReviews(context.Context, int64) ([]assessmentdomain.ReviewArchiveManualReviewItem, error) {
	if r != nil && r.manualReviews != nil {
		return r.manualReviews, nil
	}
	return []assessmentdomain.ReviewArchiveManualReviewItem{}, nil
}

type testAssessmentProfileReader struct {
	resp *dto.SkillProfileResp
}

func (r *testAssessmentProfileReader) GetSkillProfileWithContext(context.Context, int64) (*dto.SkillProfileResp, error) {
	if r == nil || r.resp == nil {
		return &dto.SkillProfileResp{}, nil
	}
	return r.resp, nil
}

func intPtr(value int) *int {
	return &value
}

func newTestSQLiteDB(t *testing.T) *gorm.DB {
	t.Helper()

	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", name)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return db
}

func findObservation(items []assessmentdomain.ReviewArchiveObservation, key string) *assessmentdomain.ReviewArchiveObservation {
	for index := range items {
		if items[index].Key == key {
			return &items[index]
		}
	}
	return nil
}

func TestWritePersonalPDFCreatesPDFFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "personal-report.pdf")
	data := &personalReportData{
		User: &assessmentdomain.ReportUser{
			ID:        1,
			Username:  "alice",
			ClassName: "class-a",
		},
		SkillProfile: []*dto.SkillDimension{
			{Dimension: "web", Score: 0.8},
			{Dimension: "crypto", Score: 0.5},
		},
		Stats: &assessmentdomain.PersonalReportStats{
			TotalScore:    400,
			TotalSolved:   4,
			TotalAttempts: 7,
			Rank:          2,
		},
		DimensionStats: []assessmentdomain.ReportDimensionStat{
			{Dimension: "web", Solved: 2, Total: 3},
			{Dimension: "crypto", Solved: 1, Total: 2},
		},
	}

	if err := writePersonalPDF(path, data); err != nil {
		t.Fatalf("writePersonalPDF() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(content) < 4 || string(content[:4]) != "%PDF" {
		t.Fatalf("expected PDF header, got %q", string(content[:min(4, len(content))]))
	}
}

func TestWritePersonalExcelCreatesWorkbook(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "personal-report.xlsx")
	data := &personalReportData{
		User: &assessmentdomain.ReportUser{
			ID:        1,
			Username:  "alice",
			ClassName: "class-a",
		},
		SkillProfile: []*dto.SkillDimension{
			{Dimension: "web", Score: 0.8},
			{Dimension: "crypto", Score: 0.5},
		},
		Stats: &assessmentdomain.PersonalReportStats{
			TotalScore:    400,
			TotalSolved:   4,
			TotalAttempts: 7,
			Rank:          2,
		},
		DimensionStats: []assessmentdomain.ReportDimensionStat{
			{Dimension: "web", Solved: 2, Total: 3},
			{Dimension: "crypto", Solved: 1, Total: 2},
		},
	}

	if err := writePersonalExcel(path, data); err != nil {
		t.Fatalf("writePersonalExcel() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(content) < 2 || content[0] != 'P' || content[1] != 'K' {
		t.Fatalf("expected ZIP header, got %q", string(content[:min(2, len(content))]))
	}
}

func TestReportFilePathUsesXLSXExtensionForExcel(t *testing.T) {
	t.Parallel()

	service := &ReportService{
		config: config.ReportConfig{
			StorageDir: t.TempDir(),
		},
	}

	path, err := service.reportFilePath(42, "class", "excel")
	if err != nil {
		t.Fatalf("reportFilePath() error = %v", err)
	}

	if filepath.Ext(path) != ".xlsx" {
		t.Fatalf("expected .xlsx extension, got %s", filepath.Ext(path))
	}
}

func TestReportFileExtension(t *testing.T) {
	t.Parallel()

	if got := reportFileExtension("json"); got != "json" {
		t.Fatalf("expected json extension for json, got %s", got)
	}
	if got := reportFileExtension("excel"); got != "xlsx" {
		t.Fatalf("expected xlsx extension for excel, got %s", got)
	}
	if got := reportFileExtension("pdf"); got != "pdf" {
		t.Fatalf("expected pdf extension for pdf, got %s", got)
	}
	if got := reportFileExtension("json"); got != "json" {
		t.Fatalf("expected json extension for json, got %s", got)
	}
}

func TestReportDownloadFileNameUsesRealExtension(t *testing.T) {
	t.Parallel()

	report := &model.Report{
		ID:     7,
		Type:   model.ReportTypeClass,
		Format: model.ReportFormatExcel,
	}

	if got := reportDownloadFileName(report); got != "class-report-7.xlsx" {
		t.Fatalf("expected xlsx download filename, got %s", got)
	}
}

func TestWriteJSONReportCreatesJSONFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "archive.json")

	if err := writeJSONReport(path, map[string]any{"type": "contest_export", "ok": true}); err != nil {
		t.Fatalf("writeJSONReport() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(content) == 0 || content[0] != '{' {
		t.Fatalf("expected json object content, got %q", string(content))
	}
}

func TestValidateStudentReviewArchiveAccess(t *testing.T) {
	t.Parallel()

	teacher := &assessmentdomain.ReportUser{ID: 1, Role: model.RoleTeacher, ClassName: "class-a"}
	admin := &assessmentdomain.ReportUser{ID: 2, Role: model.RoleAdmin}
	student := &assessmentdomain.ReportUser{ID: 3, Role: model.RoleStudent, ClassName: "class-a"}
	otherStudent := &assessmentdomain.ReportUser{ID: 4, Role: model.RoleStudent, ClassName: "class-b"}

	if err := validateStudentReviewArchiveAccess(teacher, student); err != nil {
		t.Fatalf("expected same-class teacher access, got %v", err)
	}
	if err := validateStudentReviewArchiveAccess(admin, otherStudent); err != nil {
		t.Fatalf("expected admin access, got %v", err)
	}

	err := validateStudentReviewArchiveAccess(teacher, otherStudent)
	appErr, ok := err.(*errcode.AppError)
	if !ok || appErr.Code != errcode.ErrForbidden.Code {
		t.Fatalf("expected forbidden error, got %#v", err)
	}
}

func TestBuildStudentReviewArchiveDataIncludesTeachingObservations(t *testing.T) {
	t.Parallel()

	submittedAt := time.Date(2026, 4, 1, 9, 12, 0, 0, time.UTC)
	reviewedAt := submittedAt.Add(8 * time.Minute)
	lastEventAt := time.Date(2026, 4, 1, 9, 20, 0, 0, time.UTC)
	wrong := false
	correct := true

	repo := &testReportRepository{
		users: map[int64]*assessmentdomain.ReportUser{
			7: {
				ID:        7,
				Username:  "alice",
				Name:      "Alice",
				ClassName: "class-a",
				Role:      model.RoleStudent,
			},
		},
		personalStats: &assessmentdomain.PersonalReportStats{
			TotalScore:    100,
			TotalSolved:   1,
			TotalAttempts: 4,
			Rank:          2,
		},
		totalChallenges: 5,
		timeline: []assessmentdomain.ReviewArchiveTimelineEvent{
			{
				Type:        "hint_unlock",
				ChallengeID: 11,
				Title:       "web-1",
				Timestamp:   submittedAt,
				Detail:      "解锁第 1 级提示",
			},
			{
				Type:        "flag_submit",
				ChallengeID: 11,
				Title:       "web-1",
				Timestamp:   submittedAt.Add(3 * time.Minute),
				IsCorrect:   &wrong,
				Detail:      "提交未命中 Flag",
			},
			{
				Type:        "flag_submit",
				ChallengeID: 11,
				Title:       "web-1",
				Timestamp:   lastEventAt,
				IsCorrect:   &correct,
				Points:      intPtr(100),
				Detail:      "提交命中 Flag",
			},
		},
		evidence: []assessmentdomain.ReviewArchiveEvidenceEvent{
			{
				Type:        "instance_access",
				ChallengeID: 11,
				Title:       "web-1",
				Timestamp:   submittedAt.Add(1 * time.Minute),
				Detail:      "访问攻击目标",
				Meta:        map[string]any{"event_stage": "access"},
			},
			{
				Type:        "instance_proxy_request",
				ChallengeID: 11,
				Title:       "web-1",
				Timestamp:   submittedAt.Add(2 * time.Minute),
				Detail:      "经平台代理发起 POST /login",
				Meta:        map[string]any{"event_stage": "exploit", "method": "POST"},
			},
			{
				Type:        "challenge_hint_unlock",
				ChallengeID: 11,
				Title:       "web-1",
				Timestamp:   submittedAt,
				Detail:      "解锁第 1 级提示",
				Meta:        map[string]any{"event_stage": "analysis"},
			},
			{
				Type:        "challenge_submission",
				ChallengeID: 11,
				Title:       "web-1",
				Timestamp:   lastEventAt,
				Detail:      "提交命中 Flag",
				Meta:        map[string]any{"event_stage": "submit", "is_correct": true, "points": 100},
			},
		},
		writeups: []assessmentdomain.ReviewArchiveWriteupItem{
			{
				ID:               1,
				ChallengeID:      11,
				ChallengeTitle:   "web-1",
				Title:            "从回显到 flag",
				SubmissionStatus: "published",
				VisibilityStatus: "visible",
				IsRecommended:    true,
				PublishedAt:      &submittedAt,
				UpdatedAt:        reviewedAt,
			},
		},
		manualReviews: []assessmentdomain.ReviewArchiveManualReviewItem{
			{
				ID:             2,
				ChallengeID:    12,
				ChallengeTitle: "misc-essay",
				Answer:         "完整答案正文",
				ReviewStatus:   "approved",
				SubmittedAt:    submittedAt,
				ReviewedAt:     &reviewedAt,
				ReviewComment:  "通过",
				Score:          100,
				ReviewerName:   "teacher-a",
			},
		},
	}

	service := NewReportService(
		repo,
		&testAssessmentProfileReader{
			resp: &dto.SkillProfileResp{
				UserID: 7,
				Dimensions: []*dto.SkillDimension{
					{Dimension: "web", Score: 0.8},
				},
				UpdatedAt: submittedAt.Format(time.RFC3339),
			},
		},
		config.ReportConfig{
			StorageDir:    t.TempDir(),
			DefaultFormat: model.ReportFormatPDF,
			MaxWorkers:    1,
		},
		nil,
	)

	data, err := service.buildStudentReviewArchiveData(context.Background(), 7)
	if err != nil {
		t.Fatalf("buildStudentReviewArchiveData() error = %v", err)
	}

	summaryPayload, err := json.Marshal(data.Summary)
	if err != nil {
		t.Fatalf("marshal summary: %v", err)
	}
	if bytes.Contains(summaryPayload, []byte(`"hint_unlock_count"`)) {
		t.Fatalf("expected summary payload to omit hint_unlock_count, got %s", string(summaryPayload))
	}
	if data.Summary.CorrectSubmissionCount != 1 {
		t.Fatalf("expected 1 correct submission, got %d", data.Summary.CorrectSubmissionCount)
	}
	if data.Summary.WriteupCount != 1 {
		t.Fatalf("expected 1 writeup, got %d", data.Summary.WriteupCount)
	}
	if data.Summary.LastActivityAt == nil || !data.Summary.LastActivityAt.Equal(lastEventAt) {
		t.Fatalf("expected last activity at %s, got %#v", lastEventAt, data.Summary.LastActivityAt)
	}

	if len(data.TeacherObservations.Items) == 0 {
		t.Fatal("expected teaching observations to be generated")
	}

	closure := findObservation(data.TeacherObservations.Items, "training_closure")
	if closure == nil || closure.Level != "good" {
		t.Fatalf("expected training_closure observation, got %#v", closure)
	}

	hint := findObservation(data.TeacherObservations.Items, "hint_usage")
	if hint == nil || hint.Level != "good" {
		t.Fatalf("expected hint_usage good observation, got %#v", hint)
	}
}

func TestReportDownloadFileNameUsesJSONExtension(t *testing.T) {
	t.Parallel()

	report := &model.Report{
		ID:     9,
		Type:   model.ReportTypeContestExport,
		Format: model.ReportFormatJSON,
	}

	if got := reportDownloadFileName(report); got != "contest_export-report-9.json" {
		t.Fatalf("expected json download filename, got %s", got)
	}
}

func TestReportServiceCreateAWDReviewArchiveExportStartsProcessingTask(t *testing.T) {
	t.Parallel()

	db := newTestSQLiteDB(t)
	if err := db.AutoMigrate(&model.User{}, &model.Report{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}

	teacher := &model.User{
		ID:        11,
		Username:  "teacher-awd",
		Role:      model.RoleTeacher,
		ClassName: "class-a",
		Status:    model.UserStatusActive,
	}
	if err := db.Create(teacher).Error; err != nil {
		t.Fatalf("seed teacher: %v", err)
	}
	contest := &model.Contest{
		ID:        21,
		Title:     "awd-ended",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusEnded,
		StartTime: time.Date(2026, 4, 12, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 4, 12, 11, 0, 0, 0, time.UTC),
	}
	repo := &testReportRepository{
		db: db,
		contests: map[int64]*model.Contest{
			contest.ID: contest,
		},
	}

	service := NewReportService(
		repo,
		nil,
		config.ReportConfig{
			StorageDir:    t.TempDir(),
			DefaultFormat: model.ReportFormatPDF,
			MaxWorkers:    1,
		},
		nil,
	)
	t.Cleanup(func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = service.Close(closeCtx)
	})

	resp, err := service.CreateTeacherAWDReviewArchive(context.Background(), teacher.ID, contest.ID, &dto.CreateTeacherAWDReviewExportReq{
		RoundNumber: intPtr(2),
	})
	if err != nil {
		t.Fatalf("CreateTeacherAWDReviewArchive() error = %v", err)
	}
	if resp.Status != model.ReportStatusProcessing {
		t.Fatalf("expected processing status, got %+v", resp)
	}

	var report model.Report
	if err := db.First(&report, "id = ?", resp.ReportID).Error; err != nil {
		t.Fatalf("load report: %v", err)
	}
	if report.Type != model.ReportTypeAWDReviewArchive {
		t.Fatalf("expected report type %s, got %+v", model.ReportTypeAWDReviewArchive, report)
	}
	if report.Format != model.ReportFormatZIP {
		t.Fatalf("expected report format %s, got %+v", model.ReportFormatZIP, report)
	}
	if report.UserID == nil || *report.UserID != teacher.ID {
		t.Fatalf("expected report user_id %d, got %+v", teacher.ID, report)
	}
}

func TestReportServiceCreateAWDReviewReportExportRejectsRunningContest(t *testing.T) {
	t.Parallel()

	db := newTestSQLiteDB(t)
	if err := db.AutoMigrate(&model.User{}, &model.Report{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}

	teacher := &model.User{
		ID:        12,
		Username:  "teacher-running",
		Role:      model.RoleTeacher,
		ClassName: "class-a",
		Status:    model.UserStatusActive,
	}
	if err := db.Create(teacher).Error; err != nil {
		t.Fatalf("seed teacher: %v", err)
	}
	contest := &model.Contest{
		ID:        22,
		Title:     "awd-running",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusRunning,
		StartTime: time.Date(2026, 4, 12, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 4, 12, 12, 0, 0, 0, time.UTC),
	}
	repo := &testReportRepository{
		db: db,
		contests: map[int64]*model.Contest{
			contest.ID: contest,
		},
	}

	service := NewReportService(
		repo,
		nil,
		config.ReportConfig{
			StorageDir:    t.TempDir(),
			DefaultFormat: model.ReportFormatPDF,
			MaxWorkers:    1,
		},
		nil,
	)

	_, err := service.CreateTeacherAWDReviewReport(context.Background(), teacher.ID, contest.ID, &dto.CreateTeacherAWDReviewExportReq{})
	appErr, ok := err.(*errcode.AppError)
	if !ok || appErr.Code != errcode.ErrInvalidParams.Code {
		t.Fatalf("expected invalid params error, got %#v", err)
	}

	var count int64
	if err := db.Model(&model.Report{}).Count(&count).Error; err != nil {
		t.Fatalf("count reports: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no report rows to be created, got %d", count)
	}
}

func TestReportDownloadFileNameUsesZIPForAWDReviewArchive(t *testing.T) {
	t.Parallel()

	report := &model.Report{
		ID:     10,
		Type:   model.ReportTypeAWDReviewArchive,
		Format: model.ReportFormatJSON,
	}

	if got := reportDownloadFileName(report); got != "awd_review_archive-report-10.zip" {
		t.Fatalf("expected zip download filename, got %s", got)
	}
}

func TestReportDownloadFileNameUsesPDFForAWDReviewReport(t *testing.T) {
	t.Parallel()

	report := &model.Report{
		ID:     11,
		Type:   model.ReportTypeAWDReviewReport,
		Format: model.ReportFormatJSON,
	}

	if got := reportDownloadFileName(report); got != "awd_review_report-report-11.pdf" {
		t.Fatalf("expected pdf download filename, got %s", got)
	}
}

func TestValidateClassReportAccess(t *testing.T) {
	t.Parallel()

	teacher := &assessmentdomain.ReportUser{ID: 1, Role: model.RoleTeacher, ClassName: "class-a"}
	admin := &assessmentdomain.ReportUser{ID: 2, Role: model.RoleAdmin, ClassName: ""}

	if err := validateClassReportAccess(teacher, "class-a"); err != nil {
		t.Fatalf("expected same-class teacher access, got %v", err)
	}
	if err := validateClassReportAccess(admin, "class-b"); err != nil {
		t.Fatalf("expected admin access, got %v", err)
	}

	err := validateClassReportAccess(teacher, "class-b")
	appErr, ok := err.(*errcode.AppError)
	if !ok || appErr.Code != errcode.ErrForbidden.Code {
		t.Fatalf("expected forbidden error, got %#v", err)
	}
}

func TestCreateClassReportRejectsCrossClassTeacherRequest(t *testing.T) {
	t.Parallel()

	db := newTestSQLiteDB(t)
	if err := db.AutoMigrate(&model.User{}, &model.Report{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}

	teacher := &model.User{
		ID:        1,
		Username:  "teacher-a",
		Role:      model.RoleTeacher,
		ClassName: "class-a",
		Status:    model.UserStatusActive,
	}
	if err := db.Create(teacher).Error; err != nil {
		t.Fatalf("seed teacher: %v", err)
	}

	service := NewReportService(
		&testReportRepository{db: db},
		nil,
		config.ReportConfig{
			StorageDir:    t.TempDir(),
			DefaultFormat: model.ReportFormatPDF,
			MaxWorkers:    1,
		},
		nil,
	)

	_, err = service.CreateClassReport(context.Background(), teacher.ID, &dto.CreateClassReportReq{
		ClassName: "class-b",
		Format:    model.ReportFormatPDF,
	})
	appErr, ok := err.(*errcode.AppError)
	if !ok || appErr.Code != errcode.ErrForbidden.Code {
		t.Fatalf("expected forbidden error, got %#v", err)
	}

	var count int64
	if err := db.Model(&model.Report{}).Count(&count).Error; err != nil {
		t.Fatalf("count reports: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no report rows to be created, got %d", count)
	}
}

func TestReportServiceCloseCancelsAsyncTasks(t *testing.T) {
	t.Parallel()

	service := NewReportService(
		nil,
		nil,
		config.ReportConfig{
			StorageDir:    t.TempDir(),
			DefaultFormat: model.ReportFormatPDF,
			MaxWorkers:    1,
			ClassTimeout:  time.Minute,
		},
		nil,
	)

	var started atomic.Int32
	startedCh := make(chan struct{})
	service.runAsyncReport(1, func(ctx context.Context) error {
		started.Add(1)
		close(startedCh)
		<-ctx.Done()
		return ctx.Err()
	})

	select {
	case <-startedCh:
	case <-time.After(time.Second):
		t.Fatal("expected async task to start")
	}

	deadlineCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := service.Close(deadlineCtx); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if started.Load() != 1 {
		t.Fatalf("expected async task to start once, got %d", started.Load())
	}
}

func TestReportServiceWithPersonalTimeoutUsesConfiguredDeadline(t *testing.T) {
	t.Parallel()

	service := &ReportService{
		config: config.ReportConfig{
			PersonalTimeout: 2 * time.Second,
		},
	}

	ctx, cancel := service.withPersonalTimeout(context.Background())
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected deadline to be set")
	}
	remaining := time.Until(deadline)
	if remaining <= time.Second || remaining > 2*time.Second+200*time.Millisecond {
		t.Fatalf("unexpected remaining timeout: %s", remaining)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
