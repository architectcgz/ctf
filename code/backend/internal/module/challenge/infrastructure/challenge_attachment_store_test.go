package infrastructure

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"

	challengeports "ctf-platform/internal/module/challenge/ports"
	platformsharedfs "ctf-platform/internal/platform/storage/sharedfs"
)

func TestChallengeAttachmentStorePersistsSingleAttachment(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	store := NewChallengeAttachmentStore(platformsharedfs.NewStore(root), "")

	sourceDir := t.TempDir()
	sourcePath := filepath.Join(sourceDir, "readme.txt")
	if err := os.WriteFile(sourcePath, []byte("read me"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	url, err := store.PersistImportedAttachmentBundle(ctx, challengeports.ChallengeImportedAttachmentBundleRequest{
		PackageSlug: "web-demo",
		Attachments: []challengeports.ChallengeImportedAttachment{
			{Name: "../flag.txt", Path: "files/readme.txt", AbsolutePath: sourcePath},
		},
	})
	if err != nil {
		t.Fatalf("PersistImportedAttachmentBundle() error = %v", err)
	}
	if url != "/api/v1/challenges/attachments/imports/web-demo/flag.txt" {
		t.Fatalf("url = %q", url)
	}
	content, err := os.ReadFile(filepath.Join(root, "imports", "web-demo", "flag.txt"))
	if err != nil {
		t.Fatalf("read persisted attachment: %v", err)
	}
	if string(content) != "read me" {
		t.Fatalf("content = %q", content)
	}
}

func TestChallengeAttachmentStoreBundlesMultipleAttachments(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	store := NewChallengeAttachmentStore(platformsharedfs.NewStore(root), "")

	sourceDir := t.TempDir()
	first := filepath.Join(sourceDir, "first.txt")
	second := filepath.Join(sourceDir, "second.txt")
	if err := os.WriteFile(first, []byte("first"), 0o644); err != nil {
		t.Fatalf("write first: %v", err)
	}
	if err := os.WriteFile(second, []byte("second"), 0o644); err != nil {
		t.Fatalf("write second: %v", err)
	}

	url, err := store.PersistImportedAttachmentBundle(ctx, challengeports.ChallengeImportedAttachmentBundleRequest{
		PackageSlug: "web-demo",
		Attachments: []challengeports.ChallengeImportedAttachment{
			{Name: "first.txt", Path: "files/first.txt", AbsolutePath: first},
			{Name: "../second.txt", Path: "files/second.txt", AbsolutePath: second},
		},
	})
	if err != nil {
		t.Fatalf("PersistImportedAttachmentBundle() error = %v", err)
	}
	if url != "/api/v1/challenges/attachments/imports/web-demo/web-demo-attachments.zip" {
		t.Fatalf("url = %q", url)
	}

	archive, err := zip.OpenReader(filepath.Join(root, "imports", "web-demo", "web-demo-attachments.zip"))
	if err != nil {
		t.Fatalf("open bundle: %v", err)
	}
	defer archive.Close()

	names := make([]string, 0, len(archive.File))
	for _, file := range archive.File {
		names = append(names, file.Name)
	}
	if !slices.Contains(names, "first.txt") || !slices.Contains(names, "second.txt") {
		t.Fatalf("bundle names = %v", names)
	}
}

func TestChallengeAttachmentStoreOpenAttachment(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	store := NewChallengeAttachmentStore(platformsharedfs.NewStore(root), "")

	attachmentPath := filepath.Join(root, "imports", "web-demo", "readme.txt")
	if err := os.MkdirAll(filepath.Dir(attachmentPath), 0o755); err != nil {
		t.Fatalf("mkdir attachment dir: %v", err)
	}
	if err := os.WriteFile(attachmentPath, []byte("shared attachment"), 0o644); err != nil {
		t.Fatalf("write attachment: %v", err)
	}

	download, err := store.OpenAttachment(ctx, "imports/web-demo/readme.txt")
	if err != nil {
		t.Fatalf("OpenAttachment() error = %v", err)
	}
	defer download.Reader.Close()

	content, err := io.ReadAll(download.Reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(content) != "shared attachment" {
		t.Fatalf("content = %q", string(content))
	}
	if download.FileName != "readme.txt" {
		t.Fatalf("file name = %q", download.FileName)
	}
}
