package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBaselineMigrationSeedsDemoUsers(t *testing.T) {
	path := filepath.Join("..", "..", "migrations", "000001_init_schema.up.sql")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read baseline migration: %v", err)
	}

	sql := string(content)
	expectedSnippets := []string{
		"INSERT INTO public.users",
		"'admin'",
		"'teacher'",
		"'student'",
		"'student2'",
		"INSERT INTO public.user_roles",
	}
	for _, snippet := range expectedSnippets {
		if !strings.Contains(sql, snippet) {
			t.Fatalf("expected baseline migration to contain %q", snippet)
		}
	}
}
