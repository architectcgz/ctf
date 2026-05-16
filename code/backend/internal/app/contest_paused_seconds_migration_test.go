package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContestPausedSecondsMigrationContract(t *testing.T) {
	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000007_add_contest_paused_seconds.up.sql"))
	if err != nil {
		t.Fatalf("read paused seconds up migration: %v", err)
	}
	upSQL := string(up)
	for _, snippet := range []string{
		"ALTER TABLE public.contests",
		"ADD COLUMN paused_seconds bigint DEFAULT 0 NOT NULL",
		"ADD COLUMN runtime_recovery_key varchar(191) DEFAULT '' NOT NULL",
		"ADD COLUMN runtime_recovery_applied_seconds bigint DEFAULT 0 NOT NULL",
	} {
		if !strings.Contains(upSQL, snippet) {
			t.Fatalf("paused seconds up migration should contain %q, got:\n%s", snippet, upSQL)
		}
	}

	down, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000007_add_contest_paused_seconds.down.sql"))
	if err != nil {
		t.Fatalf("read paused seconds down migration: %v", err)
	}
	for _, snippet := range []string{
		"ALTER TABLE public.contests",
		"DROP COLUMN IF EXISTS runtime_recovery_applied_seconds",
		"DROP COLUMN IF EXISTS runtime_recovery_key",
		"DROP COLUMN IF EXISTS paused_seconds",
	} {
		if !strings.Contains(string(down), snippet) {
			t.Fatalf("paused seconds down migration should contain %q, got:\n%s", snippet, string(down))
		}
	}
}
