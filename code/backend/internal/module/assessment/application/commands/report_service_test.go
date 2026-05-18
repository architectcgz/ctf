package commands

import (
	"archive/zip"
	"bytes"
	"compress/zlib"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	queryports "ctf-platform/internal/module/teaching_query/ports"
	teachingadvice "ctf-platform/internal/teaching/advice"
	"ctf-platform/internal/teaching/classwindow"
	"ctf-platform/internal/teaching/evidence"
	"ctf-platform/pkg/errcode"
)

type testReportRepository struct {
	db                *gorm.DB
	users             map[int64]*assessmentdomain.ReportUser
	contests          map[int64]*contestcontracts.Contest
	personalStats     *assessmentdomain.PersonalReportStats
	totalChallenges   int64
	classSummary      *queryports.ClassSummary
	classTrend        *queryports.ClassTrend
	classSnapshots    []teachingadvice.StudentFactSnapshot
	categoryStats     []assessmentdomain.ClassDistributionStat
	difficultyStats   []assessmentdomain.ClassDistributionStat
	contestSummary    *assessmentdomain.ClassContestMigrationSummary
	lastSummarySince  time.Time
	lastTrendSince    time.Time
	lastTrendDays     int
	lastSnapshotSince time.Time
	timeline          []assessmentdomain.ReviewArchiveTimelineEvent
	evidence          []assessmentdomain.ReviewArchiveEvidenceEvent
	writeups          []assessmentdomain.ReviewArchiveWriteupItem
	manualReviews     []assessmentdomain.ReviewArchiveManualReviewItem
}

func (r *testReportRepository) Create(ctx context.Context, report *assessmententity.Report) error {
	if r == nil || r.db == nil {
		return nil
	}
	return r.db.WithContext(ctx).Create(report).Error
}

func (r *testReportRepository) FindByID(context.Context, int64) (*assessmententity.Report, error) {
	return nil, assessmentports.ErrAssessmentReportNotFound
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
			return nil, assessmentports.ErrAssessmentReportNotFound
		}
		return user, nil
	}
	if r == nil || r.db == nil {
		return nil, assessmentports.ErrAssessmentReportNotFound
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
		return nil, assessmentports.ErrAssessmentReportNotFound
	}
	return &user, nil
}

