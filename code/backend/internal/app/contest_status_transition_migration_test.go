package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContestStatusTransitionContractInBaseline(t *testing.T) {
	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000001_init_schema.up.sql"))
	if err != nil {
		t.Fatalf("read baseline migration: %v", err)
	}
	upSQL := string(up)
	for _, snippet := range []string{
		"CREATE TABLE public.contests",
		"status_version bigint DEFAULT 0 NOT NULL",
		"CREATE TABLE public.contest_status_transitions",
		"status_version bigint NOT NULL",
		"CREATE UNIQUE INDEX uk_contest_status_transitions_contest_version",
		"(contest_id, status_version)",
		"CREATE INDEX idx_contest_status_transitions_occurred_at",
	} {
		if !strings.Contains(upSQL, snippet) {
			t.Fatalf("baseline migration should contain %q, got:\n%s", snippet, upSQL)
		}
	}

	down, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000001_init_schema.down.sql"))
	if err != nil {
		t.Fatalf("read baseline down migration: %v", err)
	}
	downSQL := string(down)
	if !strings.Contains(downSQL, "DROP SCHEMA IF EXISTS public CASCADE") {
		t.Fatalf("baseline down migration should reset public schema, got:\n%s", downSQL)
	}
}
