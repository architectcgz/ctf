package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeNodeContractInBaseline(t *testing.T) {
	t.Parallel()

	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000001_init_schema.up.sql"))
	if err != nil {
		t.Fatalf("read baseline migration: %v", err)
	}
	upSQL := string(up)
	for _, snippet := range []string{
		"CREATE TABLE public.runtime_nodes",
		"CREATE TABLE public.instances",
		"node_id bigint",
		"ALTER TABLE ONLY public.instances",
		"ADD CONSTRAINT instances_node_id_fkey",
		"FOREIGN KEY (node_id) REFERENCES public.runtime_nodes(id)",
		"CREATE INDEX idx_instances_node_id ON public.instances USING btree (node_id)",
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

func TestRuntimeNodeLastSeenMigration(t *testing.T) {
	t.Parallel()

	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000019_add_runtime_node_last_seen_at.up.sql"))
	if err != nil {
		t.Fatalf("read runtime node last_seen_at up migration: %v", err)
	}
	upSQL := string(up)
	for _, snippet := range []string{
		"ALTER TABLE public.runtime_nodes",
		"ADD COLUMN last_seen_at timestamp with time zone",
		"idx_runtime_nodes_health_schedulable_seen",
	} {
		if !strings.Contains(upSQL, snippet) {
			t.Fatalf("runtime node last_seen_at up migration should contain %q, got:\n%s", snippet, upSQL)
		}
	}

	down, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000019_add_runtime_node_last_seen_at.down.sql"))
	if err != nil {
		t.Fatalf("read runtime node last_seen_at down migration: %v", err)
	}
	downSQL := string(down)
	for _, snippet := range []string{
		"DROP INDEX IF EXISTS public.idx_runtime_nodes_health_schedulable_seen",
		"DROP COLUMN IF EXISTS last_seen_at",
	} {
		if !strings.Contains(downSQL, snippet) {
			t.Fatalf("runtime node last_seen_at down migration should contain %q, got:\n%s", snippet, downSQL)
		}
	}
}
