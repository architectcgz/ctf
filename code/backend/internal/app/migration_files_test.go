package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImagesDeletedAtIndexMigrationExists(t *testing.T) {
	t.Parallel()

	upFiles, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_images_deleted_at_index.up.sql"))
	if err != nil {
		t.Fatalf("glob up migration: %v", err)
	}
	downFiles, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_images_deleted_at_index.down.sql"))
	if err != nil {
		t.Fatalf("glob down migration: %v", err)
	}
	if len(upFiles) != 1 || len(downFiles) != 1 {
		t.Fatalf("expected one up/down images deleted_at index migration, got up=%v down=%v", upFiles, downFiles)
	}

	up, err := os.ReadFile(upFiles[0])
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	down, err := os.ReadFile(downFiles[0])
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	if !strings.Contains(string(up), "idx_images_deleted_at") || !strings.Contains(string(up), "deleted_at") {
		t.Fatalf("up migration should create idx_images_deleted_at on deleted_at, got:\n%s", string(up))
	}
	if !strings.Contains(string(down), "idx_images_deleted_at") {
		t.Fatalf("down migration should drop idx_images_deleted_at, got:\n%s", string(down))
	}
}
