package infrastructure

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"

	challengeports "ctf-platform/internal/module/challenge/ports"
)

func TestChallengePackageStoragePersistsImportedPackageSource(t *testing.T) {
	ctx := context.Background()
	sourceRoot := t.TempDir()
	exportRoot := t.TempDir()
	store := NewChallengePackageStorage(ChallengePackageStorageConfig{
		SourceRoot: sourceRoot,
		ExportRoot: exportRoot,
	})

	sourceDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "challenge.yml"), []byte("kind: challenge\n"), 0o644); err != nil {
		t.Fatalf("write source manifest: %v", err)
	}
	archivePath := filepath.Join(t.TempDir(), "upload.zip")
	if err := os.WriteFile(archivePath, []byte("archive"), 0o644); err != nil {
		t.Fatalf("write archive: %v", err)
	}

	stored, err := store.PersistImportedPackageSource(ctx, challengeports.ChallengeImportedPackageSourceRequest{
		ChallengeID:        7,
		RevisionNo:         2,
		PackageSlug:        "web-demo",
		SourceDir:          sourceDir,
		PreviewArchivePath: archivePath,
		PreviewArchiveName: "../package.zip",
	})
	if err != nil {
		t.Fatalf("PersistImportedPackageSource() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(stored.SourceDir, "challenge.yml")); err != nil {
		t.Fatalf("copied source manifest: %v", err)
	}
	if filepath.Base(stored.ArchivePath) != "package.zip" {
		t.Fatalf("archive path = %q", stored.ArchivePath)
	}
	if content, err := os.ReadFile(stored.ArchivePath); err != nil || string(content) != "archive" {
		t.Fatalf("copied archive content = %q, err = %v", content, err)
	}
}

func TestChallengePackageStoragePersistsImportedImageBuildSource(t *testing.T) {
	ctx := context.Background()
	imageBuildRoot := t.TempDir()
	store := NewChallengePackageStorage(ChallengePackageStorageConfig{
		ImageBuildSourceRoot: imageBuildRoot,
	})

	sourceDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(sourceDir, "docker"), 0o755); err != nil {
		t.Fatalf("create docker dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "docker", "Dockerfile"), []byte("FROM nginx"), 0o644); err != nil {
		t.Fatalf("write Dockerfile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "docker", "app.py"), []byte("print('ok')"), 0o644); err != nil {
		t.Fatalf("write app: %v", err)
	}

	stored, err := store.PersistImportedImageBuildSource(ctx, challengeports.ChallengeImportedImageBuildSourceRequest{
		ChallengeMode:  "jeopardy",
		PackageSlug:    "web-demo",
		PreviewID:      "preview-1",
		RootDir:        sourceDir,
		DockerfilePath: filepath.Join(sourceDir, "docker", "Dockerfile"),
		ContextPath:    filepath.Join(sourceDir, "docker"),
	})
	if err != nil {
		t.Fatalf("PersistImportedImageBuildSource() error = %v", err)
	}
	if stored == nil {
		t.Fatal("expected stored build source")
	}
	if stored.RootDir != filepath.Join(imageBuildRoot, "jeopardy", "web-demo", "preview-1") {
		t.Fatalf("RootDir = %q", stored.RootDir)
	}
	if _, err := os.Stat(stored.DockerfilePath); err != nil {
		t.Fatalf("expected copied Dockerfile: %v", err)
	}
	if stored.ContextPath != filepath.Join(stored.SourceDir, "docker") {
		t.Fatalf("ContextPath = %q", stored.ContextPath)
	}
}

func TestChallengePackageStorageBuildsExportArchive(t *testing.T) {
	ctx := context.Background()
	sourceRoot := t.TempDir()
	exportRoot := t.TempDir()
	store := NewChallengePackageStorage(ChallengePackageStorageConfig{
		SourceRoot: sourceRoot,
		ExportRoot: exportRoot,
	})

	baseSource := t.TempDir()
	if err := os.WriteFile(filepath.Join(baseSource, "challenge.yml"), []byte("old"), 0o644); err != nil {
		t.Fatalf("write base manifest: %v", err)
	}

	workspace, err := store.PrepareExportWorkspace(ctx, challengeports.ChallengePackageExportWorkspaceRequest{
		ChallengeID: 7,
		RevisionNo:  3,
		PackageSlug: "web-demo",
		SourceDir:   baseSource,
		FileName:    "../web-demo.zip",
	})
	if err != nil {
		t.Fatalf("PrepareExportWorkspace() error = %v", err)
	}
	if err := store.WriteTextFile(ctx, workspace.SourceDir, "challenge.yml", "new"); err != nil {
		t.Fatalf("WriteTextFile() error = %v", err)
	}
	if content, err := store.ReadTextFile(ctx, workspace.SourceDir, "challenge.yml"); err != nil || content != "new" {
		t.Fatalf("ReadTextFile() = %q, err = %v", content, err)
	}
	if err := store.BuildExportArchive(ctx, *workspace); err != nil {
		t.Fatalf("BuildExportArchive() error = %v", err)
	}
	fileName, err := store.EnsureArchiveExists(ctx, workspace.ArchivePath)
	if err != nil {
		t.Fatalf("EnsureArchiveExists() error = %v", err)
	}
	if fileName != "web-demo.zip" {
		t.Fatalf("fileName = %q", fileName)
	}

	archive, err := zip.OpenReader(workspace.ArchivePath)
	if err != nil {
		t.Fatalf("open export archive: %v", err)
	}
	defer archive.Close()
	if len(archive.File) != 1 || archive.File[0].Name != "challenge.yml" {
		t.Fatalf("archive files = %+v", archive.File)
	}

	if err := store.DeletePath(ctx, workspace.ExportRoot); err != nil {
		t.Fatalf("DeletePath() error = %v", err)
	}
	if _, err := os.Stat(workspace.ExportRoot); !os.IsNotExist(err) {
		t.Fatalf("export root still exists or unexpected error: %v", err)
	}
}
