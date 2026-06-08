package commands

import (
	"context"
	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	"testing"
	"time"
)

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
