package commands

import (
	"archive/zip"
	"bytes"
	"compress/zlib"
	"context"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	teachingadvice "ctf-platform/internal/teaching/advice"
	"ctf-platform/internal/teaching/evidence"
	"errors"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf16"
)

type testReportRepository struct {
	db                *gorm.DB
	users             map[int64]*assessmentdomain.ReportUser
	contests          map[int64]*contestcontracts.Contest
	personalStats     *assessmentdomain.PersonalReportStats
	totalChallenges   int64
	classSummary      *assessmentdomain.ClassInsightSummary
	classTrend        *assessmentdomain.ClassInsightTrend
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

type testReportOutputStore struct {
	root string
}

func newTestReportOutputStore(t *testing.T) assessmentports.ReportOutputStore {
	t.Helper()
	return &testReportOutputStore{root: t.TempDir()}
}

func (s *testReportOutputStore) PrepareReportOutput(ctx context.Context, fileName string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err := os.MkdirAll(s.root, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(s.root, filepath.Base(fileName)), nil
}

func (s *testReportOutputStore) ResolveReportDownloadPath(ctx context.Context, path string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", assessmentports.ErrAssessmentReportOutputNotFound
		}
		return "", err
	}
	return path, nil
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

func (r *testReportRepository) GetClassSummary(_ context.Context, _ string, since time.Time) (*assessmentdomain.ClassInsightSummary, error) {
	if r != nil {
		r.lastSummarySince = since
	}
	if r != nil && r.classSummary != nil {
		return r.classSummary, nil
	}
	return &assessmentdomain.ClassInsightSummary{}, nil
}

func (r *testReportRepository) GetClassTrend(_ context.Context, _ string, since time.Time, days int) (*assessmentdomain.ClassInsightTrend, error) {
	if r != nil {
		r.lastTrendSince = since
		r.lastTrendDays = days
	}
	if r != nil && r.classTrend != nil {
		return r.classTrend, nil
	}
	return &assessmentdomain.ClassInsightTrend{}, nil
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
