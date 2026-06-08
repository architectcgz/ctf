package commands

import (
	"bytes"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

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
