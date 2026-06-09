package commands

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	challengeports "ctf-platform/internal/module/challenge/ports"
)

func TestArtifactGCServicePlanFilesProtectsReferencedPaths(t *testing.T) {
	root := t.TempDir()
	attachmentRoot := filepath.Join(root, "challenge-attachments")
	referenced := filepath.Join(attachmentRoot, "imports", "keep", "asset.zip")
	stale := filepath.Join(attachmentRoot, "imports", "stale", "asset.zip")
	mustWriteFileAt(t, referenced, "keep")
	mustWriteFileAt(t, stale, "stale")

	old := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mustChtimes(t, referenced, old)
	mustChtimes(t, stale, old)

	service := NewArtifactGCService(ArtifactGCConfig{
		Now:                 time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC),
		AttachmentRoot:      attachmentRoot,
		AttachmentRetention: 24 * time.Hour,
	}, staticArtifactReferences{
		attachmentPaths: []string{referenced},
	})

	report, err := service.PlanFiles(context.Background())
	if err != nil {
		t.Fatalf("PlanFiles() error = %v", err)
	}

	protected := findArtifactGCCandidate(report.Candidates, referenced)
	if protected == nil {
		t.Fatalf("expected referenced candidate in report: %+v", report.Candidates)
	}
	if !protected.Protected || protected.Reason != ArtifactGCReasonReferenced {
		t.Fatalf("expected referenced candidate protected, got %+v", protected)
	}

	deletable := findArtifactGCCandidate(report.Candidates, stale)
	if deletable == nil {
		t.Fatalf("expected stale candidate in report: %+v", report.Candidates)
	}
	if deletable.Protected || deletable.Reason != ArtifactGCReasonExpired {
		t.Fatalf("expected stale candidate deletable, got %+v", deletable)
	}
}

func TestArtifactGCServiceDryRunDoesNotRemoveFiles(t *testing.T) {
	root := t.TempDir()
	previewRoot := filepath.Join(root, "challenge-import-previews")
	stalePreview := filepath.Join(previewRoot, "preview-1")
	mustWriteFileAt(t, filepath.Join(stalePreview, "preview.json"), "{}")
	mustChtimes(t, stalePreview, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))

	service := NewArtifactGCService(ArtifactGCConfig{
		Now:              time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC),
		PreviewRoots:     []string{previewRoot},
		PreviewRetention: 24 * time.Hour,
	}, staticArtifactReferences{})

	report, err := service.CollectFiles(context.Background(), ArtifactGCExecution{DryRun: true})
	if err != nil {
		t.Fatalf("CollectFiles(dry-run) error = %v", err)
	}
	if report.DeletedCount != 0 {
		t.Fatalf("dry-run deleted count = %d", report.DeletedCount)
	}
	if _, err := os.Stat(stalePreview); err != nil {
		t.Fatalf("dry-run removed preview dir: %v", err)
	}
}

func TestArtifactGCServiceProtectsActiveImageBuildSourceParent(t *testing.T) {
	root := t.TempDir()
	buildRoot := filepath.Join(root, "challenge-image-build-sources")
	candidateDir := filepath.Join(buildRoot, "jeopardy", "web-source-audit", "preview-1")
	sourceDir := filepath.Join(candidateDir, "source")
	mustWriteFileAt(t, filepath.Join(sourceDir, "Dockerfile"), "FROM scratch\n")

	old := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mustChtimes(t, sourceDir, old)
	mustChtimes(t, candidateDir, old)

	service := NewArtifactGCService(ArtifactGCConfig{
		Now:                       time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC),
		ImageBuildSourceRoot:      buildRoot,
		ImageBuildSourceRetention: 24 * time.Hour,
	}, staticArtifactReferences{
		imageBuildSourceDirs: []string{sourceDir},
	})

	report, err := service.CollectFiles(context.Background(), ArtifactGCExecution{DryRun: false})
	if err != nil {
		t.Fatalf("CollectFiles(execute) error = %v", err)
	}
	protected := findArtifactGCCandidate(report.Candidates, candidateDir)
	if protected == nil {
		t.Fatalf("expected image build candidate in report: %+v", report.Candidates)
	}
	if !protected.Protected || protected.Reason != ArtifactGCReasonReferenced {
		t.Fatalf("expected active build source parent protected, got %+v", protected)
	}
	if report.DeletedCount != 0 {
		t.Fatalf("expected no deletions, got %d", report.DeletedCount)
	}
	if _, statErr := os.Stat(filepath.Join(sourceDir, "Dockerfile")); statErr != nil {
		t.Fatalf("active build source should remain: %v", statErr)
	}
}

func TestArtifactGCServiceRejectsDeleteOutsideConfiguredRoots(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.txt")
	mustWriteFileAt(t, outside, "outside")

	service := NewArtifactGCService(ArtifactGCConfig{
		Now:              time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC),
		PreviewRoots:     []string{filepath.Join(root, "challenge-import-previews")},
		PreviewRetention: time.Hour,
	}, staticArtifactReferences{})

	err := service.DeleteFileCandidate(context.Background(), ArtifactGCCandidate{
		Kind: ArtifactGCKindPreview,
		Path: outside,
	})
	if err == nil {
		t.Fatal("expected outside-root delete to fail")
	}
	if _, statErr := os.Stat(outside); statErr != nil {
		t.Fatalf("outside path should remain: %v", statErr)
	}
}

func TestArtifactGCServiceRejectsDeleteConfiguredRoot(t *testing.T) {
	root := t.TempDir()
	previewRoot := filepath.Join(root, "challenge-import-previews")
	mustWriteFileAt(t, filepath.Join(previewRoot, "preview-1", "preview.json"), "{}")

	service := NewArtifactGCService(ArtifactGCConfig{
		Now:              time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC),
		PreviewRoots:     []string{previewRoot},
		PreviewRetention: time.Hour,
	}, staticArtifactReferences{})

	err := service.DeleteFileCandidate(context.Background(), ArtifactGCCandidate{
		Kind: ArtifactGCKindPreview,
		Path: previewRoot,
	})
	if err == nil {
		t.Fatal("expected configured root delete to fail")
	}
	if _, statErr := os.Stat(previewRoot); statErr != nil {
		t.Fatalf("configured root should remain: %v", statErr)
	}
}

type staticArtifactReferences struct {
	attachmentPaths      []string
	imageBuildSourceDirs []string
}

func (s staticArtifactReferences) ListArtifactReferences(ctx context.Context) (challengeports.ArtifactReferences, error) {
	return challengeports.ArtifactReferences{
		AttachmentPaths:      s.attachmentPaths,
		ImageBuildSourceDirs: s.imageBuildSourceDirs,
	}, nil
}

func findArtifactGCCandidate(candidates []ArtifactGCCandidate, path string) *ArtifactGCCandidate {
	clean := filepath.Clean(path)
	for index := range candidates {
		if filepath.Clean(candidates[index].Path) == clean {
			return &candidates[index]
		}
	}
	return nil
}

func mustWriteFileAt(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustChtimes(t *testing.T, path string, at time.Time) {
	t.Helper()
	if err := os.Chtimes(path, at, at); err != nil {
		t.Fatalf("chtimes %s: %v", path, err)
	}
}
