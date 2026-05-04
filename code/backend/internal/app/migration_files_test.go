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
	fixUp, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000004_fix_active_host_port_index.up.sql"))
	if err != nil {
		t.Fatalf("read host port index migration: %v", err)
	}

	for name, sql := range map[string]string{
		"baseline": string(up),
		"000004":   string(fixUp),
	} {
		if !strings.Contains(sql, "host_port > 0") {
			t.Fatalf("%s migration should only enforce active host_port uniqueness for positive ports", name)
		}
		if strings.Contains(sql, "uk_instances_active_host_port ON public.instances USING btree (host_port) WHERE ((host_port IS NOT NULL)") {
			t.Fatalf("%s migration should not treat host_port=0 as an active published port", name)
		}
	}
}
