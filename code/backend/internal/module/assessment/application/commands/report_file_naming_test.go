package commands

import (
	"ctf-platform/internal/config"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	"path/filepath"
	"testing"
)

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
