package assessment

import (
	"os"
	"path/filepath"
	"testing"

	"ctf-platform/internal/dto"
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
