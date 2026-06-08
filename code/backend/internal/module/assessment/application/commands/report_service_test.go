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
	"unicode/utf16"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	queryports "ctf-platform/internal/module/teaching_query/ports"
	"ctf-platform/internal/shared/taxonomy"
	teachingadvice "ctf-platform/internal/teaching/advice"
	"ctf-platform/internal/teaching/classwindow"
	"ctf-platform/internal/teaching/evidence"
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
		Model(&identitycontracts.User{}).
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

type testAWDReviewArchiveReader struct {
	archives []*assessmentqry.TeacherAWDReviewArchiveResp
	inputs   []assessmentqry.GetTeacherAWDReviewArchiveInput
}

func (r *testAWDReviewArchiveReader) GetContestArchive(ctx context.Context, requesterID, contestID int64, req assessmentqry.GetTeacherAWDReviewArchiveInput) (*assessmentqry.TeacherAWDReviewArchiveResp, error) {
	_ = ctx
	_ = requesterID
	_ = contestID
	r.inputs = append(r.inputs, req)
	if len(r.archives) == 0 {
		return nil, fmt.Errorf("unexpected GetContestArchive call")
	}
	archive := r.archives[0]
	r.archives = r.archives[1:]
	return archive, nil
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
	needleUTF16 := utf16BEBytes(token)
	if bytes.Contains(content, needle) {
		return true
	}
	if len(needleUTF16) > 0 && bytes.Contains(content, needleUTF16) {
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
			if readErr == nil {
				if bytes.Contains(decoded, needle) {
					return true
				}
				if len(needleUTF16) > 0 && bytes.Contains(decoded, needleUTF16) {
					return true
				}
			}
		}
		pos = start + endOffset + len("endstream")
	}

	return false
}

