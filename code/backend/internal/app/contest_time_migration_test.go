package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContestTimeContractMigrationExists(t *testing.T) {
	upFiles, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_use_timestamptz_for_contest_times.up.sql"))
	if err != nil {
		t.Fatalf("glob up migration: %v", err)
	}
	downFiles, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_use_timestamptz_for_contest_times.down.sql"))
	if err != nil {
		t.Fatalf("glob down migration: %v", err)
	}
	if len(upFiles) != 1 || len(downFiles) != 1 {
		t.Fatalf("expected one contest time contract migration, got up=%v down=%v", upFiles, downFiles)
	}

	up, err := os.ReadFile(upFiles[0])
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	upSQL := string(up)
	for _, snippet := range []string{
		"ALTER TABLE public.contests",
		"start_time TYPE timestamp with time zone USING start_time AT TIME ZONE 'UTC'",
		"end_time TYPE timestamp with time zone USING end_time AT TIME ZONE 'UTC'",
		"freeze_time TYPE timestamp with time zone USING freeze_time AT TIME ZONE 'UTC'",
	} {
		if !strings.Contains(upSQL, snippet) {
			t.Fatalf("up migration should contain %q, got:\n%s", snippet, upSQL)
		}
	}

	down, err := os.ReadFile(downFiles[0])
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	downSQL := string(down)
	for _, snippet := range []string{
		"ALTER TABLE public.contests",
		"start_time TYPE timestamp without time zone USING start_time AT TIME ZONE 'UTC'",
		"end_time TYPE timestamp without time zone USING end_time AT TIME ZONE 'UTC'",
		"freeze_time TYPE timestamp without time zone USING freeze_time AT TIME ZONE 'UTC'",
	} {
		if !strings.Contains(downSQL, snippet) {
			t.Fatalf("down migration should contain %q, got:\n%s", snippet, downSQL)
		}
	}
}
