package infrastructure

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

func TestChallengeImportPreviewStoreCreatesWorkspaceAndRecords(t *testing.T) {
	ctx := context.Background()
	store := NewChallengeImportPreviewStore(t.TempDir())

	archive := buildPreviewStoreZip(t, map[string]string{
		"challenge.yml": "apiVersion: v1\nkind: challenge\n",
		"statement.md":  "hello",
	}, nil)
	workspace, err := store.CreateWorkspace(ctx, "preview-1", "package.zip", bytes.NewReader(archive))
	if err != nil {
		t.Fatalf("CreateWorkspace() error = %v", err)
	}
	if workspace.ID != "preview-1" {
		t.Fatalf("workspace ID = %q", workspace.ID)
	}
	if _, err := os.Stat(filepath.Join(workspace.SourceDir, "challenge.yml")); err != nil {
		t.Fatalf("extracted challenge.yml: %v", err)
	}
	if _, err := os.Stat(workspace.ArchivePath); err != nil {
		t.Fatalf("saved archive: %v", err)
	}

	createdAt := time.Date(2026, 6, 11, 1, 2, 3, 0, time.UTC)
	record := &challengeports.ChallengeImportPreviewRecord{
		ID:          workspace.ID,
		FileName:    workspace.FileName,
		ArchivePath: workspace.ArchivePath,
		SourceDir:   workspace.SourceDir,
		CreatedBy:   1001,
		CreatedAt:   createdAt,
		Preview: challengecontracts.ChallengeImportPreviewResp{
			ID:        workspace.ID,
			FileName:  workspace.FileName,
			Slug:      "web-demo",
			CreatedAt: createdAt,
		},
	}
	if err := store.SaveRecord(ctx, record); err != nil {
		t.Fatalf("SaveRecord() error = %v", err)
	}

	loaded, err := store.LoadRecord(ctx, "preview-1")
	if err != nil {
		t.Fatalf("LoadRecord() error = %v", err)
	}
	if loaded.ID != record.ID || loaded.ArchivePath != record.ArchivePath || loaded.Preview.Slug != "web-demo" {
		t.Fatalf("loaded record = %+v", loaded)
	}

	records, err := store.ListRecords(ctx)
	if err != nil {
		t.Fatalf("ListRecords() error = %v", err)
	}
	if len(records) != 1 || records[0].ID != "preview-1" {
		t.Fatalf("records = %+v", records)
	}

	if err := store.DeleteWorkspace(ctx, "preview-1"); err != nil {
		t.Fatalf("DeleteWorkspace() error = %v", err)
	}
	if _, err := os.Stat(workspace.SourceDir); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("workspace source still exists or unexpected error: %v", err)
	}
}

func TestAWDChallengeImportPreviewStoreCreatesWorkspaceAndRecords(t *testing.T) {
	ctx := context.Background()
	store := NewAWDChallengeImportPreviewStore(t.TempDir())

	archive := buildPreviewStoreZip(t, map[string]string{
		"challenge.yml": "apiVersion: v1\nkind: awd-challenge\n",
	}, nil)
	workspace, err := store.CreateWorkspace(ctx, "awd-preview-1", "awd-package.zip", bytes.NewReader(archive))
	if err != nil {
		t.Fatalf("CreateWorkspace() error = %v", err)
	}
	if workspace.ID != "awd-preview-1" {
		t.Fatalf("workspace ID = %q", workspace.ID)
	}
	if _, err := os.Stat(filepath.Join(workspace.SourceDir, "challenge.yml")); err != nil {
		t.Fatalf("extracted challenge.yml: %v", err)
	}

	createdAt := time.Date(2026, 6, 11, 1, 2, 3, 0, time.UTC)
	record := &challengeports.AWDChallengeImportPreviewRecord{
		ID:          workspace.ID,
		FileName:    workspace.FileName,
		ArchivePath: workspace.ArchivePath,
		SourceDir:   workspace.SourceDir,
		CreatedBy:   2001,
		CreatedAt:   createdAt,
		Preview: challengecontracts.AWDChallengeImportPreviewResp{
			ID:        workspace.ID,
			FileName:  workspace.FileName,
			Slug:      "awd-demo",
			CreatedAt: createdAt,
		},
	}
	if err := store.SaveRecord(ctx, record); err != nil {
		t.Fatalf("SaveRecord() error = %v", err)
	}

	loaded, err := store.LoadRecord(ctx, "awd-preview-1")
	if err != nil {
		t.Fatalf("LoadRecord() error = %v", err)
	}
	if loaded.ID != record.ID || loaded.ArchivePath != record.ArchivePath || loaded.Preview.Slug != "awd-demo" {
		t.Fatalf("loaded record = %+v", loaded)
	}

	records, err := store.ListRecords(ctx)
	if err != nil {
		t.Fatalf("ListRecords() error = %v", err)
	}
	if len(records) != 1 || records[0].ID != "awd-preview-1" {
		t.Fatalf("records = %+v", records)
	}

	if err := store.DeleteWorkspace(ctx, "awd-preview-1"); err != nil {
		t.Fatalf("DeleteWorkspace() error = %v", err)
	}
	if _, err := os.Stat(workspace.SourceDir); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("workspace source still exists or unexpected error: %v", err)
	}
}

func TestChallengeImportPreviewStoreRejectsUnsafeArchives(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	store := NewChallengeImportPreviewStore(root)
	tooManyFiles := make(map[string]string, 129)
	for i := range 129 {
		tooManyFiles[filepath.Join("files", strconv.Itoa(i)+".txt")] = "x"
	}

	tests := []struct {
		name    string
		archive []byte
	}{
		{
			name: "zip slip path",
			archive: buildPreviewStoreZip(t, map[string]string{
				"../escape.txt": "escape",
			}, nil),
		},
		{
			name: "symlink entry",
			archive: buildPreviewStoreZip(t, nil, map[string]string{
				"link": "challenge.yml",
			}),
		},
		{
			name:    "too many files",
			archive: buildPreviewStoreZip(t, tooManyFiles, nil),
		},
		{
			name: "single file too large",
			archive: buildPreviewStoreZip(t, map[string]string{
				"challenge.yml": string(bytes.Repeat([]byte("a"), 16<<20+1)),
			}, nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			previewID := "preview-" + tt.name
			if _, err := store.CreateWorkspace(ctx, previewID, "package.zip", bytes.NewReader(tt.archive)); err == nil {
				t.Fatal("expected unsafe archive to be rejected")
			}
			if _, err := os.Stat(filepath.Join(root, previewID)); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("expected rejected preview workspace to be removed, stat err = %v", err)
			}
		})
	}
}

func buildPreviewStoreZip(t *testing.T, files map[string]string, symlinks map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for name, content := range files {
		fileWriter, err := writer.Create(name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := fileWriter.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry: %v", err)
		}
	}
	for name, target := range symlinks {
		header := &zip.FileHeader{Name: name}
		header.SetMode(os.ModeSymlink | 0o777)
		fileWriter, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatalf("create symlink zip entry: %v", err)
		}
		if _, err := fileWriter.Write([]byte(target)); err != nil {
			t.Fatalf("write symlink zip entry: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}
