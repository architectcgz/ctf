package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImagesDeletedAtIndexInBaseline(t *testing.T) {
	t.Parallel()

	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000001_init_schema.up.sql"))
	if err != nil {
		t.Fatalf("read baseline migration: %v", err)
	}
	down, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000001_init_schema.down.sql"))
	if err != nil {
		t.Fatalf("read baseline down migration: %v", err)
	}
	if !strings.Contains(string(up), "CREATE INDEX idx_images_deleted_at ON public.images USING btree (deleted_at)") {
		t.Fatalf("baseline migration should create idx_images_deleted_at on deleted_at, got:\n%s", string(up))
	}
	if strings.Contains(string(up), "\\restrict ") || strings.Contains(string(up), "\\unrestrict ") {
		t.Fatalf("baseline migration must not contain psql-only restrict directives, got:\n%s", string(up))
	}
	if !strings.Contains(string(down), "DROP SCHEMA IF EXISTS public CASCADE") {
		t.Fatalf("baseline down migration should reset public schema, got:\n%s", string(down))
	}
}

func TestActiveHostPortIndexIgnoresZeroPort(t *testing.T) {
	t.Parallel()

	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000001_init_schema.up.sql"))
	if err != nil {
		t.Fatalf("read baseline migration: %v", err)
	}
	upSQL := string(up)
	if !strings.Contains(upSQL, "host_port > 0") {
		t.Fatalf("baseline migration should only enforce active host_port uniqueness for positive ports")
	}
	if strings.Contains(upSQL, "uk_instances_active_host_port ON public.instances USING btree (host_port) WHERE ((host_port IS NOT NULL)") {
		t.Fatalf("baseline migration should not treat host_port=0 as an active published port")
	}
}

func TestBaselineSeedsDefaultLocalUsers(t *testing.T) {
	t.Parallel()

	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000001_init_schema.up.sql"))
	if err != nil {
		t.Fatalf("read baseline migration: %v", err)
	}
	upSQL := string(up)
	for _, snippet := range []string{
		"INSERT INTO public.roles (code, name, description)",
		"INSERT INTO public.users (",
		"INSERT INTO public.user_roles (user_id, role_id)",
		"'Platform Admin'",
		"'Demo Teacher'",
		"'Demo Student'",
		"'admin'",
		"'teacher'",
		"'student'",
		"'student2'",
	} {
		if !strings.Contains(upSQL, snippet) {
			t.Fatalf("baseline migration should keep default local seed snippet %q", snippet)
		}
	}
}
