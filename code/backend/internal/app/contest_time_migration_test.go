package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContestTimeContractInBaseline(t *testing.T) {
	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000001_init_schema.up.sql"))
	if err != nil {
		t.Fatalf("read baseline migration: %v", err)
	}
	upSQL := string(up)
	for _, snippet := range []string{
		"start_time timestamp with time zone NOT NULL",
		"end_time timestamp with time zone NOT NULL",
		"freeze_time timestamp with time zone",
	} {
		if !strings.Contains(upSQL, snippet) {
			t.Fatalf("baseline migration should contain %q, got:\n%s", snippet, upSQL)
		}
	}

	down, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000001_init_schema.down.sql"))
	if err != nil {
		t.Fatalf("read baseline down migration: %v", err)
	}
	if !strings.Contains(string(down), "DROP SCHEMA IF EXISTS public CASCADE") {
		t.Fatalf("baseline down migration should reset public schema, got:\n%s", string(down))
	}
}
