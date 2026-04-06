package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ctf-platform/internal/dto"
)

func TestServiceListChallengeImportsSortsAndFiltersByActor(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)

	mustWriteChallengeImportPreviewRecord(t, tempDir, storedChallengeImportPreview{
		ID:        "older-owned",
		FileName:  "older.zip",
		CreatedBy: 1001,
		CreatedAt: time.Date(2026, 4, 6, 8, 0, 0, 0, time.UTC),
		Preview: dto.ChallengeImportPreviewResp{
			ID:         "older-owned",
			FileName:   "older.zip",
			Title:      "Older Owned",
			CreatedAt:  time.Date(2026, 4, 6, 8, 0, 0, 0, time.UTC),
			Category:   "web",
			Difficulty: "easy",
			Points:     100,
			Flag:       dto.ChallengeImportFlagResp{Type: "static"},
		},
	})
	mustWriteChallengeImportPreviewRecord(t, tempDir, storedChallengeImportPreview{
		ID:        "newer-owned",
		FileName:  "newer.zip",
		CreatedBy: 1001,
		CreatedAt: time.Date(2026, 4, 6, 9, 0, 0, 0, time.UTC),
		Preview: dto.ChallengeImportPreviewResp{
			ID:         "newer-owned",
			FileName:   "newer.zip",
			Title:      "Newer Owned",
			CreatedAt:  time.Date(2026, 4, 6, 9, 0, 0, 0, time.UTC),
			Category:   "misc",
			Difficulty: "medium",
			Points:     150,
			Flag:       dto.ChallengeImportFlagResp{Type: "dynamic"},
		},
	})
	mustWriteChallengeImportPreviewRecord(t, tempDir, storedChallengeImportPreview{
		ID:        "other-user",
		FileName:  "other.zip",
		CreatedBy: 2002,
		CreatedAt: time.Date(2026, 4, 6, 10, 0, 0, 0, time.UTC),
		Preview: dto.ChallengeImportPreviewResp{
			ID:         "other-user",
			FileName:   "other.zip",
			Title:      "Other User",
			CreatedAt:  time.Date(2026, 4, 6, 10, 0, 0, 0, time.UTC),
			Category:   "crypto",
			Difficulty: "hard",
			Points:     300,
			Flag:       dto.ChallengeImportFlagResp{Type: "static"},
		},
	})

	service := &ChallengeService{}

	previews, err := service.ListChallengeImports(1001)
	if err != nil {
		t.Fatalf("ListChallengeImports() error = %v", err)
	}

	if len(previews) != 2 {
		t.Fatalf("expected 2 previews, got %d", len(previews))
	}
	if previews[0].ID != "newer-owned" {
		t.Fatalf("expected newest preview first, got %s", previews[0].ID)
	}
	if previews[1].ID != "older-owned" {
		t.Fatalf("expected older preview second, got %s", previews[1].ID)
	}
}

func TestServiceListChallengeImportsReturnsEmptyWhenPreviewRootMissing(t *testing.T) {
	tempDir := filepath.Join(t.TempDir(), "missing")
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)

	service := &ChallengeService{}

	previews, err := service.ListChallengeImports(1001)
	if err != nil {
		t.Fatalf("ListChallengeImports() error = %v", err)
	}
	if len(previews) != 0 {
		t.Fatalf("expected no previews, got %d", len(previews))
	}
}

func mustWriteChallengeImportPreviewRecord(t *testing.T, root string, record storedChallengeImportPreview) {
	t.Helper()

	previewDir := filepath.Join(root, record.ID)
	if err := os.MkdirAll(previewDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	content, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(previewDir, "preview.json"), content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
