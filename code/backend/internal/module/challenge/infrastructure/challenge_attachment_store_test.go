package infrastructure

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	challengeports "ctf-platform/internal/module/challenge/ports"
)

func TestChallengeAttachmentStorePersistsSingleAttachment(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	store := NewChallengeAttachmentStore(root)

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
	store := NewChallengeAttachmentStore(root)

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
