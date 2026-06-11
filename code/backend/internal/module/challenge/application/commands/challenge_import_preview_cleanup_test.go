package commands

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"os"
	"testing"

	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
)

func TestPreviewChallengeImportDeletesWorkspaceWhenParseFails(t *testing.T) {
	previewRoot := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", previewRoot)

	service := NewChallengeImportService(
		challengeinfra.NewChallengeImportPreviewStore(""),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	_, err := service.PreviewChallengeImport(context.Background(), 1001, "invalid.zip", bytes.NewReader(invalidImportArchive(t)))
	if err == nil {
		t.Fatal("expected preview parse error")
	}

	assertDirectoryEmptyOrMissing(t, previewRoot)
}

func TestAWDChallengeImportPreviewDeletesWorkspaceWhenParseFails(t *testing.T) {
	previewRoot := t.TempDir()
	t.Setenv("AWD_CHALLENGE_IMPORT_PREVIEW_DIR", previewRoot)

	service := NewAWDChallengeImportService(
		nil,
		challengeinfra.NewAWDChallengeImportPreviewStore(""),
		nil,
		nil,
	)

	_, err := service.PreviewImport(context.Background(), 1001, "invalid.zip", bytes.NewReader(invalidImportArchive(t)))
	if err == nil {
		t.Fatal("expected awd preview parse error")
	}

	assertDirectoryEmptyOrMissing(t, previewRoot)
}

func invalidImportArchive(t *testing.T) []byte {
	t.Helper()

	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	fileWriter, err := writer.Create("readme.txt")
	if err != nil {
		t.Fatalf("Create(readme.txt) error = %v", err)
	}
	if _, err := fileWriter.Write([]byte("not a challenge package")); err != nil {
		t.Fatalf("Write(readme.txt) error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	return buffer.Bytes()
}

func assertDirectoryEmptyOrMissing(t *testing.T, path string) {
	t.Helper()

	entries, err := os.ReadDir(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return
		}
		t.Fatalf("ReadDir(%s) error = %v", path, err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected %s to be empty after failed preview, got %d entries", path, len(entries))
	}
}
