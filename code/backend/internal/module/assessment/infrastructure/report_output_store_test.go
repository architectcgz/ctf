package infrastructure

import (
	"context"
	"io"
	"os"
	"testing"

	platformsharedfs "ctf-platform/internal/platform/storage/sharedfs"
)

func TestReportOutputStoreSupportsCrossReplicaRead(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	storeA := NewReportOutputStore(platformsharedfs.NewStore(root), "reports")
	storeB := NewReportOutputStore(platformsharedfs.NewStore(root), "reports")

	output, err := storeA.PrepareReportOutput(context.Background(), "class-report.pdf")
	if err != nil {
		t.Fatalf("PrepareReportOutput() error = %v", err)
	}
	if err := os.WriteFile(output.LocalPath, []byte("replica-report"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	download, err := storeB.OpenReportDownload(context.Background(), output.StorageKey)
	if err != nil {
		t.Fatalf("OpenReportDownload() error = %v", err)
	}
	defer download.Reader.Close()

	content, err := io.ReadAll(download.Reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(content) != "replica-report" {
		t.Fatalf("content = %q", string(content))
	}
	if download.Size != int64(len("replica-report")) {
		t.Fatalf("size = %d", download.Size)
	}
	if download.StorageKey != output.StorageKey {
		t.Fatalf("storage key = %q, want %q", download.StorageKey, output.StorageKey)
	}
}