func utf16BEBytes(value string) []byte {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	encoded := utf16.Encode([]rune(value))
	buf := make([]byte, 0, len(encoded)*2)
	for _, code := range encoded {
		buf = append(buf, byte(code>>8), byte(code))
	}
	return buf
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
	if err := db.AutoMigrate(&identitycontracts.User{}, &assessmententity.Report{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}

	teacher := &identitycontracts.User{
		ID:        11,
		Username:  "teacher-awd",
		Role:      identitycontracts.RoleTeacher,
		ClassName: "class-a",
		Status:    identitycontracts.UserStatusActive,
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

func TestTeacherAWDReviewExportBuilderSelectsFocusRoundWhenRoundMissing(t *testing.T) {
	t.Parallel()

	reader := &testAWDReviewArchiveReader{
		archives: []*assessmentqry.TeacherAWDReviewArchiveResp{
			{
				Contest: assessmentqry.TeacherAWDReviewContestMetaResp{
					ID:     21,
					Title:  "awd-review",
					Status: contestcontracts.ContestStatusEnded,
				},
				Rounds: []assessmentqry.TeacherAWDReviewRoundResp{
					{RoundNumber: 1, ServiceCount: 2, AttackCount: 0, TrafficCount: 1},
					{RoundNumber: 2, ServiceCount: 2, AttackCount: 3, TrafficCount: 2},
				},
			},
			{
				Contest: assessmentqry.TeacherAWDReviewContestMetaResp{
					ID:     21,
					Title:  "awd-review",
					Status: contestcontracts.ContestStatusEnded,
				},
				Rounds: []assessmentqry.TeacherAWDReviewRoundResp{
					{RoundNumber: 1, ServiceCount: 2, AttackCount: 0, TrafficCount: 1},
					{RoundNumber: 2, ServiceCount: 2, AttackCount: 3, TrafficCount: 2},
				},
				SelectedRound: &assessmentqry.TeacherAWDSelectedRoundResp{
					Round: assessmentqry.TeacherAWDReviewRoundResp{
						RoundNumber:  2,
						ServiceCount: 2,
						AttackCount:  3,
						TrafficCount: 2,
					},
				},
			},
		},
	}

	builder := NewAWDReviewExportBuilder(reader)
	archive, err := builder.BuildArchive(context.Background(), 11, 21, nil)
	if err != nil {
		t.Fatalf("BuildArchive() error = %v", err)
	}
	if archive == nil || archive.SelectedRound == nil || archive.SelectedRound.Round.RoundNumber != 2 {
		t.Fatalf("expected selected round 2, got %+v", archive)
	}
	if len(reader.inputs) != 2 {
		t.Fatalf("expected 2 archive reads, got %d", len(reader.inputs))
	}
	if reader.inputs[0].RoundNumber != nil {
		t.Fatalf("expected first archive read without round filter, got %+v", reader.inputs[0])
	}
	if reader.inputs[1].RoundNumber == nil || *reader.inputs[1].RoundNumber != 2 {
		t.Fatalf("expected second archive read with round 2, got %+v", reader.inputs[1])
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
		[]byte("awd-review"),
		[]byte("blue"),
		[]byte("red"),
		[]byte("/health"),
	} {
		if !pdfContainsText(content, string(token)) {
			t.Fatalf("expected PDF to contain %q", token)
		}
	}
}

func TestNewReportPDFRegistersBoldFont(t *testing.T) {
	t.Parallel()

	pdf, err := newReportPDF()
	if err != nil {
		t.Fatalf("newReportPDF() error = %v", err)
	}

	setReportPDFFont(pdf, "B", 14)
	if err := pdf.Error(); err != nil {
		t.Fatalf("expected report pdf bold font to be available, got %v", err)
	}
}

func TestHottestRoundPrefersAttackDenseRound(t *testing.T) {
	t.Parallel()

	round := hottestRound([]assessmentqry.TeacherAWDReviewRoundResp{
		{RoundNumber: 1, ServiceCount: 2, AttackCount: 0, TrafficCount: 4},
		{RoundNumber: 2, ServiceCount: 1, AttackCount: 2, TrafficCount: 1},
		{RoundNumber: 3, ServiceCount: 5, AttackCount: 0, TrafficCount: 0},
	})
	if round == nil || round.RoundNumber != 2 {
		t.Fatalf("expected hottest round 2, got %+v", round)
	}
}

func TestTopRiskyServicePrefersCompromisedService(t *testing.T) {
	t.Parallel()

	service := topRiskyService([]assessmentqry.TeacherAWDReviewServiceResp{
		{TeamName: "blue", AWDChallengeTitle: "web", ServiceStatus: contestcontracts.AWDServiceStatusUp, AttackReceived: 4},
		{TeamName: "red", AWDChallengeTitle: "api", ServiceStatus: contestcontracts.AWDServiceStatusCompromised, AttackReceived: 1},
	})
	if service == nil || service.TeamName != "red" {
		t.Fatalf("expected compromised red service to be top risk, got %+v", service)
	}
}

func TestBuildAWDReviewSuggestionsIncludesTrafficOnlyHint(t *testing.T) {
	t.Parallel()

	suggestions := buildAWDReviewSuggestions(
		[]assessmentqry.TeacherAWDReviewRoundResp{
			{RoundNumber: 4, AttackCount: 0, TrafficCount: 3, ServiceCount: 1},
		},
		&assessmentqry.TeacherAWDSelectedRoundResp{
			Round: assessmentqry.TeacherAWDReviewRoundResp{RoundNumber: 4},
			Services: []assessmentqry.TeacherAWDReviewServiceResp{
				{TeamName: "blue", AWDChallengeTitle: "web", ServiceStatus: contestcontracts.AWDServiceStatusUp, AttackReceived: 2},
			},
			Traffic: []assessmentqry.TeacherAWDReviewTrafficResp{
				{Path: "/health"},
			},
		},
	)

	if len(suggestions) == 0 {
		t.Fatalf("expected suggestions, got empty")
	}
	joined := strings.Join(suggestions, "\n")
	if !strings.Contains(joined, "访问流量") {
		t.Fatalf("expected traffic-only hint, got %s", joined)
	}
	if !strings.Contains(joined, "第 4 轮") {
		t.Fatalf("expected key round hint, got %s", joined)
	}
}

func TestReportServiceCreateAWDReviewReportExportRejectsRunningContest(t *testing.T) {
	t.Parallel()

	db := newTestSQLiteDB(t)
	if err := db.AutoMigrate(&identitycontracts.User{}, &assessmententity.Report{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}

	teacher := &identitycontracts.User{
		ID:        12,
		Username:  "teacher-running",
		Role:      identitycontracts.RoleTeacher,
		ClassName: "class-a",
		Status:    identitycontracts.UserStatusActive,
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
	appErr, ok := err.(*apperror.AppError)
	if !ok || appErr.Code != apperror.ErrInvalidParams.Code {
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

	teacher := &assessmentdomain.ReportUser{ID: 1, Role: identitycontracts.RoleTeacher, ClassName: "class-a"}
	admin := &assessmentdomain.ReportUser{ID: 2, Role: identitycontracts.RoleAdmin, ClassName: ""}

	if err := validateClassReportAccess(teacher, "class-a"); err != nil {
		t.Fatalf("expected same-class teacher access, got %v", err)
	}
	if err := validateClassReportAccess(admin, "class-b"); err != nil {
		t.Fatalf("expected admin access, got %v", err)
	}

	err := validateClassReportAccess(teacher, "class-b")
	appErr, ok := err.(*apperror.AppError)
	if !ok || appErr.Code != apperror.ErrForbidden.Code {
		t.Fatalf("expected forbidden error, got %#v", err)
	}
}

func TestCreateClassReportRejectsCrossClassTeacherRequest(t *testing.T) {
	t.Parallel()

	db := newTestSQLiteDB(t)
	if err := db.AutoMigrate(&identitycontracts.User{}, &assessmententity.Report{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}

	teacher := &identitycontracts.User{
		ID:        1,
		Username:  "teacher-a",
		Role:      identitycontracts.RoleTeacher,
		ClassName: "class-a",
		Status:    identitycontracts.UserStatusActive,
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
	appErr, ok := err.(*apperror.AppError)
	if !ok || appErr.Code != apperror.ErrForbidden.Code {
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
			{Key: taxonomy.DifficultyEasy, TotalChallenges: 8, CoveredChallenges: 2, SolvedStudents: 2},
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
	if len(data.CategoryDistribution) != len(taxonomy.AllDimensions) {
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
