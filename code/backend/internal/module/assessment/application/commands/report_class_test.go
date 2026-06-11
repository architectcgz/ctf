package commands

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	"ctf-platform/internal/shared/taxonomy"
	teachingadvice "ctf-platform/internal/teaching/advice"
	"ctf-platform/internal/teaching/classwindow"
)

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
		newTestReportOutputStore(t),
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
		classSummary: &assessmentdomain.ClassInsightSummary{
			ClassName:          "class-a",
			StudentCount:       2,
			AverageSolved:      2,
			ActiveStudentCount: 1,
			ActiveRate:         50,
			RecentEventCount:   6,
		},
		classTrend: &assessmentdomain.ClassInsightTrend{
			ClassName: "class-a",
			Points: []assessmentdomain.ClassInsightTrendPoint{
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
		newTestReportOutputStore(t),
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
