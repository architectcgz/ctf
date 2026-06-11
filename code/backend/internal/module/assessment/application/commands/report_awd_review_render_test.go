package commands

import (
	"archive/zip"
	"bytes"
	"context"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

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
