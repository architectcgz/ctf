package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContestStatusTransitionMigrationContract(t *testing.T) {
	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000003_create_contest_status_transitions.up.sql"))
	if err != nil {
		t.Fatalf("read transition up migration: %v", err)
	}
	upSQL := string(up)
	for _, snippet := range []string{
		"ALTER TABLE public.contests",
		"ADD COLUMN status_version bigint DEFAULT 0 NOT NULL",
		"CREATE TABLE public.contest_status_transitions",
		"CREATE UNIQUE INDEX uk_contest_status_transitions_contest_version",
		"(contest_id, status_version)",
		"CREATE INDEX idx_contest_status_transitions_occurred_at",
	} {
		if !strings.Contains(upSQL, snippet) {
			t.Fatalf("transition up migration should contain %q, got:\n%s", snippet, upSQL)
		}
	}

	down, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000003_create_contest_status_transitions.down.sql"))
	if err != nil {
		t.Fatalf("read transition down migration: %v", err)
	}
	downSQL := string(down)
	for _, snippet := range []string{
		"DROP INDEX IF EXISTS public.idx_contest_status_transitions_occurred_at",
		"DROP INDEX IF EXISTS public.uk_contest_status_transitions_contest_version",
		"DROP TABLE IF EXISTS public.contest_status_transitions",
		"ALTER TABLE public.contests DROP COLUMN IF EXISTS status_version",
	} {
		if !strings.Contains(downSQL, snippet) {
			t.Fatalf("transition down migration should contain %q, got:\n%s", snippet, downSQL)
		}
	}
}