func (r *testReportRepository) FindContestByID(ctx context.Context, contestID int64) (*contestcontracts.Contest, error) {
	if r != nil && r.contests != nil {
		contest, ok := r.contests[contestID]
		if !ok {
			return nil, assessmentports.ErrAssessmentContestNotFound
		}
		return contest, nil
	}
	return nil, assessmentports.ErrAssessmentContestNotFound
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

func (r *testReportRepository) ListClassCategoryDistribution(context.Context, string) ([]assessmentdomain.ClassDistributionStat, error) {
	if r != nil && r.categoryStats != nil {
		return r.categoryStats, nil
	}
	return []assessmentdomain.ClassDistributionStat{}, nil
}

func (r *testReportRepository) ListClassDifficultyDistribution(context.Context, string) ([]assessmentdomain.ClassDistributionStat, error) {
	if r != nil && r.difficultyStats != nil {
		return r.difficultyStats, nil
	}
	return []assessmentdomain.ClassDistributionStat{}, nil
}

func (r *testReportRepository) GetClassContestMigrationSummary(context.Context, string) (*assessmentdomain.ClassContestMigrationSummary, error) {
	if r != nil && r.contestSummary != nil {
		return r.contestSummary, nil
	}
	return &assessmentdomain.ClassContestMigrationSummary{}, nil
}

func (r *testReportRepository) GetClassSummary(_ context.Context, _ string, since time.Time) (*queryports.ClassSummary, error) {
	if r != nil {
		r.lastSummarySince = since
	}
	if r != nil && r.classSummary != nil {
		return r.classSummary, nil
	}
	return &queryports.ClassSummary{}, nil
}

func (r *testReportRepository) GetClassTrend(_ context.Context, _ string, since time.Time, days int) (*queryports.ClassTrend, error) {
	if r != nil {
		r.lastTrendSince = since
		r.lastTrendDays = days
	}
	if r != nil && r.classTrend != nil {
		return r.classTrend, nil
	}
	return &queryports.ClassTrend{}, nil
}

func (r *testReportRepository) ListClassTeachingFactSnapshots(_ context.Context, _ string, since time.Time) ([]teachingadvice.StudentFactSnapshot, error) {
	if r != nil {
		r.lastSnapshotSince = since
	}
	if r != nil && r.classSnapshots != nil {
		return r.classSnapshots, nil
	}
	return []teachingadvice.StudentFactSnapshot{}, nil
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

func (r *testReportRepository) GetStudentEvidence(context.Context, int64, evidence.Query) ([]assessmentdomain.ReviewArchiveEvidenceEvent, error) {
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
	resp *assessmentcontracts.SkillProfile
}

func (r *testAssessmentProfileReader) GetSkillProfile(context.Context, int64) (*assessmentcontracts.SkillProfile, error) {
	if r == nil || r.resp == nil {
		return &assessmentcontracts.SkillProfile{}, nil
	}
	return r.resp, nil
}

type testAWDReviewExportBuilder struct {
	wait    <-chan struct{}
	archive *assessmentqry.TeacherAWDReviewArchiveResp
}

func (b *testAWDReviewExportBuilder) BuildArchive(ctx context.Context, requesterID, contestID int64, roundNumber *int) (*assessmentqry.TeacherAWDReviewArchiveResp, error) {
	if b != nil && b.wait != nil {
		select {
		case <-b.wait:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if b != nil && b.archive != nil {
		return b.archive, nil
	}

	generatedAt := time.Date(2026, 4, 12, 12, 0, 0, 0, time.UTC)
	selectedRoundNumber := 0
	if roundNumber != nil {
		selectedRoundNumber = *roundNumber
	}
	if selectedRoundNumber <= 0 {
		selectedRoundNumber = 1
	}

	return &assessmentqry.TeacherAWDReviewArchiveResp{
		GeneratedAt: generatedAt,
		Scope: assessmentqry.TeacherAWDReviewScopeResp{
			SnapshotType: "final",
			RequestedBy:  requesterID,
			RequestedID:  contestID,
		},
		Contest: assessmentqry.TeacherAWDReviewContestMetaResp{
			ID:         contestID,
			Title:      "awd-review",
			Mode:       contestcontracts.ContestModeAWD,
			Status:     contestcontracts.ContestStatusEnded,
			RoundCount: 1,
			TeamCount:  1,
		},
		Overview: &assessmentqry.TeacherAWDReviewOverviewResp{
			RoundCount:   1,
			TeamCount:    1,
			ServiceCount: 1,
			AttackCount:  1,
			TrafficCount: 1,
		},
		Rounds: []assessmentqry.TeacherAWDReviewRoundResp{{
			ID:           1,
			ContestID:    contestID,
			RoundNumber:  selectedRoundNumber,
			Status:       contestcontracts.AWDRoundStatusFinished,
			ServiceCount: 1,
			AttackCount:  1,
			TrafficCount: 1,
		}},
		SelectedRound: &assessmentqry.TeacherAWDSelectedRoundResp{
			Round: assessmentqry.TeacherAWDReviewRoundResp{
				ID:           1,
				ContestID:    contestID,
				RoundNumber:  selectedRoundNumber,
				Status:       contestcontracts.AWDRoundStatusFinished,
				ServiceCount: 1,
				AttackCount:  1,
				TrafficCount: 1,
			},
			Teams: []assessmentqry.TeacherAWDReviewTeamResp{{
				TeamID:      1,
				TeamName:    "blue",
				CaptainID:   1,
				TotalScore:  100,
				MemberCount: 1,
			}},
			Services: []assessmentqry.TeacherAWDReviewServiceResp{{
				ID:                1,
				RoundID:           1,
				TeamID:            1,
				TeamName:          "blue",
				AWDChallengeID:    1,
				AWDChallengeTitle: "web",
				ServiceStatus:     contestcontracts.AWDServiceStatusUp,
			}},
			Attacks: []assessmentqry.TeacherAWDReviewAttackResp{{
				ID:                1,
				RoundID:           1,
				AttackerTeamID:    1,
				AttackerTeamName:  "blue",
				VictimTeamID:      2,
				VictimTeamName:    "red",
				AWDChallengeID:    1,
				AWDChallengeTitle: "web",
				AttackType:        contestcontracts.AWDAttackTypeFlagCapture,
				Source:            contestcontracts.AWDAttackSourceManual,
			}},
			Traffic: []assessmentqry.TeacherAWDReviewTrafficResp{{
				ID:                1,
				ContestID:         contestID,
				RoundID:           1,
				AttackerTeamID:    1,
				AttackerTeamName:  "blue",
				VictimTeamID:      2,
				VictimTeamName:    "red",
				AWDChallengeID:    1,
				AWDChallengeTitle: "web",
				Method:            "GET",
				Path:              "/health",
				StatusCode:        200,
				Source:            contestcontracts.AWDAttackSourceSubmission,
			}},
		},
	}, nil
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

func readTestZIPEntry(t *testing.T, archive *zip.Reader, name string) []byte {
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

func pdfContainsText(content []byte, token string) bool {
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

func findObservation(items []assessmentdomain.ReviewArchiveObservation, code string) *assessmentdomain.ReviewArchiveObservation {
	for index := range items {
		if items[index].Code == code {
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
		SkillProfile: []*assessmentcontracts.SkillDimension{
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
		SkillProfile: []*assessmentcontracts.SkillDimension{
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

	report := &assessmententity.Report{
		ID:     7,
		Type:   assessmententity.ReportTypeClass,
		Format: assessmententity.ReportFormatExcel,
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

func TestWriteJSONReportPreservesSkillProfileFieldNames(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "archive.json")
	payload := ReviewArchiveData{
		GeneratedAt: time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC),
		Student: ReviewArchiveStudent{
			ID:       7,
			Username: "alice",
		},
		SkillProfile: []*assessmentcontracts.SkillDimension{
			{Dimension: "web", Score: 0.8},
		},
	}

	if err := writeJSONReport(path, payload); err != nil {
		t.Fatalf("writeJSONReport() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !bytes.Contains(content, []byte(`"skill_profile"`)) {
		t.Fatalf("expected skill_profile key, got %s", string(content))
	}

	var decoded map[string]any
	if err := json.Unmarshal(content, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	profiles, ok := decoded["skill_profile"].([]any)
	if !ok || len(profiles) != 1 {
		t.Fatalf("expected one skill_profile item, got %#v", decoded["skill_profile"])
	}
	first, ok := profiles[0].(map[string]any)
	if !ok {
		t.Fatalf("expected skill_profile item object, got %#v", profiles[0])
	}
	if first["dimension"] != "web" {
		t.Fatalf("expected dimension key to stay dimension=web, got %#v", first)
	}
	if score, ok := first["score"].(float64); !ok || score != 0.8 {
		t.Fatalf("expected score key to stay score=0.8, got %#v", first)
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
				Category:    "web",
				Title:       "web-1",
				Timestamp:   submittedAt.Add(1 * time.Minute),
				Detail:      "访问攻击目标",
				Meta:        map[string]any{"event_stage": "access"},
			},
			{
				Type:        "instance_proxy_request",
				ChallengeID: 11,
				Category:    "web",
				Title:       "web-1",
				Timestamp:   submittedAt.Add(2 * time.Minute),
				Detail:      "经平台代理发起 POST /login",
				Meta:        map[string]any{"event_stage": "exploit", "method": "POST"},
			},
			{
				Type:        "challenge_hint_unlock",
				ChallengeID: 11,
				Category:    "web",
				Title:       "web-1",
				Timestamp:   submittedAt,
				Detail:      "解锁第 1 级提示",
				Meta:        map[string]any{"event_stage": "analysis"},
			},
			{
				Type:        "challenge_submission",
				ChallengeID: 11,
				Category:    "web",
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
				Category:         "web",
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
				Category:       "misc",
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
		repo,
		repo,
		repo,
		repo,
		repo,
		repo,
		repo,
		&testAssessmentProfileReader{
			resp: &assessmentcontracts.SkillProfile{
				UserID: 7,
				Dimensions: []*assessmentcontracts.SkillDimension{
					{Dimension: "web", Score: 0.8},
				},
				UpdatedAt: submittedAt.Format(time.RFC3339),
			},
		},
		config.ReportConfig{
			StorageDir:    t.TempDir(),
			DefaultFormat: assessmententity.ReportFormatPDF,
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
	if closure == nil || closure.Severity != "good" {
		t.Fatalf("expected training_closure observation, got %#v", closure)
	}

	handsOn := findObservation(data.TeacherObservations.Items, "hands_on_depth")
	if handsOn == nil || handsOn.Severity != "good" {
		t.Fatalf("expected hands_on_depth good observation, got %#v", handsOn)
	}
}

func TestBuildReviewArchiveSummaryCountsAWDAttackEvents(t *testing.T) {
	t.Parallel()

	success := true
	now := time.Date(2026, 4, 13, 15, 0, 0, 0, time.UTC)

	summary := buildReviewArchiveSummary(
		6,
		&assessmentdomain.PersonalReportStats{
			TotalScore:    150,
			TotalSolved:   1,
			TotalAttempts: 3,
			Rank:          1,
		},
		[]assessmentdomain.ReviewArchiveTimelineEvent{
			{
				Type:        "awd_attack_submit",
				ChallengeID: 91,
				Title:       "awd-web",
				Timestamp:   now,
				IsCorrect:   &success,
				Points:      intPtr(150),
				Detail:      "AWD 攻击命中 blue-team，得分 150",
			},
		},
		nil,
		nil,
		nil,
	)

	if summary.CorrectSubmissionCount != 1 {
		t.Fatalf("expected AWD timeline event counted as correct submission, got %+v", summary)
	}
	if summary.LastActivityAt == nil || !summary.LastActivityAt.Equal(now) {
		t.Fatalf("expected last activity at %s, got %+v", now, summary.LastActivityAt)
	}
}

func TestBuildReviewArchiveSummaryCombinesTimelineAndEvidenceSuccessesWhenSourcesAreSplit(t *testing.T) {
	t.Parallel()

	success := true
	now := time.Date(2026, 4, 13, 15, 0, 0, 0, time.UTC)

	summary := buildReviewArchiveSummary(
		6,
		&assessmentdomain.PersonalReportStats{
			TotalScore:    150,
			TotalSolved:   1,
			TotalAttempts: 2,
			Rank:          1,
		},
		[]assessmentdomain.ReviewArchiveTimelineEvent{
			{
				Type:        "flag_submit",
				ChallengeID: 11,
				Title:       "web-flag",
				Timestamp:   now,
				IsCorrect:   &success,
			},
		},
		[]assessmentdomain.ReviewArchiveEvidenceEvent{
			{
				Type:      "awd_attack_submission",
				Category:  "web",
				Timestamp: now.Add(time.Minute),
				Meta:      map[string]any{"is_success": true},
			},
		},
		nil,
		nil,
	)

	if summary.CorrectSubmissionCount != 2 {
		t.Fatalf("expected timeline and evidence successes to both count, got %+v", summary)
	}
}

func TestBuildReviewArchiveObservationsTreatsAWDAttacksAsHandsOnEvidence(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 13, 15, 30, 0, 0, time.UTC)
	evidence := []assessmentdomain.ReviewArchiveEvidenceEvent{
		{
			Type:        "awd_attack_submission",
			ChallengeID: 101,
			Category:    "pwn",
			Title:       "awd-pwn",
			Timestamp:   now.Add(-2 * time.Minute),
			Detail:      "AWD 攻击未命中 red-team",
			Meta:        map[string]any{"is_success": false, "event_stage": "exploit"},
		},
		{
			Type:        "awd_attack_submission",
			ChallengeID: 101,
			Category:    "pwn",
			Title:       "awd-pwn",
			Timestamp:   now.Add(-1 * time.Minute),
			Detail:      "AWD 攻击未命中 blue-team",
			Meta:        map[string]any{"is_success": false, "event_stage": "exploit"},
		},
	}

	observations := buildReviewArchiveObservations(
		assessmentdomain.ReviewArchiveSummary{
			TotalAttempts:          2,
			CorrectSubmissionCount: 0,
		},
		[]*assessmentcontracts.SkillDimension{
			{Dimension: "pwn", Score: 0.3},
		},
		nil,
		evidence,
		nil,
		nil,
	)

	if findObservation(observations.Items, "submission_stability") == nil {
		t.Fatalf("expected submission_stability observation from repeated AWD failures, got %+v", observations.Items)
	}
	if findObservation(observations.Items, "hands_on_depth") == nil {
		t.Fatalf("expected hands_on_depth observation from AWD exploit evidence, got %+v", observations.Items)
	}
}

func TestBuildReviewArchiveTeachingFactSnapshotOnlyMarksDimensionWithRealEvidence(t *testing.T) {
	t.Parallel()

	snapshot := buildReviewArchiveTeachingFactSnapshot(
		assessmentdomain.ReviewArchiveSummary{
			TotalAttempts:          2,
			CorrectSubmissionCount: 0,
		},
		[]*assessmentcontracts.SkillDimension{
			{Dimension: "web", Score: 0.28},
			{Dimension: "pwn", Score: 0.24},
		},
		nil,
		[]assessmentdomain.ReviewArchiveEvidenceEvent{
			{
				Type:        "challenge_submission",
				ChallengeID: 11,
				Category:    "web",
				Title:       "web-1",
				Timestamp:   time.Date(2026, 4, 13, 15, 0, 0, 0, time.UTC),
				Detail:      "提交未命中 Flag",
				Meta:        map[string]any{"is_correct": false},
			},
			{
				Type:        "instance_proxy_request",
				ChallengeID: 11,
				Category:    "web",
				Title:       "web-1",
				Timestamp:   time.Date(2026, 4, 13, 15, 2, 0, 0, time.UTC),
				Detail:      "经平台代理发起 POST /login",
				Meta:        map[string]any{"event_stage": "exploit"},
			},
		},
		nil,
		nil,
	)

	evaluation := teachingadvice.EvaluateStudent(snapshot)
	if len(evaluation.WeakDimensions) != 0 {
		t.Fatalf("expected no explicit weak dimensions for sparse archive evidence, got %+v", evaluation.WeakDimensions)
	}
	if len(evaluation.RecommendationTargets) == 0 || evaluation.RecommendationTargets[0].Dimension != "web" {
		t.Fatalf("expected web to remain the primary recommendation target, got %+v", evaluation.RecommendationTargets)
	}
	for _, item := range evaluation.RecommendationTargets {
		if item.Dimension == "pwn" {
			t.Fatalf("expected pwn to stay out of recommendation targets without archive evidence, got %+v", evaluation.RecommendationTargets)
		}
	}
}

func TestBuildReviewArchiveTeachingFactSnapshotUsesExplicitTrackedSubmissionCounts(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC)
	snapshot := buildReviewArchiveTeachingFactSnapshot(
		assessmentdomain.ReviewArchiveSummary{
			TotalAttempts:          2,
			CorrectSubmissionCount: 2,
			LastActivityAt:         &now,
		},
		[]*assessmentcontracts.SkillDimension{
			{Dimension: "web", Score: 0.42},
		},
		nil,
		[]assessmentdomain.ReviewArchiveEvidenceEvent{
			{
				Type:      "challenge_submission",
				Category:  "web",
				Timestamp: now.Add(-4 * time.Minute),
				Meta:      map[string]any{"is_correct": false},
			},
			{
				Type:      "challenge_submission",
				Category:  "web",
				Timestamp: now.Add(-3 * time.Minute),
				Meta:      map[string]any{"is_correct": true},
			},
			{
				Type:      "awd_attack_submission",
				Category:  "web",
				Timestamp: now.Add(-2 * time.Minute),
				Meta:      map[string]any{"is_success": false},
			},
			{
				Type:      "awd_attack_submission",
				Category:  "web",
				Timestamp: now.Add(-1 * time.Minute),
				Meta:      map[string]any{"is_success": true},
			},
		},
		nil,
		nil,
	)

	if snapshot.CorrectSubmissionCount != 2 {
		t.Fatalf("expected 2 successful archive events, got %+v", snapshot)
	}
	if snapshot.WrongSubmissionCount != 2 {
		t.Fatalf("expected explicit tracked failures to be counted, got %+v", snapshot)
	}
	if snapshot.ChallengeSuccessCount != 1 {
		t.Fatalf("expected 1 challenge success, got %+v", snapshot)
	}
	if snapshot.SubmissionSuccessCount != 2 || snapshot.SubmissionFailureCount != 2 {
		t.Fatalf("expected explicit success/failure breakdown, got %+v", snapshot)
	}
	if snapshot.AWDSuccessCount != 1 {
		t.Fatalf("expected 1 awd success, got %+v", snapshot)
	}
}

func TestRecentReviewArchiveActivityStatsUsesEvidenceAndWriteupsWithinSevenDays(t *testing.T) {
	t.Parallel()

	referenceTime := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)
	publishedAt := referenceTime.Add(-2 * time.Hour)

	recentEventCount, activeDays := recentReviewArchiveActivityStats(
		referenceTime,
		[]assessmentdomain.ReviewArchiveTimelineEvent{
			{
				Type:      "challenge_submission",
				Timestamp: referenceTime.Add(-24 * time.Hour),
			},
		},
		[]assessmentdomain.ReviewArchiveEvidenceEvent{
			{
				Type:      "instance_proxy_request",
				Timestamp: referenceTime.Add(-3 * time.Hour),
			},
			{
				Type:      "challenge_submission",
				Timestamp: referenceTime.AddDate(0, 0, -10),
			},
		},
		[]assessmentdomain.ReviewArchiveWriteupItem{
			{
				PublishedAt: &publishedAt,
			},
		},
		nil,
	)

	if recentEventCount != 2 {
		t.Fatalf("expected 2 recent events from evidence/writeup within 7 days, got %d", recentEventCount)
	}
	if activeDays != 1 {
		t.Fatalf("expected 1 active day, got %d", activeDays)
	}
}

func TestBuildReviewArchiveTeachingFactSnapshotCountsRecentManualReviewsAsActivity(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	manualReviews := []assessmentdomain.ReviewArchiveManualReviewItem{
		{SubmittedAt: now.Add(-12 * time.Hour), Category: "web"},
		{SubmittedAt: now.Add(-48 * time.Hour), Category: "web"},
		{SubmittedAt: now.Add(-96 * time.Hour), Category: "web"},
	}

	snapshot := buildReviewArchiveTeachingFactSnapshot(
		assessmentdomain.ReviewArchiveSummary{
			LastActivityAt: &now,
		},
		[]*assessmentcontracts.SkillDimension{
			{Dimension: "web", Score: 0.42},
		},
		nil,
		nil,
		nil,
		manualReviews,
	)

	if snapshot.RecentEventCount7d != 3 {
		t.Fatalf("expected recent manual reviews to count as 3 recent events, got %+v", snapshot)
	}
	if snapshot.ActiveDays7d != 3 {
		t.Fatalf("expected recent manual reviews to span 3 active days, got %+v", snapshot)
	}

	observations := buildReviewArchiveObservations(
		assessmentdomain.ReviewArchiveSummary{
			LastActivityAt: &now,
		},
		[]*assessmentcontracts.SkillDimension{
			{Dimension: "web", Score: 0.42},
		},
		nil,
		nil,
		nil,
		manualReviews,
	)
	if observation := findObservation(observations.Items, "low_activity"); observation != nil {
		t.Fatalf("expected no low_activity observation when recent manual reviews keep the student active, got %+v", observation)
	}
}

func TestReportDownloadFileNameUsesJSONExtension(t *testing.T) {
	t.Parallel()

	report := &assessmententity.Report{
		ID:     9,
		Type:   assessmententity.ReportTypeContestExport,
		Format: assessmententity.ReportFormatJSON,
	}

	if got := reportDownloadFileName(report); got != "contest_export-report-9.json" {
		t.Fatalf("expected json download filename, got %s", got)
	}
}

func TestReportServiceCreateAWDReviewArchiveExportStartsProcessingTask(t *testing.T) {
	t.Parallel()

	db := newTestSQLiteDB(t)
	if err := db.AutoMigrate(&model.User{}, &assessmententity.Report{}); err != nil {
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
	contest := &contestcontracts.Contest{
		ID:        21,
		Title:     "awd-ended",
		Mode:      contestcontracts.ContestModeAWD,
		Status:    contestcontracts.ContestStatusEnded,
		StartTime: time.Date(2026, 4, 12, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 4, 12, 11, 0, 0, 0, time.UTC),
	}
	repo := &testReportRepository{
		db: db,
		contests: map[int64]*contestcontracts.Contest{
			contest.ID: contest,
		},
	}

	service := NewReportService(
		repo,
		repo,
		repo,
		repo,
		repo,
		repo,
		repo,
		repo,
		nil,
		config.ReportConfig{
			StorageDir:    t.TempDir(),
			DefaultFormat: assessmententity.ReportFormatPDF,
			MaxWorkers:    1,
		},
		nil,
	)
	service.StartBackgroundTasks(context.Background())
	releaseBuilder := make(chan struct{})
	service.SetAWDReviewExportBuilder(&testAWDReviewExportBuilder{wait: releaseBuilder})
	t.Cleanup(func() {
		close(releaseBuilder)
		closeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = service.Close(closeCtx)
	})

	resp, err := service.CreateTeacherAWDReviewArchive(context.Background(), teacher.ID, contest.ID, CreateTeacherAWDReviewExportInput{
		RoundNumber: intPtr(2),
	})
	if err != nil {
		t.Fatalf("CreateTeacherAWDReviewArchive() error = %v", err)
	}
	if resp.Status != assessmententity.ReportStatusProcessing {
		t.Fatalf("expected processing status, got %+v", resp)
	}

	var report assessmententity.Report
	if err := db.First(&report, "id = ?", resp.ReportID).Error; err != nil {
		t.Fatalf("load report: %v", err)
	}
	if report.Type != assessmententity.ReportTypeAWDReviewArchive {
		t.Fatalf("expected report type %s, got %+v", assessmententity.ReportTypeAWDReviewArchive, report)
	}
	if report.Format != assessmententity.ReportFormatZIP {
		t.Fatalf("expected report format %s, got %+v", assessmententity.ReportFormatZIP, report)
	}
	if report.UserID == nil || *report.UserID != teacher.ID {
		t.Fatalf("expected report user_id %d, got %+v", teacher.ID, report)
	}
}

func TestRenderAWDReviewArchiveZipPreservesTeacherAWDReviewJSONFields(t *testing.T) {
	t.Parallel()

	archive, err := (&testAWDReviewExportBuilder{}).BuildArchive(context.Background(), 11, 21, intPtr(2))
	if err != nil {
		t.Fatalf("BuildArchive() error = %v", err)
	}

	path := filepath.Join(t.TempDir(), "awd-review.zip")
	if err := RenderAWDReviewArchiveZip(path, archive); err != nil {
		t.Fatalf("RenderAWDReviewArchiveZip() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("zip.NewReader() error = %v", err)
	}

	manifestJSON := readTestZIPEntry(t, reader, "manifest.json")
	if !bytes.Contains(manifestJSON, []byte(`"snapshot_type"`)) || !bytes.Contains(manifestJSON, []byte(`"selected_round"`)) {
		t.Fatalf("expected manifest json tags to be preserved, got %s", manifestJSON)
	}
	var manifest map[string]any
	if err := json.Unmarshal(manifestJSON, &manifest); err != nil {
		t.Fatalf("Unmarshal(manifest.json) error = %v", err)
	}
	if got := int(manifest["contest_id"].(float64)); got != 21 {
		t.Fatalf("expected contest_id=21, got %v", manifest["contest_id"])
	}
	if got := int(manifest["selected_round"].(float64)); got != 2 {
		t.Fatalf("expected selected_round=2, got %v", manifest["selected_round"])
	}

	selectedRoundJSON := readTestZIPEntry(t, reader, "selected-round.json")
	for _, key := range []string{`"round_number"`, `"team_id"`, `"service_status"`, `"attack_type"`, `"status_code"`} {
		if !bytes.Contains(selectedRoundJSON, []byte(key)) {
			t.Fatalf("expected selected-round.json to contain key %s, got %s", key, selectedRoundJSON)
		}
	}
	var selectedRound assessmentqry.TeacherAWDSelectedRoundResp
	if err := json.Unmarshal(selectedRoundJSON, &selectedRound); err != nil {
		t.Fatalf("Unmarshal(selected-round.json) error = %v", err)
	}
	if selectedRound.Round.RoundNumber != 2 {
		t.Fatalf("expected selected round number 2, got %+v", selectedRound.Round)
	}
	if len(selectedRound.Teams) != 1 || selectedRound.Teams[0].TeamID != 1 {
		t.Fatalf("expected selected round teams to be preserved, got %+v", selectedRound.Teams)
	}
}

func TestRenderAWDReviewReportPDFIncludesSelectedRoundSummary(t *testing.T) {
	t.Parallel()

	archive, err := (&testAWDReviewExportBuilder{}).BuildArchive(context.Background(), 11, 21, intPtr(2))
	if err != nil {
		t.Fatalf("BuildArchive() error = %v", err)
	}

	path := filepath.Join(t.TempDir(), "awd-review.pdf")
	if err := RenderAWDReviewReportPDF(path, archive); err != nil {
		t.Fatalf("RenderAWDReviewReportPDF() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(content) < 4 || string(content[:4]) != "%PDF" {
		t.Fatalf("expected PDF header, got %q", string(content[:min(4, len(content))]))
	}
	for _, token := range [][]byte{
		[]byte("Teacher AWD Review Report"),
		[]byte("Selected Round"),
		[]byte("awd-review"),
	} {
		if !pdfContainsText(content, string(token)) {
			t.Fatalf("expected PDF to contain %q", token)
		}
	}
}

func TestReportServiceCreateAWDReviewReportExportRejectsRunningContest(t *testing.T) {
	t.Parallel()

	db := newTestSQLiteDB(t)
	if err := db.AutoMigrate(&model.User{}, &assessmententity.Report{}); err != nil {
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
	contest := &contestcontracts.Contest{
		ID:        22,
		Title:     "awd-running",
		Mode:      contestcontracts.ContestModeAWD,
		Status:    contestcontracts.ContestStatusRunning,
		StartTime: time.Date(2026, 4, 12, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 4, 12, 12, 0, 0, 0, time.UTC),
	}
	repo := &testReportRepository{
		db: db,
		contests: map[int64]*contestcontracts.Contest{
			contest.ID: contest,
		},
	}

	service := NewReportService(
		repo,
		repo,
		repo,
		repo,
		repo,
		repo,
		repo,
		repo,
		nil,
		config.ReportConfig{
			StorageDir:    t.TempDir(),
			DefaultFormat: assessmententity.ReportFormatPDF,
			MaxWorkers:    1,
		},
		nil,
	)

	_, err := service.CreateTeacherAWDReviewReport(context.Background(), teacher.ID, contest.ID, CreateTeacherAWDReviewExportInput{})
	appErr, ok := err.(*errcode.AppError)
	if !ok || appErr.Code != errcode.ErrInvalidParams.Code {
		t.Fatalf("expected invalid params error, got %#v", err)
	}

	var count int64
	if err := db.Model(&assessmententity.Report{}).Count(&count).Error; err != nil {
		t.Fatalf("count reports: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no report rows to be created, got %d", count)
	}
}

func TestReportDownloadFileNameUsesZIPForAWDReviewArchive(t *testing.T) {
	t.Parallel()

	report := &assessmententity.Report{
		ID:     10,
		Type:   assessmententity.ReportTypeAWDReviewArchive,
		Format: assessmententity.ReportFormatJSON,
	}

	if got := reportDownloadFileName(report); got != "awd_review_archive-report-10.zip" {
		t.Fatalf("expected zip download filename, got %s", got)
	}
}

func TestReportDownloadFileNameUsesPDFForAWDReviewReport(t *testing.T) {
	t.Parallel()

	report := &assessmententity.Report{
		ID:     11,
		Type:   assessmententity.ReportTypeAWDReviewReport,
		Format: assessmententity.ReportFormatJSON,
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
	if err := db.AutoMigrate(&model.User{}, &assessmententity.Report{}); err != nil {
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
		&testReportRepository{db: db},
		&testReportRepository{db: db},
		&testReportRepository{db: db},
		&testReportRepository{db: db},
		&testReportRepository{db: db},
		&testReportRepository{db: db},
		&testReportRepository{db: db},
		nil,
		config.ReportConfig{
			StorageDir:    t.TempDir(),
			DefaultFormat: assessmententity.ReportFormatPDF,
			MaxWorkers:    1,
		},
		nil,
	)

	_, err := service.CreateClassReport(context.Background(), teacher.ID, CreateClassReportInput{
		ClassName: "class-b",
		Format:    assessmententity.ReportFormatPDF,
	})
	appErr, ok := err.(*errcode.AppError)
	if !ok || appErr.Code != errcode.ErrForbidden.Code {
		t.Fatalf("expected forbidden error, got %#v", err)
	}

	var count int64
	if err := db.Model(&assessmententity.Report{}).Count(&count).Error; err != nil {
		t.Fatalf("count reports: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no report rows to be created, got %d", count)
	}
}

func TestBuildClassReportDataUsesSharedWindowedClassInsight(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	repo := &testReportRepository{
		classSummary: &queryports.ClassSummary{
			ClassName:          "class-a",
			StudentCount:       2,
			AverageSolved:      2,
			ActiveStudentCount: 1,
			ActiveRate:         50,
			RecentEventCount:   6,
		},
		classTrend: &queryports.ClassTrend{
			ClassName: "class-a",
			Points: []queryports.ClassTrendPoint{
				{Date: "2026-05-01", ActiveStudentCount: 1, EventCount: 2, SolveCount: 1},
				{Date: "2026-05-03", ActiveStudentCount: 1, EventCount: 4, SolveCount: 2},
			},
		},
		classSnapshots: []teachingadvice.StudentFactSnapshot{
			{
				UserID:                 1,
				Username:               "alice",
				ActiveDays7d:           1,
				RecentEventCount7d:     1,
				CorrectSubmissionCount: 0,
				MaxWrongStreak:         4,
				Dimensions: []teachingadvice.DimensionFact{
					{Dimension: "web", ProfileScore: 0.2, AttemptCount: 4, SuccessCount: 0, EvidenceCount: 4},
				},
			},
			{
				UserID:                 2,
				Username:               "bob",
				ActiveDays7d:           2,
				RecentEventCount7d:     2,
				CorrectSubmissionCount: 1,
				WriteupCount:           0,
				Dimensions: []teachingadvice.DimensionFact{
					{Dimension: "web", ProfileScore: 0.3, AttemptCount: 3, SuccessCount: 1, EvidenceCount: 3},
				},
			},
		},
		categoryStats: []assessmentdomain.ClassDistributionStat{
			{Key: "web", TotalChallenges: 12, CoveredChallenges: 3, SolvedStudents: 2},
		},
		difficultyStats: []assessmentdomain.ClassDistributionStat{
			{Key: model.ChallengeDifficultyEasy, TotalChallenges: 8, CoveredChallenges: 2, SolvedStudents: 2},
		},
		contestSummary: &assessmentdomain.ClassContestMigrationSummary{
			ParticipatingStudents: 2,
			SuccessfulStudents:    1,
			AttackCount:           5,
			SuccessCount:          2,
			SuccessDimensions:     []string{"web"},
		},
	}

	service := NewReportService(
		repo,
		repo,
		repo,
		repo,
		repo,
		repo,
		repo,
		repo,
		nil,
		config.ReportConfig{
			StorageDir:    t.TempDir(),
			DefaultFormat: assessmententity.ReportFormatPDF,
			MaxWorkers:    1,
		},
		nil,
	)

	data, err := service.buildClassReportData(context.Background(), "class-a", classwindow.Range{
		FromDate:   "2026-05-01",
		ToDate:     "2026-05-03",
		Days:       3,
		Since:      start,
		StartOfDay: start,
	})
	if err != nil {
		t.Fatalf("buildClassReportData() error = %v", err)
	}

	if data.Window.FromDate != "2026-05-01" || data.Window.ToDate != "2026-05-03" || data.Window.Days != 3 {
		t.Fatalf("unexpected window: %+v", data.Window)
	}
	if !repo.lastSummarySince.Equal(start) || !repo.lastTrendSince.Equal(start) || !repo.lastSnapshotSince.Equal(start) || repo.lastTrendDays != 3 {
		t.Fatalf("unexpected class insight window propagation: summary=%s trend=%s days=%d snapshots=%s", repo.lastSummarySince, repo.lastTrendSince, repo.lastTrendDays, repo.lastSnapshotSince)
	}
	if data.Summary == nil || data.Trend == nil || data.Review == nil {
		t.Fatalf("expected summary/trend/review to be populated, got %+v", data)
	}
	if len(data.Review.Items) == 0 {
		t.Fatalf("expected review items, got %+v", data.Review)
	}
	if len(data.CategoryDistribution) != len(model.AllDimensions) {
		t.Fatalf("expected filled category distribution, got %+v", data.CategoryDistribution)
	}
	if len(data.DifficultyDistribution) != len(assessmentdomain.ClassReportDifficultyOrder()) {
		t.Fatalf("expected filled difficulty distribution, got %+v", data.DifficultyDistribution)
	}
	if data.ContestMigration.SuccessCount != 2 || len(data.ContestMigration.SuccessDimensions) != 1 {
		t.Fatalf("unexpected contest migration summary: %+v", data.ContestMigration)
	}
}

func TestReportServiceCloseCancelsAsyncTasks(t *testing.T) {
	t.Parallel()

	service := NewReportService(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		config.ReportConfig{
			StorageDir:    t.TempDir(),
			DefaultFormat: assessmententity.ReportFormatPDF,
			MaxWorkers:    1,
			ClassTimeout:  time.Minute,
		},
		nil,
	)
	service.StartBackgroundTasks(context.Background())

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

func TestReportServiceCloseRejectsNilContext(t *testing.T) {
	t.Parallel()

	service := NewReportService(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		config.ReportConfig{
			StorageDir:    t.TempDir(),
			DefaultFormat: assessmententity.ReportFormatPDF,
			MaxWorkers:    1,
		},
		nil,
	)

	if err := service.Close(nil); err == nil {
		t.Fatal("expected Close(nil) to reject missing context")
	}
}

func TestCreatePersonalReportRejectsNilContext(t *testing.T) {
	t.Parallel()

	service := NewReportService(
		&testReportRepository{},
		&testReportRepository{},
		&testReportRepository{},
		&testReportRepository{},
		&testReportRepository{},
		&testReportRepository{},
		&testReportRepository{},
		&testReportRepository{},
		nil,
		config.ReportConfig{
			StorageDir:    t.TempDir(),
			DefaultFormat: assessmententity.ReportFormatPDF,
			MaxWorkers:    1,
		},
		nil,
	)

	_, err := service.CreatePersonalReport(nil, 1, CreatePersonalReportInput{Format: assessmententity.ReportFormatPDF})
	if err == nil {
		t.Fatal("expected CreatePersonalReport(nil) to reject missing context")
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
