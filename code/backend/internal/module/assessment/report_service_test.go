package assessment

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

func TestWritePersonalPDFCreatesPDFFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "personal-report.pdf")
	data := &personalReportData{
		User: &ReportUser{
			ID:        1,
			Username:  "alice",
			ClassName: "class-a",
		},
		SkillProfile: []*dto.SkillDimension{
			{Dimension: "web", Score: 0.8},
			{Dimension: "crypto", Score: 0.5},
		},
		Stats: &PersonalReportStats{
			TotalScore:    400,
			TotalSolved:   4,
			TotalAttempts: 7,
			Rank:          2,
		},
		DimensionStats: []ReportDimensionStat{
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
		User: &ReportUser{
			ID:        1,
			Username:  "alice",
			ClassName: "class-a",
		},
		SkillProfile: []*dto.SkillDimension{
			{Dimension: "web", Score: 0.8},
			{Dimension: "crypto", Score: 0.5},
		},
		Stats: &PersonalReportStats{
			TotalScore:    400,
			TotalSolved:   4,
			TotalAttempts: 7,
			Rank:          2,
		},
		DimensionStats: []ReportDimensionStat{
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

	if got := reportFileExtension("excel"); got != "xlsx" {
		t.Fatalf("expected xlsx extension for excel, got %s", got)
	}
	if got := reportFileExtension("pdf"); got != "pdf" {
		t.Fatalf("expected pdf extension for pdf, got %s", got)
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

func TestValidateClassReportAccess(t *testing.T) {
	t.Parallel()

	teacher := &ReportUser{ID: 1, Role: model.RoleTeacher, ClassName: "class-a"}
	admin := &ReportUser{ID: 2, Role: model.RoleAdmin, ClassName: ""}

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

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
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
		NewReportRepository(db),
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
