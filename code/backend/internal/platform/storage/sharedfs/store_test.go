package sharedfs

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	platformstorage "ctf-platform/internal/platform/storage"
)

func TestSharedFSStorePrepareLocalWriteIsReadableAcrossInstances(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writer := NewStore(root)
	reader := NewStore(root)

	localPath, err := writer.PrepareLocalWrite(context.Background(), "reports/class/report.pdf")
	if err != nil {
		t.Fatalf("PrepareLocalWrite() error = %v", err)
	}
	if err := os.WriteFile(localPath, []byte("report-bytes"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	opened, err := reader.Open(context.Background(), "reports/class/report.pdf")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer opened.Close()

	content, err := io.ReadAll(opened)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(content) != "report-bytes" {
		t.Fatalf("content = %q", string(content))
	}

	info, err := reader.Stat(context.Background(), "reports/class/report.pdf")
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.Size != int64(len("report-bytes")) {
		t.Fatalf("size = %d", info.Size)
	}
	if info.Key != "reports/class/report.pdf" {
		t.Fatalf("key = %q", info.Key)
	}

	expectedPath := filepath.Join(root, "reports", "class", "report.pdf")
	if localPath != expectedPath {
		t.Fatalf("local path = %q, want %q", localPath, expectedPath)
	}
}

func TestSharedFSStorePutOpenAndStat(t *testing.T) {
	t.Parallel()

	store := NewStore(t.TempDir())
	if err := store.Put(context.Background(), "attachments/imports/demo/readme.txt", stringsReader("hello attachment")); err != nil {
		t.Fatalf("Put() error = %v", err)
	}

	opened, err := store.Open(context.Background(), "attachments/imports/demo/readme.txt")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer opened.Close()

	content, err := io.ReadAll(opened)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(content) != "hello attachment" {
		t.Fatalf("content = %q", string(content))
	}

	info, err := store.Stat(context.Background(), "attachments/imports/demo/readme.txt")
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.Size != int64(len("hello attachment")) {
		t.Fatalf("size = %d", info.Size)
	}
}

func TestSharedFSStoreRejectsUnsafeKeys(t *testing.T) {
	t.Parallel()

	store := NewStore(t.TempDir())
	unsafeKeys := []string{"../escape.txt", "/absolute/path", "reports/../../escape.txt", "reports\\..\\escape.txt"}

	for _, key := range unsafeKeys {
		key := key
		t.Run(key, func(t *testing.T) {
			t.Parallel()

			if _, err := store.PrepareLocalWrite(context.Background(), key); !errors.Is(err, platformstorage.ErrUnsafeKey) {
				t.Fatalf("PrepareLocalWrite() error = %v, want ErrUnsafeKey", err)
			}
			if err := store.Put(context.Background(), key, stringsReader("bad")); !errors.Is(err, platformstorage.ErrUnsafeKey) {
				t.Fatalf("Put() error = %v, want ErrUnsafeKey", err)
			}
			if _, err := store.Open(context.Background(), key); !errors.Is(err, platformstorage.ErrUnsafeKey) {
				t.Fatalf("Open() error = %v, want ErrUnsafeKey", err)
			}
		})
	}
}

func TestSharedFSStoreMissingObjectReturnsNotFound(t *testing.T) {
	t.Parallel()

	store := NewStore(t.TempDir())
	if _, err := store.Open(context.Background(), "reports/missing.pdf"); !errors.Is(err, platformstorage.ErrNotFound) {
		t.Fatalf("Open() error = %v, want ErrNotFound", err)
	}
	if _, err := store.Stat(context.Background(), "reports/missing.pdf"); !errors.Is(err, platformstorage.ErrNotFound) {
		t.Fatalf("Stat() error = %v, want ErrNotFound", err)
	}
}

func stringsReader(value string) io.Reader {
	return &staticStringReader{remaining: []byte(value)}
}

type staticStringReader struct {
	remaining []byte
}

func (r *staticStringReader) Read(p []byte) (int, error) {
	if len(r.remaining) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.remaining)
	r.remaining = r.remaining[n:]
	return n, nil
}
